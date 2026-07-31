package math

import (
	"fmt"
	"strconv"
	"strings"
)

// Matrix is a port of pocketmine\math\Matrix.
//
// PHP's ArrayAccess ($matrix[$r]) has no Go equivalent for a struct (only maps/slices support
// []); use Row/SetElement/GetElement instead.
type Matrix struct {
	rows, columns int
	data          [][]float64
}

// NewMatrix constructs a rows x columns matrix, optionally seeded from set (missing entries
// default to 0, matching the PHP original).
func NewMatrix(rows, columns int, set [][]float64) *Matrix {
	if rows < 1 {
		rows = 1
	}
	if columns < 1 {
		columns = 1
	}
	m := &Matrix{rows: rows, columns: columns}
	m.Set(set)
	return m
}

func (m *Matrix) Set(set [][]float64) {
	m.data = make([][]float64, m.rows)
	for r := 0; r < m.rows; r++ {
		m.data[r] = make([]float64, m.columns)
		if r < len(set) {
			for c := 0; c < m.columns && c < len(set[r]); c++ {
				m.data[r][c] = set[r][c]
			}
		}
	}
}

func (m *Matrix) GetRows() int    { return m.rows }
func (m *Matrix) GetColumns() int { return m.columns }

// Row returns the underlying slice for row r, mirroring PHP's $matrix[$r] access.
func (m *Matrix) Row(r int) []float64 { return m.data[r] }

func (m *Matrix) SetElement(row, column int, value float64) error {
	if row >= m.rows || row < 0 || column >= m.columns || column < 0 {
		return fmt.Errorf("row or column out of bounds (have %d rows %d columns)", m.rows, m.columns)
	}
	m.data[row][column] = value
	return nil
}

func (m *Matrix) GetElement(row, column int) (float64, error) {
	if row >= m.rows || row < 0 || column >= m.columns || column < 0 {
		return 0, fmt.Errorf("row or column out of bounds (have %d rows %d columns)", m.rows, m.columns)
	}
	return m.data[row][column], nil
}

func (m *Matrix) IsSquare() bool { return m.rows == m.columns }

func (m *Matrix) Add(other *Matrix) (*Matrix, error) {
	if m.rows != other.rows || m.columns != other.columns {
		return nil, fmt.Errorf("matrix does not have the same number of rows and/or columns")
	}
	result := NewMatrix(m.rows, m.columns, nil)
	for r := 0; r < m.rows; r++ {
		for c := 0; c < m.columns; c++ {
			result.data[r][c] = m.data[r][c] + other.data[r][c]
		}
	}
	return result, nil
}

func (m *Matrix) Subtract(other *Matrix) (*Matrix, error) {
	if m.rows != other.rows || m.columns != other.columns {
		return nil, fmt.Errorf("matrix does not have the same number of rows and/or columns")
	}
	result := m.clone()
	for r := 0; r < m.rows; r++ {
		for c := 0; c < m.columns; c++ {
			result.data[r][c] = m.data[r][c] - other.data[r][c]
		}
	}
	return result, nil
}

func (m *Matrix) MultiplyScalar(n float64) *Matrix {
	result := m.clone()
	for r := 0; r < m.rows; r++ {
		for c := 0; c < m.columns; c++ {
			result.data[r][c] = m.data[r][c] * n
		}
	}
	return result
}

func (m *Matrix) DivideScalar(n float64) *Matrix {
	result := m.clone()
	for r := 0; r < m.rows; r++ {
		for c := 0; c < m.columns; c++ {
			result.data[r][c] = m.data[r][c] / n
		}
	}
	return result
}

func (m *Matrix) Transpose() *Matrix {
	result := NewMatrix(m.columns, m.rows, nil)
	for r := 0; r < m.rows; r++ {
		for c := 0; c < m.columns; c++ {
			result.data[c][r] = m.data[r][c]
		}
	}
	return result
}

// Product computes the naive matrix product, O(n^3).
func (m *Matrix) Product(other *Matrix) (*Matrix, error) {
	if m.columns != other.rows {
		return nil, fmt.Errorf("expected a matrix with %d rows", m.columns)
	}
	result := NewMatrix(m.rows, other.columns, nil)
	for i := 0; i < m.rows; i++ {
		for j := 0; j < other.columns; j++ {
			sum := 0.0
			for k := 0; k < m.columns; k++ {
				sum += m.data[i][k] * other.data[k][j]
			}
			result.data[i][j] = sum
		}
	}
	return result, nil
}

// Determinant computes the determinant of a 1x1, 2x2 or 3x3 matrix.
func (m *Matrix) Determinant() (float64, error) {
	if !m.IsSquare() {
		return 0, fmt.Errorf("cannot calculate determinant of a non-square matrix")
	}
	d := m.data
	switch m.rows {
	case 1:
		return d[0][0], nil
	case 2:
		return d[0][0]*d[1][1] - d[0][1]*d[1][0], nil
	case 3:
		return d[0][0]*d[1][1]*d[2][2] +
			d[0][1]*d[1][2]*d[2][0] +
			d[0][2]*d[1][0]*d[2][1] -
			d[2][0]*d[1][1]*d[0][2] -
			d[2][1]*d[1][2]*d[0][0] -
			d[2][2]*d[1][0]*d[0][1], nil
	default:
		return 0, fmt.Errorf("not implemented for %dx%d matrices", m.rows, m.columns)
	}
}

func (m *Matrix) clone() *Matrix {
	result := &Matrix{rows: m.rows, columns: m.columns, data: make([][]float64, m.rows)}
	for r, row := range m.data {
		result.data[r] = append([]float64(nil), row...)
	}
	return result
}

func (m *Matrix) String() string {
	var rows []string
	for _, row := range m.data {
		cells := make([]string, len(row))
		for i, v := range row {
			cells[i] = strconv.FormatFloat(v, 'g', -1, 64)
		}
		rows = append(rows, strings.Join(cells, ","))
	}
	return fmt.Sprintf("Matrix(%dx%d;%s)", m.rows, m.columns, strings.Join(rows, ";"))
}
