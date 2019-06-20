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
	"reflect"
	"time"
)

// #include "wrappers.h"
import "C"

// Time is a type with underlying type time.Time that knows how to convert
// itself from and to an Emacs time value.
type Time time.Time

// String formats the time as a string.  It calls time.Time.String.
func (t Time) String() string { return time.Time(t).String() }

// Emacs returns an Emacs timestamp as a pair (ticks . hz) or a quadruple
// (high low μs ps) in the same format as the Emacs function current-time.
func (t Time) Emacs(e Env) (Value, error) {
	x := time.Time(t)
	return e.makeTime(x.Unix(), x.Nanosecond())
}

// FromEmacs sets *t to the Go equivalent of the Emacs time value in v.  v can
// be any time value: nil (for the current time), a number of seconds since the
// epoch, a pair (ticks . hz), a pair (high low), a triple (high low μs), or a
// quadruple (high low μs ps).  The picosecond value is truncated to nanosecond
// precision.  If the Emacs time value would overflow a Go time, FromEmacs
// returns an error.
func (t *Time) FromEmacs(e Env, v Value) error {
	r, err := e.Time(v)
	if err != nil {
		return err
	}
	*t = Time(r)
	return nil
}

// Time returns the Go equivalent of the Emacs time value in v.  v can be any
// time value: nil (for the current time), a number of seconds since the epoch,
// a pair (ticks . hz), a pair (high low), a triple (high low μs), or a
// quadruple (high low μs ps).  The picosecond value is truncated to nanosecond
// precision.  If the Emacs time value would overflow a Go time, Time returns
// an error.
func (e Env) Time(v Value) (time.Time, error) {
	s, ns, err := e.extractTime(v)
	return time.Unix(s, int64(ns)), err
}

// Duration is a type with underlying type time.Duration that knows how to
// convert itself from and to an Emacs time value.
type Duration time.Duration

// String formats the duration as a string.  It calls time.Duration.String.
func (d Duration) String() string { return time.Duration(d).String() }

// Emacs returns an Emacs timestamp as a pair (ticks . hz) or a quadruple
// (high low μs ps) in the same format as the Emacs function current-time.
func (d Duration) Emacs(e Env) (Value, error) {
	x := time.Duration(d)
	s, ns := int64(x/time.Second), int(x%time.Second)
	if ns < 0 {
		s--
		ns += int(time.Second)
	}
	return e.makeTime(s, ns)
}

// FromEmacs sets *d to the Go equivalent of the Emacs time value in v,
// interpreted as a duration.  v can be any time value: nil (for the current
// time), a number of seconds, a pair (ticks . hz), a pair (high low), a triple
// (high low μs), or a quadruple (high low μs ps).  The picosecond value is
// truncated to nanosecond precision.  If the Emacs time value would overflow a
// Go duration, FromEmacs returns an error.
func (d *Duration) FromEmacs(e Env, v Value) error {
	r, err := e.Duration(v)
	if err != nil {
		return err
	}
	*d = Duration(r)
	return nil
}

// Duration returns the Go equivalent of the Emacs time value in v, interpreted
// as a duration.  v can be any time value: nil (for the current time), a
// number of seconds, a pair (ticks . hz) a pair (high low), a triple
// (high low μs), or a quadruple (high low μs ps).  The picosecond value is
// truncated to nanosecond precision.  If the Emacs time value would overflow a
// Go duration, Duration returns an error.
func (e Env) Duration(v Value) (time.Duration, error) {
	s, ns, err := e.extractTime(v)
	if err != nil {
		return 0, err
	}
	if s <= math.MinInt64/int64(time.Second) || s >= math.MaxInt64/int64(time.Second) {
		return 0, OverflowError(fmt.Sprintf("%d.%09ds", s, ns))
	}
	return time.Duration(s)*time.Second + time.Duration(ns), nil
}

func (e Env) makeTime(s int64, ns int) (Value, error) {
	return e.checkValue(C.make_time(e.raw(), C.struct_timespec{C.time_t(s), C.long(ns)}))
}

func (e Env) extractTime(v Value) (s int64, ns int, err error) {
	r := C.extract_time(e.raw(), v.r)
	return int64(r.value.tv_sec), int(r.value.tv_nsec), e.check(r.base)
}

func timeIn(v reflect.Value) In   { return Time(v.Interface().(time.Time)) }
func timeOut(v reflect.Value) Out { return (*Time)(v.Interface().(*time.Time)) }

func durationIn(v reflect.Value) In   { return Duration(v.Interface().(time.Duration)) }
func durationOut(v reflect.Value) Out { return (*Duration)(v.Interface().(*time.Duration)) }
