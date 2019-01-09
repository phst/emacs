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
// emacs_value intern(emacs_env *env, _GoString_ symbol_name) {
//   return env->intern(env, _GoStringPtr(symbol_name));
// }
import "C"

import (
	"fmt"
	"unicode/utf8"
)

// Symbol represents an Emacs symbol.  Call interns Symbol values instead of
// converting them to an Emacs string.
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

// Name is a Symbol that names a definition such as a function or error symbol.
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
	// See https://phst.github.io/emacs-modules#intern.
	if isNonNullASCII(string(s)) {
		return e.checkValue(e.uncheckedIntern(s))
	}
	return e.Call("intern", s)
}

// MaybeIntern returns nameOrValue as-is if it’s a Value and calls Intern if
// it’s a Symbol, Name, or string.  Otherwise it returns an error.
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
	return e.checkValue(e.uncheckedNil())
}

// uncheckedNil returns the interned symbol nil.  Unlike Nil, it doesn’t clear
// the error flag on e.  Use uncheckedNil only if you call either e.check or
// e.signal immediately afterwards.
func (e Env) uncheckedNil() Value {
	return e.uncheckedIntern("nil")
}

// uncheckedIntern interns the ASCII symbol s in the default obarray.  Unlike
// Intern, it doesn’t clear the error flag on e.  s must not contain non-ASCII
// characters or null bytes.  Use uncheckedIntern only if you call either
// e.check or e.signal immediately afterwards.
func (e Env) uncheckedIntern(s Symbol) Value {
	return Value{C.intern(e.raw(), string(s)+"\x00")}
}
