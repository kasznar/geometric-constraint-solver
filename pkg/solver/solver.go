package solver

import (
	. "equation-solver/pkg/math"
	"fmt"
	"math"
)

type Result string

const (
	CONVERGED    Result = "CONVERGED"
	UNDERDEFINED Result = "UNDERDEFINED"
	OVERDEFINED  Result = "OVERDEFINED"
)

type SystemParameters struct {
	list []*SParam
}

func (sp *SystemParameters) add(newParam SParam) {
	sp.list = append(sp.list, &newParam)
	parameters[newParam.name] = newParam.value
}

func (sp *SystemParameters) Add(name string, value float64) {
	p := &SParam{name, value}
	sp.list = append(sp.list, p)
	parameters[p.name] = p.value
}

func (sp *SystemParameters) Get(name string) float64 {
	for _, p := range sp.list {
		if p.name == name {
			return p.value
		}
	}

	panic("No param like that")
}

func (sp *SystemParameters) getVec() Vector {
	vector := make(Vector, len(sp.list))

	for i, p := range sp.list {
		vector[i] = p.value
	}

	return vector
}

func (sp *SystemParameters) save() {
	paramList = paramList[:0]

	for _, p := range sp.list {
		parameters[p.name] = p.value
		paramList = append(paramList, p.name)
	}

}

func (sp *SystemParameters) saveVec(vector Vector) {
	for i, value := range vector {
		current := sp.list[i]
		current.value = value
		parameters[current.name] = value
	}
}

func (sp *SystemParameters) clear() {
	parameters = map[string]float64{}
}

func (sp *SystemParameters) Format() string {
	result := ""
	for _, p := range sp.list {
		result += fmt.Sprintf("%s: %v\n", p.name, p.value)
	}
	return result
}

type SParam struct {
	name  string
	value float64
}

// Global variable containing the parameter values
var parameters = make(map[string]float64)
var paramList = make([]string, 0)

type EquationSystem struct {
	coefficients Matrix
	constants    Vector
}

// Solve attempts to solve the given matrix equation using Gaussian elimination.
//
// Parameters:
//   - matrix: The augmented matrix representing the system of equations.
//
// Returns:
//   - Result: The outcome of the solving process (e.g., CONVERGED, UNDERDEFINED, OVERDEFINED).
//   - Vector: The solution vector if a solution is found, otherwise nil.
func Solve(system EquationSystem) (Result, Vector) {
	coefficients := system.coefficients
	constants := NewMatrixFromColVec(system.constants)

	rows, cols := coefficients.Size()
	if rows > cols {
		//return OVERDEFINED, nil
		transpose := coefficients.Copy().Transpose()
		coefficients = transpose.MultiplyRight(coefficients)
		constants = transpose.MultiplyRight(constants)
	}

	if cols > rows {
		return UNDERDEFINED, nil
	}

	matrix := coefficients
	matrix.Augment(constants)

	rows = matrix.Rows()

	gaussEliminate(matrix, rows)

	solution := make(Vector, rows)
	backSubstitute(matrix, rows, solution)

	fmt.Println("\nRow echelon form\n", matrix)

	return CONVERGED, solution
}

func SolveGauss(coefficients Matrix, constants Vector) Vector {
	rows := coefficients.Rows()

	matrix := coefficients
	matrix.AugmentVec(constants)

	gaussEliminate(matrix, rows)

	solution := make(Vector, rows)
	backSubstitute(matrix, rows, solution)

	fmt.Println("\nRow echelon form\n", matrix)

	return solution
}

func SolveSystem(equationSystem []*Expr, params *SystemParameters) {
	params.save()
	defer func() { params.clear() }()

	for i := 0; i < 100; i++ {
		J := createJacobian(equationSystem)
		J_x := evalJacobian(J)
		F_x := evalSystem(equationSystem)

		d := SolveGauss(J_x, F_x)

		converged := true
		for _, v := range d {
			if math.Abs(v) > 1e-6 {
				converged = false
			}
		}

		if converged {
			fmt.Printf("Converged after %d iterations\n", i+1)
			break
		}

		x := params.getVec()
		next := x.Subtract(d)
		params.saveVec(next)
	}
}

// performs gaussian elimination with partial pivot
func gaussEliminate(A Matrix, n int) {
	for i := 0; i < n; i++ {
		pivotRow := i

		for j := i + 1; j < n; j++ {
			if math.Abs(A[j][i]) > math.Abs(A[pivotRow][i]) {
				pivotRow = j
			}
		}

		if pivotRow != i {
			A.SwapRows(pivotRow, i)
		}

		for j := i + 1; j < n; j++ {
			factor := A[j][i] / A[i][i]
			A[j] = A[j].Subtract(A[i].Multiply(factor))
		}
	}
}

func backSubstitute(A Matrix, n int, x Vector) {
	for i := n - 1; i >= 0; i-- {
		sum := 0.0
		for j := i + 1; j < n; j++ {
			sum += A[i][j] * x[j]
		}
		x[i] = (A[i][n] - sum) / A[i][i]
	}
}

func createJacobian(equations []*Expr) [][]*Expr {
	rows := len(equations)
	cols := len(parameters)

	// Jacobian
	J := make([][]*Expr, rows)

	for i, e := range equations {
		J[i] = make([]*Expr, cols)
		j := 0
		for _, p := range paramList {
			J[i][j] = e.PartialDiff(p)
			j++
		}
	}

	return J
}

func evalJacobian(jacobian [][]*Expr) Matrix {
	rows := len(jacobian)
	cols := len(jacobian[0])

	m := NewMatrix(rows, cols)

	for i := 0; i < cols; i++ {
		for j := 0; j < rows; j++ {
			m[i][j] = jacobian[i][j].Eval()
		}
	}

	return m

}

func evalSystem(system []*Expr) Vector {
	result := make(Vector, len(system))

	for i, e := range system {
		result[i] = e.Eval()
	}

	return result
}
