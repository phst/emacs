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
	"log"
	"net"
	"os"
	"path/filepath"
	"sync"
)

func ExampleAsync() {
	Export(mersennePrimeAsyncP, Doc("Return a promise that resolves to a Boolean value indicating whether 2^N − 1 is probably prime."), Usage("N"))
	Export(asyncSocket, Doc("Return a filename of a socket to connect to"))
	Export(asyncFlush, Doc("Return a vector of asynchronous promise results"))
}

func mersennePrimeAsyncP(n uint16) AsyncHandle {
	h, ch := async.Start()
	boolCh := make(chan bool)
	go testMersennePrime(n, boolCh)
	go wrapBool(ch, boolCh)
	return h
}

func wrapBool(ch chan<- Result, boolCh <-chan bool) {
	ch <- Result{Bool(<-boolCh), nil}
}

var (
	asyncOnce sync.Once
	async     *Async
	socket    string
)

func asyncSocket() (string, error) {
	asyncOnce.Do(initAsync)
	if socket == "" {
		return "", errors.New("error initializing asynchronous socket")
	}
	return socket, nil
}

func asyncFlush() []AsyncData {
	return async.Flush()
}

func initAsync() {
	dir, err := os.MkdirTemp("", "emacs-")
	if err != nil {
		panic(err)
	}
	sock := filepath.Join(dir, "socket")
	listener, err := net.Listen("unix", sock)
	if err != nil {
		panic(err)
	}
	log.Printf("Created socket %s for asynchronous communication", sock)
	async = NewAsync(NotifyListener(listener))
	socket = sock
}

func init() {
	ExampleAsync()
}
