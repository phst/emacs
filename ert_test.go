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

import "log"

func ExampleERTTest() {
	ERTTest(exampleERTTest, Name("example-ert-test"), Doc("Run an example ERT test."))
}

// Once Emacs has successfully loaded the module, ERT will see a test named
// example-ert-test that calls exampleERTTest.
func exampleERTTest(e Env) error {
	log.Print("running example test")
	return nil
}

func init() {
	// We would normally call ExampleERTTest here, but the test runner
	// already calls it for us.
}
