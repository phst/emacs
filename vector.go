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

// Vector represents an Emacs vector.
type Vector []In

// Emacs creates a new vector from the given elements.
func (v Vector) Emacs(e Env) (Value, error) {
	return e.Call("vector", v...)
}

// VectorOut is an Out that converts an Emacs vector to the slice Data.  The
// concrete element type is determined by the return value of the New function.
type VectorOut struct {
	// New must return a new vector element each time it’s called.
	New func() Out

	// FromEmacs fills Data with the elements from the vector.
	Data []Out
}

// FromEmacs sets v.Data to a new slice containing the elements of the Emacs
// vector u.  It returns an error if u is not a vector.  FromEmacs calls v.New
// for each element in u.  v.New must return a new Out value for the element.
// If FromEmacs returns an error, it doesn’t modify v.Data.
func (v *VectorOut) FromEmacs(e Env, u Value) error {
	n, err := e.VecSize(u)
	if err != nil {
		return err
	}
	r := make([]Out, n)
	for i := range r {
		o := v.New()
		if err := e.VecGetOut(u, i, o); err != nil {
			return err
		}
		r[i] = o
	}
	v.Data = r
	return nil
}

// UnpackVector represents a destructuring binding on an Emacs vector.
type UnpackVector []Out

// FromEmacs fills *u with elements from the vector v.  It returns an error if
// v is not a vector.  If the vector is shorter than *u, FromEmacs sets the
// length of *u to the length of the vector.  If the vector is longer than *u,
// FromEmacs converts only the first len(*u) elements of the vector and ignores
// the rest.  If FromEmacs returns an error, the contents of *u are
// unspecified.
func (u *UnpackVector) FromEmacs(e Env, v Value) error {
	s := *u
	n, err := e.VecSize(v)
	if err != nil {
		return err
	}
	if n > len(s) {
		n = len(s)
	}
	s = s[:n]
	for i, o := range s {
		if err := e.VecGetOut(v, i, o); err != nil {
			return err
		}
	}
	*u = s
	return nil
}

// MakeVector creates and returns an Emacs vector of size n.  It initializes
// all elements to init.
func (e Env) MakeVector(n int, init Value) (Value, error) {
	return e.Call("make-vector", Int(n), init)
}

// VecGet returns the i-th element of vector.  It returns an error if vector is
// not a vector.
func (e Env) VecGet(vector Value, i int) (Value, error) {
	// d, err := intToCptrdiff(i)
	// if err != nil {
	// 	return Value{}, err
	// }
	return e.checkRaw(C.vec_get(e.raw(), vector.r, C.int64_t(i)))
}

// VecGetOut sets elem to the value of the i-th element of vector.  It returns
// an error if vector is not a vector.
func (e Env) VecGetOut(vector Value, i int, elem Out) error {
	o, err := e.VecGet(vector, i)
	if err != nil {
		return err
	}
	return elem.FromEmacs(e, o)
}

// VecSet sets the i-th element of the given Emacs vector.
func (e Env) VecSet(v Value, i int, elem Value) error {
	// ri, err := intToCptrdiff(i)
	// if err != nil {
	// 	return err
	// }
	C.vec_set(e.raw(), v.r, C.int64_t(i), elem.r)
	return e.check()
}

// VecSize returns the size of the given Emacs vector.
func (e Env) VecSize(v Value) (int, error) {
	r := int64(C.vec_size(e.raw(), v.r))
	if err := e.check(); err != nil {
		return -1, err
	}
	return int64ToInt(r)
}

type vectorIn struct{ elem InFunc }

func (i vectorIn) call(v reflect.Value) In {
	return makeVector{i, v}
}

type makeVector struct {
	vectorIn
	reflect.Value
}

func (m makeVector) Emacs(e Env) (Value, error) {
	if !m.IsValid() {
		return Value{}, WrongTypeArgument("go-valid-reflect-p", String(m.String()))
	}
	init, err := e.Nil()
	if err != nil {
		return Value{}, err
	}
	n := m.Len()
	r, err := e.MakeVector(n, init)
	if err != nil {
		return Value{}, err
	}
	for i := 0; i < n; i++ {
		o, err := m.elem(m.Index(i)).Emacs(e)
		if err != nil {
			return Value{}, err
		}
		if err := e.VecSet(r, i, o); err != nil {
			return Value{}, err
		}
	}
	return r, nil
}

type vectorOut struct{ elem OutFunc }

func (o vectorOut) call(v reflect.Value) Out {
	return getVector{o, v}
}

type getVector struct {
	vectorOut
	reflect.Value
}

func (g getVector) FromEmacs(e Env, v Value) error {
	n, err := e.VecSize(v)
	if err != nil {
		return err
	}
	s := reflect.MakeSlice(g.Type(), n, n)
	for i := 0; i < n; i++ {
		if err := e.VecGetOut(v, i, g.elem(s.Index(i))); err != nil {
			return err
		}
	}
	g.Set(s)
	return nil
}
