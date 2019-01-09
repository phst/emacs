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
	"math/big"
	"testing"
	"testing/quick"
	"time"
)

func TestTime(t *testing.T) {
	f := func(s, ns int64) bool {
		u := time.Unix(s, ns)
		var ps Picoseconds
		ps.FromTime(u)
		v, err := ps.Time()
		if err != nil {
			t.Logf("u = %s, ps = %s, err = %s", u, &ps, err)
			return false
		}
		if !u.Equal(v) {
			t.Logf("u = %s, v = %s, ps = %s", u, v, &ps)
			return false
		}
		return true
	}
	if err := quick.Check(f, nil); err != nil {
		t.Error(err)
	}
}

func TestDuration(t *testing.T) {
	f := func(u time.Duration) bool {
		var ps Picoseconds
		ps.FromDuration(u)
		v, err := ps.Duration()
		if err != nil {
			t.Logf("u = %s, ps = %s, err = %s", u, &ps, err)
			return false
		}
		if u != v {
			t.Logf("u = %s, v = %s, ps = %s", u, v, &ps)
			return false
		}
		return true
	}
	if err := quick.Check(f, nil); err != nil {
		t.Error(err)
	}
}

func TestQuad(t *testing.T) {
	f := func(high int64, low, μsec, psec uint16) bool {
		u := [4]int64{high, int64(low), int64(μsec), int64(psec)}
		var ps Picoseconds
		ps.fromQuad(u[0], u[1], u[2], u[3])
		var v [4]big.Int
		ps.quad(&v[0], &v[1], &v[2], &v[3])
		for i := range u {
			if !v[i].IsInt64() || u[i] != v[i].Int64() {
				t.Logf("u[%d] = %d, v[%d] = %s, ps = %s", i, u[i], i, &v[i], &ps)
				return false
			}
		}
		return true
	}
	if err := quick.Check(f, nil); err != nil {
		t.Error(err)
	}
}
