package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/AstromechZA/gaze/conf"
)

func RunCmdBehaviour(report *GazeReport, config *conf.GazeBehaviourConfig) error {
	data, _ := json.Marshal(report)
	commandToRun := config.Settings["command"].(string)
	argsForCommandRaw := config.Settings["args"].([]interface{})
	argsForCommand := make([]string, len(argsForCommandRaw))
	for i, r := range argsForCommandRaw {
		argsForCommand[i] = r.(string)
	}
	cmd := exec.Command(commandToRun, argsForCommand...)
	cmd.Stdin = bytes.NewReader(data)
	err := cmd.Start()
	if err != nil {
		return err
	}
	return cmd.Wait()
}

func RunWebBehaviour(report *GazeReport, config *conf.GazeBehaviourConfig) error {
	// convert to json
	data, _ := json.Marshal(report)

	// pull stuff out of settings
	url := config.Settings["url"].(string)
	method := config.Settings["method"].(string)
	extraHeaders := config.Settings["headers"].(map[string]string)

	// construct the thing
	log.Infof("Making %v request to %v..", method, url)
	req, err := http.NewRequest(method, url, bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	for headerName, headerContent := range extraHeaders {
		req.Header.Set(headerName, headerContent)
	}

	// do the thing
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// must read body but pretty much discard it
	body, _ := ioutil.ReadAll(resp.Body)

	// throw error if necessary
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		log.Infof("Request returned code %v: %v", resp.Status, body)
		return fmt.Errorf("%v request to %v failed with code %v", method, url, resp.StatusCode)
	}

	return nil
}

func checkLogDirectoryExists(directoryPath string) error {
	dstat, err := os.Stat(directoryPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("Directory '%v' does not exist!", directoryPath)
		}
		return fmt.Errorf("Could not read directory '%v': %v", directoryPath, err.Error())
	}
	if !dstat.IsDir() {
		return fmt.Errorf("Path '%v' is a file, we cannot create a logfile", directoryPath)
	}
	return nil
}

func ensureLogFileExists(fullFilePath string) error {
	_, err := os.Stat(fullFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			f, err := os.Create(fullFilePath)
			if err != nil {
				return err
			}
			defer f.Close()
		}
		return fmt.Errorf("Could not check file '%v': %v", fullFilePath, err.Error())
	}
	return nil
}

func RunLogBehaviour(report *GazeReport, config *conf.GazeBehaviourConfig) error {

	// pull stuff from settings
	logDirectory := config.Settings["directory"].(string)
	logFileName := config.Settings["filename"].(string)
	logFormat := config.Settings["format"].(string)

	if err := checkLogDirectoryExists(logDirectory); err != nil {
		return err
	}
	logFilePath := filepath.Join(logDirectory, logFileName)
	if err := ensureLogFileExists(logFilePath); err != nil {
		return err
	}
	log.Infof("Appending data to '%v'..", logFilePath)
	f, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_RDWR, 06666)
	if err != nil {
		return err
	}
	defer f.Close()

	if logFormat == "machine" {
		data, _ := json.Marshal(report)
		f.Write(data)
		f.WriteString("\n")
	} else {
		f.WriteString("---- ---- ---- ----\n")
		f.WriteString(fmt.Sprintf("Start Time:       %v\n", report.StartTime))
		f.WriteString(fmt.Sprintf("End Time:         %v\n", report.EndTime))
		f.WriteString(fmt.Sprintf("Exit Code:        %v\n", report.ExitCode))
		f.WriteString(fmt.Sprintf("Exit Description: %v\n", report.ExitDescription))
		f.WriteString(fmt.Sprintf("Captured Output:  %v\n", report.CapturedOutput))
	}

	return nil
}
