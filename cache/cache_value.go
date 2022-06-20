package cache

import (
	"sync"
	"time"
)

type CacheValue[T comparable] struct {
	mutex    sync.Mutex
	value    T
	last     time.Time
	validity time.Duration
}

func NewCacheValue[T comparable](validity time.Duration) *CacheValue[T] {
	if validity == 0 {
		validity = 1 * time.Minute
	}
	return &CacheValue[T]{
		mutex:    sync.Mutex{},
		validity: validity,
	}
}

func (v *CacheValue[T]) HasValue() bool {
	v.mutex.Lock()
	defer v.mutex.Unlock()

	return v.hasValue()
}

func (v *CacheValue[T]) Set(value T) {
	v.mutex.Lock()
	defer v.mutex.Unlock()

	v.set(value)
}

func (v *CacheValue[T]) GetCached() (T, bool) {
	v.mutex.Lock()
	defer v.mutex.Unlock()

	if !v.hasValue() {
		return *new(T), false
	}
	return v.value, true
}

func (v *CacheValue[T]) Get(origin func() (T, error)) (T, error) {
	v.mutex.Lock()
	defer v.mutex.Unlock()

	if v.hasValue() {
		return v.value, nil
	}
	value, err := origin()
	if err != nil {
		return *new(T), err
	}
	v.set(value)
	return v.value, nil
}

func (v *CacheValue[T]) hasValue() bool {
	if v.last.IsZero() {
		return false
	}
	return v.last.Add(v.validity).After(time.Now())
}

func (v *CacheValue[T]) set(value T) {
	v.last = time.Now()
	v.value = value
}
