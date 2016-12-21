package main

import (
	"bufio"
	"fmt"
	"os/exec"
	"syscall"
	"time"

	"os"

	"bytes"

	"io"

	"github.com/AstromechZA/gaze/conf"
)

type GazeReport struct {
	StartTime       time.Time `json:"start_time"`
	EndTime         time.Time `json:"start_time"`
	ExitCode        int       `json:"start_time"`
	CapturedOutput  string    `json:"captured_output"`
	ExitDescription string    `json:"exit_description"`
}

func streamToFile(r io.Reader, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	w := bufio.NewWriter(f)
	defer w.Flush()
	_, err = io.Copy(w, r)
	if err != nil {
		return err
	}
	return nil

}

func setupReadAll(stdoutPipe, stderrPipe io.Reader, buff *bytes.Buffer) error {
	stdOutDoneChan := make(chan error, 1)
	stdErrDoneChan := make(chan error, 1)

	// also, Tee everything to this processes' stdout/stderr
	cmdStderrTee := io.TeeReader(stderrPipe, os.Stderr)
	cmdStdoutTee := io.TeeReader(stdoutPipe, os.Stdout)

	// spawn goroutines to read from stdout/stderr
	go func() {
		if err := streamToFile(cmdStdoutTee, "stdout.log"); err != nil {
			stdOutDoneChan <- err
		} else {
			stdOutDoneChan <- nil
		}

	}()

	go func() {
		if err := streamToFile(cmdStderrTee, "stderr.log"); err != nil {
			stdErrDoneChan <- err
		} else {
			stdErrDoneChan <- nil
		}

	}()

	// wait for goroutines
	stdOutResult := <-stdOutDoneChan
	stdErrResult := <-stdErrDoneChan

	// check for errors
	results := []error{stdOutResult, stdErrResult}
	for _, result := range results {
		if result != nil {
			return fmt.Errorf("Saving cmd output failed: %v", result)
		}
	}

	return nil
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

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		output.ExitCode = -1
		output.ExitDescription = fmt.Sprintf("Failed to bind stdout pipe: %v", err.Error())
		return output, err
	}

	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		output.ExitCode = -1
		output.ExitDescription = fmt.Sprintf("Failed to bind stderr pipe: %v", err.Error())
		return output, err
	}

	err = cmd.Start()
	if err != nil {
		fmt.Printf(">>> %v\n", err.Error())
	} else {

		var outputBuffer *bytes.Buffer
		err = setupReadAll(stdoutPipe, stderrPipe, outputBuffer)
		if err != nil {
			output.ExitCode = -1
			output.ExitDescription = fmt.Sprintf("Failed to setup read channels: %v", err.Error())
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
				fmt.Println(string(outputBuffer.String()))
				output.ExitDescription = "Execution finished with no error"
			}
		}
	}

	output.EndTime = time.Now()
	return output, nil
}
