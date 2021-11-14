package cache

import (
	"sync"
	"time"
)

type StringValue struct {
	mutex    sync.Mutex
	value    string
	last     time.Time
	validity time.Duration
}

func NewStringValue(validity time.Duration) *StringValue {
	if validity == 0 {
		validity = 1 * time.Minute
	}
	return &StringValue{
		mutex:    sync.Mutex{},
		validity: validity,
	}
}

func (v *StringValue) HasValue() bool {
	v.mutex.Lock()
	defer v.mutex.Unlock()

	return v.hasValue()
}

func (v *StringValue) Set(value string) {
	v.mutex.Lock()
	defer v.mutex.Unlock()

	v.set(value)
}

func (v *StringValue) GetCached() (string, bool) {
	v.mutex.Lock()
	defer v.mutex.Unlock()

	if !v.hasValue() {
		return "", false
	}
	return v.value, true
}

func (v *StringValue) Get(origin func() (string, error)) (string, error) {
	v.mutex.Lock()
	defer v.mutex.Unlock()

	if v.hasValue() {
		return v.value, nil
	}
	value, err := origin()
	if err != nil {
		return "", err
	}
	v.set(value)
	return v.value, nil
}

func (v *StringValue) hasValue() bool {
	if v.last.IsZero() {
		return false
	}
	return v.last.Add(v.validity).After(time.Now())
}

func (v *StringValue) set(value string) {
	v.last = time.Now()
	v.value = value
}
