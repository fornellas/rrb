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
	cmd    *exec.Cmd
}

func NewRunner(name string, args ...string) Runner {
	escapedCmd := []string{}
	for _, s := range append([]string{name}, args...) {
		escapedCmd = append(escapedCmd, shellescape.Quote(s))
	}

	r := Runner{
		Name:   name,
		Args:   args,
		cmdStr: strings.Join(escapedCmd, " "),
	}
	return r
}

func (r *Runner) Run() error {
	// if r.cmd != nil {
	// 	// TODO kill
	// 	// wait
	// }

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

	return nil
}
