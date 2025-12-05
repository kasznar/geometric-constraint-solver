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
	}

	panic("Can't format")
}

func (e *Expr) Eval() float64 {
	switch e.Type {
	case CONSTANT:
		return e.Value
	case PARAMETER:
		return parameters[e.Name]
	case ADD:
		return e.Left.Eval() + e.Right.Eval()
	case SUBTRACT:
		return e.Left.Eval() - e.Right.Eval()
	case MULTIPLY:
		return e.Left.Eval() * e.Right.Eval()
	case SQUARE:
		return e.Left.Eval() * e.Left.Eval()
	case NEGATE:
		return e.Left.Eval() * -1.0
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

func Number(value float64) *Expr {
	return &Expr{CONSTANT, nil, nil, value, ""}
}

func Param(name string) *Expr {
	return &Expr{PARAMETER, nil, nil, 0, name}
}
