package cache

import (
	"sync"
	"time"
)

type IntValue struct {
	mutex    sync.Mutex
	value    int
	last     time.Time
	validity time.Duration
}

func NewIntValue(validity time.Duration) *IntValue {
	if validity == 0 {
		validity = 1 * time.Minute
	}
	return &IntValue{
		mutex:    sync.Mutex{},
		validity: validity,
	}
}

func (v *IntValue) HasValue() bool {
	v.mutex.Lock()
	defer v.mutex.Unlock()

	return v.hasValue()
}

func (v *IntValue) Set(value int) {
	v.mutex.Lock()
	defer v.mutex.Unlock()

	v.set(value)
}

func (v *IntValue) GetCached() (int, bool) {
	v.mutex.Lock()
	defer v.mutex.Unlock()

	if !v.hasValue() {
		return 0, false
	}
	return v.value, true
}

func (v *IntValue) Get(origin func() (int, error)) (int, error) {
	v.mutex.Lock()
	defer v.mutex.Unlock()

	if v.hasValue() {
		return v.value, nil
	}
	value, err := origin()
	if err != nil {
		return 0, err
	}
	v.set(value)
	return v.value, nil
}

func (v *IntValue) hasValue() bool {
	if v.last.IsZero() {
		return false
	}
	return v.last.Add(v.validity).After(time.Now())
}

func (v *IntValue) set(value int) {
	v.last = time.Now()
	v.value = value
}
