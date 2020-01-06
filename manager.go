// Copyright 2020 Google LLC
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
	"fmt"
	"sync"
)

// Manager simplifies registering and maintaining Emacs entities.  Often,
// entities such as functions, variables, or error symbols should be registered
// during initialization of the Go package, when Emacs hasn’t yet loaded the
// module.  In a second phase, these entities then have to be defined as soon
// as Emacs has loaded the module.  To support this form of two-phase
// initialization, Manager maintains a queue of QueuedItem objects for later
// definition.
//
// Optionally, you can use a Manager to ensure that entities are named and/or
// that names are unique within the scope of the manager.
//
// You can use NewManager to create Manager objects.  The
// zero Manager is valid and equivalent to NewManager(0).  Manager objects
// can’t be copied once they are initialized.
//
// Usually you’d create one global manager object per entity type.  All methods
// of Manager are safe for concurrent use, assuming the QueueItems aren’t
// modified after registering them, and QueueItem.Define is safe for concurrent
// use.
type Manager struct {
	mu    sync.Mutex
	flag  ManagerFlag
	queue []QueuedItem
	names map[Name]struct{}
}

// NewManager creates a new Manager object with the given flags.  If flags
// includes RequireName, then all entities registered on this manager must have
// a nonempty name.  If flags includes RequireUniqueName, then all named
// entities must have unique names.  If flags includes DefineOnInit, NewManager
// arranges for DefineQueued to be called when the module is initialized.
// NewManager(0) is equivalent to new(Manager).
//
// Having flags include RequireUniqueName but not RequireName is valid; in this
// case names are optional (i. e. the registration functions accept empty
// names), but if an entity is named, its name must be unique.
func NewManager(flags ManagerFlag) *Manager {
	m := &Manager{flag: flags}
	if flags&RequireUniqueName != 0 {
		m.names = make(map[Name]struct{})
	}
	if flags&DefineOnInit != 0 {
		OnInit(m.DefineQueued)
	}
	return m
}

// ManagerFlag defines flags for NewManager.
type ManagerFlag uint

const (
	// RequireName causes a Manager to fail if the name of an entity to be
	// managed is empty.
	RequireName ManagerFlag = 1 << iota

	// RequireUniqueName causes a Manager to fail if two named entities
	// have the same name.
	RequireUniqueName

	// DefineOnInit arranges for Manager.DefineQueued to be called when the
	// module is initialized.
	DefineOnInit

	initDone
)

// Enqueue registers a QueuedItem for later definition.  Enqueue is usually
// called from an init function, before Emacs has loaded the module.  The name
// may not be empty if the flag RequireName has been passed to NewManager.  The
// name must be unique if it’s nonempty and the flag RequireUniqueName has been
// passed to NewManager.  If neither flag was passed, the name is ignored.
// Enqueue returns an error if RegisterAndDefine has already been called.
func (m *Manager) Enqueue(name Name, item QueuedItem) error {
	return m.register(name, item, true)
}

// MustEnqueue registers a QueuedItem for later definition.  MustEnqueue is
// usually called from an init function, before Emacs has loaded the module.
// The name may not be empty if the flag RequireName has been passed to
// NewManager.  The name must be unique if it’s nonempty and the flag
// RequireUniqueName has been passed to NewManager.  If neither flag was
// passed, the name is ignored.  MustEnqueue panics if the name is invalid or
// if RegisterAndDefine has already been called.  MustEnqueue is like Enqueue,
// except that it panics on all errors.
func (m *Manager) MustEnqueue(name Name, item QueuedItem) {
	if err := m.Enqueue(name, item); err != nil {
		panic(err)
	}
}

// RegisterAndDefine registers a QueueItem and defines it immediately.  Unlike
// Enqueue and MustEnqueue, RegisterAndDefine requires a live Env object and
// therefore only works after Emacs has loaded the module.  The name may not be
// empty if the flag RequireName has been passed to NewManager.  The name must
// be unique if it’s nonempty and the flag RequireUniqueName has been passed to
// NewManager.  If neither flag was passed, the name is ignored.
func (m *Manager) RegisterAndDefine(e Env, name Name, item QueuedItem) error {
	if err := m.register(name, item, false); err != nil {
		return err
	}
	return item.Define(e)
}

// DefineQueued defines all queued items using the given environment.
// DefineQueued may be called at most once.  Usually, you should call it during
// module initialization, either using OnInit or by passing the DefineOnInit
// flag to NewManager.  Once DefineQueued has been called, Enqueue and
// MustEnqueue can no longer be used.  DefineQueued returns the error of the
// first failed definition, or nil if all definitions succeeded.
func (m *Manager) DefineQueued(e Env) error {
	for _, i := range m.drain() {
		if err := i.Define(e); err != nil {
			return err
		}
	}
	return nil
}

func (m *Manager) register(name Name, item QueuedItem, queue bool) error {
	if name != "" {
		if err := name.validate(); err != nil {
			return err
		}
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if queue && m.flag&initDone != 0 {
		return errors.New("initialization already complete")
	}
	if name == "" && m.flag&RequireName != 0 {
		return errors.New("unnamed queue item")
	}
	if name != "" && m.flag&RequireUniqueName != 0 {
		if _, dup := m.names[name]; dup {
			return fmt.Errorf("duplicate name %s", name)
		}
		// No more non-fatal errors from this point on.
		m.names[name] = struct{}{}
	}
	if queue {
		m.queue = append(m.queue, item)
	}
	return nil
}

func (m *Manager) drain() []QueuedItem {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.flag&initDone != 0 {
		panic("initialization already complete")
	}
	r := make([]QueuedItem, len(m.queue))
	copy(r, m.queue)
	m.queue = nil
	m.flag |= initDone
	return r
}

// QueuedItem is an item stored in a Manager’s queue.
type QueuedItem interface {
	// Define should define the item using the given Emacs environment.
	Define(Env) error
}
