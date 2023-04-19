// Copyright 2020, 2023 Google LLC
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
	"hash/fnv"
	"io"
	"reflect"
	"sync/atomic"
)

func init() {
	ERTTest(hashRoundtripEmpty)
	ERTTest(hashRoundtripFloatString)
	ERTTest(customHasher)
}

func hashRoundtripEmpty(e Env) error {
	a := Hash{Eq, nil}
	v, err := a.Emacs(e)
	if err != nil {
		return fmt.Errorf("couldn’t convert hash table %#v to Emacs: %s", a, e.Message(err))
	}
	b := &HashOut{New: nil} // never called
	if err := b.FromEmacs(e, v); err != nil {
		return fmt.Errorf("couldn’t convert hash table from Emacs: %s", e.Message(err))
	}
	if len(b.Data) != 0 {
		return fmt.Errorf("got %#v, want an empty map", b.Data)
	}
	return nil
}

func hashRoundtripFloatString(e Env) error {
	a := Hash{Eql, map[In]In{Float(1.0): String("foo"), Float(-0.7): String("bar")}}
	v, err := a.Emacs(e)
	if err != nil {
		return fmt.Errorf("couldn’t convert hash table %#v to Emacs: %s", a, e.Message(err))
	}
	b := &HashOut{New: func() (Out, Out) { return new(Float), new(String) }}
	if err := b.FromEmacs(e, v); err != nil {
		return fmt.Errorf("couldn’t convert hash table from Emacs: %s", e.Message(err))
	}
	got := make(map[Float]String)
	for k, v := range b.Data {
		got[*k.(*Float)] = *v.(*String)
	}
	want := map[Float]String{1.0: "foo", -0.7: "bar"}
	if !reflect.DeepEqual(got, want) {
		return fmt.Errorf("got %#v, want %#v", got, want)
	}
	return nil
}

func customHasher(e Env) error {
	callsBefore := fnvHasher.loadCalls()
	in := map[In]In{
		String("hello"): Int(123),
		String("world"): Int(456),
	}
	hash := Hash{fnvHashTest, in}
	if _, err := hash.Emacs(e); err != nil {
		return err
	}
	callsAfter := fnvHasher.loadCalls()
	gotCalls := callsAfter - callsBefore
	wantCalls := uint64(len(in))
	if gotCalls < wantCalls {
		return fmt.Errorf("not enough calls to Go hasher (got %d, want at least %d)", gotCalls, wantCalls)
	}
	return nil
}

var (
	fnvHasher   = new(stringHash)
	fnvHashTest = RegisterHashTest("go-fnv", fnvHasher)
)

// stringHash is a CustomHasher that hashes strings using the Fowler–Noll–Vo
// hash function.  It also records how often it has been called.
type stringHash struct{ calls uint64 }

func (h *stringHash) Hash(e Env, v Value) (int64, error) {
	atomic.AddUint64(&h.calls, 1)
	s, err := e.Str(v)
	if err != nil {
		return 0, err
	}
	// We use a 32-bit hasher to avoid integer overflows in Emacsen that
	// don’t support big integers.
	hash := fnv.New32()
	if _, err := io.WriteString(hash, s); err != nil {
		return 0, err
	}
	return int64(hash.Sum32()), nil
}

func (h *stringHash) Equal(e Env, a, b Value) (bool, error) {
	s, err := e.Str(a)
	if err != nil {
		return false, err
	}
	t, err := e.Str(b)
	if err != nil {
		return false, err
	}
	return s == t, nil
}

func (h *stringHash) loadCalls() uint64 {
	return atomic.LoadUint64(&h.calls)
}
