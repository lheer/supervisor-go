package main

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"strings"
)

func run_program(prg *ProgramConfig) {

	fmt.Printf("[%s]: Starting\n", prg.Command)

	splitted := strings.Split(prg.Command, " ")
	pHandle := exec.Command(splitted[0], splitted[1:]...)

	// Create pipes for stdout and stderr
	stdoutPipe, err := pHandle.StdoutPipe()
	if err != nil {
		fmt.Printf("Error creating stdout pipe: %v\n", err)
	}
	stderrPipe, err := pHandle.StderrPipe()
	if err != nil {
		fmt.Printf("Error creating stderr pipe: %v\n", err)
	}

	// Start the process
	if err := pHandle.Start(); err != nil {
		fmt.Printf("[%s]: Error starting program\n", prg.Command)
	}

	// Read stdout / stderr and print
	go func(reader io.Reader) {
		scanner := bufio.NewScanner(reader)
		for scanner.Scan() {
			fmt.Printf("[%s]: %s\n", prg.Command, scanner.Text())
		}
		if err := scanner.Err(); err != nil {
			fmt.Printf("[%s]: Error reading stdout: %v\n", prg.Command, err)
		}
	}(stdoutPipe)

	// Function to read from pipe and print to stderr with "info" prefix
	go func(reader io.Reader) {
		scanner := bufio.NewScanner(reader)
		for scanner.Scan() {
			fmt.Printf("[%s]: %s\n", prg.Command, scanner.Text())
		}
		if err := scanner.Err(); err != nil {
			fmt.Printf("[%s]: Error reading stderr: %v\n", prg.Command, err)
		}
	}(stderrPipe)

	// Wait for the command to finish
	if err := pHandle.Wait(); err != nil {
		fmt.Printf("[%s]: Finished with error: %v\n", prg.Command, err)
	}
}
