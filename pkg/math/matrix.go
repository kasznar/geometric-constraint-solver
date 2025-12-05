package math

import "strconv"

type Matrix []Vector

func NewMatrixFromColVec(vec Vector) Matrix {
	m := Matrix{}

	for i := 0; i < len(vec); i++ {
		m = append(m, Vector{vec[i]})
	}

	return m
}

func NewMatrix(rows int, cols int) Matrix {
	m := Matrix{}

	for i := 0; i < rows; i++ {
		m = append(m, make(Vector, cols))
	}

	return m
}

func (m Matrix) AugmentVec(vec Vector) {
	if len(m) != len(vec) {
		panic("dimensions don't match, can't augment matrix")
	}

	for i := 0; i < len(m); i++ {
		m[i] = append(m[i], vec[i])
	}
}

func (m Matrix) Augment(m2 Matrix) {
	if len(m) != len(m2) {
		panic("dimensions don't match, can't augment matrix")
	}

	for i := 0; i < len(m); i++ {
		m[i] = append(m[i], m2[i]...)
	}
}

func (m Matrix) Rows() int {
	return len(m)
}

func (m Matrix) Cols() int {
	return len(m[0])
}

func (m Matrix) Size() (int, int) {
	return len(m), len(m[0])
}

func (m Matrix) IsSquare() bool {
	rows, cols := m.Size()

	return rows == cols
}

func (m Matrix) SwapRows(row1 int, row2 int) {
	temp := m[row1]
	m[row1] = m[row2]
	m[row2] = temp
}

func (m Matrix) String() string {
	result := ""
	rows, cols := m.Size()
	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			result += strconv.FormatFloat(m[i][j], 'f', -1, 64)
			if j != cols-1 {
				result += " "
			}
		}

		if i != rows-1 {
			result += "\n"
		}

	}

	return result
}

func (m *Matrix) Copy() *Matrix {
	result := Matrix{}
	rows, cols := m.Size()

	for i := 0; i < rows; i++ {
		result = append(result, Vector{})

		for j := 0; j < cols; j++ {
			result[i] = append(result[i], (*m)[i][j])
		}
	}

	return &result
}

func (m *Matrix) Transpose() *Matrix {
	result := NewMatrix(m.Cols(), m.Rows())

	rows, cols := m.Size()

	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			result[j][i] = (*m)[i][j]
		}
	}

	*m = result

	return &result
}

func (left Matrix) MultiplyRight(right Matrix) Matrix {
	if left.Cols() != right.Rows() {
		panic("can't multiply, dimensions don't match")
	}

	result := NewMatrix(left.Rows(), right.Cols())

	for i := 0; i < left.Rows(); i++ {
		for j := 0; j < right.Cols(); j++ {
			// compute
			sum := 0.0
			for k := 0; k < left.Cols(); k++ {
				sum += left[i][k] * right[k][j]
			}

			result[i][j] = sum
		}
	}

	return result
}
