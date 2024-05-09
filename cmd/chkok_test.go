package main

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"
)

const (
	MaxHTTPRetries = 5
)

func TestRunCli(t *testing.T) {
	var buf bytes.Buffer
	cwd, _ := os.Getwd()
	baseDir, _ := filepath.Abs(filepath.Dir(cwd))
	writer := bufio.NewWriter(&buf)
	var confPath = filepath.Join(baseDir, "examples", "test.yaml")
	got := run(confPath, "cli", writer, false)
	if got != 0 {
		t.Errorf("want exit code 0, got %v. output: %v", got, buf.String())
	}
}

func TestRunHttp(t *testing.T) {
	var buf bytes.Buffer
	cwd, _ := os.Getwd()
	baseDir, _ := filepath.Abs(filepath.Dir(cwd))
	writer := bufio.NewWriter(&buf)
	var confPath = filepath.Join(baseDir, "examples", "test-http.yaml")

	go func() { // run the server in a goroutine
		run(confPath, ModeHTTP, writer, false)
	}()

	// Test the runner via an HTTP request
	// Create a new request with a context
	req, err := http.NewRequestWithContext(
		context.Background(), http.MethodGet, "http://127.0.0.1:51234", http.NoBody)
	if err != nil {
		t.Fatalf("Failed to create HTTP request: %v", err)
	}
	req.Header.Set("X-Server-Shutdown", "test-shutdown-signal") // shutdown the server after the request
	req.Header.Set("X-Required-Header", "required-value")
	req.Header.Set("X-Required-Header2", "anything")

	// Send the request multiple times, waiting for the server to
	// start up and respond
	var resp *http.Response
	var body []byte
	for i := 0; i < MaxHTTPRetries; i++ {
		resp, err = http.DefaultClient.Do(req)
		if err == nil && resp != nil { // server is up
			body, err = io.ReadAll(io.Reader(resp.Body))
			resp.Body.Close()
			if err != nil {
				t.Fatalf("Failed to read HTTP runner response body: %v", err)
			}
			break
		}
		time.Sleep(500 * time.Millisecond)
	}
	if err != nil {
		t.Fatalf("Failed to send HTTP request to HTTP runner after %v treis: %v", MaxHTTPRetries, err)
	}

	// Assert the response body
	want := "OK"
	if string(body) != want {
		t.Errorf("want response body %q, got %q", want, body)
	}
}
