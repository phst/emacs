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
	"math/big"
	"time"
)

func ExampleEnv_ShouldQuit() {
	Export(mersennePrimeP, Doc("Return whether 2^N − 1 is probably prime."), Usage("N"))
}

func mersennePrimeP(e Env, n uint) bool {
	tick := time.NewTicker(100 * time.Millisecond)
	defer tick.Stop()
	// Start long-running operation in another goroutine.  Note that we
	// don’t pass any Env or Value values to the goroutine.
	ch := make(chan bool)
	go testMersennePrime(n, ch)
	// Wait for either the operation to finish or the user to quit.
	for {
		select {
		case r := <-ch:
			return r
		case <-tick.C:
			if e.ProcessInput() != Continue {
				log.Print("quitting")
				return false // Emacs will ignore the return value
			}
		}
	}
}

func testMersennePrime(n uint, ch chan<- bool) {
	x := big.NewInt(1)
	x.Lsh(x, n)
	x.Sub(x, one)
	ch <- x.ProbablyPrime(10)
	log.Print("testMersennePrime finished")
}

var one = big.NewInt(1)

func init() {
	ExampleEnv_ShouldQuit()
}
