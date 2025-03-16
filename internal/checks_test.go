package chkok

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const UnavailableUID = 10375 // hopefully this won't be the uid on the running host
const UnavailableGID = 10375 // hopefully this won't be the gid on the running host
const UnavailablePort = 1023 // hopefully this port is not open on the running host

func TestCheckFile(t *testing.T) {
	var check *CheckFile
	var got, want string
	var gotStatus, wantStatus Status
	wantStatus = StatusUnknown
	check = NewCheckFile("/no/such/path/exists")
	if gotStatus = check.Status(); gotStatus != wantStatus {
		t.Errorf("invalid check dial status, want %v got %v", wantStatus, gotStatus)
	}
	if got = check.Suite(); got != want {
		t.Errorf("invalid check file suite, want empty got %v", got)
	}
	want = "any:/no/such/path/exists"
	if got = check.Name(); got != want {
		t.Errorf("invalid check file name, want %v got %v", want, got)
	}
	check.fileType = TypeDir
	want = "dir:/no/such/path/exists"
	if got = check.Name(); got != want {
		t.Errorf("invalid check dir name, want %v got %v", want, got)
	}
}

func TestCheckFileAbsentFiles(t *testing.T) {
	var check *CheckFile
	var gotStatus, wantStatus Status
	wantStatus = StatusDone
	check = NewCheckFile("/no/such/path/exists")
	if check.Run().IsOK {
		t.Error("invalid check file not exists, want not ok got ok")
	}
	if gotStatus = check.Status(); gotStatus != wantStatus {
		t.Errorf("invalid check file status, want %v got %v", wantStatus, gotStatus)
	}

	check = NewCheckFile("/no/such/path/exists")
	check.absent = true
	if !check.Run().IsOK {
		t.Error("invalid check file not exists, want ok got not ok")
	}
	if gotStatus = check.Status(); gotStatus != wantStatus {
		t.Errorf("invalid check file status, want %v got %v", wantStatus, gotStatus)
	}
}

//gocyclo:ignore
func TestCheckFileExistingFiles(t *testing.T) {
	var check *CheckFile
	var got Result
	var gotStatus, wantStatus Status
	const filePath = "../LICENSE"
	wantStatus = StatusDone

	check = NewCheckFile(filePath)
	if got = check.Run(); !got.IsOK {
		t.Error("invalid check file, want ok got not ok")
	} else if len(got.Issues) > 0 {
		t.Errorf("invalid check file, want no issues got %v", got.Issues)
	}
	if gotStatus = check.Status(); gotStatus != wantStatus {
		t.Errorf("invalid check file status, want %v got %v", wantStatus, gotStatus)
	}

	check.uid = UnavailableUID
	if got = check.Run(); got.IsOK {
		t.Error("invalid check file uid, want not ok got ok")
	} else if len(got.Issues) != 1 || !strings.Contains(got.Issues[0].Error(), "owner mismatch") {
		t.Errorf("invalid check file uid, want 1 issue got %v", got.Issues)
	}
	if gotStatus = check.Status(); gotStatus != wantStatus {
		t.Errorf("invalid check file status, want %v got %v", wantStatus, gotStatus)
	}

	check.uid = -1
	check.gid = UnavailableGID
	if got = check.Run(); got.IsOK {
		t.Error("invalid check file gid, want not ok got ok")
	} else if len(got.Issues) != 1 || !strings.Contains(got.Issues[0].Error(), "group mismatch") {
		t.Errorf("invalid check file gid, want 1 issue got %v", got.Issues)
	}
	if gotStatus = check.Status(); gotStatus != wantStatus {
		t.Errorf("invalid check file status, want %v got %v", wantStatus, gotStatus)
	}

	check = NewCheckFile(filePath)
	check.fileType = TypeDir
	if got = check.Run(); got.IsOK {
		t.Error("invalid check dir file, want not ok got ok")
	} else if len(got.Issues) != 1 || !strings.Contains(got.Issues[0].Error(), "is not a dir") {
		t.Errorf("invalid check dir file, want 1 issue got %v", got.Issues)
	}

	check = NewCheckFile(filePath)
	check.minSize = 0
	if got = check.Run(); !got.IsOK {
		t.Error("invalid check file min size, want ok got not ok")
	}
	check.minSize = 1024 * 1024 * 100
	if got = check.Run(); got.IsOK {
		t.Error("invalid check file min size, want not ok got ok")
	} else if len(got.Issues) != 1 || !strings.Contains(got.Issues[0].Error(), "file too small") {
		t.Errorf("invalid check file min size, want 1 issue got %v", got.Issues)
	}
	if gotStatus = check.Status(); gotStatus != wantStatus {
		t.Errorf("invalid check file status, want %v got %v", wantStatus, gotStatus)
	}

	check = NewCheckFile(filePath)
	check.maxSize = 1024 * 1024 * 100
	if got = check.Run(); !got.IsOK {
		t.Error("invalid check file max size, want ok got not ok")
	}
	check.maxSize = 1
	if got = check.Run(); got.IsOK {
		t.Error("invalid check file max size, want not ok got ok")
	} else if len(got.Issues) != 1 || !strings.Contains(got.Issues[0].Error(), "file too large") {
		t.Errorf("invalid check file max size, want 1 issue got %v", got.Issues)
	}
}

func TestCheckFileDirectories(t *testing.T) {
	var check *CheckFile
	var got Result
	const filePath = "../cmd"

	check = NewCheckFile(filePath)
	check.fileType = TypeDir
	if got = check.Run(); !got.IsOK {
		t.Error("invalid check dir dir, want ok got not ok")
	} else if len(got.Issues) > 0 {
		t.Errorf("invalid check file, want no issues got %v", got.Issues)
	}

	check.uid = UnavailableUID
	if got = check.Run(); got.IsOK {
		t.Error("invalid check dir uid, want not ok got ok")
	} else if len(got.Issues) != 1 || !strings.Contains(got.Issues[0].Error(), "owner mismatch") {
		t.Errorf("invalid check dir uid, want 1 issue got %v", got.Issues)
	}

	check.uid = -1
	check.gid = UnavailableGID
	if got = check.Run(); got.IsOK {
		t.Error("invalid check dir gid, want not ok got ok")
	} else if len(got.Issues) != 1 || !strings.Contains(got.Issues[0].Error(), "group mismatch") {
		t.Errorf("invalid check dir gid, want 1 issue got %v", got.Issues)
	}
}

func TestCheckFileCount(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "file-count-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create some test files
	for i := 0; i < 5; i++ {
		filename := filepath.Join(tempDir, fmt.Sprintf("file%d.txt", i))
		if err := os.WriteFile(filename, []byte("test"), 0600); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	// Test cases
	testCases := []struct {
		name       string
		minCount   int
		maxCount   int
		expectPass bool
	}{
		{"No constraints", -1, -1, true},
		{"Min count satisfied", 3, -1, true},
		{"Min count not satisfied", 10, -1, false},
		{"Max count satisfied", -1, 10, true},
		{"Max count not satisfied", -1, 3, false},
		{"Both constraints satisfied", 3, 10, true},
		{"Both constraints not satisfied", 10, 20, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			check := NewCheckFile(tempDir)
			check.fileType = TypeDir
			check.minFileCount = tc.minCount
			check.maxFileCount = tc.maxCount

			result := check.Run()

			if result.IsOK != tc.expectPass {
				t.Errorf("Expected IsOK=%v but got %v. Issues: %v",
					tc.expectPass, result.IsOK, result.Issues)
			}
		})
	}
}

func TestCheckDial(t *testing.T) {
	var check *CheckDial
	var got, want string
	check = NewCheckDial()
	if got = check.Suite(); got != want {
		t.Errorf("invalid check dial suite, want empty got %v", got)
	}
	want = "tcp:127.0.0.1:80"
	if got = check.Name(); got != want {
		t.Errorf("invalid check dial name, want %v got %v", want, got)
	}
	var gotStatus, wantStatus Status
	wantStatus = StatusUnknown
	if gotStatus = check.Status(); gotStatus != wantStatus {
		t.Errorf("invalid check dial status, want %v got %v", wantStatus, gotStatus)
	}
}

func TestCheckDialTCPPortAbsent(t *testing.T) {
	var check *CheckDial
	var got Result
	var gotStatus, wantStatus Status
	wantStatus = StatusDone
	check = NewCheckDial()
	check.Address = fmt.Sprintf("localhost:%d", UnavailablePort)
	check.timeout, _ = time.ParseDuration("500ms")
	check.Absent = true
	if got = check.Run(); !got.IsOK {
		t.Fatalf("invalid check dial, want ok got not ok")
	}
	if gotStatus = check.Status(); gotStatus != wantStatus {
		t.Errorf("invalid check dial status, want %v got %v", wantStatus, gotStatus)
	}
}

func TestCheckDialTCPPort(t *testing.T) {
	var check *CheckDial
	var got Result
	var gotStatus, wantStatus Status
	wantStatus = StatusDone
	check = NewCheckDial()
	check.timeout, _ = time.ParseDuration("500ms")
	check.Address = fmt.Sprintf("localhost:%d", UnavailablePort)
	if got = check.Run(); got.IsOK {
		t.Fatalf("invalid check dial, want not ok got ok")
	}
	if gotStatus = check.Status(); gotStatus != wantStatus {
		t.Errorf("invalid check dial status, want %v got %v", wantStatus, gotStatus)
	}
}
