package main

import (
	"equation-solver/pkg/sketch"
	"fmt"
	"strconv"
)

func main() {
	LaunchUI(func(distances []*Distance, points []*Point, update func([]*Point)) {
		fmt.Printf("Distances: %+v\n", distances)
		fmt.Printf("Points: %+v\n", points)

		s := sketch.NewSketch()

		for i, p := range points {
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

		s.SatisfyConstraints()
		s.PrintParams()

		newPoints := make([]*Point, 0)
		newPoints = append(newPoints, points...)

		for i := 2; i < len(newPoints); i++ {
			uiP := newPoints[i]
			pName := strconv.Itoa(uiP.index)

			xParam := s.GetParam(pName + "x")
			yParam := s.GetParam(pName + "y")

			uiP.X = int(xParam)
			uiP.Y = int(yParam)
		}

		update(newPoints)
	})
}
