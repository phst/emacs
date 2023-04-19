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
	"errors"
	"fmt"
	"math"
	"reflect"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"unicode"
)

// Export arranges for a Go function to be exported to Emacs.  Call Export in
// an init function.  Loading the dynamic module will then define the Emacs
// function.  The function fun can be any Go function.  When calling the
// exported function from Emacs, arguments and return values are converted as
// described in the package documentation.  If you don’t want this
// autoconversion or need more control, use ExportFunc instead and convert
// arguments yourself.  If you want to export functions to Emacs after the
// module has been initialized, use the Export method of Env instead.
//
// The function may accept any number of arguments.  Optionally, the first
// argument may be of type Env.  In this case, Emacs passes a live environment
// value that you can use to interact with Emacs.  All other arguments are
// converted from Emacs as described in the package documentation.  If not all
// arguments are convertible from Emacs values, Export panics.
//
// The function must return either zero, one, or two results.  If the last or
// only result is of type error, a non-nil value causes Emacs to trigger a
// nonlocal exit as appropriate.  There may be at most one non-error result.
// Its value will be converted to an Emacs value as described in the package
// documentation.  If the type of the non-error result can’t be converted to an
// Emacs value, Export panics.  If there are invalid result patterns, Export
// panics.
//
// By default, Export derives the function’s name from its Go name by
// Lisp-casing it.  For example, MyFunc becomes my-func.  To specify a
// different name, pass a Name option.  If there’s no name or the name is
// already registered, Export panics.
//
// By default, the function has no documentation string.  To add one, pass a
// Doc option.
//
// You can call Export safely from multiple goroutines.
func Export(fun interface{}, opts ...Option) {
	ExportFunc(AutoFunc(fun, opts...))
}

// ExportFunc arranges for a Go function to be exported to Emacs.  Call
// ExportFunc in an init function.  Loading the dynamic module will then define
// the Emacs function.  Unlike Export, functions registered by ExportFunc don’t
// automatically convert their arguments and return values to and from Emacs.
// If name is empty, ExportFunc panics.  If doc is empty, the function won’t
// have a documentation string.  If you want to export functions to Emacs after
// the module has been initialized, use the ExportFunc method of Env instead.
//
// You can call ExportFunc safely from multiple goroutines.
func ExportFunc(name Name, fun Func, arity Arity, doc Doc) {
	if name == "" {
		panic("empty function name")
	}
	funcs.mustEnqueue(&function{Lambda{fun, arity, doc}, name, 0})
}

// Export exports a Go function to Emacs.  Unlike the global Export function,
// Env.Export requires a live environment and defines the Emacs function
// immediately.  The function fun can be any Go function.  When calling the
// exported function from Emacs, arguments and return values are converted as
// described in the package documentation.  If you don’t want this
// autoconversion or need more control, use ExportFunc instead and convert
// arguments yourself.
//
// The function may accept any number of arguments.  Optionally, the first
// argument may be of type Env.  In this case, Emacs passes a live environment
// value that you can use to interact with Emacs.  All other arguments are
// converted from Emacs as described in the package documentation.  If not all
// arguments are convertible from Emacs values, Export panics.
//
// The function must return either zero, one, or two results.  If the last or
// only result is of type error, a non-nil value causes Emacs to trigger a
// non-local exit as appropriate.  There may be at most one non-error result.
// Its value will be converted to an Emacs value as described in the package
// documentation.  If the type of the non-error result can’t be converted to an
// Emacs value, Export panics.  If there are invalid result patterns, Export
// panics.
//
// By default, Export derives the function’s name from its Go name by
// Lisp-casing it.  For example, MyFunc becomes my-func.  To specify a
// different name, pass a Name option.  To make the function anonymous, pass an
// Anonymous option.  If there’s no name and the function isn’t anonymous,
// AutoFunc panics.  If the name of a non-anonymous function is already
// registered, Export panics.
//
// By default, the function has no documentation string.  To add one, pass a
// Doc option.
func (e Env) Export(fun interface{}, opts ...Option) (Value, error) {
	return e.ExportFunc(AutoFunc(fun, opts...))
}

// ExportFunc exports a Go function to Emacs.  Unlike the global ExportFunc
// function, Env.ExportFunc requires a live environment and defines the Emacs
// function immediately.  Unlike Export, functions defined by ExportFunc don’t
// automatically convert their arguments and return values to and from Emacs.
// ExportFunc returns the Emacs function object of the new function.  If name
// is empty, the function is anonymous.  If name is not empty, it is bound to
// the new function.  If doc is empty, the function won’t have a documentation
// string.
func (e Env) ExportFunc(name Name, fun Func, arity Arity, doc Doc) (Value, error) {
	f := &function{Lambda{fun, arity, doc}, name, 0}
	if err := funcs.register(f); err != nil {
		return Value{}, err
	}
	return f.define(e)
}

// AutoFunc converts an arbitrary Go function to a Func.  When calling the
// exported function from Emacs, arguments and return values are converted as
// described in the package documentation.
//
// The function may accept any number of arguments.  Optionally, the first
// argument may be of type Env.  In this case, Emacs passes a live environment
// value that you can use to interact with Emacs.  All other arguments are
// converted from Emacs as described in the package documentation.  If not all
// arguments are convertible from Emacs values, AutoFunc panics.
//
// The function must return either zero, one, or two results.  If the last or
// only result is of type error, a non-nil value causes Emacs to trigger a
// non-local exit as appropriate.  There may be at most one non-error result.
// Its value will be converted to an Emacs value as described in the package
// documentation.  If the type of the non-error result can’t be converted to an
// Emacs value, AutoFunc panics.  If there are invalid result patterns,
// AutoFunc panics.
//
// By default, Export derives the function’s name from its Go name by
// Lisp-casing it.  For example, MyFunc becomes my-func.  To specify a
// different name, pass a Name option.  To make the function anonymous, pass an
// Anonymous option.  If there’s no name and the function isn’t anonymous,
// AutoFunc panics.
//
// By default, the function has no documentation string.  To add one, pass a
// Doc option.
//
// You can call AutoFunc safely from multiple goroutines.
func AutoFunc(fun interface{}, opts ...Option) (Name, Func, Arity, Doc) {
	v := reflect.ValueOf(fun)
	if v.Kind() != reflect.Func {
		panic(fmt.Errorf("%s is not a function", v))
	}
	d := exportAuto{fun: v}
	for _, opt := range opts {
		opt.apply(&d)
	}
	anon := d.flag&exportAnonymous != 0
	if !anon && d.name == "" {
		d.name = lispName(v)
	}
	if anon && d.name != "" {
		panic(fmt.Errorf("function %s declared as anonymous, but has a name", d.name))
	}
	t := v.Type()
	numIn := t.NumIn()
	hasEnv := numIn > 0 && t.In(0) == envType
	offset := 0
	if hasEnv {
		offset = 1
	}
	var arity Arity
	var hasErr bool
	if t.IsVariadic() {
		numIn--
		arity.Min = numIn - offset
		arity.Max = -1
		conv, err := OutFuncFor(reflect.PtrTo(t.In(numIn).Elem()))
		if err != nil {
			panic(fmt.Errorf("function %s: don’t know how to convert variadic type: %s", d.name, err))
		}
		d.varConv = conv
	} else {
		arity.Min = numIn - offset
		arity.Max = arity.Min
	}
	for i := offset; i < numIn; i++ {
		conv, err := OutFuncFor(reflect.PtrTo(t.In(i)))
		if err != nil {
			panic(fmt.Errorf("function %s: don’t know how to convert argument %d: %s", d.name, i, err))
		}
		d.inConv = append(d.inConv, conv)
	}
	hasRet := false
	switch t.NumOut() {
	case 0:
	case 1:
		hasErr = t.Out(0) == errorType
		hasRet = !hasErr
	case 2:
		if t.Out(1) != errorType {
			panic(fmt.Errorf("function %s: second result must be error, but is %s", d.name, t.Out(1)))
		}
		hasErr = true
		hasRet = true
	default:
		panic(fmt.Errorf("function %s: too many results", d.name))
	}
	if hasEnv {
		d.flag |= exportHasEnv
	}
	if hasRet {
		conv, err := InFuncFor(t.Out(0))
		if err != nil {
			panic(fmt.Errorf("function %s: don’t know how to convert result: %s", d.name, err))
		}
		d.outConv = conv
	}
	if hasErr {
		d.flag |= exportHasErr
	}
	return d.name, d.call, arity, d.doc
}

// AutoLambda returns a Lambda object that exports the given function to Emacs
// as an anonymous lambda function.  When calling the exported function from
// Emacs, arguments and return values are converted as described in the package
// documentation.
//
// The function may accept any number of arguments.  Optionally, the first
// argument may be of type Env.  In this case, Emacs passes a live environment
// value that you can use to interact with Emacs.  All other arguments are
// converted from Emacs as described in the package documentation.  If not all
// arguments are convertible from Emacs values, AutoLambda panics.
//
// The function must return either zero, one, or two results.  If the last or
// only result is of type error, a non-nil value causes Emacs to trigger a
// non-local exit as appropriate.  There may be at most one non-error result.
// Its value will be converted to an Emacs value as described in the package
// documentation.  If the type of the non-error result can’t be converted to an
// Emacs value, AutoLambda panics.  If there are invalid result patterns,
// AutoLambda panics.
//
// The function is always anonymous.  Any Name option in opts is ignored.
//
// By default, the function has no documentation string.  To add one, pass a
// Doc option.
//
// You can call AutoLambda safely from multiple goroutines.
func AutoLambda(f interface{}, opts ...Option) Lambda {
	_, fun, arity, doc := AutoFunc(f, opts...)
	return Lambda{fun, arity, doc}
}

// Lambda represents an anonymous function.  When converting to Emacs, it
// exports Fun as an Emacs lambda function.
type Lambda struct {
	Fun   Func
	Arity Arity
	Doc   Doc
}

// Emacs returns a new function object for the given Go function.
func (l Lambda) Emacs(e Env) (Value, error) {
	return e.ExportFunc("", l.Fun, l.Arity, l.Doc)
}

// Lambda exports the given function to Emacs as an anonymous lambda function.
// Unlike the global AutoLambda function, Env.Lambda requires a live
// environment and defines the Emacs function immediately.  When calling the
// exported function from Emacs, arguments and return values are converted as
// described in the package documentation and the documentation for the
// AutoLambda function.
//
// When you don’t need the function any more, unregister it by calling the
// returned DeleteFunc function (typically using defer).  If you don’t call the
// delete function, the function will remain registered and require a bit of
// memory.  After calling the delete function, calling the function from Emacs
// panics.
//
// The function is always anonymous.  Any Name option in opts is ignored.
//
// By default, the function has no documentation string.  To add one, pass a
// Doc option.
//
// You can call Lambda safely from multiple goroutines.
func (e Env) Lambda(fun interface{}, opts ...Option) (Value, DeleteFunc, error) {
	l := AutoLambda(fun, opts...)
	return e.LambdaFunc(l.Fun, l.Arity, l.Doc)
}

// LambdaFunc exports the given function to Emacs as an anonymous lambda
// function.  Unlike the global Lambda function, Env.LambdaFunc requires a live
// environment and defines the Emacs function immediately.  Unlike Lambda,
// functions registered by LambdaFunc don’t automatically convert their
// arguments and return values to and from Emacs.
//
// When you don’t need the function any more, unregister it by calling the
// returned DeleteFunc function (typically using defer).  If you don’t call the
// delete function, the function will remain registered and require a bit of
// memory.  After calling the delete function, calling the function from Emacs
// panics.
//
// You can call LambdaFunc safely from multiple goroutines.
func (e Env) LambdaFunc(fun Func, arity Arity, doc Doc) (Value, DeleteFunc, error) {
	f := &function{Lambda{fun, arity, doc}, "", 0}
	if err := funcs.register(f); err != nil {
		return Value{}, nil, err
	}
	v, err := f.define(e)
	if err != nil {
		return Value{}, nil, err
	}
	return v, func() { funcs.delete(f.index) }, nil
}

// DeleteFunc is a function returned by Env.Lambda and Env.LambdaFunc.  Call
// this function to delete the created function.  After deletion the function
// can’t be called any more from Emacs.
type DeleteFunc func()

// Option is an option for Export, AutoFunc, AutoLambda, and ERTTest.  Its
// implementations are Name, Anonymous, Doc, and Usage.
type Option interface {
	apply(*exportAuto)
}

// Anonymous is an Option that tells AutoFunc and friends that the new function
// should be anonymous.  Anonymous is mutually exclusive with Name; if both are
// given, AutoFunc panics.
type Anonymous struct{}

func (Anonymous) apply(o *exportAuto) { o.flag |= exportAnonymous }
func (n Name) apply(o *exportAuto)    { o.name = n }
func (d Doc) apply(o *exportAuto)     { o.doc = d }
func (u Usage) apply(o *exportAuto)   { o.doc = o.doc.WithUsage(u) }

type exportAuto struct {
	fun     reflect.Value
	flag    exportFlag
	name    Name
	doc     Doc
	inConv  []OutFunc
	varConv OutFunc
	outConv InFunc
}

type exportFlag uint

const (
	exportAnonymous exportFlag = 1 << iota
	exportHasEnv
	exportHasErr
)

func lispName(fun reflect.Value) Name {
	name := runtime.FuncForPC(fun.Pointer()).Name()
	name = goIdentPattern.FindString(name)
	if name == "" {
		panic("can’t determine name of function")
	}
	var b strings.Builder
	for i, r := range name {
		if i > 0 && unicode.IsUpper(r) {
			b.WriteByte('-')
		}
		b.WriteRune(unicode.ToLower(r))
	}
	return Name(b.String())
}

// See https://go.dev/ref/spec#Identifiers.
var goIdentPattern = regexp.MustCompile(`[\p{L}_][\p{L}_\p{Nd}]*$`)

var (
	envType   = reflect.TypeOf(Env{})
	errorType = reflect.TypeOf((*error)(nil)).Elem()
)

func (d exportAuto) call(e Env, args []Value) (Value, error) {
	t := d.fun.Type()
	offset := 0
	if d.flag&exportHasEnv != 0 {
		offset = 1
	}
	in := make([]reflect.Value, len(args)+offset)
	if offset == 1 {
		in[0] = reflect.ValueOf(e)
	}
	numIn := len(d.inConv)
	for i, a := range args {
		j := i + offset
		var conv OutFunc
		var u reflect.Type
		if i < numIn {
			conv = d.inConv[i]
			u = t.In(j)
		} else {
			conv = d.varConv
			u = t.In(t.NumIn() - 1).Elem()
		}
		r := reflect.New(u)
		if err := conv(r).FromEmacs(e, a); err != nil {
			return Value{}, err
		}
		in[j] = r.Elem()
	}
	out := d.fun.Call(in)
	if d.flag&exportHasErr != 0 {
		if err, ok := out[len(out)-1].Interface().(error); ok && err != nil {
			return Value{}, err
		}
	}
	if d.outConv != nil {
		return d.outConv(out[0]).Emacs(e)
	}
	return e.Nil()
}

type function struct {
	Lambda
	name  Name
	index funcIndex
}

func (f *function) Define(e Env) error {
	_, err := f.define(e)
	return err
}

func (f *function) define(e Env) (Value, error) {
	v, err := e.makeFunction(f.Arity, f.Doc, uint64(f.index))
	if err != nil {
		return Value{}, err
	}
	if f.name != "" {
		if err := e.Defalias(f.name, v); err != nil {
			return Value{}, err
		}
	}
	return v, nil
}

// funcIndex is an index into funcManager.funcs.  It identifies an exported
// function.  See https://github.com/golang/go/wiki/cgo#function-variables why
// we need to use an index instead of a pointer.
type funcIndex uint64

type funcManager struct {
	mu    sync.RWMutex
	base  *Manager
	funcs map[funcIndex]Func
	next  funcIndex
}

func (m *funcManager) enqueue(f *function) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	// We split the actions performed under the lock into a preparation
	// phase (which shouldn’t modify state) and a registration phase (which
	// shouldn’t fail) to simulate a transaction across this object and the
	// base Manager.
	if err := m.prepareLocked(); err != nil {
		return err
	}
	if err := m.base.Enqueue(f.name, f); err != nil {
		return err
	}
	m.registerLocked(f)
	return nil
}

func (m *funcManager) register(f *function) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if err := m.prepareLocked(); err != nil {
		return err
	}
	m.registerLocked(f)
	return nil
}

func (m *funcManager) prepareLocked() error {
	if m.next == math.MaxUint64 {
		// This is very unlikely, but could happen if users create and
		// leak lots of functions in an unbounded loop.
		return errors.New("too many functions")
	}
	return nil
}

func (m *funcManager) registerLocked(f *function) {
	index := m.next
	m.next++
	if m.funcs == nil {
		m.funcs = make(map[funcIndex]Func)
	}
	m.funcs[index] = f.Fun
	f.index = index
}

func (m *funcManager) mustEnqueue(f *function) {
	if err := m.enqueue(f); err != nil {
		panic(err)
	}
}

func (m *funcManager) get(i funcIndex) Func {
	m.mu.RLock()
	defer m.mu.RUnlock()
	fun, ok := m.funcs[i]
	if !ok {
		panic(fmt.Errorf("attempt to access deleted function with index %d", i))
	}
	return fun
}

func (m *funcManager) delete(i funcIndex) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.funcs, i)
}

var funcs = funcManager{base: NewManager(RequireUniqueName | DefineOnInit)}
