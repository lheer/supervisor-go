package main

import (
	"flag"
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

// Start the successors of a program and return the number of started programs
func startSuccessors(g *Graph[*ProgramConfig], prg *ProgramConfig, c chan<- ProcessEvent) int {
	running := 0
	successors := g.GetSuccessors(prg)
	for _, p := range successors {
		if !p.hasRun {
			go RunProgram(p, c)
			running++
		}
	}
	return running
}

func parseConfigFile(cfgFile *string) ([]ProgramConfig, error) {
	var cfg ConfigFile
	md, err := toml.DecodeFile(*cfgFile, &cfg)
	if err != nil {
		return []ProgramConfig{}, err
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

	// If no restart counter is given, set it to -1
	for i := range programs {
		prg := &programs[i]
		defined := md.IsDefined("programs", prg.key, "startretries")
		if !defined {
			prg.Startretries = -1
		}
	}
	return programs, nil
}

func main() {
	configFile := flag.String("c", "", "Configuration file to use")
	flag.Parse()

	if *configFile == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	// Parse toml config file
	programs, err := parseConfigFile(configFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to parse config file: %s", err.Error())
		os.Exit(1)
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
			if pre == nil {
				fmt.Fprintf(os.Stderr, "Error creating execution graph while adding edge %+v, check configuration\n", programs[i])
				os.Exit(1)
			}
			ProgramGraph.AddEdge(pre, &programs[i])
		}
	}

	backchannel := make(chan ProcessEvent, len(programs))

	// Start: Launch all root nodes
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
			program.hasRun = true

			if program.Autorestart && (program.Startretries == -1 || program.Startretries != 0) {
				// Restart if configured
				program.Startretries--
				go RunProgram(program, backchannel)
				running++
				fmt.Printf("Restarted: %s\n", program.key)
			} else if event.exit_code == 0 {
				// Program has finished with exit code 0, start successors
				running += startSuccessors(ProgramGraph, program, backchannel)
			} else {
				fmt.Fprintf(os.Stderr, "Failed to start %s, giving up\n", program.key)
			}
		} else if event.new_state == Starting {
			fmt.Printf("Starting: %s\n", program.key)
		} else if event.new_state == Running {
			fmt.Printf("Running: %s\n", program.key)
			// Program is up and running, start successors
			running += startSuccessors(ProgramGraph, program, backchannel)
		} else {
			fmt.Fprintf(os.Stderr, "Internal error: Invalid event with new_state=%s", event.new_state)
		}

		if running == 0 {
			break
		}
	}

	fmt.Println("Exit")
}
