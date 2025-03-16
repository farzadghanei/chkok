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

	var reqHandlerChan = make(chan *http.Request, 1)

	httpHandler := makeHTTPRequestHandler(reqHandlerChan, conf, checkGroups, logger)
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

// makeHTTPRequestHandler creates a http request handler function used by RunModeHTTP
func makeHTTPRequestHandler(reqHandlerChan chan *http.Request,
	conf *ConfRunner, checkGroups *CheckSuites, logger *log.Logger) func(http.ResponseWriter, *http.Request) {
	maxConcurrentRequests := int32(*conf.MaxConcurrentRequests) //nolint: gosec
	responseOK, responseFailed := *conf.ResponseOK, *conf.ResponseFailed
	responseTimeout := *conf.ResponseTimeout
	responseUnavailable, responseInvalidRequest := *conf.ResponseUnavailable, *conf.ResponseInvalidRequest
	requieredHeaders := conf.RequestRequiredHeaders
	shouldCheckHeaders := len(requieredHeaders) > 0

	runner := Runner{Log: logger, Timeout: *conf.Timeout}

	var runningRequests atomic.Int32
	httpRequestHandler := func(w http.ResponseWriter, r *http.Request) {
		runningRequests.Add(1)
		if maxConcurrentRequests > 0 && runningRequests.Load() > maxConcurrentRequests {
			logger.Printf("runner reached max conccurent requests. rejecting request: %s", httpRequestAsString(r))
			w.WriteHeader(http.StatusServiceUnavailable) // 503
			fmt.Fprint(w, responseUnavailable)
			return
		}
		defer runningRequests.Add(-1)
		if shouldCheckHeaders {
			for header, value := range requieredHeaders {
				reqHeader, ok := r.Header[header]
				if !ok {
					logger.Printf("http request missing required header %s: %s", header, httpRequestAsString(r))
					w.WriteHeader(http.StatusBadRequest) // 400
					fmt.Print(w, responseInvalidRequest)
					return
				}
				if value != "" && reqHeader[0] != value {
					logger.Printf("http request doesn't match required header %s: %s", header, httpRequestAsString(r))
					w.WriteHeader(http.StatusBadRequest) // 400
					fmt.Print(w, responseInvalidRequest)
					return
				}
			}
		}

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
	return httpRequestHandler
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
