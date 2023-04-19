// Copyright 2019, 2021, 2023 Google LLC
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
// #include <stddef.h>
// #include <stdlib.h>
// #include "emacs-module.h"
// #include "wrappers.h"
// struct phst_emacs_value_result phst_emacs_make_string(emacs_env *env,
//                                                       _GoString_ contents) {
//   size_t size = _GoStringLen(contents);
//   assert(size > 0);
//   return phst_emacs_make_string_impl(env, _GoStringPtr(contents), size - 1);
// }
import "C"

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"unicode/utf8"
	"unsafe"
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
	if strings.ContainsRune(string(s), '\r') {
		return e.Let("inhibit-eol-conversion", T, func() (Value, error) {
			return e.makeString(string(s))
		})
	}
	return e.makeString(string(s))
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
// named String to avoid confusion with the [fmt.Stringer.String] method.
func (e Env) Str(v Value) (string, error) {
	r := C.phst_emacs_copy_string_contents(e.raw(), v.r)
	if err := e.check(r.base); err != nil {
		return "", err
	}
	if r.size == 0 {
		return "", nil
	}
	defer C.free(unsafe.Pointer(r.data))
	s := C.GoStringN(r.data, r.size)
	if !utf8.ValidString(s) {
		return "", WrongTypeArgument("valid-string-p", v)
	}
	return s, nil
}

func (e Env) makeString(s string) (Value, error) {
	return e.checkValue(C.phst_emacs_make_string(e.raw(), s+"\x00"))
}

// FormatMessage calls the Emacs function format-message with the given format
// string and arguments.  If the call to format-message fails, FormatMessage
// returns a descriptive error string.  Note that the syntax of the format
// string for FormatMessage is similar but not identical to the format strings
// for the [fmt.Printf] family.
func (e Env) FormatMessage(format string, args ...In) string {
	var s String
	args = append([]In{String(format)}, args...)
	if err := e.CallOut("format-message", &s, args...); err != nil {
		// Don’t return the error to the caller to avoid clutter.
		return fmt.Sprintf("<error formatting message: %s>", e.Message(err))
	}
	return string(s)
}

// Bytes is a type with underlying type []byte that knows how to convert itself
// to an Emacs unibyte string.
type Bytes []byte

// Emacs creates an Emacs unibyte string value representing the given bytes.
// It always makes a copy of the byte slice.
func (b Bytes) Emacs(e Env) (Value, error) {
	if len(b) == 0 {
		return e.checkValue(C.phst_emacs_make_unibyte_string(e.raw(), nil, 0))
	}
	return e.checkValue(C.phst_emacs_make_unibyte_string(e.raw(), unsafe.Pointer(&b[0]), C.int64_t(len(b))))
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

func byteArrayIn(v reflect.Value) In {
	r := make(Bytes, v.Len())
	reflect.Copy(reflect.ValueOf(r), v)
	return r
}

func byteArrayOut(v reflect.Value) Out { return getUnibyteFromArray(v) }

type getUnibyteFromArray reflect.Value

func (g getUnibyteFromArray) FromEmacs(e Env, v Value) error {
	b, err := e.Bytes(v)
	if err != nil {
		return err
	}
	u := reflect.Value(g).Elem()
	if len(b) != u.Len() {
		return fmt.Errorf("incompatible array length: Go array has length %d, but Emacs string has length %d", u.Len(), len(b))
	}
	reflect.Copy(u, reflect.ValueOf(b))
	return nil
}
