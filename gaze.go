package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/AstromechZA/gaze/conf"

	"github.com/go-yaml/yaml"
	logging "github.com/op/go-logging"
)

const usageString = `TODO

`

const logoImage = `
 .    '                   .  "   '
            .  .  .                 '      '
    "'       .   .
                                     '     '
  .    '      _______________
          ==c(___(o(______(_()
                  \=\
                   )=\
                  //|\\
                 //|| \\
                // ||  \\
               //  ||   \\
              //         \\
`

// GazeVersion is the version string
// format should be 'X.YZ'
// Set this at build time using the -ldflags="-X main.GazeVersion=X.YZ"
var GazeVersion = "<unofficial build>"

var logFormat = logging.MustStringFormatter(
	"%{time:2006-01-02T15:04:05.000} %{module} %{level:.4s} - %{message}",
)
var log = logging.MustGetLogger("gaze")

// MuteLogBackend is used because for some reason this logging library doesn't support globally
// disabling the logging
type MuteLogBackend struct{}

// Log function does nothing
func (b *MuteLogBackend) Log(level logging.Level, calldepth int, rec *logging.Record) error {
	return nil
}

func setupLogging(enabled bool) {
	if enabled {
		logBackend := logging.NewLogBackend(os.Stdout, "", 0)
		backend := logging.NewBackendFormatter(logBackend, logFormat)
		lvlBackend := logging.AddModuleLevel(backend)
		lvlBackend.SetLevel(logging.DEBUG, "")
		logging.SetBackend(lvlBackend)
	} else {
		logging.SetBackend(new(MuteLogBackend))
	}
}

func mainInner() error {

	// first set up config flag options
	versionFlag := flag.Bool("version", false, "Print the version string")
	configFlag := flag.String("config", "", "path to a gaze config file (default = $HOME/.config/gaze.yaml)")
	jsonFlag := flag.Bool("json", false, "mutes normal stdout and stderr and just outputs the json report on stdout")
	debugFlag := flag.Bool("debug", false, "mutes normal stdout and stderr and just outputs debug messages")
	nameFlag := flag.String("name", "", "override the auto generated name for the task")
	tagsFlag := flag.String("extra-tags", "", "comma-seperated extra tags to add to the structure")
	exampleConfigFlag := flag.Bool("example-config", false, "output an example config and exit")

	// set a more verbose usage message.
	flag.Usage = func() {
		os.Stderr.WriteString(usageString)
		flag.PrintDefaults()
	}
	// parse them
	flag.Parse()

	// first do arg checking
	if *versionFlag {
		fmt.Println("Version: " + GazeVersion)
		fmt.Println(logoImage)
		fmt.Println("Project: https://github.com/AstromechZA/gaze")
		return nil
	}

	// example config
	if *exampleConfigFlag {
		cfg := conf.GenerateExample()
		cfgBytes, err := yaml.Marshal(cfg)
		if err != nil {
			return err
		}
		fmt.Println(string(cfgBytes))
		return nil
	}

	// json and debug conflict
	if *jsonFlag && *debugFlag {
		return fmt.Errorf("Cannot specify both -debug and -json")
	}

	// args must either be present or "-"
	if flag.NArg() == 0 {
		flag.Usage()
		return nil
	}

	setupLogging(*debugFlag)
	log.Info("Logging initialised.")

	var cfg *conf.GazeConfig
	// identify config path
	configPath := (*configFlag)
	configMustExist := true
	if configPath == "" {
		configMustExist = false
		usr, _ := user.Current()
		configPath = filepath.Join(usr.HomeDir, ".config/gaze.yaml")
	}

	// load and validate config
	log.Infof("Loading config from %v", configPath)
	cfg, err := conf.Load(&configPath, configMustExist)
	if err != nil {
		return fmt.Errorf("Failed to load config: %v", err.Error())
	}
	if err = conf.ValidateAndClean(cfg); err != nil {
		return fmt.Errorf("Config failed validation: %v", err.Error())
	}

	j, err := json.MarshalIndent(cfg, "", "  ")
	log.Infof("Loaded config: %v (err: %v)", string(j), err)

	// append extra tags to the config object even though it might be nil
	if *tagsFlag != "" {
		extraTagsRaw := strings.Split(*tagsFlag, ",")
		for _, t := range extraTagsRaw {
			t = strings.TrimSpace(t)
			if len(t) > 0 {
				cfg.Tags = append(cfg.Tags, t)
			}
		}
	}

	// build command name
	var commandName string
	if *nameFlag != "" {
		commandName = *nameFlag
	} else {
		// build command name
		commandName = ""
		re := regexp.MustCompile("^\\w[\\w\\-\\.]*$")
		for _, a := range flag.Args() {
			if re.Match([]byte(a)) {
				if len(commandName) > 0 {
					commandName += "."
				}
				commandName += a
			}
		}
	}
	log.Infof("Attempting to use '%v' as commandName", commandName)
	if commandName == "" {
		return fmt.Errorf("Could not build command name from supplied args, please provide -name flag for gaze")
	}

	// run and generate report
	forwardOutputToConsole := !*jsonFlag
	report, err := runReport(flag.Args(), cfg, commandName, forwardOutputToConsole)
	if err != nil {
		return fmt.Errorf("Failed during run and report: %v", err.Error())
	}
	log.Infof("Command exited with code %v", report.ExitCode)

	if *jsonFlag {
		output, _ := json.Marshal(report)
		fmt.Println(string(output))
		return nil
	}

	commandWasSuccessful := report.ExitCode == 0
	activateBehaviours := true
	if activateBehaviours {
		for _, bref := range cfg.Behaviours {
			log.Infof("Running behaviour of type %v..", bref.Type)

			// only run at the right times
			if commandWasSuccessful && bref.When == "failures" {
				log.Info("Skipping because it only runs on failures")
				continue
			} else if !commandWasSuccessful && bref.When == "successes" {
				log.Info("Skipping because it only runs on successes")
				continue
			}

			// run the correct behaviour
			if bref.Type == "command" {
				err = RunCmdBehaviour(report, bref)
			} else if bref.Type == "logfile" {
				err = RunLogBehaviour(report, bref)
			} else if bref.Type == "web" {
				err = RunWebBehaviour(report, bref)
			} else {
				panic(fmt.Sprintf(">>> err: unknown behaviour type: %v", bref.Type))
			}
			if err == nil {
				log.Info("Behaviour completed.")
			} else {
				log.Errorf("Behaviour '%v' failed!: %v", bref.Type, err.Error())
			}
		}
	}

	return nil
}

func main() {
	if err := mainInner(); err != nil {
		os.Stderr.WriteString(err.Error() + "\n")
		os.Exit(1)
	}
}
