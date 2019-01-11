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
import "C"

import (
	"math/big"
	"reflect"
	"time"
)

// Value represents an Emacs object.  The zero Value is not valid.  Various
// functions in this package return Value values.  These functions are either
// methods on the Env type or take an Env argument.  In any case, the returned
// Value values are only valid (or “live”) as long as the Env used to create
// them is valid.  Don’t pass Value values to other goroutines.  Two different
// Value values may represent the same Emacs value.  Use Env.Eq instead of the
// == operator to compare values.  See
// https://phst.eu/emacs-modules#emacs-values for details.
type Value struct{ r C.emacs_value }

// In is a value that knows how to convert itself into an Emacs object.  You
// can implement In for your own types if you want this package to convert them
// to Emacs values automatically.  This package also defines a few wrapper
// implementations for primitive types such as In or String.
type In interface {
	// Emacs returns an Emacs object corresponding to the receiver, or an
	// error if the conversion fails.  Implementations should document
	// whether it always returns a new object or not.  They should also
	// document potential side effects.
	Emacs(Env) (Value, error)
}

// Out is a value that knows how to convert itself from an Emacs object.  You
// can implement In for your own types if you want this package to convert them
// from Emacs values automatically.  Pointers to wrapper implementations for
// In, such as *Int or *String, implement Out.
type Out interface {
	// FromEmacs sets the receiver to a Go value corresponding to the Emacs
	// object.  Implementations should document whether FromEmacs modifies
	// the receiver in case of an error.
	FromEmacs(Env, Value) error
}

// Emacs returns v.  It never returns an error.
func (v Value) Emacs(Env) (Value, error) {
	return v, nil
}

// FromEmacs sets *v to u.  It never returns an error.
func (v *Value) FromEmacs(e Env, u Value) error {
	*v = u
	return nil
}

// NewIn returns an In value that converts the dynamic type of v to an Emacs
// value.  If v already implements In, NewIn returns it directly.  Otherwise,
// if the dynamic type of v is known, NewIn returns one of the predefined In
// implementations (Int, Float, String, Reflect, …).  Otherwise, NewIn returns
// an In instance that uses reflection to convert itself to Emacs.
func NewIn(v interface{}) In {
	if i := newIn(v); i != nil {
		return i
	}
	return Reflect(reflect.ValueOf(v))
}

func newIn(v interface{}) In {
	if i, ok := v.(In); ok {
		return i
	}
	switch v := v.(type) {
	case reflect.Value:
		return Reflect(v)
	case nil:
		return Nil
	case bool:
		return Bool(v)
	case int:
		return Int(v)
	case int8:
		return Int(v)
	case int16:
		return Int(v)
	case int32:
		return Int(v)
	case int64:
		return Int(v)
	case uint8:
		return Uint(v)
	case uint16:
		return Uint(v)
	case uint32:
		return Uint(v)
	case uint64:
		return Uint(v)
	case float32:
		return Float(v)
	case float64:
		return Float(v)
	case string:
		return String(v)
	case []byte:
		return Bytes(v)
	case big.Int:
		return BigInt(v)
	case time.Time:
		return Time(v)
	case time.Duration:
		return Duration(v)
	default:
		return nil
	}
}

// NewOut returns an Out value that sets p to the Go representation of an Emacs
// value.  Typically p is a pointer.  Only a few types such as reflect.Value
// can be set without indirection.  If p already implements Out, NewOut returns
// it directly.  Otherwise, if the dynamic type of v is known, NewOut returns
// one of the predefined Out implementations (*Int, *Float, *String, Reflect,
// …).  Otherwise, NewOut returns an Out instance that uses reflection to
// convert Emacs values.
func NewOut(p interface{}) Out {
	if i := newOut(p); i != nil {
		return i
	}
	return Reflect(reflect.ValueOf(p))
}

func newOut(p interface{}) Out {
	if i, ok := p.(Out); ok {
		return i
	}
	switch p := p.(type) {
	case reflect.Value:
		return Reflect(p)
	case *reflect.Value:
		return (*Reflect)(p)
	case *bool:
		return (*Bool)(p)
	case *int64:
		return (*Int)(p)
	case *uint64:
		return (*Uint)(p)
	case *float64:
		return (*Float)(p)
	case *string:
		return (*String)(p)
	case *[]byte:
		return (*Bytes)(p)
	case *big.Int:
		return (*BigInt)(p)
	case *time.Time:
		return (*Time)(p)
	case *time.Duration:
		return (*Duration)(p)
	default:
		return nil
	}
}

// Ignore is an Out that does nothing.
type Ignore struct{}

// FromEmacs does nothing.
func (Ignore) FromEmacs(Env, Value) error {
	return nil
}

// Emacs converts v to an Emacs value and returns the value.  See NewIn for
// details of the conversion process.
func (e Env) Emacs(v interface{}) (Value, error) {
	return NewIn(v).Emacs(e)
}

// Go converts v to a Go value and stores it in p.  See NewOut for details of
// the conversion process.
func (e Env) Go(v Value, p interface{}) error {
	return NewOut(p).FromEmacs(e, v)
}

func castToIn(v reflect.Value) In   { return v.Interface().(In) }
func castToOut(v reflect.Value) Out { return v.Interface().(Out) }
