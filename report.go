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

func streamToBuffer(r io.Reader, buff *bytes.Buffer) error {
	w := bufio.NewWriter(buff)
	defer w.Flush()
	_, err := io.Copy(w, r)
	return err
}

func setupPullFromStdout(stdoutPipe io.ReadCloser, target *bytes.Buffer) error {
	stdOutDoneChan := make(chan error, 1)

	// also, Tee everything to this processes' stdout/stderr
	cmdStdoutTee := io.TeeReader(stdoutPipe, os.Stdout)

	// spawn goroutines to read from stdout/stderr
	go func() {
		stdOutDoneChan <- streamToBuffer(cmdStdoutTee, target)
	}()

	// wait for goroutines
	stdOutResult := <-stdOutDoneChan
	return stdOutResult
}

func setupPullFromStderr(stderrPipe io.ReadCloser, target *bytes.Buffer) error {
	stdErrDoneChan := make(chan error, 1)

	// also, Tee everything to this processes' stdout/stderr
	cmdStderrTee := io.TeeReader(stderrPipe, os.Stderr)

	// spawn goroutines to read from stdout/stderr
	go func() {
		stdErrDoneChan <- streamToBuffer(cmdStderrTee, target)
	}()

	// wait for goroutines
	stdErrResult := <-stdErrDoneChan
	return stdErrResult
}

func setupReadAll(stdoutPipe, stderrPipe io.ReadCloser, buff *bytes.Buffer) error {
	// wait for goroutines
	stdOutResult := setupPullFromStdout(stdoutPipe, buff)
	stdErrResult := setupPullFromStderr(stderrPipe, buff)

	// check for errors
	if stdOutResult != nil {
		return stdOutResult
	}
	return stdErrResult
}

func runReport(args []string, config *conf.GazeConfig, name string) (*GazeReport, error) {
	output := new(GazeReport)
	output.StartTime = time.Now()
	output.ExitCode = 0
	output.ExitDescription = "No description added"
	output.CapturedOutput = ""

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

		outputBuffer := new(bytes.Buffer)
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
				output.ExitDescription = "Execution finished with no error"
			}
			output.CapturedOutput = outputBuffer.String()
		}
	}

	output.EndTime = time.Now()
	return output, nil
}
