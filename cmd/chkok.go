package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/farzadghanei/chkok"
)

func main() {
	confPath := flag.String("conf", "/etc/chkok.yaml", "path to configuration file")
	verbose := flag.Bool("verbose", false, "more output, include logs")
	flag.Parse()

	os.Exit(run(confPath, os.Stderr, verbose))
}

// run app provided with main arguments, print results to output, return exit code
func run(confPath *string, output io.Writer, verbose *bool) int {
	logger := log.New(io.Discard, "", log.Lshortfile)
	if *verbose {
		logger.SetOutput(output)
	}

	conf, err := chkok.ReadConf(*confPath)
	if err != nil {
		fmt.Fprintf(output, "couldn't read YAML configuration file: %v", err)
		return chkok.ExDataErr
	}
	checkGroups, err := chkok.CheckSuitesFromSpecSuites(conf.CheckSuites)
	if err != nil {
		fmt.Fprintf(output, "invalid configurations: %v", err)
		return chkok.ExConfig
	}
	return runCli(&checkGroups, conf, output, logger)
}

// run app in CLI mode, return exit code
func runCli(checkGroups *chkok.CheckSuites, conf *chkok.Conf, output io.Writer, logger *log.Logger) int {
	runner := chkok.Runner{Log: logger}
	logger.Printf("running checks ...")
	checks := runner.RunChecks(*checkGroups, conf.Runners["default"].Timeout)
	incompeleteChecks := 0
	failed := 0
	for _, chk := range checks {
		if chk.Status() != chkok.StatusDone {
			incompeleteChecks++
		}
		if !chk.Result().IsOK {
			failed++
		}
		fmt.Fprintf(output, "check %s status %d ok: %v\n", chk.Name(), chk.Status(), chk.Result().IsOK)
	}
	if incompeleteChecks > 0 {
		fmt.Fprintf(output, "%v checks didn't get to completion", incompeleteChecks)
		return chkok.ExTempFail
	}
	if failed > 0 {
		return chkok.ExSoftware
	}
	return chkok.ExOK
}
