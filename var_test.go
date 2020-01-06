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

func ExampleVar() {
	// Var panics when trying to export a variable name twice.
	defer func() { fmt.Println("panic:", recover()) }()
	Var("go-var", nil, "")
	// Output: panic: duplicate name go-var
}

var _ = Var("go-var", String("hi"), "Example variable.")
