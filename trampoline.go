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
	"fmt"
	"runtime"
	"runtime/debug"
	"unsafe"
)

// The trampoline must be defined in a separate file.  See
// https://golang.org/cmd/cgo/#hdr-C_references_to_Go.

//export go_emacs_trampoline
func go_emacs_trampoline(env *C.emacs_env, nargs C.int64_t, args *C.emacs_value, data C.uint64_t) C.emacs_value {
	// We can’t use environments from other threads, so make sure that we
	// don’t switch threads.  See https://phst.eu/emacs-modules#threads.
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	e := Env{env}
	// Don’t allow Go panics to crash Emacs.
	defer protect(e)
	fun := funcs.get(funcIndex(data))
	var in []Value
	if nargs > 0 {
		// See
		// https://github.com/golang/go/wiki/cgo#turning-c-arrays-into-go-slices.
		argSlice := (*[1 << 40]C.emacs_value)(unsafe.Pointer(args))[:nargs:nargs]
		in = make([]Value, nargs)
		for i, a := range argSlice {
			in[i] = Value{a}
		}
	}
	r, err := callGo(e, fun, in)
	if err != nil {
		e.signal(err)
		// No normal or deferred call using e is allowed at this point.
		// That’s why we use the callGo indirection to defer
		// gcContext.restore.
	}
	return r.r
}

func protect(e Env) {
	if x := recover(); x != nil {
		// Because to Go runtime calls deferred functions in the call
		// frame of the panic, this stack trace will be useful.  This
		// stack trace and the Emacs stack trace that the backtrace
		// function prints combine nicely: The Go stack trace starts at
		// the Cgo entry point, the Emacs stack trace ends at the
		// module interface.
		debug.PrintStack()
		e.signal(errPanic.Error(String(fmt.Sprint(x))))
	}
}

var errPanic = DefineError("go-panic", "Panic while running Emacs module function", baseError)

func callGo(e Env, fun Func, in []Value) (Value, error) {
	// Inhibit Emacs garbage collector on Emacs 26 and below to work around
	// https://debbugs.gnu.org/cgi/bugreport.cgi?bug=31238.
	defer gc.inhibit(e).restore(e)
	return fun(e, in)
}
