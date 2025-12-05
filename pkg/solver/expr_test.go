package solver

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestDerivation_ConstantAdd(t *testing.T) {
	input := Number(1).Add(Number(2))
	expected := Number(0).Add(Number(0))

	result := input.PartialDiff("A")

	assert.Equal(t, expected, result)
}

func TestDerivation_Multiply(t *testing.T) {
	input := Number(2).Multiply(Param("A"))
	expected := Number(0).Multiply(Param("A")).Add(Number(2).Multiply(Number(1)))

	result := input.PartialDiff("A")

	assert.Equal(t, expected, result)
}

func TestDerivation_Exponent(t *testing.T) {
	input := Param("A").Square()

	result := input.PartialDiff("A")

	assert.Equal(t, "2*A*1", result.Format())
}

func TestDerivation_Combined(t *testing.T) {
	input := Number(16).Multiply(Param("B")).Add(Param("B").Square())

	result := input.PartialDiff("B")

	assert.Equal(t, "0*B+16*1+2*B*1", result.Format())
}

func TestExpr_PartialDiff(t *testing.T) {
	input := Number(2).Multiply(Param("A")).Add(Number(3).Multiply(Param("B")))
	expected := "0*A+2*0+0*B+3*1"

	result := input.PartialDiff("B").Format()

	assert.Equal(t, expected, result)
}

func TestExpr_Format(t *testing.T) {
	input := Number(16).Multiply(Param("B")).Add(Param("B").Square())
	expected := "16*B+B^2"

	result := input.Format()

	assert.Equal(t, expected, result)
}

func TestExpr_Eval(t *testing.T) {
	parameters = map[string]float64{"X": 5, "Y": 3}
	defer func() { parameters = map[string]float64{} }()

	assert := assert.New(t)

	// CONSTANT
	assert.Equal(3.14, Number(3.14).Eval())

	// PARAMETER
	assert.Equal(5.0, Param("X").Eval())

	// ADD
	add := Number(1).Add(Number(2))
	assert.Equal(3.0, add.Eval())

	// SUBTRACT
	sub := Number(5).Subtract(Number(3))
	assert.Equal(2.0, sub.Eval())

	// MULTIPLY
	mul := Number(2).Multiply(Param("Y"))
	assert.Equal(6.0, mul.Eval())

	// SQUARE
	sq := Param("Y").Square()
	assert.Equal(9.0, sq.Eval())

	// NEGATE
	neg := Number(7).Negate()
	assert.Equal(-7.0, neg.Eval())
}

func TestExpr_EvalSquare(t *testing.T) {
	parameters = map[string]float64{"X": 5, "Y": 3}
	defer func() { parameters = map[string]float64{} }()

	e := Param("X").Subtract(Param("Y")).Square()
	assert.Equal(t, 4.0, e.Eval())
}

func TestExpr_DerivSquare(t *testing.T) {
	parameters = map[string]float64{"X": 5, "Y": 3}
	defer func() { parameters = map[string]float64{} }()

	e := Param("X").Subtract(Param("Y")).Square()
	assert.Equal(t, -4.0, e.PartialDiff("Y").Eval())
}
