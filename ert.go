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
	"fmt"
	"sync"
)

// ERTTestFunc is a function that implements an ERT test.  Use ERTTest to
// register ERTTestFunc functions.  If the function returns an error, the ERT
// test fails.
type ERTTestFunc func(Env) error

// ERTTest arranges for a Go function to be exported as an ERT test.  Call
// ERTTest in an init function.  Loading the dynamic module will then define
// the ERT test.  If you want to define ERT tests after the module has been
// initialized, use the ERTTest method of Env instead.  If the function returns
// an error, the ERT test fails.
//
// By default, ERTTest derives the test’s name from the function’s Go name by
// Lisp-casing it.  For example, MyTest becomes my-test.  To specify a
// different name, pass a Name option.  If there’s no name or the name is
// already registered, ERTTest panics.
//
// By default, the ERT test has no documentation string.  To add one, pass a
// Doc option.
//
// You can call ERTTest safely from multiple goroutines.
func ERTTest(fun ERTTestFunc, opts ...Option) {
	name, fn, _, doc := AutoFunc(fun, opts...)
	ertTests.mustRegister(ertTest{name, fn, doc})
}

// ERTTest exports a Go function as an ERT test.  Unlike the global ERTTest
// function, Env.ERTTest requires a live environment and defines the ERT test
// immediately.  If the function returns an error, the ERT test fails.
//
// By default, ERTTest derives the test’s name from the function’s Go name by
// Lisp-casing it.  For example, MyTest becomes my-test.  To specify a
// different name, pass a Name option.  If there’s no name or the name is
// already registered, ERTTest panics.
//
// By default, the ERT test has no documentation string.  To add one, pass a
// Doc option.
func (e Env) ERTTest(fun ERTTestFunc, opts ...Option) error {
	name, fn, _, doc := AutoFunc(fun, opts...)
	t := ertTest{name, fn, doc}
	if err := ertTests.register(t); err != nil {
		return err
	}
	return t.define(e)
}

// ERTDeftest defines an ERT test with the given name and documentation string.
// The test calls the Go function fun.  It succeeds if fun returns nil.  This
// is the Go equivalent of the ert-deftest macro.
func (e Env) ERTDeftest(name Name, fun Func, doc Doc) error {
	// The Emacs function itself is anonymous and undocumented.
	f, err := e.ExportFunc("", fun, Arity{}, "")
	if err != nil {
		return err
	}
	// We don’t eval ert-deftest because its expansion is trivial and it
	// only uses the public interface of ERT (make-ert-test and
	// ert-set-test).
	args := []In{Symbol(":name"), name, Symbol(":body"), f}
	if doc != "" {
		args = append(args, Symbol(":documentation"), doc)
	}
	t, err := e.Call("make-ert-test", args...)
	if err != nil {
		return err
	}
	_, err = e.Call("ert-set-test", name, t)
	return err

}

type ertTest struct {
	name Name
	fun  Func
	doc  Doc
}

func (t ertTest) define(e Env) error {
	return e.ERTDeftest(t.name, t.fun, t.doc)
}

type ertTestManager struct {
	mu    sync.Mutex
	tests []ertTest
	names map[Name]struct{}
}

func (m *ertTestManager) register(t ertTest) error {
	if err := t.name.validate(); err != nil {
		return err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, dup := m.names[t.name]; dup {
		return fmt.Errorf("duplicate test name %s", t.name)
	}
	m.tests = append(m.tests, t)
	if m.names == nil {
		m.names = make(map[Name]struct{})
	}
	m.names[t.name] = struct{}{}
	return nil
}

func (m *ertTestManager) mustRegister(t ertTest) {
	if err := m.register(t); err != nil {
		panic(err)
	}
}

func (m *ertTestManager) define(e Env) error {
	for _, t := range m.copy() {
		if err := t.define(e); err != nil {
			return err
		}
	}
	return nil
}

func (m *ertTestManager) copy() []ertTest {
	m.mu.Lock()
	defer m.mu.Unlock()
	r := make([]ertTest, len(m.tests))
	copy(r, m.tests)
	return r
}

var ertTests ertTestManager

func init() {
	OnInit(ertTests.define)
}
