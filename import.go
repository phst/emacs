// Copyright 2019, 2023 Google LLC
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
	"reflect"
)

// Import imports an Emacs function as a Go function.  name must be the Emacs
// symbol name of the function.  fp must be a pointer to a variable of function
// type.  Import sets *fp to a function that calls the Emacs function name.
//
// Import must be called before Emacs loads the module.  Typically you should
// call Import from an init function.
//
// The dynamic type of *fp must have one of the following two forms:
//
//	func(Env, Args...) (Ret, error)
//	func(Env, Args...) error
//
// Here Args is a (possibly empty) list of argument types, and Ret is the
// return type.  If the function type has any other form, Import panics.
//
// Calling the Go function converts all arguments to Emacs, calls the Emacs
// function name, and potentially converts the return value back to the Ret
// type.  If any of the involved types can’t be represented as Emacs value,
// Import panics.
//
// If you don’t want argument autoconversion, use ImportFunc instead.
//
// You can call Import safely from multiple goroutines, provided that setting
// *fp is race-free.
func Import(name Name, fp interface{}) {
	if err := name.validate(); err != nil {
		panic(err)
	}
	r := importAuto{name: name}
	fun := reflect.ValueOf(fp).Elem()
	t := fun.Type()
	numIn := t.NumIn()
	if numIn <= 0 || t.In(0) != envType {
		panic(fmt.Errorf("can’t import %s: function doesn’t accept an Env argument", name))
	}
	if t.IsVariadic() {
		numIn--
		conv, err := InFuncFor(t.In(numIn).Elem())
		if err != nil {
			panic(fmt.Errorf("can’t import %s: don’t know how to convert variadic argument: %s", name, err))
		}
		r.varConv = conv
	}
	for i := 1; i < numIn; i++ {
		conv, err := InFuncFor(t.In(i))
		if err != nil {
			panic(fmt.Errorf("can’t import %s: don’t know how to convert argument %d: %s", name, i, err))
		}
		r.inConv = append(r.inConv, conv)
	}
	numOut := t.NumOut()
	if numOut > 2 {
		panic(fmt.Errorf("can’t import %s: function returns too many arguments", name))
	}
	if t.Out(numOut-1) != errorType {
		panic(fmt.Errorf("can’t import %s: function doesn’t return an error argument", name))
	}
	if numOut == 2 {
		r.retType = t.Out(0)
		conv, err := OutFuncFor(reflect.PtrTo(t.Out(0)))
		if err != nil {
			panic(fmt.Errorf("can’t import %s; don’t know how to convert result: %s", name, err))
		}
		r.outConv = conv
	}
	fun.Set(reflect.MakeFunc(t, r.call))
}

// ImportFunc imports an Emacs function as a Go function.  name must be the
// Emacs symbol name of the function.  ImportFunc returns a new [Func] that
// calls the Emacs function name.  Unlike [Import], there is no type
// autoconversion.
//
// ImportFunc must be called before Emacs loads the module.  Typically you
// should initialize a global variable with the return value of ImportFunc.
//
// You can call ImportFunc safely from multiple goroutines.
func ImportFunc(name Name) Func {
	if err := name.validate(); err != nil {
		panic(err)
	}
	i := importFunc{name}
	return i.call
}

type importAuto struct {
	name    Name
	retType reflect.Type
	inConv  []InFunc
	varConv InFunc
	outConv OutFunc
}

func (i importAuto) call(in []reflect.Value) (out []reflect.Value) {
	e := in[0].Interface().(Env)
	var args []In
	for i, conv := range i.inConv {
		args = append(args, conv(in[i+1]))
	}
	if conv := i.varConv; conv != nil {
		last := in[len(in)-1]
		for i := 0; i < last.Len(); i++ {
			args = append(args, conv(last.Index(i)))
		}
	}
	var ret Out
	if i.retType == nil {
		ret = Ignore{}
	} else {
		o := reflect.New(i.retType)
		out = append(out, o.Elem())
		ret = i.outConv(o)
	}
	err := e.CallOut(i.name, ret, args...)
	return append(out, reflect.ValueOf(&err).Elem())
}

type importFunc struct{ name Name }

func (i importFunc) call(e Env, args []Value) (Value, error) {
	fv, err := e.Intern(Symbol(i.name))
	if err != nil {
		return Value{}, err
	}
	return e.Funcall(fv, args)
}
