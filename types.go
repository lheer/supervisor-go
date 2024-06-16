package main

type (
	// Top-level toml config
	ConfigFile struct {
		Name     string
		Programs map[string]ProgramConfig
	}
	// A program as defined in toml file
	ProgramConfig struct {
		Command     string
		Priority    uint8
		Oneshot     bool
		Autorestart bool
		Startsecs   uint
	}

	ProcessState int

	// Struct to keep track of a process and its state
	Process struct {
		pid       int
		state     ProcessState
		exit_code int
	}
)

// The states a process can be in
const (
	Starting ProcessState = iota
	Running
	Exited
)
