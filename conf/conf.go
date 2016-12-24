package conf

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	logging "github.com/op/go-logging"
)

var log = logging.MustGetLogger("gaze.conf")

type GazeBehaviourConfig struct {
	Type          string                 `toml:"type"`
	When          string                 `toml:"when"`
	IncludeOutput bool                   `toml:"include_output"`
	Settings      map[string]interface{} `toml:"settings"`
}

type GazeConfig struct {
	Behaviours map[string]*GazeBehaviourConfig `toml:"behaviours"`
	Tags       []string                        `toml:"tags"`
}

// Load the config information from the file on disk
func Load(path *string, mustExist bool) (*GazeConfig, error) {
	var output GazeConfig

	configPath, err := filepath.Abs(*path)
	if err != nil {
		return nil, fmt.Errorf("Failed to construct config path: %v", err.Error())
	}

	_, err = toml.DecodeFile(configPath, &output)
	if err != nil {
		if os.IsNotExist(err) && !mustExist {
			log.Warningf("Config file %v does not exist, but we don't require it so using an empty config struct anyway!", configPath)
			return &output, nil
		}
		return nil, err
	}

	return &output, nil
}

func GenerateExample() *GazeConfig {
	var output GazeConfig
	output.Tags = []string{"tagA", "tagB"}
	output.Behaviours = make(map[string]*GazeBehaviourConfig)

	b1 := new(GazeBehaviourConfig)
	b1.Type = "logfile"
	b1.When = "failures"
	b1.IncludeOutput = false
	b1.Settings = make(map[string]interface{})
	b1.Settings["directory"] = "/var/log"
	b1.Settings["filename"] = "gaze.log"
	b1.Settings["format"] = "human"
	output.Behaviours["logging"] = b1

	b2 := new(GazeBehaviourConfig)
	b2.Type = "web"
	b2.When = "always"
	b2.IncludeOutput = true
	b2.Settings = make(map[string]interface{})
	b2.Settings["url"] = "http://127.0.0.1:8080"
	b2.Settings["method"] = "POST"
	h := make(map[string]string)
	h["API-TOKEN"] = "MY_TOKEN"
	b2.Settings["headers"] = h
	output.Behaviours["request"] = b2

	b3 := new(GazeBehaviourConfig)
	b3.Type = "command"
	b3.When = "successes"
	b3.IncludeOutput = true
	b3.Settings = make(map[string]interface{})
	b3.Settings["command"] = "python"
	b3.Settings["args"] = []string{"-m", "json.tool"}
	output.Behaviours["cmd"] = b3

	return &output
}

func stringIn(containee string, container *[]string) bool {
	for _, s := range *container {
		if s == containee {
			return true
		}
	}
	return false
}

func validateStringSetting(input *GazeBehaviourConfig, name string) error {
	commandV, ok := input.Settings[name]
	if ok {
		commandS, ok := commandV.(string)
		if ok {
			input.Settings[name] = commandS
			return nil
		}
	}
	return fmt.Errorf("Behaviour of type '%v' must have a '%v' string", input.Type, name)
}

func validateStringSettingWithDefault(input *GazeBehaviourConfig, name string, defaultValue string) error {
	commandV, ok := input.Settings[name]
	if ok {
		commandS, ok := commandV.(string)
		if ok {
			input.Settings[name] = commandS
			return nil
		}
		return fmt.Errorf("Behaviour of type '%v' must have a '%v' string", input.Type, name)
	}
	input.Settings[name] = defaultValue
	return nil
}

func validateStringSettingAllowed(input *GazeBehaviourConfig, name string, allowed *[]string) error {
	if err := validateStringSetting(input, name); err != nil {
		return err
	}
	if !stringIn(input.Settings[name].(string), allowed) {
		return fmt.Errorf("Behaviour of type '%v' setting '%v' must be one of '%v'", input.Type, name, &allowed)
	}
	return nil
}

func validateStringSettingWithDefaultAllowed(input *GazeBehaviourConfig, name string, defaultValue string, allowed *[]string) error {
	if err := validateStringSettingWithDefault(input, name, defaultValue); err != nil {
		return err
	}
	if !stringIn(input.Settings[name].(string), allowed) {
		return fmt.Errorf("Behaviour of type '%v' setting '%v' must be one of '%v'", input.Type, name, &allowed)
	}
	return nil
}

func ValidateGazeCommandBehaviour(input *GazeBehaviourConfig) error {
	if err := validateStringSetting(input, "command"); err != nil {
		return err
	}
	argsV, ok := input.Settings["args"]
	if ok {
		_, ok := argsV.([]interface{})
		if !ok {
			return fmt.Errorf("Behaviour of type '%v' setting 'args' must be an array", input.Type)
		}
		argsVA := argsV.([]interface{})
		for _, i := range argsVA {
			_, ok = i.(string)
			if !ok {
				return fmt.Errorf("Behaviour of type '%v' setting 'args' must contain only strings", input.Type)
			}
		}
	} else {
		return fmt.Errorf("Behaviour of type '%v' must have an 'args' key", input.Type)
	}
	return nil
}

func ValidateGazeLogFileBehaviour(input *GazeBehaviourConfig) error {
	if err := validateStringSetting(input, "directory"); err != nil {
		return err
	}
	if err := validateStringSetting(input, "filename"); err != nil {
		return err
	}
	validFormats := []string{"human", "machine"}
	if err := validateStringSettingAllowed(input, "format", &validFormats); err != nil {
		return err
	}
	return nil
}

func ValidateGazeWebBehaviour(input *GazeBehaviourConfig) error {
	if err := validateStringSetting(input, "url"); err != nil {
		return err
	}
	validMethods := []string{"POST", "PUT"}
	if err := validateStringSettingWithDefaultAllowed(input, "method", "POST", &validMethods); err != nil {
		return err
	}
	headersV, ok := input.Settings["headers"]
	if ok {
		_, ok := headersV.(map[string]interface{})
		if !ok {
			return fmt.Errorf("Behaviour of type '%v' headers must contain key-value pairs", input.Type)
		}
		headersVM := headersV.(map[string]interface{})
		for _, v := range headersVM {
			_, ok := v.(string)
			if !ok {
				return fmt.Errorf("Behaviour of type '%v' headers can only be string-string pairs", input.Type)
			}
		}
	} else {
		input.Settings["headers"] = make(map[string]string)
	}
	return nil
}

// ValidateAndClean a config that has already been loaded
func ValidateAndClean(cfg *GazeConfig) error {
	validTypes := []string{"logfile", "command", "web"}
	validWhens := []string{"always", "failures", "successes"}

	for _, behaviour := range cfg.Behaviours {
		if !stringIn(behaviour.Type, &validTypes) {
			return fmt.Errorf("Behaviour 'type' must be one of %v", validTypes)
		}
		if behaviour.When == "" {
			behaviour.When = "always"
		}
		if !stringIn(behaviour.When, &validWhens) {
			return fmt.Errorf("Behaviour 'when' must be one of %v", validWhens)
		}
		if behaviour.Type == "command" {
			if err := ValidateGazeCommandBehaviour(behaviour); err != nil {
				return err
			}
		}
		if behaviour.Type == "logfile" {
			if err := ValidateGazeLogFileBehaviour(behaviour); err != nil {
				return err
			}
		}
		if behaviour.Type == "web" {
			if err := ValidateGazeWebBehaviour(behaviour); err != nil {
				return err
			}
		}
	}

	return nil
}
