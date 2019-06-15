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
	"math/big"
	"reflect"
	"time"
)

// #include <assert.h>
// #include <inttypes.h>
// #include <limits.h>
// #include <stddef.h>
// #include <stdint.h>
// #include <stdlib.h>
// #include <time.h>
// #include <emacs-module.h>
// static_assert(PTRDIFF_MAX <= SIZE_MAX, "unsupported architecture");
// static_assert((time_t)1.5 == 1, "unsupported architecture");
// static_assert(LONG_MAX >= 1000000000, "unsupported architecture");
// struct timespec extract_time(emacs_env *env, emacs_value value) {
// #if defined EMACS_MAJOR_VERSION && EMACS_MAJOR_VERSION >= 27
//   if ((size_t)env->size > offsetof(emacs_env, extract_time)) {
//     return env->extract_time(env, value);
//   }
// #endif
//   emacs_value list = env->funcall(env, env->intern(env, "seconds-to-time"), 1, &value);
//   emacs_value car = env->intern(env, "car");
//   emacs_value cdr = env->intern(env, "cdr");
//   intmax_t parts[4] = {0};
//   for (int i = 0; i < 4; ++i) {
//     parts[i] = env->extract_integer(env, env->funcall(env, car, 1, &list));
//     list = env->funcall(env, cdr, 1, &list);
//     if (!env->is_not_nil(env, list)) break;
//   }
//   struct timespec result;
//   assert(parts[1] >= 0 && parts[1] <= 0x10000);
//   if (__builtin_mul_overflow(parts[0], 0x10000, &result.tv_sec) ||
//       __builtin_add_overflow(result.tv_sec, parts[1], &result.tv_sec)) {
//     env->non_local_exit_signal(env, env->intern(env, "overflow-error"), env->intern(env, "nil"));
//     return result;
//   }
//   assert(parts[2] >= 0 && parts[2] < 1000000);
//   assert(parts[3] >= 0 && parts[3] < 1000000);
//   result.tv_nsec = (long) parts[2] * 1000 + (long) parts[3] / 1000;
//   return result;
// }
// emacs_value make_time(emacs_env *env, struct timespec time) {
// assert(time.tv_nsec >= 0 && time.tv_nsec < 1000000000);
// #if defined EMACS_MAJOR_VERSION && EMACS_MAJOR_VERSION >= 27
//   if ((size_t)env->size > offsetof(emacs_env, make_time)) {
//     return env->make_time(env, time);
//   }
// #endif
//   imaxdiv_t seconds = imaxdiv(time.tv_sec, 0x10000);
//   if (seconds.rem < 0) {
//     --seconds.quot;
//     seconds.rem += 0x10000;
//   }
//   assert(seconds.rem >= 0 && seconds.rem <= 0x10000);
//   ldiv_t nanos = ldiv(time.tv_nsec, 1000);
//   assert(nanos.quot >= 0 && nanos.quot < 1000000);
//   assert(nanos.rem >= 0 && nanos.rem < 1000000);
//   emacs_value args[] = {
//     env->make_integer(env, seconds.quot),
//     env->make_integer(env, seconds.rem),
//     env->make_integer(env, nanos.quot),
//     env->make_integer(env, nanos.rem * 1000)
//   };
//   return env->funcall(env, env->intern(env, "list"), 4, args);
// }
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

// Picoseconds represents a number of picoseconds.
type Picoseconds big.Int

func (ps *Picoseconds) String() string {
	i := (*big.Int)(ps)
	return fmt.Sprintf("%s ps", i.String())
}

// Time returns the time at ps picoseconds after the epoch, truncated to
// nanosecond precision.  If the time value would overflow the Go time type,
// Time returns an error.
func (ps *Picoseconds) Time() (time.Time, error) {
	var s, ns big.Int
	ps.pair(&s, &ns)
	if !s.IsInt64() || !ns.IsInt64() {
		return time.Time{}, OverflowError(ps.String())
	}
	sec := s.Int64()
	nsec := ns.Int64()
	r := time.Unix(sec, nsec)
	if r.Unix() != sec || int64(r.Nanosecond()) != nsec {
		return time.Time{}, OverflowError(ps.String())
	}
	return r, nil
}

// FromTime sets *ps to the number of picoseconds between the epoch and t.
func (ps *Picoseconds) FromTime(t time.Time) {
	ps.fromPair(t.Unix(), int64(t.Nanosecond()))
}

// Duration returns the number of picoseconds in ps truncated to nanosecond
// precision.  If the duration value would overflow the Go duration type,
// Duration returns an error.
func (ps *Picoseconds) Duration() (time.Duration, error) {
	var ns big.Int
	ps.nanos(&ns)
	if !ns.IsInt64() {
		return -1, OverflowError(ps.String())
	}
	return time.Duration(ns.Int64()), nil
}

// FromDuration sets *ps to the number of picoseconds in p
func (ps *Picoseconds) FromDuration(d time.Duration) {
	ps.fromPair(0, d.Nanoseconds())
}

// Emacs returns a quadruple (high low μs ps) in the same format as the Emacs
// function current‑time.
func (ps *Picoseconds) Emacs(e Env) (Value, error) {
	var high, low, μsec, psec big.Int
	ps.quad(&high, &low, &μsec, &psec)
	return e.List(BigInt(high), BigInt(low), BigInt(μsec), BigInt(psec))
}

// FromEmacs sets *ps to the Go equivalent of the Emacs time information in v.
// v can have the same formats as for the function format‑time‑string: a number
// of seconds, a pair (high low) or (high . low), a triple (high low μs), or a
// quadruple (high low μs ps).
func (ps *Picoseconds) FromEmacs(e Env, v Value) error {
	i, err := e.Int(v)
	if err == nil {
		ps.fromPair(i, 0)
		return nil
	}
	f, err := e.Float(v)
	if err == nil {
		big.NewFloat(f * 1e9).Int((*big.Int)(ps))
		return nil
	}
	var high Int
	var cdr Value
	if err := e.UnconsOut(v, &high, &cdr); err != nil {
		return err
	}
	if low, err := e.Int(cdr); err == nil {
		ps.fromQuad(int64(high), low, 0, 0)
		return nil
	}
	var low Int
	if err := e.UnconsOut(cdr, &low, &cdr); err != nil {
		return err
	}
	car, cdr, err := e.Uncons(cdr)
	if err != nil {
		return err
	}
	var μsec, psec int64
	if e.IsNotNil(car) {
		μsec, err = e.Int(car)
		if err != nil {
			return err
		}
	}
	car, cdr, err = e.Uncons(cdr)
	if err != nil {
		return err
	}
	if e.IsNotNil(car) {
		psec, err = e.Int(car)
		if err != nil {
			return err
		}
	}
	if e.IsNotNil(cdr) {
		return WrongTypeArgument("timep", v)
	}
	ps.fromQuad(int64(high), int64(low), μsec, psec)
	return nil
}

func (ps *Picoseconds) nanos(ns *big.Int) {
	i := (*big.Int)(ps)
	ns.Div(i, thousand)
}

func (ps *Picoseconds) pair(s, ns *big.Int) {
	ps.nanos(ns)
	s.DivMod(ns, billion, ns)
}

func (ps *Picoseconds) fromPair(s, ns int64) {
	a := (*big.Int)(ps)
	a.SetInt64(s)
	a.Mul(a, billion)
	a.Add(a, big.NewInt(ns))
	a.Mul(a, thousand)
}

func (ps *Picoseconds) quad(high, low, μsec, psec *big.Int) {
	i := (*big.Int)(ps)
	μsec.DivMod(i, million, psec)
	var sec big.Int
	sec.DivMod(μsec, million, μsec)
	low.And(&sec, lowMask)
	high.Rsh(&sec, 16)
}

func (ps *Picoseconds) fromQuad(high, low, μsec, psec int64) {
	a := (*big.Int)(ps)
	a.SetInt64(high)
	a.Lsh(a, 16)
	b := big.NewInt(low)
	a.Add(a, b)
	a.Mul(a, million)
	a.Add(a, b.SetInt64(μsec))
	a.Mul(a, million)
	a.Add(a, b.SetInt64(psec))
}

var (
	thousand = big.NewInt(1e3)
	million  = big.NewInt(1e6)
	billion  = big.NewInt(1e9)
	lowMask  = big.NewInt(0xFFFF)
)

func (e Env) makeTime(s int64, ns int) (Value, error) {
	return e.checkRaw(C.make_time(e.raw(), C.struct_timespec{C.time_t(s), C.long(ns)}))
}

func (e Env) extractTime(v Value) (s int64, ns int, err error) {
	t := C.extract_time(e.raw(), v.r)
	return int64(t.tv_sec), int(t.tv_nsec), e.check()
}

func timeIn(v reflect.Value) In   { return Time(v.Interface().(time.Time)) }
func timeOut(v reflect.Value) Out { return (*Time)(v.Interface().(*time.Time)) }

func durationIn(v reflect.Value) In   { return Duration(v.Interface().(time.Duration)) }
func durationOut(v reflect.Value) Out { return (*Duration)(v.Interface().(*time.Duration)) }
