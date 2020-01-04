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
	funcs.mustRegister(function{name, fun, arity, doc})
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
// different name, pass a Name option.  If there’s no name or the name is
// already registered, Export panics.
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
	f := function{name, fun, arity, doc}
	i, err := funcs.register(f)
	if err != nil {
		return Value{}, err
	}
	return f.define(e, i)
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
// different name, pass a Name option.  If there’s no name, AutoFunc panics.
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
	if d.name == "" {
		d.name = lispName(v)
	}
	t := v.Type()
	numIn := t.NumIn()
	d.hasEnv = numIn > 0 && t.In(0) == envType
	offset := 0
	if d.hasEnv {
		offset = 1
	}
	var arity Arity
	if t.IsVariadic() {
		numIn--
		arity.Min = numIn - offset
		arity.Max = -1
		conv, err := OutFuncFor(t.In(numIn).Elem())
		if err != nil {
			panic(fmt.Errorf("function %s: don’t know how to convert variadic type: %s", d.name, err))
		}
		d.varConv = conv
	} else {
		arity.Min = numIn - offset
		arity.Max = arity.Min
	}
	for i := offset; i < numIn; i++ {
		conv, err := OutFuncFor(t.In(i))
		if err != nil {
			panic(fmt.Errorf("function %s: don’t know how to convert argument %d: %s", d.name, i, err))
		}
		d.inConv = append(d.inConv, conv)
	}
	hasRet := false
	switch t.NumOut() {
	case 0:
	case 1:
		d.hasErr = t.Out(0) == errorType
		hasRet = !d.hasErr
	case 2:
		if t.Out(1) != errorType {
			panic(fmt.Errorf("function %s: second result must be error, but is %s", d.name, t.Out(1)))
		}
		d.hasErr = true
		hasRet = true
	default:
		panic(fmt.Errorf("function %s: too many results", d.name))
	}
	if hasRet {
		conv, err := InFuncFor(t.Out(0))
		if err != nil {
			panic(fmt.Errorf("function %s: don’t know how to convert result: %s", d.name, err))
		}
		d.outConv = conv
	}
	return d.name, d.call, arity, d.doc
}

// Option is an option for Export and ERTTest.  Its implementations are Name,
// Doc, and Usage.
type Option interface {
	apply(*exportAuto)
}

func (n Name) apply(o *exportAuto)  { o.name = n }
func (d Doc) apply(o *exportAuto)   { o.doc = d }
func (u Usage) apply(o *exportAuto) { o.doc = o.doc.WithUsage(u) }

type exportAuto struct {
	fun            reflect.Value
	name           Name
	doc            Doc
	hasEnv, hasErr bool
	inConv         []OutFunc
	varConv        OutFunc
	outConv        InFunc
}

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

// See https://golang.org/ref/spec#Identifiers.
var goIdentPattern = regexp.MustCompile(`[\p{L}_][\p{L}_\p{Nd}]*$`)

var (
	envType   = reflect.TypeOf(Env{})
	errorType = reflect.TypeOf((*error)(nil)).Elem()
)

func (d exportAuto) call(e Env, args []Value) (Value, error) {
	t := d.fun.Type()
	in := make([]reflect.Value, t.NumIn())
	offset := 0
	if d.hasEnv {
		offset = 1
		in[0] = reflect.ValueOf(e)
	}
	numIn := len(d.inConv)
	for i, a := range args {
		j := i + offset
		var conv OutFunc
		if i < numIn {
			conv = d.inConv[i]
		} else {
			conv = d.varConv
		}
		r := reflect.New(t.In(j)).Elem()
		if err := conv(r).FromEmacs(e, a); err != nil {
			return Value{}, err
		}
		in[j] = r
	}
	out := d.fun.Call(in)
	if d.hasErr {
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
	name  Name
	fun   Func
	arity Arity
	doc   Doc
}

func (f function) define(e Env, index funcIndex) (Value, error) {
	v, err := e.makeFunction(f.arity, f.doc, uint64(index))
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
	funcs map[funcIndex]function
	next  funcIndex
	names map[Name]struct{}
}

func (m *funcManager) register(f function) (funcIndex, error) {
	if f.name == "" {
		return m.registerUnnamed(f)
	}
	return m.registerNamed(f)
}

func (m *funcManager) registerNamed(f function) (funcIndex, error) {
	if err := f.name.validate(); err != nil {
		return 0, err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, dup := m.names[f.name]; dup {
		return 0, fmt.Errorf("duplicate definition of %s", f.name)
	}
	index, err := m.registerLocked(f)
	if err != nil {
		return 0, err
	}
	if m.names == nil {
		m.names = make(map[Name]struct{})
	}
	m.names[f.name] = struct{}{}
	return index, nil
}

func (m *funcManager) registerUnnamed(f function) (funcIndex, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.registerLocked(f)
}

func (m *funcManager) registerLocked(f function) (funcIndex, error) {
	index := m.next
	if index == math.MaxUint64 {
		return 0, errors.New("too many functions")
	}
	m.next++
	if m.funcs == nil {
		m.funcs = make(map[funcIndex]function)
	}
	m.funcs[index] = f
	return index, nil
}

func (m *funcManager) mustRegister(f function) {
	if _, err := m.register(f); err != nil {
		panic(err)
	}
}

func (m *funcManager) get(i funcIndex) Func {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.funcs[i].fun
}

func (m *funcManager) define(e Env) error {
	for i, f := range m.copy() {
		if _, err := f.define(e, i); err != nil {
			return err
		}
	}
	return nil
}

func (m *funcManager) delete(i funcIndex) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.funcs, i)
}

func (m *funcManager) copy() map[funcIndex]function {
	m.mu.RLock()
	defer m.mu.RUnlock()
	r := make(map[funcIndex]function, len(m.funcs))
	for i, f := range m.funcs {
		r[i] = f
	}
	return r
}

var funcs funcManager

func init() {
	OnInit(funcs.define)
}
