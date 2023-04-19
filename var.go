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

// Var arranges for an Emacs dynamic variable to be defined once the module is
// loaded.  If doc is empty, the variable won’t have a documentation string.
// Var panics if the name is empty or already registered.  Var returns name so
// you can assign it directly to a Go variable if you want.
func Var(name Name, init In, doc Doc) Name {
	vars.MustEnqueue(name, variable{name, init, doc})
	return name
}

// Var is like the global [Var] function, except that it requires a live
// environment, defines the variable immediately, and returns errors instead of
// panicking.
func (e Env) Var(name Name, init In, doc Doc) error {
	v := variable{name, init, doc}
	return vars.RegisterAndDefine(e, name, v)
}

// Defvar calls the Emacs special form defvar.
func (e Env) Defvar(name Name, init In, doc Doc) error {
	// Can’t use Call because defvar is not a function.
	_, err := e.Eval(List{Symbol("defvar"), name, init, doc})
	return err
}

// Let locally binds the variable to the value within body.  This is like the
// Emacs Lisp let special form.  Let returns the value and error returned by
// body, unless some other error occurs.
func (e Env) Let(variable Name, value In, body func() (Value, error)) (Value, error) {
	return e.LetMany([]Binding{{variable, value}}, body)
}

// LetMany locally binds the given variables value within body.  This is like
// the Emacs Lisp let special form.  Let returns the value and error returned
// by body, unless some other error occurs.
func (e Env) LetMany(bindings []Binding, body func() (Value, error)) (Value, error) {
	fun, delete, err := e.Lambda(body)
	if err != nil {
		return Value{}, err
	}
	defer delete()
	binds := make(List, len(bindings))
	for i, b := range bindings {
		binds[i] = List{b.Variable, b.Value}
	}
	return e.Call("eval", List{
		Symbol("let"),
		binds,
		List{Symbol("funcall"), fun},
	}, T)
}

// Binding describes a variable binding for [LetMany].
type Binding struct {
	Variable Name
	Value    In
}

type variable struct {
	name Name
	init In
	doc  Doc
}

func (v variable) Define(e Env) error {
	return e.Defvar(v.name, v.init, v.doc)
}

var vars = NewManager(RequireName | RequireUniqueName | DefineOnInit)
