package main

import (
	"fmt"
	"os/exec"
	"syscall"
	"time"

	"os"

	"github.com/AstromechZA/gaze/conf"
)

type GazeReport struct {
	StartTime       time.Time `json:"start_time"`
	EndTime         time.Time `json:"start_time"`
	ExitCode        int       `json:"start_time"`
	CapturedOutput  string    `json:"captured_output"`
	ExitDescription string    `json:"exit_description"`
}

func runReport(args []string, config *conf.GazeConfig, name string) (*GazeReport, error) {
	output := new(GazeReport)
	output.StartTime = time.Now()
	output.ExitCode = 0
	output.ExitDescription = "No description added"

	// run command
	cmd := exec.Command(args[0], args[1:]...)

	// send process stdin to subprocess
	cmd.Stdin = os.Stdin

	// https://github.com/tleyden/elastic-thought/blob/master/os_util.go#L119

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	// TODO - how do we capture this too

	err := cmd.Start()
	if err != nil {
		fmt.Printf(">>> %v\n", err.Error())
	} else {

		err = cmd.Wait()
		if err != nil {
			exitErr, ok := err.(*exec.ExitError)
			output.ExitCode = 127
			output.ExitDescription = "Execution failed"
			if ok {
				if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
					output.ExitCode = status.ExitStatus()
					output.ExitDescription = fmt.Sprintf("Execution failed with code %d", output.ExitCode)
				}
			} else {
				output.ExitDescription = fmt.Sprintf("Unexpected error: %v", err.Error())
			}
		} else {
			output.ExitDescription = "Execution finished with no error"
		}
	}

	output.EndTime = time.Now()
	return output, nil
}
