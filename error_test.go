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

import "fmt"

func ExampleError() {
	Export(goError, Doc("Signal an error of type ‘example-error’.").WithUsage("int float vec"))

	err := exampleError.Error(Int(123), Vector{String("foo"), Float(0.7), T})
	fmt.Println(err)
	// Output: Example error: 123, ["foo" 0.7 t]
}

func goError(f float32, v []uint16) (uint8, error) {
	return 55, exampleError.Error(Float(f), NewIn(v))
}

var exampleError = DefineError("example-error", "Example error")
