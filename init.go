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

//export emacs_module_init
func emacs_module_init(rt *C.struct_emacs_runtime) C.int {
	// We can’t use environments from other threads, so make sure that we
	// don’t switch threads.  See https://phst.eu/emacs-modules#threads.
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	if rt.size < C.sizeof_struct_emacs_runtime {
		return 1
	}
	e := getEnv(rt)
	if e.ptr.size < C.sizeof_struct_emacs_env_26 {
		return 2
	}
	// Don’t allow Go panics to crash Emacs.
	defer protect(e)
	if err := inits.run(e); err != nil {
		e.signal(err)
		// We still return 0.  See
		// https://phst.eu/emacs-modules#module-loading-and-initialization.
	}
	return 0
}

//export plugin_is_GPL_compatible
func plugin_is_GPL_compatible() { panic("unused") }
