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
	KillWait time.Duration
	Name     string
	Args     []string
	cmdStr   string
	idleCn   chan struct{}
	waitCn   chan struct{}
	killCn   chan struct{}
	cmd      *exec.Cmd
}

func NewRunner(killWait time.Duration, name string, args ...string) *Runner {
	escapedCmd := []string{}
	for _, s := range append([]string{name}, args...) {
		escapedCmd = append(escapedCmd, shellescape.Quote(s))
	}

	r := Runner{
		KillWait: killWait,
		Name:     name,
		Args:     args,
		cmdStr:   strings.Join(escapedCmd, " "),
		idleCn:   make(chan struct{}),
		waitCn:   make(chan struct{}),
		killCn:   make(chan struct{}),
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

func waitChildren(waitChildrenErrCh chan error) {
	var waitStatus syscall.WaitStatus
	var rusage syscall.Rusage
	var err error
	for {
		var wpid int
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
	waitChildrenErrCh <- err
}

func (r *Runner) syncWaitAndKill(waitChildrenErrCh chan error) error {
	for {
		select {
		case <-time.After(r.KillWait):
			selfProcess, err := process.SelfProcess()
			if err != nil {
				return err
			}
			if len(selfProcess.Children) == 0 {
				return nil
			}

			for _, childProcess := range selfProcess.Children {
				// fmt.Printf("%s", childProcess.SprintTree(0))
				log.Printf("Sending SIGKILL to %s...", childProcess)
				_ = childProcess.Signal(syscall.SIGKILL)
			}

		case err := <-waitChildrenErrCh:
			return err
		}
	}
}

func (r *Runner) killChildren() error {
	selfProcess, err := process.SelfProcess()
	if err != nil {
		return err
	}

	if len(selfProcess.Children) == 0 {
		return nil
	}

	waitChildrenErrCh := make(chan error)

	go waitChildren(waitChildrenErrCh)

	log.Printf("Orphan process left behind!")
	for _, childProcess := range selfProcess.Children {
		// fmt.Printf("%s", childProcess.SprintTree(0))
		// FIXME send to process group
		log.Printf("Sending SIGTERM to %s...", childProcess)
		_ = childProcess.Signal(syscall.SIGTERM)
	}

	err = r.syncWaitAndKill(waitChildrenErrCh)
	if err != nil {
		log.Printf("wait error: %s", err)
	}

	return fmt.Errorf("Orphan process behind")
}

func (r *Runner) waitAll() {
	_ = r.cmd.Wait()

	if r.cmd.ProcessState.Success() {
		log.Printf("Success: %s", r.cmd.ProcessState)
	} else {
		log.Printf("Failure: %s", r.cmd.ProcessState)
	}

	if err := r.killChildren(); err != nil {
		if r.cmd.ProcessState.Success() {
			log.Printf("Failure: %s", err)
		} else {
			log.Printf("Warning: %s", err)
		}
	}

	r.waitCn <- struct{}{}
}

func (r *Runner) sendIdle() {
	select {
	case <-r.waitCn:
	case <-r.killCn:
		_ = r.cmd.Process.Signal(syscall.SIGTERM)
		select {
		case <-r.waitCn:
		case <-time.After(r.KillWait):
			_ = r.cmd.Process.Signal(syscall.SIGTERM)
			<-r.waitCn
		}
	}

	r.idleCn <- struct{}{}
}

func (r *Runner) Run() error {
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

	go r.waitAll()
	go r.sendIdle()

	return nil
}
