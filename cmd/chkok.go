package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/farzadghanei/chkok"
)

func main() {
	confPath := flag.String("conf", "/etc/chkok.yaml", "path to configuration file")
	flag.Parse()

	conf, err := chkok.ReadConf(*confPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "couldn't read YAML configuration file: %v", err)
		os.Exit(chkok.EX_DATAERR)
	}
	checkGroups, err := chkok.CheckSuitesFromSpecSuites(conf.CheckSuites)
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid configurations: %v", err)
		os.Exit(chkok.EX_CONFIG)
	}
	checks := chkok.RunChecks(checkGroups, int(conf.Runners["default"].MaxRunning), conf.Runners["default"].Timeout)
	incompeleteChecks := 0
	for _, chk := range checks {
		if chk.Status() != chkok.StatusDone {
			incompeleteChecks += 1
		}
		fmt.Printf("check %s status %d result: %v\n", chk.Name(), chk.Status(), chk.Result().IsOK)
	}
	if incompeleteChecks > 0 {
		fmt.Fprintf(os.Stderr, "%v checks didn't get to completion", incompeleteChecks)
		os.Exit(chkok.EX_TEMPFAIL)
	}
	fmt.Fprintf(os.Stderr, "all checks are done!")
	os.Exit(chkok.EX_OK)
}

