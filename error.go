// Copyright 2019 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package emacs

// #include "emacs-module.h"
// #include "wrappers.h"
import "C"

import (
	"fmt"
	"strings"
)

// Error is an error that causes to signal an error with the given symbol and
// data.  Signaling an error will evaluate symbol and data lazily.  The
// evaluation is best-effort since it can itself fail.  If it fails the symbol
// and/or data might be lost, but Emacs will signal some error in any case.  If
// you already have an evaluated symbol and data value, use Signal instead.
// ErrorSymbol.Error is a convenience factory function for Error values.
type Error struct {
	// The error symbol.
	Symbol ErrorSymbol

	// The unevaluated error data.
	Data List
}

// Error implements the error interface.  It does a best-effort attempt to
// mimic the Emacs error‑message‑string function.
func (x Error) Error() string {
	parts := make([]string, len(x.Data))
	for i, d := range x.Data {
		parts[i] = fmt.Sprint(d)
	}
	return fmt.Sprintf("%s: %s", x.Symbol, strings.Join(parts, ", "))
}

// ErrorSymbol represents an error symbol.  Use DefineError to create
// ErrorSymbol values.  The zero ErrorSymbol is not a valid error symbol.
type ErrorSymbol struct {
	name    Name
	message string
}

// DefineError arranges for an error symbol to be defined using the Emacs
// function define‑error.  Call it from an init function (i.e., before loading
// the dynamic module into Emacs) to define additional error symbols for your
// module.  DefineError panics if name or message is empty, or if name is
// duplicate.
func DefineError(name Name, message string, parents ...ErrorSymbol) ErrorSymbol {
	if message == "" {
		panic(fmt.Errorf("empty error message for error symbol %s", name))
	}
	errorSymbols.MustEnqueue(name, errorSymbol{name, message, parents})
	return ErrorSymbol{name, message}
}

// DefineError is like the global DefineError function, except that it requires
// a live environment, defines the error symbol immediately, and returns errors
// instead of panicking.
func (e Env) DefineError(name Name, message string, parents ...ErrorSymbol) (ErrorSymbol, error) {
	if message == "" {
		return ErrorSymbol{}, fmt.Errorf("empty error message for error symbol %s", name)
	}
	s := errorSymbol{name, message, parents}
	if err := errorSymbols.RegisterAndDefine(e, name, s); err != nil {
		return ErrorSymbol{}, err
	}
	return ErrorSymbol{name, message}, nil
}

// String returns the message of the error symbol.
func (s ErrorSymbol) String() string {
	return s.message
}

// Error returns an error that causes to signal an error with the given symbol
// and data.  The return value is of type Error.  Signaling an error will
// evaluate symbol and data lazily.  The evaluation is best-effort since it can
// itself fail.  If it fails the symbol and/or data might be lost, but Emacs
// will signal some error in any case.  If you already have an evaluated symbol
// and data value, use Signal instead.
func (s ErrorSymbol) Error(data ...In) error {
	return Error{s, data}
}

// Emacs interns the error symbol in the default obarray and returns the symbol
// value.
func (s ErrorSymbol) Emacs(e Env) (Value, error) {
	return s.name.Emacs(e)
}

func (s ErrorSymbol) match(e Env, err error) bool {
	switch x := err.(type) {
	case Signal:
		want, err := s.Emacs(e)
		// We treat an error here as silent failure to not require
		// clutter at the call sites.
		return err == nil && e.Eq(x.Symbol, want)
	case Error:
		return s == x.Symbol
	default:
		return false
	}
}

// Signal is an error that that causes Emacs to signal an error with the given
// symbol and data.  This is the equivalent to Error if you already have an
// evaluated symbol and data value.
type Signal struct{ Symbol, Data Value }

// Error implements the error interface.  Error returns a static string.  Use
// Message to return the actual Emacs error message.
func (Signal) Error() string {
	return "Emacs signal"
}

// Message returns the Emacs error message for this signal.  It returns <error>
// if determining the message failed.
func (s Signal) Message(e Env) string {
	var r String
	if err := e.CallOut("error-message-string", &r, Cons{s.Symbol, s.Data}); err != nil {
		return "<error>"
	}
	return string(r)
}

// Message returns an error message for err.  If err is an Emacs error, it uses
// error-message-string to obtain the Emacs error message.  Otherwise, it
// returns err.Error().
func (e Env) Message(err error) string {
	if s, ok := err.(Signal); ok {
		return s.Message(e)
	}
	return err.Error()
}

// Throw is an error that triggers the Emacs throw function.
type Throw struct{ Tag, Value Value }

// Error implements the error interface.
func (Throw) Error() string {
	return "Emacs throw"
}

// WrongTypeArgument returns an error that will cause Emacs to signal an error
// of type wrong‑type‑argument.  Use this if some argument has an unexpected
// type.  pred should be a predicate-like symbol such as natnump.  arg is the
// argument whose type is invalid.
func WrongTypeArgument(pred Symbol, arg In) error {
	return wrongTypeArgument.Error(pred, arg)
}

// IsWrongTypeArgument returns whether err is an Emacs signal of type
// wrong-type-argument.  This function detects both Error and Signal.
func (e Env) IsWrongTypeArgument(err error) bool {
	return wrongTypeArgument.match(e, err)
}

var wrongTypeArgument = ErrorSymbol{"wrong-type-argument", "Wrong type argument"}

// OverflowError returns an error that will cause Emacs to signal an error of
// type overflow‑error.  Use this if you notice that an integer doesn’t fit
// into its target type.  val is the string representation of the overflowing
// value.  val is a string instead of an In to avoid further overflows when
// converting it to an Emacs value.
func OverflowError(val string) error {
	return overflowError.Error(String(val))
}

// IsOverflowError returns whether err is an Emacs signal of type
// overflow-error.  This function detects both Error and Signal.
func (e Env) IsOverflowError(err error) bool {
	return overflowError.match(e, err)
}

var overflowError = ErrorSymbol{"overflow-error", "Arithmetic overflow error"}

type nonlocalExit interface {
	// signal returns a C representation of this nonlocal exit.
	signal(Env) C.struct_result_base_with_optional_error_info
}

func (x Error) signal(e Env) C.struct_result_base_with_optional_error_info {
	// If we can’t create the error symbol or data, report this fact back
	// by setting has_error_info to false.  handle_nonlocal_exit detects
	// this case and attempts to fill in generic error information.  This
	// approach prevents infinite recursion if there’s an error during
	// error handling.
	symbol, err := x.Symbol.Emacs(e)
	if err != nil {
		return C.struct_result_base_with_optional_error_info{C.emacs_funcall_exit_signal, false, nil, nil}
	}
	data, err := x.Data.Emacs(e)
	if err != nil {
		return C.struct_result_base_with_optional_error_info{C.emacs_funcall_exit_signal, false, nil, nil}
	}
	return Signal{symbol, data}.signal(e)
}

func (s Signal) signal(e Env) C.struct_result_base_with_optional_error_info {
	return C.struct_result_base_with_optional_error_info{C.emacs_funcall_exit_signal, true, s.Symbol.r, s.Data.r}
}

func (t Throw) signal(e Env) C.struct_result_base_with_optional_error_info {
	return C.struct_result_base_with_optional_error_info{C.emacs_funcall_exit_throw, true, t.Tag.r, t.Value.r}
}

// signal returns a C representation of err.
func (e Env) signal(err error) C.struct_result_base_with_optional_error_info {
	if err == nil {
		return C.struct_result_base_with_optional_error_info{C.emacs_funcall_exit_return, false, nil, nil}
	}
	if n, ok := err.(nonlocalExit); ok {
		return n.signal(e)
	}
	return Error{baseError, List{String(err.Error())}}.signal(e)
}

// check converts a pending nonlocal exit to a Go error.  If no nonlocal exit
// is set in r, check returns nil.  If a signal is set in r, check returns an
// error of dynamic type Signal.  If a throw is set in r, check returns an
// error of dynamic type Throw.
func (e Env) check(r C.struct_result_base) error {
	switch r.exit {
	case C.emacs_funcall_exit_return:
		return nil
	case C.emacs_funcall_exit_signal:
		return Signal{Value{r.error_symbol}, Value{r.error_data}}
	case C.emacs_funcall_exit_throw:
		return Throw{Value{r.error_symbol}, Value{r.error_data}}
	default:
		// This cannot really happen, but better safe than sorry.
		return WrongTypeArgument("module-funcall-exit-p", Int(r.exit))
	}
}

// checkVoid is like check, but takes a struct void_result for convenience.
func (e Env) checkVoid(r C.struct_void_result) error {
	return e.check(r.base)
}

// checkValue is like check, but takes a struct value_result and returns v for
// convenience.
func (e Env) checkValue(r C.struct_value_result) (Value, error) {
	return Value{r.value}, e.check(r.base)
}

var (
	baseError          = DefineError("go-error", "Generic Go error")
	unimplementedError = DefineError("go-unimplemented-error", "Unimplemented Go function", baseError)
)

type errorSymbol struct {
	name    Name
	message string
	parents []ErrorSymbol
}

func (s errorSymbol) Define(e Env) error {
	parents := make(List, len(s.parents))
	for i, p := range s.parents {
		parents[i] = p
	}
	_, err := e.Call("define-error", s.name, String(s.message), parents)
	return err
}

var errorSymbols = NewManager(RequireName | RequireUniqueName | DefineOnInit)
