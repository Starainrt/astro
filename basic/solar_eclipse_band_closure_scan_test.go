package basic

import (
	"math"
	"testing"
)

// 扫描式枚举是与三种子牛顿链彼此独立的取根路径：两边解出的地平闭包根必须一致。
func TestSolarEclipseClosureRootsMatchScan(t *testing.T) {
	testCases := []struct {
		name string
		date [3]int
	}{
		{"1136-06-01", [3]int{1136, 6, 1}},
		{"-1480-12-27", [3]int{-1480, 12, 27}},
		{"5705-06-17", [3]int{5705, 6, 17}},
		{"4862-09-28", [3]int{4862, 9, 28}},
		{"2024-04-08", [3]int{2024, 4, 8}},
		{"2024-10-02", [3]int{2024, 10, 2}},
		{"2026-08-12", [3]int{2026, 8, 12}},
		{"2009-07-22", [3]int{2009, 7, 22}},
		{"2012-05-21", [3]int{2012, 5, 21}},
	}
	const (
		distanceToleranceKM = 0.2
		timeTolerance       = 0.5
	)
	for _, tc := range testCases {
		seed := JDCalc(tc.date[0], tc.date[1], float64(tc.date[2]))
		footprints := SolarEclipsePartialFootprints(seed, SolarEclipsePartialFootprintOptions{
			StepDays: 2.0 / 1440.0, BoundaryPoints: 96, RiseSetStepDays: 2.0 / 1440.0,
		})
		if len(footprints.CentralBandHorizonClosures) != 2 {
			t.Fatalf("%s horizon closures=%d, want two", tc.name, len(footprints.CentralBandHorizonClosures))
		}
		solver := newSolarEclipseSolver(CalcMoonSHByJDE(seed, 0), SolarEclipseModelNASABulletinSplitK)
		riseSetCurves, _ := solver.centralBandRiseSetCurves(footprints.Eclipse, SolarEclipsePartialFootprintOptions{
			StepDays: 2.0 / 1440.0, BoundaryPoints: 96, RiseSetStepDays: 2.0 / 1440.0,
		}, nil, true)
		sides := solarEclipseCentralBandClosureSides(footprints, footprints.Eclipse)
		for index, side := range sides {
			scanned := solver.centralLimitHorizonRootsByScan(side.shadowContactJDE, side.innerContactJDE)
			if len(scanned) != 2 {
				t.Fatalf("%s side%d scan roots=%d, want two", tc.name, index, len(scanned))
			}
			// 扫描必须覆盖三种子牛顿链解出的每一个根：两条路径彼此独立。
			first, last, axisOK := solver.centralLimitHorizonRootsNearAxisContact(
				side.axisContactJDE, side.shadowContactJDE, side.innerContactJDE, side.direction,
			)
			curveRoots, _ := solver.centralLimitHorizonRootsFromCurves(
				riseSetCurves, side.shadowContactJDE, side.innerContactJDE,
			)
			unionRoots, _ := solver.centralLimitHorizonRootsFromSampledBoundary(
				side.shadowContactJDE, side.innerContactJDE, footprints.Eclipse.GreatestEclipse,
				footprints.CentralBandFootprints, riseSetCurves,
			)
			chained := append([]SolarEclipsePathPoint{}, curveRoots...)
			chained = append(chained, unionRoots...)
			if axisOK {
				chained = append(chained, first, last)
			}
			for _, root := range chained {
				nearest := math.Inf(1)
				for _, candidate := range scanned {
					nearest = math.Min(nearest, solarEclipsePathDistanceKM(candidate, root))
				}
				if nearest > distanceToleranceKM {
					t.Fatalf("%s side%d chain root %.3f km from every scan root", tc.name, index, nearest)
				}
			}
			closure := footprints.CentralBandHorizonClosures[index]
			for _, want := range []SolarEclipsePathPoint{closure[0], closure[len(closure)-1]} {
				nearest, nearestJDE := math.Inf(1), 0.0
				for _, root := range scanned {
					if distance := solarEclipsePathDistanceKM(root, want); distance < nearest {
						nearest, nearestJDE = distance, root.JDE
					}
				}
				if nearest > distanceToleranceKM {
					t.Fatalf("%s side%d scan root %.3f km from the closure root", tc.name, index, nearest)
				}
				if math.Abs(nearestJDE-want.JDE)*86400 > timeTolerance {
					t.Fatalf("%s side%d scan root %.3f s from the closure root",
						tc.name, index, math.Abs(nearestJDE-want.JDE)*86400)
				}
			}
		}
	}
}

// 扫描式枚举同样要复现「凑不齐一对根」的判定：不成对的奇点不能凭空造出解析中心带。
func TestSolarEclipseClosureScanRejectsIncompletePairs(t *testing.T) {
	seed := JDCalc(1552, 7, 21)
	footprints := SolarEclipsePartialFootprints(seed, SolarEclipsePartialFootprintOptions{
		StepDays: 2.0 / 1440.0, BoundaryPoints: 96, RiseSetStepDays: 2.0 / 1440.0,
	})
	if len(footprints.CentralBandHorizonClosures) != 0 {
		t.Fatalf("horizon closures=%d, want none", len(footprints.CentralBandHorizonClosures))
	}
	solver := newSolarEclipseSolver(CalcMoonSHByJDE(seed, 0), SolarEclipseModelNASABulletinSplitK)
	for index, side := range solarEclipseCentralBandClosureSides(footprints, footprints.Eclipse) {
		if scanned := solver.centralLimitHorizonRootsByScan(
			side.shadowContactJDE, side.innerContactJDE,
		); len(scanned) >= 2 {
			t.Fatalf("side%d scan roots=%d, want fewer than two", index, len(scanned))
		}
	}
}
