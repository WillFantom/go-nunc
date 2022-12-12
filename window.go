package nunc

import (
	"fmt"
	"sync"
)

// Window is a fixed size FIFO buffer that is thread safe. This creates a
// sliding window, where once the window reaches capacity it removes the first
// (oldest) element to add new elements.
type Window[T any] struct {
	data  []T
	lock  *sync.RWMutex
	count uint64
}

// NewWindow returns a new empty Window with the given capacity. If the capcity
// is less than or equal to 0, an error is returned.
func NewWindow[T any](capacity int) (*Window[T], error) {
	if capacity <= 0 {
		return nil, fmt.Errorf("window capacity must be greater than 0")
	}
	return &Window[T]{
		data:  make([]T, capacity),
		lock:  &sync.RWMutex{},
		count: 0,
	}, nil
}

// Push adds a new value to the window, overwritting the oldest element if the
// window is full. Returned is the count of the total number of datapoints
// pushed to the window.
func (w *Window[T]) Push(value T) uint64 {
	w.lock.Lock()
	defer w.lock.Unlock()
	w.push(value)
	return w.count
}

// PushGet performs a push to the window and returns the total count of entries
// added to window along with the window immediately after the push.
func (w *Window[T]) PushGet(value T) (uint64, []T) {
	w.lock.Lock()
	defer w.lock.Unlock()
	w.push(value)
	return w.count, w.get(false)
}

// PushGetFull performs a push to the window and returns the total count of
// entries added to window along with the window immediately after the push. If
// the window has not yet reached capacity, the data returned will be nil.
func (w *Window[T]) PushGetFull(value T) (uint64, []T) {
	w.lock.Lock()
	defer w.lock.Unlock()
	w.push(value)
	return w.count, w.get(true)
}

// Get returns the data as a single slice, where the oldest element is at the
// start and the newest is at the end.
func (w *Window[T]) Get() []T {
	w.lock.RLock()
	defer w.lock.RUnlock()
	return w.get(false)
}

// GetFull returns the data as a single slice, where the oldest element is at
// the start and the newest is at the end. If the window is not yet full, nil is
// returned.
func (w *Window[T]) GetFull() []T {
	w.lock.RLock()
	defer w.lock.RUnlock()
	return w.get(true)
}

// Len returns the number of elements that are in the window. If the window is
// full, this will be equal to the capacity.
func (w *Window[T]) Len() uint64 {
	if w.count >= uint64(cap(w.data)) {
		return uint64(cap(w.data))
	}
	return w.marker()
}

// Cap returns the window's maximum capacity. This in no way indicates how many
// elements are currently in the window.
func (w *Window[T]) Cap() int {
	return cap(w.data)
}

// Count returns the number of values that have been added to the window. Note
// that this returns a count of **all** elements that have been added, not just
// ones still present within the window.
func (w *Window[T]) Count() uint64 {
	return w.count
}

// Full returns true if the number of elements in the window equal that of the
// capacity.
func (w *Window[T]) Full() bool {
	return int(w.Len()) == cap(w.data)
}

// marker represents the index of the data slice that contains the oldest
// element.
func (w *Window[T]) marker() uint64 {
	return (w.count % uint64(cap(w.data)))
}

// push overwrites the oldest element with a new value and increases the total
// count.
func (w *Window[T]) push(value T) {
	w.data[w.marker()] = value
	w.count++
}

// get returns the data as a single slice, where the oldest element is at the
// start and the newest is at the end.
func (w *Window[T]) get(onlyFull bool) []T {
	if w.data == nil || (!w.Full() && onlyFull) {
		return nil
	}
	m := w.marker()
	return append(w.data[m:w.Len()], w.data[:m]...)
}
