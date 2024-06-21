package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"

	"github.com/looplab/fsm"
)

// Used to capture stdout/stderr from process and annotate it with some prefix
// before printing.
func pipe_output(reader io.Reader, prefix string, isStderr bool) {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		if isStderr {
			fmt.Fprintf(os.Stderr, "%s %s\n", prefix, scanner.Text())
		} else {
			fmt.Printf("%s %s\n", prefix, scanner.Text())
		}
	}
}

func newProcess(id int) *Process {
	p := &Process{}
	p.id = id
	p.FSM = fsm.NewFSM(
		string(Exited),
		fsm.Events{
			{Name: "start", Src: []string{"Exited"}, Dst: "Starting"},
			{Name: "start_fail", Src: []string{"Starting"}, Dst: "Exited"},
			{Name: "start_success", Src: []string{"Starting"}, Dst: "Running"},
			{Name: "exit", Src: []string{"Running"}, Dst: "Exited"},
		},
		fsm.Callbacks{},
	)

	return p
}

func RunProgram(prg *ProgramConfig, backchannel chan ProcessEvent) {

	fmt.Printf("[%s]: Starting\n", prg.Command)

	pHandle := exec.Command("sh", "-c", prg.Command)

	// Create pipes for stdout and stderr
	stdoutPipe, err := pHandle.StdoutPipe()
	if err != nil {
		fmt.Printf("Error creating stdout pipe: %v\n", err)
	}
	stderrPipe, err := pHandle.StderrPipe()
	if err != nil {
		fmt.Printf("Error creating stderr pipe: %v\n", err)
	}

	// Read stdout / stderr and print
	go pipe_output(stdoutPipe, "["+prg.Command+"]: ", false)
	go pipe_output(stderrPipe, "["+prg.Command+"]: ", true)

	// Start the process
	if err := pHandle.Start(); err != nil {
		fmt.Printf("[%s]: Error starting program\n", prg.Command)
		return
	}

	backchannel <- ProcessEvent{id: prg.id, pid: pHandle.Process.Pid, exit_code: 0, new_state: Starting}

	// Only after startsecs, process is considered up and running
	timer := time.NewTimer(time.Second * time.Duration(prg.Startsecs))
	go func() {
		<-timer.C
		backchannel <- ProcessEvent{id: prg.id, new_state: Running}
	}()

	// Wait for the command to finish
	if err := pHandle.Wait(); err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			backchannel <- ProcessEvent{id: prg.id, exit_code: exiterr.ExitCode(), new_state: Exited}
		}
		fmt.Printf("[%s]: Finished with error: %v\n", prg.Command, err)
	}

	timer.Stop()

	// all good, exit code 0
	backchannel <- ProcessEvent{id: prg.id, exit_code: pHandle.ProcessState.ExitCode(), new_state: Exited}
}
