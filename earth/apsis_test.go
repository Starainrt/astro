package earth

import (
	"math"
	"testing"
	"time"

	"github.com/starainrt/astro/basic"
)

func TestApsisWrappersMatchBasic(t *testing.T) {
	peri := basic.EarthPerihelion(2026)
	periWrapped := Perihelion(2026)
	if !periWrapped.Time.Equal(basic.JDE2DateByZone(peri.JDE, time.UTC, false)) {
		t.Fatalf("perihelion time mismatch: got %s want %s", periWrapped.Time.Format(time.RFC3339Nano), basic.JDE2DateByZone(peri.JDE, time.UTC, false).Format(time.RFC3339Nano))
	}
	if math.Float64bits(periWrapped.Distance) != math.Float64bits(peri.Distance) {
		t.Fatalf("perihelion distance mismatch: got %.12f want %.12f", periWrapped.Distance, peri.Distance)
	}

	aphe := basic.EarthAphelion(2026)
	apheWrapped := Aphelion(2026)
	if !apheWrapped.Time.Equal(basic.JDE2DateByZone(aphe.JDE, time.UTC, false)) {
		t.Fatalf("aphelion time mismatch: got %s want %s", apheWrapped.Time.Format(time.RFC3339Nano), basic.JDE2DateByZone(aphe.JDE, time.UTC, false).Format(time.RFC3339Nano))
	}
	if math.Float64bits(apheWrapped.Distance) != math.Float64bits(aphe.Distance) {
		t.Fatalf("aphelion distance mismatch: got %.12f want %.12f", apheWrapped.Distance, aphe.Distance)
	}
}
