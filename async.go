// Copyright 2021, 2023 Google LLC
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
	"io"
	"log"
	"net"
	"sync/atomic"
)

// Async manages asynchronous operations.  Create a new Async object using
// [NewAsync]; the zero Async isn’t valid, and Async objects may not be copied
// once created.  An asynchronous operation can use [Async.Start] to allocate
// an operation handle and a channel.  It can then start an operation in the
// background and report its result asynchronously using the channel.
//
// Async requires a way to notify Emacs about pending asynchronous results; use
// [NotifyWriter] or [NotifyListener] to create notification channels.  When
// notified about asynchronous operations, Emacs should then use [Async.Flush]
// to return the pending results.
//
// Async doesn’t prescribe any specific programming model on the Emacs side;
// the example uses the [“aio” package].
//
// [“aio” package]: https://github.com/skeeto/emacs-aio
type Async struct {
	notifyCh    chan<- struct{}
	promiseCh   chan AsyncData
	nextPromise uint64
}

// NewAsync creates a new [Async] object.  It will use the given notification
// channel to signal completion of an asynchronous operation to Emacs; use
// [NotifyWriter] or [NotifyListener] to create usable channels.
func NewAsync(notifyCh chan<- struct{}) *Async {
	if notifyCh == nil {
		panic("nil notification channel")
	}
	return &Async{
		notifyCh: notifyCh,
		// We need to use some buffering to avoid a deadlock.  Any
		// nonzero value should do the job, but a size that’s not too
		// tiny might be a bit more efficient since it allows Emacs to
		// request multiple asynchronous results at once.
		promiseCh: make(chan AsyncData, 10),
	}
}

// Start starts a new asynchronous operation.  It returns a handle for the
// operation and a channel.  Return the operation handle to Emacs so that it
// can associate pending operations with e.g. a promise structure.  The
// operation can then use the channel to write exactly one result (error or
// value); any further writes or closes are ignored and will block.  The
// typical usage pattern is:
//
//	func operation() AsyncHandle {
//	    h, ch := async.Start()
//	    go performOperation(ch)
//	    return h
//	}
//
// Here, performOperation should write the result to the channel once
// available.
func (a *Async) Start() (AsyncHandle, chan<- Result) {
	ch := make(chan Result)
	h := AsyncHandle(atomic.AddUint64(&a.nextPromise, 1))
	if h == 0 {
		panic("too many asynchronous operations")
	}
	h--
	go a.forward(h, ch)
	return h, ch
}

func (a *Async) forward(h AsyncHandle, ch <-chan Result) {
	a.promiseCh <- AsyncData{h, <-ch}
	a.notifyCh <- struct{}{}
}

// Flush returns and removes all pending asynchronous operation results.  You
// should call this method from Emacs Lisp when notified about pending
// asynchronous results.
func (a *Async) Flush() []AsyncData {
	var r []AsyncData
	for {
		select {
		case v := <-a.promiseCh:
			r = append(r, v)
		default:
			return r
		}
	}
}

// AsyncHandle is an opaque reference to a pending asynchronous operation.  Use
// [Async.Start] to create AsyncHandle objects.
type AsyncHandle uint64

// Emacs implements In.Emacs.  It returns the handle as an integer.
func (h AsyncHandle) Emacs(e Env) (Value, error) {
	return Uint(h).Emacs(e)
}

// Result contains the result of an asynchronous operation.  If Err is set,
// Value is ignored.
type Result struct {
	Value In
	Err   error
}

// AsyncData contains the result of an asynchronous operation, together with
// the associated operation handle.  It is returned by [Async.Flush].
type AsyncData struct {
	Handle AsyncHandle
	Result
}

// Emacs implements [In.Emacs].  It returns a triple (handle value error).
// If Err is set, the error element will be of the form (symbol . data).
func (d AsyncData) Emacs(e Env) (Value, error) {
	return e.List(d.Handle, d.Value, errorData(d.Err))
}

func errorData(err error) In {
	switch err := err.(type) {
	case nil:
		return Nil
	case Error:
		return Cons{err.Symbol, err.Data}
	case Signal:
		return Cons{err.Symbol, err.Data}
	default:
		return Cons{asyncError, String(err.Error())}
	}
}

var asyncError = DefineError("go-async-error", "Generic asynchronous Go error", baseError)

// NotifyWriter returns a channel that causes some arbitrary content to be
// written to the given writer whenever something is written to the channel.
// The writer will typically be either a pipe created by [Env.OpenPipe] or a
// socket connection to a Unix domain socket created by [net.Dial] or similar.
// The return value of this function is useful as argument to [NewAsync].
func NotifyWriter(w io.Writer) chan<- struct{} {
	if w == nil {
		panic("nil writer")
	}
	ch := make(chan struct{})
	go poke(w, ch)
	return ch
}

// NotifyListener returns a channel that causes some arbitrary content to be
// sent to the client of the given listener whenever something is written to
// the channel.  NotifyListener will wait (in the background) for exactly one
// client to connect to the server and then close the listener.  You can create
// the listener using [net.Listen] or similar.  A common use case is a Unix
// domain socket server; the socket name should be reported to Emacs Lisp using
// other means.  The return value of this function is useful as argument to
// [NewAsync].
func NotifyListener(listener net.Listener) chan<- struct{} {
	if listener == nil {
		panic("nil listener")
	}
	ch := make(chan struct{})
	go notifyListener(listener, ch)
	return ch
}

func notifyListener(listener net.Listener, ch <-chan struct{}) {
	defer listener.Close()
	conn := acceptOne(listener)
	defer conn.Close()
	poke(conn, ch)
}

func acceptOne(listener net.Listener) net.Conn {
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("error accepting connection: %s", err)
			continue
		}
		if err := listener.Close(); err != nil {
			log.Printf("error closing listener: %s", err)
		}
		return conn
	}
}

func poke(w io.Writer, ch <-chan struct{}) {
	b := []byte{'.'}
	for range ch {
		if _, err := w.Write(b); err != nil {
			log.Printf("can’t write to notifier: %s", err)
		}
	}
}
