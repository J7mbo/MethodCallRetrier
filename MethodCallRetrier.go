package MethodCallRetrier

import (
	"math/rand"
	"reflect"
	"time"
)

/* Handles the retrying of a function call when an error is given up to X times - useful for http requests. */
type MethodCallRetrier struct {
	/* Wait time in seconds between each unsuccessful call. */
	waitTime int64

	/* Maximum number of retries to attempt before returning an error. */
	maxRetries int64

	/* Useful for incremental increases in the sleep time. Defaults to no exponent. */
	exponent int64

	/* Store the current number of retries; is always reset after ExecuteWithRetry() has finished. */
	currentRetries int64

	/* Store the errors retrieved as they may be different on each subsequent retry. */
	errorList []error
}

/* MethodCallRetrier.New returns a new MethodCallRetrier. */
func New(waitTime int64, maxRetries int64, exponent *int64) *MethodCallRetrier {
	if exponent == nil {
		defaultInt := int64(1)
		exponent = &defaultInt
	}

	if maxRetries <= 0 {
		maxRetries = 0
	}

	return &MethodCallRetrier{waitTime: waitTime, maxRetries: maxRetries, exponent: *exponent}
}

/* Retries the call to object.methodName(...args) with a maximum number of retries and a wait time. */
func (r *MethodCallRetrier) ExecuteWithRetry(
	object interface{}, methodName string, args ...interface{},
) ([]reflect.Value, []error) {
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

	returnValues := r.callMethodOnObject(object, methodName, args)
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

	results := make([]reflect.Value, returnValueCount)

	for i := range results {
		results[i] = returnValues[i]
	}

	return results, nil
}

/* callMethodOnObject calls a method dynamically on an object with arguments. */
func (r *MethodCallRetrier) callMethodOnObject(object interface{}, methodName string, args []interface{}) []reflect.Value {
	var method reflect.Value

	if objectIsAPointer(object) {
		method = reflect.ValueOf(object).MethodByName(methodName)
	} else {
		method = reflect.New(reflect.TypeOf(object)).MethodByName(methodName)
	}

	arguments := make([]reflect.Value, method.Type().NumIn())

	for i := 0; i < method.Type().NumIn(); i++ {
		arguments[i] = reflect.ValueOf(args[i])
	}

	return method.Call(arguments)
}

/* calculateJitter adds randomness to avoid deterministic algorithm causing retry collisions for multiple consumers. */
func calculateJitter(waitTime time.Duration) time.Duration {
	if int64(waitTime) == 0 {
		return waitTime
	}

	jitter := time.Duration(rand.Int63n(int64(waitTime)))

	return waitTime + jitter / 2
}

/* If it's a pointer, we need to call the concrete instead */
func objectIsAPointer(object interface{}) bool {
	return reflect.ValueOf(object).Kind() == reflect.Ptr
}

/* Sleep for the given wait time and increment the retry count by 1. */
func (r *MethodCallRetrier) sleepAndIncrementRetries() {
    time.Sleep(calculateJitter(time.Duration(r.waitTime) * time.Second))

	r.waitTime *= r.exponent

	r.currentRetries++
}

/* Reset the current retries back to zero so that we can re-use this object elsewhere. */
func (r *MethodCallRetrier) resetCurrentRetries() {
	r.currentRetries = 0
}

/* Reset the error list back to zero so that we can re-use this object elsewhere. */
func (r *MethodCallRetrier) resetErrorList() {
	r.errorList = nil
}
