package chkok

import (
	"maps"
	"testing"
	"time"
)

// TestReadConfErrors tests the ReadConf function for error handling
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
	baseRunner := GetBaseConfRunner()
	runners = ConfRunners{}
	defaultRunner = GetDefaultConfRunner(&runners)
	if *defaultRunner.Timeout != *baseRunner.Timeout {
		t.Errorf("Timeout want %v, got %v", *baseRunner.Timeout, *defaultRunner.Timeout)
	}
	if defaultRunner.ListenAddress != baseRunner.ListenAddress {
		t.Errorf("ListenAddress want %v, got %s", baseRunner.ListenAddress, defaultRunner.ListenAddress)
	}
	if *defaultRunner.ResponseOK != *baseRunner.ResponseOK {
		t.Errorf("ResponseOK want %v, got %s", *baseRunner.ResponseOK, *defaultRunner.ResponseOK)
	}
	if *defaultRunner.ResponseFailed != *baseRunner.ResponseFailed {
		t.Errorf("ResponseFailed want %v, got %s", *baseRunner.ResponseFailed, *defaultRunner.ResponseFailed)
	}
	if *defaultRunner.ResponseTimeout != *baseRunner.ResponseTimeout {
		t.Errorf("ResponseTimeout want %v, got %s", *baseRunner.ResponseTimeout, *defaultRunner.ResponseTimeout)
	}
	if *defaultRunner.MaxHeaderBytes != *baseRunner.MaxHeaderBytes {
		t.Errorf("MaxHeaderBytes want %v, got %v", *baseRunner.MaxHeaderBytes, *defaultRunner.MaxHeaderBytes)
	}
	if *defaultRunner.MaxConcurrentRequests != *baseRunner.MaxConcurrentRequests {
		t.Errorf("MaxConcurrentRequests want %v, got %v", *baseRunner.MaxConcurrentRequests,
			*defaultRunner.MaxConcurrentRequests)
	}
}

func TestGetConfRunner(t *testing.T) {
	var shutdownSignalHeader = "Test-Shutdown"
	var fiveSecond, tenSecond time.Duration = 5 * time.Second, 10 * time.Second
	var ok, failed, timeout string = "OK", "FAILED", "TIMEOUT"

	runners := ConfRunners{
		"default": ConfRunner{
			Timeout:                &fiveSecond,
			ListenAddress:          "localhost:8080",
			ResponseWriteTimeout:   &tenSecond,
			RequestRequiredHeaders: map[string]string{"X-Test-Default": "test"},
		},
		"testMinimalHttpRunner": ConfRunner{},
		"testHttpRunner": ConfRunner{
			Timeout:                &tenSecond,
			ShutdownSignalHeader:   &shutdownSignalHeader,
			ListenAddress:          "localhost:9090",
			RequestReadTimeout:     &fiveSecond,
			ResponseWriteTimeout:   &fiveSecond,
			RequestRequiredHeaders: map[string]string{"X-Test-2": "http-test"},
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
				Timeout:                &tenSecond,
				ShutdownSignalHeader:   &shutdownSignalHeader,
				ListenAddress:          "localhost:9090",
				RequestReadTimeout:     &fiveSecond,
				RequestRequiredHeaders: map[string]string{"X-Test-2": "http-test", "X-Test-Default": "test"},
				ResponseWriteTimeout:   &fiveSecond,
				ResponseOK:             &ok,
				ResponseFailed:         &failed,
				ResponseTimeout:        &timeout,
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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner, exists := GetConfRunner(&runners, tt.runnerName)
			if exists != tt.expectedExists {
				t.Errorf("exists want %v got %v", tt.expectedExists, exists)
			}
			expRunner := tt.expectedRunner
			if *runner.Timeout != *expRunner.Timeout {
				t.Errorf("timout want %v got %+v", *expRunner.Timeout, runner.Timeout)
			}
			if *runner.RequestReadTimeout != *expRunner.RequestReadTimeout {
				t.Errorf("read timeout want %+v got %+v", *expRunner.Timeout, runner.RequestReadTimeout)
			}
			if *runner.ResponseWriteTimeout != *expRunner.ResponseWriteTimeout {
				t.Errorf("write timeout want %+v got %+v", *expRunner.ResponseWriteTimeout, runner.ResponseWriteTimeout)
			}
			if runner.ListenAddress != expRunner.ListenAddress {
				t.Errorf("listen address want %s got %s", expRunner.ListenAddress, runner.ListenAddress)
			}
			if *runner.ResponseOK != *expRunner.ResponseOK {
				t.Errorf("response ok want %s got %s", *expRunner.ResponseOK, *runner.ResponseOK)
			}
			if *runner.ResponseFailed != *expRunner.ResponseFailed {
				t.Errorf("response failed want %s got %s", *expRunner.ResponseFailed, *runner.ResponseFailed)
			}
			if *runner.ResponseTimeout != *expRunner.ResponseTimeout {
				t.Errorf("response timeout want %s got %s", *expRunner.ResponseTimeout, *runner.ResponseTimeout)
			}
			if !maps.Equal(expRunner.RequestRequiredHeaders, runner.RequestRequiredHeaders) {
				t.Errorf("request headers want %+v got %+v", expRunner.RequestRequiredHeaders, runner.RequestRequiredHeaders)
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
