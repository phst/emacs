// Copyright 2019, 2021, 2023, 2024 Google LLC
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
// struct phst_emacs_value_result phst_emacs_intern(emacs_env *env,
//                                                  _GoString_ symbol_name) {
//   return phst_emacs_intern_impl(env, _GoStringPtr(symbol_name));
// }
import "C"

import (
	"fmt"
	"unicode/utf8"
)

// Symbol represents an Emacs symbol.  [Env.Call] interns Symbol values instead
// of converting them to an Emacs string.
type Symbol string

// Symbol returns the symbol name.
func (s Symbol) String() string {
	return string(s)
}

// Emacs interns the given symbol name in the default obarray and returns the
// symbol object.
func (s Symbol) Emacs(e Env) (Value, error) {
	return e.Intern(s)
}

// FromEmacs sets *s to the name of the Emacs symbol v.  It returns an error if
// v doesn’t represent a symbol.
func (s *Symbol) FromEmacs(e Env, v Value) error {
	r, err := e.Symbol(v)
	if err != nil {
		return err
	}
	*s = r
	return nil
}

// Symbol returns the name of the Emacs symbol v.  It returns an error if v is
// not a symbol.
func (e Env) Symbol(v Value) (Symbol, error) {
	nv, err := e.Call("symbol-name", v)
	if err != nil {
		return "", err
	}
	n, err := e.Str(nv)
	if err != nil {
		return "", err
	}
	return Symbol(n), err
}

// Name is a [Symbol] that names a definition such as a function or error
// symbol.  You can use a Name as an [Option] in [Export] and [ERTTest] to set
// the function or test name.
type Name Symbol

// Name returns the symbol name.
func (n Name) String() string {
	return Symbol(n).String()
}

// Emacs interns the given symbol name in the default obarray and returns the
// symbol object.
func (n Name) Emacs(e Env) (Value, error) {
	return Symbol(n).Emacs(e)
}

// FromEmacs sets *n to the name of the Emacs symbol v.  It returns an error if
// v is not a symbol.
func (n *Name) FromEmacs(e Env, v Value) error {
	return (*Symbol)(n).FromEmacs(e, v)
}

func (n Name) validate() error {
	if n == "" {
		return WrongTypeArgument("nonempty-string-p", n)
	}
	if !utf8.ValidString(string(n)) {
		return WrongTypeArgument("valid-string-p", n)
	}
	return nil
}

// Intern interns the given symbol name in the default obarray and returns the
// symbol object.
func (e Env) Intern(s Symbol) (Value, error) {
	// See
	// https://www.gnu.org/software/emacs/manual/html_node/elisp/Module-Misc.html#index-intern-1.
	if isNonNullASCII(string(s)) {
		return e.internASCII(s)
	}
	return e.Call("intern", s)
}

// MaybeIntern returns nameOrValue as-is if it’s a [Value] and calls
// [Env.Intern] if it’s a [Symbol], [Name], or string.  Otherwise it returns an
// error.
func (e Env) MaybeIntern(nameOrValue interface{}) (Value, error) {
	switch v := nameOrValue.(type) {
	case Value:
		return v, nil
	case Symbol:
		return e.Intern(v)
	case Name:
		return e.Intern(Symbol(v))
	case string:
		return e.Intern(Symbol(v))
	default:
		return Value{}, WrongTypeArgument("stringp", String(fmt.Sprint(v)))
	}
}

// Nil returns the interned symbol nil.  It fails only if interning nil fails.
func (e Env) Nil() (Value, error) {
	return e.internASCII("nil")
}

// internASCII interns the ASCII symbol s in the default obarray.  s must not
// contain non-ASCII characters or null bytes.
func (e Env) internASCII(s Symbol) (Value, error) {
	return e.checkValue(C.phst_emacs_intern(e.raw(), string(s)+"\x00"))
}
