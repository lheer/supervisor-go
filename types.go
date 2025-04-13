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
		Period       string
		key          string
		hasRun       bool
	}

	ProcessState string

	ProcessEvent struct {
		key       string
		pid       int // only valid in state starting and running
		exitCode int // only valid in state exited
		newState ProcessState
	}
)

// The states a process can be in
const (
	NotRunning ProcessState = "not_running"
	Starting   ProcessState = "starting"
	Running    ProcessState = "running"
	Exited     ProcessState = "exited"
)

// JSON interface type definitions
type (
	ProcessInfo struct {
		State    ProcessState `json:"state"`
		ExitCode string          `json:"exit_code"`
	}

 	SystemState struct {
		mu    sync.Mutex
		state map[string]ProcessInfo
	}
)
