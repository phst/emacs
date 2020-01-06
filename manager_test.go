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

import "fmt"

// We assume that we want to allow module authors to register a Foo entity.
// Registration should be possible eagerly (if an Emacs environment is
// available) as well as lazily (before Emacs has loaded the module).  For that
// we use a Manager and pass DefineOnInit.
var foos = NewManager(RequireUniqueName | DefineOnInit)

// Define a type to hold the necessary information about a Foo.  The type needs
// to implement the QueuedItem interface.
type foo struct {
	name    Name
	message string
}

func (f foo) Define(e Env) error {
	return e.Invoke("message", Ignore{}, "defining foo %s: %s", f.name, f.message)
	// In real code you’d probably do something like
	// return e.Invoke("define-foo", Ignore{}, f.name, f.message)
}

// Define convenience functions to register definition of a Foo.
func DefineFooNow(e Env, name Name, message string) error {
	return foos.RegisterAndDefine(e, name, foo{name, message})
}

func DefineFooEventually(name Name, message string) error {
	return foos.Enqueue(name, foo{name, message})
}

func ExampleManager() {
	for _, n := range []Name{"foo-1", "foo-1", ""} {
		if err := DefineFooEventually(n, fmt.Sprintf("hi from foo %q", n)); err != nil {
			fmt.Println(err)
		}
	}
	// Output: duplicate name foo-1
}

func init() {
	// We would normally call ExampleManager here, but the test runner
	// already calls it for us.
}
