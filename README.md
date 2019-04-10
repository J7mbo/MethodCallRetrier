MethodCallRetrier
-

[![Build Status](https://travis-ci.org/J7mbo/MethodCallRetrier.svg?branch=master)](https://travis-ci.org/J7mbo/MethodCallRetrier)
[![GoDoc](https://godoc.org/github.com/J7mbo/MethodCallRetrier?status.svg)](https://godoc.org/github.com/J7mbo/MethodCallRetrier)
[![codecov](https://img.shields.io/codecov/c/github/j7mbo/MethodCallRetrier.svg)](https://codecov.io/gh/J7mbo/MethodCallRetrier)

Features
-

Retry your method calls automatically when an `error` is returned up to a specified number of times, with a given wait time.

Extremely useful for retrying HTTP calls in distributed systems, or anything else over the network that can error sporadically.

Installation
-

`go get github.com/j7mbo/MethodCallRetrier/v2`

Usage
-

Initialise the object with some options:

```
MethodCallRetrier.New(waitTime int64, maxRetries int64, exponent int64) 
```

Call `ExecuteWithRetry` with your object and method you want to retry:

```
ExecuteWithRetry(
	object interface{}, methodName string, args ...interface{},
) ([]reflect.Value, []error)
```

Alternatively, call `ExecuteFuncWithRetry` and pass in a function that returns `error` to retry.

```
ExecuteFuncWithRetry(func() error) []error
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

Or, to maintain type safety in userland, you can pass a function in instead and do all your retriable work in there:

```
var json string

functionToRetry := func() error {
    json, err = DoSomeFlakeyHttpCall()
    
    if err != nil {
        return err
    }
    
    return nil
}

if errs := retrier.ExecuteFuncWithRetry(funcToRetry); len(errs) > 0 {
    /* Do something because we failed 3 times */
    return
}

fmt.Println(json)
```

Be aware that this choice now requires you to work within the function scope, and that if you want to use the results of
the call outside of the scope of the function then you must declare it using `var result type` as shown above; but this
does remove the requirement of a type assertion that you have to do with `ExecuteWithRetry`.
