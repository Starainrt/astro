package basic

import (
	"fmt"
	"math"
	"testing"
)

func TestSolarEclipsePolarCentralEnvelope(t *testing.T) {
	for _, sample := range []struct {
		year, month, day int
		direction        RiseSetDirection
	}{
		{2061, 10, 13, RiseSetDirectionRise},
		{2026, 2, 17, RiseSetDirectionSet},
	} {
		t.Run(fmt.Sprintf("%04d-%02d-%02d", sample.year, sample.month, sample.day), func(t *testing.T) {
			seed := JDCalc(sample.year, sample.month, float64(sample.day))
			result := SolarEclipsePartialFootprints(seed, SolarEclipsePartialFootprintOptions{
				StepDays: 2.0 / 1440, BoundaryPoints: 96,
			})
			if result.Eclipse.Type != SolarEclipseAnnular || len(result.CentralBandSegments) != 1 ||
				len(result.CentralBandHorizonClosures) != 2 {
				t.Fatalf("type=%s rings=%d closures=%d, want annular with a continuous closed envelope",
					result.Eclipse.Type, len(result.CentralBandSegments), len(result.CentralBandHorizonClosures))
			}
			solver := newSolarEclipseSolver(CalcMoonSHByJDE(seed, 0), SolarEclipseModelNASABulletinSplitK)
			for index, closure := range result.CentralBandHorizonClosures {
				if len(closure) < 2 {
					t.Fatalf("closure %d has no arc", index)
				}
				for _, root := range []SolarEclipsePathPoint{closure[0], closure[len(closure)-1]} {
					evaluation := solver.magnitudeEvaluationAt(root.JDE)
					_, key, ok := evaluation.classify(root.Longitude, root.Latitude, true)
					if !ok || key.direction != sample.direction {
						t.Fatalf("closure %d direction=%s valid=%v, want %s", index, key.direction, ok, sample.direction)
					}
					residual, _, ok := solarEclipseCentralLimitHorizonJacobian(
						solver, [3]float64{root.Longitude, root.Latitude, root.JDE},
						solver.magnitudeEvaluationAt,
					)
					if !ok || math.Abs(residual[0]) > 1e-9 || math.Abs(residual[1]) > 1e-7 || math.Abs(residual[2]) > 1e-9 {
						t.Fatalf("closure %d root residual=%v valid=%v", index, residual, ok)
					}
				}
				for pointIndex, point := range closure {
					matched := false
					for _, curve := range result.RiseSetCurves {
						if curve.Phase != RiseSetPhaseGreatest || curve.Direction != sample.direction {
							continue
						}
						for _, segment := range curve.Segments {
							for _, candidate := range segment {
								if math.Abs(candidate.JDE-point.JDE) < solarEclipsePathDuplicateTimeDays &&
									solarEclipsePathDistanceKM(candidate, point) < 0.001 {
									matched = true
								}
							}
						}
					}
					if !matched {
						t.Fatalf("closure %d point %d is missing from greatest/%s", index, pointIndex, sample.direction)
					}
				}
			}
			ring := result.CentralBandSegments[0]
			if len(ring) < 100 || solarEclipsePathDistanceKM(ring[0], ring[len(ring)-1]) > 0.001 {
				t.Fatal("central envelope is not densely sampled and closed")
			}
			for index, point := range ring {
				if index > 0 && solarEclipsePathDistanceKM(ring[index-1], point) > solarEclipseCentralEnvelopeMaxSpacingKM {
					t.Fatalf("envelope edge %d exceeds spacing limit", index)
				}
				if solarEclipseCentralEnvelopePointOnHorizonClosure(point, result.CentralBandHorizonClosures) {
					continue
				}
				residual, ok := solarEclipseNonCentralBandBoundaryResidualAt(solver.magnitudeEvaluationAt(point.JDE), point.Longitude, point.Latitude)
				if !ok || math.Abs(residual[0]) > 1e-6 || math.Abs(residual[1]) > solarEclipseNonCentralBandDerivativeTolerance || point.SunAltitude < 0 {
					t.Fatalf("envelope point %d altitude=%g residual=%v valid=%v", index, point.SunAltitude, residual, ok)
				}
			}
		})
	}
}
