package sketch

import (
	"math"
	"testing"

	. "equation-solver/pkg/utils"
)

func TestSketch3D_Distance(t *testing.T) {
	s := NewSketch3D()

	s.AddOrigin3D("O1", 0, 0, 0)
	s.AddOrigin3D("O2", 4, 0, 0)
	s.AddOrigin3D("O3", 0, 4, 0)

	s.AddPoint3D("A", 1, 1, 1)

	d := 2 * math.Sqrt(3)
	s.SetDistance3D("O1", "A", d)
	s.SetDistance3D("O2", "A", d)
	s.SetDistance3D("O3", "A", d)

	before := captureSketch3D(s)
	s.SatisfyConstraints()
	writeSnapshot3D(s, before, t.Name(), "testdata/"+t.Name()+".json")

	AssertAlmost(t, s.GetParam("Ax"), 2)
	AssertAlmost(t, s.GetParam("Ay"), 2)
	AssertAlmost(t, s.GetParam("Az"), 2)
}

func TestSketch3D_PointOnPlane(t *testing.T) {
	s := NewSketch3D()

	s.AddOrigin3D("O1", 0, 5, 5)
	s.AddOrigin3D("O2", 5, 0, 5)

	s.AddPoint3D("A", 3, 3, 3)

	s.SetPointOnPlane("A", 0, 0, 1, 0, 0, 5)
	s.SetDistance3D("O1", "A", 5)
	s.SetDistance3D("O2", "A", 5)

	before := captureSketch3D(s)
	s.SatisfyConstraints()
	writeSnapshot3D(s, before, t.Name(), "testdata/"+t.Name()+".json",
		snapshotPlane{NX: 0, NY: 0, NZ: 1, PX: 0, PY: 0, PZ: 5})

	AssertAlmost(t, s.GetParam("Ax"), 5)
	AssertAlmost(t, s.GetParam("Ay"), 5)
	AssertAlmost(t, s.GetParam("Az"), 5)
}
