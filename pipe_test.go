// Copyright 2020, 2022, 2025 Google LLC
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
	"fmt"
	"io"
	"log"
)

func init() {
	ERTTest(pipe)
}

func pipe(e Env) error {
	buffer, err := e.Call("generate-new-buffer", String(" *temp*"))
	if err != nil {
		return err
	}
	defer e.Call("kill-buffer", buffer)
	proc, err := e.Call(
		"make-pipe-process",
		Symbol(":name"), String("test"),
		Symbol(":buffer"), buffer,
		Symbol(":coding"), Symbol("utf-8-unix"),
		Symbol(":noquery"), T,
		Symbol(":sentinel"), Nil,
	)
	if err != nil {
		return err
	}
	fd, err := e.OpenPipe(proc)
	if err != nil {
		return err
	}
	defer fd.Close()
	go write(fd, "hi from Go")
	if _, err = e.Call("accept-process-output", proc); err != nil {
		return err
	}
	v, err := e.Eval(List{Symbol("with-current-buffer"), buffer, List{Symbol("buffer-string")}})
	if err != nil {
		return err
	}
	got, err := e.Str(v)
	if err != nil {
		return err
	}
	const want = "hi from Go"
	if got != want {
		return fmt.Errorf("got %q, want %q", got, want)
	}
	return nil
}

func write(w io.WriteCloser, s string) {
	if _, err := io.WriteString(w, s); err != nil {
		log.Printf("error writing to pipe: %s", err)
	}
	if err := w.Close(); err != nil {
		log.Printf("error closing pipe: %s", err)
	}
}
