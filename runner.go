package chkok

import (
	"log"
	"sync"
	"time"
)

// Runner runs all the checks logging details
type Runner struct {
	Log *log.Logger
}

// RunChecks runs all the checks in the suite and returns slice of check
func (r *Runner) RunChecks(suites CheckSuites, timeout time.Duration) []Check {
	var now, deadline time.Time
	deadline = time.Now().Add(timeout)
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
		// TODO: find a way to consider global timeout to the check,
		// maybe set timeout to max of remaining global time and per test timeout
		if now.Before(deadline) {
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
