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
	"flag"
	"os"
	"os/exec"
	"testing"
)

func TestEmacs(t *testing.T) {
	emacs := os.Getenv("EMACS")
	if emacs == "" {
		emacs = "emacs"
	}
	cmd := exec.Command(
		emacs, "--quick", "--batch", "--module-assertions",
		"--load=ert", "--load="+*moduleFlag, "--load="+*ertTestsFlag,
		"--funcall=ert-run-tests-batch-and-exit",
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	// Set HOME to a nonempty value to work around
	// https://debbugs.gnu.org/cgi/bugreport.cgi?bug=36263.  Remove this
	// once that bug is either fixed on Emacs 26, or we don’t support Emacs
	// 26 any more.
	cmd.Env = append(os.Environ(), "HOME=/")
	if err := cmd.Run(); err != nil {
		t.Errorf("Emacs failed: %v", err)
	}
}

func TestMain(m *testing.M) {
	flag.Parse()
	os.Exit(m.Run())
}

var (
	moduleFlag   = flag.String("module", "", "filename of test module")
	ertTestsFlag = flag.String("ert_tests", "", "filename of Emacs Lisp file containing ERT tests")
)
