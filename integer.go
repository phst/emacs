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

// #include <assert.h>
// #include <emacs-module.h>
// static_assert(INTMAX_MIN == INT64_MIN, "unsupported architecture");
// static_assert(INTMAX_MAX == INT64_MAX, "unsupported architecture");
// int64_t extract_integer(emacs_env *env, emacs_value value) {
//   return env->extract_integer(env, value);
// }
// emacs_value make_integer(emacs_env *env, int64_t value) {
//   return env->make_integer(env, value);
// }
import "C"

import (
	"fmt"
	"math"
	"math/big"
	"reflect"
)

// Int is a type with underlying type int64 that knows how to convert itself
// into an Emacs value.
type Int int64

// Emacs creates an Emacs value representing the given integer.  It returns an
// error if the integer value is too big for Emacs.
func (i Int) Emacs(e Env) (Value, error) {
	return e.checkRaw(C.make_integer(e.raw(), C.int64_t(i)))
}

// FromEmacs sets *i to the integer stored in v.  It returns an error if v is
// not an integer, or if doesn’t fit into an int64.
func (i *Int) FromEmacs(e Env, v Value) error {
	r, err := e.Int(v)
	if err != nil {
		return err
	}
	*i = Int(r)
	return nil
}

// Int returns the integer stored in v.  It returns an error if v is not an
// integer, or if it doesn’t fit into an int64.
func (e Env) Int(v Value) (int64, error) {
	i := C.extract_integer(e.raw(), v.r)
	return int64(i), e.check()
}

// Uint is a type with underlying type uint64 that knows how to convert itself
// into an Emacs value.
type Uint uint64

// Emacs creates an Emacs value representing the given integer.  It returns an
// error if the integer value is too big for Emacs.
func (i Uint) Emacs(e Env) (Value, error) {
	if i > math.MaxInt64 {
		return Value{}, OverflowError(fmt.Sprint(i))
	}
	return Int(i).Emacs(e)
}

// FromEmacs sets *i to the integer stored in v.  It returns an error if v is
// not an integer, or if it doesn’t fit into an uint64.
func (i *Uint) FromEmacs(e Env, v Value) error {
	r, err := e.Uint(v)
	if err != nil {
		return err
	}
	*i = Uint(r)
	return nil
}

// Uint returns the integer stored in v.  It returns an error if v is not an
// integer, or if it doesn’t fit into an uint64.
func (e Env) Uint(v Value) (uint64, error) {
	i, err := e.Int(v)
	if err != nil {
		return 0, err
	}
	if i < 0 {
		return 0, WrongTypeArgument("natnump", String(fmt.Sprint(i)))
	}
	return uint64(i), nil
}

// BigInt is a type with underlying type big.Int that knows how to convert
// itself to and from an Emacs value.
type BigInt big.Int

// String formats the big integer as a string.  It calls big.Int.String.
func (i *BigInt) String() string { return (*big.Int)(i).String() }

// Emacs creates an Emacs value representing the given integer.  It returns an
// error if the integer value is too big for Emacs.
func (i BigInt) Emacs(e Env) (Value, error) {
	b := big.Int(i)
	if !b.IsInt64() {
		return Value{}, OverflowError(b.String())
	}
	return Int(b.Int64()).Emacs(e)
}

// FromEmacs sets *i to the integer stored in v.  It returns an error if v is
// not an integer.
func (i *BigInt) FromEmacs(e Env, v Value) error {
	n, err := e.Int(v)
	if err != nil {
		return err
	}
	(*big.Int)(i).SetInt64(n)
	return nil
}

func intIn(v reflect.Value) In   { return Int(reflect.Value(v).Int()) }
func intOut(v reflect.Value) Out { return reflectInt(v) }

type reflectInt reflect.Value

func (r reflectInt) FromEmacs(e Env, v Value) error {
	i, err := e.Int(v)
	if err != nil {
		return err
	}
	s := reflect.Value(r)
	if s.OverflowInt(i) {
		return OverflowError(fmt.Sprint(i))
	}
	s.SetInt(i)
	return nil
}

func uintIn(v reflect.Value) In   { return Uint(reflect.Value(v).Uint()) }
func uintOut(v reflect.Value) Out { return reflectUint(v) }

type reflectUint reflect.Value

func (r reflectUint) FromEmacs(e Env, v Value) error {
	i, err := e.Uint(v)
	if err != nil {
		return err
	}
	s := reflect.Value(r)
	if s.OverflowUint(i) {
		return OverflowError(fmt.Sprint(i))
	}
	s.SetUint(i)
	return nil
}

func bigIntIn(v reflect.Value) In   { return BigInt(v.Interface().(big.Int)) }
func bigIntOut(v reflect.Value) Out { return (*BigInt)(v.Interface().(*big.Int)) }
