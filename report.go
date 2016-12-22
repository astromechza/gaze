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
	EndTime         time.Time `json:"end_time"`
	ExitCode        int       `json:"exit_code"`
	CapturedOutput  string    `json:"captured_output"`
	ExitDescription string    `json:"exit_description"`
}

func streamToBuffer(r io.Reader, buff *bytes.Buffer) error {
	w := bufio.NewWriter(buff)
	defer w.Flush()
	_, err := io.Copy(w, r)
	return err
}

func setupPullFromStdout(stdoutPipe io.ReadCloser, target *bytes.Buffer, forwardOutput bool) error {
	stdOutDoneChan := make(chan error, 1)

	if forwardOutput {
		// also, Tee everything to this processes' stdout/stderr
		cmdStdoutTee := io.TeeReader(stdoutPipe, os.Stdout)

		// spawn goroutines to read from stdout/stderr
		go func() {
			stdOutDoneChan <- streamToBuffer(cmdStdoutTee, target)
		}()
	} else {
		go func() {
			stdOutDoneChan <- streamToBuffer(stdoutPipe, target)
		}()
	}

	// wait for goroutines
	stdOutResult := <-stdOutDoneChan
	return stdOutResult
}

func setupPullFromStderr(stderrPipe io.ReadCloser, target *bytes.Buffer, forwardOutput bool) error {
	stdErrDoneChan := make(chan error, 1)

	if forwardOutput {
		// also, Tee everything to this processes' stdout/stderr
		cmdStderrTee := io.TeeReader(stderrPipe, os.Stderr)

		// spawn goroutines to read from stdout/stderr
		go func() {
			stdErrDoneChan <- streamToBuffer(cmdStderrTee, target)
		}()
	} else {
		go func() {
			stdErrDoneChan <- streamToBuffer(stderrPipe, target)
		}()
	}

	// wait for goroutines
	stdErrResult := <-stdErrDoneChan
	return stdErrResult
}

func setupReadAll(stdoutPipe, stderrPipe io.ReadCloser, buff *bytes.Buffer, forwardOutput bool) error {
	// wait for goroutines
	stdOutResult := setupPullFromStdout(stdoutPipe, buff, forwardOutput)
	stdErrResult := setupPullFromStderr(stderrPipe, buff, forwardOutput)

	// check for errors
	if stdOutResult != nil {
		return stdOutResult
	}
	return stdErrResult
}

func runReport(args []string, config *conf.GazeConfig, name string, forwardOutput bool) (*GazeReport, error) {
	output := new(GazeReport)
	output.StartTime = time.Now()
	output.ExitCode = 0
	output.ExitDescription = "No description added"
	output.CapturedOutput = ""

	defer func() {
		output.EndTime = time.Now()
	}()

	// run command
	cmd := exec.Command(args[0], args[1:]...)

	// send process stdin to subprocess
	cmd.Stdin = os.Stdin

	var err error
	var stdoutPipe io.ReadCloser
	var stderrPipe io.ReadCloser

	stdoutPipe, err = cmd.StdoutPipe()
	if err != nil {
		output.ExitCode = -1
		output.ExitDescription = fmt.Sprintf("Failed to bind stdout pipe: %v", err.Error())
		return output, err
	}

	stderrPipe, err = cmd.StderrPipe()
	if err != nil {
		output.ExitCode = -1
		output.ExitDescription = fmt.Sprintf("Failed to bind stderr pipe: %v", err.Error())
		return output, err
	}

	err = cmd.Start()
	if err != nil {
		output.ExitCode = 127
		output.ExitDescription = err.Error()
		return output, nil
	}

	outputBuffer := new(bytes.Buffer)

	err = setupReadAll(stdoutPipe, stderrPipe, outputBuffer, forwardOutput)
	if err != nil {
		output.ExitCode = -1
		output.ExitDescription = fmt.Sprintf("Failed to setup read channels: %v", err.Error())
		return output, err
	}

	err = cmd.Wait()
	output.CapturedOutput = outputBuffer.String()

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

	return output, nil
}
