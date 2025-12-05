package sketch

import (
	. "equation-solver/pkg/utils"
	"testing"
)

func TestSketch_Distance(t *testing.T) {
	s := NewSketch()

	s.AddOrigin("O1", 0, 0)
	s.AddOrigin("O2", 10, 0)

	s.AddPoint("A", 5, 3)

	s.SetDistance("O1", "A", 7)
	s.SetDistance("O2", "A", 7)

	s.SatisfyConstraints()

	println(s.parameters.Format())

	AssertAlmost(t, s.GetParam("Ax"), 5)
	AssertAlmost(t, s.GetParam("Ay"), 4.898979485566356)
}

func TestSketch_Distance2(t *testing.T) {
	s := NewSketch()

	s.AddOrigin("O1", 0, 0)
	s.AddOrigin("O2", 10, 0)

	s.AddPoint("A", 5, 3)

	s.SetDistance("O1", "A", 5)
	s.SetDistance("O2", "A", 11.18)

	s.SatisfyConstraints()

	AssertAlmost(t, s.GetParam("Ax"), 0)
	AssertAlmost(t, s.GetParam("Ay"), 5)
}

func TestSketch_Distance3(t *testing.T) {
	s := NewSketch()

	s.AddOrigin("O1", 100, 0)
	s.AddOrigin("O2", 0, 100)

	s.AddPoint("A", 200, 150)

	s.SetDistance("O1", "A", 100)
	s.SetDistance("O2", "A", 100)

	s.SatisfyConstraints()

	println(s.parameters.Format())
}