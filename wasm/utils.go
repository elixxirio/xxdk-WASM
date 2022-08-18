////////////////////////////////////////////////////////////////////////////////
// Copyright © 2020 xx network SEZC                                           //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file                                                               //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm

package wasm

import (
	"encoding/json"
	"syscall/js"
)

// CopyBytesToGo copies the Uint8Array stored in the js.Value to []byte. This is
// a wrapper for js.CopyBytesToGo to make it more convenient.
func CopyBytesToGo(src js.Value) []byte {
	b := make([]byte, src.Length())
	js.CopyBytesToGo(b, src)
	return b
}

// CopyBytesToJS copies the []byte to a Uint8Array stored in a js.Value. This is
// a wrapper for js.CopyBytesToJS to make it more convenient.
func CopyBytesToJS(src []byte) js.Value {
	dst := js.Global().Get("Uint8Array").New(len(src))
	js.CopyBytesToJS(dst, src)
	return dst
}

// WrapCB wraps a js.Call so that it can be called later with only the arguments
// and without specifying the function name
func WrapCB(call func(m string, args ...interface{}) js.Value, m string) func(args ...interface{}) js.Value {
	return func(args ...interface{}) js.Value {
		return call(m, args)
	}
}

// JsonToJS converts a marshalled JSON bytes to a Javascript object.
func JsonToJS(src []byte) js.Value {
	var inInterface map[string]interface{}
	err := json.Unmarshal(src, &inInterface)
	if err != nil {
		Throw(TypeError, err.Error())
		return js.ValueOf(nil)
	}

	return js.ValueOf(inInterface)
}

// Throw function stub to throws Javascript exceptions. The exception must be
// one of the defined Exception below. Any other error types will result in an
// error.
func Throw(exception Exception, message string)

// Exception are the possible Javascript error types that can be thrown.
type Exception string

const (
	// EvalError occurs when error has occurred in the eval() function.
	//
	// Deprecated: This exception is not thrown by JavaScript anymore, however
	// the EvalError object remains for compatibility.
	EvalError Exception = "EvalError"

	// RangeError occurs when a numeric variable or parameter is outside its
	// valid range.
	RangeError Exception = "RangeError"

	// ReferenceError occurs when a variable that does not exist (or hasn't yet
	// been initialized) in the current scope is referenced.
	ReferenceError Exception = "ReferenceError"

	// SyntaxError occurs when trying to interpret syntactically invalid code.
	SyntaxError Exception = "SyntaxError"

	// TypeError occurs when an operation could not be performed, typically (but
	// not exclusively) when a value is not of the expected type.
	//
	// A TypeError may be thrown when:
	//
	//  - an operand or argument passed to a function is incompatible with the
	//    type expected by that operator or function; or
	//  - when attempting to modify a value that cannot be changed; or
	//  - when attempting to use a value in an inappropriate way.
	TypeError Exception = "TypeError"

	// URIError occurs when a global URI handling function was used in a wrong
	// way.
	URIError Exception = "URIError"
)
