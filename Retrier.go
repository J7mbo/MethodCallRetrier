package MethodCallRetrier

/* Represents an object capable of retrying a call on an object X times after receiving an error. */
type Retrier interface {
	ExecuteWithRetry(
		object interface{}, methodName string, args ...interface{},
	) (results []interface{}, errs []error, wasSuccesful bool)

	ExecuteFuncWithRetry(function func() error) (errs []error, wasSuccessful bool)
}
