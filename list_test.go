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
	"errors"
	"fmt"
)

func ExampleIter() {
	// Assumes that env is a valid Env value.
	sum := 0
	var elem Int
	var err error
	for i := env.Iter(list, &elem, &err); i.Ok(); i.Next() {
		sum += int(elem)
	}
	if err != nil {
		panic(err)
	}
	fmt.Printf("sum is %d\n", sum)
}

func ExampleEnv_Dolist() {
	// Assumes that env is a valid Env value.
	var sum int64
	if err := env.Dolist(list, func(x Value) error {
		i, err := env.Int(x)
		if err != nil {
			return err
		}
		sum += i
		return nil
	}); err != nil {
		panic(err)
	}
	fmt.Printf("sum is %d\n", sum)
}

var list Value // invalid, for exposition only

func init() {
	ERTTest(nilIter)
	ERTTest(listIter)
}

func nilIter(e Env) error {
	list, err := e.Nil()
	if err != nil {
		return err
	}
	var got Int
	iter := e.Iter(list, &got, &err)
	if iter.Ok() {
		return errors.New("iterator for nil is OK")
	}
	if err != nil {
		return fmt.Errorf("iterator returned error %s", e.Message(err))
	}
	return nil
}

func listIter(e Env) error {
	list, err := e.List(Int(10), Int(20), Int(30))
	if err != nil {
		return err
	}
	var got Int
	iter := e.Iter(list, &got, &err)
	wantElems := []Int{10, 20, 30}
	for i, want := range wantElems {
		if !iter.Ok() {
			return fmt.Errorf("index %d: iterator is not OK", i)
		}
		if err != nil {
			return fmt.Errorf("index %d: got error %s", i, e.Message(err))
		}
		if got != want {
			return fmt.Errorf("index %d: got %d, want %d", i, got, want)
		}
		iter.Next()
	}
	if iter.Ok() {
		return errors.New("after iteration: iterator is still OK")
	}
	if err != nil {
		return fmt.Errorf("after iteration: got error %s", e.Message(err))
	}
	return nil
}
