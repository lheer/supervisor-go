package main

import (
	"github.com/looplab/fsm"
)

type (
	// Top-level toml config
	ConfigFile struct {
		Name     string
		Programs map[string]ProgramConfig
	}
	// A program as defined in toml file
	ProgramConfig struct {
		Command     string
		Priority    int
		Oneshot     bool
		Autorestart bool
		Startsecs   uint
		id int
	}

	ProcessState string

	// Struct to keep track of a process and its state
	Process struct {
		pid       int
		id int
		FSM *fsm.FSM
	}

	ProcessEvent struct {
		id int
		pid int
		exit_code int // only valid in 
		new_state ProcessState
	}
)

// The states a process can be in
const (
	Starting ProcessState = "starting"
	Running ProcessState = "running"
	Exited ProcessState = "exited"
)
