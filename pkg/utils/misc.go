package utils

import (
	"sync"
	"time"
)

// https://stackoverflow.com/a/30716481
func Ptr[T any](v T) *T {
	return &v
}

// WaitTimeout waits for the waitgroup for the specified max timeout.
// Returns true if waiting timed out.
// https://stackoverflow.com/a/32843750
func WaitTimeout(wg *sync.WaitGroup, timeout time.Duration) bool {
	c := make(chan struct{})
	go func() {
		defer close(c)
		wg.Wait()
	}()
	select {
	case <-c:
		return false // completed normally
	case <-time.After(timeout):
		return true // timed out
	}
}
