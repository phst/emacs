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
// bool is_not_nil(emacs_env *env, emacs_value value) {
//   return env->is_not_nil(env, value);
// }
import "C"

// Bool is a type with underlying type bool that knows how to convert itself to
// and from an Emacs value.
type Bool bool

// Common symbols.
const (
	Nil Symbol = "nil"
	T   Symbol = "t"
)

// Symbol returns Nil or T depending on the value of b.
func (b Bool) Symbol() Symbol {
	if b {
		return T
	}
	return Nil
}

// Emacs returns t if v is true and nil if it’s false.  It can only fail if
// interning nil or t fails.
func (b Bool) Emacs(e Env) (Value, error) {
	return b.Symbol().Emacs(e)
}

// FromEmacs sets *b to false if v is nil and to true otherwise.  It never
// fails.
func (b *Bool) FromEmacs(e Env, v Value) error {
	*b = Bool(e.IsNotNil(v))
	return nil
}

// IsNotNil returns false if and only if the given Emacs value is nil.
func (e Env) IsNotNil(v Value) bool {
	return bool(C.is_not_nil(e.raw(), v.r))
}

// IsNil returns true if and only if the given Emacs value is nil.
func (e Env) IsNil(v Value) bool {
	return !e.IsNotNil(v)
}
