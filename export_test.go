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
	"strings"
)

func ExampleExport() {
	Export(goUppercase, Doc("Return the uppercase version of STRING.").WithUsage("string"))

	// Export panics when encountering an invalid type.
	defer func() { fmt.Println("panic:", recover()) }()
	Export(func(chan int) {}, Name("invalid-function"))
	// Output:
	// panic: function invalid-function: don’t know how to convert argument 0: Wrong type argument: go-known-type-p, "chan int"
}

// Once Emacs has successfully loaded the module, Emacs Lisp code can call
// exampleFunc under the name go-uppercase.  For example, (go-uppercase "hi")
// will return "HI".
func goUppercase(s string) string {
	return strings.ToUpper(s)
}

func init() {
	// We would normally call ExampleExport here, but the test runner
	// already calls it for us.
}
