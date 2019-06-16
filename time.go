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
	"math/big"
	"reflect"
	"time"
)

// Time is a type with underlying type time.Time that knows how to convert
// itself from and to an Emacs time value.
type Time time.Time

// String formats the time as a string.  It calls time.Time.String.
func (t Time) String() string { return time.Time(t).String() }

// Emacs returns a quadruple (high low μs ps) in the same format as the Emacs
// function current-time.
func (t Time) Emacs(e Env) (Value, error) {
	var ps Picoseconds
	ps.FromTime(time.Time(t))
	return ps.Emacs(e)
}

// FromEmacs sets *t to the Go equivalent of the Emacs time information in v.
// v can have the same formats as for the function format‑time‑string: a number
// of seconds since the epoch, a pair (high low) or (high . low), a triple
// (high low μs), or a quadruple (high low μs ps).  The picosecond value is
// truncated to nanosecond precision.  If the Emacs time value would overflow a
// Go time, FromEmacs returns an error.
func (t *Time) FromEmacs(e Env, v Value) error {
	var ps Picoseconds
	if err := ps.FromEmacs(e, v); err != nil {
		return err
	}
	r, err := ps.Time()
	if err != nil {
		return err
	}
	*t = Time(r)
	return nil
}

// Duration is a type with underlying type time.Duration that knows how to
// convert itself from and to an Emacs time value.
type Duration time.Duration

// String formats the duration as a string.  It calls time.Duration.String.
func (d Duration) String() string { return time.Duration(d).String() }

// Emacs returns a quadruple (high low μs ps) in the same format as the Emacs
// function current-time.
func (d Duration) Emacs(e Env) (Value, error) {
	var ps Picoseconds
	ps.FromDuration(time.Duration(d))
	return ps.Emacs(e)
}

// FromEmacs sets *d to the Go equivalent of the Emacs time information in v.
// v can have the same formats as for the function format‑time‑string: a number
// of seconds, a pair (high low) or (high . low), a triple (high low μs), or a
// quadruple (high low μs ps).  The picosecond value is truncated to nanosecond
// precision.  If the Emacs time value would overflow a Go duration, FromEmacs
// returns an error.
func (d *Duration) FromEmacs(e Env, v Value) error {
	var ps Picoseconds
	if err := ps.FromEmacs(e, v); err != nil {
		return err
	}
	r, err := ps.Duration()
	if err != nil {
		return err
	}
	*d = Duration(r)
	return nil
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

func timeIn(v reflect.Value) In   { return Time(v.Interface().(time.Time)) }
func timeOut(v reflect.Value) Out { return (*Time)(v.Interface().(*time.Time)) }

func durationIn(v reflect.Value) In   { return Duration(v.Interface().(time.Duration)) }
func durationOut(v reflect.Value) Out { return (*Duration)(v.Interface().(*time.Duration)) }
