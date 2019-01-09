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
	"time"
)

func ExampleImport() {
	Import("format-time-string", &formatTimeString)
	Export(goPrintNow)

	// Import panics when encountering nonconvertible types.
	defer func() { fmt.Println("panic:", recover()) }()
	Import("format-time-string", &invalidType)
	// Output:
	// panic: can’t import format-time-string: don’t know how to convert argument 1: Wrong type argument: go-known-type-p, "chan int"
}

var (
	formatTimeString func(Env, string, time.Time, bool) (string, error)

	// Can’t convert channel types to Emacs.
	invalidType func(Env, chan int) error
)

func goPrintNow(e Env, format string) (string, error) {
	// Functions that have access to a live environment can now call the
	// Emacs message function, like so:
	r, err := formatTimeString(e, format, time.Now(), true)
	if err != nil {
		return "", err
	}
	_, err = fmt.Println("Time from Emacs:", r)
	return r, err
}

func init() {
	// We would normally call ExampleExport here, but the test runner
	// already calls it for us.
}
