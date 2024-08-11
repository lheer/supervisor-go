package main

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
)

func getProgramByKey(prgs []ProgramConfig, key string) *ProgramConfig {
	for i := range prgs {
		if prgs[i].key == key {
			return &prgs[i]
		}
	}
	return nil
}

func main() {
	f := "examples/config.toml"

	// Parse toml config file
	var cfg ConfigFile
	_, err := toml.DecodeFile(f, &cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// Copy program identifier over to struct
	for k := range cfg.Programs {
		entry := cfg.Programs[k]
		entry.key = k
		cfg.Programs[k] = entry
	}

	// Create slice of programs (to populate graph later)
	var programs []ProgramConfig
	for _, prg := range cfg.Programs {
		programs = append(programs, prg)
	}

	// Create graph
	ProgramGraph := NewGraph[*ProgramConfig]()
	for i := range programs {
		err = ProgramGraph.AddVertex(&programs[i])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating execution graph while adding vertex %+v\n", programs[i])
			os.Exit(1)
		}
	}

	for i := range programs {
		prg := &programs[i]
		if prg.After != "" {
			// Has successor, get predeccessor and add edge in graph
			pre := getProgramByKey(programs, prg.After)
			ProgramGraph.AddEdge(pre, prg)
		}
	}

	// Start: Launch all root nodes
	backchannel := make(chan ProcessEvent, len(cfg.Programs))

	running := 0
	for _, prg := range ProgramGraph.GetRootNodes() {
		go RunProgram(prg, backchannel)
		running++
	}

	for {
		event := <-backchannel
		program := getProgramByKey(programs, event.key)

		if event.new_state == Exited {
			fmt.Printf("Exited: %s\n", program.key)
			running--
			program.Startretries--

			if program.Autorestart && program.Startretries > 0 {
				go RunProgram(program, backchannel)
				running++
				fmt.Printf("Restarted: %s\n", program.key)
			}
		} else if event.new_state == Starting {
			fmt.Printf("Starting: %s\n", program.key)
		} else if event.new_state == Running {
			fmt.Printf("Running: %s\n", program.key)
		} else {
			fmt.Fprintf(os.Stderr, "Internal error: Invalid event with new_state=%s", event.new_state)
		}

		if running == 0 {
			break
		}
	}

	fmt.Println("Exit")
}
