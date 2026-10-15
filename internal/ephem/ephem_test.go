package ephem

import "testing"

type testEngine struct {
	marker float64
}

func (engine testEngine) State(target, center Body, jdTDB float64) (pos, vel [3]float64, err error) {
	pos[0] = engine.marker + float64(target-center)
	vel[0] = jdTDB
	return pos, vel, nil
}

func (engine testEngine) Coverage(target Body) (start, end float64, ok bool) {
	return engine.marker, engine.marker + float64(target), true
}

func (engine testEngine) Apparent(target Body, jdTT float64) (dir [3]float64, distKM, tauDays float64, err error) {
	dir[0] = engine.marker + float64(target)
	return dir, jdTT, engine.marker, nil
}

func TestInstanceUsesItsOwnEngine(t *testing.T) {
	first := NewInstance(testEngine{marker: 1})
	second := NewInstance(testEngine{marker: 2})
	firstPos, _, err := first.State(Mars, Earth, 2451545)
	if err != nil {
		t.Fatal(err)
	}
	secondPos, _, err := second.State(Mars, Earth, 2451545)
	if err != nil {
		t.Fatal(err)
	}
	if firstPos[0] == secondPos[0] || first.Engine() == second.Engine() {
		t.Fatal("instances are not isolated")
	}
	if got, want := first.Model(), Custom; got != want {
		t.Fatalf("custom model = %v, want %v", got, want)
	}
}

func TestInstanceApparentAndCoverage(t *testing.T) {
	instance := NewInstance(testEngine{marker: 3})
	start, end, ok := instance.Coverage(Mars)
	if !ok || start != 3 || end != 7 {
		t.Fatalf("coverage = %v, %v, %v", start, end, ok)
	}
	dir, distance, tau, err := instance.Apparent(Mars, 2451545)
	if err != nil || dir[0] != 7 || distance != 2451545 || tau != 3 {
		t.Fatalf("apparent = %v, %v, %v, %v", dir, distance, tau, err)
	}
}

func TestInstanceWithoutEngineReportsExplicitErrors(t *testing.T) {
	instance := &Instance{model: BuildDefault}
	if _, _, err := instance.State(Mars, Earth, 2451545); err != ErrNoEngine {
		t.Fatalf("State error = %v, want %v", err, ErrNoEngine)
	}
	if _, _, _, err := instance.Apparent(Mars, 2451545); err != ErrNoEngine {
		t.Fatalf("Apparent error = %v, want %v", err, ErrNoEngine)
	}
}
