package process

import (
	"fmt"
	"os"
	"strings"

	"github.com/prometheus/procfs"
)

type PID int
type Session int

type Process struct {
	// The process ID.
	PID PID
	// The filename of the executable.
	Comm string
	// The PID of the parent of this process.
	PPID PID
	// The process group ID of the process.
	PGRP PID
	// The session ID of the process.
	Session Session
	// Children processes.
	Children []*Process
}

func (p *Process) String() string {
	return fmt.Sprintf("%s(%d)", p.Comm, p.PID)
}

func (p *Process) SprintTree(indent int) string {
	res := fmt.Sprintf("%s%s\n", strings.Repeat("  ", indent), p)
	for _, childProcess := range p.Children {
		res += childProcess.SprintTree(indent + 1)
	}
	return res
}

func (p *Process) Signal(sig os.Signal) error {
	osProc, err := os.FindProcess(int(p.PID))
	if err != nil {
		return err
	}
	err = osProc.Signal(sig)
	if err != nil {
		return err
	}
	return nil
}

type ProcessMap map[PID]*Process

func GetProcessMap() (ProcessMap, error) {
	fs, err := procfs.NewDefaultFS()
	if err != nil {
		return nil, err
	}

	allProcs, err := fs.AllProcs()
	if err != nil {
		return nil, err
	}

	processMap := ProcessMap{}
	for _, proc := range allProcs {
		procStat, err := proc.Stat()
		if err != nil {
			continue
		}
		pid := PID(proc.PID)
		process := Process{
			PID:      pid,
			Comm:     procStat.Comm,
			PPID:     PID(procStat.PPID),
			PGRP:     PID(procStat.PGRP),
			Session:  Session(procStat.Session),
			Children: []*Process{},
		}
		processMap[pid] = &process
	}

	for _, process := range processMap {
		parentProcess, ok := processMap[process.PPID]
		if !ok {
			continue
		}
		parentProcess.Children = append(parentProcess.Children, process)
	}

	return processMap, nil
}

func SelfProcess() (*Process, error) {
	fs, err := procfs.NewDefaultFS()
	if err != nil {
		return nil, err
	}

	selfProc, err := fs.Self()
	if err != nil {
		return nil, err
	}

	processMap, err := GetProcessMap()
	if err != nil {
		return nil, err
	}
	return processMap[PID(selfProc.PID)], nil
}
