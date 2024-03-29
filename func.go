// Copyright 2019, 2021, 2023, 2024 Google LLC
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
// struct phst_emacs_value_result phst_emacs_make_function(emacs_env *env,
//                                                         int64_t min_arity,
//                                                         int64_t max_arity,
//                                                         _GoString_ documentation,
//                                                         uint64_t data) {
//   size_t length = _GoStringLen(documentation);
//   const char *doc = length == 0 ? NULL : _GoStringPtr(documentation);
//   return phst_emacs_make_function_impl(env, min_arity, max_arity, doc, data);
// }
import "C"

import (
	"fmt"
	"regexp"
	"strings"
)

// Doc contains a documentation string for a function or variable.  An empty
// doc string becomes nil.  As described in [Documentation Strings of
// Functions], a documentation string can contain usage information.  Use
// [Doc.SplitUsage] to extract the usage information from a documentation
// string.  Use [Doc.WithUsage] to add usage information to a documentation
// string.  Documentation strings must be valid UTF-8 strings without embedded
// null bytes.  You can use a Doc as an [Option] in [Export] and [ERTTest] to
// set the function or test documentation string.
//
// [Documentation Strings of Functions]: https://www.gnu.org/software/emacs/manual/html_node/elisp/Function-Documentation.html
type Doc string

// Emacs returns nil if d is empty and an Emacs string otherwise.
func (d Doc) Emacs(e Env) (Value, error) {
	if d == "" {
		return e.Nil()
	}
	return String(d).Emacs(e)
}

func (d Doc) validate() error {
	if isNonNullUTF8(string(d)) {
		return nil
	}
	return WrongTypeArgument("valid-string-p", d)
}

// SplitUsage splits d into the actual docstring and the usage information.
// hasUsage specifies whether a usage information is present.  Absence of usage
// information is not the same as an empty usage.
func (d Doc) SplitUsage() (actualDoc Doc, hasUsage bool, usage Usage) {
	s := string(d)
	result := usagePattern.FindStringSubmatchIndex(s)
	if result == nil {
		return d, false, ""
	}
	actualDoc = d[:result[0]]
	hasUsage = true
	if i, j := result[2], result[3]; i >= 0 {
		usage = Usage(strings.Trim(s[i:j], " "))
	}
	return
}

// See the implementation of the Emacs function help-split-fundoc.
var usagePattern = regexp.MustCompile(`\n\n\(fn( .*)?\)$`)

// WithUsage returns d with the usage string appended.  If d already contains
// usage information, WithUsage replaces it.
func (d Doc) WithUsage(u Usage) Doc {
	if err := u.validate(); err != nil {
		panic(fmt.Errorf("invalid usage %q: %v", u, err))
	}
	s := strings.Trim(string(u), " ")
	if s != "" {
		s = " " + s
	}
	d, _, _ = d.SplitUsage()
	// See
	// https://www.gnu.org/software/emacs/manual/html_node/elisp/Function-Documentation.html.
	return Doc(fmt.Sprintf("%s\n\n(fn%s)", d, s))
}

// Usage contains a list of argument names to be added to a documentation
// string.  It should contain a plain space-separated list of argument names
// without enclosing parentheses.  See [Documentation Strings of Functions].
// Usage strings must be valid UTF-8 strings without embedded null characters
// or newlines.  You can use a Usage as an [Option] in [Export] to set the
// argument list in the documentation string.
//
// [Documentation Strings of Functions]: https://www.gnu.org/software/emacs/manual/html_node/elisp/Function-Documentation.html
type Usage string

func (u Usage) validate() error {
	if isNonNullUTF8(string(u)) && strings.IndexByte(string(u), '\n') < 0 {
		return nil
	}
	return WrongTypeArgument("valid-string-p", String(u))
}

// Arity contains how many arguments an Emacs function accepts.  Min must be
// nonnegative.  Max must either be negative (indicating a variadic function)
// or at least Min.
type Arity struct{ Min, Max int }

// Variadic returns whether the function is variadic, i.e., whether a.Max is
// negative.
func (a Arity) Variadic() bool {
	return a.Max < 0
}

// Func is a Go function exported to Emacs.  It has access to a live
// environment, takes arguments as a slice, and can return a value or an error.
type Func func(Env, []Value) (Value, error)

// Defalias calls the Emacs function defalias.
func (e Env) Defalias(name Name, def Value) error {
	_, err := e.Call("defalias", name, def)
	return err
}

func (e Env) makeFunction(arity Arity, doc Doc, data uint64) (Value, error) {
	min := C.int64_t(arity.Min)
	var max C.int64_t
	if arity.Variadic() {
		max = C.emacs_variadic_function
	} else {
		max = C.int64_t(arity.Max)
	}
	if doc != "" {
		if err := doc.validate(); err != nil {
			return Value{}, err
		}
		doc += "\x00"
	}
	return e.checkValue(C.phst_emacs_make_function(e.raw(), min, max, string(doc), C.uint64_t(data)))
}

// Funcall calls the Emacs function fun with the given arguments.  Both
// function and arguments must be Emacs values.  Use [Env.Call] or [Env.Invoke]
// if you want them to be autoconverted.
func (e Env) Funcall(fun Value, args []Value) (Value, error) {
	nargs := len(args)
	var ptr *C.emacs_value
	if nargs > 0 {
		rawArgs := make([]C.emacs_value, nargs)
		for i, a := range args {
			rawArgs[i] = a.r
		}
		ptr = &rawArgs[0]
	}
	return e.checkValue(C.phst_emacs_funcall(e.raw(), fun.r, C.int64_t(nargs), ptr))
}

// MakeInteractive sets the interactive specification of the given function.
// The function must refer to a module function.
func (e Env) MakeInteractive(fun, spec Value) error {
	return e.checkVoid(C.phst_emacs_make_interactive(e.raw(), fun.r, spec.r))
}
