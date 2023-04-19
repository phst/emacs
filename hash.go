// Copyright 2019, 2023 Google LLC
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

import "reflect"

// HashTest represents a test function for an Emacs hash table.
type HashTest Symbol

// String returns the hash test symbol name.
func (t HashTest) String() string {
	return Symbol(t).String()
}

// Emacs interns the hash test symbol in the default obarray and returns the
// symbol object.
func (t HashTest) Emacs(e Env) (Value, error) {
	return Symbol(t).Emacs(e)
}

// Predefined hash table tests.  To define your own test, use
// [RegisterHashTest].
const (
	Eq    HashTest = "eq"
	Eql   HashTest = "eql"
	Equal HashTest = "equal"
)

// HashTestFor returns a hash table test that is appropriate for the given
// type.  It returns [Eq] for integral types, [Eql] for floating-point types,
// and [Equal] otherwise.  HashTestFor ignores custom hash tests registered
// with [RegisterHashTest].
func HashTestFor(t reflect.Type) HashTest {
	switch t.Kind() {
	case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return Eq
	case reflect.Float32, reflect.Float64:
		return Eql
	default:
		return Equal
	}
}

// RegisterHashTest registers a new custom hash table test.  The name must be a
// unique nonempty name for the test, and hash must define the hash function
// and equality predicate.  RegisterHashTest returns name.
func RegisterHashTest(name HashTest, hash CustomHasher) HashTest {
	customHashTests.MustEnqueue(Name(name), customHashTest{name, hash})
	return name
}

// CustomHasher defines the hashing and equality functions for a custom hash
// test.  Use [RegisterHashTest] to register such a custom hash test.  The
// hashing and equality functions must be compatible, i. e., if Equal(e, a, b)
// is true, then Hash(e, a) = Hash(e, b) must be fulfilled.
type CustomHasher interface {
	// Hash returns a hash code for the given value.
	Hash(Env, Value) (int64, error)

	// Equal returns whether the given values are considered equal
	// according to this hash test.
	Equal(Env, Value, Value) (bool, error)
}

// Hash represents the data for an Emacs hash table.
type Hash struct {
	Test HashTest
	Data map[In]In
}

// Emacs creates a new hash table using the test and data in h.
func (h Hash) Emacs(e Env) (Value, error) {
	r, err := e.MakeHash(h.Test, len(h.Data))
	if err != nil {
		return Value{}, err
	}
	for key, val := range h.Data {
		if err := e.Puthash(key, val, r); err != nil {
			return Value{}, err
		}
	}
	return r, nil
}

// HashOut is an [Out] that converts an Emacs hash table to the map Data.  The
// concrete key and value types are determined by the return values of the
// [HashOut.New] function.
type HashOut struct {
	// New must return a new key and value each time it’s called.
	New func() (Out, Out)

	// FromEmacs fills Data with the pairs from the hash table.
	Data map[Out]Out
}

// FromEmacs sets h.Data to a new map containing the same key–value pairs as
// the Emacs hash table in v.  It returns an error if v is not a hash table.
// FromEmacs calls h.New for each key–value pair in v.  h.New must return a new
// pair of Out values for the pair’s key and value.  If FromEmacs returns an
// error, h.Data is not modified.
//
// FromEmacs ignores the Emacs hash table test for v.  This means that there may
// be multiple Emacs keys mapping to a single Go key if the hash functions
// aren’t consistent.  For example, an Emacs hash table with string keys and
// hash test eq may contain two keys that are equal when converted to Go
// strings.  In such a case, FromEmacs returns an error.
func (h *HashOut) FromEmacs(e Env, v Value) error {
	m := make(map[Out]Out)
	f := func(rawKey, rawVal Value) error {
		key, val := h.New()
		if err := key.FromEmacs(e, rawKey); err != nil {
			return err
		}
		if err := val.FromEmacs(e, rawVal); err != nil {
			return err
		}
		if _, dup := m[key]; dup {
			return duplicateKey.Error(rawKey, v)
		}
		m[key] = val
		return nil
	}
	if err := e.Maphash(f, v); err != nil {
		return err
	}
	h.Data = m
	return nil
}

var duplicateKey = DefineError("go-duplicate-key", "Duplicate map key", baseError)

// MakeHash returns a new hash table with the given test and size hint.
func (e Env) MakeHash(test HashTest, sizeHint int) (Value, error) {
	return e.Call("make-hash-table", Symbol(":test"), test, Symbol(":size"), Int(sizeHint))
}

// Gethash returns the hash table value with the given key.  ok specifies
// whether the key is present.
func (e Env) Gethash(key In, table Value) (value Value, ok bool, err error) {
	// Unique dummy object.
	dummy, err := e.Cons(nil, nil)
	if err != nil {
		return Value{}, false, err
	}
	val, err := e.GethashDef(key, table, dummy)
	if err != nil {
		return Value{}, false, err
	}
	if e.Eq(val, dummy) {
		return Value{}, false, nil
	}
	return val, true, nil
}

// GethashDef returns the hash table value with the given key.  If the key is
// not present, it returns def.
func (e Env) GethashDef(key In, table Value, def Value) (Value, error) {
	return e.Call("gethash", key, table, def)
}

// Puthash sets the value of key in table to value.
func (e Env) Puthash(key, value In, table Value) error {
	_, err := e.Call("puthash", key, value, table)
	return err
}

// Maphash calls fun for each key–value pair in table.  The order is arbitrary.
// If table is modified during iteration, the results are unpredictable.
func (e Env) Maphash(fun func(key, val Value) error, table Value) error {
	fv, delete, err := e.Lambda(fun)
	if err != nil {
		return err
	}
	defer delete()
	_, err = e.Call("maphash", fv, table)
	return err
}

type hashIn struct {
	test       HashTest
	key, value InFunc
}

func (i hashIn) call(v reflect.Value) In {
	return makeHash{i, v}
}

type makeHash struct {
	hashIn
	reflect.Value
}

func (m makeHash) Emacs(e Env) (Value, error) {
	if !m.IsValid() {
		return Value{}, WrongTypeArgument("go-valid-reflect-p", String(m.String()))
	}
	r, err := e.MakeHash(m.test, m.Len())
	if err != nil {
		return Value{}, err
	}
	for _, key := range m.MapKeys() {
		val := m.MapIndex(key)
		if err := e.Puthash(m.key(key), m.value(val), r); err != nil {
			return Value{}, err
		}
	}
	return r, nil
}

type hashOut struct{ key, value OutFunc }

func (o hashOut) call(v reflect.Value) Out {
	return getHash{o, v}
}

type getHash struct {
	hashOut
	reflect.Value
}

func (g getHash) FromEmacs(e Env, v Value) error {
	u := g.Elem()
	t := u.Type()
	m := reflect.MakeMap(t)
	f := func(rawKey, rawVal Value) error {
		key := reflect.New(t.Key())
		if err := g.key(key).FromEmacs(e, rawKey); err != nil {
			return err
		}
		val := reflect.New(t.Elem())
		if err := g.value(val).FromEmacs(e, rawVal); err != nil {
			return err
		}
		m.SetMapIndex(key.Elem(), val.Elem())
		return nil
	}
	if err := e.Maphash(f, v); err != nil {
		return err
	}
	u.Set(m)
	return nil
}

var customHashTests = NewManager(RequireName | RequireUniqueName | DefineOnInit)

type customHashTest struct {
	name HashTest
	hash CustomHasher
}

func (t customHashTest) Define(e Env) error {
	test := AutoLambda(t.hash.Equal, Doc("Hash equality function for "+t.name), Usage("a b"))
	hash := AutoLambda(t.hash.Hash, Doc("Hash function for "+t.name), Usage("o"))
	_, err := e.Call("define-hash-table-test", t.name, test, hash)
	return err
}
