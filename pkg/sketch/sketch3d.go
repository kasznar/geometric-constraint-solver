package sketch

import . "equation-solver/pkg/solver"

type Point3D struct {
	Name    string
	X, Y, Z *Expr
}

func NewPoint3D(name string) *Point3D {
	return &Point3D{name, Param(name + "x"), Param(name + "y"), Param(name + "z")}
}

func NewOrigin3D(name string, x, y, z float64) *Point3D {
	return &Point3D{name, Number(x), Number(y), Number(z)}
}

type Sketch3D struct {
	system      []*Expr
	constraints []Constraint
	points      map[string]*Point3D
	parameters  *SystemParameters
}

func NewSketch3D() *Sketch3D {
	return &Sketch3D{
		system:     []*Expr{},
		points:     map[string]*Point3D{},
		parameters: &SystemParameters{},
	}
}

func (s *Sketch3D) AddPoint3D(name string, x, y, z float64) {
	p := NewPoint3D(name)
	s.parameters.Add(p.X.Name, x)
	s.parameters.Add(p.Y.Name, y)
	s.parameters.Add(p.Z.Name, z)
	s.points[name] = p
}

func (s *Sketch3D) AddOrigin3D(name string, x, y, z float64) {
	s.points[name] = NewOrigin3D(name, x, y, z)
}

func (s *Sketch3D) SetDistance3D(A, B string, d float64) {
	a, b := s.points[A], s.points[B]
	e := a.X.Subtract(b.X).Square().
		Add(a.Y.Subtract(b.Y).Square()).
		Add(a.Z.Subtract(b.Z).Square()).
		Subtract(Number(d).Square())
	s.system = append(s.system, e)
	s.constraints = append(s.constraints, Constraint{"distance", A, B, d})
}

func (s *Sketch3D) SetPointOnPlane(P string, nx, ny, nz, p0x, p0y, p0z float64) {
	p := s.points[P]
	e := Number(nx).Multiply(p.X.Subtract(Number(p0x))).
		Add(Number(ny).Multiply(p.Y.Subtract(Number(p0y)))).
		Add(Number(nz).Multiply(p.Z.Subtract(Number(p0z))))
	s.system = append(s.system, e)
}

func (s *Sketch3D) SatisfyConstraints() SolveResult {
	return SolveSystem(s.system, s.parameters, nil)
}

func (s *Sketch3D) GetParam(name string) float64 {
	return s.parameters.Get(name)
}
