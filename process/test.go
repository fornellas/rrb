package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"syscall"
	"time"

	"github.com/prometheus/procfs"
	"github.com/williammartin/subreaper"
)

type PID int

type Process struct {
	PID  PID
	Proc procfs.Proc
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
		procStatus, err := proc.NewStatus()
		if err != nil {
			continue
		}
		if procStatus.UIDs[0] != "1000" {
			continue
		}

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
		process := &Process{
			PID:  PID(proc.PID),
			Proc: proc,
		}
		processGroup.Processes[process.PID] = process
	}

	return &sessions, nil
}

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

func main() {
	if err := subreaper.Set(); err != nil {
		log.Println("Fail: subreaper.Set()")
		log.Fatal(err)
	}

	cmd := exec.Command(
		"bash",
		"-c",
		"{ setsid sleep 2222 && echo sleep bg ; } & echo fg",
	)
	cmd.Env = os.Environ()
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid: true,
	}

	log.Println("Start()")
	if err := cmd.Start(); err != nil {
		log.Println("Fail: Start()")
		log.Fatal(err)
	}

	log.Println("Wait()")
	if err := cmd.Wait(); err != nil {
		log.Println("Fail: Wait()")
		log.Fatal(err)
	}

	if !cmd.ProcessState.Success() {
		log.Fatal("!Success()")
	}
	log.Println("Success()")

	sessions, err := GetSessions()
	if err != nil {
		log.Println("Fail: GetSessions()")
		log.Fatal(err)
	}
	sessions.Print()

	time.Sleep(time.Hour)

	log.Println("Exit")
}
