package basic

import (
	"testing"
	"time"
)

func TestPlanetOccultationContactPhaseJunctionsClose(t *testing.T) {
	for _, sample := range []struct {
		year, month, day int
		planet           OccultationPlanet
	}{
		{2027, 4, 15, OccultationJupiter},
		{2027, 1, 8, OccultationMercury},
		// 2025-01-05 月掩海王星在环极折点处残差与零相切，符号扫描漏根会让
		// 「终掩在月升/月落」分支断开，这里固定该回归。
		// The 2025-01-05 Neptune occultation tangents the zero residual at its
		// polar fold; a sign-only scan drops that root and breaks the end phase
		// branches apart, so pin the regression here.
		{2025, 1, 5, OccultationNeptune},
	} {
		t.Run(string(sample.planet), func(t *testing.T) {
			day := time.Date(sample.year, time.Month(sample.month), sample.day, 0, 0, 0, 0, time.UTC)
			paths, err := FindPlanetOccultationPaths(day.Add(-12*time.Hour), day.Add(36*time.Hour), sample.planet,
				OccultationPathOptions{Step: 20 * time.Minute, TargetSpacingKM: 900, DisableFootprints: true, RiseSetStep: time.Minute})
			if err != nil || len(paths) != 1 {
				t.Fatalf("paths=%d err=%v", len(paths), err)
			}
			for _, curves := range [][]OccultationRiseSetCurve{paths[0].RiseSetCurves, paths[0].TotalRiseSetCurves} {
				for ci, curve := range curves {
					if curve.Phase == RiseSetPhaseGreatest {
						continue
					}
					for si, segment := range curve.Segments {
						if len(segment) < 2 {
							continue
						}
						for _, endpoint := range []OccultationPathPoint{segment[0], segment[len(segment)-1]} {
							shared := false
							for oi, other := range curves {
								if other.Phase == RiseSetPhaseGreatest {
									continue
								}
								for os, points := range other.Segments {
									if ci == oi && si == os || len(points) < 2 {
										continue
									}
									for _, point := range []OccultationPathPoint{points[0], points[len(points)-1]} {
										dt := point.Time.Sub(endpoint.Time)
										if dt >= -time.Second && dt <= time.Second && occultationPathDistanceKM(endpoint, point) < 0.01 {
											shared = true
										}
									}
								}
							}
							if !shared {
								t.Errorf("unclosed %s/%s endpoint lon=%.9f lat=%.9f at=%s", curve.Phase, curve.Direction, endpoint.Longitude, endpoint.Latitude, endpoint.Time)
							}
						}
					}
				}
			}
		})
	}
}
