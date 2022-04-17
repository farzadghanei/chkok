package chkok

import (
	"testing"
)

func TestReadConfErrors(t *testing.T) {
	var conf *Conf
	var err error
	_, err = ReadConf("/no/such/path/exists")
	if err == nil {
		t.Errorf("invalid read conf, want err got nil")
	}
	conf, err = ReadConf("LISENSE")
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
	conf, err = ReadConf("examples/config.yaml")
	if err != nil {
		t.Errorf("invalid read conf, want nil err got err %v", err)
	}
	runner, ok := conf.Runners["default"]
	if !ok {
		t.Errorf("invalid read conf, want default runner")
	}
	wantMaxRunning := 3
	if runner.MaxRunning != int32(wantMaxRunning) {
		t.Errorf("invalid read conf default runner, want %v max running got %v", wantMaxRunning, runner.MaxRunning)
	}
	wantMinutes := 5
	if runner.Timeout.Minutes() != float64(wantMinutes) {
		t.Errorf("invalid read conf default runner, want %v timeout got %v", wantMinutes, runner.Timeout.Minutes())
	}
	etcChecks, ok := conf.Checks["etc"]
	if !ok {
		t.Errorf("read conf found no etc checks")
	}
	if len(etcChecks) != 3 {
		t.Errorf("read conf etc checks failed, want 3 checks found %v", len(etcChecks))
	}
	defaultChecks, ok := conf.Checks["default"]
	if !ok {
		t.Errorf("read conf found no default checks")
	}
	if len(defaultChecks) != 3 {
		t.Errorf("read conf default checks failed, want 3 checks found %v", len(etcChecks))
	}
	invalidChecks, ok := conf.Checks["invalid"]
	if ok {
		t.Errorf("read conf found invalid check group but shouldn't")
	}
	if len(invalidChecks) > 0 {
		t.Errorf("read conf found invalid checks but shouldn't")
	}
}
