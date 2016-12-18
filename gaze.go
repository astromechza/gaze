package main

import (
	"flag"
	"fmt"
	"os"
)

// mattermost bot client

const usageString = `
TODO

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

	return nil
}

func main() {
	if err := mainInner(); err != nil {
		os.Stderr.WriteString(err.Error() + "\n")
		os.Exit(1)
	}
}
