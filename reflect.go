// Copyright 2019, 2023, 2024 Google LLC
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

import (
	"fmt"
	"math/big"
	"reflect"
	"time"
)

// Reflect is a type with underlying type [reflect.Value] that knows how to
// convert itself to and from an Emacs value.
type Reflect reflect.Value

// Emacs attempts to convert r to an Emacs value.
func (r Reflect) Emacs(e Env) (Value, error) {
	v := reflect.Value(r)
	if !v.IsValid() {
		return Value{}, WrongTypeArgument("go-valid-reflect-p", String(fmt.Sprintf("%#v", v)))
	}
	if v.Kind() == reflect.Interface {
		u := v.Elem()
		if !u.IsValid() {
			return Value{}, WrongTypeArgument("go-not-nil-p", String(fmt.Sprintf("%#v", v)))
		}
		v = u
	}
	fun, err := InFuncFor(v.Type())
	if err != nil {
		return Value{}, err
	}
	return fun(v).Emacs(e)
}

// FromEmacs sets r to the Go representation of v.  It returns an error if r
// isn’t settable.
func (r Reflect) FromEmacs(e Env, v Value) error {
	s := reflect.Value(r)
	if !s.IsValid() {
		return WrongTypeArgument("go-valid-reflect-p", String(fmt.Sprintf("%#v", s)))
	}
	conv, err := OutFuncFor(s.Type())
	if err != nil {
		return err
	}
	return conv(s).FromEmacs(e, v)
}

type (
	// InFunc is a function that returns an [In] for the given value.
	InFunc func(reflect.Value) In

	// OutFunc is a function that returns an [Out] for the given value.
	OutFunc func(reflect.Value) Out
)

// InFuncFor returns an [InFunc] for the given type.  If there’s no known
// conversion from t to Emacs, InFuncFor returns an error.
func InFuncFor(t reflect.Type) (InFunc, error) {
	if t.Implements(inType) {
		return castToIn, nil
	}
	if u := valueTypes[t]; u != nil {
		return func(v reflect.Value) In { return castToIn(v.Convert(u)) }, nil
	}
	if u := pointerTypes[t]; u != nil {
		return func(v reflect.Value) In { return castToIn(v.Convert(u)) }, nil
	}
	switch t.Kind() {
	case reflect.Array, reflect.Slice:
		if t.Elem().Kind() == reflect.Uint8 {
			return byteArrayIn, nil
		}
		elem, err := InFuncFor(t.Elem())
		if err != nil {
			return nil, err
		}
		return vectorIn{elem}.call, nil
	case reflect.Map:
		key, err := InFuncFor(t.Key())
		if err != nil {
			return nil, err
		}
		value, err := InFuncFor(t.Elem())
		if err != nil {
			return nil, err
		}
		return hashIn{HashTestFor(t.Key()), key, value}.call, nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return intIn, nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return uintIn, nil
	case reflect.Float32, reflect.Float64:
		return floatIn, nil
	default:
		return nil, WrongTypeArgument("go-known-type-p", String(t.String()))
	}
}

// OutFuncFor returns an [OutFunc] for the given type.  If there’s no known
// conversion from Emacs to t, OutFuncFor returns an error.
func OutFuncFor(t reflect.Type) (OutFunc, error) {
	if t.Implements(outType) {
		return castToOut, nil
	}
	if u := pointerTypes[t]; u != nil {
		return func(v reflect.Value) Out { return castToOut(v.Convert(u)) }, nil
	}
	if t.Kind() != reflect.Ptr {
		return nil, WrongTypeArgument("go-pointer-type-p", String(t.String()))
	}
	t = t.Elem()
	if u := valueTypes[t]; u != nil {
		f := func(v reflect.Value) Out {
			return castToOut(v.Convert(reflect.PtrTo(u)))
		}
		return f, nil
	}
	switch t.Kind() {
	case reflect.Array, reflect.Slice:
		if t.Elem().Kind() == reflect.Uint8 {
			return byteArrayOut, nil
		}
		elem, err := OutFuncFor(reflect.PtrTo(t.Elem()))
		if err != nil {
			return nil, err
		}
		return vectorOut{elem}.call, nil
	case reflect.Map:
		key, err := OutFuncFor(reflect.PtrTo(t.Key()))
		if err != nil {
			return nil, err
		}
		value, err := OutFuncFor(reflect.PtrTo(t.Elem()))
		if err != nil {
			return nil, err
		}
		return hashOut{key, value}.call, nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return intOut, nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return uintOut, nil
	case reflect.Float32, reflect.Float64:
		return floatOut, nil
	default:
		return nil, WrongTypeArgument("go-known-type-p", String(t.String()))
	}
}

var (
	inType  = reflect.TypeOf((*In)(nil)).Elem()
	outType = reflect.TypeOf((*Out)(nil)).Elem()
)

// Map basic types to aliases that support Emacs.  Each value has to be a
// defined type with the corresponding key as underlying type.  The value types
// have to implement [In], and pointers to them have to implement [Out].
var valueTypes = map[reflect.Type]reflect.Type{
	reflect.TypeOf(bool(false)):      reflect.TypeOf(Bool(false)),
	reflect.TypeOf(""):               reflect.TypeOf(String("")),
	reflect.TypeOf(([]byte)(nil)):    reflect.TypeOf(Bytes(nil)),
	reflect.TypeOf(time.Time{}):      reflect.TypeOf(Time{}),
	reflect.TypeOf(time.Duration(0)): reflect.TypeOf(Duration(0)),
}

// Map basic types to aliases that support Emacs.  Each value type has to be
// convertible to its corresponding key type.  The value types have to
// implement both [In] and [Out].  This is a variant of valueTypes for types
// that are always used as pointers.
var pointerTypes = map[reflect.Type]reflect.Type{
	reflect.TypeOf((*big.Int)(nil)): reflect.TypeOf((*BigInt)(nil)),
}
