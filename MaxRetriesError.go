package MethodCallRetrier

import "fmt"

/* Custom error for differentiation. */
type MaxRetriesError struct {
	methodName string
	waitTime   int64
	maxRetries int64
}

/* Create a new instance of MaxRetriesError. */
func (e *MaxRetriesError) Error() string {
	return fmt.Sprintf(
		"Tried calling: '%s' %d times but reached max retries of: %d", e.methodName, e.waitTime, e.maxRetries,
	)
}
