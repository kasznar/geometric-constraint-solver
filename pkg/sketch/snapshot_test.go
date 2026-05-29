package sketch

import (
	"encoding/json"
	"os"
	"path/filepath"

	"equation-solver/pkg/solver"
)

type sketchSnap map[string][2]float64

func captureSketch(s *Sketch) sketchSnap {
	snap := make(sketchSnap)
	for name, p := range s.points {
		if p.X.Type == solver.CONSTANT {
			snap[name] = [2]float64{p.X.Value, p.Y.Value}
		} else {
			snap[name] = [2]float64{s.parameters.Get(p.X.Name), s.parameters.Get(p.Y.Name)}
		}
	}
	return snap
}

type sketchSnap3D map[string][3]float64

func captureSketch3D(s *Sketch3D) sketchSnap3D {
	snap := make(sketchSnap3D)
	for name, p := range s.points {
		if p.X.Type == solver.CONSTANT {
			snap[name] = [3]float64{p.X.Value, p.Y.Value, p.Z.Value}
		} else {
			snap[name] = [3]float64{
				s.parameters.Get(p.X.Name),
				s.parameters.Get(p.Y.Name),
				s.parameters.Get(p.Z.Name),
			}
		}
	}
	return snap
}

type snapshotPoint struct {
	X     float64 `json:"x"`
	Y     float64 `json:"y"`
	Z     float64 `json:"z,omitempty"`
	Fixed bool    `json:"fixed"`
}

type snapshotLine struct {
	Name string `json:"name"`
	A    string `json:"a"`
	B    string `json:"b"`
}

type snapshotPlane struct {
	NX float64 `json:"nx"`
	NY float64 `json:"ny"`
	NZ float64 `json:"nz"`
	PX float64 `json:"px"`
	PY float64 `json:"py"`
	PZ float64 `json:"pz"`
}

type snapConstraint struct {
	Kind  string  `json:"kind"`
	A     string  `json:"a"`
	B     string  `json:"b,omitempty"`
	Value float64 `json:"value,omitempty"`
}

type snapshotFile struct {
	Name        string                   `json:"name"`
	Dimension   string                   `json:"dimension"`
	Before      map[string]snapshotPoint `json:"before"`
	After       map[string]snapshotPoint `json:"after"`
	Lines       []snapshotLine           `json:"lines,omitempty"`
	Planes      []snapshotPlane          `json:"planes,omitempty"`
	Constraints []snapConstraint         `json:"constraints,omitempty"`
}

func writeSnapshot(s *Sketch, before sketchSnap, name, path string) {
	after := captureSketch(s)

	toPoint := func(snap sketchSnap, pname string) snapshotPoint {
		coords := snap[pname]
		pt := s.points[pname]
		return snapshotPoint{X: coords[0], Y: coords[1], Fixed: pt.X.Type == solver.CONSTANT}
	}

	sf := snapshotFile{
		Name:      name,
		Dimension: "2d",
		Before:    make(map[string]snapshotPoint),
		After:     make(map[string]snapshotPoint),
	}
	for pname := range s.points {
		sf.Before[pname] = toPoint(before, pname)
		sf.After[pname] = toPoint(after, pname)
	}
	for _, l := range s.lines {
		sf.Lines = append(sf.Lines, snapshotLine{Name: l.name, A: l.A, B: l.B})
	}
	for _, c := range s.constraints {
		sf.Constraints = append(sf.Constraints, snapConstraint{Kind: c.Kind, A: c.A, B: c.B, Value: c.Value})
	}

	_ = os.MkdirAll(filepath.Dir(path), 0o755)
	f, err := os.Create(path)
	if err != nil {
		return
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	_ = enc.Encode(sf)
}

func writeSnapshot3D(s *Sketch3D, before sketchSnap3D, name, path string, planes ...snapshotPlane) {
	after := captureSketch3D(s)

	toPoint := func(snap sketchSnap3D, pname string) snapshotPoint {
		coords := snap[pname]
		pt := s.points[pname]
		return snapshotPoint{X: coords[0], Y: coords[1], Z: coords[2], Fixed: pt.X.Type == solver.CONSTANT}
	}

	sf := snapshotFile{
		Name:      name,
		Dimension: "3d",
		Before:    make(map[string]snapshotPoint),
		After:     make(map[string]snapshotPoint),
		Planes:    planes,
	}
	for pname := range s.points {
		sf.Before[pname] = toPoint(before, pname)
		sf.After[pname] = toPoint(after, pname)
	}
	for _, c := range s.constraints {
		sf.Constraints = append(sf.Constraints, snapConstraint{Kind: c.Kind, A: c.A, B: c.B, Value: c.Value})
	}

	_ = os.MkdirAll(filepath.Dir(path), 0o755)
	f, err := os.Create(path)
	if err != nil {
		return
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	_ = enc.Encode(sf)
}
