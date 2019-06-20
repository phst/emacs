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

// #include "wrappers.h"
import "C"

import "reflect"

// Float is a type with underlying type float64 that knows how to convert
// itself into an Emacs value.
type Float float64

// Emacs creates an Emacs value representing the given floating-point number.
func (f Float) Emacs(e Env) (Value, error) {
	return e.checkValue(C.make_float(e.raw(), C.double(f)))
}

// FromEmacs sets *f to the floating-point number stored in v.  It returns an
// error if v is not a floating-point value.
func (f *Float) FromEmacs(e Env, v Value) error {
	r, err := e.Float(v)
	if err != nil {
		return err
	}
	*f = Float(r)
	return nil
}

// Float returns the floating-point number stored in v.  It returns an error if
// v is not a floating-point value.
func (e Env) Float(v Value) (float64, error) {
	r := C.extract_float(e.raw(), v.r)
	if err := e.check(r.base); err != nil {
		return 0, err
	}
	return float64(r.value), nil
}

func floatIn(v reflect.Value) In   { return Float(reflect.Value(v).Float()) }
func floatOut(v reflect.Value) Out { return reflectFloat(v) }

type reflectFloat reflect.Value

func (r reflectFloat) FromEmacs(e Env, v Value) error {
	f, err := e.Float(v)
	if err != nil {
		return err
	}
	reflect.Value(r).SetFloat(f)
	return nil
}
