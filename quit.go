// Copyright 2019, 2021, 2023 Google LLC
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

// #include "wrappers.h"
import "C"

// ShouldQuit returns whether the user has requested a quit.  If ShouldQuit
// returns true, the caller should return to Emacs as soon as possible to allow
// Emacs to process the quit.  Once Emacs regains control, it will quit and
// ignore the return value.
//
// Deprecated: Use [Env.ProcessInput] instead.
func (e Env) ShouldQuit() bool {
	return bool(C.phst_emacs_should_quit(e.raw()))
}

// ProcessInput processes pending input and returns whether the user has
// requested a quit.  If ProcessInput returns an error, the caller should
// return the error to Emacs as soon as possible to allow Emacs to process the
// quit.  Once Emacs regains control, it will quit and ignore the return value.
// Note that processing input can run arbitrary Lisp code, so don’t rely on
// global state staying the same after calling ProcessInput.
func (e Env) ProcessInput() error {
	return e.checkVoid(C.phst_emacs_process_input(e.raw()))
}
