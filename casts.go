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
	"fmt"
	"math"
)

func int64ToInt(i int64) (int, error) {
	if i < minInt || i > maxInt {
		return 0, OverflowError(fmt.Sprint(i))
	}
	return int(i), nil
}

func int64ToByte(i int64) (byte, error) {
	if i < 0 || i > math.MaxUint8 {
		return 0, OverflowError(fmt.Sprint(i))
	}
	return uint8(i), nil
}

const (
	minInt = -maxInt - 1
	maxInt = int64(int(^uint(0) >> 1))
)
