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
	Timeout time.Duration
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
