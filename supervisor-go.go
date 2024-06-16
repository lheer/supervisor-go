package main

import (
	"cmp"
	"fmt"
	"os"
	"slices"

	"github.com/BurntSushi/toml"
)

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

	// Start programs ordered by priority
	for _, ptr := range progPointers {
		fmt.Printf("[main]: Starting %s\n", ptr.Command)
		go run_program(ptr)
	}
}
