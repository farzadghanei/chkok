package chkok

import (
	"errors"
	"fmt"
	"net"
	"os"
	"syscall"
	"time"
)

// Status is status of a given check (use Status* contants)
type Status uint8

const (
	// StatusUnknown is when a check has not started
	StatusUnknown Status = iota
	// StatusRunning is when a check started to run
	StatusRunning
	// StatusStopped is when a check stopped before completion (timeout, etc.)
	StatusStopped
	// StatusDone is when a check successfully ran
	StatusDone
)

// FileType is the type of a file resources, use Type* contants
type FileType uint8

const (
	// TypeAny is when a file type is not important
	TypeAny FileType = iota
	// TypeFile is used when a path should be a regular file
	TypeFile
	// TypeDir is used when a path should be a directory
	TypeDir
)

// Result is the results of a Check
type Result struct {
	IsOK   bool
	Issues []error
}

// Check is the interface that all checks implement
type Check interface {
	Suite() string
	Name() string
	Run() Result
	Result() Result
	Status() Status
}

// TimedCheck is the interface for checks that accept a timeout
type TimedCheck interface {
	GetTimeout() time.Duration
	SetTimeout(t time.Duration)
}

// CheckSuites is list of checks, grouped by suite name
type CheckSuites map[string][]Check

type baseCheck struct {
	suite  string
	name   string
	status Status
	result Result
}

func (bc *baseCheck) Suite() string {
	return bc.suite
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

// CheckFile checks for file/dir existence/type/uid/gid/size/file count
type CheckFile struct {
	baseCheck
	path         string
	fileType     FileType
	uid          int32 // -1 to skip
	gid          int32 // -1 to skip
	absent       bool
	minSize      int32 // -1 to skip
	maxSize      int64 // -1 to skip
	minFileCount int   // -1 to skip
	maxFileCount int   // -1 to skip
}

// NewCheckFile returns a new checkFile without a uid/gid/size/file count checks
func NewCheckFile(path string) *CheckFile {
	return &CheckFile{
		path:         path,
		fileType:     TypeAny,
		uid:          -1,
		gid:          -1,
		absent:       false,
		minSize:      -1,
		maxSize:      -1,
		minFileCount: -1,
		maxFileCount: -1,
	}
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

// Name returns the unique name of the check
func (chk *CheckFile) Name() string {
	return fmt.Sprintf("%v:%v", chk.typeString(), chk.path)
}

// Run runs the check
func (chk *CheckFile) Run() Result {
	if chk.path == "" {
		panic("check file path is empty")
	}

	chk.status = StatusRunning
	chk.result = Result{IsOK: true, Issues: []error{}}
	finfo, err := os.Lstat(chk.path)
	if chk.absent { // file is not there
		chk.status = StatusDone
		return chk.result
	}
	if err != nil {
		chk.result.IsOK = false
		chk.result.Issues = append(chk.result.Issues, err)
		chk.status = StatusDone
		return chk.result
	}

	var fstat *syscall.Stat_t = finfo.Sys().(*syscall.Stat_t)

	switch chk.fileType {
	case TypeDir:
		if !finfo.IsDir() {
			chk.result.IsOK = false
			chk.result.Issues = append(chk.result.Issues, errors.New("is not a directory"))
		} else if chk.minFileCount > -1 || chk.maxFileCount > -1 {
			// Only check file counts if it's a directory and we have count constraints
			chk.checkFileCount(&chk.result)
		}
	case TypeFile:
		if !finfo.Mode().IsRegular() {
			chk.result.IsOK = false
			chk.result.Issues = append(chk.result.Issues, errors.New("is not a regular file"))
		}
	}

	chk.checkUIDGID(fstat, &chk.result)
	chk.checkSize(finfo.Size(), &chk.result)
	chk.status = StatusDone
	return chk.result
}

// CheckFile.checkUIDGID checks for file uid/gid attrs updates the provided result
func (chk *CheckFile) checkUIDGID(fstat *syscall.Stat_t, result *Result) {
	if chk.uid > -1 {
		if fstat == nil {
			result.IsOK = false
			result.Issues = append(result.Issues, fmt.Errorf("check for file owner is not supported on this system"))
		} else if uint32(chk.uid) != fstat.Uid { //nolint: gosec
			result.IsOK = false
			result.Issues = append(result.Issues, fmt.Errorf("owner mismatch. want %v got %v", chk.uid, fstat.Uid))
		}
	}

	if chk.gid > -1 {
		if fstat == nil {
			result.IsOK = false
			result.Issues = append(result.Issues, fmt.Errorf("check for file group is not supported on this system"))
		} else if uint32(chk.gid) != fstat.Gid { //nolint: gosec
			result.IsOK = false
			result.Issues = append(result.Issues, fmt.Errorf("group mismatch. want %v got %v", chk.gid, fstat.Gid))
		}
	}
}

// checkSize checks for file min/max size and updates the provided result
func (chk *CheckFile) checkSize(size int64, result *Result) {
	if chk.minSize > -1 && size <= int64(chk.minSize) {
		result.IsOK = false
		result.Issues = append(result.Issues, fmt.Errorf(
			"file too small, size %v is less than min size %v", size, chk.minSize))
	}
	if chk.maxSize > -1 && size >= chk.maxSize {
		result.IsOK = false
		result.Issues = append(result.Issues, fmt.Errorf(
			"file too large, size %v is more than max size %v", size, chk.maxSize))
	}
}

// countFilesInDir counts the number of files in a directory
func countFilesInDir(dirPath string) (int, error) {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return 0, err
	}
	return len(entries), nil
}

// checkFileCount checks for directory min/max file count and updates the provided result
func (chk *CheckFile) checkFileCount(result *Result) {
	count, err := countFilesInDir(chk.path)
	if err != nil {
		result.IsOK = false
		result.Issues = append(result.Issues, fmt.Errorf("failed to count files: %v", err))
		return
	}

	if chk.minFileCount > -1 && count < chk.minFileCount {
		result.IsOK = false
		result.Issues = append(result.Issues, fmt.Errorf(
			"directory contains too few files, found %v but minimum is %v", count, chk.minFileCount))
	}

	if chk.maxFileCount > -1 && count > chk.maxFileCount {
		result.IsOK = false
		result.Issues = append(result.Issues, fmt.Errorf(
			"directory contains too many files, found %v but maximum is %v", count, chk.maxFileCount))
	}
}

// CheckDial checks for a net resource by dialing
type CheckDial struct {
	baseCheck
	Network string
	Address string
	Absent  bool
	timeout time.Duration
}

// NewCheckDial returns a checkDial for local http availability by default
func NewCheckDial() *CheckDial {
	timeout, err := time.ParseDuration("5s")
	if err != nil {
		panic("err creating check dial default timeout")
	}
	chk := CheckDial{Network: "tcp", Address: "127.0.0.1:80", Absent: false}
	chk.SetTimeout(timeout)
	return &chk
}

// Name returns the unique name of the check
func (chk *CheckDial) Name() string {
	return fmt.Sprintf("%v:%v", chk.Network, chk.Address)
}

// GetTimeout gets the max duration for the check to timeout
func (chk *CheckDial) GetTimeout() time.Duration {
	return chk.timeout
}

// SetTimeout sets the max duration for the check to timeout
func (chk *CheckDial) SetTimeout(timeout time.Duration) {
	chk.timeout = timeout
}

// Run runs the check and returns the results
func (chk *CheckDial) Run() Result {
	if chk.Network == "" {
		panic("check dial network is empty")
	}
	if chk.Address == "" {
		panic("check dial address is empty")
	}

	start := time.Now()
	chk.result = Result{IsOK: true, Issues: []error{}}
	conn, err := net.DialTimeout(chk.Network, chk.Address, chk.timeout)
	if err != nil { // no connection
		if chk.Absent {
			chk.status = StatusDone
			return chk.result
		}
		chk.result.IsOK = false
		chk.result.Issues = append(chk.result.Issues, err)
		chk.status = StatusDone
		return chk.result
	}
	defer conn.Close()
	elapsed := time.Since(start)
	if elapsed > chk.timeout {
		chk.status = StatusStopped
		chk.result.IsOK = false
		chk.result.Issues = append(chk.result.Issues, fmt.Errorf("check dial timed out after %v seconds", elapsed.Seconds()))
	} else {
		chk.status = StatusDone
	}
	return chk.result
}
