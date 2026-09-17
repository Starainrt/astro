package basic

import (
	"math"
	"testing"
	"time"
)

func TestOccultationBoundarySamplingNearTemporalFolds(t *testing.T) {
	for _, algorithm := range []OccultationPathAlgorithm{OccultationPathAlgorithmExact, OccultationPathAlgorithmOptimized} {
		t.Run(string(algorithm), func(t *testing.T) {
			day := time.Date(3627, 9, 20, 0, 0, 0, 0, time.UTC)
			paths, err := FindPlanetOccultationPaths(day.Add(-12*time.Hour), day.Add(36*time.Hour), OccultationJupiter,
				OccultationPathOptions{Algorithm: algorithm, Step: 20 * time.Minute, TargetSpacingKM: 900, DisableFootprints: true, RiseSetStep: time.Minute})
			if err != nil || len(paths) != 1 {
				t.Fatalf("paths=%d err=%v", len(paths), err)
			}
			for _, contours := range [][][]OccultationPathPoint{paths[0].PartialBandContours, paths[0].TotalBandContours} {
				for ci, contour := range contours {
					for i := 1; i+1 < len(contour); i++ {
						a, b, c := contour[i-1], contour[i], contour[i+1]
						first, second := occultationPathDistanceKM(a, b), occultationPathDistanceKM(b, c)
						if first > 0.05 && second > 0.05 && first+second > 2 && occultationRiseSetTurnAngleDegrees(a, b, c) < 30 {
							t.Errorf("contour %d point %d reverses between %.3f/%.3f km edges at %.9f,%.9f", ci, i, first, second, b.Longitude, b.Latitude)
						}
					}
				}
			}
		})
	}
}

func TestOccultationPhaseJunctionSamplingResolvesCurvature(t *testing.T) {
	day := time.Date(1227, 5, 19, 0, 0, 0, 0, time.UTC)
	paths, err := FindPlanetOccultationPaths(day.Add(-12*time.Hour), day.Add(36*time.Hour), OccultationVenus,
		OccultationPathOptions{Algorithm: OccultationPathAlgorithmExact, Step: 20 * time.Minute, TargetSpacingKM: 900, DisableFootprints: true, RiseSetStep: time.Minute})
	if err != nil || len(paths) != 1 {
		t.Fatalf("paths=%d err=%v", len(paths), err)
	}
	config, _ := planetOccultationConfigFor(OccultationVenus)
	cache := newPlanetOccultationEventCache(config)
	checked := 0
	for ci, curve := range paths[0].RiseSetCurves {
		if curve.Phase == RiseSetPhaseGreatest {
			continue
		}
		for _, segment := range curve.Segments {
			if len(segment) < 2 {
				continue
			}
			for _, atStart := range []bool{true, false} {
				endpoint := occultationRiseSetSegmentEndpoint(segment, atStart)
				if !occultationRiseSetEndpointSharesPhaseJunction(paths[0].RiseSetCurves, ci, endpoint) {
					continue
				}
				index := len(segment) - 2
				if atStart {
					index = 0
				}
				a, b := segment[index], segment[index+1]
				middle, ok := occultationRiseSetPhaseMidpoint(a, b, curve.Phase, curve.Direction, time.UTC, cache.riseSetCache)
				if !ok || occultationPathDistanceKM(a, b) < 0.05 {
					continue
				}
				checked++
				deviation := planetOccultationPointSegmentDistanceKM(middle, a, b)
				if math.IsNaN(deviation) || deviation > 0.05 {
					t.Errorf("%s/%s endpoint chord misses the physical arc by %.6f km", curve.Phase, curve.Direction, deviation)
				}
			}
		}
	}
	if checked < 2 {
		t.Fatalf("checked only %d junction chords", checked)
	}
}
