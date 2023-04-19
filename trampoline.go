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

// #include "emacs-module.h"
// #include "wrappers.h"
import "C"

import (
	"fmt"
	"runtime"
	"runtime/debug"
	"unsafe"
)

// The trampoline must be defined in a separate file.  See
// https://pkg.go.dev/cmd/cgo#hdr-C_references_to_Go.

//export phst_emacs_trampoline
func phst_emacs_trampoline(env *C.emacs_env, nargs C.int64_t, args *C.emacs_value, data C.uint64_t) (r C.struct_phst_emacs_trampoline_result) {
	// We can’t use environments from other threads, so make sure that we
	// don’t switch threads.  See
	// https://www.gnu.org/software/emacs/manual/html_node/elisp/Module-Functions.html.
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	e := Env{env}
	// Don’t allow Go panics to crash Emacs.
	defer protect(e, &r.base)
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
	v, err := fun(e, in)
	return C.struct_phst_emacs_trampoline_result{e.signal(err), v.r}
}

//export phst_emacs_function_finalizer
func phst_emacs_function_finalizer(data C.uint64_t) {
	funcs.delete(funcIndex(data))
}

func protect(e Env, r *C.struct_result_base_with_optional_error_info) {
	if x := recover(); x != nil {
		// Because to Go runtime calls deferred functions in the call
		// frame of the panic, this stack trace will be useful.  This
		// stack trace and the Emacs stack trace that the backtrace
		// function prints combine nicely: The Go stack trace starts at
		// the Cgo entry point, the Emacs stack trace ends at the
		// module interface.
		debug.PrintStack()
		*r = e.signal(errPanic.Error(String(fmt.Sprint(x))))
	}
}

var errPanic = DefineError("go-panic", "Panic while running Emacs module function", baseError)
