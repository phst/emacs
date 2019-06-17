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

// #cgo CPPFLAGS: -DEMACS_MODULE_GMP
// #cgo LDFLAGS: -lgmp
// #include <assert.h>
// #include <limits.h>
// #include <stddef.h>
// #include <stdint.h>
// #include <gmp.h>
// #include <emacs-module.h>
// static_assert(PTRDIFF_MAX <= SIZE_MAX, "unsupported architecture");
// static_assert(INTMAX_MIN == INT64_MIN, "unsupported architecture");
// static_assert(INTMAX_MAX == INT64_MAX, "unsupported architecture");
// static_assert(LONG_MIN >= INT64_MIN, "unsupported architecture");
// static_assert(LONG_MAX <= INT64_MAX, "unsupported architecture");
// static_assert(ULONG_MAX <= UINT64_MAX, "unsupported architecture");
// int64_t extract_integer(emacs_env *env, emacs_value value) {
//   return env->extract_integer(env, value);
// }
// void extract_big_integer(emacs_env *env, emacs_value value, mpz_t result) {
// #if defined EMACS_MAJOR_VERSION && EMACS_MAJOR_VERSION >= 27
//   if ((size_t)env->size > offsetof(emacs_env, extract_big_integer)) {
//     struct emacs_mpz temp = {*result};
//     env->extract_big_integer(env, value, &temp);
//     *result = *temp.value;
//     return;
//   }
// #endif
//   int64_t i = env->extract_integer(env, value);
//   if (i >= 0 && (uint64_t)i <= ULONG_MAX) {
//     mpz_set_ui(result, i);
//     return;
//   }
//   if (i >= LONG_MIN && i <= LONG_MAX) {
//     mpz_set_si(result, i);
//     return;
//   }
//   uint64_t u;
//   // Set u = abs(i).  See https://stackoverflow.com/a/17313717.
//   if (i >= 0) {
//     u = i;
//   } else {
//     u = -(uint64_t)i;
//   }
//   enum {
//     count = 1,
//     order = 1,
//     size = sizeof u,
//     endian = 0,
//     nails = 0
//   };
//   mpz_import(result, count, order, size, endian, nails, &u);
//   if (i < 0) mpz_neg(result, result);
// }
// emacs_value make_integer(emacs_env *env, int64_t value) {
//   return env->make_integer(env, value);
// }
// emacs_value make_big_integer(emacs_env *env, const mpz_t value) {
// #if defined EMACS_MAJOR_VERSION && EMACS_MAJOR_VERSION >= 27
//   if ((size_t)env->size > offsetof(emacs_env, make_big_integer)) {
//     struct emacs_mpz temp = {*value};
//     return env->make_big_integer(env, &temp);
//   }
// #endif
//   // The code below always calls make_integer if possible,
//   // so this can only overflow.
//   env->non_local_exit_signal(env, env->intern(env, "overflow-error"), env->intern(env, "nil"));
//   return NULL;
// }
// // This wrapper function is needed because mpz_sgn is a macro.
// int emacs_mpz_sgn(const mpz_t value) {
//   return mpz_sgn(value);
// }
import "C"

import (
	"fmt"
	"math"
	"math/big"
	"reflect"
	"unsafe"
)

// Int is a type with underlying type int64 that knows how to convert itself
// into an Emacs value.
type Int int64

// Emacs creates an Emacs value representing the given integer.  It returns an
// error if the integer value is too big for Emacs.
func (i Int) Emacs(e Env) (Value, error) {
	return e.checkRaw(C.make_integer(e.raw(), C.int64_t(i)))
}

// FromEmacs sets *i to the integer stored in v.  It returns an error if v is
// not an integer, or if doesn’t fit into an int64.
func (i *Int) FromEmacs(e Env, v Value) error {
	r, err := e.Int(v)
	if err != nil {
		return err
	}
	*i = Int(r)
	return nil
}

// Int returns the integer stored in v.  It returns an error if v is not an
// integer, or if it doesn’t fit into an int64.
func (e Env) Int(v Value) (int64, error) {
	i := C.extract_integer(e.raw(), v.r)
	return int64(i), e.check()
}

// BigInt sets z to the integer stored in v.  It returns an error if v is not
// an integer.
func (e Env) BigInt(v Value, z *big.Int) error {
	var i C.mpz_t
	C.mpz_init(&i[0])
	defer C.mpz_clear(&i[0])
	C.extract_big_integer(e.raw(), v.r, &i[0])
	if err := e.check(); err != nil {
		return err
	}
	if C.mpz_fits_ulong_p(&i[0]) != 0 {
		z.SetUint64(uint64(C.mpz_get_ui(&i[0])))
		return nil
	}
	if C.mpz_fits_slong_p(&i[0]) != 0 {
		z.SetInt64(int64(C.mpz_get_si(&i[0])))
		return nil
	}
	// See
	// https://gmplib.org/manual/Integer-Import-and-Export.html#index-Export.
	const numb = 8*gmpSize - gmpNails
	count := (C.mpz_sizeinbase(&i[0], 2) + numb - 1) / numb
	p := make([]byte, count)
	var written C.size_t
	C.mpz_export(unsafe.Pointer(&p[0]), &written, gmpOrder, gmpSize, gmpEndian, gmpNails, &i[0])
	if written != count {
		return fmt.Errorf("unexpected number of bytes exported: got %d, want %d", written, count)
	}
	z.SetBytes(p)
	if C.emacs_mpz_sgn(&i[0]) == -1 {
		z.Neg(z)
	}
	return nil
}

// Uint is a type with underlying type uint64 that knows how to convert itself
// into an Emacs value.
type Uint uint64

// Emacs creates an Emacs value representing the given integer.  It returns an
// error if the integer value is too big for Emacs.
func (i Uint) Emacs(e Env) (Value, error) {
	if i > math.MaxInt64 {
		z := new(big.Int)
		z.SetUint64(uint64(i))
		return (*BigInt)(z).Emacs(e)
	}
	return Int(i).Emacs(e)
}

// FromEmacs sets *i to the integer stored in v.  It returns an error if v is
// not an integer, or if it doesn’t fit into an uint64.
func (i *Uint) FromEmacs(e Env, v Value) error {
	r, err := e.Uint(v)
	if err != nil {
		return err
	}
	*i = Uint(r)
	return nil
}

// Uint returns the integer stored in v.  It returns an error if v is not an
// integer, or if it doesn’t fit into an uint64.
func (e Env) Uint(v Value) (uint64, error) {
	var z big.Int
	if err := e.BigInt(v, &z); err != nil {
		return 0, err
	}
	if !z.IsUint64() {
		return 0, WrongTypeArgument("natnump", String(z.String()))
	}
	return z.Uint64(), nil
}

// BigInt is a type with underlying type big.Int that knows how to convert
// itself to and from an Emacs value.
type BigInt big.Int

// String formats the big integer as a string.  It calls big.Int.String.
func (i *BigInt) String() string { return (*big.Int)(i).String() }

// Emacs creates an Emacs value representing the given integer.  It returns an
// error if the integer value is too big for Emacs.
func (i *BigInt) Emacs(e Env) (Value, error) {
	b := (*big.Int)(i)
	if b.IsInt64() {
		return Int(b.Int64()).Emacs(e)
	}
	var v C.mpz_t
	C.mpz_init(&v[0])
	defer C.mpz_clear(&v[0])
	p := b.Bytes()
	C.mpz_import(&v[0], C.size_t(len(p)), gmpOrder, gmpSize, gmpEndian, gmpNails, unsafe.Pointer(&p[0]))
	if b.Sign() == -1 {
		C.mpz_neg(&v[0], &v[0])
	}
	return e.checkRaw(C.make_big_integer(e.raw(), &v[0]))
}

// FromEmacs sets *i to the integer stored in v.  It returns an error if v is
// not an integer.
func (i *BigInt) FromEmacs(e Env, v Value) error {
	return e.BigInt(v, (*big.Int)(i))
}

// Constants for mpz_import and mpz_export.  See
// https://gmplib.org/manual/Integer-Import-and-Export.html.
const (
	gmpSize   = 1
	gmpOrder  = 1
	gmpEndian = 1
	gmpNails  = 0
)

func intIn(v reflect.Value) In   { return Int(reflect.Value(v).Int()) }
func intOut(v reflect.Value) Out { return reflectInt(v) }

type reflectInt reflect.Value

func (r reflectInt) FromEmacs(e Env, v Value) error {
	i, err := e.Int(v)
	if err != nil {
		return err
	}
	s := reflect.Value(r)
	if s.OverflowInt(i) {
		return OverflowError(fmt.Sprint(i))
	}
	s.SetInt(i)
	return nil
}

func uintIn(v reflect.Value) In   { return Uint(reflect.Value(v).Uint()) }
func uintOut(v reflect.Value) Out { return reflectUint(v) }

type reflectUint reflect.Value

func (r reflectUint) FromEmacs(e Env, v Value) error {
	i, err := e.Uint(v)
	if err != nil {
		return err
	}
	s := reflect.Value(r)
	if s.OverflowUint(i) {
		return OverflowError(fmt.Sprint(i))
	}
	s.SetUint(i)
	return nil
}

func bigIntIn(v reflect.Value) In   { return (*BigInt)(v.Interface().(*big.Int)) }
func bigIntOut(v reflect.Value) Out { return (*BigInt)(v.Interface().(*big.Int)) }
