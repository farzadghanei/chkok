package chkok

import (
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type ConfRunners map[string]ConfRunner
type ConfCheckSpecSuites map[string][]ConfCheckSpec

type Conf struct {
	Runners ConfRunners
	Checks  ConfCheckSpecSuites
}

type ConfRunner struct {
	MaxRunning int32 `yaml:"max_running"`
	Timeout    time.Duration
}

type ConfCheckSpec struct {
	Type    string
	Path    string
	Mode    *uint32
	User    *string
	Group   *string
	MinSize int32 `yaml:"min_size"`
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
