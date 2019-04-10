package MethodCallRetrier

import (
	"fmt"
	"time"
)

/* MaxRetriesError is a custom error for differentiation. */
type MaxRetriesError struct {
	methodName string
	waitTime   time.Duration
	maxRetries int64
}

/* Error is for implementing the error interface. */
func (e *MaxRetriesError) Error() string {
	return fmt.Sprintf(
		"Tried calling: '%s' %d times but reached max retries of: %d", e.methodName, e.waitTime, e.maxRetries,
	)
}
