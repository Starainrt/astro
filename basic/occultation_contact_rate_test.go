package basic

import (
	"math"
	"testing"
	"time"
)

// 解析速率与中心差分必须给出同一个接触度量导数：点源目标曾把零位置当站心位置，
// 减掉观测者速度后得到日尺度的假速率，掩带边界因此无法加密。
// The analytic contact rate must agree with the central difference for every target
// kind: the point-source branch once treated a zero position as a topocentric vector
// and produced a spurious diurnal rate that stopped the band boundary from refining.
func TestOccultationContactRateMatchesCentralDifference(t *testing.T) {
	const tolerance = 0.01 // 度/日
	star := hr4799OccultationCoordinateForTest()
	infinite := star
	infinite.ParallaxMas = 0
	parallaxStar := star
	parallaxStar.ParallaxMas = 10
	planetConfig, _ := planetOccultationConfigFor(OccultationMercury)
	planetCache := newPlanetOccultationEventCache(planetConfig)

	starEvaluation := func(coordinate StarCoordinate) func(float64) occultationRiseSetEvaluation {
		cache := newStarOccultationEventCache(coordinate)
		return newOccultationRiseSetEvaluationCache(cache.riseSetContextAt).evaluation
	}
	cases := []struct {
		name       string
		sampleTime time.Time
		evaluation func(float64) occultationRiseSetEvaluation
	}{
		{"point-source star", time.Date(2025, 6, 5, 10, 0, 0, 0, time.UTC), starEvaluation(infinite)},
		{"finite star", time.Date(2025, 6, 5, 10, 0, 0, 0, time.UTC), starEvaluation(parallaxStar)},
		{"planet", time.Date(2024, 8, 21, 10, 0, 0, 0, time.UTC), newOccultationRiseSetEvaluationCache(planetCache.riseSetContextAt).evaluation},
	}
	sites := [][2]float64{{0, 0}, {115, 33}, {-74, 40}, {179, 80}, {-123, -64}}
	for _, testCase := range cases {
		tt := occultationTimeToTT(testCase.sampleTime)
		evaluation := testCase.evaluation(tt)
		for _, site := range sites {
			analytic := evaluation.contactRateAt(site[0], site[1])
			reference := evaluation.contactDerivative(site[0], site[1])
			if !finite(analytic) || !finite(reference) {
				t.Fatalf("%s lon=%g lat=%g: analytic=%g reference=%g", testCase.name, site[0], site[1], analytic, reference)
			}
			if math.Abs(analytic-reference) > tolerance {
				t.Errorf("%s lon=%g lat=%g: analytic=%.6f reference=%.6f degrees/day",
					testCase.name, site[0], site[1], analytic, reference)
			}
		}
	}
}

// 无视差恒星的掩带边界必须加密到请求的地面间距：修复前相邻点可达 150 km。
func TestStarOccultationBandContourSpacing(t *testing.T) {
	star := hr4799OccultationCoordinateForTest()
	start := time.Date(2025, 6, 5, 0, 0, 0, 0, time.UTC)
	paths, err := FindStarOccultationPaths(start, start.Add(24*time.Hour), star,
		OccultationPathOptions{Step: 5 * time.Minute, TargetSpacingKM: 30, DisableFootprints: true, DisableRiseSet: true})
	if err != nil {
		t.Fatal(err)
	}
	const limitKM = 45.0
	sampled := 0
	for _, path := range paths {
		for branch, contour := range path.BandContours {
			for index := 1; index < len(contour); index++ {
				spacing := occultationPathDistanceKM(contour[index-1], contour[index])
				sampled++
				if spacing > limitKM {
					t.Fatalf("branch %d sample %d: spacing %.3f km exceeds %.0f km", branch, index, spacing, limitKM)
				}
			}
		}
	}
	if sampled == 0 {
		t.Fatal("no band contour samples")
	}
}
