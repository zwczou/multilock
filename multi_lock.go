// https://medium.com/@kf99916/multiple-lock-based-on-input-in-golang-74931a3c8230
package multilock

import (
	"sync"
	"sync/atomic"
)

type refCounter struct {
	counter int64
	lock    *sync.RWMutex
}

// MultiLock is the main interface for lock base on key
type MultiLock interface {
	// Lock base on the key
	Lock(interface{})

	// RLock lock the rw for reading
	RLock(interface{})

	// Unlock the key
	Unlock(interface{})

	// RUnlock the the read lock
	RUnlock(interface{})
}

// A multi lock type
type lock struct {
	inUse sync.Map
	pool  *sync.Pool
}

func (l *lock) Lock(key interface{}) {
	m := l.getLocker(key)
	atomic.AddInt64(&m.counter, 1)
	m.lock.Lock()
}

func (l *lock) RLock(key interface{}) {
	m := l.getLocker(key)
	atomic.AddInt64(&m.counter, 1)
	m.lock.RLock()
}

func (l *lock) Unlock(key interface{}) {
	m := l.getLocker(key)
	m.lock.Unlock()
	l.putBackInPool(key, m)
}

func (l *lock) RUnlock(key interface{}) {
	m := l.getLocker(key)
	m.lock.RUnlock()
	l.putBackInPool(key, m)
}

func (l *lock) putBackInPool(key interface{}, m *refCounter) {
	atomic.AddInt64(&m.counter, -1)
	if m.counter <= 0 {
		l.pool.Put(m.lock)
		l.inUse.Delete(key)
	}
}

func (l *lock) getLocker(key interface{}) *refCounter {
	res, _ := l.inUse.LoadOrStore(key, &refCounter{
		counter: 0,
		lock:    l.pool.Get().(*sync.RWMutex),
	})

	return res.(*refCounter)
}

// NewMultiLock create a new multiple lock
func NewMultiLock() MultiLock {
	return &lock{
		pool: &sync.Pool{
			New: func() interface{} {
				return &sync.RWMutex{}
			},
		},
	}
}

var (
	defaultLock = NewMultiLock()
)

func Lock(key interface{}) {
	defaultLock.Lock(key)
}

func RLock(key interface{}) {
	defaultLock.RLock(key)
}

func Unlock(key interface{}) {
	defaultLock.Unlock(key)
}

func RUnlock(key interface{}) {
	defaultLock.RUnlock(key)
}
