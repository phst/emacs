# Go bindings for the Emacs module API

This package implements [Go](https://golang.org/) bindings for the [GNU Emacs
module API](https://phst.eu/emacs-modules).

This is not an officially supported Google product.

## Usage

See the package documentation.

## Implementation notes

The package assumes that `ptrdiff_t` and `intmax_t` have the same domain as
`int64_t`.  It also assumes that `uintptr_t` has the same domain as `uint64_t`.
The C code uses `static_assert` to check these assumptions.  The C standard
[guarantess](https://en.cppreference.com/w/c/types/integer) that `int64_t` and
`uint64_t` have exactly 64 bits without padding bits and that they use two’s
complement representation.  The corresponding Go types `int64` and `uint64`
have the same [representation](https://golang.org/ref/spec#Numeric_types).
This means that converting `int64_t` and `uint64_t` to and from the Go types
can never lose data.

CGo [doesn’t support calling C function
pointers](https://golang.org/cmd/cgo/#hdr-Go_references_to_C).  Therefore, the
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
