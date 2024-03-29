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

/*
Package emacs contains infrastructure to write dynamic modules for Emacs in Go.
See [Emacs Dynamic Modules] and [Writing Dynamically-Loaded Modules] for
background on Emacs modules.

To build an Emacs module, you have to build your Go code as a shared C library,
e.g., using go build ‑buildmode=c‑shared.  If you import the emacs package, the
shared library is loadable as an Emacs module.

This package contains high-level as well as lower-level functions.  The
high-level functions help reducing boilerplate when exporting functions to
Emacs and calling Emacs functions from Go.  The lower-level functions are more
type-safe, support more exotic use cases, and have less overhead.

# Export and Import

At the highest level, use the [Export] function to export Go functions to
Emacs, and the [Import] function to import Emacs functions so that they can be
called from Go.  These functions automatically convert between Go and Emacs
types as necessary.  This export functionality is unrelated to exported Go
names or the Cgo export functionality.  Functions exported to Emacs don’t have
to be exported in the Go or Cgo sense.

The automatic type conversion behaves as follows.  Go bool values are become
the Emacs symbols nil and t.  When converting to Go bool, only nil becomes
false, any other value becomes true.  This matches the Emacs convention that
all non-nil values represent a logically true value.  Go integral values become
Emacs integer values and vice versa.  Go floating-point values become Emacs
floating-point values and vice versa.  Go strings become Emacs strings and vice
versa.  Go []byte arrays and slices become Emacs unibyte strings.  Emacs
unibyte strings become Go []byte slices.  Other Go arrays and slices become
Emacs vectors.  Emacs vectors become Go slices.  Go maps become Emacs hash
tables and vice versa.  All types that implement [In] can be converted to
Emacs.  All types that implement [Out] can be converted from Emacs.  You can
implement [In] or [Out] yourself to extend the type conversion machinery.  A
[reflect.Value] behaves like its underlying value.

Functions exported via [Export] don’t have a documentation string by default.
To add one, pass a [Doc] value to [Export].  Since argument names aren’t
available at runtime, the documentation by default lacks argument names.  Use
[Usage] to add argument names.

As an alternative to [Import], you can call functions directly using
[Env.Invoke].  [Env.Invoke] uses the same autoconversion rules as [Import], but
allows you to specify an arbitrary function value.

At a slightly lower level, you can use [Env.Call] and [Env.CallOut] to call
Emacs functions.  These functions use the [In] and [Out] interfaces to convert
from and to Emacs values.  The primary disadvantage of this approach is that
you can’t use primitive types like int or string directly.  Use wrapper types
like [Int] and [String] instead.  On the other hand, [Env.Call] and
[Env.CallOut] are more type-safe than [Invoke].  If you use [Call] or
[CallOut], the compiler will detect unsupported types.  By contrast, when using
[Export], [Import], or [Invoke], they will only be detected at runtime and
cause runtime panics or errors.

To reduce boilerplate when using [Env.Call] and [Env.CallOut], this package
contains several convenience types that implement [In] or [Out].  Most
primitive types have corresponding wrapper types, such as [Int], [Float], or
[String].  Types such as [List], [Cons], or [Hash] allow you to pass common
Lisp structures without much boilerplate.  There are also some destructuring
types such as [ListOut] or [Uncons].

At an even lower level, you can use [ExportFunc], [ImportFunc], and
[Env.Funcall] as alternatives to [Export], [Import], and [Env.Call],
respectively.  They have the same behavior, but don’t do any type conversion at
all.

# Env and Value

The fundamental types for interacting with Emacs are [Env] and [Value].  They
represent Emacs module environments and values as described in [Writing Module
Functions].  These types are opaque, and their zero values are invalid.  You
can’t use [Env] and [Value] values once they are no longer live.  This is
described in [Writing Module Functions] and [Conversion Between Lisp and Module
Values].  As a best practice, don’t let these values escape exported functions.
You also can’t interact with Emacs from other threads, cf. [Writing Module
Functions].  These rules are a bit subtle, but you are usually on the safe side
if you don’t store [Env] and [Value] values in struct fields or global
variables, and don’t pass them to other goroutines.

# Error handling

All functions in this package translate between Go errors and Emacs nonlocal
exits.  See [Nonlocal Exits in Modules].  This package represents Emacs
nonlocal exits as ordinary Go errors.

Each call to a function fetches and clears nonlocal exit information after the
actual call and converts it to an error of type [Signal] or [Throw].  This
means that the Go bindings don’t exhibit the saturating error behavior
described at [Nonlocal Exits in Modules].  Instead, they behave like normal Go
functions: an erroneous return doesn’t affect future function calls.

When returning from an exported function, this package converts errors back to
Emacs nonlocal exits.  If you return a [Signal] or [Error], Emacs will raise a
signal using the signal function.  If you return a [Throw], Emacs will throw to
a catch using the throw function.  If you return any other type of error, Emacs
will signal an error of type go‑error, with the error string as signal data.

You can define your own error symbols using [DefineError].  There are also a
couple of factory functions for builtin errors such as [WrongTypeArgument] and
[OverflowError].

# Variables

You can use [Var] to define a dynamic variable.

# User pointers and handles

This package intentionally doesn’t support wrapping pointers to arbitrary Go
values in Emacs user pointer objects.  Attempting to do that wouldn’t work well
with Go’s garbage collection and CGo’s pointer-passing rules; see [Passing
pointers].  Instead, prefer using handles, e.g. simple integers as map keys.
See the “Handles” example.

# Long-running operations

A long-running operation should periodically call [Env.ProcessInput] to process
pending input and to check whether the user wants to quit the operation.  If
so, you should cancel the operation as soon as possible.  See the documentation
of [Env.ProcessInput] for a concrete example.

As an alternative, this package provides limited support for asynchronous
operations.  Such operations are represented using the [AsyncHandle] type.  You
can use the [Async] type to create and manage asynchronous operations.  [Async]
requires a way to notify Emacs about a pending asynchronous result; this
package supports notification using pipes or sockets.

# Initialization

If you want to run code while Emacs is loading the module, use [OnInit] to
register initialization functions.  Loading the module will call all
initialization functions in order.

# ERT tests

You can use [ERTTest] to define ERT tests backed by Go functions.  This works
similar to [Export], but defines ERT tests instead of functions.

[Emacs Dynamic Modules]: https://www.gnu.org/software/emacs/manual/html_node/elisp/Dynamic-Modules.html
[Writing Dynamically-Loaded Modules]: https://www.gnu.org/software/emacs/manual/html_node/elisp/Writing-Dynamic-Modules.html
[Writing Module Functions]: https://www.gnu.org/software/emacs/manual/html_node/elisp/Module-Functions.html
[Conversion Between Lisp and Module Values]: https://www.gnu.org/software/emacs/manual/html_node/elisp/Module-Values.html
[Nonlocal Exits in Modules]: https://www.gnu.org/software/emacs/manual/html_node/elisp/Module-Nonlocal.html
[Passing pointers]: https://pkg.go.dev/cmd/cgo#hdr-Passing_pointers
*/
package emacs
