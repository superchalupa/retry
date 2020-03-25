// Copyright @2018 Saddam Hossain.  All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Package retry is a simple and easy retry mechanism package for Go
package retry

import (
	"errors"
	"math/rand"
	"reflect"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// DoFuncBAckoff runs provided func until it returns nil error or attempts exhausted, running the provided backoff strategy function between successive attempts
func DoFuncBackoff(attempt uint, sleep time.Duration, backoffStrategy func(time.Duration) time.Duration, fn func() error) (err error) {
	for {
		err = fn(); 
    if attempt--; attempt == 0 || err == nil{
      break
    }
    sleep = backoffStrategy(sleep)
	}
	return err
}

// DoFunc try to execute the function, it only expect that the function will return an error only
// default backoff strategy is exponential 
func DoFunc(attempt uint, sleep time.Duration, fn func() error) (err error) {
  backoff := func(sleep time.Duration) time.Duration {
		// Add jitter to prevent Thundering Herd problem (https://en.wikipedia.org/wiki/Thundering_herd_problem)
		sleep += (time.Duration(rand.Int63n(int64(sleep)))) / 2
		time.Sleep(sleep)
		// multiplicative for next loop
		return 2*sleep
  }
  
  return DoFuncBackoff(attempt, sleep, backoff, fn)
}

// Do try to execute the function by its value, function can take variadic arguments and return multiple return.
// You must put error as the last return value so that DoFunc can take decision that the call failed or not
func Do(attempt uint, sleep time.Duration, fn interface{}, args ...interface{}) ([]interface{}, error) {

	if attempt == 0 {
		return nil, errors.New("retry: attempt should be greater than 0")
	}

	vfn := reflect.ValueOf(fn)

	// if the fn is not a function then return error
	if vfn.Type().Kind() != reflect.Func {
		return nil, errors.New("retry: fn is not a function")
	}

	// if the functions in not variadic then return the argument missmatch error
	if !vfn.Type().IsVariadic() && (vfn.Type().NumIn() != len(args)) {
		return nil, errors.New("retry: fn argument mismatch")
	}

	// if the function does not return anything, we can't catch if an error occur or not
	if vfn.Type().NumOut() <= 0 {
		return nil, errors.New("retry: fn return's can not empty, at least an error")
	}

	// build args for reflect value Call
	in := make([]reflect.Value, len(args))
	for k, a := range args {
		in[k] = reflect.ValueOf(a)
	}

	var lastErr error
	for attempt > 0 {
		// call the fn with arguments
		out := []interface{}{}
		for _, o := range vfn.Call(in) {
			out = append(out, o.Interface())
		}

		// if the last value is not error then return an error
		err, ok := out[len(out)-1].(error)
		if !ok && out[len(out)-1] != nil {
			return nil, errors.New("retry: fn return's right most value must be an error")
		}

		if err == nil {
			return out[:len(out)-1], nil
		}
		lastErr = err
		attempt--
		// Add jitter to prevent Thundering Herd problem (https://en.wikipedia.org/wiki/Thundering_herd_problem)
		sleep += (time.Duration(rand.Int63n(int64(sleep)))) / 2
		time.Sleep(sleep)
		sleep *= 2
	}

	return nil, lastErr
}
