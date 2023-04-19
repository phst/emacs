// Copyright 2019, 2022, 2023 Google LLC
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

// List represents an Emacs list.
type List []In

// Emacs creates a new list from the given elements.
func (l List) Emacs(e Env) (Value, error) {
	return e.List(l...)
}

// ListOut is an [Out] that converts an Emacs list to the slice Data.  The
// concrete element type is determined by the return value of the New function.
type ListOut struct {
	// New must return a new list element each time it’s called.
	New func() Out

	// FromEmacs fills Data with the elements from the list.
	Data []Out
}

// FromEmacs sets l.Data to a new slice containing the elements of the Emacs
// list v.  It returns an error if v is not a list.  FromEmacs calls l.New for
// each element in v.  l.New must return a new Out value for the element.  If
// FromEmacs returns an error, it doesn’t modify l.Data.  FromEmacs assumes
// that v is a true list.  In particular, it may loop forever if v is circular.
func (l *ListOut) FromEmacs(e Env, v Value) error {
	var r []Out
	f := func(car Value) error {
		o := l.New()
		if err := o.FromEmacs(e, car); err != nil {
			return err
		}
		r = append(r, o)
		return nil
	}
	if err := e.Dolist(v, f); err != nil {
		return err
	}
	l.Data = r
	return nil
}

// UnpackList represents a destructuring binding on an Emacs list.
type UnpackList []Out

// FromEmacs fills *u with elements from the list v.  It returns an error if v
// is not a list.  If the list is shorter than *u, FromEmacs sets the length of
// *u to the length of the list.  If the list is longer than *u, FromEmacs
// converts only the first len(*u) elements of the list and ignores the rest.
// FromEmacs assumes that v is a true list.  If v is dotted or cyclic, the
// contents of *u are unspecified.  If FromEmacs returns an error, the contents
// of *u are also unspecified.
func (u *UnpackList) FromEmacs(e Env, v Value) error {
	s := *u
	if len(s) == 0 {
		return nil
	}
	n := 0
	var o Value
	var err error
	for i := e.Iter(v, &o, &err); i.Ok(); i.Next() {
		if err1 := s[n].FromEmacs(e, o); err1 != nil {
			return err1
		}
		n++
		if n >= len(s) {
			break
		}
	}
	if err != nil {
		return err
	}
	*u = s[:n]
	return nil
}

// List creates and returns a new Emacs list containing the given values.
func (e Env) List(os ...In) (Value, error) {
	if len(os) == 0 {
		// Trivial micro-optimization since this is so frequent.
		return e.Nil()
	}
	return e.Call("list", os...)
}

// Car is an [In] that represents the car of List.
type Car struct{ List In }

// Emacs returns the car of c.List.  It returns an error if c.List is not a
// list.
func (c Car) Emacs(e Env) (Value, error) {
	return e.Car(c.List)
}

// Car returns the car of list.  It returns an error if list is not a list.
func (e Env) Car(list In) (Value, error) {
	return e.Call("car", list)
}

// CarOut sets car to the car of list.  It returns an error if list is not a
// list.
func (e Env) CarOut(list In, car Out) error {
	return e.CallOut("car", car, list)
}

// Cdr is an [In] that represents the cdr of List.
type Cdr struct{ List In }

// Emacs returns the cdr of c.List.  It returns an error if c.List is not a
// list.
func (c Cdr) Emacs(e Env) (Value, error) {
	return e.Cdr(c.List)
}

// Cdr returns the cdr of list.  It returns an error if list is not a list.
func (e Env) Cdr(list In) (Value, error) {
	return e.Call("cdr", list)
}

// CdrOut sets cdr to the cdr of list.  It returns an error if list is not a
// list.
func (e Env) CdrOut(list In, cdr Out) error {
	return e.CallOut("cdr", cdr, list)
}

// Cons represents a cons cell.
type Cons struct{ Car, Cdr In }

// Emacs creates and returns a new cons cell.
func (c Cons) Emacs(e Env) (Value, error) {
	return e.Cons(c.Car, c.Cdr)
}

// Cons creates and returns a cons cell (car . cdr).
func (e Env) Cons(car, cdr In) (Value, error) {
	return e.Call("cons", car, cdr)
}

// Uncons represents a destructuring binding of a cons cell.
type Uncons struct{ Car, Cdr Out }

// FromEmacs sets c.Car and c.Cdr to the values in cons.  It returns an error
// if cons is not a cons cell.  In case of an error, the value of c.Car and
// c.Cdr is unspecified.
func (c Uncons) FromEmacs(e Env, cons Value) error {
	return e.UnconsOut(cons, c.Car, c.Cdr)
}

// Uncons returns the car and cdr of cons.  It returns an error if cons is not
// a cons cell.
func (e Env) Uncons(cons Value) (car Value, cdr Value, err error) {
	if e.IsNil(cons) {
		return Value{}, Value{}, WrongTypeArgument("consp", cons)
	}
	car, err = e.Car(cons)
	if err != nil {
		return
	}
	cdr, err = e.Cdr(cons)
	return
}

// UnconsOut sets car and cdr to the car and cdr of cons.  It returns an error
// if cons is not a cons cell.  If UnconsOut fails, it is unspecified whether
// car and cdr have been modified.
func (e Env) UnconsOut(cons Value, car, cdr Out) error {
	if e.IsNil(cons) {
		return WrongTypeArgument("consp", cons)
	}
	if err := e.CarOut(cons, car); err != nil {
		return err
	}
	return e.CdrOut(cons, cdr)
}

// Length returns the length of the sequence represented by seq.  It returns an
// error if seq is not a sequence.  Just like the Emacs length function, it may
// loop forever if seq is a circular object.
func (e Env) Length(seq Value) (int, error) {
	r, err := e.Call("length", seq)
	if err != nil {
		return -1, err
	}
	i, err := e.Int(r)
	if err != nil {
		return -1, err
	}
	return int64ToInt(i)
}

// Iter is an iterator over a list.  Use [Env.Iter] to create Iter values.  The
// zero Iter is not a valid iterator.  Iter values can’t outlive the
// environment that created them.  Don’t pass Iter values to other goroutines.
//
// Typical use, with *T implementing the [Out] interface:
//
//	var elem T
//	var err error
//	for i := env.Iter(list, &elem, &err); i.Ok(); i.Next() {
//		// …
//	}
//	if err != nil {
//		return err
//	}
type Iter struct {
	env  Env
	tail Value
	elem Out
	err  *error
}

// Iter creates an [Iter] value that iterates over list.  [Iter.Next] will set
// elem to the elements of the list, and *err to any error.  Iter assumes that
// list is a true list.  If list is circular, the iteration may never
// terminate.
//
// Typical use, with *T implementing the [Out] interface:
//
//	var elem T
//	var err error
//	for i := env.Iter(list, &elem, &err); i.Ok(); i.Next() {
//		// …
//	}
//	if err != nil {
//		return err
//	}
//
// See [Dolist] for a simpler iteration method.
func (e Env) Iter(list Value, elem Out, err *error) *Iter {
	i := &Iter{e, list, elem, err}
	i.setElem()
	return i
}

// Ok returns whether the iterator is still valid.  It is valid if no error is
// set and there are still elements left in the list.
func (i *Iter) Ok() bool {
	return *i.err == nil && i.env.IsNotNil(i.tail)
}

// Next sets the elem passed to [Env.Iter] to the next element in the list.  If
// Next fails, it sets the error passed to [Env.Iter], and [Iter.Ok] will
// return false.
func (i *Iter) Next() {
	i.tail, *i.err = i.env.Cdr(i.tail)
	i.setElem()
}

func (i *Iter) setElem() {
	if i.Ok() {
		*i.err = i.env.CarOut(i.tail, i.elem)
	}
}

// Dolist calls f for each element in list.  It returns an error if list is not
// a list.  If f returns an error, the loop terminates and Dolist returns the
// same error.  If list is a circular list, Dolist may loop forever.
func (e Env) Dolist(list Value, f func(Value) error) error {
	var elem Value
	var err error
	for i := e.Iter(list, &elem, &err); i.Ok(); i.Next() {
		if err1 := f(elem); err1 != nil {
			return err1
		}
	}
	return err
}
