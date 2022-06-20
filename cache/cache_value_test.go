package cache

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEmptyHasNoStringValue(t *testing.T) {
	value := NewCacheValue[string](0)
	assert.False(t, value.HasValue())
}

func TestSetAndGetCachedStringValue(t *testing.T) {
	value := NewCacheValue[string](0)
	value.Set("TestSetAndGetCachedValue")
	assert.True(t, value.HasValue())
	result, ok := value.GetCached()
	assert.True(t, ok)
	assert.Equal(t, "TestSetAndGetCachedValue", result)
}

func TestGetStringValueFromOrigin(t *testing.T) {
	value := NewCacheValue[string](0)
	result, err := value.Get(func() (string, error) {
		return "TestGetValueFromOrigin", nil
	})
	assert.NoError(t, err)
	assert.Equal(t, "TestGetValueFromOrigin", result)
	assert.True(t, value.HasValue())
}

func TestGetErrorFromOriginStringValue(t *testing.T) {
	value := NewCacheValue[string](0)
	result, err := value.Get(func() (string, error) {
		return "something", errors.New("error")
	})
	assert.Error(t, err)
	assert.Empty(t, result)
	assert.False(t, value.HasValue())
}

func TestGetStringValueFromOriginOnlyOnce(t *testing.T) {
	value := NewCacheValue[string](0)
	result, err := value.Get(func() (string, error) {
		return "TestGetValueFromOriginOnlyOnce", nil
	})
	assert.NoError(t, err)
	assert.Equal(t, "TestGetValueFromOriginOnlyOnce", result)
	assert.True(t, value.HasValue())

	// there's a valid value in cache so origin function should not be called
	result, err = value.Get(func() (string, error) {
		return "error", errors.New("error")
	})
	assert.NoError(t, err)
	assert.Equal(t, "TestGetValueFromOriginOnlyOnce", result)
	assert.True(t, value.HasValue())
}

func TestEmptyHasNoIntValue(t *testing.T) {
	value := NewCacheValue[int](0)
	assert.False(t, value.HasValue())
}

func TestSetAndGetCachedIntValue(t *testing.T) {
	value := NewCacheValue[int](0)
	value.Set(111)
	assert.True(t, value.HasValue())
	result, ok := value.GetCached()
	assert.True(t, ok)
	assert.Equal(t, 111, result)
}

func TestGetIntValueFromOrigin(t *testing.T) {
	value := NewCacheValue[int](0)
	result, err := value.Get(func() (int, error) {
		return 112, nil
	})
	assert.NoError(t, err)
	assert.Equal(t, 112, result)
	assert.True(t, value.HasValue())
}

func TestGetErrorFromOriginIntValue(t *testing.T) {
	value := NewCacheValue[int](0)
	result, err := value.Get(func() (int, error) {
		return 11, errors.New("error")
	})
	assert.Error(t, err)
	assert.Empty(t, result)
	assert.False(t, value.HasValue())
}

func TestGetIntValueFromOriginOnlyOnce(t *testing.T) {
	value := NewCacheValue[int](0)
	result, err := value.Get(func() (int, error) {
		return 113, nil
	})
	assert.NoError(t, err)
	assert.Equal(t, 113, result)
	assert.True(t, value.HasValue())

	// there's a valid value in cache so origin function should not be called
	result, err = value.Get(func() (int, error) {
		return 0, errors.New("error")
	})
	assert.NoError(t, err)
	assert.Equal(t, 113, result)
	assert.True(t, value.HasValue())
}
