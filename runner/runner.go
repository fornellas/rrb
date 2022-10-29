package runner

import (
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/alessio/shellescape"
)

type Runner struct {
	Name   string
	Args   []string
	cmdStr string
	waitCn chan struct{}
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
		waitCn: make(chan struct{}),
	}
	go func() { r.waitCn <- struct{}{} }()
	return &r
}

func (r *Runner) Run() error {
	// log.Printf("Run()")

	if r.cmd != nil && r.cmd.ProcessState == nil {
		// TODO kill
		log.Printf("Killing")
	}

	<-r.waitCn

	cmd := exec.Command(r.Name, r.Args...)
	cmd.Env = os.Environ()
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

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
		r.waitCn <- struct{}{}
	}()

	return nil
}
