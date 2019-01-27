Features
-

Retry your method calls automatically when an `error` is returned up to a specified number of times, with a given wait time.

Extremely useful for retrying HTTP calls in distributed systems, or anything else over the network that can error sporadically.

Installation
-

`go get github.com/j7mbo/MethodCallRetrier`

Usage
-

Initialise the object with some options:

```
MethodCallRetrier.New(waitTime int64, maxRetries int64, exponent *int64, onError *func(err error, retryCount int64) 
```

Call `ExecuteWithRetry` with your object and method you want to retry:

```
ExecuteWithRetry(
	object interface{}, methodName string, args ...interface{},
) ([]reflect.Value, []error) {
```

You can use it as follows:

```
results, errs := retrier.ExecuteWithRetry(yourObject, "MethodToCall", "Arg1", "Arg2", "etc")
```

The results are an array of `reflect.Value` objects, (used for the dynamic method call), and an array of all errors.
To use the results, you must typecast the result to the expected type. In the case of an `int64`, for example:

```
myInt := results[0].Interface().(int64)
```