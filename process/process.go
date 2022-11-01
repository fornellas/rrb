package main

import (
	"fmt"

	"github.com/prometheus/procfs"
)

type PID int

type Process struct {
	PID      PID
	Children map[PID]*Process
	Proc     procfs.Proc
}

type PGID int

type ProcessGroup struct {
	PGID      PGID
	Processes map[PID]*Process
}

type SID int

type Session struct {
	SID           SID
	ProcessGroups map[PGID]*ProcessGroup
}

func (s *Session) Print() {
	fmt.Printf("Session %d:\n", s.SID)
	for _, processGroup := range s.ProcessGroups {
		fmt.Printf("  Process group %d:\n", processGroup.PGID)
		for _, process := range processGroup.Processes {
			comm, err := process.Proc.Comm()
			if err != nil {
				continue
			}
			fmt.Printf("    Process %d: %s\n", process.PID, comm)
		}
	}
}

type Sessions map[SID]*Session

func (sessions *Sessions) Print() {
	for _, session := range *sessions {
		session.Print()
	}
}

func (sessions *Sessions) GetSession(pid PID) *Session {
	for _, session := range *sessions {
		for _, processGroup := range session.ProcessGroups {
			if _, ok := processGroup.Processes[pid]; ok {
				return session
			}
		}
	}
	return nil
}

func GetSessions() (*Sessions, error) {
	fs, err := procfs.NewDefaultFS()
	if err != nil {
		return nil, err
	}

	procs, err := fs.AllProcs()
	if err != nil {
		return nil, err
	}

	sessions := Sessions{}

	for _, proc := range procs {
		procStat, err := proc.Stat()
		if err != nil {
			continue
		}

		sid := SID(procStat.Session)

		session, ok := sessions[sid]
		if !ok {
			session = &Session{
				SID:           sid,
				ProcessGroups: map[PGID]*ProcessGroup{},
			}
			sessions[sid] = session
		}

		pgid := PGID(procStat.PGRP)
		processGroup, ok := session.ProcessGroups[pgid]
		if !ok {
			processGroup = &ProcessGroup{
				PGID:      pgid,
				Processes: map[PID]*Process{},
			}
			session.ProcessGroups[pgid] = processGroup
		}
		pid := PID(proc.PID)
		processGroup.Processes[pid] = &Process{
			PID:      pid,
			Children: map[PID]*Process{}, // FIXME
			Proc:     proc,
		}
	}

	return &sessions, nil
}
