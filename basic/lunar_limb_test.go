package basic

import (
	"math"
	"testing"
	"time"

	"github.com/starainrt/astro/internal/geodata"
	"github.com/starainrt/astro/moon/limb"
)

// recordingLunarLimb 记录最后一次查询参数，返回固定半径。
type recordingLunarLimb struct {
	lon, lat, angle float64
	radiusKM        float64
}

func (limb *recordingLunarLimb) LimbRadiusKM(lon, lat, angle float64) (float64, bool) {
	limb.lon, limb.lat, limb.angle = lon, lat, angle
	return limb.radiusKM, true
}

func lunarLimbAngleDiff(a, b float64) float64 {
	d := math.Mod(a-b, 360)
	if d > 180 {
		d -= 360
	}
	if d < -180 {
		d += 360
	}
	return d
}

// 地心观测者所见的月面经纬必须还原该时刻的总天平动；目标位于月球北极投影方向时位置角为 0，
// 向天东转 90° 时为 90°。
func TestLunarLimbFrameRoundTrip(t *testing.T) {
	for _, jde := range []float64{2460409.25, 2451545.0, 2462000.7} {
		_, moon := solarEclipseSunMoonEquatorial(jde)
		moonXYZ := solarEclipseLLRToXYZ(moon[0], moon[1], moon[2])
		limb := &recordingLunarLimb{radiusKM: 1737.4}
		frame := newLunarLimbFrame(limb, jde, moonXYZ)
		if !frame.valid {
			t.Fatalf("jde %v: frame invalid", jde)
		}
		physical := MoonPhysical(jde)
		moonUnit, _ := lunarLimbNormalize(moonXYZ)
		north, east, _ := lunarLimbSkyBasis(moonUnit)

		for _, offset := range []float64{0, 90, 200} {
			pa := (physical.PositionAngle + offset) * rad
			// 目标放在月心旁 0.5° 处、位置角为 P + offset。
			sky := lunarLimbCombine(math.Cos(pa), north, math.Sin(pa), east)
			target := lunarLimbCombine(math.Cos(0.5*rad), moonUnit, math.Sin(0.5*rad), sky)
			target = [3]float64{target[0] * 1.5e8, target[1] * 1.5e8, target[2] * 1.5e8}
			if _, ok := frame.radiusKM(moonXYZ, target, false); !ok {
				t.Fatalf("jde %v: radiusKM failed", jde)
			}
			if d := lunarLimbAngleDiff(limb.lon, physical.LibrationLongitude); math.Abs(d) > 1e-6 {
				t.Errorf("jde %v: libration longitude %.9f, want %.9f", jde, limb.lon, physical.LibrationLongitude)
			}
			if d := limb.lat - physical.LibrationLatitude; math.Abs(d) > 1e-6 {
				t.Errorf("jde %v: libration latitude %.9f, want %.9f", jde, limb.lat, physical.LibrationLatitude)
			}
			if d := lunarLimbAngleDiff(limb.angle, -offset); math.Abs(d) > 1e-6 {
				t.Errorf("jde %v offset %v: position angle %.9f", jde, offset, limb.angle)
			}
			if _, ok := frame.radiusKM(moonXYZ, target, true); !ok {
				t.Fatalf("jde %v: away radiusKM failed", jde)
			}
			if d := lunarLimbAngleDiff(limb.angle, 180-offset); math.Abs(d) > 1e-6 {
				t.Errorf("jde %v offset %v: away position angle %.9f", jde, offset, limb.angle)
			}
		}
	}
}

// 参考球半径等于 k·R⊕ 的常数月缘只能让结果在末位附近变化，用来确认替换路径本身不引入偏差。
func TestLocalSolarEclipseConstantLimbMatchesCircular(t *testing.T) {
	const seed, lon, lat = 2460409.25, -96.8, 32.8 // 2024-04-08，达拉斯附近
	options := SolarEclipseOptions{RadiusModel: SolarEclipseModelIAUSingleK}
	circular := LocalSolarEclipseWithOptionsWithInstance(seed, lon, lat, 0, options, nil)
	if !circular.HasTotal {
		t.Fatalf("expected total eclipse, got %v", circular.Type)
	}
	options.LunarLimb = &recordingLunarLimb{
		radiusKM: geodata.EarthEquatorialRadiusKM * solarEclipseIAUSingleRadiusK * localSolarMoonRadiusScale,
	}
	limbed := LocalSolarEclipseWithOptionsWithInstance(seed, lon, lat, 0, options, nil)
	for _, pair := range [][2]float64{
		{circular.PartialStart, limbed.PartialStart},
		{circular.CentralStart, limbed.CentralStart},
		{circular.CentralEnd, limbed.CentralEnd},
		{circular.PartialEnd, limbed.PartialEnd},
	} {
		if d := math.Abs(pair[0]-pair[1]) * 86400; d > 1e-3 {
			t.Errorf("contact moved by %.6f s with a constant limb", d)
		}
	}
}

func TestLunarEclipseConstantLimbMatchesCircular(t *testing.T) {
	limb := &recordingLunarLimb{radiusKM: lunarMoonRadiusScale}
	for _, chauvenet := range []bool{false, true} {
		circular := LunarEclipseDanjonWithInstance(2460926.0, nil)
		if chauvenet {
			circular = LunarEclipseChauvenetWithInstance(2460926.0, nil)
		}
		limbed := LunarEclipseWithLunarLimb(circular, limb, nil)
		for _, pair := range [][2]float64{
			{circular.PenumbralStart, limbed.PenumbralStart},
			{circular.PartialStart, limbed.PartialStart},
			{circular.TotalStart, limbed.TotalStart},
			{circular.TotalEnd, limbed.TotalEnd},
			{circular.PartialEnd, limbed.PartialEnd},
			{circular.PenumbralEnd, limbed.PenumbralEnd},
		} {
			if d := math.Abs(pair[0]-pair[1]) * 86400; d > 0.05 {
				t.Errorf("chauvenet=%v: contact moved by %.4f s with a constant limb", chauvenet, d)
			}
		}
		if limbed.Maximum != circular.Maximum || limbed.Magnitude != circular.Magnitude {
			t.Error("maximum or magnitude changed")
		}
	}
	if got := LunarEclipseWithLunarLimb(LunarEclipseDanjonWithInstance(2460926.0, nil), nil, nil); got != LunarEclipseDanjonWithInstance(2460926.0, nil) {
		t.Error("nil limb must return the input unchanged")
	}
}

func TestOccultationConstantLimbMatchesCircular(t *testing.T) {
	limb := &recordingLunarLimb{radiusKM: moonEquatorialRadiusKM}
	start := time.Date(2024, time.August, 21, 1, 30, 0, 0, time.UTC)
	end := time.Date(2024, time.August, 21, 4, 0, 0, 0, time.UTC)
	observer := Observer{Longitude: -30.072, Latitude: 16.21}
	plain, err := FindPlanetOccultationsWithInstance(start, end, OccultationSaturn,
		observer.Longitude, observer.Latitude, 0, OccultationSearchOptions{}, nil)
	if err != nil || len(plain) != 1 {
		t.Fatalf("circular planet search: %v, %d events", err, len(plain))
	}
	limbed, err := FindPlanetOccultationsWithInstance(start, end, OccultationSaturn,
		observer.Longitude, observer.Latitude, 0, OccultationSearchOptions{LunarLimb: limb}, nil)
	if err != nil || len(limbed) != 1 {
		t.Fatalf("limb planet search: %v, %d events", err, len(limbed))
	}
	a, b := plain[0], limbed[0]
	for i, pair := range [][2]time.Time{
		{a.ExternalImmersion, b.ExternalImmersion}, {a.InternalImmersion, b.InternalImmersion},
		{a.InternalEmersion, b.InternalEmersion}, {a.ExternalEmersion, b.ExternalEmersion},
	} {
		if d := math.Abs(pair[0].Sub(pair[1]).Seconds()); d > 0.01 {
			t.Errorf("planet contact %d moved by %.4f s with a constant limb", i, d)
		}
	}
	if !a.Greatest.Equal(b.Greatest) || a.Type != b.Type {
		t.Error("greatest or type changed")
	}
}

func TestGlobalSolarEclipseConstantLimbMatchesCircular(t *testing.T) {
	// IAU Single-K 的内外 k 相同，常数月缘可同时复现外切与内切半径。
	limb := &recordingLunarLimb{radiusKM: geodata.EarthEquatorialRadiusKM * solarEclipseIAUSingleRadiusK * localSolarMoonRadiusScale}
	options := SolarEclipseOptions{RadiusModel: SolarEclipseModelIAUSingleK}
	for _, seed := range []float64{2460409.0, 2460232.0} {
		circular := SolarEclipseWithOptionsWithInstance(seed, options, nil)
		withLimb := options
		withLimb.LunarLimb = limb
		limbed := SolarEclipseWithOptionsWithInstance(seed, withLimb, nil)
		for i, pair := range [][2]float64{
			{circular.PartialBeginOnEarth, limbed.PartialBeginOnEarth},
			{circular.PartialEndOnEarth, limbed.PartialEndOnEarth},
		} {
			if d := math.Abs(pair[0]-pair[1]) * 86400; d > 0.05 {
				t.Errorf("seed %.0f contact %d moved by %.4f s with a constant limb", seed, i, d)
			}
		}
		if limbed.GreatestEclipse != circular.GreatestEclipse || limbed.Type != circular.Type ||
			limbed.CentralBeginOnEarth != circular.CentralBeginOnEarth || limbed.PathWidthKM != circular.PathWidthKM {
			t.Errorf("seed %.0f: non-contact fields changed", seed)
		}
		if d := math.Abs(limbed.CentralDurationDays-circular.CentralDurationDays) * 86400; d > 0.05 {
			t.Errorf("seed %.0f: central duration moved by %.4f s with a constant limb", seed, d)
		}
	}
}

func TestCentralPathConstantLimbKeepsLimits(t *testing.T) {
	limb := &recordingLunarLimb{radiusKM: geodata.EarthEquatorialRadiusKM * solarEclipseUmbralK * localSolarMoonRadiusScale}
	for _, seed := range []float64{2460409.0, 2460232.0} {
		circular := SolarEclipseCentralPathNASABulletinSplitKWithInstance(seed, SolarEclipsePathOptions{}, nil)
		limbed := SolarEclipseCentralPathNASABulletinSplitKWithInstance(seed, SolarEclipsePathOptions{LunarLimb: limb}, nil)
		if len(circular.NorthernLimit) == 0 || len(limbed.NorthernLimit) != len(circular.NorthernLimit) ||
			len(limbed.SouthernLimit) != len(circular.SouthernLimit) {
			t.Fatalf("seed %.0f: limit lengths %d/%d vs %d/%d", seed,
				len(limbed.NorthernLimit), len(limbed.SouthernLimit), len(circular.NorthernLimit), len(circular.SouthernLimit))
		}
		worst := 0.0
		for _, pair := range [][2][]SolarEclipsePathPoint{{circular.NorthernLimit, limbed.NorthernLimit}, {circular.SouthernLimit, limbed.SouthernLimit}} {
			for i := range pair[0] {
				d := math.Hypot((pair[0][i].Longitude-pair[1][i].Longitude)*math.Cos(pair[0][i].Latitude*rad), pair[0][i].Latitude-pair[1][i].Latitude) * 111.2
				worst = math.Max(worst, d)
			}
		}
		if worst > 0.05 {
			t.Errorf("seed %.0f: constant limb moved a limit by %.3f km", seed, worst)
		}
	}
}

func TestOccultationPathConstantLimbKeepsLimits(t *testing.T) {
	limb := &recordingLunarLimb{radiusKM: moonEquatorialRadiusKM}
	star := StarCoordinate{
		ID: "HR 4799", RA: 189.1975, Dec: -5.831944444444,
		Epoch: time.Date(2000, 1, 1, 12, 0, 0, 0, time.UTC), Frame: CoordinateFrameJ2000,
		ProperMotionRACosDecMasPerYear: -28, ProperMotionDecMasPerYear: -18,
	}
	start := time.Date(2025, 6, 4, 16, 0, 0, 0, time.UTC)
	end := time.Date(2025, 6, 5, 16, 0, 0, 0, time.UTC)
	options := OccultationPathOptions{Step: 5 * time.Minute, TargetSpacingKM: 200, DisableFootprints: true, DisableRiseSet: true}
	circular, err := FindStarOccultationPathsWithInstance(start, end, star, options, nil)
	if err != nil || len(circular) != 1 {
		t.Fatalf("circular: %v %d", err, len(circular))
	}
	options.LunarLimb = limb
	limbed, err := FindStarOccultationPathsWithInstance(start, end, star, options, nil)
	if err != nil || len(limbed) != 1 {
		t.Fatalf("limbed: %v %d", err, len(limbed))
	}
	worst := 0.0
	for _, pair := range [][2][]OccultationPathPoint{{circular[0].NorthernLimit, limbed[0].NorthernLimit}, {circular[0].SouthernLimit, limbed[0].SouthernLimit}} {
		if len(pair[0]) == 0 || len(pair[0]) != len(pair[1]) {
			t.Fatalf("limit lengths %d vs %d", len(pair[0]), len(pair[1]))
		}
		for i := range pair[0] {
			d := math.Hypot((pair[0][i].Longitude-pair[1][i].Longitude)*math.Cos(pair[0][i].Latitude*rad), pair[0][i].Latitude-pair[1][i].Latitude) * 111.2
			worst = math.Max(worst, d)
		}
	}
	if worst > 0.05 {
		t.Errorf("constant limb moved a graze limit by %.3f km", worst)
	}
	if limbed[0].LunarLimbApplied {
		t.Fatal("constant circular limb set LunarLimbApplied = true")
	}
}

func TestCentralPathConstantLimbKeepsWidthsAndBand(t *testing.T) {
	limb := &recordingLunarLimb{radiusKM: geodata.EarthEquatorialRadiusKM * solarEclipseUmbralK * localSolarMoonRadiusScale}
	for _, seed := range []float64{2460409.0, 2460232.0} {
		circular := SolarEclipseCentralPathNASABulletinSplitKWithInstance(seed, SolarEclipsePathOptions{}, nil)
		limbed := SolarEclipseCentralPathNASABulletinSplitKWithInstance(seed, SolarEclipsePathOptions{LunarLimb: limb}, nil)
		if len(limbed.CenterLine) != len(circular.CenterLine) || len(limbed.CentralBandSegments) != len(circular.CentralBandSegments) {
			t.Fatalf("seed %.0f: sample counts changed", seed)
		}
		worstWidth := 0.0
		for i := range circular.CenterLine {
			worstWidth = math.Max(worstWidth, math.Abs(limbed.CenterLine[i].WidthKM-circular.CenterLine[i].WidthKM))
		}
		if worstWidth > 0.1 {
			t.Errorf("seed %.0f: constant limb moved a width by %.3f km", seed, worstWidth)
		}
		if d := math.Abs(limbed.Greatest.WidthKM - circular.Greatest.WidthKM); d > 0.1 {
			t.Errorf("seed %.0f: greatest width moved by %.3f km", seed, d)
		}
		if d := math.Abs(limbed.MaxCentralDurationDays-circular.MaxCentralDurationDays) * 86400; d > 0.05 {
			t.Errorf("seed %.0f: max duration moved by %.4f s", seed, d)
		}
		worstBand := 0.0
		for s := range circular.CentralBandSegments {
			if len(limbed.CentralBandSegments[s]) != len(circular.CentralBandSegments[s]) {
				t.Fatalf("seed %.0f: band segment %d length changed", seed, s)
			}
			for i, a := range circular.CentralBandSegments[s] {
				b := limbed.CentralBandSegments[s][i]
				d := math.Hypot((a.Longitude-b.Longitude)*math.Cos(a.Latitude*rad), a.Latitude-b.Latitude) * 111.2
				worstBand = math.Max(worstBand, d)
			}
		}
		if worstBand > 0.1 {
			t.Errorf("seed %.0f: constant limb moved a band vertex by %.3f km", seed, worstBand)
		}
	}
}

// failingLunarLimb 处处无值：月缘修正必须整体退回圆月面且不置 LunarLimbApplied。
type failingLunarLimb struct{}

func (failingLunarLimb) LimbRadiusKM(_, _, _ float64) (float64, bool) { return 0, false }

func TestFailingLimbKeepsEclipsesCircular(t *testing.T) {
	const seed, lon, lat = 2460409.25, -96.8, 32.8 // 2024-04-08，达拉斯附近
	limb := failingLunarLimb{}
	for _, model := range []SolarEclipseRadiusModel{SolarEclipseModelNASABulletinSplitK, SolarEclipseModelIAUSingleK} {
		options := SolarEclipseOptions{RadiusModel: model}
		withLimb := options
		withLimb.LunarLimb = limb

		circularLocal := LocalSolarEclipseWithOptionsWithInstance(seed, lon, lat, 0, options, nil)
		limbedLocal := LocalSolarEclipseWithOptionsWithInstance(seed, lon, lat, 0, withLimb, nil)
		if limbedLocal != circularLocal {
			t.Errorf("model %v: failing limb changed the local result", model)
		}

		circularGlobal := SolarEclipseWithOptionsWithInstance(seed, options, nil)
		limbedGlobal := SolarEclipseWithOptionsWithInstance(seed, withLimb, nil)
		if limbedGlobal.LunarLimbApplied ||
			limbedGlobal.PartialBeginOnEarth != circularGlobal.PartialBeginOnEarth ||
			limbedGlobal.PartialEndOnEarth != circularGlobal.PartialEndOnEarth ||
			limbedGlobal.CentralDurationDays != circularGlobal.CentralDurationDays {
			t.Errorf("model %v: failing limb changed the global result or set LunarLimbApplied", model)
		}
	}

	circularLunar := LunarEclipseDanjonWithInstance(2460926.0, nil)
	if got := LunarEclipseWithLunarLimb(circularLunar, limb, nil); got != circularLunar {
		t.Error("failing limb changed the lunar eclipse result")
	}
}

func TestPartialBoundaryLimbCorrection(t *testing.T) {
	const seed = 2460409.25 // 2024-04-08
	circularOptions := SolarEclipsePartialFootprintOptions{}
	circular := solarEclipsePartialFootprints(seed, SolarEclipseModelIAUSingleK, circularOptions, nil)
	if len(circular.PartialBandContours) == 0 || len(circular.RiseSetCurves) == 0 {
		t.Fatal("circular partial boundary missing")
	}
	samePoints := func(a, b []SolarEclipsePathPoint, toleranceKM float64) float64 {
		worst := 0.0
		for index := range a {
			km := math.Hypot(signedAngleDifference(a[index].Longitude, b[index].Longitude)*111.19*math.Cos(a[index].Latitude*rad),
				(a[index].Latitude-b[index].Latitude)*111.19)
			worst = math.Max(worst, km)
		}
		return worst
	}
	compare := func(name string, limb LunarLimb, wantApplied bool, maxShiftKM float64) {
		options := circularOptions
		options.LunarLimb = limb
		limbed := solarEclipsePartialFootprints(seed, SolarEclipseModelIAUSingleK, options, nil)
		if limbed.PartialBoundaryLimbApplied != wantApplied {
			t.Errorf("%s: PartialBoundaryLimbApplied = %v, want %v", name, limbed.PartialBoundaryLimbApplied, wantApplied)
		}
		if limbed.RiseSetTopologyDegraded {
			t.Errorf("%s: rise/set topology degraded", name)
		}
		if len(limbed.PartialBandContours) != len(circular.PartialBandContours) || len(limbed.RiseSetCurves) != len(circular.RiseSetCurves) {
			t.Fatalf("%s: boundary structure changed", name)
		}
		worst := 0.0
		for index := range circular.PartialBandContours {
			worst = math.Max(worst, samePoints(circular.PartialBandContours[index], limbed.PartialBandContours[index], 0))
		}
		for index, curve := range circular.RiseSetCurves {
			for segment := range curve.Segments {
				shift := samePoints(curve.Segments[segment], limbed.RiseSetCurves[index].Segments[segment], 0)
				if curve.Phase == RiseSetPhaseGreatest && shift != 0 {
					t.Errorf("%s: greatest rise/set curve moved by %.3f km", name, shift)
				}
				worst = math.Max(worst, shift)
			}
		}
		if worst > maxShiftKM {
			t.Errorf("%s: boundary moved by %.3f km, want <= %.3f km", name, worst, maxShiftKM)
		}
		if wantApplied && worst < 0.5 {
			t.Errorf("%s: boundary moved only %.3f km", name, worst)
		}
	}
	compare("failing", failingLunarLimb{}, false, 0)
	compare("constant", &recordingLunarLimb{
		radiusKM: geodata.EarthEquatorialRadiusKM * solarEclipseIAUSingleRadiusK * localSolarMoonRadiusScale,
	}, false, 0.05)
	compare("profile", sawtoothLunarLimb{}, true, limbLimitMaxShiftKM)
}

// sawtoothLunarLimb 以位置角 10° 为周期在参考半径上叠加 ±3 km 起伏，模拟月缘山谷。
type sawtoothLunarLimb struct{}

func (sawtoothLunarLimb) LimbRadiusKM(_, _, angle float64) (float64, bool) {
	phase := math.Mod(angle, 10) / 10
	return geodata.EarthEquatorialRadiusKM*solarEclipseIAUSingleRadiusK*localSolarMoonRadiusScale + 6*math.Abs(phase-0.5) - 1.5, true
}

// TestGlobalSolarEclipseLimbContactMatchesTime 检查月缘修正后的偏食始末接触点与修正时刻同源：
// 接触点必须落在该时刻重解的半影外切点上，而不是沿地球边缘扫描网格上的样点。
func TestGlobalSolarEclipseLimbContactMatchesTime(t *testing.T) {
	for _, seed := range []float64{2460409.25, 2460232.25, 2461275.25} {
		options := SolarEclipseOptions{RadiusModel: SolarEclipseModelNASABulletinSplitK, LunarLimb: limb.Default()}
		result := SolarEclipseWithOptionsWithInstance(seed, options, nil)
		if !result.LunarLimbApplied {
			t.Fatalf("seed %.2f: the default limb profile changed nothing", seed)
		}
		solver := newSolarEclipseSolverWithOptions(CalcMoonSHByJDE(seed, 0), options, nil)
		for _, pair := range []struct {
			name    string
			jde     float64
			contact SolarEclipseLimbContact
		}{
			{"P1", result.PartialBeginOnEarth, result.PartialBeginContact},
			{"P4", result.PartialEndOnEarth, result.PartialEndContact},
		} {
			if !pair.contact.Defined {
				t.Errorf("seed %.2f %s: the corrected contact location is missing", seed, pair.name)
				continue
			}
			point, ok := solver.shadowContactPointAt(pair.jde, solarEclipsePenumbralShadow, false)
			if !ok {
				t.Errorf("seed %.2f %s: no penumbral tangent point at the corrected time", seed, pair.name)
				continue
			}
			longitude := signedAngleDifference(pair.contact.Longitude, point.Longitude) * math.Cos(point.Latitude*rad)
			if d := math.Hypot(longitude, pair.contact.Latitude-point.Latitude) * geodata.EarthKMPerDegree; d > 0.05 {
				t.Errorf("seed %.2f %s: contact location is %.3f km from the tangent point at its own time", seed, pair.name, d)
			}
		}
	}
}

// TestLimbFlagsIgnoreNoiseWithConstantLimb 检查常数月缘等价于圆月面时，各层月缘标志都不因离散残差误报。
func TestLimbFlagsIgnoreNoiseWithConstantLimb(t *testing.T) {
	iau := &recordingLunarLimb{radiusKM: geodata.EarthEquatorialRadiusKM * solarEclipseIAUSingleRadiusK * localSolarMoonRadiusScale}
	umbral := &recordingLunarLimb{radiusKM: geodata.EarthEquatorialRadiusKM * solarEclipseUmbralK * localSolarMoonRadiusScale}
	for _, seed := range []float64{2460409.25, 2460232.25, 2461275.25} {
		global := SolarEclipseWithOptionsWithInstance(seed, SolarEclipseOptions{RadiusModel: SolarEclipseModelIAUSingleK, LunarLimb: iau}, nil)
		if global.LunarLimbApplied || global.PartialBeginContact.Defined || global.PartialEndContact.Defined {
			t.Errorf("seed %.2f: constant limb reported a global correction", seed)
		}
		central := SolarEclipseCentralPathNASABulletinSplitKWithInstance(seed, SolarEclipsePathOptions{LunarLimb: umbral}, nil)
		if central.CentralLimitsLimbApplied {
			t.Errorf("seed %.2f: constant limb reported moved central limits", seed)
		}
		local := LocalSolarEclipseWithOptionsWithInstance(seed, -96.8, 32.8, 0, SolarEclipseOptions{RadiusModel: SolarEclipseModelIAUSingleK, LunarLimb: iau}, nil)
		if local.LunarLimbApplied {
			t.Errorf("seed %.2f: constant limb reported a local correction", seed)
		}
	}
	real := SolarEclipseCentralPathNASABulletinSplitKWithInstance(2460409.25, SolarEclipsePathOptions{LunarLimb: limb.Default()}, nil)
	if !real.CentralLimitsLimbApplied {
		t.Error("default limb profile must move the central limits")
	}
}

// TestLocalSolarEclipseLimbFlagCoversMagnitude 检查只改变食分、遮蔽率而不改变接触时刻时站心标志仍为真。
func TestLocalSolarEclipseLimbFlagCoversMagnitude(t *testing.T) {
	circular := LocalSolarEclipseResult{Type: SolarEclipsePartial, PartialStart: 1, PartialEnd: 2, Magnitude: 0.5, Obscuration: 0.4}
	limbed := circular
	if localSolarEclipseLimbChanged(limbed, circular) {
		t.Fatal("identical results reported a change")
	}
	limbed.Magnitude += 1e-12
	if localSolarEclipseLimbChanged(limbed, circular) {
		t.Error("floating-point noise in magnitude reported a change")
	}
	limbed.Magnitude = circular.Magnitude + 1e-6
	if !localSolarEclipseLimbChanged(limbed, circular) {
		t.Error("a visible magnitude change was not reported")
	}
	limbed = circular
	limbed.Obscuration += 1e-6
	if !localSolarEclipseLimbChanged(limbed, circular) {
		t.Error("a visible obscuration change was not reported")
	}
	limbed = circular
	limbed.Type = SolarEclipseAnnular
	if !localSolarEclipseLimbChanged(limbed, circular) {
		t.Error("a type change was not reported")
	}
}

// steppedLunarLimb 报告指定网格步长的常半径剖面。
type steppedLunarLimb struct{ step float64 }

func (limb steppedLunarLimb) LimbRadiusKM(lon, lat, angle float64) (float64, bool) {
	return 1737.4, true
}
func (limb steppedLunarLimb) RadiusRangeKM() (float64, float64) { return 1737.4, 1737.4 }
func (limb steppedLunarLimb) AngleStepDeg() float64             { return limb.step }

func TestLunarLimbFrameRejectsTinyStep(t *testing.T) {
	_, moon := solarEclipseSunMoonEquatorial(2460409.25)
	moonXYZ := solarEclipseLLRToXYZ(moon[0], moon[1], moon[2])
	for _, c := range []struct{ step, want float64 }{
		{1e-300, lunarLimbDefaultStepDeg}, {1e-4, lunarLimbDefaultStepDeg}, {0.001, 0.001}, {0.5, 0.5}, {2, lunarLimbDefaultStepDeg},
	} {
		frame := newLunarLimbFrame(steppedLunarLimb{c.step}, 2460409.25, moonXYZ)
		if !frame.valid || frame.stepDeg != c.want {
			t.Errorf("step %g: got %g valid=%v, want %g", c.step, frame.stepDeg, frame.valid, c.want)
		}
	}
}
