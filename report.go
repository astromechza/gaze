package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"syscall"
	"time"

	"github.com/AstromechZA/gaze/conf"
)

type GazeReport struct {
	Name    string   `json:"name"`
	Command []string `json:"command"`

	StartTime      time.Time `json:"start_time"`
	EndTime        time.Time `json:"end_time"`
	ElapsedSeconds float32   `json:"elapsed_seconds"`

	ExitCode        int    `json:"exit_code"`
	ExitDescription string `json:"exit_description"`

	CapturedOutput string `json:"captured_output"`

	Hostname string `json:"hostname"`
}

func streamToBuffer(r io.Reader, buff *bytes.Buffer) error {
	w := bufio.NewWriter(buff)
	defer w.Flush()
	_, err := io.Copy(w, r)
	return err
}

func beginBufferTee(inputPipe io.ReadCloser, target *bytes.Buffer, forwardOutput bool, forwardTarget *os.File) error {
	doneChan := make(chan error, 1)

	var bufferSource io.Reader = inputPipe
	if forwardOutput {
		// also, Tee everything to this processes' stdout/stderr
		bufferSource = io.TeeReader(bufferSource, forwardTarget)
	}

	// start goroutine to run copy routine
	go func() {
		doneChan <- streamToBuffer(bufferSource, target)
	}()

	// wait for goroutine
	return <-doneChan
}

func setupReadAll(stdoutPipe, stderrPipe io.ReadCloser, buff *bytes.Buffer, forwardOutput bool) error {
	// wait for goroutines
	stdOutResult := beginBufferTee(stdoutPipe, buff, forwardOutput, os.Stdout)
	stdErrResult := beginBufferTee(stderrPipe, buff, forwardOutput, os.Stderr)

	// check for errors
	if stdOutResult != nil {
		return stdOutResult
	}
	return stdErrResult
}

func runReport(args []string, config *conf.GazeConfig, name string, forwardOutput bool) (*GazeReport, error) {
	output := new(GazeReport)
	output.Name = name
	output.StartTime = time.Now()
	output.ExitCode = 0
	output.ExitDescription = "No description added"
	output.CapturedOutput = ""
	output.ElapsedSeconds = 0
	output.Command = args

	hn, err := os.Hostname()
	if err == nil {
		output.Hostname = hn
	}

	defer func() {
		output.EndTime = time.Now()
		output.ElapsedSeconds = float32(output.EndTime.Sub(output.StartTime)) / float32(time.Second)
	}()

	// run command
	cmd := exec.Command(args[0], args[1:]...)

	// send process stdin to subprocess
	cmd.Stdin = os.Stdin

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
