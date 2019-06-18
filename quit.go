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

// #include "wrappers.h"
import "C"

// ShouldQuit returns whether the user has requested a quit.  If ShouldQuit
// returns true, the caller should return to Emacs as soon as possible to allow
// Emacs to process the quit.  Once Emacs regains control, it will quit and
// ignore the return value.
//
// Deprecated: Use ProcessInput instead.
func (e Env) ShouldQuit() bool {
	return bool(C.should_quit(e.raw()))
}

// ProcessInput processes pending input and returns whether the user has
// requested a quit.  If ProcessInput returns Quit, the caller should return to
// Emacs as soon as possible to allow Emacs to process the quit.  Once Emacs
// regains control, it will quit and ignore the return value.  Note that
// processing input can run arbitrary Lisp code, so don’t rely on global state
// staying the same after calling ProcessInput.
func (e Env) ProcessInput() ProcessInput {
	return ProcessInput(C.process_input(e.raw()))
}

// ProcessInput is the possible return values of Env.ProcessInput.
type ProcessInput int

const (
	// Keep these constants in sync with emacs-module.h

	// Continue means that the caller may continue.
	Continue ProcessInput = 0

	// Quit means that the caller should return to Emacs as quickly as
	// possible.
	Quit ProcessInput = 1
)
