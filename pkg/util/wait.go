package util

import (
	"fmt"
	"time"
)

// WaitFor() waits until f() returns false or err != nil
// f() returns <wait as bool, or err>.
func WaitFor(timeout time.Duration, period time.Duration, f func() (bool, error)) error {
	timeoutChan := time.After(timeout)
	var (
		wait bool = true
		err  error
	)
	for wait {
		select {
		case <-timeoutChan:
			return fmt.Errorf("Timed out")
		default:
			wait, err = f()
			if err != nil {
				return err
			}
			time.Sleep(period)
		}
	}

	return nil
}
