package basic

import (
	"math"
	"testing"
)

// 未发生阶段的接触时刻必须是 NaN：JD 0 是 −4713-11-24 的真实时刻，不能当“无此阶段”的哨兵。

func TestLunarEclipseAbsentPhasesAreNaN(t *testing.T) {
	eclipse := LunarEclipse(2458860.0)
	if eclipse.Type != LunarEclipsePenumbral {
		t.Fatalf("type = %s, want %s", eclipse.Type, LunarEclipsePenumbral)
	}
	if !eclipse.HasPenumbral || eclipse.HasPartial || eclipse.HasTotal {
		t.Fatalf("unexpected flags: %+v", eclipse)
	}
	for name, value := range map[string]float64{
		"PartialStart": eclipse.PartialStart,
		"PartialEnd":   eclipse.PartialEnd,
		"TotalStart":   eclipse.TotalStart,
		"TotalEnd":     eclipse.TotalEnd,
	} {
		if !math.IsNaN(value) {
			t.Fatalf("%s = %.12f, want NaN", name, value)
		}
	}
	for name, value := range map[string]float64{
		"PenumbralStart": eclipse.PenumbralStart,
		"PenumbralEnd":   eclipse.PenumbralEnd,
		"Maximum":        eclipse.Maximum,
	} {
		if !isFiniteFloat(value) {
			t.Fatalf("%s = %v, want a finite contact", name, value)
		}
	}
}

func TestLunarEclipseWithoutEclipseHasNoContacts(t *testing.T) {
	// 2025-01-13 的望月没有月食（半影食分 < 0）。
	eclipse := LunarEclipse(JDECalc(2025, 1, 13))
	if eclipse.Type != LunarEclipseNone {
		t.Fatalf("type = %s, want %s", eclipse.Type, LunarEclipseNone)
	}
	if eclipse.HasPenumbral || eclipse.HasPartial || eclipse.HasTotal {
		t.Fatalf("unexpected flags: %+v", eclipse)
	}
	for name, value := range map[string]float64{
		"PenumbralStart": eclipse.PenumbralStart,
		"PenumbralEnd":   eclipse.PenumbralEnd,
		"PartialStart":   eclipse.PartialStart,
		"PartialEnd":     eclipse.PartialEnd,
		"TotalStart":     eclipse.TotalStart,
		"TotalEnd":       eclipse.TotalEnd,
	} {
		if !math.IsNaN(value) {
			t.Fatalf("%s = %.12f, want NaN", name, value)
		}
	}
	if !isFiniteFloat(eclipse.Maximum) {
		t.Fatalf("Maximum = %v, want the geometric extremum", eclipse.Maximum)
	}
}

func TestLunarEclipseTotalKeepsAllContactsFinite(t *testing.T) {
	eclipse := LunarEclipse(JDECalc(2025, 3, 14))
	if eclipse.Type != LunarEclipseTotal || !eclipse.HasPenumbral || !eclipse.HasPartial || !eclipse.HasTotal {
		t.Fatalf("unexpected result: %+v", eclipse)
	}
	contacts := []float64{
		eclipse.PenumbralStart, eclipse.PartialStart, eclipse.TotalStart,
		eclipse.Maximum,
		eclipse.TotalEnd, eclipse.PartialEnd, eclipse.PenumbralEnd,
	}
	for i, value := range contacts {
		if !isFiniteFloat(value) {
			t.Fatalf("contact %d = %v, want finite", i, value)
		}
		if i > 0 && value <= contacts[i-1] {
			t.Fatalf("contacts not increasing at %d: %v", i, contacts)
		}
	}
}

func TestLunarEclipseDiagramSkipsAbsentPhases(t *testing.T) {
	diagram := LunarEclipseDiagram(2458860.0, LunarEclipseDiagramOptions{})
	if len(diagram.Points) == 0 {
		t.Fatalf("penumbral eclipse diagram should still have path points")
	}
	for _, point := range diagram.Points {
		if !isFiniteFloat(point.JDE) || !isFiniteFloat(point.X) || !isFiniteFloat(point.Y) {
			t.Fatalf("non-finite diagram point: %+v", point)
		}
	}
	if empty := LunarEclipseDiagram(JDECalc(2025, 1, 13), LunarEclipseDiagramOptions{}); len(empty.Points) != 0 {
		t.Fatalf("eclipse-free diagram should have no points, got %d", len(empty.Points))
	}
}
