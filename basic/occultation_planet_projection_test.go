package basic

import (
	"math"
	"testing"
	"time"
)

func TestMars20250729FiniteContactContoursRespectWebMercatorSpacing(t *testing.T) {
	zone := time.FixedZone("UTC+8", 8*60*60)
	start := time.Date(2025, time.July, 29, 0, 0, 0, 0, zone)
	paths, err := FindPlanetOccultationPaths(start, start.Add(24*time.Hour), OccultationMars, OccultationPathOptions{
		Step: 20 * time.Minute, TargetSpacingKM: 900, RiseSetStep: time.Minute,
		DisableFootprints: true, IncludeFootprintTimeline: true, FootprintTimelineStep: 5 * time.Minute,
	})
	if err != nil || len(paths) != 1 {
		t.Fatalf("FindPlanetOccultationPaths() paths=%d err=%v, want one", len(paths), err)
	}
	for _, item := range []struct {
		name     string
		contours [][]OccultationPathPoint
	}{
		{name: "partial", contours: paths[0].PartialBandContours},
		{name: "total", contours: paths[0].TotalBandContours},
	} {
		for contourIndex, contour := range item.contours {
			for pointIndex := 1; pointIndex < len(contour); pointIndex++ {
				spacing := webMercatorGeoPointSpacingKM(contour[pointIndex-1], contour[pointIndex])
				if spacing > 45 {
					t.Fatalf("%s contour %d edge %d has %.1f km Web Mercator spacing, want <=45 km", item.name, contourIndex, pointIndex, spacing)
				}
			}
		}
	}
}

func webMercatorGeoPointSpacingKM(first, second OccultationPathPoint) float64 {
	const maxLatitude = 85.05112878
	const radiusKM = 6378.1366
	clamp := func(value float64) float64 { return math.Max(-maxLatitude, math.Min(maxLatitude, value)) }
	longitude := math.Remainder(second.Longitude-first.Longitude, 360) * math.Pi / 180
	firstLatitude := clamp(first.Latitude) * math.Pi / 180
	secondLatitude := clamp(second.Latitude) * math.Pi / 180
	firstY := math.Log(math.Tan(math.Pi/4 + firstLatitude/2))
	secondY := math.Log(math.Tan(math.Pi/4 + secondLatitude/2))
	return radiusKM * math.Hypot(longitude, secondY-firstY)
}
