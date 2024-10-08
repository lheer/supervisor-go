package main

import "sync"

type (
	// Top-level toml config
	ConfigFile struct {
		Name     string
		Server   string
		Programs map[string]ProgramConfig
	}
	// A program as defined in toml file
	ProgramConfig struct {
		Command      string
		Autorestart  bool
		Startsecs    uint
		Startretries int
		After        string
		key          string
		hasRun       bool
	}

	ProcessState string

	ProcessEvent struct {
		key       string
		pid       int // only valid in state starting and running
		exit_code int // only valid in state exited
		new_state ProcessState
	}
)

// The states a process can be in
const (
	NotRunning ProcessState = "not_running"
	Starting   ProcessState = "starting"
	Running    ProcessState = "running"
	Exited     ProcessState = "exited"
)

type SystemState struct {
	mu    sync.Mutex
	state map[string]ProcessState
}
