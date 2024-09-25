package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
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

// Given a slice of programs, create an execution graph from it
func createExecutionGraph(programs []ProgramConfig) (*Graph[*ProgramConfig], error) {
	ProgramGraph := NewGraph[*ProgramConfig]()
	for i := range programs {
		err := ProgramGraph.AddVertex(&programs[i])
		if err != nil {
			return nil, fmt.Errorf("error creating execution graph while adding vertex %+v, error: %s", programs[i], err.Error())
		}
	}

	for i := range programs {
		prg := &programs[i]
		if prg.After != "" {
			// Has successor, get predeccessor and add edge in graph
			pre := getProgramByKey(programs, prg.After)
			if pre == nil {
				return nil, fmt.Errorf("error creating execution graph while adding edge %+v, check configuration", programs[i])
			}
			ProgramGraph.AddEdge(pre, &programs[i])
		}
	}
	return ProgramGraph, nil
}

func HTTPHandler(programstate *SystemState) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		programstate.mu.Lock()
		defer programstate.mu.Unlock()
		jsonstate, err := json.Marshal(programstate.state)
		if err != nil {
			fmt.Fprint(w, err.Error())
		} else {
			fmt.Fprint(w, string(jsonstate))
		}
	}
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

	// Create graph containing pointers to programs
	ProgramGraph, err := createExecutionGraph(programs)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error while creating execution graph: %s", err.Error())
		os.Exit(1)
	}

	// Keep track of program states
	statemap := &SystemState{
		state: make(map[string]ProcessState),
	}
	for _, v := range programs {
		statemap.state[v.key] = NotRunning
	}

	mux := http.NewServeMux()
	mux.Handle("GET /state", HTTPHandler(statemap))

	server := &http.Server{
		Addr:    ":8888",
		Handler: mux,
	}
	go server.ListenAndServe()

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
		if program == nil {
			fmt.Fprintf(os.Stderr, "Internal error: Unable to retrieve program with key=%s", event.key)
			continue
		}

		// Update state
		ret := func() bool {
			statemap.mu.Lock()
			defer statemap.mu.Unlock()
			_, ok := statemap.state[program.key]
			if !ok {
				fmt.Fprintf(os.Stderr, "Internal error: Unable to retrieve program state with key=%s", program.key)
				return false
			}
			statemap.state[program.key] = event.new_state
			return true
		}()
		if !ret {
			continue
		}

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

	server.Close()
	fmt.Println("Exit")
}
