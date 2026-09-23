package basic

import (
	"fmt"
	"testing"
)

func TestSolarEclipseAnnularEnvelopeContainsShadow(t *testing.T) {
	for _, date := range [][3]int{{4005, 4, 22}, {4329, 6, 12}} {
		t.Run(fmt.Sprintf("%04d-%02d-%02d", date[0], date[1], date[2]), func(t *testing.T) {
			result := SolarEclipsePartialFootprints(JDCalc(date[0], date[1], float64(date[2])), SolarEclipsePartialFootprintOptions{
				StepDays: 2.0 / 1440, BoundaryPoints: 96, CentralShadowStepDays: 1.0 / 1440,
			})
			if len(result.CentralBandSegments) == 0 {
				t.Fatal("missing central envelope")
			}
			// The static horizon caps describe visible local greatest. Shadow
			// slices at the limb also include sites whose greatest is below the
			// horizon; use fully risen samples to audit the two side envelopes.
			var visible []SolarEclipsePartialFootprint
			for _, footprint := range result.CentralShadowFootprints {
				for _, boundary := range footprint.Boundaries {
					for _, point := range boundary {
						if point.SunAltitude > 0.5 {
							visible = append(visible, SolarEclipsePartialFootprint{Boundaries: [][]SolarEclipsePathPoint{{point}}})
						}
					}
				}
			}
			if len(visible) == 0 {
				t.Fatal("missing visible shadow witnesses")
			}
			miss := solarEclipseBandFootprintMissDistanceKM(result.CentralBandSegments, visible)
			if miss > 2 {
				t.Fatalf("central envelope excludes instantaneous shadow by %.3f km", miss)
			}
		})
	}
}

func BenchmarkSolarEclipseCentralEnvelopePath(b *testing.B) {
	for _, date := range [][3]int{{4005, 4, 22}, {4329, 6, 12}, {2024, 4, 8}} {
		b.Run(fmt.Sprint(date), func(b *testing.B) {
			seed := JDCalc(date[0], date[1], float64(date[2]))
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				SolarEclipseCentralPath(seed, SolarEclipsePathOptions{StepDays: 2.0 / 1440, TargetSpacingKM: 150})
			}
		})
	}
}
