package chkok

import (
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// ConfRunners is a map of runner name to its config ConfRunner
type ConfRunners map[string]ConfRunner

// ConfCheckSpecSuites is list of ConfCheckSpec grouped by name
type ConfCheckSpecSuites map[string][]ConfCheckSpec

// Conf is app configurations struct
type Conf struct {
	Runners     ConfRunners
	CheckSuites ConfCheckSpecSuites `yaml:"check_suites"`
}

// ConfRunner is config for the check runners
type ConfRunner struct {
	Timeout               *time.Duration
	ShutdownSignalHeader  *string        `yaml:"shutdown_signal_header"`
	MaxHeaderBytes        *int           `yaml:"max_header_bytes"`
	MaxConcurrentRequests *int           `yaml:"max_concurrent_requests"`
	ListenAddress         string         `yaml:"listen_address"`
	RequestReadTimeout    *time.Duration `yaml:"request_read_timeout"`
	ResponseWriteTimeout  *time.Duration `yaml:"response_write_timeout"`
	ResponseOK            *string        `yaml:"response_ok"`
	ResponseFailed        *string        `yaml:"response_failed"`
	ResponseTimeout       *string        `yaml:"response_timeout"`
	ResponseUnavailable   *string        `yaml:"response_unavailable"`
}

// ConfCheckSpec is the spec for each check configuration
type ConfCheckSpec struct {
	Type    string
	Path    string
	Mode    *uint32
	User    *string
	Group   *string
	MinSize int32  `yaml:"min_size"`
	MaxSize *int64 `yaml:"max_size"`
	Absent  bool
	Network string
	Address string
	Timeout time.Duration
}

// ReadConf reads the configuration file and returns a pointer to Conf struct
func ReadConf(path string) (*Conf, error) {
	var conf Conf
	contents, err := os.ReadFile(path)
	if err != nil {
		return &conf, err
	}
	err = yaml.Unmarshal(contents, &conf)
	return &conf, err
}

// GetBaseConfRunner returns a base ConfRunner with default literal values
func GetBaseConfRunner() ConfRunner {
	var timeout, readTimeout, writeTimout time.Duration = 5 * time.Minute, 30 * time.Second, 30 * time.Second
	var maxHeaderBytes int = 8 * 1024
	var MaxConcurrentRequests int = 1
	var respOK, respFailed, respTimeout string = "OK", "FAILED", "TIMEOUT"
	var respUnavailable string = "UNAVAILABLE"

	baseConf := ConfRunner{
		Timeout:               &timeout,
		ShutdownSignalHeader:  nil,
		MaxHeaderBytes:        &maxHeaderBytes,
		ListenAddress:         "127.0.0.1:8880",
		RequestReadTimeout:    &readTimeout,
		ResponseWriteTimeout:  &writeTimout,
		ResponseOK:            &respOK,
		ResponseFailed:        &respFailed,
		ResponseTimeout:       &respTimeout,
		ResponseUnavailable:   &respUnavailable,
		MaxConcurrentRequests: &MaxConcurrentRequests,
	}
	return baseConf
}

// GetDefaultConfRunner returns a ConfRunner based on the default configuration
func GetDefaultConfRunner(runners *ConfRunners) ConfRunner {
	baseConf := GetBaseConfRunner()

	if defaultConf, defaultExists := (*runners)["default"]; defaultExists {
		baseConf = MergedConfRunners(&baseConf, &defaultConf)
	}

	return baseConf
}

// GetConfRunner returns the runner config for the name merged with the default, and bool if it exists
func GetConfRunner(runners *ConfRunners, name string) (ConfRunner, bool) {
	defaultConf := GetDefaultConfRunner(runners)
	namedConf, namedExists := (*runners)[name]

	if !namedExists {
		return defaultConf, false
	}

	// Merge the requested runner with the default runner
	mergedConf := MergedConfRunners(&defaultConf, &namedConf)

	return mergedConf, true
}

// MergedConfRunners merges the baseConf with the overrideConf and returns the merged ConfRunner
func MergedConfRunners(baseConf, overrideConf *ConfRunner) ConfRunner {
	mergedConf := CopyConfRunner(overrideConf)

	if mergedConf.Timeout == nil {
		mergedConf.Timeout = baseConf.Timeout
	}

	if mergedConf.ShutdownSignalHeader == nil {
		mergedConf.ShutdownSignalHeader = baseConf.ShutdownSignalHeader
	}

	if mergedConf.MaxHeaderBytes == nil {
		mergedConf.MaxHeaderBytes = baseConf.MaxHeaderBytes
	}

	if mergedConf.ListenAddress == "" {
		mergedConf.ListenAddress = baseConf.ListenAddress
	}

	if mergedConf.RequestReadTimeout == nil {
		mergedConf.RequestReadTimeout = baseConf.RequestReadTimeout
	}

	if mergedConf.ResponseWriteTimeout == nil {
		mergedConf.ResponseWriteTimeout = baseConf.ResponseWriteTimeout
	}

	if mergedConf.ResponseOK == nil {
		mergedConf.ResponseOK = baseConf.ResponseOK
	}

	if mergedConf.ResponseFailed == nil {
		mergedConf.ResponseFailed = baseConf.ResponseFailed
	}

	if mergedConf.ResponseTimeout == nil {
		mergedConf.ResponseTimeout = baseConf.ResponseTimeout
	}

	if mergedConf.ResponseUnavailable == nil {
		mergedConf.ResponseUnavailable = baseConf.ResponseUnavailable
	}

	if mergedConf.MaxConcurrentRequests == nil {
		mergedConf.MaxConcurrentRequests = baseConf.MaxConcurrentRequests
	}

	return mergedConf
}

// CopyConfRunner returns a copy of the ConfRunner with the same values
func CopyConfRunner(conf *ConfRunner) ConfRunner {
	return ConfRunner{
		Timeout:               conf.Timeout,
		ShutdownSignalHeader:  conf.ShutdownSignalHeader,
		ListenAddress:         conf.ListenAddress,
		RequestReadTimeout:    conf.RequestReadTimeout,
		ResponseWriteTimeout:  conf.ResponseWriteTimeout,
		ResponseOK:            conf.ResponseOK,
		ResponseFailed:        conf.ResponseFailed,
		ResponseTimeout:       conf.ResponseTimeout,
		ResponseUnavailable:   conf.ResponseUnavailable,
		MaxHeaderBytes:        conf.MaxHeaderBytes,
		MaxConcurrentRequests: conf.MaxConcurrentRequests,
	}
}
