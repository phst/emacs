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

import "time"

func ExampleEnv_Call() {
	// Assumes that env is a valid Env value.
	_, err := env.Call("message", String("It is %s"), Time(time.Now()))
	if err != nil {
		panic(err)
	}
}

func ExampleEnv_CallOut() {
	// Assumes that env is a valid Env value.
	var now Time
	err := env.CallOut("current-time", &now)
	if err != nil {
		panic(err)
	}
}

func ExampleEnv_Invoke() {
	// Assumes that env is a valid Env value.
	var now time.Time
	err := env.Invoke("current-time", &now)
	if err != nil {
		panic(err)
	}
}

var env Env // invalid, for exposition only
