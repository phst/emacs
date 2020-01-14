// Copyright 2020 Google LLC
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

import (
	"bytes"
	"log"
	"math/rand"
	"reflect"
	"strings"
	"testing/quick"
)

func init() {
	ERTTest(stringRoundtrip)
	ERTTest(bytesRoundtrip)
	ERTTest(asciiRoundtrip)
}

func stringRoundtrip(e Env) error {
	f := func(a string) bool {
		v, err := String(a).Emacs(e)
		if err != nil {
			log.Printf("couldn’t convert string %q to Emacs: %s", a, e.Message(err))
			return false
		}
		b, err := e.Str(v)
		if err != nil {
			log.Printf("couldn’t convert string from Emacs: %s", e.Message(err))
			return false
		}
		equal := a == b
		if !equal {
			log.Printf("string roundtrip: got %q, want %q", b, a)
		}
		return equal
	}
	return quick.Check(f, nil)
}

func bytesRoundtrip(e Env) error {
	f := func(a []byte) bool {
		v, err := Bytes(a).Emacs(e)
		if err != nil {
			log.Printf("couldn’t convert bytes %q to Emacs: %s", a, e.Message(err))
			return false
		}
		b, err := e.Bytes(v)
		if err != nil {
			log.Printf("couldn’t convert bytes from Emacs: %s", e.Message(err))
			return false
		}
		equal := bytes.Equal(a, b)
		if !equal {
			log.Printf("bytes roundtrip: got %q, want %q", b, a)
		}
		return equal
	}
	return quick.Check(f, nil)
}

func asciiRoundtrip(e Env) error {
	f := func(a asciiString) bool {
		v, err := String(a).Emacs(e)
		if err != nil {
			log.Printf("couldn’t convert string %q to Emacs: %s", a, e.Message(err))
			return false
		}
		b, err := e.Str(v)
		if err != nil {
			log.Printf("couldn’t convert string from Emacs: %s", e.Message(err))
			return false
		}
		equal := string(a) == b
		if !equal {
			log.Printf("string roundtrip: got %q, want %q", b, a)
		}
		return equal
	}
	return quick.Check(f, nil)
}

type asciiString string

func (asciiString) Generate(rand *rand.Rand, size int) reflect.Value {
	n := rand.Intn(size)
	var b strings.Builder
	for i := 0; i < n; i++ {
		b.WriteByte(byte(rand.Intn(0x80)))
	}
	return reflect.ValueOf(asciiString(b.String()))
}
