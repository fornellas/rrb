package runner

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/alessio/shellescape"
	"github.com/williammartin/subreaper"

	"github.com/fornellas/rrb/process"
)

type Runner struct {
	Name   string
	Args   []string
	cmdStr string
	idleCn chan struct{}
	cmd    *exec.Cmd
}

func NewRunner(name string, args ...string) *Runner {
	escapedCmd := []string{}
	for _, s := range append([]string{name}, args...) {
		escapedCmd = append(escapedCmd, shellescape.Quote(s))
	}

	r := Runner{
		Name:   name,
		Args:   args,
		cmdStr: strings.Join(escapedCmd, " "),
		idleCn: make(chan struct{}),
	}
	go func() { r.idleCn <- struct{}{} }()
	return &r
}

func waitStatusStr(waitStatus syscall.WaitStatus) string {
	res := ""
	switch {
	case waitStatus.Exited():
		if waitStatus.ExitStatus() == 0 {
			res = "exited successfully"
		} else {
			res = "exited with status " + strconv.Itoa(waitStatus.ExitStatus())
		}
	case waitStatus.Signaled():
		res = "received signal: " + waitStatus.Signal().String()
	case waitStatus.Stopped():
		res = "received stop signal: " + waitStatus.StopSignal().String()
		if waitStatus.StopSignal() == syscall.SIGTRAP && waitStatus.TrapCause() != 0 {
			res += " (trap " + strconv.Itoa(waitStatus.TrapCause()) + ")"
		}
	case waitStatus.Continued():
		res = "continued"
	}
	if waitStatus.CoreDump() {
		res += " (core dumped)"
	}
	return res
}

func terminateChildren() error {
	selfProcess, err := process.SelfProcess()
	if err != nil {
		return err
	}

	if len(selfProcess.Children) == 0 {
		return nil
	}

	waitErrCh := make(chan error)

	go func() {
		var waitStatus syscall.WaitStatus
		var rusage syscall.Rusage
		var err error
		for {
			var wpid int
			log.Printf("Waiting for children...")
			wpid, err = syscall.Wait4(-1, &waitStatus, 0, &rusage)
			if err != nil {
				if err == syscall.ECHILD {
					err = nil
				}
				break
			}
			if wpid == -1 {
				break
			}
			log.Printf("Child %d %s", wpid, waitStatusStr(waitStatus))
		}
		log.Printf("No more children left")
		waitErrCh <- err
	}()

	log.Printf("Orphan process left behind, terminating them")
	for _, childProcess := range selfProcess.Children {
		fmt.Printf("%s", childProcess.SprintTree(0))
		// FIXME send to process group
		log.Printf("Sending SIGTERM to %s", childProcess)
		_ = childProcess.Signal(syscall.SIGTERM)
	}

no_more_children:
	for {
		select {
		case <-time.After(500 * time.Millisecond):
			selfProcess, err := process.SelfProcess()
			if err != nil {
				return err
			}
			if len(selfProcess.Children) == 0 {
				break no_more_children
			}

			log.Printf("Orphan process still alive, sending SIGKILL")
			for _, childProcess := range selfProcess.Children {
				fmt.Printf("%s", childProcess.SprintTree(0))
				log.Printf("Sending SIGKILL to %s", childProcess)
				_ = childProcess.Signal(syscall.SIGTERM)
			}
		case err = <-waitErrCh:
			break no_more_children
		}
	}

	if err != nil {
		log.Printf("wait error: %s", err)
	}

	return fmt.Errorf("Main process left orphan process behind")
}

func (r *Runner) Run() error {
	log.Printf("Run()")

ready:
	for {
		log.Printf("waiting for idle...")
		select {
		case <-r.idleCn:
			log.Printf("Idle! We can run!")
			break ready
		default:
			log.Printf("Still running, we have to wait!")
			// TODO kill
		}
	}

	// This ensures that orphan process will become children of the process
	// that called Start(), so we can babysit then.
	if err := subreaper.Set(); err != nil {
		return err
	}

	cmd := exec.Command(r.Name, r.Args...)
	cmd.Env = os.Environ()
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid: true,
	}
	log.Printf("> %s", r.cmdStr)
	if err := cmd.Start(); err != nil {
		return err
	}
	r.cmd = cmd

	go func() {
		if err := r.cmd.Wait(); err != nil {
			log.Printf("Wait(): %s", err)
		}
		if r.cmd.ProcessState.Success() {
			log.Printf("Success!")
		} else {
			log.Printf("Failure!")
		}

		if err := terminateChildren(); err != nil {
			log.Printf("terminateChildren: %s", err)
		}

		r.idleCn <- struct{}{}
	}()

	return nil
}
