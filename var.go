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

// Var arranges for an Emacs dynamic variable to be defined once the module is
// loaded.  If doc is empty, the variable won’t have a documentation string.
// Var panics if the name is empty or already registered.  Var returns name so
// you can assign it directly to a Go variable if you want..
func Var(name Name, init In, doc Doc) Name {
	vars.mustRegister(variable{name, init, doc})
	return name
}

// Var is like the global Var function, except that it requires a live
// environment, defines the variable immediately, and returns errors instead of
// panicking.
func (e Env) Var(name Name, init In, doc Doc) error {
	v := variable{name, init, doc}
	if err := vars.register(v); err != nil {
		return err
	}
	return v.define(e)
}

// Defvar calls the Emacs special form defvar.
func (e Env) Defvar(name Name, init In, doc Doc) error {
	// Can’t use Call because defvar is not a function.
	_, err := e.Eval(List{Symbol("defvar"), name, init, doc})
	return err
}

type variable struct {
	name Name
	init In
	doc  Doc
}

func (v variable) define(e Env) error {
	return e.Defvar(v.name, v.init, v.doc)
}

type varManager struct {
	mu    sync.Mutex
	vars  []variable
	names map[Name]struct{}
}

func (m *varManager) register(v variable) error {
	if err := v.name.validate(); err != nil {
		return err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, dup := m.names[v.name]; dup {
		return fmt.Errorf("duplicate variable name %s", v.name)
	}
	m.vars = append(m.vars, v)
	if m.names == nil {
		m.names = make(map[Name]struct{})
	}
	m.names[v.name] = struct{}{}
	return nil
}

func (m *varManager) mustRegister(v variable) {
	if err := m.register(v); err != nil {
		panic(err)
	}
}

func (m *varManager) define(e Env) error {
	for _, v := range m.copy() {
		if err := v.define(e); err != nil {
			return err
		}
	}
	return nil
}

func (m *varManager) copy() []variable {
	m.mu.Lock()
	defer m.mu.Unlock()
	r := make([]variable, len(m.vars))
	copy(r, m.vars)
	return r
}

var vars varManager

func init() {
	OnInit(vars.define)
}
