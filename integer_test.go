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

import (
	"log"
	"math/big"
	"math/rand"
	"reflect"
	"testing/quick"
)

func init() {
	ERTTest(intRoundtrip)
	ERTTest(bigIntRoundtrip)
}

func intRoundtrip(e Env) error {
	canOverflow := MajorVersion() < 27
	f := func(a int64) bool {
		v, err := Int(a).Emacs(e)
		if canOverflow && e.IsOverflowError(err) {
			return true
		}
		if err != nil {
			log.Printf("couldn’t convert integer %d to Emacs: %s", a, e.Message(err))
			return true
		}
		b, err := e.Int(v)
		if err != nil {
			log.Printf("couldn’t convert integer from Emacs: %s", e.Message(err))
			return false
		}
		if a != b {
			log.Printf("integer roundtrip: got %d, want %d", b, a)
		}
		return a == b
	}
	return quick.Check(f, nil)
}

func bigIntRoundtrip(e Env) error {
	canOverflow := MajorVersion() < 27
	f := func(a *BigInt) bool {
		v, err := a.Emacs(e)
		if canOverflow && e.IsOverflowError(err) {
			return true
		}
		if err != nil {
			log.Printf("couldn’t convert big integer %s to Emacs: %s", a, e.Message(err))
			return false
		}
		b := new(big.Int)
		if err := e.BigInt(v, b); err != nil {
			log.Printf("couldn’t convert big integer from Emacs: %s", e.Message(err))
			return false
		}
		equal := (*big.Int)(a).Cmp(b) == 0
		if !equal {
			log.Printf("big integer roundtrip: got %s, want %s", b, a)
		}
		return equal
	}
	return quick.Check(f, nil)
}

func (*BigInt) Generate(rand *rand.Rand, size int) reflect.Value {
	// Generate values that are rarely zero, occasionally negative, and
	// frequently positive.
	sign := rand.Intn(10)
	if sign == 0 {
		return reflect.ValueOf(new(BigInt))
	}
	z := big.NewInt(1)
	// Generate values with exponentially-distributed magnitude that
	// occasionally fit into an int64.
	z.Lsh(z, uint(rand.Intn(100)))
	z.Rand(rand, z)
	if sign < 4 {
		z.Neg(z)
	}
	return reflect.ValueOf((*BigInt)(z))
}
