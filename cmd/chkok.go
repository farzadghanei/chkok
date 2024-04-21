package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	chkok "github.com/farzadghanei/chkok/internal"
)

// ModeHTTP run checks in http server mode
const ModeHTTP = "http"

func main() {
	var confPath string
	var mode string
	var verbose bool
	flag.StringVar(&confPath, "conf", "/etc/chkok.yaml", "path to configuration file")
	flag.StringVar(&mode, "mode", "cli", "running mode: cli,http")
	flag.BoolVar(&verbose, "verbose", false, "more output, include logs")
	flag.Parse()

	os.Exit(run(confPath, mode, os.Stderr, verbose))
}

// run app provided with main arguments, print results to output, return exit code
func run(confPath, mode string, output io.Writer, verbose bool) int {
	logger := log.New(io.Discard, "", log.Lshortfile)
	if verbose {
		logger.SetOutput(output)
	}

	conf, err := chkok.ReadConf(confPath)
	if err != nil {
		fmt.Fprintf(output, "couldn't read YAML configuration file: %v", err)
		return chkok.ExDataErr
	}
	checkGroups, err := chkok.CheckSuitesFromSpecSuites(conf.CheckSuites)
	if err != nil {
		fmt.Fprintf(output, "invalid configurations: %v", err)
		return chkok.ExConfig
	}
	if mode == ModeHTTP {
		return chkok.RunModeHTTP(&checkGroups, conf, logger)
	}
	return chkok.RunModeCLI(&checkGroups, conf, output, logger)
}
