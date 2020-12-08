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
	"math"
	"sync/atomic"
)

// MajorVersion returns the major version of the Emacs instance in which the
// module is loaded.  It can only be called after the module has been loaded.
// Module functions and functions registered by OnInit can call MajorVersion if
// they have been called from Emacs.  MajorVersion panics if the module isn’t
// yet initialized.
func MajorVersion() int {
	v := majorVersion.load()
	if v == 0 {
		panic("module not yet initialized")
	}
	return v
}

// Stores the Emacs major version.
type versionManager struct{ version int32 }

var majorVersion versionManager

func (m *versionManager) init(e Env) error {
	var ver Int
	if err := e.CallOut("symbol-value", &ver, Symbol("emacs-major-version")); err != nil {
		return err
	}
	if ver < 27 || ver > math.MaxInt32 {
		return fmt.Errorf("unsupported Emacs version %d", ver)
	}
	if ok := atomic.CompareAndSwapInt32(&m.version, 0, int32(ver)); !ok {
		return errors.New("major version initialized twice")
	}
	return nil
}

func (m *versionManager) load() int {
	return int(atomic.LoadInt32(&m.version))
}
