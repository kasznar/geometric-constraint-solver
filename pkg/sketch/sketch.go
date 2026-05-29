package sketch

import (
	"math"

	. "equation-solver/pkg/solver"
)

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

type Constraint struct {
	Kind  string
	A, B  string
	Value float64
}

type Sketch struct {
	system      []*Expr
	constraints []Constraint
	points      map[string]*Point
	lines       map[string]*Line
	parameters  *SystemParameters
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
	s.constraints = append(s.constraints, Constraint{"distance", A, B, d})
}

func (s *Sketch) SetParallel(lineA, lineB string) {
	la := s.lines[lineA]
	lb := s.lines[lineB]
	a, b := s.points[la.A], s.points[la.B]
	c, d := s.points[lb.A], s.points[lb.B]

	e := b.X.Subtract(a.X).Multiply(d.Y.Subtract(c.Y)).
		Subtract(b.Y.Subtract(a.Y).Multiply(d.X.Subtract(c.X)))
	s.system = append(s.system, e)
	s.constraints = append(s.constraints, Constraint{"parallel", lineA, lineB, 0})
}

func (s *Sketch) SetPerpendicular(lineA, lineB string) {
	la := s.lines[lineA]
	lb := s.lines[lineB]
	a, b := s.points[la.A], s.points[la.B]
	c, d := s.points[lb.A], s.points[lb.B]

	e := b.X.Subtract(a.X).Multiply(d.X.Subtract(c.X)).
		Add(b.Y.Subtract(a.Y).Multiply(d.Y.Subtract(c.Y)))
	s.system = append(s.system, e)
	s.constraints = append(s.constraints, Constraint{"perpendicular", lineA, lineB, 0})
}

func (s *Sketch) SetAngle(lineA, lineB string, angle float64) {
	la := s.lines[lineA]
	lb := s.lines[lineB]
	a, b := s.points[la.A], s.points[la.B]
	c, d := s.points[lb.A], s.points[lb.B]

	abx := b.X.Subtract(a.X)
	aby := b.Y.Subtract(a.Y)
	cdx := d.X.Subtract(c.X)
	cdy := d.Y.Subtract(c.Y)

	dot := abx.Multiply(cdx).Add(aby.Multiply(cdy))
	abLen := abx.Square().Add(aby.Square()).Sqrt()
	cdLen := cdx.Square().Add(cdy.Square()).Sqrt()

	e := dot.Divide(abLen.Multiply(cdLen)).Subtract(Number(math.Cos(angle)))
	s.system = append(s.system, e)
	s.constraints = append(s.constraints, Constraint{"angle", lineA, lineB, angle})
}

func (s *Sketch) SatisfyConstraints() SolveResult {
	return SolveSystem(s.system, s.parameters, nil)
}

func (s *Sketch) SatisfyConstraintsWithInspector(ins *Inspector) {
	SolveSystem(s.system, s.parameters, ins)
}

func (s *Sketch) PrintParams() {
	println(s.parameters.Format())
}

func (s *Sketch) GetParam(name string) float64 {
	return s.parameters.Get(name)
}
