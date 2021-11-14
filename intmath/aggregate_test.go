package intmath

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAggregateMax(t *testing.T) {
	testData := []struct {
		input []int
		max   int
	}{
		{nil, 0},
		{[]int{}, 0},
		{[]int{1}, 1},
		{[]int{1, 2}, 2},
		{[]int{2, 1}, 2},
	}

	for _, testItem := range testData {
		assert.Equal(t, testItem.max, Aggregate(testItem.input, AggregateMax))
	}
}

func TestAggregateMin(t *testing.T) {
	testData := []struct {
		input []int
		max   int
	}{
		{nil, 0},
		{[]int{}, 0},
		{[]int{1}, 1},
		{[]int{1, 2}, 1},
		{[]int{2, 1}, 1},
	}

	for _, testItem := range testData {
		assert.Equal(t, testItem.max, Aggregate(testItem.input, AggregateMin))
	}
}

func TestAggregateAvg(t *testing.T) {
	testData := []struct {
		input []int
		max   int
	}{
		{nil, 0},
		{[]int{}, 0},
		{[]int{1}, 1},
		{[]int{1, 3}, 2},
		{[]int{3, 1}, 2},
	}

	for _, testItem := range testData {
		assert.Equal(t, testItem.max, Aggregate(testItem.input, AggregateAverage))
	}
}
