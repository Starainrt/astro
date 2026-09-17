package occultationgeo

import (
	"testing"
	"time"

	"github.com/starainrt/astro/basic"
)

func TestDensifyRiseSetCurvesUsesProjectedSpacingAndPreservesEndpoints(t *testing.T) {
	zone := time.FixedZone("UTC+8", 8*60*60)
	firstTime := time.Date(2025, time.July, 29, 0, 0, 0, 0, zone)
	lastTime := firstTime.Add(10 * time.Minute)
	curves := []basic.OccultationRiseSetCurve{{
		Phase:     basic.RiseSetPhaseStart,
		Direction: basic.RiseSetDirectionRise,
		Segments: [][]basic.OccultationPathPoint{{
			{Time: firstTime, Longitude: -130, Latitude: -70},
			{Time: lastTime, Longitude: -128, Latitude: -78},
		}},
	}}
	dense := DensifyRiseSetCurves(curves, 35)
	if len(dense) != 1 || len(dense[0].Segments) != 1 {
		t.Fatalf("curve topology changed: %#v", dense)
	}
	segment := dense[0].Segments[0]
	if len(segment) <= 2 {
		t.Fatalf("points=%d, want projected densification", len(segment))
	}
	if segment[0] != curves[0].Segments[0][0] || segment[len(segment)-1] != curves[0].Segments[0][1] {
		t.Fatalf("endpoints changed: first=%+v last=%+v", segment[0], segment[len(segment)-1])
	}
	for index := 1; index < len(segment); index++ {
		if spacing := projectedGeoPointDistanceKM(segment[index-1], segment[index]); spacing > 35.000001 {
			t.Fatalf("edge %d projected spacing=%.3f km, want <=35 km", index, spacing)
		}
	}
}

func TestDensifyRiseSetCurvesDoesNotMutateInput(t *testing.T) {
	point := basic.OccultationPathPoint{Time: time.Unix(0, 0), Longitude: 1, Latitude: 2}
	curves := []basic.OccultationRiseSetCurve{{Segments: [][]basic.OccultationPathPoint{{point, point}}}}
	_ = DensifyRiseSetCurves(curves, 35)
	if len(curves) != 1 || len(curves[0].Segments) != 1 || len(curves[0].Segments[0]) != 2 {
		t.Fatalf("input mutated: %#v", curves)
	}
}
