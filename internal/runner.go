package chkok

import (
	"log"
	"sync"
	"time"
)

// Runner runs all the checks logging details
type Runner struct {
	Log     *log.Logger
	Timeout time.Duration
}

// RunChecks runs all the checks in the suite and returns slice of check
func (r *Runner) RunChecks(suites CheckSuites) []Check {
	var now, deadline time.Time
	deadline = time.Now().Add(r.Timeout)
	var checks []Check
	var wg sync.WaitGroup
	// TODO: implement check groups sequencial runs, wait for each item
	// in a group before submitting the next one
	for _, groupChecks := range suites {
		checks = append(checks, groupChecks...)
	}
	r.Log.Printf("going to run %d checks", len(checks))
	var result []Check
	for index := range checks {
		chk := checks[index]
		result = append(result, chk)
		now = time.Now()
		if now.Before(deadline) {
			// adjust timeout for timed checks based on remaining timeout of the runner
			remaining := deadline.Sub(now)
			if timedCheck, ok := chk.(TimedCheck); ok {
				if timedCheck.GetTimeout() > remaining {
					timedCheck.SetTimeout(remaining)
				}
			}
			// schedule the check to run
			wg.Add(1)
			go func() {
				defer wg.Done()
				chk.Run()
			}()
		}
	}
	wg.Wait()
	return result
}
