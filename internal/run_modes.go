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
	// DefaultListenAddress is the default http server listen address
	DefaultListenAddress        = ":8080"
	DefaultRequestReadTimeout   = 5 * time.Second
	DefaultResponseWriteTimeout = 5 * time.Second
)

// RunModeCLI run app in CLI mode using the provided configs, return exit code
func RunModeCLI(checkGroups *CheckSuites, conf *ConfRunner, output io.Writer, logger *log.Logger) int {
	runner := Runner{Log: logger, Timeout: conf.Timeout}
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
	timeout := conf.Timeout
	shutdownSignalHeaderValue := ""
	if conf.ShutdownSignalHeader != nil {
		shutdownSignalHeaderValue = *conf.ShutdownSignalHeader
	}
	listenAddress := conf.ListenAddress
	requestReadTimeout := conf.RequestReadTimeout
	responseWriteTimeout := conf.ResponseWriteTimeout

	if listenAddress == "" {
		logger.Printf("no http listen address provided, using default: %s", DefaultListenAddress)
		listenAddress = DefaultListenAddress
	}
	if requestReadTimeout == 0 {
		logger.Printf("no http request read timeout provided, using default: %s", DefaultRequestReadTimeout)
		requestReadTimeout = DefaultRequestReadTimeout
	}
	if responseWriteTimeout == 0 {
		logger.Printf("no http response write timeout provided, using default: %s", DefaultResponseWriteTimeout)
		responseWriteTimeout = DefaultResponseWriteTimeout
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
		Addr:         listenAddress,
		Handler:      nil, // use http.DefaultServeMux
		ReadTimeout:  requestReadTimeout,
		WriteTimeout: responseWriteTimeout,
		IdleTimeout:  0 * time.Second, // set to 0 so uses read timeout
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
				}
				return
			}
		}
	}()

	logger.Printf("starting http server ...")
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
