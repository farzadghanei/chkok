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

	os.Exit(run(confPath, verbose))
}

// run app provided with main arguments, return exit code
func run(confPath *string, verbose *bool) int {
	logger := log.New(io.Discard, "", log.Lshortfile)
	if *verbose {
		logger.SetOutput(os.Stderr)
	}

	conf, err := chkok.ReadConf(*confPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "couldn't read YAML configuration file: %v", err)
		return chkok.ExDataErr
	}
	checkGroups, err := chkok.CheckSuitesFromSpecSuites(conf.CheckSuites)
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid configurations: %v", err)
		return chkok.ExConfig
	}
	return runCli(&checkGroups, conf, logger)
}

// run app in CLI mode, return exit code
func runCli(checkGroups *chkok.CheckSuites, conf *chkok.Conf, logger *log.Logger) int {
	runner := chkok.Runner{Log: logger}
	logger.Printf("running checks ...")
	checks := runner.RunChecks(*checkGroups, conf.Runners["default"].Timeout)
	incompeleteChecks := 0
	for _, chk := range checks {
		if chk.Status() != chkok.StatusDone {
			incompeleteChecks++
		}
		fmt.Printf("check %s status %d ok: %v\n", chk.Name(), chk.Status(), chk.Result().IsOK)
	}
	if incompeleteChecks > 0 {
		fmt.Fprintf(os.Stderr, "%v checks didn't get to completion", incompeleteChecks)
		return chkok.ExTempFail
	}
	fmt.Fprintf(os.Stderr, "all checks are done!")
	return chkok.ExOK
}
