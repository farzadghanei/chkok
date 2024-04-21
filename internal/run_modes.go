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
func RunModeCLI(checkGroups *CheckSuites, conf *Conf, output io.Writer, logger *log.Logger) int {
	runner := Runner{Log: logger, Timeout: conf.Runners["default"].Timeout}
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
func RunModeHTTP(checkGroups *CheckSuites, conf *Conf, logger *log.Logger) int {
	timeout := conf.Runners["default"].Timeout
	shutdownAfterRequests := conf.Runners["default"].ShutdownAfterRequests

	// override default runner config if with http runner config if provided
	if httpRunnerConf, ok := conf.Runners["http"]; ok {
		if httpRunnerConf.Timeout > 0 {
			timeout = httpRunnerConf.Timeout
		}
		if httpRunnerConf.ShutdownAfterRequests > 0 {
			shutdownAfterRequests = httpRunnerConf.ShutdownAfterRequests
		}
	}
	runner := Runner{Log: logger, Timeout: timeout}

	var reqHandlerChan = make(chan *http.Request, 1)

	httpHandler := func(w http.ResponseWriter, r *http.Request) {
		// TODO: custmize return codes and messages from configuration
		logger.Printf("processing http request: %s", httpRequestAsString(r))
		_, failed, timedout := runChecks(&runner, checkGroups, logger)
		if timedout > 0 {
			w.WriteHeader(http.StatusGatewayTimeout) // 504
			fmt.Fprintf(w, "TIMEDOUT")
		} else if failed > 0 {
			w.WriteHeader(http.StatusInternalServerError) // 500
			fmt.Fprintf(w, "FAILED")
		} else {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, "OK")
		}
		reqHandlerChan <- r
	}

	http.HandleFunc("/", httpHandler)
	// TODO: allow to set server timeouts from configuration
	server := &http.Server{
		Addr:         ":8080",
		Handler:      nil, // use http.DefaultServeMux
		ReadTimeout:  2 * time.Second,
		WriteTimeout: 5 * time.Second,
		IdleTimeout:  2 * time.Second,
	}

	var count uint32 = 0

	go func() {
		var request *http.Request
		timeoutCtx, cancel := context.WithTimeout(context.Background(), ShutdownTimeout)
		defer cancel()
		for request = range reqHandlerChan {
			atomic.AddUint32(&count, 1)
			logger.Printf("request [%v] is processed: %v", count, httpRequestAsString(request))
			if shutdownAfterRequests > 0 && atomic.LoadUint32(&count) >= shutdownAfterRequests {
				if err := server.Shutdown(timeoutCtx); err != nil {
					logger.Printf("http server shutdown failed: %v", err)
				}
				return
			}
		}
	}()

	logger.Printf("starting http server ...")
	err := server.ListenAndServe()
	close(reqHandlerChan)
	if err != nil {
		if atomic.LoadUint32(&count) < 1 { // server didn't handle any requests
			logger.Printf("http server failed to start: %v", err)
			return ExSoftware
		}
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
