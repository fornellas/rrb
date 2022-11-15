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
	waitCn chan struct{}
	killCn chan struct{}
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
		waitCn: make(chan struct{}),
		killCn: make(chan struct{}),
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
		// FIXME configurable value
		case <-time.After(3 * time.Second):
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
				_ = childProcess.Signal(syscall.SIGKILL)
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

	select {
	case <-r.idleCn:
		break
	default:
		log.Printf("Killing...")
		r.killCn <- struct{}{}
		<-r.idleCn
	}

	// This ensures that orphan process will become children of the process
	// that called Start(), so we can babysit then.
	if err := subreaper.Set(); err != nil {
		r.idleCn <- struct{}{}
		return err
	}

	r.cmd = exec.Command(r.Name, r.Args...)
	r.cmd.Env = os.Environ()
	r.cmd.Stdin = os.Stdin
	r.cmd.Stdout = os.Stdout
	r.cmd.Stderr = os.Stderr
	r.cmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid: true,
	}
	log.Printf("> %s", r.cmdStr)
	if err := r.cmd.Start(); err != nil {
		r.idleCn <- struct{}{}
		return err
	}

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
			if r.cmd.ProcessState.Success() {
				log.Printf("Error: %s", err)
			} else {
				log.Printf("Warning: %s", err)
			}
		}

		r.waitCn <- struct{}{}
	}()

	go func() {
		select {
		case <-r.waitCn:
		case <-r.killCn:
			_ = r.cmd.Process.Signal(syscall.SIGTERM)
			select {
			case <-r.waitCn:
			// FIXME configurable value
			case <-time.After(3 * time.Second):
				_ = r.cmd.Process.Signal(syscall.SIGTERM)
				<-r.waitCn
			}
		}

		r.idleCn <- struct{}{}
	}()

	return nil
}
