package main

import (
	"flag"
	"fmt"
	"os"
	"os/user"
	"path/filepath"

	"regexp"

	"encoding/json"

	"github.com/AstromechZA/gaze/conf"
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

func mainInner() error {

	// first set up config flag options
	versionFlag := flag.Bool("version", false, "Print the version string")
	configFlag := flag.String("config", "", "path to a gaze config file (default = $HOME/.config/gaze.json)")
	jsonFlag := flag.Bool("json", false, "mutes normal stdout and stderr and just outputs the json report on stdout")
	debugFlag := flag.Bool("debug", false, "mutes normal stdout and stderr and just outputs debug messages")
	nameFlag := flag.String("name", "", "override the auto generated name for the task")

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

	// json and debug conflict
	if *jsonFlag && *debugFlag {
		return fmt.Errorf("Cannot specify both -debug and -json")
	}

	// args must either be present or "-"
	if flag.NArg() == 0 {
		flag.Usage()
		return nil
	}

	// identify config path
	configPath := (*configFlag)
	if configPath == "" {
		usr, _ := user.Current()
		configPath = filepath.Join(usr.HomeDir, ".config/gaze.toml")
	}
	configPath, err := filepath.Abs(configPath)
	if err != nil {
		return fmt.Errorf("Failed to identify config path: %v", err.Error())
	}
	// load and validate config
	cfg, err := conf.Load(&configPath)
	if err != nil {
		return fmt.Errorf("Failed to load config: %v", err.Error())
	}
	if err = conf.ValidateAndClean(cfg); err != nil {
		return fmt.Errorf("Config failed validation: %v", err.Error())
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
	// TODO possibly more validation for this name
	if commandName == "" {
		return fmt.Errorf("Could not build command name from supplied args, please provide -name flag for gaze")
	}

	forwardOutputToConsole := !*jsonFlag

	// run and generate report
	report, err := runReport(flag.Args(), cfg, commandName, forwardOutputToConsole)
	if err != nil {
		return fmt.Errorf("Failed during run and report: %v", err.Error())
	}

	activateBehaviours := true

	if *jsonFlag {
		output, _ := json.Marshal(report)
		fmt.Println(string(output))
		return nil
	}

	commandWasSuccessful := report.ExitCode == 0
	if activateBehaviours {
		for _, bref := range cfg.Behaviours {
			// only run at the right times
			if commandWasSuccessful && bref.When == "failures" {
				continue
			} else if !commandWasSuccessful && bref.When == "successes" {
				continue
			}

			// run the correct behaviour
			if bref.Type == "command" {
				_ = RunCmdBehaviour(report, bref)
			} else if bref.Type == "logfile" {
				_ = RunLogBehaviour(report, bref)
			} else if bref.Type == "web" {
				_ = RunWebBehaviour(report, bref)
			} else {
				fmt.Printf(">>> err: unknown behaviour type: %v", bref.Type)
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
