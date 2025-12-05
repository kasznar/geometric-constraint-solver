package math

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVector_Subtract(t *testing.T) {
	v := Vector{1, 2, 3}
	v2 := Vector{1, 1, 1}

	got := v.Subtract(v2)

	expected := Vector{0, 1, 2}

	assert.Equal(t, expected, got)
}

func TestVector_Add(t *testing.T) {
	v := Vector{1, 2, 3}
	v2 := Vector{1, 1, 1}

	got := v.Add(v2)

	expected := Vector{2, 3, 4}

	assert.Equal(t, expected, got)
}

func TestVector_Multiply(t *testing.T) {
	v := Vector{1, 2, 3}

	got := v.Multiply(2)

	expected := Vector{2, 4, 6}

	assert.Equal(t, expected, got)
}

func TestVector_Divide(t *testing.T) {
	v := Vector{1, 2, 3}

	got := v.Divide(2)

	expected := Vector{0.5, 1, 1.5}

	assert.Equal(t, expected, got)
}

func TestVector_Format(t *testing.T) {
	v := Vector{1, 2, 3}

	got := v.Format()

	expected := "[1, 2, 3]"

	assert.Equal(t, expected, got)
}
