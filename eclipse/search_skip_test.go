package eclipse

import (
	"testing"
	"time"

	"github.com/starainrt/astro/basic"
)

func TestEclipseSearchStepDoesNotSkipPotentialCandidates(t *testing.T) {
	testCases := []struct {
		name             string
		phaseType        int
		synodicMonthDays float64
		potential        func(float64) bool
	}{
		{
			name:             "solar",
			phaseType:        0,
			synodicMonthDays: solarEclipseSynodicMonthDays,
			potential:        isPotentialSolarEclipse,
		},
		{
			name:             "lunar",
			phaseType:        1,
			synodicMonthDays: lunarEclipseSynodicMonthDays,
			potential:        isPotentialLunarEclipse,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			candidates := eclipseSearchTestCandidates(1600, 800, tc.phaseType, tc.synodicMonthDays)

			for index, candidateTT := range candidates {
				for _, direction := range []int{-1, 1} {
					step := eclipseSearchStep(candidateTT, direction, tc.synodicMonthDays)
					for offset := 1; offset < step; offset++ {
						skippedIndex := index + direction*offset
						if skippedIndex < 0 || skippedIndex >= len(candidates) {
							continue
						}
						if tc.potential(candidates[skippedIndex]) {
							t.Fatalf(
								"%s skip crosses potential candidate: index=%d direction=%d step=%d offset=%d jd=%.8f",
								tc.name,
								index,
								direction,
								step,
								offset,
								candidates[skippedIndex],
							)
						}
					}
				}
			}
		})
	}
}

func eclipseSearchTestCandidates(startYear, years, phaseType int, synodicMonthDays float64) []float64 {
	startTT := basic.Date2JDE(time.Date(startYear, 1, 1, 0, 0, 0, 0, time.UTC))
	endTT := basic.Date2JDE(time.Date(startYear+years, 1, 1, 0, 0, 0, 0, time.UTC))
	candidateTT := basic.CalcMoonSHByJDE(startTT, phaseType)
	candidates := make([]float64, 0, years*13)
	for candidateTT < endTT {
		if candidateTT >= startTT {
			candidates = append(candidates, candidateTT)
		}
		candidateTT = basic.CalcMoonSHByJDE(candidateTT+synodicMonthDays, phaseType)
	}
	return candidates
}
