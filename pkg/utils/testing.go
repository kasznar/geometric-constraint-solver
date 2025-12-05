package utils

import (
	"math"

	"github.com/stretchr/testify/assert"
)

func AlmostEqual(a, b, epsilon float64) bool {
	return math.Abs(a-b) < epsilon
}

func AssertAlmost(t assert.TestingT, a, b float64) {
	assert.Equal(t, AlmostEqual(a, b, 1e-3), true)
}
