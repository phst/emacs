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
	"fmt"
	"log"
	"math/big"
	"math/rand"
	"reflect"
	"testing"
	"testing/quick"
	"time"
)

func init() {
	ERTTest(reflectRoundtrip)
	ERTTest(reflectValue)
	ERTTest(reflectInterface)
}

func reflectRoundtrip(e Env) error {
	f := func(a Reflect) bool {
		v, err := a.Emacs(e)
		if err != nil {
			log.Printf("couldn’t convert reflected value %#v to Emacs: %s", reflect.Value(a), e.Message(err))
			return false
		}
		want := reflect.Value(a)
		p := Reflect(reflect.New(want.Type()))
		if err := p.FromEmacs(e, v); err != nil {
			log.Printf("couldn’t convert reflected value from Emacs: %s", e.Message(err))
			return false
		}
		got := reflect.Value(p).Elem()
		equal := reflect.DeepEqual(got.Interface(), want.Interface())
		if !equal {
			log.Printf("reflected value roundtrip: got %#v, want %#v", got, want)
		}
		return equal
	}
	return quick.Check(f, nil)
}

func reflectValue(e Env) error {
	a, err := Int(123).Emacs(e)
	if err != nil {
		return fmt.Errorf("couldn’t convert integer to Emacs: %s", e.Message(err))
	}
	r := Reflect(reflect.ValueOf(a))
	b, err := r.Emacs(e)
	if err != nil {
		return fmt.Errorf("couldn’t convert reflected value %#v to Emacs: %s", reflect.Value(r), e.Message(err))
	}
	if !e.Eq(a, b) {
		return fmt.Errorf("got %v, want %v", b, a)
	}

	s := Reflect(reflect.ValueOf(new(Value)))
	if err := s.FromEmacs(e, b); err != nil {
		return fmt.Errorf("couldn’t convert reflected value from Emacs: %s", e.Message(err))
	}
	c := *reflect.Value(s).Interface().(*Value)
	if !e.Eq(b, c) {
		return fmt.Errorf("got %v, want %v", c, b)
	}
	return nil
}

func reflectInterface(e Env) error {
	var temp interface{} = 123
	in := Reflect(reflect.ValueOf(&temp).Elem())
	v, err := in.Emacs(e)
	if err != nil {
		return fmt.Errorf("couldn’t convert reflected value %#v to Emacs: %s", reflect.Value(in), e.Message(err))
	}
	got, err := e.Int(v)
	if err != nil {
		return fmt.Errorf("couldn’t convert integer from Emacs: %s", e.Message(err))
	}
	const want = 123
	if got != want {
		return fmt.Errorf("got %d, want %d", got, want)
	}
	return nil
}

func (Reflect) Generate(rand *rand.Rand, size int) reflect.Value {
	t := randomType(rand, size)
	v, ok := quick.Value(t, rand)
	if !ok {
		panic(fmt.Errorf("can’t generate value of type %s", t))
	}
	return reflect.ValueOf(Reflect(v))
}

func randomType(rand *rand.Rand, size int) reflect.Type {
	if size < 0 {
		panic("negative size")
	}
	// The commented-out kinds are type kinds that this library doesn’t
	// support (yet).  In addition, we don’t generate the 64-bit integral
	// types to avoid overflow errors if either Emacs or emacs-module.h
	// doesn’t support big integers.
	kinds := []reflect.Kind{
		// reflect.Invalid,
		reflect.Bool,
		// reflect.Int,
		reflect.Int8,
		reflect.Int16,
		reflect.Int32,
		// reflect.Int64,
		// reflect.Uint,
		reflect.Uint8,
		reflect.Uint16,
		reflect.Uint32,
		// reflect.Uint64,
		// reflect.Uintptr,
		reflect.Float32,
		reflect.Float64,
		// reflect.Complex64,
		// reflect.Complex128,
		reflect.String,
		// reflect.UnsafePointer,
	}
	if size > 0 {
		// If we have complexity left, allow non-scalar types.
		kinds = append(kinds,
			reflect.Array,
			// reflect.Chan,
			// reflect.Func,
			// reflect.Interface,
			reflect.Map,
			// reflect.Ptr,
			reflect.Slice,
			// reflect.Struct,
		)
	}
	// How fast to shrink non-scalar types.  Must be greater than 1.  If
	// this is too small, the test is likely to time out.
	const factor = 10
	size /= factor
	kind := kinds[rand.Intn(len(kinds))]
	switch kind {
	case reflect.Bool:
		return reflect.TypeOf(bool(false))
	case reflect.Int:
		return reflect.TypeOf(int(0))
	case reflect.Int8:
		return reflect.TypeOf(int8(0))
	case reflect.Int16:
		return reflect.TypeOf(int16(0))
	case reflect.Int32:
		return reflect.TypeOf(int32(0))
	case reflect.Int64:
		return reflect.TypeOf(int64(0))
	case reflect.Uint:
		return reflect.TypeOf(uint(0))
	case reflect.Uint8:
		return reflect.TypeOf(uint8(0))
	case reflect.Uint16:
		return reflect.TypeOf(uint16(0))
	case reflect.Uint32:
		return reflect.TypeOf(uint32(0))
	case reflect.Uint64:
		return reflect.TypeOf(uint64(0))
	case reflect.Uintptr:
		return reflect.TypeOf(uintptr(0))
	case reflect.Float32:
		return reflect.TypeOf(float32(0))
	case reflect.Float64:
		return reflect.TypeOf(float64(0))
	case reflect.Array:
		return reflect.ArrayOf(rand.Intn(50), randomType(rand, size))
	case reflect.Map:
		key := randomType(rand, size)
		for !key.Comparable() {
			key = randomType(rand, size)
		}
		elem := randomType(rand, size)
		return reflect.MapOf(key, elem)
	case reflect.Slice:
		return reflect.SliceOf(randomType(rand, size))
	case reflect.String:
		return reflect.TypeOf("")
	default:
		panic("this can’t happen")
	}
}

func TestInOutFunc(t *testing.T) {
	type want uint
	const (
		inErr want = 1 << iota
		outErr
	)
	var temp interface{} = 1
	for _, tc := range []struct {
		val interface{}
		want
	}{
		{Value{}, 0},
		{false, 0},
		{true, 0},
		{int(1), 0},
		{float32(1), 0},
		{"hi", 0},
		{time.Unix(1, 0), 0},
		{time.Second, 0},
		{[0]int{}, 0},
		{[1]int{1}, 0},
		{[]int{}, 0},
		{[]int{1}, 0},
		{map[string]float64{}, 0},
		{map[string]float64{"hi": 1}, 0},
		{big.Int{}, inErr},
		{big.NewInt(1), outErr},
		{temp, 0},
		{func() {}, inErr | outErr},
		{struct{ F int }{1}, inErr | outErr},
	} {
		t.Run(fmt.Sprintf("%#v", tc.val), func(t *testing.T) {
			inVal := reflect.ValueOf(tc.val)
			inType := inVal.Type()
			t.Run(fmt.Sprintf("InFuncFor(%s)", inType), func(t *testing.T) {
				inFunc, err := InFuncFor(inType)
				gotErr := err != nil
				wantErr := tc.want&inErr != 0
				switch {
				case gotErr && !wantErr:
					t.Fatalf("got error %s", err)
				case !gotErr && wantErr:
					t.Fatalf("got %#v, want error", inFunc)
				case gotErr && wantErr:
					return
				}
				if inFunc(inVal) == nil {
					t.Error("got nil")
				}
			})
			outVal := reflect.New(inType)
			outType := outVal.Type()
			t.Run(fmt.Sprintf("OutFuncFor(%s)", outType), func(t *testing.T) {
				outFunc, err := OutFuncFor(outType)
				gotErr := err != nil
				wantErr := tc.want&outErr != 0
				switch {
				case gotErr && !wantErr:
					t.Fatalf("got error %s", err)
				case !gotErr && wantErr:
					t.Fatalf("got %#v, want error", outFunc)
				case gotErr && wantErr:
					return
				}
				if outFunc(outVal) == nil {
					t.Error("got nil")
				}
			})
		})
	}
}
