package math

import (
	"fmt"
	"strings"
)

type Vector []float64

func (v Vector) Add(v2 Vector) Vector {
	if len(v) != len(v2) {
		panic("vectors must be the same length")
	}

	result := make(Vector, len(v))

	for i := range v {
		result[i] = v[i] + v2[i]
	}

	return result
}

func (v Vector) Subtract(v2 Vector) Vector {
	if len(v) != len(v2) {
		panic("vectors must be the same length")
	}

	result := make(Vector, len(v))

	for i := range v {
		result[i] = v[i] - v2[i]
	}

	return result
}

func (v Vector) Multiply(factor float64) Vector {
	result := make(Vector, len(v))

	for i := range v {
		result[i] = v[i] * factor
	}

	return result
}

func (v Vector) Divide(factor float64) Vector {
	result := make(Vector, len(v))

	for i := range v {
		result[i] = v[i] / factor
	}

	return result
}

func (v Vector) Format() string {
	parts := make([]string, len(v))
	for i, val := range v {
		parts[i] = fmt.Sprintf("%.4g", val)
	}
	return "[" + strings.Join(parts, ", ") + "]"
}
