package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"

	"encoding/json"

	"github.com/AstromechZA/gaze/conf"
)

func RunCmdBehaviour(report *GazeReport, config *conf.GazeBehaviourConfig) error {
	// TODO actually run the thing
	fmt.Println("<mock run cmd behaviour>")
	return nil
}

func RunWebBehaviour(report *GazeReport, config *conf.GazeBehaviourConfig) error {
	// convert to json
	data, _ := json.Marshal(report)

	// pull stuff out of settings
	url := config.Settings["url"].(string)
	method := config.Settings["method"].(string)
	extraHeaders := config.Settings["headers"].(map[string]string)

	req, err := http.NewRequest(method, url, bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	for headerName, headerContent := range extraHeaders {
		req.Header.Set(headerName, headerContent)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// must read body
	ioutil.ReadAll(resp.Body)

	// throw error if necessary
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("request failed with code %v", resp.StatusCode)
	}

	return nil
}

func RunLogBehaviour(report *GazeReport, config *conf.GazeBehaviourConfig) error {
	// TODO actually run the thing
	fmt.Println("<mock run log behaviour>")
	return nil
}
