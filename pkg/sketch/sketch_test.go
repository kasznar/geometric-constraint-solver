package sketch

import (
	"fmt"
	"math"
	"testing"

	. "equation-solver/pkg/utils"
)

func TestSketch_Distance(t *testing.T) {
	s := NewSketch()

	s.AddOrigin("O1", 0, 0)
	s.AddOrigin("O2", 10, 0)

	s.AddPoint("A", 5, 3)

	s.SetDistance("O1", "A", 7)
	s.SetDistance("O2", "A", 7)

	before := captureSketch(s)
	s.SatisfyConstraints()
	writeSnapshot(s, before, t.Name(), "testdata/"+t.Name()+".json")

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

	before := captureSketch(s)
	s.SatisfyConstraints()
	writeSnapshot(s, before, t.Name(), "testdata/"+t.Name()+".json")

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

	before := captureSketch(s)
	s.SatisfyConstraints()
	writeSnapshot(s, before, t.Name(), "testdata/"+t.Name()+".json")

	println(s.parameters.Format())
}

func TestSketch_Parallel(t *testing.T) {
	s := NewSketch()

	s.AddOrigin("O1", 0, 0)
	s.AddOrigin("O2", 4, 0)
	s.AddLine("L1", "O1", "O2")

	s.AddOrigin("O3", 0, 3)
	s.AddPoint("D", 2, 1)
	s.AddLine("L2", "O3", "D")

	s.SetParallel("L1", "L2")
	s.SetDistance("O3", "D", 4)

	before := captureSketch(s)
	s.SatisfyConstraints()
	writeSnapshot(s, before, t.Name(), "testdata/"+t.Name()+".json")

	AssertAlmost(t, s.GetParam("Dx"), 4)
	AssertAlmost(t, s.GetParam("Dy"), 3)
}

func TestSketch_Perpendicular(t *testing.T) {
	s := NewSketch()

	s.AddOrigin("O1", 0, 0)
	s.AddOrigin("O2", 3, 0)
	s.AddLine("L1", "O1", "O2")

	s.AddOrigin("O3", 1, 0)
	s.AddPoint("D", 3, 2)
	s.AddLine("L2", "O3", "D")

	s.SetPerpendicular("L1", "L2")
	s.SetDistance("O3", "D", 4)

	before := captureSketch(s)
	s.SatisfyConstraints()
	writeSnapshot(s, before, t.Name(), "testdata/"+t.Name()+".json")

	AssertAlmost(t, s.GetParam("Dx"), 1)
	AssertAlmost(t, s.GetParam("Dy"), 4)
}

func TestSketch_Angle(t *testing.T) {
	s := NewSketch()

	s.AddOrigin("O1", 0, 0)
	s.AddOrigin("O2", 1, 0)
	s.AddLine("L1", "O1", "O2")

	s.AddOrigin("O3", 1, 1)
	s.AddPoint("D", 2, 2)
	s.AddLine("L2", "O3", "D")

	s.SetAngle("L1", "L2", math.Pi/4)
	s.SetDistance("O3", "D", 2)

	before := captureSketch(s)
	s.SatisfyConstraints()
	writeSnapshot(s, before, t.Name(), "testdata/"+t.Name()+".json")

	expected := 1 + math.Sqrt2
	AssertAlmost(t, s.GetParam("Dx"), expected)
	AssertAlmost(t, s.GetParam("Dy"), expected)
}

func TestSketch_DiagnoseFarFromSolution(t *testing.T) {
	s := NewSketch()
	s.AddOrigin("O1", 0, 0)
	s.AddOrigin("O2", 6, 0)
	s.AddPoint("A", 0.001, 0.001)
	s.SetDistance("O1", "A", 5)
	s.SetDistance("O2", "A", 5)

	result := s.SatisfyConstraints()
	t.Logf("far-from-solution: Diverged=%v Converged=%v Iterations=%d",
		result.Diverged, result.Converged, result.Iterations)
}

func TestSketch_RemediationBetterInitialGuess(t *testing.T) {
	s := NewSketch()
	s.AddOrigin("O1", 0, 0)
	s.AddOrigin("O2", 6, 0)
	s.AddPoint("A", 3, 1)
	s.SetDistance("O1", "A", 5)
	s.SetDistance("O2", "A", 5)

	result := s.SatisfyConstraints()
	if !result.Converged {
		t.Errorf("expected convergence with better initial guess, got %+v", result)
	}
}

func TestSketch_DiagnoseNearSingularRedundant(t *testing.T) {
	s := NewSketch()
	s.AddOrigin("O1", 0, 0)
	s.AddOrigin("O2", 10, 0)
	s.AddPoint("A", 5, 3)
	s.SetDistance("O1", "A", 7)
	s.SetDistance("O1", "A", 7)

	result := s.SatisfyConstraints()
	t.Logf("redundant constraint: NearSingular=%v Converged=%v",
		result.NearSingular, result.Converged)
}

func TestSketch_RemediationDistinctConstraints(t *testing.T) {
	s := NewSketch()
	s.AddOrigin("O1", 0, 0)
	s.AddOrigin("O2", 10, 0)
	s.AddPoint("A", 5, 3)
	s.SetDistance("O1", "A", 7)
	s.SetDistance("O2", "A", 7)

	result := s.SatisfyConstraints()
	if !result.Converged {
		t.Errorf("expected convergence with distinct constraints, got %+v", result)
	}
}

func TestSketch_DiagnoseNearSingularDegenerate(t *testing.T) {
	s := NewSketch()
	s.AddOrigin("O1", 0, 0)
	s.AddOrigin("O2", 0, 10)
	s.AddPoint("A", 0, 0)
	s.SetDistance("O1", "A", 5)
	s.SetDistance("O2", "A", 5)

	result := s.SatisfyConstraints()
	t.Logf("degenerate position: NearSingular=%v Converged=%v",
		result.NearSingular, result.Converged)
}

func TestSketch_RemediationShiftedInitialGuess(t *testing.T) {
	s := NewSketch()
	s.AddOrigin("O1", 0, 0)
	s.AddOrigin("O2", 0, 10)
	s.AddPoint("A", 1, 0)
	s.SetDistance("O1", "A", 5)
	s.SetDistance("O2", "A", 5)

	result := s.SatisfyConstraints()
	if !result.Converged {
		t.Errorf("expected convergence after shifting initial guess, got %+v", result)
	}
}

func TestSketch_DiagnoseOverconstrained(t *testing.T) {
	s := NewSketch()
	s.AddOrigin("O1", 0, 0)
	s.AddOrigin("O2", 10, 0)
	s.AddOrigin("O3", 5, 8)
	s.AddPoint("A", 5, 3)
	s.SetDistance("O1", "A", 7)
	s.SetDistance("O2", "A", 7)
	s.SetDistance("O3", "A", 5)

	result := s.SatisfyConstraints()
	t.Logf("overconstrained: Overconstrained=%v Converged=%v",
		result.Overconstrained, result.Converged)
}

func TestSketch_RemediationRemoveExtraConstraint(t *testing.T) {
	s := NewSketch()
	s.AddOrigin("O1", 0, 0)
	s.AddOrigin("O2", 10, 0)
	s.AddPoint("A", 5, 3)
	s.SetDistance("O1", "A", 7)
	s.SetDistance("O2", "A", 7)

	result := s.SatisfyConstraints()
	if !result.Converged {
		t.Errorf("expected convergence after removing extra constraint, got %+v", result)
	}
}

func TestSketch_DiagnoseUnderconstrained(t *testing.T) {
	s := NewSketch()
	s.AddOrigin("O1", 0, 0)
	s.AddPoint("A", 3, 4)
	s.SetDistance("O1", "A", 5)

	result := s.SatisfyConstraints()
	t.Logf("underconstrained: Underconstrained=%v Converged=%v",
		result.Underconstrained, result.Converged)
}

func TestSketch_RemediationAddMissingConstraint(t *testing.T) {
	s := NewSketch()
	s.AddOrigin("O1", 0, 0)
	s.AddOrigin("O2", 10, 0)
	s.AddPoint("A", 3, 4)
	s.SetDistance("O1", "A", 5)
	s.SetDistance("O2", "A", 7.07)

	result := s.SatisfyConstraints()
	if !result.Converged {
		t.Errorf("expected convergence after adding missing constraint, got %+v", result)
	}
}

func TestSketch_PrintMeasurements(t *testing.T) {
	type example struct {
		name  string
		build func() *Sketch
	}

	examples := []example{
		{
			"Ex1 symmetric circles",
			func() *Sketch {
				s := NewSketch()
				s.AddOrigin("O1", 0, 0)
				s.AddOrigin("O2", 10, 0)
				s.AddPoint("A", 5, 3)
				s.SetDistance("O1", "A", 7)
				s.SetDistance("O2", "A", 7)
				return s
			},
		},
		{
			"Ex2 asymmetric circles",
			func() *Sketch {
				s := NewSketch()
				s.AddOrigin("O1", 0, 0)
				s.AddOrigin("O2", 10, 0)
				s.AddPoint("A", 5, 3)
				s.SetDistance("O1", "A", 5)
				s.SetDistance("O2", "A", 11.18)
				return s
			},
		},
		{
			"Ex3 equilateral triangle",
			func() *Sketch {
				s := NewSketch()
				s.AddOrigin("O1", 0, 0)
				s.AddOrigin("O2", 5, 0)
				s.AddPoint("C", 2, 1)
				s.SetDistance("O1", "C", 5)
				s.SetDistance("O2", "C", 5)
				return s
			},
		},
		{
			"Ex4 parallelism",
			func() *Sketch {
				s := NewSketch()
				s.AddOrigin("O1", 0, 0)
				s.AddOrigin("O2", 4, 0)
				s.AddLine("L1", "O1", "O2")
				s.AddOrigin("O3", 0, 3)
				s.AddPoint("D", 2, 1)
				s.AddLine("L2", "O3", "D")
				s.SetParallel("L1", "L2")
				s.SetDistance("O3", "D", 4)
				return s
			},
		},
		{
			"Ex5 perpendicularity",
			func() *Sketch {
				s := NewSketch()
				s.AddOrigin("O1", 0, 0)
				s.AddOrigin("O2", 3, 0)
				s.AddLine("L1", "O1", "O2")
				s.AddOrigin("O3", 1, 0)
				s.AddPoint("D", 3, 2)
				s.AddLine("L2", "O3", "D")
				s.SetPerpendicular("L1", "L2")
				s.SetDistance("O3", "D", 4)
				return s
			},
		},
		{
			"Ex6 angle constraint",
			func() *Sketch {
				s := NewSketch()
				s.AddOrigin("O1", 0, 0)
				s.AddOrigin("O2", 1, 0)
				s.AddLine("L1", "O1", "O2")
				s.AddOrigin("O3", 1, 1)
				s.AddPoint("D", 2, 2)
				s.AddLine("L2", "O3", "D")
				s.SetAngle("L1", "L2", math.Pi/4)
				s.SetDistance("O3", "D", 2)
				return s
			},
		},
	}

	t.Log("=== Quantitative Results ===")
	t.Log("Example                  | n | k  | time       | κ∞(J₀)  | ‖F₀‖      | ‖F*‖")
	t.Log("-------------------------|---|----|-----------:|--------:|----------:|----------:")
	for _, ex := range examples {
		s := ex.build()
		r := s.SatisfyConstraints()
		t.Logf("%-25s| 2 | %2d | %10v | %7.2f | %.4e | %.4e",
			ex.name, r.Iterations, r.SolveTime.Round(1),
			r.ConditionNumber, r.InitialResidualNorm, r.FinalResidualNorm)
	}

	t.Log("")
	t.Log("=== Per-iteration residual history (for convergence plot) ===")
	for _, ex := range examples {
		s := ex.build()
		r := s.SatisfyConstraints()
		t.Logf("--- %s ---", ex.name)
		for k, rk := range r.ResidualHistory {
			t.Logf("  k=%d  ‖F‖=%.6e  ‖d‖=%.6e", k+1, rk, r.StepHistory[k])
		}
	}

	chainSizes := []int{1, 2, 5, 10, 20}
	t.Log("")
	t.Log("=== Scalability: chain sketch ===")
	t.Log("n_pts | n_params | k  | time       | κ∞(J₀)     | converged")
	t.Log("------|----------|----|------------|------------|----------")
	for _, n := range chainSizes {
		s := buildChainSketch(n)
		r := s.SatisfyConstraints()
		t.Logf("%5d | %8d | %2d | %10v | %10.2f | %v",
			n, 2*n, r.Iterations, r.SolveTime.Round(1),
			r.ConditionNumber, r.Converged)
	}
}

func TestSketch_Chain_1(t *testing.T) {
	s := buildChainSketch(1)
	before := captureSketch(s)
	r := s.SatisfyConstraints()
	writeSnapshot(s, before, t.Name(), "testdata/"+t.Name()+".json")
	if !r.Converged {
		t.Fatalf("chain n=1 did not converge: %+v", r)
	}
	AssertAlmost(t, s.GetParam("P1x"), 2.0)
	AssertAlmost(t, s.GetParam("P1y"), 0.0)
}

func TestSketch_Chain_2(t *testing.T) {
	s := buildChainSketch(2)
	before := captureSketch(s)
	r := s.SatisfyConstraints()
	writeSnapshot(s, before, t.Name(), "testdata/"+t.Name()+".json")
	if !r.Converged {
		t.Fatalf("chain n=2 did not converge: %+v", r)
	}
	AssertAlmost(t, s.GetParam("P1x"), 2.0)
	AssertAlmost(t, s.GetParam("P1y"), 0.0)
	AssertAlmost(t, s.GetParam("P2x"), 4.0)
	AssertAlmost(t, s.GetParam("P2y"), 0.0)
}

func TestSketch_Chain_5(t *testing.T) {
	s := buildChainSketch(5)
	before := captureSketch(s)
	r := s.SatisfyConstraints()
	writeSnapshot(s, before, t.Name(), "testdata/"+t.Name()+".json")
	if !r.Converged {
		t.Fatalf("chain n=5 did not converge: %+v", r)
	}
	for i := 1; i <= 5; i++ {
		AssertAlmost(t, s.GetParam(fmt.Sprintf("P%dx", i)), float64(2*i))
		AssertAlmost(t, s.GetParam(fmt.Sprintf("P%dy", i)), 0.0)
	}
}

func TestSketch_Chain_10(t *testing.T) {
	s := buildChainSketch(10)
	before := captureSketch(s)
	r := s.SatisfyConstraints()
	writeSnapshot(s, before, t.Name(), "testdata/"+t.Name()+".json")
	if !r.Converged {
		t.Fatalf("chain n=10 did not converge: %+v", r)
	}
	for i := 1; i <= 10; i++ {
		AssertAlmost(t, s.GetParam(fmt.Sprintf("P%dx", i)), float64(2*i))
		AssertAlmost(t, s.GetParam(fmt.Sprintf("P%dy", i)), 0.0)
	}
}

func TestSketch_Chain_20(t *testing.T) {
	s := buildChainSketch(20)
	before := captureSketch(s)
	r := s.SatisfyConstraints()
	writeSnapshot(s, before, t.Name(), "testdata/"+t.Name()+".json")
	if !r.Converged {
		t.Fatalf("chain n=20 did not converge: %+v", r)
	}
	for i := 1; i <= 20; i++ {
		AssertAlmost(t, s.GetParam(fmt.Sprintf("P%dx", i)), float64(2*i))
		AssertAlmost(t, s.GetParam(fmt.Sprintf("P%dy", i)), 0.0)
	}
}

func buildChainSketch(n int) *Sketch {
	s := NewSketch()
	s.AddOrigin("O", 0, 0)
	for i := 1; i <= n; i++ {
		prev := "O"
		if i > 1 {
			prev = fmt.Sprintf("P%d", i-1)
		}
		li := fmt.Sprintf("L%d", i)
		pi := fmt.Sprintf("P%d", i)
		s.AddOrigin(li, float64(2*i), 1.0)
		s.AddPoint(pi, float64(2*i)-0.5, 0.5)
		s.SetDistance(prev, pi, 2.0)
		s.SetDistance(li, pi, 1.0)
	}
	return s
}