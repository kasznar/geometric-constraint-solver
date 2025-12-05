package math

import (
	"fmt"
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestMatrix_FromColVec(t *testing.T) {
	v := Vector{1, 2, 3}
	e := Matrix{
		Vector{1},
		Vector{2},
		Vector{3},
	}

	got := NewMatrixFromColVec(v)

	assert.Equal(t, e, got)
}

func TestMatrix_New(t *testing.T) {
	e := Matrix{
		Vector{0,0,0},
		Vector{0,0,0},
		Vector{0,0,0},
	}

	got := NewMatrix(3, 3)

	assert.Equal(t, e, got)
}

func TestMatrix_AugmentVec(t *testing.T) {
	m := Matrix{
		{1, 2},
		{3, 4},
	}
	vec := Vector{5, 6}

	expected := Matrix{
		{1, 2, 5},
		{3, 4, 6},
	}

	m.AugmentVec(vec)

	assert.Equal(t, expected, m)
}

func TestMatrix_Augment(t *testing.T) {
	m := Matrix{
		{1, 2},
		{3, 4},
	}
	m2 := Matrix{
		{5, 6},
		{7, 8},
	}

	expected := Matrix{
		{1, 2, 5, 6},
		{3, 4, 7, 8},
	}

	m.Augment(m2)

	assert.Equal(t, expected, m)

}

func TestSwapMatrixRows(t *testing.T) {
	m := Matrix{
		{1, 2},
		{3, 4},
		{5, 6},
	}

	expected := Matrix{
		{1, 2},
		{5, 6},
		{3, 4},
	}

	m.SwapRows(1, 2)

	assert.Equal(t, expected, m)
}

func TestMatrixString(t *testing.T) {
	m := Matrix{
		{1, 2},
		{3, 4},
	}
	expected := "1 2\n3 4"
	assert.Equal(t, expected, m.String())
}

func TestMatrix_Copy(t *testing.T) {
	m := Matrix{
		{1, 2},
		{3, 4},
	}
	m2 := m.Copy()
	(*m2)[0][0] = 0
	(*m2)[1][1] = 0
	assert.Equal(t, Matrix{{0, 2}, {3, 0}}, *m2)
	assert.Equal(t, Matrix{{1, 2}, {3, 4}}, m)
}

func TestMatrix_Transpose(t *testing.T) {
	m := Matrix{
		{1, 2},
		{3, 4},
		{5, 6},
	}
	m.Transpose()

	fmt.Println(m)

	expected := Matrix{
		{1, 3, 5},
		{2, 4, 6},
	}

	assert.Equal(t, expected, m)
}

func TestMatrix_Multiply(t *testing.T) {
	left := Matrix{
		{1, 3, 5},
		{2, 4, 6},
	}
	right := Matrix{
		{1, 2},
		{3, 4},
		{5, 6},
	}

	expected := Matrix{
		{35, 44},
		{44, 56},
	}

	got := left.MultiplyRight(right)

	assert.Equal(t, expected, got)
}
