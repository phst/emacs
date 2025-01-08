# Go bindings for the Emacs module API

This package implements [Go][] bindings for the [GNU Emacs module API][].

This is not an officially supported Google product.

[Go]: https://go.dev/
[GNU Emacs module API]: (https://www.gnu.org/software/emacs/manual/html_node/elisp/Writing-Dynamic-Modules.html)

## Usage

See the [package documentation][].

[package documentation]: https://pkg.go.dev/github.com/phst/emacs

## Implementation notes

The package assumes that `ptrdiff_t` and `intmax_t` have the same domain as
`int64_t`.  It also assumes that `uintptr_t` has the same domain as `uint64_t`.
The C code uses `static_assert` to check these assumptions.  The C standard
[guarantees](https://en.cppreference.com/w/c/types/integer) that `int64_t` and
`uint64_t` have exactly 64 bits without padding bits and that they use two’s
complement representation.  The corresponding Go types `int64` and `uint64` have
the same [representation](https://go.dev/ref/spec#Numeric_types).  This means
that converting `int64_t` and `uint64_t` to and from the Go types can never lose
data.

The package requires at least Emacs 28.  It provides optional support for some
Emacs 29 features.  Such optional features always require two additional
checks:

```c
// Check that emacs-module.h has static support for Emacs 29.
#if defined EMACS_MAJOR_VERSION && EMACS_MAJOR_VERSION >= 29
// Make sure the cast below doesn’t lose information.
static_assert(SIZE_MAX >= PTRDIFF_MAX, "unsupported architecture");
// Check that the Emacs that has loaded this module supports the function.
if ((size_t)env->size > offsetof(emacs_env, function)) {
  return env->function(env, …);
}
#endif
// Some workaround in case Emacs 29 isn’t available.
```

CGo [doesn’t support calling C function
pointers](https://pkg.go.dev/cmd/cgo#hdr-Go_references_to_C).  Therefore, the
code wraps all function pointers in the `emacs_runtime` and `emacs_env`
structures in ordinary wrapper functions.  The wrapper functions use only the
`int64_t` and `uint64_t` types in their interfaces.

The package uses [a global
registry](https://github.com/golang/go/wiki/cgo#function-variables) to identify
exported functions.

You can compile the examples in the test files to form an example module.  In
addition to the examples, this also needs the file `example_test.go` to
initialize the examples.  The Emacs Lisp file `test.el` runs the examples as
[ERT](https://www.gnu.org/software/emacs/manual/html_node/ert/index.html)
tests.

To build and run all tests, install [Bazel][] and run

```shell
bazel test //...
```

[Bazel]: https://bazel.build/
