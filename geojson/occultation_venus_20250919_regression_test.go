package geojson_test

import (
	"testing"
	"time"

	"github.com/starainrt/astro/geojson"
	"github.com/starainrt/astro/moon"
)

func TestMarshalPlanetOccultationVenus20250919OptimizedContoursAreStrictlyOrdered(t *testing.T) {
	skipHeavyOnGo116386(t)
	day := time.Date(2025, 9, 19, 0, 0, 0, 0, time.UTC)
	paths, err := moon.FindPlanetOccultationPaths(
		day.Add(-12*time.Hour), day.Add(36*time.Hour), moon.OccultationVenus,
		moon.OccultationPathOptions{
			Step:              5 * time.Minute,
			TargetSpacingKM:   300,
			DisableFootprints: true,
			RiseSetStep:       30 * time.Second,
		},
	)
	if err != nil || len(paths) != 1 {
		t.Fatalf("FindPlanetOccultationPaths paths=%d err=%v, want one path", len(paths), err)
	}
	if len(paths[0].PartialBandContours) != 2 {
		t.Fatalf("partial contours=%d, want two", len(paths[0].PartialBandContours))
	}
	for contourIndex, contour := range paths[0].PartialBandContours {
		for index := 1; index < len(contour); index++ {
			if !contour[index].Time.After(contour[index-1].Time) {
				t.Fatalf("contour %d sample %d is not strictly increasing: %s then %s", contourIndex, index, contour[index-1].Time, contour[index].Time)
			}
		}
	}
	if _, err := geojson.MarshalPlanetOccultation(paths[0]); err != nil {
		t.Fatalf("MarshalPlanetOccultation: %v", err)
	}
}
