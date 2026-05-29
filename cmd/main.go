package main

import (
	"equation-solver/pkg/sketch"
	"equation-solver/pkg/solver"
	"fmt"
	"strconv"
)

func main() {
	LaunchUI(func(
		distances []*Distance,
		uiPoints []*Point,
		inspEnabled bool,
		update func([]*Point),
		setInspector func(*solver.Inspector),
	) {
		fmt.Printf("Distances: %+v\n", distances)
		fmt.Printf("Points: %+v\n", uiPoints)

		s := sketch.NewSketch()

		for i, p := range uiPoints {
			name := strconv.Itoa(p.index)
			if i < 2 {
				s.AddOrigin(name, float64(p.X), float64(p.Y))
			} else {
				s.AddPoint(name, float64(p.X), float64(p.Y))
			}
		}

		for _, d := range distances {
			A := strconv.Itoa(d.P1.index)
			B := strconv.Itoa(d.P2.index)
			s.SetDistance(A, B, d.Value)
		}

		// applyFinal reads solved param values back into the UI point slice.
		applyFinal := func() {
			newPoints := make([]*Point, len(uiPoints))
			copy(newPoints, uiPoints)
			for i := 2; i < len(newPoints); i++ {
				p := newPoints[i]
				name := strconv.Itoa(p.index)
				p.X = int(s.GetParam(name + "x"))
				p.Y = int(s.GetParam(name + "y"))
			}
			update(newPoints)
		}

		if !inspEnabled {
			s.SatisfyConstraints()
			s.PrintParams()
			applyFinal()
			return
		}

		ins := solver.NewInspector(true)
		ins.OnCheckpoint = applyFinal

		setInspector(ins)

		go func() {
			s.SatisfyConstraintsWithInspector(ins)
			s.PrintParams()
			applyFinal()
		}()
	})
}
