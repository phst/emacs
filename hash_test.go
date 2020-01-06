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
	"reflect"
)

func init() {
	ERTTest(hashRoundtripEmpty)
	ERTTest(hashRoundtripFloatString)
}

func hashRoundtripEmpty(e Env) error {
	a := Hash{Eq, nil}
	v, err := a.Emacs(e)
	if err != nil {
		return fmt.Errorf("couldn’t convert hashtable %#v to Emacs: %s", a, e.Message(err))
	}
	b := &HashOut{New: nil} // never called
	if err := b.FromEmacs(e, v); err != nil {
		return fmt.Errorf("couldn’t convert hashtable from Emacs: %s", e.Message(err))
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
		return fmt.Errorf("couldn’t convert hashtable %#v to Emacs: %s", a, e.Message(err))
	}
	b := &HashOut{New: func() (Out, Out) { return new(Float), new(String) }}
	if err := b.FromEmacs(e, v); err != nil {
		return fmt.Errorf("couldn’t convert hashtable from Emacs: %s", e.Message(err))
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
