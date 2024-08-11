package main

type (
	// Top-level toml config
	ConfigFile struct {
		Name     string
		Programs map[string]ProgramConfig
	}
	// A program as defined in toml file
	ProgramConfig struct {
		Command      string
		Oneshot      bool
		Autorestart  bool
		Startsecs    uint
		Startretries uint
		After        string
		key          string
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
	Starting ProcessState = "starting"
	Running  ProcessState = "running"
	Exited   ProcessState = "exited"
)
