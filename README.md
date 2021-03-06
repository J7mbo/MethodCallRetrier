MethodCallRetrier
-

[![Build Status](https://travis-ci.org/J7mbo/MethodCallRetrier.svg?branch=master)](https://travis-ci.org/J7mbo/MethodCallRetrier)
[![GoDoc](https://godoc.org/github.com/J7mbo/MethodCallRetrier?status.svg)](https://godoc.org/github.com/J7mbo/MethodCallRetrier)
[![codecov](https://img.shields.io/codecov/c/github/j7mbo/MethodCallRetrier.svg)](https://codecov.io/gh/J7mbo/MethodCallRetrier)

Features
-

Retry your method calls automatically when an `error` is returned up to a specified number of times, with a given wait time.

Extremely useful for retrying HTTP calls in distributed systems, or anything else over the network that can error sporadically.

An example of using this library for everything from retrying connections to redis, to logging with elasticsearch, can
be found [here](https://github.com/J7mbo/palmago-streetview).

Installation
-
```bash
go get github.com/j7mbo/MethodCallRetrier/v2
```

Usage
-

Initialise the object with some options:

```go
MethodCallRetrier.New(waitTime time.Duration, maxRetries int64, exponent int64) 
```

Call `ExecuteWithRetry` with your object and method you want to retry:

```go
ExecuteWithRetry(
	object interface{}, methodName string, args ...interface{},
) (results []interface{}, errs []error, wasSuccessful bool)
```

Alternatively, call `ExecuteFuncWithRetry` and pass in a function that returns `error` to retry.

```go
ExecuteFuncWithRetry(func() error) (errs []error, wasSuccessful bool)
```

You can use it as follows:

```go
results, errs, wasSuccessful := retrier.ExecuteWithRetry(yourObject, "MethodToCall", "Arg1", "Arg2", "etc")
```

The results are an array of `interface{}` objects, (used for the dynamic method call), and an array of all errors.
To use the results, you must typecast the result to the expected type. In the case of an `int64`, for example:

```go
myInt := results[0].(int64)
```

Or, to maintain type safety in userland, you can pass a function in instead and do all your retriable work in there:

```go
var json string

functionToRetry := func() error {
    json, err = DoSomeFlakeyHttpCall()
    
    if err != nil {
        return err
    }
    
    return nil
}

if wasSuccessful, errs := retrier.ExecuteFuncWithRetry(funcToRetry); !wasSuccesful {
    /* Do something with errs because we failed 3 times */
    return
}

fmt.Println(json)
```

Be aware that this choice now requires you to work within the function scope, and that if you want to use the results of
the call outside of the scope of the function then you must declare it using `var result type` as shown above; but this
does remove the requirement of a type assertion that you have to do with `ExecuteWithRetry`.
