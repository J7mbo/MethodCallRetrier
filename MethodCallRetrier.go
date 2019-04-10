package MethodCallRetrier

import (
	"errors"
	"fmt"
	"math/rand"
	"reflect"
	"time"
)

/* MethodCallRetrier handles the retrying of a function when an error is given up to X times - eg. for http requests. */
type MethodCallRetrier struct {
	/* waitTime is the wait time in seconds between each unsuccessful call. */
	waitTime time.Duration

	/* maxRetries is the maximum number of retries to attempt before returning an error. */
	maxRetries int64

	/* exponent is useful for incremental increases in the sleep time. Defaults to no exponent increase. */
	exponent int64

	/* currentRetries stores the current number of retries; is always reset after ExecuteWithRetry() has finished. */
	currentRetries int64

	/* errorList stores the errors retrieved as they may be different on each subsequent retry. */
	errorList []error
}

/* New returns a new MethodCallRetrier. */
func New(waitTime time.Duration, maxRetries int64, exponent int64) *MethodCallRetrier {
	if exponent < 1 {
		exponent = 1
	}

	if maxRetries < 1 {
		maxRetries = 1
	}

	if waitTime <= 0 {
		waitTime = 0
	}

	return &MethodCallRetrier{waitTime: waitTime, maxRetries: maxRetries, exponent: exponent}
}

/*
ExecuteFuncWithRetry retries a function with a maximum number of retries and a wait time. Functionally equivalent to
ExecuteWithRetry() but accepts a function to maintain type safety in userland instead and removes the requirement of a
user type assertion.
*/
func (r *MethodCallRetrier) ExecuteFuncWithRetry(function func() error) []error {
	defer func() {
		r.resetCurrentRetries()
		r.resetErrorList()
	}()

	if r.currentRetries >= r.maxRetries {
		r.errorList = append(
			r.errorList, &MaxRetriesError{
				methodName: reflect.TypeOf(function).String(),
				waitTime: r.waitTime,
				maxRetries: r.maxRetries,
			},
		)

		return r.errorList
	}

	err := function()

	if err != nil {
		r.errorList = append(r.errorList, err)

		r.sleepAndIncrementRetries()

		return r.ExecuteFuncWithRetry(function)
	}

	return r.errorList
}

/* ExecuteWithRetry retries the call to object.methodName(...args) with a maximum number of retries and a wait time. */
func (r *MethodCallRetrier) ExecuteWithRetry(
	object interface{}, methodName string, args ...interface{},
) ([]interface{}, []error) {
	defer func() {
		r.resetCurrentRetries()
		r.resetErrorList()
	}()

	if r.currentRetries >= r.maxRetries {
		r.errorList = append(
			r.errorList, &MaxRetriesError{methodName: methodName, waitTime: r.waitTime, maxRetries: r.maxRetries},
		)

		return nil, r.errorList
	}

	returnValues, err := r.callMethodOnObject(object, methodName, args)

	if err != nil {
		return nil, []error{err}
	}

	returnValueCount := len(returnValues)

	errorFound := false

	for i := 0; i < returnValueCount; i++ {
		if err, ok := returnValues[i].Interface().(error); ok && err != nil {
			r.errorList = append(r.errorList, err)

			errorFound = true
		}
	}

	if errorFound == true {
		r.sleepAndIncrementRetries()

		return r.ExecuteWithRetry(object, methodName, args...)
	}

	results := make([]interface{}, returnValueCount)

	for i := range results {
		/* Convert from reflect.Value to a magical anything. */
		results[i] = returnValues[i].Interface()
	}

	return results, nil
}

/* callMethodOnObject calls a method dynamically on an object with arguments. */
func (r *MethodCallRetrier) callMethodOnObject(
	object interface{},
	methodName string,
	args []interface{},
) ([]reflect.Value, error) {
	var method reflect.Value

	if objectIsAPointer(object) {
		method = reflect.ValueOf(object).MethodByName(methodName)
	} else {
		method = reflect.New(reflect.TypeOf(object)).MethodByName(methodName)
	}

	if !method.IsValid() {
		return nil, errors.New(
			fmt.Sprintf("method with name: '%s' does not exist on object: '%T'", methodName, object),
		)
	}

	arguments := make([]reflect.Value, method.Type().NumIn())

	for i := 0; i < method.Type().NumIn(); i++ {
		arguments[i] = reflect.ValueOf(args[i])
	}

	return method.Call(arguments), nil
}

/* calculateJitter adds randomness to avoid deterministic algorithm causing retry collisions for multiple consumers. */
func calculateJitter(waitTime time.Duration) time.Duration {
	if int64(waitTime) == 0 {
		return waitTime
	}

	jitter := time.Duration(rand.Int63n(int64(waitTime)))

	return waitTime + jitter / 2
}

/* objectIsAPointer decides whether or not an object is a pointer and so we would need to call the concrete instead. */
func objectIsAPointer(object interface{}) bool {
	return reflect.ValueOf(object).Kind() == reflect.Ptr
}

/* sleepAndIncrementRetries sleeps for the given wait time and increments the retry count by 1. */
func (r *MethodCallRetrier) sleepAndIncrementRetries() {
    time.Sleep(calculateJitter(time.Duration(r.waitTime) * time.Second))

	r.waitTime = time.Duration(int64(r.waitTime) * r.exponent)

	r.currentRetries++
}

/* resetCurrentRetries resets the current retries back to zero so that we can re-use this object elsewhere. */
func (r *MethodCallRetrier) resetCurrentRetries() {
	r.currentRetries = 0
}

/* resetErrorList resets the error list back to zero so that we can re-use this object elsewhere. */
func (r *MethodCallRetrier) resetErrorList() {
	r.errorList = nil
}
