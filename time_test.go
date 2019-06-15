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
	"fmt"
	"log"
	"math/big"
	"math/rand"
	"reflect"
	"testing"
	"testing/quick"
	"time"
)

func TestTime(t *testing.T) {
	f := func(s, ns int64) bool {
		u := time.Unix(s, ns)
		var ps Picoseconds
		ps.FromTime(u)
		v, err := ps.Time()
		if err != nil {
			t.Logf("u = %s, ps = %s, err = %s", u, &ps, err)
			return false
		}
		if !u.Equal(v) {
			t.Logf("u = %s, v = %s, ps = %s", u, v, &ps)
			return false
		}
		return true
	}
	if err := quick.Check(f, nil); err != nil {
		t.Error(err)
	}
}

func TestDuration(t *testing.T) {
	f := func(u time.Duration) bool {
		var ps Picoseconds
		ps.FromDuration(u)
		v, err := ps.Duration()
		if err != nil {
			t.Logf("u = %s, ps = %s, err = %s", u, &ps, err)
			return false
		}
		if u != v {
			t.Logf("u = %s, v = %s, ps = %s", u, v, &ps)
			return false
		}
		return true
	}
	if err := quick.Check(f, nil); err != nil {
		t.Error(err)
	}
}

func TestQuad(t *testing.T) {
	f := func(high int64, low, μsec, psec uint16) bool {
		u := [4]int64{high, int64(low), int64(μsec), int64(psec)}
		var ps Picoseconds
		ps.fromQuad(u[0], u[1], u[2], u[3])
		var v [4]big.Int
		ps.quad(&v[0], &v[1], &v[2], &v[3])
		for i := range u {
			if !v[i].IsInt64() || u[i] != v[i].Int64() {
				t.Logf("u[%d] = %d, v[%d] = %s, ps = %s", i, u[i], i, &v[i], &ps)
				return false
			}
		}
		return true
	}
	if err := quick.Check(f, nil); err != nil {
		t.Error(err)
	}
}

func init() {
	ERTTest(goTimeRoundtrip)
	ERTTest(goDurationRoundtrip)
	ERTTest(emacsTimeRoundtrip)
	ERTTest(emacsDurationRoundtrip)
}

func goTimeRoundtrip(e Env) error {
	f := func(a Time) bool {
		v, err := a.Emacs(e)
		if err != nil {
			log.Printf("couldn’t convert time %s to Emacs: %v", a, e.Message(err))
			return false
		}
		b, err := e.Time(v)
		if err != nil {
			log.Printf("couldn’t convert time from Emacs: %v", e.Message(err))
			return false
		}
		return time.Time(a).Equal(b)
	}
	return quick.Check(f, nil)
}

func (Time) Generate(rand *rand.Rand, size int) reflect.Value {
	return reflect.ValueOf(Time(time.Unix(rand.Int63(), rand.Int63())))
}

func goDurationRoundtrip(e Env) error {
	f := func(a Duration) bool {
		v, err := a.Emacs(e)
		if err != nil {
			log.Printf("couldn’t convert duration %s to Emacs: %v", a, e.Message(err))
			return false
		}
		b, err := e.Duration(v)
		if err != nil {
			log.Printf("couldn’t convert duration from Emacs: %v", e.Message(err))
			return false
		}
		return time.Duration(a) == b
	}
	return quick.Check(f, nil)
}

func emacsTimeRoundtrip(e Env) error {
	f := func(q quad) bool {
		a, trunc, ok := q.lists(e)
		if !ok {
			return false
		}
		t, err := e.Time(a)
		if e.IsOverflowError(err) {
			// Treat overflows as success.
			return true
		}
		if err != nil {
			log.Printf("couldn’t convert time from Emacs: %v", e.Message(err))
			return false
		}
		b, err := Time(t).Emacs(e)
		if err != nil {
			log.Printf("couldn’t convert time %s to Emacs: %v", t, e.Message(err))
			return false
		}
		// We need to compare the truncated value.
		return equalTime(e, trunc, b)
	}
	return quick.Check(f, &quick.Config{MaxCountScale: 10})
}

func emacsDurationRoundtrip(e Env) error {
	f := func(q quad) bool {
		a, trunc, ok := q.lists(e)
		if !ok {
			return false
		}
		d, err := e.Duration(a)
		if e.IsOverflowError(err) {
			// Treat overflows as success.
			return true
		}
		if err != nil {
			log.Printf("couldn’t convert duration from Emacs: %v", e.Message(err))
			return false
		}
		b, err := Duration(d).Emacs(e)
		if err != nil {
			log.Printf("couldn’t convert duration %s to Emacs: %v", d, e.Message(err))
			return false
		}
		// We need to compare the truncated value.
		return equalTime(e, trunc, b)
	}
	return quick.Check(f, &quick.Config{MaxCountScale: 10})
}

type quad struct {
	hi         int64
	lo, μs, ps uint16
}

func (quad) Generate(rand *rand.Rand, size int) reflect.Value {
	return reflect.ValueOf(quad{
		// We intentionally leave the upper 16 bit out.  On typical
		// systems, time_t is a 64-bit signed integer, so filling all
		// the bits would cause almost 100% overflow errors and no real
		// tests.  We cast to signed first to have some negative
		// numbers.
		int64(rand.Uint64()) >> 16,
		// The other parts have technically a smaller domain than
		// uint16, but the functions should accept denormalized values
		// as well.
		uint16(rand.Uint32()), uint16(rand.Uint32()), uint16(rand.Uint32()),
	})
}

func (q quad) String() string {
	return fmt.Sprintf("(%d %d %d %d)", q.hi, q.lo, q.μs, q.ps)
}

func (q quad) lists(e Env) (precise, truncated Value, ok bool) {
	precise, ok = q.list(e)
	if !ok {
		return
	}
	q.ps -= q.ps % 1000
	truncated, ok = q.list(e)
	return
}

func (q quad) list(e Env) (v Value, ok bool) {
	v, err := e.List(Int(q.hi), Int(q.lo), Int(q.μs), Int(q.ps))
	if err != nil {
		log.Printf("couldn’t convert time quadruple %s to list: %v", q, e.Message(err))
	}
	return v, err == nil
}

// equalTime returns whether the two Emacs time values in a and b are equal.
func equalTime(e Env, a, b Value) bool {
	// Rewrite: a = b ⟺ a − b = 0 ⟺ ¬(a − b < 0 ∨ 0 < a − b)
	d, err := e.Call("time-subtract", a, b)
	if err != nil {
		log.Printf("couldn’t subtract time values: %s", e.Message(err))
		return false
	}
	// Emacs doesn’t have a time-equal-p function or similar.
	pos, err := timeLess(e, Int(0), d)
	if err != nil {
		// Don’t bother returning the error since the test functions
		// above treat different times as an error.
		log.Printf("couldn’t compare times: %s", e.Message(err))
		return false
	}
	neg, err := timeLess(e, d, Int(0))
	if err != nil {
		// Don’t bother returning the error since the test functions
		// above treat different times as an error.
		log.Printf("couldn’t compare times: %s", e.Message(err))
		return false
	}
	equal := !(pos || neg)
	if !equal {
		log.Print(e.FormatMessage("time values %s and %s aren’t equal, difference is %s", a, b, d))
		// Don’t bother with returning the error since the test
		// functions above treat different times as an error.
	}
	return equal
}

// timeLess returns whether a < b for two Emacs time values.
func timeLess(e Env, a, b In) (bool, error) {
	var less Bool
	err := e.CallOut("time-less-p", &less, a, b)
	return bool(less), err
}
