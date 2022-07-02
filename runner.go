package chkok

import (
	"sync"
	"time"

	"github.com/panjf2000/ants/v2"
)

// RunChecks runs all the checks in the suite and returns slice of check
func RunChecks(suites CheckSuites, poolSize int, timeout time.Duration) []Check {
	var now, deadline time.Time
	deadline = time.Now().Add(timeout)
	var checks []Check
	var wg sync.WaitGroup
	// TODO: implement check groups sequencial runs, wait for each item
	// in a group before submitting the next one
	for _, groupChecks := range suites {
		checks = append(checks, groupChecks...)
	}

	pool, _ := ants.NewPool(poolSize)
	for _, chk := range checks {
		now = time.Now()
		// TODO: find a way to consider global timeout to the check,
		// maybe set timeout to max of remaining global time and per test timeout
		if now.After(deadline) {
			break
		}
		wg.Add(1)
		task := func() {
			defer wg.Done()
			chk.Run()
		}
		if err := pool.Submit(task); err != nil {
			panic(err)
		}
	}
	wg.Wait()
	return checks
}
