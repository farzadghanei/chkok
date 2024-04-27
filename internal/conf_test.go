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
	runners := ConfRunners{
		"default": ConfRunner{
			Timeout:              10 * time.Second,
			ShutdownSignalHeader: nil,
			ListenAddress:        "localhost:8081",
			RequestReadTimeout:   wantReadTimeout,
			ResponseWriteTimeout: wantWriteTimeout,
			ResponseOK:           "YES",
			ResponseFailed:       "NO",
			ResponseTimeout:      "MAYBE",
		},
	}

	defaultRunner := GetDefaultConfRunner(&runners)
	if defaultRunner.Timeout != 10*time.Second {
		t.Errorf("Expected Timeout to be 10s, got %v", defaultRunner.Timeout)
	}
	if defaultRunner.ListenAddress != "localhost:8081" {
		t.Errorf("Expected ListenAddress to be localhost:8081, got %s", defaultRunner.ListenAddress)
	}
	if defaultRunner.RequestReadTimeout != wantReadTimeout {
		t.Errorf("Expected RequestReadTimeout to be %v, got %v", wantReadTimeout,
			defaultRunner.RequestReadTimeout)
	}
	if defaultRunner.ResponseWriteTimeout != wantWriteTimeout {
		t.Errorf("Expected ResponseWriteTimeout to be %v, got %v", wantWriteTimeout,
			defaultRunner.ResponseWriteTimeout)
	}
	if defaultRunner.ResponseOK != "YES" {
		t.Errorf("Expected ResponseOK to be YES, got %s", defaultRunner.ResponseOK)
	}
	if defaultRunner.ResponseFailed != "NO" {
		t.Errorf("Expected ResponseFailed to be NO, got %s", defaultRunner.ResponseFailed)
	}
	if defaultRunner.ResponseTimeout != "MAYBE" {
		t.Errorf("Expected ResponseTimeout to be MAYBE, got %s", defaultRunner.ResponseTimeout)
	}

	// Test case where default key does not exist
	runners = ConfRunners{}
	defaultRunner = GetDefaultConfRunner(&runners)
	if defaultRunner.Timeout != 0 {
		t.Errorf("Expected Timeout to be 0, got %v", defaultRunner.Timeout)
	}
	if defaultRunner.ListenAddress != "127.0.0.1:8880" {
		t.Errorf("Expected ListenAddress to be 127.0.0.1:8080, got %s", defaultRunner.ListenAddress)
	}
	if defaultRunner.ResponseOK != "OK" {
		t.Errorf("Expected ResponseOK to be OK, got %s", defaultRunner.ResponseOK)
	}
	if defaultRunner.ResponseFailed != "FAILED" {
		t.Errorf("Expected ResponseFailed to be FAILED, got %s", defaultRunner.ResponseFailed)
	}
	if defaultRunner.ResponseTimeout != "TIMEOUT" {
		t.Errorf("Expected ResponseTimeout to be TIMEOUT, got %s", defaultRunner.ResponseTimeout)
	}
}

func TestGetConfRunner(t *testing.T) {
	var shutdownSignalHeader = "Test-Shutdown"

	runners := ConfRunners{
		"default": ConfRunner{
			Timeout:       5 * time.Second,
			ListenAddress: "localhost:8080",
		},
		"testMinimalHttpRunner": ConfRunner{},
		"testHttpRunner": ConfRunner{
			Timeout:              10 * time.Second,
			ShutdownSignalHeader: &shutdownSignalHeader,
			ListenAddress:        "localhost:9090",
			RequestReadTimeout:   5 * time.Second,
			ResponseWriteTimeout: 5 * time.Second,
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
				Timeout:              10 * time.Second,
				ShutdownSignalHeader: &shutdownSignalHeader,
				ListenAddress:        "localhost:9090",
				RequestReadTimeout:   5 * time.Second,
				ResponseWriteTimeout: 5 * time.Second,
				ResponseOK:           "OK",
				ResponseFailed:       "FAILED",
				ResponseTimeout:      "TIMEOUT",
			},
			expectedExists: true,
		},
		{
			name:           "Non-existing runner",
			runnerName:     "nonExistingRunner",
			expectedRunner: defaultRunner,
			expectedExists: false,
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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner, exists := GetConfRunner(&runners, tt.runnerName)
			if exists != tt.expectedExists {
				t.Errorf("expected runner exists to be %v, got %v", tt.expectedExists, exists)
			}
			if runner != tt.expectedRunner {
				t.Errorf("expected runner to be %+v, got %+v", tt.expectedRunner, runner)
			}
		})
	}
}
