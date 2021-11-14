package intmath

type Aggregation int

const (
	AggregateAverage Aggregation = iota
	AggregateMax
	AggregateMin
)

func Aggregate(values []int, method Aggregation) int {
	switch method {
	case AggregateAverage:
		return Avg(values)
	case AggregateMax:
		return Max(values)
	case AggregateMin:
		return Min(values)
	}
	return 0
}

func Avg(values []int) int {
	total := 0
	if len(values) == 0 {
		return total
	}
	for _, value := range values {
		total += value
	}
	return total / len(values)
}

func Max(values []int) int {
	if len(values) == 0 {
		return 0
	}
	max := values[0]
	for i := 1; i < len(values); i++ {
		if values[i] > max {
			max = values[i]
		}
	}
	return max
}

func Min(values []int) int {
	if len(values) == 0 {
		return 0
	}
	min := values[0]
	for i := 1; i < len(values); i++ {
		if values[i] < min {
			min = values[i]
		}
	}
	return min
}
