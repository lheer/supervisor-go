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
		Priority    int
		Oneshot     bool
		Autorestart bool
		Startsecs   uint
		id          int
	}

	ProcessState string

	ProcessEvent struct {
		id        int
		pid       int
		exit_code int // only valid in
		new_state ProcessState
	}
)

// The states a process can be in
const (
	Starting ProcessState = "starting"
	Running  ProcessState = "running"
	Exited   ProcessState = "exited"
)
