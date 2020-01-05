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
// #include "wrappers.h"
import "C"

import (
	"runtime"
	"sync"
)

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
	inits.register(i)
}

type initManager struct {
	mu    sync.Mutex
	inits []InitFunc
}

var inits initManager

func (m *initManager) register(i InitFunc) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.inits = append(m.inits, i)
}

func (m *initManager) run(e Env) error {
	if err := majorVersion.init(e); err != nil {
		return err
	}
	inits := m.copy()
	for _, i := range inits {
		if err := i(e); err != nil {
			return err
		}
	}
	return nil
}

func (m *initManager) copy() []InitFunc {
	m.mu.Lock()
	defer m.mu.Unlock()
	r := make([]InitFunc, len(m.inits))
	copy(r, m.inits)
	return r
}

//export go_emacs_init
func go_emacs_init(env *C.emacs_env) (r C.struct_init_result) {
	// We can’t use environments from other threads, so make sure that we
	// don’t switch threads.  See https://phst.eu/emacs-modules#threads.
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	e := Env{env}
	// Don’t allow Go panics to crash Emacs.
	defer protect(e, &r.base)
	// Inhibit Emacs garbage collector on Emacs 26 and below to work around
	// https://debbugs.gnu.org/cgi/bugreport.cgi?bug=31238.
	defer gc.inhibit(e).restore(e)
	err := inits.run(e)
	return C.struct_init_result{e.signal(err)}
}

//export plugin_is_GPL_compatible
func plugin_is_GPL_compatible() { panic("unused") }
