package chkok

import (
	"fmt"
	"strings"
)

// CheckSuitesFromSpecSuites creates checksuites from config check spec suites
func CheckSuitesFromSpecSuites(specSuites ConfCheckSpecSuites) (CheckSuites, error) {
	var checkSuites CheckSuites = make(map[string][]Check)
	var check Check
	var err error
	for suite, specs := range specSuites {
		checkSuites[suite] = []Check{}
		for _, spec := range specs {
			check, err = CheckFromSpec(spec)
			if err != nil {
				return checkSuites, err
			}
			checkSuites[suite] = append(checkSuites[suite], check)
		}
	}
	return checkSuites, err
}

// CheckFromSpec creates a check from a given ConfCheckSpec
func CheckFromSpec(spec ConfCheckSpec) (Check, error) {
	var err error
	var check Check
	switch checkType := strings.ToLower(spec.Type); checkType {
	case "file", "dir":
		check, err = CheckFileFromSpec(spec)
	case "dial":
		check, err = CheckDialFromSpec(spec)
	default:
		check = NewCheckFile("/")
		err = fmt.Errorf("invalid check type '%v'", checkType)
	}
	return check, err
}

// CheckFileFromSpec creates a CheckFile from a ConfCheckSpec
func CheckFileFromSpec(spec ConfCheckSpec) (*CheckFile, error) {
	var err error
	var id int
	check := NewCheckFile(spec.Path)
	if specType := strings.ToLower(spec.Type); specType == "dir" {
		check.fileType = TypeDir
	} else if specType == "file" {
		check.fileType = TypeFile
	}
	check.absent = spec.Absent
	check.path = spec.Path
	check.minSize = spec.MinSize

	check.maxSize = -1
	check.uid = -1
	check.gid = -1
	if spec.MaxSize != nil {
		check.maxSize = *spec.MaxSize
	}
	if spec.User != nil {
		id, err = getUID(*spec.User)
		if err == nil {
			check.uid = int32(id)
		}
	}
	if spec.Group != nil {
		id, err = getGID(*spec.Group)
		if err == nil {
			check.gid = int32(id)
		}
	}
	return check, err
}

// CheckDialFromSpec creates a CheckDial from a ConfCheckSpec
func CheckDialFromSpec(spec ConfCheckSpec) (*CheckDial, error) {
	var err error
	check := NewCheckDial()
	check.Absent = spec.Absent
	switch network := strings.ToLower(spec.Network); network {
	case "tcp":
		check.Network = network
	default:
		err = fmt.Errorf("dial check network '%v' is not supported", spec.Network)
	}
	check.Address = spec.Address
	check.Timeout = spec.Timeout
	return check, err
}
