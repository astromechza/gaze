package conf

import "github.com/BurntSushi/toml"
import "fmt"

type GazeBehaviour struct {
	Type         string                 `toml:"type"`
	When         string                 `toml:"when"`
	StdoutPolicy string                 `toml:"stdoutpolicy"`
	StderrPolicy string                 `toml:"stderrpolicy"`
	Settings     map[string]interface{} `toml:"settings"`
}

type GazeConfig struct {
	Behaviours []*GazeBehaviour `toml:"behaviours"`
}

// Load the config information from the file on disk
func Load(path *string) (*GazeConfig, error) {
	var output GazeConfig

	_, err := toml.DecodeFile(*path, &output)
	if err != nil {
		return nil, err
	}

	return &output, nil
}

func stringIn(containee string, container *[]string) bool {
	for _, s := range *container {
		if s == containee {
			return true
		}
	}
	return false
}

func validateStringSetting(input *GazeBehaviour, name string) error {
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

func validateStringSettingWithDefault(input *GazeBehaviour, name string, defaultValue string) error {
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

func validateStringSettingAllowed(input *GazeBehaviour, name string, allowed *[]string) error {
	if err := validateStringSetting(input, name); err != nil {
		return err
	}
	if !stringIn(input.Settings[name].(string), allowed) {
		return fmt.Errorf("Behaviour of type '%v' setting '%v' must be one of '%v'", input.Type, name, &allowed)
	}
	return nil
}

func validateStringSettingWithDefaultAllowed(input *GazeBehaviour, name string, defaultValue string, allowed *[]string) error {
	if err := validateStringSettingWithDefault(input, name, defaultValue); err != nil {
		return err
	}
	if !stringIn(input.Settings[name].(string), allowed) {
		return fmt.Errorf("Behaviour of type '%v' setting '%v' must be one of '%v'", input.Type, name, &allowed)
	}
	return nil
}

func ValidateGazeCommandBehaviour(input *GazeBehaviour) error {
	return validateStringSetting(input, "command")
}

func ValidateGazeLogFileBehaviour(input *GazeBehaviour) error {
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

func ValidateGazeWebBehaviour(input *GazeBehaviour) error {
	if err := validateStringSetting(input, "url"); err != nil {
		return err
	}
	validMethods := []string{"POST", "PUT"}
	if err := validateStringSettingWithDefaultAllowed(input, "method", "POST", &validMethods); err != nil {
		return err
	}
	commandV, ok := input.Settings["headers"]
	if ok {
		_, ok := commandV.(map[string]string)
		if !ok {
			return fmt.Errorf("Behaviour of type '%v' headers must be string-string key-values", input.Type)
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
	validPolicies := []string{"capture", "ignore"}

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
		if behaviour.StdoutPolicy == "" {
			behaviour.StdoutPolicy = "ignore"
		}
		if !stringIn(behaviour.StdoutPolicy, &validPolicies) {
			return fmt.Errorf("Behaviour 'stdoutpolicy' must be one of %v", validPolicies)
		}
		if behaviour.StderrPolicy == "" {
			behaviour.StderrPolicy = "ignore"
		}
		if !stringIn(behaviour.StderrPolicy, &validPolicies) {
			return fmt.Errorf("Behaviour 'stderrpolicy' must be one of %v", validPolicies)
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
