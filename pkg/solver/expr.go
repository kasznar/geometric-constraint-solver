package solver

import (
	"fmt"
	"math"
)

type ExprType string

const (
	CONSTANT  ExprType = "CONSTANT"
	PARAMETER ExprType = "PARAMETER"
	ADD       ExprType = "ADD"
	SUBTRACT  ExprType = "SUBTRACT"
	MULTIPLY  ExprType = "MULTIPLY"
	SQUARE    ExprType = "SQUARE"
	NEGATE    ExprType = "NEGATE"
	DIVIDE    ExprType = "DIVIDE"
	SQRT      ExprType = "SQRT"
)

type Expr struct {
	Type  ExprType
	Left  *Expr
	Right *Expr
	Value float64
	Name  string
}

func (e *Expr) PartialDiff(by string) *Expr {

	switch e.Type {
	case CONSTANT:
		return Number(0)
	case PARAMETER:
		if e.Name == by {
			return Number(1)
		} else {
			return Number(0)
		}
	case ADD:
		left := e.Left
		right := e.Right
		dLeft := left.PartialDiff(by)
		dRight := right.PartialDiff(by)

		return dLeft.Add(dRight)
	case SUBTRACT:
		left := e.Left
		right := e.Right
		dLeft := left.PartialDiff(by)
		dRight := right.PartialDiff(by)

		return dLeft.Subtract(dRight)
	case MULTIPLY:
		left := e.Left
		right := e.Right
		dLeft := left.PartialDiff(by)
		dRight := right.PartialDiff(by)

		return dLeft.Multiply(right).Add(left.Multiply(dRight))
	case SQUARE:
		return Number(2).Multiply(e.Left).Multiply(e.Left.PartialDiff(by))
	case DIVIDE:
		f, g := e.Left, e.Right
		df, dg := f.PartialDiff(by), g.PartialDiff(by)
		return df.Multiply(g).Subtract(f.Multiply(dg)).Divide(g.Square())
	case SQRT:
		df := e.Left.PartialDiff(by)
		return df.Divide(Number(2).Multiply(e))
	}

	panic("Can't differentiate")
}

func (e *Expr) Format() string {
	switch e.Type {
	case CONSTANT:
		v := e.Value
		if math.Mod(v, 1) == 0 {
			return fmt.Sprintf("%.0f", v)
		}
		return fmt.Sprintf("%.2f", v)
	case PARAMETER:
		return e.Name
	case ADD:
		return e.Left.Format() + "+" + e.Right.Format()
	case SUBTRACT:
		return e.Left.Format() + "-" + e.Right.Format()
	case MULTIPLY:
		return e.Left.Format() + "*" + e.Right.Format()
	case SQUARE:
		return e.Left.Format() + "^2"
	case NEGATE:
		return "-" + e.Left.Format()
	case DIVIDE:
		return "(" + e.Left.Format() + ")/(" + e.Right.Format() + ")"
	case SQRT:
		return "sqrt(" + e.Left.Format() + ")"
	}

	panic("Can't format")
}

func (e *Expr) Eval(params *SystemParameters) float64 {
	switch e.Type {
	case CONSTANT:
		return e.Value
	case PARAMETER:
		return params.Get(e.Name)
	case ADD:
		return e.Left.Eval(params) + e.Right.Eval(params)
	case SUBTRACT:
		return e.Left.Eval(params) - e.Right.Eval(params)
	case MULTIPLY:
		return e.Left.Eval(params) * e.Right.Eval(params)
	case SQUARE:
		return e.Left.Eval(params) * e.Left.Eval(params)
	case NEGATE:
		return e.Left.Eval(params) * -1.0
	case DIVIDE:
		return e.Left.Eval(params) / e.Right.Eval(params)
	case SQRT:
		return math.Sqrt(e.Left.Eval(params))
	}

	panic("Can't eval")
}

func (e *Expr) Add(right *Expr) *Expr {
	return &Expr{ADD, e, right, 0, ""}
}

func (e *Expr) Subtract(right *Expr) *Expr {
	return &Expr{SUBTRACT, e, right, 0, ""}
}

func (e *Expr) Multiply(right *Expr) *Expr {
	return &Expr{MULTIPLY, e, right, 0, ""}
}

func (e *Expr) Square() *Expr {
	return &Expr{SQUARE, e, nil, 0, ""}
}

func (e *Expr) Negate() *Expr {
	return &Expr{NEGATE, e, nil, 0, ""}
}

func (e *Expr) Divide(right *Expr) *Expr {
	return &Expr{DIVIDE, e, right, 0, ""}
}

func (e *Expr) Sqrt() *Expr {
	return &Expr{SQRT, e, nil, 0, ""}
}

func Number(value float64) *Expr {
	return &Expr{CONSTANT, nil, nil, value, ""}
}

func Param(name string) *Expr {
	return &Expr{PARAMETER, nil, nil, 0, name}
}
