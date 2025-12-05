package sketch

import . "equation-solver/pkg/solver"

type Point struct {
	Name string
	X    *Expr
	Y    *Expr
}

func NewPoint(name string) *Point {

	x := Param(name + "x")
	y := Param(name + "y")

	p := &Point{name, x, y}

	return p
}

func NewOrigin(name string, x float64, y float64) *Point {
	xExpr := Number(x)
	yExpr := Number(y)

	p := &Point{name, xExpr, yExpr}

	return p
}

type Line struct {
	name string
	A    string
	B    string
}

type Sketch struct {
	system     []*Expr
	points     map[string]*Point
	lines      map[string]*Line
	parameters *SystemParameters
}

func NewSketch() *Sketch {

	return &Sketch{
		system:     []*Expr{},
		points:     map[string]*Point{},
		lines:      map[string]*Line{},
		parameters: &SystemParameters{},
	}
}

func (s *Sketch) AddPoint(name string, x float64, y float64) {
	p := NewPoint(name)

	s.parameters.Add(p.X.Name, x)
	s.parameters.Add(p.Y.Name, y)

	s.points[name] = p
}

func (s *Sketch) AddOrigin(name string, x float64, y float64) {
	p := NewOrigin(name, x, y)

	s.points[name] = p
}

func (s *Sketch) AddLine(name string, A string, B string) {
	s.lines[name] = &Line{name, A, B}
}

func (s *Sketch) SetDistance(A string, B string, d float64) {
	a := s.points[A]
	b := s.points[B]

	e := a.X.Subtract(b.X).Square().
		Add(a.Y.Subtract(b.Y).Square()).
		Subtract(Number(d).Square())

	s.system = append(s.system, e)
}

func (s *Sketch) SetAngle(A Line, B Line, angle float64) {
	panic("todo")
}

func (s *Sketch) SatisfyConstraints() {
	SolveSystem(s.system, s.parameters)
}

func (s *Sketch) PrintParams() {
	println(s.parameters.Format())
}

func (s *Sketch) GetParam(name string) float64 {
	return s.parameters.Get(name)
}
