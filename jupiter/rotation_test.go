package jupiter_test

import (
	"math"
	"testing"
	"time"

	"github.com/starainrt/astro"
	"github.com/starainrt/astro/basic"
	"github.com/starainrt/astro/internal/ephem"
	"github.com/starainrt/astro/jupiter"
)

func TestRotationPublicConsistencyPaths(t *testing.T) {
	previous := basic.GetEphemerisModel()
	defer basic.SetEphemerisModel(previous)
	for _, model := range []astro.Model{astro.ModelLegacy, astro.ModelSemianalytic} {
		if err := astro.SetModel(model); err != nil {
			t.Fatal(err)
		}
		inst := ephem.NewModelInstance(ephem.Model(model))
		module := jupiter.NewWithInstance(inst)
		context, err := astro.NewContext(model)
		if err != nil {
			t.Fatal(err)
		}
		paths := []func(time.Time, jupiter.RotationSystem) float64{jupiter.RotationAngle, module.RotationAngle, context.Jupiter().RotationAngle}
		for _, date := range []time.Time{time.Date(1965, 5, 1, 0, 0, 0, 0, time.UTC), time.Date(2000, 1, 1, 12, 0, 0, 0, time.UTC), time.Date(2026, 11, 14, 6, 7, 8, 900000000, time.UTC), time.Date(2105, 1, 1, 0, 0, 0, 0, time.UTC)} {
			jd := basic.UTC2TT(basic.Date2JD(date.UTC()))
			for _, system := range []jupiter.RotationSystem{jupiter.SystemI, jupiter.SystemII, jupiter.SystemIII} {
				want, err := basic.JupiterRotationAngleWithInstance(jd, system, inst)
				if err != nil {
					t.Fatal(err)
				}
				for _, zone := range []*time.Location{time.UTC, time.FixedZone("east", 14*3600), time.FixedZone("west", -12*3600)} {
					for _, query := range paths {
						got := query(date.In(zone), system)
						if math.IsNaN(got) || math.IsInf(got, 0) || math.Float64bits(got) != math.Float64bits(want) {
							t.Fatalf("system=%d date=%v bits %.17g/%.17g", system, date, got, want)
						}
					}
				}
			}
		}
	}
	date := time.Date(2026, 11, 14, 0, 0, 0, 0, time.UTC)
	for _, query := range []func(time.Time, jupiter.RotationSystem) float64{jupiter.RotationAngle, jupiter.NewWithInstance(nil).RotationAngle} {
		for _, system := range []jupiter.RotationSystem{-1, 3, 99} {
			if !math.IsNaN(query(date, system)) {
				t.Fatalf("invalid system=%d did not return NaN", system)
			}
		}
	}
}
