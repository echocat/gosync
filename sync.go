package gosync

import (
	"runtime"
	"sync"
)

// Interruptable represents an object that could be interrupted.
type Interruptable interface {
	Interrupt()
	IsInterrupted() bool
	Reset() error
}

// Group is a couple of tools (like sleep, locks, conditions, ...) that are grouped
// together and could be interrupted by calling Interrupt() method.
type Group struct {
	interruptables map[Interruptable]int
	control        *sync.Mutex
	interrupted    bool
}

// TimeoutError occurs if a timeout condition is reached.
type TimeoutError struct{}

func (instance TimeoutError) Error() string {
	return "Timeout."
}

func IsTimeout(err error) bool {
	if err == nil {
		return false
	}
	if _, ok := err.(TimeoutError); ok {
		return true
	}
	if _, ok := err.(*TimeoutError); ok {
		return true
	}
	return false
}

// InterruptedError occurs if someone has called the Interrupt() method.
type InterruptedError struct{}

func (instance InterruptedError) Error() string {
	return "Interrupted."
}

func IsInterrupted(err error) bool {
	if err == nil {
		return false
	}
	if _, ok := err.(InterruptedError); ok {
		return true
	}
	if _, ok := err.(*InterruptedError); ok {
		return true
	}
	return false
}

// NewGroup creates a new SyncGroup instance.
func NewGroup() (*Group, error) {
	result := &Group{
		interruptables: map[Interruptable]int{},
		control:        new(sync.Mutex),
		interrupted:    false,
	}
	runtime.SetFinalizer(result, finalizeSyncGroup)
	return result, nil
}

func finalizeSyncGroup(instance *Group) {
	instance.Interrupt()
}

// Interrupt interrupts every action on this SyncGroup.
// After calling this method the instance is no longer usable anymore.
func (instance *Group) Interrupt() {
	instance.control.Lock()
	defer instance.control.Unlock()

	instance.interrupted = true

	for interruptable := range instance.interruptables {
		interruptable.Interrupt()
	}
}

func (instance *Group) IsInterrupted() bool {
	instance.control.Lock()
	defer instance.control.Unlock()

	return instance.interrupted
}

func (instance *Group) Reset() error {
	instance.control.Lock()
	defer instance.control.Unlock()

	for interruptable := range instance.interruptables {
		err := interruptable.Reset()
		if err != nil {
			return err
		}
	}
	return nil
}

// NewGroup creates a new sub instance of this instance.
func (instance *Group) NewGroup() (*Group, error) {
	instance.control.Lock()
	defer instance.control.Unlock()

	if instance.interrupted {
		return nil, InterruptedError{}
	}

	result, err := NewGroup()
	if err != nil {
		return nil, err
	}
	instance.add(result)
	return result, nil
}

func (instance *Group) add(what Interruptable) {
	if existing, ok := instance.interruptables[what]; ok {
		instance.interruptables[what] = existing + 1
	} else {
		instance.interruptables[what] = 1
	}
}

func (instance *Group) removeAndReturn(what Interruptable, result error) error {
	instance.control.Lock()
	defer instance.control.Unlock()
	if existing, ok := instance.interruptables[what]; ok {
		if existing <= 1 {
			delete(instance.interruptables, what)
		} else {
			instance.interruptables[what] = existing - 1
		}
	}
	return result
}

func closeChannel(c chan bool) {
	defer func() {
		p := recover()
		if p != nil {
			if s, ok := p.(string); ok {
				if s != "close of closed channel" {
					panic(p)
				}
			} else {
				panic(p)
			}
		}

	}()
	close(c)
}
