package chkok

import (
	"testing"
	"time"
)

func TestReadConfErrors(t *testing.T) {
	var conf *Conf
	var err error
	_, err = ReadConf("/no/such/path/exists")
	if err == nil {
		t.Errorf("invalid read conf, want err got nil")
	}
	conf, err = ReadConf("LICENSE")
	if err == nil {
		t.Errorf("invalid read conf, want err got nil")
	}
	_, ok := conf.Runners["default"]
	if ok {
		t.Errorf("read conf invaid file returned default runner, should not")
	}
}

func TestReadConf(t *testing.T) {
	var conf *Conf
	var err error
	conf, err = ReadConf("../examples/config.yaml")
	if err != nil {
		t.Errorf("invalid read conf, want nil err got err %v", err)
	}
	runner, ok := conf.Runners["default"]
	if !ok {
		t.Errorf("invalid read conf, want default runner")
	}
	wantMinutes := 5
	if runner.Timeout.Minutes() != float64(wantMinutes) {
		t.Errorf("invalid read conf default runner, want %v timeout got %v", wantMinutes, runner.Timeout.Minutes())
	}
	etcChecks, ok := conf.CheckSuites["etc"]
	if !ok {
		t.Errorf("read conf found no etc checks")
	}
	if len(etcChecks) != 3 {
		t.Errorf("read conf etc checks failed, want 3 checks found %v", len(etcChecks))
	}
	defaultChecks, ok := conf.CheckSuites["default"]
	if !ok {
		t.Errorf("read conf found no default checks")
	}
	if len(defaultChecks) != 3 {
		t.Errorf("read conf default checks failed, want 3 checks found %v", len(etcChecks))
	}
	invalidChecks, ok := conf.CheckSuites["invalid"]
	if ok {
		t.Errorf("read conf found invalid check group but shouldn't")
	}
	if len(invalidChecks) > 0 {
		t.Errorf("read conf found invalid checks but shouldn't")
	}
}

func TestGetDefaultConfRunner(t *testing.T) {
	// Test case where default key exists
	wantReadTimeout := 60 * time.Second
	wantWriteTimeout := 30 * time.Second
	wantTimeout := 10 * time.Second

	respYes := "YES"
	respNo := "NO"
	respMaybe := "MAYBE"

	runners := ConfRunners{
		"default": ConfRunner{
			Timeout:              &wantTimeout,
			ShutdownSignalHeader: nil,
			ListenAddress:        "localhost:8081",
			RequestReadTimeout:   &wantReadTimeout,
			ResponseWriteTimeout: &wantWriteTimeout,
			ResponseOK:           &respYes,
			ResponseFailed:       &respNo,
			ResponseTimeout:      &respMaybe,
		},
	}

	defaultRunner := GetDefaultConfRunner(&runners)
	if *defaultRunner.Timeout != wantTimeout {
		t.Errorf("Expected Timeout to be %v, got %v", wantTimeout, defaultRunner.Timeout)
	}
	if defaultRunner.ListenAddress != "localhost:8081" {
		t.Errorf("Expected ListenAddress to be localhost:8081, got %s", defaultRunner.ListenAddress)
	}
	if *defaultRunner.RequestReadTimeout != wantReadTimeout {
		t.Errorf("Expected RequestReadTimeout to be %v, got %v", wantReadTimeout,
			defaultRunner.RequestReadTimeout)
	}
	if *defaultRunner.ResponseWriteTimeout != wantWriteTimeout {
		t.Errorf("Expected ResponseWriteTimeout to be %v, got %v", wantWriteTimeout,
			defaultRunner.ResponseWriteTimeout)
	}
	if *defaultRunner.ResponseOK != "YES" {
		t.Errorf("Expected ResponseOK to be YES, got %s", *defaultRunner.ResponseOK)
	}
	if *defaultRunner.ResponseFailed != "NO" {
		t.Errorf("Expected ResponseFailed to be NO, got %s", *defaultRunner.ResponseFailed)
	}
	if *defaultRunner.ResponseTimeout != "MAYBE" {
		t.Errorf("Expected ResponseTimeout to be MAYBE, got %s", *defaultRunner.ResponseTimeout)
	}

	// Test case where default key does not exist
	runners = ConfRunners{}
	defaultRunner = GetDefaultConfRunner(&runners)
	if *defaultRunner.Timeout != 0 {
		t.Errorf("Expected Timeout to be 0, got %v", *defaultRunner.Timeout)
	}
	if defaultRunner.ListenAddress != "127.0.0.1:8880" {
		t.Errorf("Expected ListenAddress to be 127.0.0.1:8080, got %s", defaultRunner.ListenAddress)
	}
	if *defaultRunner.ResponseOK != "OK" {
		t.Errorf("Expected ResponseOK to be OK, got %s", *defaultRunner.ResponseOK)
	}
	if *defaultRunner.ResponseFailed != "FAILED" {
		t.Errorf("Expected ResponseFailed to be FAILED, got %s", *defaultRunner.ResponseFailed)
	}
	if *defaultRunner.ResponseTimeout != "TIMEOUT" {
		t.Errorf("Expected ResponseTimeout to be TIMEOUT, got %s", *defaultRunner.ResponseTimeout)
	}
}

func TestGetConfRunner(t *testing.T) {
	var shutdownSignalHeader = "Test-Shutdown"
	var fiveSecond, tenSecond time.Duration = 5 * time.Second, 10 * time.Second
	var ok, failed, timeout string = "OK", "FAILED", "TIMEOUT"

	runners := ConfRunners{
		"default": ConfRunner{
			Timeout:              &fiveSecond,
			ListenAddress:        "localhost:8080",
			ResponseWriteTimeout: &tenSecond,
		},
		"testMinimalHttpRunner": ConfRunner{},
		"testHttpRunner": ConfRunner{
			Timeout:              &tenSecond,
			ShutdownSignalHeader: &shutdownSignalHeader,
			ListenAddress:        "localhost:9090",
			RequestReadTimeout:   &fiveSecond,
			ResponseWriteTimeout: &fiveSecond,
		},
	}

	defaultRunner := GetDefaultConfRunner(&runners)

	tests := []struct {
		name           string
		runnerName     string
		expectedRunner ConfRunner
		expectedExists bool
	}{
		{
			name:       "Existing runner",
			runnerName: "testHttpRunner",
			expectedRunner: ConfRunner{
				Timeout:              &tenSecond,
				ShutdownSignalHeader: &shutdownSignalHeader,
				ListenAddress:        "localhost:9090",
				RequestReadTimeout:   &fiveSecond,
				ResponseWriteTimeout: &fiveSecond,
				ResponseOK:           &ok,
				ResponseFailed:       &failed,
				ResponseTimeout:      &timeout,
			},
			expectedExists: true,
		},
		{
			name:           "Minimal http runner",
			runnerName:     "testMinimalHttpRunner",
			expectedRunner: defaultRunner,
			expectedExists: true,
		},
		{
			name:           "Default runner",
			runnerName:     "default",
			expectedRunner: defaultRunner,
			expectedExists: true,
		},
	}

	var wantTimeout, wantReadTimeout, wantWriteTimeout time.Duration = 0, 0, 0
	var wantResponseOK, wantResponseFailed, wantResponseTimeout, wantListenAddr string = "", "", "", ""

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner, exists := GetConfRunner(&runners, tt.runnerName)
			if exists != tt.expectedExists {
				t.Errorf("expected runner exists to be %v, got %v", tt.expectedExists, exists)
			}
			wantTimeout = *tt.expectedRunner.Timeout
			wantReadTimeout = *tt.expectedRunner.RequestReadTimeout
			wantWriteTimeout = *tt.expectedRunner.ResponseWriteTimeout
			wantListenAddr = tt.expectedRunner.ListenAddress
			wantResponseOK = *tt.expectedRunner.ResponseOK
			wantResponseFailed = *tt.expectedRunner.ResponseFailed
			wantResponseTimeout = *tt.expectedRunner.ResponseTimeout
			if *runner.Timeout != wantTimeout {
				t.Errorf("expected runner timeout to be %+v, got %+v", wantTimeout, runner.Timeout)
			}
			if *runner.RequestReadTimeout != wantReadTimeout {
				t.Errorf("expected runner read timeout to be %+v, got %+v", wantReadTimeout, runner.RequestReadTimeout)
			}
			if *runner.ResponseWriteTimeout != wantWriteTimeout {
				t.Errorf("expected runner write timeout to be %+v, got %+v", wantWriteTimeout, runner.ResponseWriteTimeout)
			}
			if runner.ListenAddress != wantListenAddr {
				t.Errorf("expected runner listen address to be %s, got %s", wantListenAddr, runner.ListenAddress)
			}
			if *runner.ResponseOK != wantResponseOK {
				t.Errorf("expected runner response ok to be %s, got %s", wantResponseOK, *runner.ResponseOK)
			}
			if *runner.ResponseFailed != wantResponseFailed {
				t.Errorf("expected runner response failed to be %s, got %s", wantResponseFailed, *runner.ResponseFailed)
			}
			if *runner.ResponseTimeout != wantResponseTimeout {
				t.Errorf("expected runner response timeout to be %s, got %s", wantResponseTimeout, *runner.ResponseTimeout)
			}
		})
	}
}

func TestGetConfRunnerAllowOverridesWithZeroValue(t *testing.T) {
	wantTimeout := 0 * time.Second
	wantReadTimeout := 0 * time.Second
	wantWriteTimeout := 0 * time.Second
	wantResponseOK := ""
	wantResponseFailed := ""
	wantResponseTimeout := ""
	var ok, failed, timeout string = "OK", "FAILED", "TIMEOUT"

	var fiveSecond, tenSecond time.Duration = 5 * time.Second, 10 * time.Second

	runners := ConfRunners{
		"default": ConfRunner{
			Timeout:              &fiveSecond,
			RequestReadTimeout:   &tenSecond,
			ResponseWriteTimeout: &wantWriteTimeout, // not overridden
			ResponseOK:           &ok,
			ResponseFailed:       &failed,
			ResponseTimeout:      &timeout,
		},
		"testRunner": ConfRunner{ // timeouts and reponses can be all have empty values
			Timeout:            &wantTimeout,
			RequestReadTimeout: &wantReadTimeout,
			ResponseOK:         &wantResponseOK,
			ResponseFailed:     &wantResponseFailed,
			ResponseTimeout:    &wantResponseTimeout,
		},
	}

	runner, exists := GetConfRunner(&runners, "testRunner")
	if !exists {
		t.Errorf("expected runner 'testRunner' to exist")
	}
	if *runner.Timeout != wantTimeout {
		t.Errorf("expected Timeout to be %v, got %v", wantTimeout, runner.Timeout)
	}
	if *runner.RequestReadTimeout != wantReadTimeout {
		t.Errorf("expected RequestReadTimeout to be %v, got %v", wantReadTimeout, runner.RequestReadTimeout)
	}
	if *runner.ResponseWriteTimeout != wantWriteTimeout {
		t.Errorf("expected ResponseWriteTimeout to be %v, got %v", wantWriteTimeout, runner.ResponseWriteTimeout)
	}
	if *runner.ResponseOK != wantResponseOK {
		t.Errorf("expected ResponseOK to be %v, got %v", wantResponseOK, runner.ResponseOK)
	}
	if *runner.ResponseFailed != wantResponseFailed {
		t.Errorf("expected ResponseFailed to be %v, got %v", wantResponseFailed, runner.ResponseFailed)
	}
	if *runner.ResponseTimeout != wantResponseTimeout {
		t.Errorf("expected ResponseTimeout to be %v, got %v", wantResponseTimeout, runner.ResponseTimeout)
	}
	if runner.ListenAddress == "" {
		t.Errorf("expected ListenAddress to be set by default if not configured")
	}
}
