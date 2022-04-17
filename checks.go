package chkok

import (
	"errors"
	"fmt"
	"net"
	"os"
	"syscall"
	"time"
)

type Status uint8

const (
	StatusUnknown Status = iota
	StatusRunning
	StatusStopped
	StatusDone
)

type FileType uint8

const (
	TypeAny FileType = iota
	TypeFile
	TypeDir
)

type Result struct {
	IsOK   bool
	Issues []error
}

type Check interface {
	Group() string
	Name() string
	Run() Result
	Result() Result
	Status() Status
}

type CheckGroups map[string][]Check

type baseCheck struct {
	group  string  // TODO: rename group to chain/suite/line to avoid clashing with file group
	name   string
	status Status
	result Result
}

func (bc *baseCheck) Group() string {
	return bc.group
}

func (bc *baseCheck) Name() string {
	return bc.name
}

func (bc *baseCheck) Result() Result {
	return bc.result
}

func (bc *baseCheck) Status() Status {
	return bc.status
}

// CheckFile checks for file/dir existence/type/uid/gid/size
type CheckFile struct {
	baseCheck
	path     string
	fileType FileType
	uid      int32 // -1 to skip
	gid      int32 // -1 to skip
	absent   bool
	minSize  int32 // -1 to sktip
	maxSize  int64 // 0 to skip
}

// NewCheckFile returns a new checkFile without a uid/gid/size checks
func NewCheckFile(path string) *CheckFile {
	return &CheckFile{path: path, fileType: TypeAny, uid: -1, gid: -1, absent: false, minSize: -1, maxSize: -1}
}

func (chk *CheckFile) typeString() string {
	switch chk.fileType {
	case TypeFile:
		return "file"
	case TypeDir:
		return "dir"
	}
	return "any"
}

func (chk *CheckFile) Name() string {
	return fmt.Sprintf("%v:%v", chk.typeString(), chk.path)
}

// checkFile.Run runs the check
func (chk *CheckFile) Run() Result {
	if chk.path == "" {
		panic("check file path is empty")
	}

	chk.status = StatusRunning
	result := Result{IsOK: true, Issues: []error{}}
	finfo, err := os.Lstat(chk.path)
	if chk.absent { // file is not there
		chk.status = StatusDone
		return result
	}
	if err != nil {
		result.IsOK = false
		result.Issues = append(result.Issues, err)
		chk.status = StatusDone
		return result
	}

	var fstat *syscall.Stat_t = finfo.Sys().(*syscall.Stat_t)

	switch chk.fileType {
	case TypeDir:
		if !finfo.IsDir() {
			result.IsOK = false
			result.Issues = append(result.Issues, errors.New("is not a directory"))
		}
	case TypeFile:
		if !finfo.Mode().IsRegular() {
			result.IsOK = false
			result.Issues = append(result.Issues, errors.New("is not a regular file"))
		}
	}

	chk.checkUidGid(fstat, &result)
	chk.checkSize(finfo.Size(), &result)
	chk.status = StatusDone
	return result
}

// checkFile.checkUidGid checks for file uid/gid attrs updates the provided result
func (chk *CheckFile) checkUidGid(fstat *syscall.Stat_t, result *Result) {
	if chk.uid > -1 {
		if fstat == nil {
			result.IsOK = false
			result.Issues = append(result.Issues, fmt.Errorf("check for file owner is not supported on this system"))
		} else if uint32(chk.uid) != fstat.Uid {
			result.IsOK = false
			result.Issues = append(result.Issues, fmt.Errorf("owner mismatch. want %v got %v", chk.uid, fstat.Uid))
		}
	}

	if chk.gid > -1 {
		if fstat == nil {
			result.IsOK = false
			result.Issues = append(result.Issues, fmt.Errorf("check for file group is not supported on this system"))
		} else if uint32(chk.gid) != fstat.Gid {
			result.IsOK = false
			result.Issues = append(result.Issues, fmt.Errorf("group mismatch. want %v got %v", chk.gid, fstat.Gid))
		}
	}
}

// checkFile.checkSize checks for file min/max size and updates the provided result
func (chk *CheckFile) checkSize(size int64, result *Result) {
	if chk.minSize > -1 && size <= int64(chk.minSize) {
		result.IsOK = false
		result.Issues = append(result.Issues, fmt.Errorf("file too small, size %v is less than min size %v", size, chk.minSize))
	}
	if chk.maxSize > -1 && size >= chk.maxSize {
		result.IsOK = false
		result.Issues = append(result.Issues, fmt.Errorf("file too large, size %v is more than max size %v", size, chk.maxSize))
	}
}

// CheckDial checks for a net resource by dialing
type CheckDial struct {
	baseCheck
	Network string
	Address string
	Absent  bool
	Timeout time.Duration
}

// NewCheckDial returns a checkDial for local http availablity by default
func NewCheckDial() *CheckDial {
	timeout, err := time.ParseDuration("5s")
	if err != nil {
		panic("err creating check dial default timeout")
	}
	check := CheckDial{Network: "tcp", Address: "127.0.0.1:80", Absent: false}
	check.Timeout = timeout
	return &check
}

func (chk *CheckDial) Name() string {
	return fmt.Sprintf("%v:%v", chk.Network, chk.Address)
}

// checkDial.Run runs the check and returns the results
func (chk *CheckDial) Run() Result {
	if chk.Network == "" {
		panic("check dial network is empty")
	}
	if chk.Address == "" {
		panic("check dial address is empty")
	}

	start := time.Now()
	result := Result{IsOK: true, Issues: []error{}}
	conn, err := net.DialTimeout(chk.Network, chk.Address, chk.Timeout)
	if err != nil { // no connection
		if chk.Absent {
			chk.status = StatusDone
			return result
		} else {
			result.IsOK = false
			result.Issues = append(result.Issues, err)
			chk.status = StatusDone
			return result
		}
	}
	defer conn.Close()
	elapsed := time.Since(start)
	if elapsed > chk.Timeout {
		chk.status = StatusStopped
		result.IsOK = false
		result.Issues = append(result.Issues, fmt.Errorf("check dial timed out after %v seconds", elapsed.Seconds()))
	} else {
		chk.status = StatusDone
	}
	return result
}
