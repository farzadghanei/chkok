package main

import (
	"bufio"
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestRunCli(t *testing.T) {
	var buf bytes.Buffer
	cwd, _ := os.Getwd()
	baseDir, _ := filepath.Abs(filepath.Dir(cwd))
	writer := bufio.NewWriter(&buf)
	var confPath = filepath.Join(baseDir, "examples", "test.yaml")
	verbose := false
	got := run(&confPath, writer, &verbose)
	if got != 0 {
		t.Errorf("want exit code 0, got %v. output: %v", got, buf.String())
	}
}
