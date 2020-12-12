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

// #include "wrappers.h"
import "C"

import (
	"fmt"
	"os"
)

// OpenPipe opens a writable pipe to the given Emacs pipe process.  The pipe
// process must have been created with make-pipe-process.  You can write to the
// returned pipe to provide input to the pipe process.
func (e Env) OpenPipe(process Value) (*os.File, error) {
	i := C.open_channel(e.raw(), process.r)
	if err := e.check(i.base); err != nil {
		return nil, err
	}
	fd := uintptr(i.value)
	return os.NewFile(fd, fmt.Sprintf("/dev/fd/%d", fd)), nil
}
