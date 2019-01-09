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

// #include <assert.h>
// #include <limits.h>
// #include <stdint.h>
// #include <emacs-module.h>
// static_assert(CHAR_BIT == 8, "unsupported architecture");
// static_assert(PTRDIFF_MIN == INT64_MIN, "unsupported architecture");
// static_assert(PTRDIFF_MAX == INT64_MAX, "unsupported architecture");
// bool copy_string_contents(emacs_env *env, emacs_value value, uint8_t *buffer, int64_t *size) {
//   // It’s fine to cast uint8_t * to char *.  See
//   // https://en.cppreference.com/w/c/language/object#Strict_aliasing.
//   ptrdiff_t size_ptrdiff = *size;
//   bool success = env->copy_string_contents(env, value, (char *) buffer, &size_ptrdiff);
//   *size = size_ptrdiff;
//   return success;
// }
// emacs_value make_string(emacs_env *env, _GoString_ contents) {
//   return env->make_string(env, _GoStringPtr(contents), _GoStringLen(contents) - 1);
// }
import "C"

import (
	"fmt"
	"reflect"
	"strconv"
	"unicode/utf8"
)

// String is a type with underlying type string that knows how to convert
// itself to an Emacs string.
type String string

// String quotes the string to mimic the Emacs printed representation.
func (s String) String() string {
	return strconv.Quote(string(s))
}

// Emacs creates an Emacs value representing the given string.  It returns an
// error if the string isn’t a valid UTF-8 string.
func (s String) Emacs(e Env) (Value, error) {
	if !utf8.ValidString(string(s)) {
		return Value{}, WrongTypeArgument("valid-string-p", String(fmt.Sprintf("%+q", s)))
	}
	return e.checkRaw(C.make_string(e.raw(), string(s)+"\x00"))
}

// FromEmacs sets *s to the string stored in v.  It returns an error if v is
// not a string, or if it’s not a valid Unicode scalar value sequence.
func (s *String) FromEmacs(e Env, v Value) error {
	r, err := e.Str(v)
	if err != nil {
		return err
	}
	*s = String(r)
	return nil
}

// Str returns the string stored in v.  It returns an error if v is not a
// string, or if it’s not a valid Unicode scalar value sequence.  Str is not
// named String to avoid confusion with the String method of the Stringer
// interface.
func (e Env) Str(v Value) (string, error) {
	// See https://phst.github.io/emacs-modules#copy_string_contents.
	var size C.int64_t
	if ok := C.copy_string_contents(e.raw(), v.r, nil, &size); !ok {
		return "", e.check()
	}
	if size == 0 {
		return "", nil
	}
	buffer := make([]byte, size)
	if ok := C.copy_string_contents(e.raw(), v.r, (*C.uint8_t)(&buffer[0]), &size); !ok {
		return "", e.check()
	}
	r := string(buffer[:size-1])
	if !utf8.ValidString(r) {
		return "", WrongTypeArgument("valid-string-p", v)
	}
	return r, nil
}

// Bytes is a type with underlying type []byte that knows how to convert itself
// to an Emacs unibyte string.
type Bytes []byte

// Emacs creates an Emacs unibyte string value representing the given bytes.
// It always makes a copy of the byte slice.
func (b Bytes) Emacs(e Env) (Value, error) {
	args := make([]In, len(b))
	for i, b := range b {
		args[i] = Int(b)
	}
	return e.Call("unibyte-string", args...)
}

// FromEmacs sets *b to the unibyte string stored in v.  It returns an error if
// v is not a unibyte string.
func (b *Bytes) FromEmacs(e Env, v Value) error {
	r, err := e.Bytes(v)
	if err != nil {
		return err
	}
	*b = r
	return nil
}

// Bytes returns the unibyte string stored in v.  It returns an error if v is
// not a unibyte string.
func (e Env) Bytes(v Value) ([]byte, error) {
	var isString Bool
	if err := e.CallOut("stringp", &isString, v); err != nil {
		return nil, err
	}
	if !isString {
		return nil, WrongTypeArgument("stringp", v)
	}
	var isMultibyte Bool
	if err := e.CallOut("multibyte-string-p", &isMultibyte, v); err != nil {
		return nil, err
	}
	if isMultibyte {
		return nil, WrongTypeArgument("unibyte-string-p", v)
	}
	vec := VectorOut{New: func() Out { return new(Int) }}
	if err := e.CallOut("vconcat", &vec, v); err != nil {
		return nil, err
	}
	r := make([]byte, len(vec.Data))
	for i, o := range vec.Data {
		b, err := int64ToByte(int64(*o.(*Int)))
		if err != nil {
			return nil, err
		}
		r[i] = b
	}
	return r, nil
}

func stringIn(v reflect.Value) In   { return String(reflect.Value(v).String()) }
func stringOut(v reflect.Value) Out { return reflectString(v) }

type reflectString reflect.Value

func (r reflectString) FromEmacs(e Env, v Value) error {
	s, err := e.Str(v)
	if err != nil {
		return err
	}
	reflect.Value(r).SetString(s)
	return nil
}

func bytesIn(v reflect.Value) In   { return Bytes(reflect.Value(v).Bytes()) }
func bytesOut(v reflect.Value) Out { return reflectBytes(v) }

type reflectBytes reflect.Value

func (r reflectBytes) FromEmacs(e Env, v Value) error {
	b, err := e.Bytes(v)
	if err != nil {
		return err
	}
	reflect.Value(r).SetBytes(b)
	return nil
}
