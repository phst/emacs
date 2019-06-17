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
	"log"
	"sync"
)

type gcManager struct{ warningOnce sync.Once }

var gc gcManager

// Inhibit implements a partial workaround for
// https://debbugs.gnu.org/cgi/bugreport.cgi?bug=31238.  Before calling into
// user code, entry points should temporarily inhibit automatic garbage
// collection:
//
//     defer gc.inhibit(e).restore(e)
//
// If that bug is present (Emacs 26 and below), inhibit will attempt to inhibit
// automatic garbage collection by setting gc-cons-threshold to a large value.
// This isn’t perfect; for example, the user can still call garbage-collect
// manually and trigger the bug.
func (m *gcManager) inhibit(e Env) *gcContext {
	if majorVersion.load() >= 27 {
		// Emacs 27, nothing required.
		return nil
	}
	// Swallow errors since there’s nothing we can do about them, but they
	// shouldn’t cause the initialization to spuriously fail.
	old, err := e.Call("symbol-value", Symbol("gc-cons-threshold"))
	if err != nil {
		log.Printf("couldn’t read gc-cons-threshold: %s", e.Message(err))
		return nil
	}
	new, err := e.Call("symbol-value", Symbol("most-positive-fixnum"))
	if err != nil {
		log.Printf("couldn’t read most-positive-fixnum: %s", e.Message(err))
		return nil
	}
	m.warningOnce.Do(printGCWarning)
	setGCConsThreshold(e, new)
	return &gcContext{old}
}

type gcContext struct {
	// Previous value of gc-cons-threshold.
	threshold Value
}

func (c *gcContext) restore(e Env) {
	if c == nil {
		// Emacs 27, nothing required.
		return
	}
	setGCConsThreshold(e, c.threshold)
}

func setGCConsThreshold(e Env, v Value) {
	if _, err := e.Call("set", Symbol("gc-cons-threshold"), v); err != nil {
		// Swallow this error since there’s nothing we can do about it,
		// but it shouldn’t cause functions to spuriously fail.
		log.Printf("couldn’t set gc-cons-threshold: %s", e.Message(err))
	}
}

func printGCWarning() {
	log.Print("disabling automatic garbage collection to work around https://debbugs.gnu.org/cgi/bugreport.cgi?bug=31238")
}
