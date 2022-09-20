////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm

package utils

import (
	"github.com/pkg/errors"
	jww "github.com/spf13/jwalterweatherman"
	"syscall/js"
)

var (
	// Error is the Javascript Error object. It used to create new Javascript
	// errors.
	Error = js.Global().Get("Error")

	// JSON is the Javascript JSON object. It is used to perform JSON operations
	// on the Javascript layer.
	JSON = js.Global().Get("JSON")

	// Promise is the Javascript Promise object. It is used to generate new
	// promises.
	Promise = js.Global().Get("Promise")

	// Uint8Array is the Javascript Uint8Array object. It is used to create new
	// Uint8Array.
	Uint8Array = js.Global().Get("Uint8Array")
)

// WrapCB wraps a Javascript function in an object so that it can be called
// later with only the arguments and without specifying the function name.
//
// Panics if m is not a function.
func WrapCB(parent js.Value, m string) func(args ...interface{}) js.Value {
	if parent.Get(m).Type() != js.TypeFunction {
		// Create the error separate from the print so stack trace is printed
		err := errors.Errorf("Function %q is not of type %s", m, js.TypeFunction)
		jww.FATAL.Panicf("%+v", err)
	}

	return func(args ...interface{}) js.Value {
		return parent.Call(m, args...)
	}
}

// PromiseFn converts the Javascript Promise construct into Go.
//
// Call resolve with the return of the function on success. Call reject with an
// error on failure.
type PromiseFn func(resolve, reject func(args ...interface{}) js.Value)

// CreatePromise creates a Javascript promise to return the value of a blocking
// Go function to Javascript.
func CreatePromise(f PromiseFn) interface{} {
	// Create handler for promise (this will be a Javascript function)
	handler := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		// Spawn a new go routine to perform the blocking function
		go func(resolve, reject js.Value) {
			f(resolve.Invoke, reject.Invoke)
		}(args[0], args[1])

		return nil
	})

	// Create and return the Promise object
	return Promise.New(handler)
}