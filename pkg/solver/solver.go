package solver

import (
	. "equation-solver/pkg/math"
	"fmt"
	"math"
	"time"
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
}

func (sp *SystemParameters) Add(name string, value float64) {
	p := &SParam{name, value}
	sp.list = append(sp.list, p)
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

func (sp *SystemParameters) saveVec(vector Vector) {
	for i, value := range vector {
		sp.list[i].value = value
	}
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


type EquationSystem struct {
	coefficients Matrix
	constants    Vector
}

func Solve(system EquationSystem) (Result, Vector) {
	coefficients := system.coefficients
	constants := NewMatrixFromColVec(system.constants)

	rows, cols := coefficients.Size()
	if rows > cols {
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

func SolveGauss(coefficients Matrix, constants Vector) (Vector, bool) {
	rows := coefficients.Rows()

	matrix := coefficients
	matrix.AugmentVec(constants)

	nearSingular := gaussEliminate(matrix, rows)

	solution := make(Vector, rows)
	backSubstitute(matrix, rows, solution)

	fmt.Println("\nRow echelon form\n", matrix)

	return solution, nearSingular
}

type SolveResult struct {
	Converged           bool
	Iterations          int
	FinalResidualNorm   float64
	FinalStepNorm       float64
	InitialResidualNorm float64
	Diverged            bool
	NearSingular        bool
	Overconstrained     bool
	Underconstrained    bool
	ResidualHistory     []float64
	StepHistory         []float64
	ConditionNumber     float64
	SolveTime           time.Duration
}

func SolveSystem(equationSystem []*Expr, params *SystemParameters, ins *Inspector) SolveResult {
	const maxIter = 100
	const tol = 1e-6

	result := SolveResult{}

	nEq, nVar := len(equationSystem), len(params.list)
	if nEq > nVar {
		result.Overconstrained = true
		fmt.Printf("Warning: overconstrained system (%d equations, %d unknowns).\n", nEq, nVar)
		fmt.Println("  No exact solution exists. Remove redundant or conflicting constraints until equations == unknowns.")
		return result
	}
	if nEq < nVar {
		result.Underconstrained = true
		fmt.Printf("Warning: underconstrained system (%d equations, %d unknowns).\n", nEq, nVar)
		fmt.Println("  Infinitely many solutions exist; the result depends entirely on the initial guess.")
		fmt.Println("  Add constraints to remove the remaining degrees of freedom.")
		return result
	}

	f0 := evalSystem(equationSystem, params)
	initialResid := 0.0
	for _, v := range f0 {
		initialResid += v * v
	}
	result.InitialResidualNorm = math.Sqrt(initialResid)
	prevResidNorm := result.InitialResidualNorm

	J_sym := createJacobian(equationSystem, params)
	start := time.Now()
	for i := 0; i < maxIter; i++ {
		J_x := evalJacobian(J_sym, params)
		F_x := evalSystem(equationSystem, params)
		if i == 0 {
			result.ConditionNumber = condNumberInf(*J_x.Copy())
		}

		checkpointInspector(ins, i, J_x, F_x, params)

		d, singular := SolveGauss(J_x, F_x)
		if singular {
			result.NearSingular = true
		}

		stepNorm := 0.0
		for _, v := range d {
			stepNorm += v * v
		}
		stepNorm = math.Sqrt(stepNorm)

		residNorm := 0.0
		for _, v := range F_x {
			residNorm += v * v
		}
		residNorm = math.Sqrt(residNorm)

		if residNorm > prevResidNorm*10 {
			result.Diverged = true
			fmt.Printf("Warning: residual is growing (‖F‖ %.2e → %.2e at iteration %d).\n", prevResidNorm, residNorm, i+1)
			fmt.Println("  The initial configuration may be too far from the solution.")
			fmt.Println("  Recommended: move the initial guess closer to the expected result,")
			fmt.Println("  or solve simpler sub-problems first to get a better starting point.")
		}
		prevResidNorm = residNorm

		result.ResidualHistory = append(result.ResidualHistory, residNorm)
		result.StepHistory = append(result.StepHistory, stepNorm)
		result.Iterations = i + 1
		result.FinalStepNorm = stepNorm
		result.FinalResidualNorm = residNorm

		if stepNorm < tol && residNorm < tol {
			result.Converged = true
			fmt.Printf("Converged after %d iterations (‖d‖=%.2e, ‖F‖=%.2e)\n", i+1, stepNorm, residNorm)
			break
		}

		x := params.getVec()
		next := x.Subtract(d)
		params.saveVec(next)
	}

	result.SolveTime = time.Since(start)

	if !result.Converged {
		fmt.Printf("Did not converge after %d iterations (‖d‖=%.2e, ‖F‖=%.2e)\n",
			result.Iterations, result.FinalStepNorm, result.FinalResidualNorm)
	}

	if ins != nil {
		close(ins.Done)
	}

	return result
}

func checkpointInspector(ins *Inspector, i int, J_x Matrix, F_x []float64, params *SystemParameters) {
	if ins == nil {
		return
	}
	rows, cols := J_x.Size()
	jacSlice := make([][]float64, rows)
	for r := 0; r < rows; r++ {
		jacSlice[r] = make([]float64, cols)
		copy(jacSlice[r], J_x[r])
	}
	residSlice := make([]float64, len(F_x))
	copy(residSlice, F_x)
	namesCopy := make([]string, len(params.list))
	for k, p := range params.list {
		namesCopy[k] = p.name
	}
	ins.Checkpoint(i, namesCopy, jacSlice, residSlice)
}

func gaussEliminate(A Matrix, n int) bool {
	const singularTol = 1e-10
	nearSingular := false

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

		if math.Abs(A[i][i]) < singularTol {
			nearSingular = true
			fmt.Printf("Warning: near-zero pivot (%.2e) at column %d — Jacobian is near-singular.\n", A[i][i], i)
			fmt.Println("  Possible causes: redundant constraint, conflicting constraints, or initial")
			fmt.Println("  point at a degenerate position (e.g., two coincident points).")
			fmt.Println("  Recommended: remove duplicate constraints, or shift the initial guess off the degenerate position.")
		}

		for j := i + 1; j < n; j++ {
			factor := A[j][i] / A[i][i]
			A[j] = A[j].Subtract(A[i].Multiply(factor))
		}
	}

	return nearSingular
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

func condNumberInf(A Matrix) float64 {
	n := A.Rows()
	normA := 0.0
	for i := 0; i < n; i++ {
		rowSum := 0.0
		for j := 0; j < n; j++ {
			rowSum += math.Abs(A[i][j])
		}
		if rowSum > normA {
			normA = rowSum
		}
	}
	Ainv := NewMatrix(n, n)
	for col := 0; col < n; col++ {
		e := make(Vector, n)
		e[col] = 1.0
		x, _ := SolveGauss(*A.Copy(), e)
		for row := 0; row < n; row++ {
			Ainv[row][col] = x[row]
		}
	}
	normAinv := 0.0
	for i := 0; i < n; i++ {
		rowSum := 0.0
		for j := 0; j < n; j++ {
			rowSum += math.Abs(Ainv[i][j])
		}
		if rowSum > normAinv {
			normAinv = rowSum
		}
	}
	return normA * normAinv
}

func createJacobian(equations []*Expr, params *SystemParameters) [][]*Expr {
	rows := len(equations)
	cols := len(params.list)

	J := make([][]*Expr, rows)

	for i, e := range equations {
		J[i] = make([]*Expr, cols)
		for j, p := range params.list {
			J[i][j] = e.PartialDiff(p.name)
		}
	}

	return J
}

func evalJacobian(jacobian [][]*Expr, params *SystemParameters) Matrix {
	rows := len(jacobian)
	cols := len(jacobian[0])

	m := NewMatrix(rows, cols)

	for i := 0; i < cols; i++ {
		for j := 0; j < rows; j++ {
			m[i][j] = jacobian[i][j].Eval(params)
		}
	}

	return m
}

func evalSystem(system []*Expr, params *SystemParameters) Vector {
	result := make(Vector, len(system))

	for i, e := range system {
		result[i] = e.Eval(params)
	}

	return result
}
