package chkok

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync/atomic"
	"time"
)

const (
	// ShutdownTimeout is the time to wait for server to shutdown
	ShutdownTimeout = 5 * time.Second
)

// RunModeCLI run app in CLI mode using the provided configs, return exit code
func RunModeCLI(checkGroups *CheckSuites, conf *ConfRunner, output io.Writer, logger *log.Logger) int {
	runner := Runner{Log: logger, Timeout: *conf.Timeout}
	passed, failed, timedout := runChecks(&runner, checkGroups, logger)
	total := passed + failed + timedout
	if timedout > 0 {
		fmt.Fprintf(output, "%v/%v checks timedout", timedout, total)
		return ExTempFail
	}
	if failed > 0 {
		fmt.Fprintf(output, "%v/%v checks failed", failed, total)
		return ExSoftware
	}
	fmt.Fprintf(output, "%v checks passed", total)
	return ExOK
}

func httpRequestAsString(r *http.Request) string {
	return fmt.Sprintf("%s - %s %s %s", r.RemoteAddr, r.Proto, r.Method, r.URL)
}

// RunModeHTTP runs app in http server mode using the provided config, return exit code
func RunModeHTTP(checkGroups *CheckSuites, conf *ConfRunner, logger *log.Logger) int {
	shutdownSignalHeaderValue := ""
	if conf.ShutdownSignalHeader != nil {
		shutdownSignalHeaderValue = *conf.ShutdownSignalHeader
	}
	maxConcurrentRequests := *conf.MaxConcurrentRequests
	responseOK := *conf.ResponseOK
	responseFailed := *conf.ResponseFailed
	responseTimeout := *conf.ResponseTimeout
	responseUnavailable := *conf.ResponseUnavailable

	runner := Runner{Log: logger, Timeout: *conf.Timeout}

	var runningRequests atomic.Int32
	var reqHandlerChan = make(chan *http.Request, 1)

	httpHandler := func(w http.ResponseWriter, r *http.Request) {
		runningRequests.Add(1)
		if maxConcurrentRequests > 0 && runningRequests.Load() > int32(maxConcurrentRequests) {
			w.WriteHeader(http.StatusServiceUnavailable) // 503
			fmt.Fprint(w, responseUnavailable)
		}
		defer runningRequests.Add(-1)
		logger.Printf("processing http request: %s", httpRequestAsString(r))
		_, failed, timedout := runChecks(&runner, checkGroups, logger)
		if timedout > 0 {
			w.WriteHeader(http.StatusGatewayTimeout) // 504
			fmt.Fprint(w, responseTimeout)
		} else if failed > 0 {
			w.WriteHeader(http.StatusInternalServerError) // 500
			fmt.Fprint(w, responseFailed)
		} else {
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, responseOK)
		}
		reqHandlerChan <- r
	}

	http.HandleFunc("/", httpHandler)

	server := &http.Server{
		Addr:           conf.ListenAddress,
		Handler:        nil, // use http.DefaultServeMux
		ReadTimeout:    *conf.RequestReadTimeout,
		WriteTimeout:   *conf.ResponseWriteTimeout,
		IdleTimeout:    0 * time.Second, // set to 0 so uses read timeout
		MaxHeaderBytes: *conf.MaxHeaderBytes,
	}

	var count uint32 = 0

	go func() {
		var request *http.Request
		timeoutCtx, cancel := context.WithTimeout(context.Background(), ShutdownTimeout)
		defer cancel()
		for request = range reqHandlerChan {
			atomic.AddUint32(&count, 1)
			logger.Printf("request [%v] is processed: %v", count, httpRequestAsString(request))
			if shutdownSignalHeaderValue != "" && request.Header.Get("X-Server-Shutdown") == shutdownSignalHeaderValue {
				if err := server.Shutdown(timeoutCtx); err != nil {
					logger.Printf("http server shutdown failed: %v", err)
				} else {
					logger.Printf("http server shutdown signal received!")
				}
				return
			}
		}
	}()

	logger.Printf("starting http server listening on %s", conf.ListenAddress)
	err := server.ListenAndServe()
	close(reqHandlerChan)
	if err != nil && err != http.ErrServerClosed {
		logger.Printf("http server failed to start: %v", err)
		return ExSoftware
	}
	logger.Printf("http server shutdown!")
	return ExOK
}

// runChecks runs checks with logs, and returns number of passed, failed and timedout checks
func runChecks(runner *Runner, checkGroups *CheckSuites, logger *log.Logger) (passed, failed, timedout int) {
	checks := runner.RunChecks(*checkGroups)
	for _, chk := range checks {
		if chk.Status() != StatusDone {
			timedout++
		} else {
			if chk.Result().IsOK {
				passed++
			} else {
				failed++
			}
		}
		logger.Printf("check %s status %d ok: %v", chk.Name(), chk.Status(), chk.Result().IsOK)
	}
	logger.Printf("%v checks done. passed: %v - failed: %v - timedout: %v", len(checks), passed, failed, timedout)
	return passed, failed, timedout
}
