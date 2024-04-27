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
			},
			expectedExists: true,
		},
		{
			name:           "Non-existing runner",
			runnerName:     "nonExistingRunner",
			expectedRunner: runners["default"],
			expectedExists: true,
		},
		{
			name:       "Minimal http runner",
			runnerName: "testMinimalHttpRunner",
			expectedRunner: ConfRunner{
				Timeout:              5 * time.Second,
				ShutdownSignalHeader: nil,
				ListenAddress:        "localhost:8080",
				RequestReadTimeout:   0,
				ResponseWriteTimeout: 0,
			},
			expectedExists: true,
		},
		{
			name:           "Default runner",
			runnerName:     "default",
			expectedRunner: runners["default"],
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
