package basic

import (
	"math"
	"testing"
	"time"
)

// 掩带两条横向口径的契约：WidthKM 是中心线采样处的地面横向宽度，南北限是外接触锥的擦边切点轨迹，
// 限线间距由 LimitSeparationKM 单独给出，两者不相等也不可互相换算。

const (
	occultationLimitTangencyToleranceArcsec = 5e-3
	occultationLimitEndpointToleranceArcsec = 0.5
	occultationLimitIndependentTolerance    = 1e-4
	occultationLimitExactTolerance          = 1e-9
	occultationLimitAnchorToleranceKM       = 0.5
	occultationLimitRatioTolerance          = 0.002
	// 掠射端点锚点的容差：只钉住端点补采带来的百公里级缺口，不追求末位。
	occultationLimitGrazingAnchorToleranceKM = 0.01
)

type occultationLimitFixture struct {
	name              string
	planet            OccultationPlanet
	start             time.Time
	widthKM           float64
	limitSeparationKM float64
	totalWidthKM      float64
	limitRatio        float64
	centerLineSamples int
}

func occultationLimitFixtures() []occultationLimitFixture {
	return []occultationLimitFixture{
		{
			name: "Jupiter 1962-10-10", planet: OccultationJupiter,
			start:             time.Date(1962, time.October, 10, 0, 0, 0, 0, time.UTC),
			widthKM:           5319.263900,
			limitSeparationKM: 3593.197135,
			totalWidthKM:      3385.147026,
			limitRatio:        0.675506,
			centerLineSamples: 12,
		},
		{
			name: "Mars 2025-07-29", planet: OccultationMars,
			start:             time.Date(2025, time.July, 29, 0, 0, 0, 0, time.FixedZone("UTC+8", 8*60*60)),
			widthKM:           2190.685431,
			limitSeparationKM: 1819.273229,
			totalWidthKM:      1788.503498,
			limitRatio:        0.830458,
		},
	}
}

func occultationLimitOptions() OccultationPathOptions {
	return OccultationPathOptions{
		Step: 10 * time.Minute, TargetSpacingKM: 900,
		DisableFootprints: true, DisableRiseSet: true,
	}
}

func occultationLimitPath(t *testing.T, fixture occultationLimitFixture) PlanetOccultationPath {
	t.Helper()
	paths, err := FindPlanetOccultationPaths(
		fixture.start, fixture.start.Add(24*time.Hour), fixture.planet, occultationLimitOptions(),
	)
	if err != nil || len(paths) != 1 {
		t.Fatalf("%s: paths=%d err=%v, want one", fixture.name, len(paths), err)
	}
	return paths[0]
}

func occultationLimitExactFrame(t *testing.T, planet OccultationPlanet, centerTT float64) occultationPathFrameFunc {
	t.Helper()
	config, ok := planetOccultationConfigFor(planet)
	if !ok {
		t.Fatalf("%v config is unavailable", planet)
	}
	cache := newPlanetOccultationEventCache(config)
	cache.preparePathEphemeris(centerTT, OccultationPathAlgorithmExact)
	return cache.outerFrameAt
}

func occultationLimitExternalContactGap(t *testing.T, planet OccultationPlanet, point OccultationPathPoint) (float64, bool) {
	t.Helper()
	config, ok := planetOccultationConfigFor(planet)
	if !ok {
		t.Fatalf("%v config is unavailable", planet)
	}
	observer := Observer{Longitude: point.Longitude, Latitude: point.Latitude}
	state := planetOccultationStateAt(centerTimeTT(point.Time), config, &observer, -1)
	if !state.valid {
		return 0, false
	}
	return state.externalContactMetric, true
}

// occultationLimitIndependentGroundWidth 按定义独立重算地面横向宽度：自建接触弧网格并自算地面横向方向。
func occultationLimitIndependentGroundWidth(tt float64, frameAt occultationPathFrameFunc) (float64, bool) {
	frame, ok := frameAt(tt)
	if !ok {
		return 0, false
	}
	before, beforeOK := frameAt(tt - occultationPathVelocityStepDays)
	after, afterOK := frameAt(tt + occultationPathVelocityStepDays)
	if !beforeOK || !afterOK {
		return 0, false
	}
	center, centerOK := occultationPathTrackReference(frame)
	beforeCenter, beforeCenterOK := occultationPathTrackReference(before)
	afterCenter, afterCenterOK := occultationPathTrackReference(after)
	if !centerOK || !beforeCenterOK || !afterCenterOK {
		return 0, false
	}
	rotation := occultationPathEarthRotationAt(tt)
	centerFixed := occultationPathEarthFixedVectorWithRotation(center, rotation)
	beforeFixed := occultationPathEarthFixedVector(tt-occultationPathVelocityStepDays, beforeCenter)
	afterFixed := occultationPathEarthFixedVector(tt+occultationPathVelocityStepDays, afterCenter)
	polar := occultationPathEarthPolarRatio * occultationPathEarthPolarRatio
	normal := occultationPathUnit(occultationPathVector{x: centerFixed.x, y: centerFixed.y, z: centerFixed.z / polar})
	track := occultationPathSub(afterFixed, beforeFixed)
	track = occultationPathSub(track, occultationPathScale(normal, occultationPathDot(track, normal)))
	if occultationPathNorm(track) <= 1e-12 {
		return 0, false
	}
	cross := occultationPathUnit(occultationPathCross(normal, occultationPathUnit(track)))
	offset := func(vector occultationPathVector) float64 {
		fixed := occultationPathEarthFixedVectorWithRotation(vector, rotation)
		return occultationPathDot(occultationPathSub(fixed, centerFixed), cross)
	}
	minimum, maximum := math.Inf(1), math.Inf(-1)
	consider := func(vector occultationPathVector) {
		value := offset(vector)
		minimum = math.Min(minimum, value)
		maximum = math.Max(maximum, value)
	}
	const probeSamples = 4096
	for _, interval := range occultationPathBoundaryThetaIntervals(frame) {
		for index := 0; index <= probeSamples; index++ {
			theta := interval.left + (interval.right-interval.left)*float64(index)/probeSamples
			if vector, _, pointOK := occultationPathBoundaryVector(frame, theta); pointOK {
				consider(vector)
			}
		}
		// 掠射时横向极值恰好落在可见 θ 区间的相切端点上，端点判别式为 0：等差写法的末样本
		// 与 interval.right 相差 1 ULP，足以把判别式推成负值而丢掉端点，端点必须按原值补采。
		if vector, _, pointOK := occultationPathBoundaryVector(frame, interval.right); pointOK {
			consider(vector)
		}
	}
	for index := 0; index < occultationPathBoundaryScanPoints; index++ {
		if vector, _, pointOK := occultationPathBoundaryVector(frame, 2*math.Pi*float64(index)/float64(occultationPathBoundaryScanPoints)); pointOK {
			consider(vector)
		}
	}
	if !finite(minimum) || !finite(maximum) {
		return 0, false
	}
	return maximum - minimum, true
}

func TestPlanetOccultationLimitTracksAreGrazingTangencies(t *testing.T) {
	for _, fixture := range occultationLimitFixtures() {
		path := occultationLimitPath(t, fixture)
		if len(path.NorthernLimit) != len(path.SouthernLimit) || len(path.NorthernLimit) < 32 {
			t.Fatalf("%s: limit sample counts=%d/%d, want equal and at least 32",
				fixture.name, len(path.NorthernLimit), len(path.SouthernLimit))
		}
		checked := 0
		for _, track := range [][]OccultationPathPoint{path.NorthernLimit, path.SouthernLimit} {
			for index, point := range track {
				gap, ok := occultationLimitExternalContactGap(t, fixture.planet, point)
				if !ok {
					t.Fatalf("%s: limit point %d at %v has no valid contact state", fixture.name, index, point.Time)
				}
				// 首末采样是南北限的公共端点（构造上重合），残差略大，单独放宽。
				if index == 0 || index == len(track)-1 {
					if math.Abs(gap) > occultationLimitEndpointToleranceArcsec {
						t.Fatalf("%s: closure endpoint %d external contact gap=%.6f arcsec, want at most %.6f",
							fixture.name, index, gap, occultationLimitEndpointToleranceArcsec)
					}
					continue
				}
				if index >= 3 && index < len(track)-3 {
					checked++
				}
				if math.Abs(gap) > occultationLimitTangencyToleranceArcsec {
					t.Fatalf("%s: limit point %d at %v external contact gap=%.6f arcsec, want at most %.6f",
						fixture.name, index, point.Time, gap, occultationLimitTangencyToleranceArcsec)
				}
			}
		}
		if checked < 100 {
			t.Fatalf("%s: checked %d interior limit points, want at least 100", fixture.name, checked)
		}
		if separation := occultationPathDistanceKM(path.NorthernLimit[0], path.SouthernLimit[0]); separation != 0 {
			t.Fatalf("%s: start closure separation=%.6f km, want the shared endpoint", fixture.name, separation)
		}
		last := len(path.NorthernLimit) - 1
		if separation := occultationPathDistanceKM(path.NorthernLimit[last], path.SouthernLimit[last]); separation != 0 {
			t.Fatalf("%s: end closure separation=%.6f km, want the shared endpoint", fixture.name, separation)
		}
	}
}

func TestPlanetOccultationWidthKMIsGroundCrossTrackWidth(t *testing.T) {
	for _, fixture := range occultationLimitFixtures() {
		path := occultationLimitPath(t, fixture)
		if len(path.CenterLine) != fixture.centerLineSamples {
			t.Fatalf("%s: center line samples=%d, want %d", fixture.name, len(path.CenterLine), fixture.centerLineSamples)
		}
		frameAt := occultationLimitExactFrame(t, fixture.planet, centerTimeTT(path.Greatest.Time))
		for _, point := range append([]OccultationPathPoint{path.Greatest}, path.CenterLine...) {
			tt := centerTimeTT(point.Time)
			_, _, exact, ok := occultationPathLimitsAndWidthForFrame(tt, frameAt)
			if !ok || exact <= 0 {
				t.Fatalf("%s: width at %v is unavailable", fixture.name, point.Time)
			}
			if difference := math.Abs(point.WidthKM - exact); difference > occultationLimitExactTolerance*exact {
				t.Fatalf("%s: width at %v field=%.9f recomputed=%.9f, want within %.3e relative",
					fixture.name, point.Time, point.WidthKM, exact, occultationLimitExactTolerance)
			}
			independent, ok := occultationLimitIndependentGroundWidth(tt, frameAt)
			if !ok || independent <= 0 {
				t.Fatalf("%s: independent ground width at %v is unavailable", fixture.name, point.Time)
			}
			if difference := math.Abs(point.WidthKM - independent); difference > occultationLimitIndependentTolerance*independent {
				t.Fatalf("%s: width at %v field=%.9f independent ground cross-track width=%.9f, want within %.3e relative",
					fixture.name, point.Time, point.WidthKM, independent, occultationLimitIndependentTolerance)
			}
		}
		for _, point := range path.CenterLine {
			if point.LimitSeparationKM != 0 {
				t.Fatalf("%s: center-line point at %v carries limit separation %.6f, want zero outside limit tracks",
					fixture.name, point.Time, point.LimitSeparationKM)
			}
		}
	}
}

func TestPlanetOccultationLimitSeparationMatchesExportedLimits(t *testing.T) {
	for _, fixture := range occultationLimitFixtures() {
		path := occultationLimitPath(t, fixture)
		nearestIndex, nearestDelta := 0, math.Inf(1)
		for index := range path.NorthernLimit {
			if !path.NorthernLimit[index].Time.Equal(path.SouthernLimit[index].Time) {
				t.Fatalf("%s: limit sample %d times differ between tracks", fixture.name, index)
			}
			separation := occultationPathDistanceKM(path.NorthernLimit[index], path.SouthernLimit[index])
			if path.NorthernLimit[index].LimitSeparationKM != separation ||
				path.SouthernLimit[index].LimitSeparationKM != separation {
				t.Fatalf("%s: limit sample %d separations=%.9f/%.9f, want the exported pair distance %.9f",
					fixture.name, index, path.NorthernLimit[index].LimitSeparationKM,
					path.SouthernLimit[index].LimitSeparationKM, separation)
			}
			if delta := math.Abs(centerTimeTT(path.NorthernLimit[index].Time) - centerTimeTT(path.Greatest.Time)); delta < nearestDelta {
				nearestIndex, nearestDelta = index, delta
			}
		}
		nearest := occultationPathDistanceKM(path.NorthernLimit[nearestIndex], path.SouthernLimit[nearestIndex])
		if path.GreatestLimitSeparationKM != nearest {
			t.Fatalf("%s: greatest limit separation=%.9f, want the nearest limit sample pair %.9f at %v",
				fixture.name, path.GreatestLimitSeparationKM, nearest, path.NorthernLimit[nearestIndex].Time)
		}
		for index := range path.NorthernTotalLimit {
			separation := occultationPathDistanceKM(path.NorthernTotalLimit[index], path.SouthernTotalLimit[index])
			if path.NorthernTotalLimit[index].LimitSeparationKM != separation {
				t.Fatalf("%s: total limit sample %d separation=%.9f, want %.9f",
					fixture.name, index, path.NorthernTotalLimit[index].LimitSeparationKM, separation)
			}
		}
		for _, anchor := range []struct {
			name string
			got  float64
			want float64
		}{
			{name: "greatest width", got: path.Greatest.WidthKM, want: fixture.widthKM},
			{name: "greatest limit separation", got: path.GreatestLimitSeparationKM, want: fixture.limitSeparationKM},
			{name: "greatest total width", got: path.GreatestTotalWidthKM, want: fixture.totalWidthKM},
		} {
			if difference := math.Abs(anchor.got - anchor.want); difference > occultationLimitAnchorToleranceKM {
				t.Fatalf("%s: %s=%.6f km, want %.6f km within %.3f km",
					fixture.name, anchor.name, anchor.got, anchor.want, occultationLimitAnchorToleranceKM)
			}
		}
		ratio := path.GreatestLimitSeparationKM / path.Greatest.WidthKM
		if math.Abs(ratio-fixture.limitRatio) > occultationLimitRatioTolerance {
			t.Fatalf("%s: limit separation/width=%.6f, want %.6f within %.4f",
				fixture.name, ratio, fixture.limitRatio, occultationLimitRatioTolerance)
		}
		if path.GreatestTotalWidthKM >= path.Greatest.WidthKM {
			t.Fatalf("%s: total width=%.6f must stay below the outer width=%.6f",
				fixture.name, path.GreatestTotalWidthKM, path.Greatest.WidthKM)
		}
	}
}

func TestStarOccultationLimitSeparationMatchesExportedLimits(t *testing.T) {
	start := time.Date(2025, time.June, 5, 0, 0, 0, 0, time.UTC)
	paths, err := FindStarOccultationPaths(
		start, start.Add(24*time.Hour), hr4799OccultationCoordinateForTest(),
		OccultationPathOptions{
			Step: 20 * time.Minute, TargetSpacingKM: 900,
			DisableFootprints: true, DisableRiseSet: true,
		},
	)
	if err != nil || len(paths) != 1 {
		t.Fatalf("FindStarOccultationPaths() paths=%d err=%v, want one", len(paths), err)
	}
	path := paths[0]
	nearestIndex, nearestDelta := 0, math.Inf(1)
	for index := range path.NorthernLimit {
		separation := occultationPathDistanceKM(path.NorthernLimit[index], path.SouthernLimit[index])
		if path.NorthernLimit[index].LimitSeparationKM != separation {
			t.Fatalf("limit sample %d separation=%.9f, want %.9f",
				index, path.NorthernLimit[index].LimitSeparationKM, separation)
		}
		if delta := math.Abs(centerTimeTT(path.NorthernLimit[index].Time) - centerTimeTT(path.Greatest.Time)); delta < nearestDelta {
			nearestIndex, nearestDelta = index, delta
		}
	}
	nearest := occultationPathDistanceKM(path.NorthernLimit[nearestIndex], path.SouthernLimit[nearestIndex])
	if path.GreatestLimitSeparationKM != nearest {
		t.Fatalf("greatest limit separation=%.9f, want %.9f", path.GreatestLimitSeparationKM, nearest)
	}
	if math.Abs(path.Greatest.WidthKM-3582.363599) > occultationLimitAnchorToleranceKM ||
		math.Abs(path.GreatestLimitSeparationKM-3666.576690) > occultationLimitAnchorToleranceKM {
		t.Fatalf("greatest width=%.6f limit separation=%.6f, want 3582.363599 and 3666.576690 within %.3f km",
			path.Greatest.WidthKM, path.GreatestLimitSeparationKM, occultationLimitAnchorToleranceKM)
	}
	ratio := path.GreatestLimitSeparationKM / path.Greatest.WidthKM
	if math.Abs(ratio-1.023508) > occultationLimitRatioTolerance {
		t.Fatalf("greatest limit separation/width=%.6f, want 1.023508 within %.4f", ratio, occultationLimitRatioTolerance)
	}
}

// TestPlanetOccultationGrazingEndpointWidthAnchors 钉住两个掠射相切端点样本的带宽。
//
// 这两个时刻的横向极值恰好落在可见 θ 区间的相切端点上，只能靠端点原值取到：等差写法
// left+(right-left)*n/n 与端点相差 1 ULP，该 ULP 会把端点判别式推成负值而被判"无地面交点"，
// 于是 Saturn 2024-08-21 的 CenterLine[160] 少 47.9 km（3810.7491 对 3858.7641）、
// Venus 1227-05-19 的 CenterLine[15] 少 158.6 km（6047.4858 对 6206.0535）。
func TestPlanetOccultationGrazingEndpointWidthAnchors(t *testing.T) {
	testCases := []struct {
		name    string
		planet  OccultationPlanet
		start   time.Time
		options OccultationPathOptions
		tt      float64
		widthKM float64
	}{
		{
			name: "Saturn 2024-08-21 center-line[160]", planet: OccultationSaturn,
			start:   time.Date(2024, time.August, 21, 0, 0, 0, 0, time.UTC),
			tt:      2460543.661859489,
			widthKM: 3858.7641,
		},
		{
			name: "Venus 1227-05-19 center-line[15]", planet: OccultationVenus,
			start: time.Date(1227, time.May, 18, 12, 0, 0, 0, time.UTC),
			options: OccultationPathOptions{
				Algorithm: OccultationPathAlgorithmExact, Step: 20 * time.Minute,
				TargetSpacingKM: 900, DisableFootprints: true, RiseSetStep: time.Minute,
			},
			tt:      2169357.811307699,
			widthKM: 6206.0535,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			paths, err := FindPlanetOccultationPaths(tc.start, tc.start.Add(24*time.Hour), tc.planet, tc.options)
			if err != nil || len(paths) != 1 {
				t.Fatalf("paths=%d err=%v, want one", len(paths), err)
			}
			path := paths[0]
			frameAt := occultationLimitExactFrame(t, tc.planet, centerTimeTT(path.Greatest.Time))
			matched := 0
			for _, point := range append([]OccultationPathPoint{path.Greatest}, path.CenterLine...) {
				tt := centerTimeTT(point.Time)
				if math.Abs(tt-tc.tt) > 1e-6 {
					continue
				}
				matched++
				if difference := math.Abs(point.WidthKM - tc.widthKM); difference > occultationLimitGrazingAnchorToleranceKM {
					t.Fatalf("width at %v = %.6f km, want %.4f within %.4f km",
						point.Time, point.WidthKM, tc.widthKM, occultationLimitGrazingAnchorToleranceKM)
				}
				independent, ok := occultationLimitIndependentGroundWidth(tt, frameAt)
				if !ok || independent <= 0 {
					t.Fatalf("independent ground width at %v is unavailable", point.Time)
				}
				if difference := math.Abs(point.WidthKM - independent); difference > occultationLimitIndependentTolerance*independent {
					t.Fatalf("width at %v field=%.9f independent=%.9f, want within %.3e relative",
						point.Time, point.WidthKM, independent, occultationLimitIndependentTolerance)
				}
			}
			if matched != 1 {
				t.Fatalf("matched %d samples at tt=%.9f, want exactly one", matched, tc.tt)
			}
		})
	}
}
