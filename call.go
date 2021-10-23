// Copyright 2019, 2021 Google LLC
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

// Call calls the Emacs named fun with the given arguments.
func (e Env) Call(fun Name, in ...In) (Value, error) {
	fv, err := e.Intern(Symbol(fun))
	if err != nil {
		return Value{}, err
	}
	vals := make([]Value, len(in))
	for i, a := range in {
		if a == nil {
			a = Nil
		}
		vals[i], err = a.Emacs(e)
		if err != nil {
			return Value{}, err
		}
	}
	return e.Funcall(fv, vals)
}

// CallOut calls a the Emacs function named fun with the given arguments and
// assigns the result to out.
func (e Env) CallOut(fun Name, out Out, in ...In) error {
	v, err := e.Call(fun, in...)
	if err != nil {
		return err
	}
	return out.FromEmacs(e, v)
}

// Invoke calls a named Emacs function or function value.  fun may be a string,
// Symbol, Name, or Value.  If it’s not a value, Invoke interns it first.
// Invoke then calls the Emacs functions with the given arguments and assigns
// the result to out.  It converts arguments and the return value as described
// in the package documentation.
func (e Env) Invoke(fun interface{}, out interface{}, in ...interface{}) error {
	fv, err := e.MaybeIntern(fun)
	if err != nil {
		return err
	}
	vals := make([]Value, len(in))
	for i, a := range in {
		vals[i], err = e.Emacs(a)
		if err != nil {
			return err
		}
	}
	rv, err := e.Funcall(fv, vals)
	if err != nil {
		return err
	}
	return e.Go(rv, out)
}
