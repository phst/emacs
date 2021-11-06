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

// #include <stdbool.h>
// #include "emacs-module.h"
// bool phst_emacs_eq(emacs_env *env, emacs_value a, emacs_value b) {
//   return env->eq(env, a, b);
// }
import "C"

// Env represents an Emacs module environment.  The zero Env is not valid.
// Exported functions and module initializers will receive a valid Env value.
// That Env value only remains valid (or “live”) while the exported function or
// module initializer is active.  Env values are only valid in the same
// goroutine as the exported function or module initializer.  So don’t store
// them or pass them to other goroutines.  See
// https://phst.eu/emacs-modules#environments for details.
type Env struct{ ptr *C.emacs_env }

// Eq returns true if and only if the two values represent the same Emacs
// object.
func (e Env) Eq(a, b Value) bool {
	return a == b || bool(C.phst_emacs_eq(e.raw(), a.r, b.r))
}

// Eval evaluates form using the Emacs function eval.  The binding is always
// lexical.
func (e Env) Eval(form In) (Value, error) {
	return e.Call("eval", form, T)
}

func (e Env) raw() *C.emacs_env {
	if e.ptr == nil {
		panic("nil environment")
	}
	return e.ptr
}
