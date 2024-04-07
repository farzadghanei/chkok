package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/farzadghanei/chkok"
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
		return runHTTP(&checkGroups, conf, output, logger)
	}
	return runCli(&checkGroups, conf, output, logger)
}

// run app in CLI mode, return exit code
func runCli(checkGroups *chkok.CheckSuites, conf *chkok.Conf, output io.Writer, logger *log.Logger) int {
	runner := chkok.Runner{Log: logger, Timeout: conf.Runners["default"].Timeout}
	logger.Printf("running checks ...")
	checks := runner.RunChecks(*checkGroups)
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

// run app in http server mode, return exit code
func runHTTP(checkGroups *chkok.CheckSuites, conf *chkok.Conf, output io.Writer, logger *log.Logger) int {
	runner := chkok.Runner{Log: logger, Timeout: conf.Runners["default"].Timeout}
	logger.Printf("running checks ...")

	httpHandler := func(w http.ResponseWriter, r *http.Request) {
		logger.Printf("http request: %v", r)
		checks := runner.RunChecks(*checkGroups)
		incompeleteChecks := 0
		failed := 0
		passed := 0
		for _, chk := range checks {
			if chk.Status() != chkok.StatusDone {
				incompeleteChecks++
			} else {
				if chk.Result().IsOK {
					passed++
				} else {
					failed++
				}
			}
			logger.Printf("check %s status %d ok: %v", chk.Name(), chk.Status(), chk.Result().IsOK)
		}
		logger.Printf("Run checks done. passed: %v - failed: %v - timedout: %v", passed, failed, incompeleteChecks)
		if incompeleteChecks > 0 {
			w.WriteHeader(http.StatusGatewayTimeout)  // 504
			fmt.Fprintf(w, "TIMEDOUT")
		} else if failed > 0 {
			w.WriteHeader(http.StatusInternalServerError)  // 500
			fmt.Fprintf(w, "FAILED")
		} else {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, "OK")
		}
	}

	http.HandleFunc("/", httpHandler)
	logger.Printf("starting http server ...")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		logger.Printf("http server failed to start: %v", err)
		return chkok.ExSoftware
	}
	return chkok.ExOK
}
