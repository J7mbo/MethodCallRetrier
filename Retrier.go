package MethodCallRetrier

import "reflect"

/* Represents an object capable of retrying a call on an object X times after receiving an error. */
type Retrier interface {
	ExecuteWithRetry(
		maxRetries int64, waitTime int64, object interface{}, methodName string, args ...interface{},
	) ([]reflect.Value, error)
}
