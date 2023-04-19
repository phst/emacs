// Copyright 2019, 2021, 2023 Google LLC
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

// #include <stdlib.h>
// #include "wrappers.h"
import "C"

import (
	"fmt"
	"math"
	"math/big"
	"reflect"
	"unsafe"
)

// Int is a type with underlying type int64 that knows how to convert itself
// into an Emacs value.
type Int int64

// Emacs creates an Emacs value representing the given integer.  It returns an
// error if the integer value is too big for Emacs.
func (i Int) Emacs(e Env) (Value, error) {
	return e.checkValue(C.phst_emacs_make_integer(e.raw(), C.int64_t(i)))
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
	i := C.phst_emacs_extract_integer(e.raw(), v.r)
	return int64(i.value), e.check(i.base)
}

// BigInt sets z to the integer stored in v.  It returns an error if v is not
// an integer.
func (e Env) BigInt(v Value, z *big.Int) error {
	r := C.phst_emacs_extract_big_integer(e.raw(), v.r)
	if err := e.check(r.base); err != nil {
		return err
	}
	if r.sign == 0 {
		z.SetInt64(0)
		return nil
	}
	defer C.free(unsafe.Pointer(r.data))
	z.SetBytes(C.GoBytes(unsafe.Pointer(r.data), r.size))
	if r.sign == -1 {
		z.Neg(z)
	}
	return nil
}

// Uint is a type with underlying type uint64 that knows how to convert itself
// into an Emacs value.
type Uint uint64

// Emacs creates an Emacs value representing the given integer.  It returns an
// error if the integer value is too big for Emacs.
func (i Uint) Emacs(e Env) (Value, error) {
	if i > math.MaxInt64 {
		z := new(big.Int)
		z.SetUint64(uint64(i))
		return (*BigInt)(z).Emacs(e)
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
	var z big.Int
	if err := e.BigInt(v, &z); err != nil {
		return 0, err
	}
	if !z.IsUint64() {
		return 0, WrongTypeArgument("natnump", String(z.String()))
	}
	return z.Uint64(), nil
}

// BigInt is a type with underlying type [big.Int] that knows how to convert
// itself to and from an Emacs value.
type BigInt big.Int

// String formats the big integer as a string.  It calls big.Int.String.
func (i *BigInt) String() string { return (*big.Int)(i).String() }

// Emacs creates an Emacs value representing the given integer.  It returns an
// error if the integer value is too big for Emacs.
func (i *BigInt) Emacs(e Env) (Value, error) {
	b := (*big.Int)(i)
	if b.IsInt64() {
		return Int(b.Int64()).Emacs(e)
	}
	p := b.Bytes()
	return e.checkValue(C.phst_emacs_make_big_integer(e.raw(), C.int(b.Sign()), (*C.uint8_t)(&p[0]), C.int64_t(len(p))))
}

// FromEmacs sets *i to the integer stored in v.  It returns an error if v is
// not an integer.
func (i *BigInt) FromEmacs(e Env, v Value) error {
	return e.BigInt(v, (*big.Int)(i))
}

func intIn(v reflect.Value) In   { return Int(v.Int()) }
func intOut(v reflect.Value) Out { return reflectInt(v) }

type reflectInt reflect.Value

func (r reflectInt) FromEmacs(e Env, v Value) error {
	i, err := e.Int(v)
	if err != nil {
		return err
	}
	s := reflect.Value(r).Elem()
	if s.OverflowInt(i) {
		return OverflowError(fmt.Sprint(i))
	}
	s.SetInt(i)
	return nil
}

func uintIn(v reflect.Value) In   { return Uint(v.Uint()) }
func uintOut(v reflect.Value) Out { return reflectUint(v) }

type reflectUint reflect.Value

func (r reflectUint) FromEmacs(e Env, v Value) error {
	i, err := e.Uint(v)
	if err != nil {
		return err
	}
	s := reflect.Value(r).Elem()
	if s.OverflowUint(i) {
		return OverflowError(fmt.Sprint(i))
	}
	s.SetUint(i)
	return nil
}
