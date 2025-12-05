package solver

import (
	. "equation-solver/pkg/math"
	. "equation-solver/pkg/utils"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSolver_Solve(t *testing.T) {
	m := Matrix{
		{3, 2, -4},
		{2, 3, 3},
		{5, -3, 1},
	}
	c := Vector{3, 15, 14}
	expected := Vector{2.9999999999999996, 0.9999999999999996, 2}

	state, result := Solve(EquationSystem{m, c})

	assert.Equal(t, CONVERGED, state)
	assert.Equal(t, expected, result)
}

func TestSolver_Solve_Overconstrained(t *testing.T) {
	m := Matrix{
		{1, 1},
		{1, 2},
		{1, 3},
	}
	c := Vector{1, 3, 2}
	expected := Vector{1, 0.5}

	state, result := Solve(EquationSystem{m, c})

	assert.Equal(t, CONVERGED, state)
	assert.Equal(t, expected, result)
}

func TestSolver_CreateJacobian(t *testing.T) {
	p := &SystemParameters{}

	p.add(SParam{"A", 1})
	p.add(SParam{"B", 2})

	p.save()

	defer func() { parameters = map[string]float64{} }()

	es := []*Expr{
		Param("A"),
		Param("B"),
	}

	J := createJacobian(es)

	expected := [][]*Expr{
		{Number(1), Number(0)},
		{Number(0), Number(1)},
	}

	assert.Equal(t, expected, J)
}

func TestSolver_CreateJacobian2(t *testing.T) {
	p := &SystemParameters{}

	p.add(SParam{"A", 1})
	p.add(SParam{"B", 2})

	p.save()

	defer func() { parameters = map[string]float64{} }()

	es := []*Expr{
		Param("A").Square(),
		Param("B").Square(),
	}

	J := createJacobian(es)

	print(J[0][0].Format())
	print(" ")
	println(J[0][1].Format())
	print(J[1][0].Format())
	print(" ")
	println(J[1][1].Format())

	expected := [][]*Expr{
		{
			Number(2).Multiply(Param("A")).Multiply(Number(1)),
			Number(2).Multiply(Param("A")).Multiply(Number(0)),
		},
		{
			Number(2).Multiply(Param("B")).Multiply(Number(0)),
			Number(2).Multiply(Param("B")).Multiply(Number(1)),
		},
	}

	assert.Equal(t, expected, J)
}

func TestSolver_EvalJacobian(t *testing.T) {
	parameters = map[string]float64{"A": 1, "B": 2}
	defer func() { parameters = map[string]float64{} }()

	es := []*Expr{
		Param("A"),
		Param("B"),
	}

	J := createJacobian(es)
	result := evalJacobian(J)

	expected := Matrix{
		{1, 0},
		{0, 1},
	}

	assert.Equal(t, expected, result)
}

func TestSolver_EvalJacobian2(t *testing.T) {
	parameters = map[string]float64{"A": 1, "B": 2}
	defer func() { parameters = map[string]float64{} }()

	es := []*Expr{
		Param("A").Square(),
		Param("B").Square(),
	}

	for _, v := range es {
		println(v.Format())
	}

	J := createJacobian(es)

	for _, r := range J {
		for _, e := range r {
			println(e.Format())
		}
	}

	result := evalJacobian(J)

	/*
		2*A*1 2*A*0
		2*B*0 2*B*1
	*/
	expected := Matrix{
		{2, 0},
		{0, 4},
	}

	assert.Equal(t, expected, result)
}

func TestSolver_EvalSystem(t *testing.T) {
	parameters = map[string]float64{"A": 1, "B": 2}
	defer func() { parameters = map[string]float64{} }()

	es := []*Expr{
		Param("A").Square(),
		Param("B").Square(),
	}
	got := evalSystem(es)

	expected := Vector{1, 4}

	assert.Equal(t, expected, got)
}

func TestSolver_Params(t *testing.T) {
	defer func() { parameters = map[string]float64{} }()

	p := &SystemParameters{}

	p.add(SParam{"A", 1})

	assert.Equal(t, parameters["A"], 1.0)
}

func TestSolver_ParamsGetVec(t *testing.T) {
	defer func() { parameters = map[string]float64{} }()

	p := &SystemParameters{}

	p.add(SParam{"A", 1})
	p.add(SParam{"B", 2})
	p.add(SParam{"C", 3})

	got := p.getVec()

	expected := Vector{1, 2, 3}

	assert.Equal(t, expected, got)
}

func TestSolver_SaveVec(t *testing.T) {
	defer func() { parameters = map[string]float64{} }()

	p := &SystemParameters{}

	p.add(SParam{"A", 1})
	p.add(SParam{"B", 2})
	p.add(SParam{"C", 3})

	vec := Vector{4, 5, 6}

	p.saveVec(vec)

	assert.Equal(t, 4.0, parameters["A"])
	assert.Equal(t, 5.0, parameters["B"])
	assert.Equal(t, 6.0, parameters["C"])

	assert.Equal(t, 4.0, p.list[0].value)
	assert.Equal(t, 5.0, p.list[1].value)
	assert.Equal(t, 6.0, p.list[2].value)
}

// this order is not guaranteed
func TestSolver_ParamIteration(t *testing.T) {
	parameters = map[string]float64{"A": 1, "B": 2, "C": 3}
	defer func() { parameters = map[string]float64{} }()

	var result []string

	for p := range parameters {
		result = append(result, p)
	}

	assert.Equal(t, []string{"A", "B", "C"}, result)
}

func TestSolver_SolveSystem(t *testing.T) {
	defer func() {
		parameters = map[string]float64{}
	}()

	p := &SystemParameters{}

	p.add(SParam{"x", 1})
	p.add(SParam{"y", 1})

	sys := []*Expr{
		Param("x").Square().Add(Param("y")),
		Param("y").Square().Add(Param("x")).Subtract(Number(1)),
	}

	SolveSystem(sys, p)

	fmt.Printf("%#v\n", parameters)

	assert.Equal(t, AlmostEqual(p.list[0].value, 0.7244919590005157, 1e-9), true)
	assert.Equal(t, AlmostEqual(p.list[1].value, -0.5248885986564048, 1e-9), true)
}
