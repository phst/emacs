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

// #include <emacs-module.h>
// void non_local_exit_clear(emacs_env *env) {
//   env->non_local_exit_clear(env);
// }
// enum emacs_funcall_exit non_local_exit_get(emacs_env *env, emacs_value *symbol, emacs_value *data) {
//   return env->non_local_exit_get(env, symbol, data);
// }
// void non_local_exit_signal(emacs_env *env, emacs_value symbol, emacs_value data) {
//   env->non_local_exit_signal(env, symbol, data);
// }
// void non_local_exit_throw(emacs_env *env, emacs_value tag, emacs_value value) {
//   env->non_local_exit_throw(env, tag, value);
// }
import "C"

import (
	"fmt"
	"strings"
	"sync"
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
	errorSymbols.mustRegister(errorSymbol{name, message, parents})
	return ErrorSymbol{name, message}
}

// DefineError is like the global DefineError function, except that it requires
// a live environment, defines the error symbol immediately, and returns errors
// instead of panicking.
func (e Env) DefineError(name Name, message string, parents ...ErrorSymbol) (ErrorSymbol, error) {
	s := errorSymbol{name, message, parents}
	if err := errorSymbols.register(s); err != nil {
		return ErrorSymbol{}, err
	}
	if err := s.define(e); err != nil {
		return ErrorSymbol{}, err
	}
	return ErrorSymbol{name, message}, nil
}

// String returns the message of the error symbol.
func (s ErrorSymbol) String() string {
	return s.message
}

// Error returns an error that causes to signal an error with the given symbol
// and data.  The return value is of type LazySignalError.  Signaling an error
// will evaluate symbol and data lazily.  The evaluation is best-effort since
// it can itself fail.  If it fails the symbol and/or data might be lost, but
// Emacs will signal some error in any case.  If you already have an evaluated
// symbol and data value, use Signal instead.
func (s ErrorSymbol) Error(data ...In) error {
	return Error{s, data}
}

// Emacs interns the error symbol in the default obarray and returns the symbol
// value.
func (s ErrorSymbol) Emacs(e Env) (Value, error) {
	return s.name.Emacs(e)
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

var wrongTypeArgument = ErrorSymbol{"wrong-type-argument", "Wrong type argument"}

// OverflowError returns an error that will cause Emacs to signal an error of
// type overflow‑error.  Use this if you notice that an integer doesn’t fit
// into its target type.  val is the string representation of the overflowing
// value.  val is a string instead of an In to avoid further overflows when
// converting it to an Emacs value.
func OverflowError(val string) error {
	return overflowError.Error(String(val))
}

var overflowError = ErrorSymbol{"overflow-error", "Arithmetic overflow error"}

type nonlocalExit interface {
	// signal sets the nonlocal exit in the environment.  Call it only
	// immediately before returning control to Emacs.
	signal(Env)
}

func (x Error) signal(e Env) {
	// We need to be careful here to not call other signal functions, as
	// that easily leads to infinite recursion.
	symbol, err := x.Symbol.Emacs(e)
	if err != nil {
		symbol = e.uncheckedIntern(Symbol(baseError.name))
	}
	data, err := x.Data.Emacs(e)
	if err != nil {
		data = e.uncheckedNil()
	}
	Signal{symbol, data}.signal(e)
}

func (s Signal) signal(e Env) {
	C.non_local_exit_signal(e.raw(), s.Symbol.r, s.Data.r)
}

func (t Throw) signal(e Env) {
	C.non_local_exit_throw(e.raw(), t.Tag.r, t.Value.r)
}

// signal sets the nonlocal exit state in e.  Call it only immediately before
// returning control to Emacs.
func (e Env) signal(err error) {
	if err == nil {
		return
	}
	if n, ok := err.(nonlocalExit); ok {
		n.signal(e)
		return
	}
	Error{baseError, List{String(err.Error())}}.signal(e)
}

// check converts a pending nonlocal exit to a Go error.  If no nonlocal exit
// is pending, check returns nil.  If a signal is pending, check returns an
// error of dynamic type Signal.  If a throw is pending, check returns an error
// of dynamic type Throw.
func (e Env) check() error {
	var a, b Value
	switch k := C.non_local_exit_get(e.raw(), &a.r, &b.r); k {
	case C.emacs_funcall_exit_return:
		return nil
	case C.emacs_funcall_exit_signal:
		C.non_local_exit_clear(e.raw())
		return Signal{a, b}
	case C.emacs_funcall_exit_throw:
		C.non_local_exit_clear(e.raw())
		return Throw{a, b}
	default:
		// This cannot really happen, but better safe than sorry.
		return WrongTypeArgument("module-funcall-exit-p", Int(k))
	}
}

// checkValue is like check, but also returns v for convenience.
func (e Env) checkValue(v Value) (Value, error) {
	return e.checkRaw(v.r)
}

// checkRaw is like check, but also wraps v in Value for convenience.
func (e Env) checkRaw(v C.emacs_value) (Value, error) {
	return Value{v}, e.check()
}

var baseError = DefineError("go-error", "Generic Go error")

type errorSymbol struct {
	name    Name
	message string
	parents []ErrorSymbol
}

func (s errorSymbol) define(e Env) error {
	parents := make(List, len(s.parents))
	for i, p := range s.parents {
		parents[i] = p
	}
	_, err := e.Call("define-error", s.name, String(s.message), parents)
	return err
}

type errorManager struct {
	mu    sync.Mutex
	syms  []errorSymbol
	names map[Name]struct{}
}

func (m *errorManager) register(s errorSymbol) error {
	if err := s.name.validate(); err != nil {
		return err
	}
	if s.message == "" {
		return fmt.Errorf("empty error message for error symbol %s", s.name)
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, dup := m.names[s.name]; dup {
		return fmt.Errorf("duplicate error symbol %s", s.name)
	}
	m.syms = append(m.syms, s)
	if m.names == nil {
		m.names = make(map[Name]struct{})
	}
	m.names[s.name] = struct{}{}
	return nil
}

func (m *errorManager) mustRegister(s errorSymbol) {
	if err := m.register(s); err != nil {
		panic(err)
	}
}

func (m *errorManager) define(e Env) error {
	for _, s := range m.copy() {
		if err := s.define(e); err != nil {
			return err
		}
	}
	return nil
}

func (m *errorManager) copy() []errorSymbol {
	m.mu.Lock()
	defer m.mu.Unlock()
	r := make([]errorSymbol, len(m.syms))
	copy(r, m.syms)
	return r
}

var errorSymbols errorManager

func init() {
	OnInit(errorSymbols.define)
}
