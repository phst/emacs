// Copyright 2021 Google LLC
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
	"os"
	"sync"
)

func Example_handles() {
	Export(createGoFile, Doc(`Create a file with the given NAME and return a handle to it.`), Usage("NAME"))
	Export(closeGoFile, Doc(`Close the file with the given HANDLE.
HANDLE must be opened using ‘create-go-file’.`), Usage("HANDLE"))
	Export(writeGoFile, Doc(`Write CONTENTS to the file with the given HANDLE.
Return the number of bytes written.
HANDLE must be opened using ‘create-go-file’.
CONTENTS must be a unibyte string.`), Usage("HANDLE CONTENTS"))
}

type fileHandle Uint

func createGoFile(name string) (fileHandle, error) {
	fd, err := os.Create(name)
	if err != nil {
		return 0, err
	}
	h, err := registerFile(fd)
	if err != nil {
		fd.Close()
		return 0, err
	}
	return h, nil
}

func closeGoFile(h fileHandle) error {
	fd := popFile(h)
	if fd == nil {
		return fmt.Errorf("close-go-file: invalid file handle %d", h)
	}
	return fd.Close()
}

func writeGoFile(h fileHandle, b []byte) (int, error) {
	fd := getFile(h)
	if fd == nil {
		return 0, fmt.Errorf("write-go-file: invalid file handle %d", h)
	}
	return fd.Write(b)
}

var (
	fileMu   sync.Mutex
	files    = make(map[fileHandle]*os.File)
	nextFile fileHandle
)

func registerFile(fd *os.File) (fileHandle, error) {
	fileMu.Lock()
	defer fileMu.Unlock()
	h := nextFile
	if h == math.MaxUint64 {
		fd.Close()
		return 0, errors.New("too many files")
	}
	nextFile++
	files[h] = fd
	return h, nil
}

func getFile(h fileHandle) *os.File {
	fileMu.Lock()
	defer fileMu.Unlock()
	return files[h]
}

func popFile(h fileHandle) *os.File {
	fileMu.Lock()
	defer fileMu.Unlock()
	fd := files[h]
	delete(files, h)
	return fd
}

func init() {
	Example_handles()
}
