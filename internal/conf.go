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
	Timeout              time.Duration
	ShutdownSignalHeader *string       `yaml:"shutdown_signal_header"`
	ListenAddress        string        `yaml:"listen_address"`
	RequestReadTimeout   time.Duration `yaml:"request_read_timeout"`
	ResponseWriteTimeout time.Duration `yaml:"response_write_timeout"`
	ResponseOK           string        `yaml:"response_ok"`
	ResponseFailed       string        `yaml:"response_failed"`
	ResponseTimeout      string        `yaml:"response_timeout"`
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

// GetDefaultConfRunner returns a ConfRunner based on the default configuration
func GetDefaultConfRunner(runners *ConfRunners) ConfRunner {
	defaultRunner := ConfRunner{
		Timeout:              0,
		ShutdownSignalHeader: nil,
		ListenAddress:        "127.0.0.1:8880",
		RequestReadTimeout:   30 * time.Second,
		ResponseWriteTimeout: 30 * time.Second,
		ResponseOK:           "OK",
		ResponseFailed:       "FAILED",
		ResponseTimeout:      "TIMEOUT",
	}

	if defaultConf, defaultExists := (*runners)["default"]; defaultExists {
		if defaultConf.Timeout != 0 {
			defaultRunner.Timeout = defaultConf.Timeout
		}
		if defaultConf.ShutdownSignalHeader != nil {
			defaultRunner.ShutdownSignalHeader = defaultConf.ShutdownSignalHeader
		}
		if defaultConf.ListenAddress != "" {
			defaultRunner.ListenAddress = defaultConf.ListenAddress
		}
		if defaultConf.RequestReadTimeout != 0 {
			defaultRunner.RequestReadTimeout = defaultConf.RequestReadTimeout
		}
		if defaultConf.ResponseWriteTimeout != 0 {
			defaultRunner.ResponseWriteTimeout = defaultConf.ResponseWriteTimeout
		}
		if defaultConf.ResponseOK != "" {
			defaultRunner.ResponseOK = defaultConf.ResponseOK
		}
		if defaultConf.ResponseFailed != "" {
			defaultRunner.ResponseFailed = defaultConf.ResponseFailed
		}
		if defaultConf.ResponseTimeout != "" {
			defaultRunner.ResponseTimeout = defaultConf.ResponseTimeout
		}
	}

	return defaultRunner
}

// GetConfRunner returns the runner config for the name merged with the default, and bool if it exists
func GetConfRunner(runners *ConfRunners, name string) (ConfRunner, bool) {
	defaultConf := GetDefaultConfRunner(runners)
	namedConf, namedExists := (*runners)[name]

	if !namedExists {
		return defaultConf, false
	}

	// Merge the requested runner with the default runner
	mergedConf := ConfRunner{
		Timeout:              namedConf.Timeout,
		ShutdownSignalHeader: namedConf.ShutdownSignalHeader,
		ListenAddress:        namedConf.ListenAddress,
		RequestReadTimeout:   namedConf.RequestReadTimeout,
		ResponseWriteTimeout: namedConf.ResponseWriteTimeout,
		ResponseOK:           namedConf.ResponseOK,
		ResponseFailed:       namedConf.ResponseFailed,
		ResponseTimeout:      namedConf.ResponseTimeout,
	}

	if mergedConf.Timeout == 0 {
		mergedConf.Timeout = defaultConf.Timeout
	}
	if mergedConf.ShutdownSignalHeader == nil {
		mergedConf.ShutdownSignalHeader = defaultConf.ShutdownSignalHeader
	}
	if mergedConf.ListenAddress == "" {
		mergedConf.ListenAddress = defaultConf.ListenAddress
	}
	if mergedConf.RequestReadTimeout == 0 {
		mergedConf.RequestReadTimeout = defaultConf.RequestReadTimeout
	}
	if mergedConf.ResponseWriteTimeout == 0 {
		mergedConf.ResponseWriteTimeout = defaultConf.ResponseWriteTimeout
	}
	if mergedConf.ResponseOK == "" {
		mergedConf.ResponseOK = defaultConf.ResponseOK
	}
	if mergedConf.ResponseFailed == "" {
		mergedConf.ResponseFailed = defaultConf.ResponseFailed
	}
	if mergedConf.ResponseTimeout == "" {
		mergedConf.ResponseTimeout = defaultConf.ResponseTimeout
	}

	return mergedConf, true
}
