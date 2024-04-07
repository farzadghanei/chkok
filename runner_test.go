package chkok

import (
	"io"
	"log"
	"testing"
	"time"
)

func TestTimedoutRunnerWontSubmitChecks(t *testing.T) {
	logger := log.New(io.Discard, "", log.Lshortfile)
	timeout, _ := time.ParseDuration("0s")
	checks := make(CheckSuites)
	checks["default"] = []Check{NewCheckFile("examples"), NewCheckDial()}
	runner := Runner{Log: logger, Timeout: timeout}
	results := runner.RunChecks(checks)
	checkDial := results[1]
	if checkDial.Status() != StatusUnknown {
		t.Errorf("wanted 2nd check not run due to timedout")
	}
}

func TestTimedoutRunnerAdjustsTimeouts(t *testing.T) {
	logger := log.New(io.Discard, "", log.Lshortfile)
	timeout, _ := time.ParseDuration("1s")
	duration10, _ := time.ParseDuration("10s")
	checks := make(CheckSuites)
	checkDial := NewCheckDial()
	checkDial.SetTimeout(duration10)
	checks["default"] = []Check{checkDial}
	runner := Runner{Log: logger, Timeout: timeout}
	runner.RunChecks(checks)
	if checkDial.GetTimeout() > timeout {
		t.Errorf("wanted check dial's timeout adjusted to 1s, got %v", checkDial.GetTimeout())
	}
}
