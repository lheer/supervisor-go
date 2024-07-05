package main

import (
	"cmp"
	"fmt"
	"os"
	"slices"

	"github.com/BurntSushi/toml"
)

func getProgramById(prgs []*ProgramConfig, id int) *ProgramConfig {
	for _, p := range prgs {
		if p.id == id {
			return p
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

	// Create slice of pointers and sort by priority
	var progPointers []*ProgramConfig
	for _, p := range cfg.Programs {
		p := p
		progPointers = append(progPointers, &p)
	}

	slices.SortFunc(progPointers, func(a, b *ProgramConfig) int {
		return cmp.Compare(a.Priority, b.Priority)
	})

	backchannel := make(chan ProcessEvent, len(cfg.Programs))

	running := 0
	// Start programs ordered by priority
	for i, ptr := range progPointers {
		ptr.id = i
		go RunProgram(ptr, backchannel)
		running++
	}

	for {
		event := <-backchannel

		program := getProgramById(progPointers, event.id)
		if program == nil {
			// Fatal
			fmt.Fprintf(os.Stderr, "Internal error: Could not get program with id=%d", event.id)
			break
		}

		if event.new_state == Exited {
			fmt.Printf("Exited: %s\n", program.Command)
			running--; program.Startretries--

			if program.Autorestart && program.Startretries > 0 {
				go RunProgram(program, backchannel)
				running++
				fmt.Printf("Restarted: %s\n", program.Command)
			}
		} else if event.new_state == Starting {
			fmt.Printf("Starting: %s\n", program.Command)
		} else if event.new_state == Running {
			fmt.Printf("Running: %s\n", program.Command)
		} else {
			fmt.Fprintf(os.Stderr, "Internal error: Invalid event with new_state=%s", event.new_state)
		}

		if running == 0 {
			break
		}
	}

	fmt.Println("Exit")
}
