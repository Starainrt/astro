package basic

import (
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/starainrt/astro/tools"
)

// 本文件锁定掩星恒星坐标的距离契约：无距离走二维、给光年或视差走三维。

func occStar3DTestCoordinate(ra, dec float64) StarCoordinate {
	return StarCoordinate{
		ID:    "3D contract",
		RA:    ra,
		Dec:   dec,
		Epoch: time.Date(2000, time.January, 1, 12, 0, 0, 0, time.UTC),
		Frame: CoordinateFrameJ2000,
	}
}

func TestStarCoordinateDistanceGateKeepsTwoDimensions(t *testing.T) {
	star := occStar3DTestCoordinate(189.1975, -5.831944444444)
	star.ProperMotionRACosDecMasPerYear = -28
	star.ProperMotionDecMasPerYear = -18
	star.RadialVelocityKmPerSecond = -45
	if got := star.parallaxMas(); got != 0 {
		t.Fatalf("parallaxMas() = %v, want 0 without any distance", got)
	}
	tt := occultationTimeToTT(time.Date(2050, 1, 1, 0, 0, 0, 0, time.UTC))
	gotRA, gotDec, gotDistance := starProperMotionRaDec(tt, star, star.RA, star.Dec, starEpochDistanceAU(star.parallaxMas()))
	if gotDistance != 0 {
		t.Fatalf("no-distance propagation reported distance %v, want 0", gotDistance)
	}
	cosDec := math.Cos(star.Dec * math.Pi / 180)
	years := (tt - occultationTimeToTT(star.Epoch)) / 365.25
	wantRA := star.RA + years*star.ProperMotionRACosDecMasPerYear/(3600000*cosDec)
	wantDec := star.Dec + years*star.ProperMotionDecMasPerYear/3600000
	if gotRA != wantRA || gotDec != wantDec {
		t.Fatalf("no-distance propagation = %.12f %.12f, want 2D %.12f %.12f", gotRA, gotDec, wantRA, wantDec)
	}
}

func TestStarCoordinateLightYearDistanceMatchesParallax(t *testing.T) {
	const lightYears = 10
	byLightYear := func() StarCoordinate {
		star := occStar3DTestCoordinate(224.366667, -21.415556)
		star.ProperMotionRACosDecMasPerYear = 1045
		star.ProperMotionDecMasPerYear = -1729
		star.RadialVelocityKmPerSecond = 20
		star.DistanceLightYear = lightYears
		return star
	}()
	byParallax := byLightYear
	byParallax.DistanceLightYear = 0
	byParallax.ParallaxMas = byLightYear.parallaxMas()
	if byParallax.ParallaxMas <= 0 || byParallax.ParallaxMas > 400 {
		t.Fatalf("derived parallax = %v mas, want a positive sub-arcsecond value", byParallax.ParallaxMas)
	}
	tt := occultationTimeToTT(time.Date(2050, 1, 1, 0, 0, 0, 0, time.UTC))
	lightRA, lightDec := starApparentRaDecGeocentric(tt, byLightYear)
	parallaxRA, parallaxDec := starApparentRaDecGeocentric(tt, byParallax)
	if lightRA != parallaxRA || lightDec != parallaxDec {
		t.Fatalf("light-year path = %.12f %.12f, want bit-identical to parallax path %.12f %.12f",
			lightRA, lightDec, parallaxRA, parallaxDec)
	}
	if separation := starSepArcsec(lightRA, lightDec, byLightYear.RA, byLightYear.Dec); separation < 60 {
		t.Fatalf("proper motion over 50 years moved the star only %.6f arcsec", separation)
	}
}

func TestStarCoordinateParallaxTakesPriorityOverLightYear(t *testing.T) {
	star := occStar3DTestCoordinate(120, 15)
	star.DistanceLightYear = 10
	if got, want := star.parallaxMas(), 1000/tools.DistanceToParsecs(10, tools.DistanceLightYear); math.Abs(got-want) > 1e-9 {
		t.Fatalf("parallaxMas() = %.12f, want %.12f derived from light-years", got, want)
	}
	star.ParallaxMas = 25
	if got := star.parallaxMas(); got != 25 {
		t.Fatalf("parallaxMas() = %v, want the explicit 25 mas to win", got)
	}
}

func TestStarCoordinateRadialVelocityContract(t *testing.T) {
	base := occStar3DTestCoordinate(120, 15)
	for _, velocity := range []float64{0, -500, 500, starRadialVelocityLimitKmPerSecond} {
		star := base
		star.RadialVelocityKmPerSecond = velocity
		if err := star.Validate(); err != nil {
			t.Fatalf("radial velocity %v rejected: %v", velocity, err)
		}
	}
	for _, velocity := range []float64{-5000, 5000, math.NaN(), math.Inf(1), math.Inf(-1)} {
		star := base
		star.RadialVelocityKmPerSecond = velocity
		if err := star.Validate(); err == nil {
			t.Fatalf("radial velocity %v accepted, want a contract error", velocity)
		}
	}
	for _, distance := range []float64{-1, math.NaN(), math.Inf(1)} {
		star := base
		star.DistanceLightYear = distance
		if err := star.Validate(); err == nil {
			t.Fatalf("light-year distance %v accepted, want a contract error", distance)
		}
	}
}

func TestStarCoordinateThreeDimensionsShiftsOccultationTimingWithinBudget(t *testing.T) {
	// HR 5568 于 2026-01-13 的一次真实全掩；同一站、同一窗口下比较三种口径的真实接触时刻。
	star := occStar3DTestCoordinate(224.366667, -21.415556)
	star.ProperMotionRACosDecMasPerYear = 1045
	star.ProperMotionDecMasPerYear = -1729
	star.ParallaxMas = 173
	star.RadialVelocityKmPerSecond = 20
	noRadial := star
	noRadial.RadialVelocityKmPerSecond = 0
	noDistance := star
	noDistance.ParallaxMas = 0

	const longitude, latitude = 97.471, -45.720
	start := time.Date(2026, 1, 12, 18, 0, 0, 0, time.UTC)
	end := time.Date(2026, 1, 13, 6, 0, 0, 0, time.UTC)

	event := func(c StarCoordinate) StarOccultationInfo {
		t.Helper()
		events, err := FindStarOccultations(start, end, c, longitude, latitude, 0, OccultationSearchOptions{})
		if err != nil {
			t.Fatalf("FindStarOccultations: %v", err)
		}
		if len(events) != 1 {
			t.Fatalf("event count = %d, want 1", len(events))
		}
		return events[0]
	}
	// 无距离确实走的是二维分支，否则下面的对照没有意义。
	if noDistance.parallaxMas() != 0 || star.parallaxMas() != 173 {
		t.Fatalf("parallax gate = %v / %v, want 0 for the 2D branch and 173 for the 3D branch",
			noDistance.parallaxMas(), star.parallaxMas())
	}
	unlimited := event(star)
	limited := event(noDistance)
	if unlimited.Type != OccultationTotal || limited.Type != OccultationTotal {
		t.Fatalf("event types = %v / %v, want both total", unlimited.Type, limited.Type)
	}

	milliseconds := func(a, b time.Time) float64 { return math.Abs(a.Sub(b).Seconds()) * 1000 }
	// 纯空间运动项：只把径向速度清零，几何仍为三维。这是本轮三维改动引入的那一项。
	// 实测约 14 ms（路径距离也随径向项变化后由 8.6 ms 升到 14 ms），阈值留约 1.8 倍余量，
	// 目的是抓口径回退，不是卡精度指标。
	if got := milliseconds(unlimited.Immersion, event(noRadial).Immersion); got > 25 {
		t.Fatalf("space-motion term shifts immersion by %.1f ms, want the 25 ms budget respected", got)
	}
	// 距离有无会额外启用站心视差修正，属于既有几何而非本轮改动，量级单独记录。
	fmt.Printf("空间运动项 %.1f ms；距离项 %.1f ms\n",
		milliseconds(unlimited.Immersion, event(noRadial).Immersion),
		milliseconds(unlimited.Immersion, limited.Immersion))
}

func TestStarCoordinateTwoAndThreeDimensionsAreDistinctGeometries(t *testing.T) {
	// 径向速度反号必须让三维结果分居两侧，且二维分支确实给出不同的角距，
	// 否则"三维对二维"的对照就是空保证。
	star := occStar3DTestCoordinate(224.366667, -21.415556)
	star.ProperMotionRACosDecMasPerYear = 1045
	star.ProperMotionDecMasPerYear = -1729
	star.ParallaxMas = 173
	star.RadialVelocityKmPerSecond = -100
	away := star
	away.RadialVelocityKmPerSecond = 100
	noDistance := star
	noDistance.ParallaxMas = 0

	tt := occultationTimeToTT(time.Date(2026, 1, 13, 0, 14, 7, 0, time.UTC))
	observer := Observer{Longitude: 97.471, Latitude: -45.720}
	approaching := starMoonSeparationArcsec(tt, star, observer)
	receding := starMoonSeparationArcsec(tt, away, observer)
	twoDimension := starMoonSeparationArcsec(tt, noDistance, observer)
	if math.Abs(approaching-receding) < 0.005 {
		t.Fatalf("opposite radial velocities separate by only %.6f arcsec, want a measurable space-motion term",
			math.Abs(approaching-receding))
	}
	if math.Abs(twoDimension-approaching) < 0.05 {
		t.Fatalf("2D separation %.6f arcsec is indistinguishable from the 3D one %.6f", twoDimension, approaching)
	}
	fmt.Printf("同一时刻角距：三维 rv=-100 为 %.4f\"，二维为 %.4f\"，三维 rv=+100 为 %.4f\"\n",
		approaching, twoDimension, receding)
}

func TestStarCoordinatePathDistanceAcceptsLightYears(t *testing.T) {
	// 光年与等价视差必须进入同一套路径几何，不能一个当无穷远。
	byParallax := occStar3DTestCoordinate(224.366667, -21.415556)
	byParallax.ParallaxMas = 173
	byLightYear := byParallax
	byLightYear.ParallaxMas = 0
	byLightYear.DistanceLightYear = (1000.0 / 173) * tools.AstronomicalUnitKilometers * (648000 / math.Pi) / tools.LightYearKilometers

	tt := occultationTimeToTT(time.Date(2026, 1, 13, 0, 0, 0, 0, time.UTC))
	fromParallax := starOccultationEphemerisStateAt(tt, byParallax)
	fromLightYear := starOccultationEphemerisStateAt(tt, byLightYear)
	if fromParallax.starDistanceKM <= 0 {
		t.Fatal("parallax input should yield a finite path distance")
	}
	// 两种等价输入只允许差一个浮点往返的量级。
	if relative := math.Abs(fromLightYear.starDistanceKM-fromParallax.starDistanceKM) / fromParallax.starDistanceKM; relative > 1e-12 {
		t.Fatalf("path distance = %g km from light-years, want %g km from the equivalent parallax (relative %.3g)",
			fromLightYear.starDistanceKM, fromParallax.starDistanceKM, relative)
	}
	kept := newStarOccultationLocalEphemeris(tt, byLightYear)
	if relative := math.Abs(kept.starDistanceKM()-fromParallax.starDistanceKM) / fromParallax.starDistanceKM; relative > 1e-12 {
		t.Fatalf("local ephemeris distance = %g km, want %g km (relative %.3g)", kept.starDistanceKM(), fromParallax.starDistanceKM, relative)
	}
}

func TestStarCoordinatePathGeometryUsesPropagatedDistance(t *testing.T) {
	// 路径与本地星历必须用当日的距离，而不是历元距离。1 pc、径向 -100 km/s，1000 年后
	// 距离缩短约 10%，两种口径给出的路径距离必须能区分开。
	star := occStar3DTestCoordinate(189.1975, -5.831944444444)
	star.ParallaxMas = 1000
	star.RadialVelocityKmPerSecond = -100
	epochDistanceKM := 206264806.247 / star.ParallaxMas * occultationPathAstronomicalUnitKM

	for _, years := range []float64{26, 1000, 3000} {
		tt := occultationTimeToTT(star.Epoch) + years*365.25
		state := starOccultationEphemerisStateAt(tt, star)
		if !state.valid {
			t.Fatalf("ephemeris state invalid at %.0f years", years)
		}
		_, _, distanceAU := starApparentRaDecDistanceGeocentric(tt, star)
		want := distanceAU * occultationPathAstronomicalUnitKM
		if math.Abs(state.starDistanceKM-want) > want*1e-12 {
			t.Fatalf("path distance at %.0f years = %g km, want the propagated %g km", years, state.starDistanceKM, want)
		}
		if years >= 1000 && math.Abs(state.starDistanceKM-epochDistanceKM) < want*0.01 {
			t.Fatalf("path distance at %.0f years matches the epoch distance, so the radial term is still dropped", years)
		}
		if got := newStarOccultationLocalEphemeris(tt, star).starDistanceKM(); math.Abs(got-want) > want*1e-12 {
			t.Fatalf("local ephemeris distance at %.0f years = %g km, want %g km", years, got, want)
		}
	}

	// 径向速度为零时两种口径必须重合，否则说明引入了与运动无关的偏移。
	star.RadialVelocityKmPerSecond = 0
	still := starOccultationEphemerisStateAt(occultationTimeToTT(star.Epoch)+1000*365.25, star)
	if math.Abs(still.starDistanceKM-epochDistanceKM) > epochDistanceKM*1e-9 {
		t.Fatalf("path distance without radial motion = %g km, want the epoch %g km", still.starDistanceKM, epochDistanceKM)
	}
}

// TestStarCoordinateOptimizedEphemerisKeepsFiniteDistance 固定优化分支（密集星历）也带当日距离：
// 节点按每个采样时刻的推进距离装配，状态回读若丢掉它，有限距离会悄悄退化成无穷远框架，
// 而且 exact 与 optimized 两条分支会在同一颗带视差恒星上给出不同的帧。
func TestStarCoordinateOptimizedEphemerisKeepsFiniteDistance(t *testing.T) {
	star := occStar3DTestCoordinate(189.1975, -5.831944444444)
	star.ProperMotionRACosDecMasPerYear = -28
	star.ProperMotionDecMasPerYear = -18
	star.ParallaxMas = 768.5
	star.RadialVelocityKmPerSecond = -30

	cst := time.FixedZone("CST", 8*3600)
	start := time.Date(2025, 6, 5, 0, 0, 0, 0, cst)
	center := occultationTimeToTT(start.Add(12 * time.Hour))
	_, _, distanceAU := starApparentRaDecDistanceGeocentric(center, star)
	want := distanceAU * occultationPathAstronomicalUnitKM

	cache := newStarOccultationEventCache(star)
	cache.preparePathEphemeris(center, OccultationPathAlgorithmOptimized)
	if cache.local == nil || !cache.local.dense {
		t.Fatalf("dense ephemeris not selected, the optimized branch is not exercised")
	}
	if got := cache.local.starDistanceKM(); math.Abs(got-want) > want*1e-9 {
		t.Fatalf("optimized ephemeris distance = %g km, want the propagated %g km", got, want)
	}
	state, ok := cache.local.stateAt(center)
	if !ok {
		t.Fatal("optimized state unavailable")
	}
	if math.Abs(state.starDistanceKM-want) > want*1e-6 {
		t.Fatalf("optimized state distance = %g km, want the propagated %g km", state.starDistanceKM, want)
	}
	denseFrame, ok := starOccultationPathFrameFromState(state)
	if !ok {
		t.Fatal("optimized frame unavailable")
	}
	exactFrame, ok := starOccultationPathFrameFromState(starOccultationEphemerisStateAt(center, star))
	if !ok {
		t.Fatal("exact frame unavailable")
	}
	if angle := occultationPathNorm(occultationPathCross(denseFrame.axis, exactFrame.axis)) * 180 / math.Pi * 3600 * 1000; angle > 1e-6 {
		t.Fatalf("optimized and exact frame axes differ by %.3e mas, want only the dense table's own error", angle)
	}

	// 距离未知时仍按 0 上报，几何保持无穷远，不要退化成"1 km 处"的假视差。
	unknown := star
	unknown.ParallaxMas = 0
	unknownCache := newStarOccultationEventCache(unknown)
	unknownCache.preparePathEphemeris(center, OccultationPathAlgorithmOptimized)
	if unknownCache.local == nil {
		t.Fatal("local ephemeris missing for the distance-free star")
	}
	if got := unknownCache.local.starDistanceKM(); got != 0 {
		t.Fatalf("distance-free ephemeris distance = %g, want 0", got)
	}
	if state, ok := unknownCache.local.stateAt(center); !ok || state.starDistanceKM != 0 {
		t.Fatalf("distance-free state distance = %g (ok=%v), want 0", state.starDistanceKM, ok)
	}

	// 路径层：两条分支必须给出同一条路径。
	options := OccultationPathOptions{Step: 5 * time.Minute, TargetSpacingKM: 200}
	end := start.Add(24 * time.Hour)
	exactOptions, optimizedOptions := options, options
	exactOptions.Algorithm = OccultationPathAlgorithmExact
	optimizedOptions.Algorithm = OccultationPathAlgorithmOptimized
	exactPaths, err := FindStarOccultationPaths(start, end, star, exactOptions)
	if err != nil || len(exactPaths) == 0 {
		t.Fatalf("exact path: %v (paths=%d)", err, len(exactPaths))
	}
	optimizedPaths, err := FindStarOccultationPaths(start, end, star, optimizedOptions)
	if err != nil || len(optimizedPaths) == 0 {
		t.Fatalf("optimized path: %v (paths=%d)", err, len(optimizedPaths))
	}
	exact, optimized := exactPaths[0], optimizedPaths[0]
	for _, moment := range []struct {
		name           string
		exact, optimiz time.Time
	}{
		{"start", exact.Start.Time, optimized.Start.Time},
		{"greatest", exact.Greatest.Time, optimized.Greatest.Time},
		{"end", exact.End.Time, optimized.End.Time},
	} {
		if delta := math.Abs(moment.exact.Sub(moment.optimiz).Seconds()); delta > 1e-3 {
			t.Fatalf("%s differs by %.6f s between the two ephemeris branches", moment.name, delta)
		}
	}
	if delta := math.Abs(exact.Greatest.WidthKM - optimized.Greatest.WidthKM); delta > 1e-3 {
		t.Fatalf("greatest width differs by %.6f km between the two ephemeris branches", delta)
	}
}
