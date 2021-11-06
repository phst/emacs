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

// #include "emacs-module.h"
// #include "wrappers.h"
import "C"

import "runtime"

// InitFunc is an initializer function that should be run during module
// initialization.  Use OnInit to register InitFunc functions.  If an
// initialization function returns an error, the module loading itself will
// fail.
type InitFunc func(Env) error

// OnInit arranges for the given function to run while Emacs is loading the
// module.  Initialization functions registered with OnInit will be called in
// sequence, in the same order in which they’ve been registered.  You need to
// call OnInit before loading the module for the initializer to run.
// Typically, you should call OnInit in an init function.  You can call OnInit
// safely from multiple goroutines.
func OnInit(i InitFunc) {
	if i == nil {
		panic("nil initializer")
	}
	inits.MustEnqueue("", initFunc(i))
}

var inits Manager

type initFunc InitFunc

func (i initFunc) Define(e Env) error {
	return i(e)
}

//export phst_emacs_init
func phst_emacs_init(env *C.emacs_env) (r C.struct_phst_emacs_init_result) {
	// We can’t use environments from other threads, so make sure that we
	// don’t switch threads.  See https://phst.eu/emacs-modules#threads.
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	e := Env{env}
	// Don’t allow Go panics to crash Emacs.
	defer protect(e, &r.base)
	if err := majorVersion.init(e); err != nil {
		return C.struct_phst_emacs_init_result{e.signal(err)}
	}
	err := inits.DefineQueued(e)
	return C.struct_phst_emacs_init_result{e.signal(err)}
}

//export plugin_is_GPL_compatible
func plugin_is_GPL_compatible() { panic("unused") }
