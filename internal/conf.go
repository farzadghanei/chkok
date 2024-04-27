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
	Timeout               time.Duration
	ShutdownAfterRequests uint32        `yaml:"shutdown_after_requests"`
	ListenAddress         string        `yaml:"listen_address"`
	RequestReadTimeout    time.Duration `yaml:"request_read_timeout"`
	ResponseWriteTimeout  time.Duration `yaml:"response_write_timeout"`
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

// GetConfRunner returns the runner config for the name merged with the default, and bool if it exists
func GetConfRunner(runners *ConfRunners, name string) (ConfRunner, bool) {
	defaultConf, defaultExists := (*runners)["default"]
	namedConf, namedExists := (*runners)[name]

	if !defaultExists && !namedExists {
		return ConfRunner{}, false
	}

	if !defaultExists {
		return namedConf, true
	}

	if !namedExists {
		return defaultConf, true
	}

	// Merge the requested runner with the default runner
	mergedConf := ConfRunner{
		Timeout:               namedConf.Timeout,
		ShutdownAfterRequests: namedConf.ShutdownAfterRequests,
		ListenAddress:         namedConf.ListenAddress,
		RequestReadTimeout:    namedConf.RequestReadTimeout,
		ResponseWriteTimeout:  namedConf.ResponseWriteTimeout,
	}

	if mergedConf.Timeout == 0 {
		mergedConf.Timeout = defaultConf.Timeout
	}
	if mergedConf.ShutdownAfterRequests == 0 {
		mergedConf.ShutdownAfterRequests = defaultConf.ShutdownAfterRequests
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

	return mergedConf, true
}
