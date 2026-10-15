package basic

// 恒星月掩按月缘剖面判定成掩：标准结果的掩始、掩终等于精密剖面的首次遮挡与最后复现，常数剖面退化为圆月面。

import (
	"math"
	"testing"
	"time"

	"github.com/starainrt/astro/moon/limb"
)

var limbTestAntares = StarCoordinate{ID: "Antares", RA: 247.3516666666667, Dec: -26.431944444444444,
	Epoch: time.Date(2000, 1, 1, 12, 0, 0, 0, time.UTC), Frame: CoordinateFrameJ2000,
	ProperMotionRACosDecMasPerYear: -10, ProperMotionDecMasPerYear: -20, ParallaxMas: 24}

func limbTestAntaresEvent(t *testing.T, lon, lat float64, profile LunarLimb) []StarOccultationInfo {
	t.Helper()
	start := time.Date(2026, 2, 10, 12, 0, 0, 0, time.UTC)
	events, err := FindStarOccultationsWithInstance(start, start.Add(36*time.Hour), limbTestAntares, lon, lat, 0,
		OccultationSearchOptions{LunarLimb: profile}, nil)
	if err != nil {
		t.Fatal(err)
	}
	return events
}

func TestStarOccultationLimbGrazeHasMultipleIntervals(t *testing.T) {
	const lon, lat = 41.5907, -55.07036
	if circular := limbTestAntaresEvent(t, lon, lat, nil); len(circular) != 0 {
		t.Fatalf("circular limb: %d events, want none just outside the circular limit", len(circular))
	}
	events := limbTestAntaresEvent(t, lon, lat, limb.Default())
	if len(events) != 1 || !events[0].LunarLimbApplied || events[0].Type != OccultationGrazing {
		t.Fatalf("limb events = %+v, want one limb-applied graze", events)
	}
	profile, err := StarOccultationLimbProfileWithInstance(events[0], limbTestAntares, limb.Default(), StarOccultationLimbProfileOptions{}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(profile.Hidden) < 2 {
		t.Fatalf("graze hidden intervals = %d, want several", len(profile.Hidden))
	}
	first, last := profile.Hidden[0], profile.Hidden[len(profile.Hidden)-1]
	if !first.Start.Equal(events[0].Immersion) || !last.End.Equal(events[0].Emersion) {
		t.Errorf("standard contacts %v–%v differ from profile %v–%v", events[0].Immersion, events[0].Emersion, first.Start, last.End)
	}
	for index, interval := range profile.Hidden {
		if !interval.End.After(interval.Start) || (index > 0 && !interval.Start.After(profile.Hidden[index-1].End)) {
			t.Fatalf("interval %d not ordered: %+v", index, profile.Hidden)
		}
	}
	for _, segment := range profile.Segments {
		for _, sample := range segment {
			inside := profile.Occultation.Immersion.Before(sample.Time) && sample.Time.Before(profile.Occultation.Emersion)
			if sample.Hidden && !inside {
				t.Fatalf("hidden sample %v outside %v–%v", sample.Time, profile.Occultation.Immersion, profile.Occultation.Emersion)
			}
		}
	}
}

func TestStarOccultationLimbCanRemoveCircularEvent(t *testing.T) {
	const lon, lat = 24.4934, -49.66243
	if circular := limbTestAntaresEvent(t, lon, lat, nil); len(circular) != 1 {
		t.Fatalf("circular limb: %d events, want one", len(circular))
	}
	if events := limbTestAntaresEvent(t, lon, lat, limb.Default()); len(events) != 0 {
		t.Fatalf("limb profile: %d events, want none (star passes a limb valley)", len(events))
	}
}

func TestStarOccultationConstantLimbMatchesCircular(t *testing.T) {
	constant := &recordingLunarLimb{radiusKM: moonEquatorialRadiusKM}
	for _, site := range [][2]float64{{41.5907, -55.2}, {24.4934, -50.5}} {
		circular := limbTestAntaresEvent(t, site[0], site[1], nil)
		limbed := limbTestAntaresEvent(t, site[0], site[1], constant)
		if len(circular) != 1 || len(limbed) != 1 {
			t.Fatalf("site %v: %d circular, %d constant-limb events", site, len(circular), len(limbed))
		}
		for _, pair := range [][2]time.Time{{circular[0].Immersion, limbed[0].Immersion}, {circular[0].Emersion, limbed[0].Emersion}} {
			if d := math.Abs(pair[0].Sub(pair[1]).Seconds()); d > 0.01 {
				t.Errorf("site %v: constant limb moved a contact by %.4f s", site, d)
			}
		}
	}
}

func TestStarOccultationLimbProfileRejectsBadInput(t *testing.T) {
	info := StarOccultationInfo{Greatest: time.Date(2026, 2, 11, 3, 47, 0, 0, time.UTC), Observer: Observer{Longitude: 41.5907, Latitude: -55.07}}
	if _, err := StarOccultationLimbProfileWithInstance(info, limbTestAntares, nil, StarOccultationLimbProfileOptions{}, nil); err == nil {
		t.Error("nil limb accepted")
	}
	if _, err := StarOccultationLimbProfileWithInstance(info, limbTestAntares, limb.Default(), StarOccultationLimbProfileOptions{Step: -time.Second}, nil); err == nil {
		t.Error("negative step accepted")
	}
	if _, err := StarOccultationLimbProfileWithInstance(StarOccultationInfo{Observer: info.Observer}, limbTestAntares, limb.Default(), StarOccultationLimbProfileOptions{}, nil); err == nil {
		t.Error("zero greatest accepted")
	}
}

func TestStarLimbEphemerisMatchesExact(t *testing.T) {
	observer := Observer{Longitude: 41.5907, Latitude: -55.07036}
	greatest := occultationTimeToTT(time.Date(2026, 2, 11, 3, 47, 20, 0, time.UTC))
	profile := limb.Default()
	ephemeris, ok := newStarLimbEphemeris(greatest, limbTestAntares, profile, nil)
	if !ok {
		t.Fatal("ephemeris failed")
	}
	for _, offsetMinutes := range []float64{-200, -61.3, -7.7, 0, 3.1, 44.4, 199} {
		tt := greatest + offsetMinutes/1440
		got, ok := ephemeris.stateAt(tt, observer)
		if !ok {
			t.Fatalf("%+.1f min: interpolation failed", offsetMinutes)
		}
		geoRA, geoDec := HMoonGeocentricApparentRaDecNWithInstance(tt, -1, nil)
		geoDistance := HMoonAwayNWithInstance(tt, -1, nil)
		topoRA, topoDec := moonTopocentricApparentRaDec(tt, observer, -1, nil)
		topoDistance := topocentricDistanceKM(geoRA, geoDec, geoDistance, observer, TT2UTC(tt))
		starRA, starDec := starApparentRaDec(tt, limbTestAntares, observer, nil)
		frame := newLunarLimbFrame(profile, tt, lunarEclipseLimbVector(geoRA, geoDec, geoDistance), nil)
		radiusKM, ok := frame.radiusKM(lunarEclipseLimbVector(topoRA, topoDec, topoDistance), lunarEclipseLimbVector(starRA, starDec, 1), false)
		if !ok {
			t.Fatal("exact limb failed")
		}
		separation := angularSeparationDegrees(topoRA, topoDec, starRA, starDec) * 3600
		limbRadius := angularSemidiameterArcsec(radiusKM, topoDistance)
		if d := math.Abs(got.separation - separation); d > 1e-3 {
			t.Errorf("%+.1f min: separation off by %.2g arcsec", offsetMinutes, d)
		}
		if d := math.Abs(got.limbRadius - limbRadius); d > 2e-3 {
			t.Errorf("%+.1f min: limb radius off by %.2g arcsec", offsetMinutes, d)
		}
	}
}

func TestStarOccultationLimbTypes(t *testing.T) {
	deep := limbTestAntaresEvent(t, 41.5907, -57.0, limb.Default())
	if len(deep) != 1 || deep[0].Type != OccultationTotal {
		t.Fatalf("deep occultation = %+v, want total", deep)
	}
	info := StarOccultationInfo{Greatest: time.Date(2026, 2, 11, 3, 47, 15, 0, time.UTC), Observer: Observer{Longitude: 41.5907, Latitude: -55.07036 + 0.4/111.2}}
	profile, err := StarOccultationLimbProfileWithInstance(info, limbTestAntares, limb.Default(), StarOccultationLimbProfileOptions{}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(profile.Hidden) == 0 || len(profile.Segments) != 1 || profile.Occultation.Type != OccultationGrazing {
		t.Fatalf("0.4 km outside the circular limit: %d intervals, %d segments, type %v; want a limb graze", len(profile.Hidden), len(profile.Segments), profile.Occultation.Type)
	}
	info.Observer.Latitude = -55.07036 + 0.8/111.2
	profile, err = StarOccultationLimbProfileWithInstance(info, limbTestAntares, limb.Default(), StarOccultationLimbProfileOptions{}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(profile.Hidden) != 0 || profile.Occultation.Type != OccultationNone || !profile.Occultation.Immersion.IsZero() || profile.Occultation.ContactsComplete {
		t.Fatalf("missed graze reported %d intervals, type %v", len(profile.Hidden), profile.Occultation.Type)
	}
}

func TestStarOccultationLimbProfileLargeStepKeepsGraze(t *testing.T) {
	events := limbTestAntaresEvent(t, 41.5907, -55.07036, limb.Default())
	if len(events) != 1 {
		t.Fatal("graze missing")
	}
	profile, err := StarOccultationLimbProfileWithInstance(events[0], limbTestAntares, limb.Default(), StarOccultationLimbProfileOptions{Step: time.Hour}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(profile.Hidden) == 0 || profile.Step >= time.Hour {
		t.Fatalf("hour step: %d intervals with effective step %v, want the graze kept", len(profile.Hidden), profile.Step)
	}
}

func TestStarLimbEphemerisFrameIsOrthonormal(t *testing.T) {
	greatest := occultationTimeToTT(time.Date(2026, 2, 11, 3, 47, 20, 0, time.UTC))
	ephemeris, ok := newStarLimbEphemeris(greatest, limbTestAntares, limb.Default(), nil)
	if !ok {
		t.Fatal("ephemeris failed")
	}
	for _, offsetMinutes := range []float64{-230, -97.3, 0, 41.1, 233} {
		axisX, axisZ, ok := interpolateLocalEphemerisVectors(ephemeris.axes, greatest+offsetMinutes/1440)
		if !ok {
			t.Fatal("interpolation failed")
		}
		axisZ, _ = lunarLimbNormalize(axisZ)
		axisX, _ = lunarLimbNormalize(lunarLimbReject(axisX, axisZ))
		if d := math.Abs(lunarLimbDot(axisX, axisZ)); d > 1e-12 {
			t.Errorf("%+.1f min: axes not orthogonal (%.2g)", offsetMinutes, d)
		}
	}
}

func TestPlanetOccultationLimbHandlesGraze(t *testing.T) {
	start := time.Date(2024, time.August, 21, 1, 30, 0, 0, time.UTC)
	end := time.Date(2024, time.August, 21, 4, 0, 0, 0, time.UTC)
	const lon, edge = -30.072, 43.562592
	find := func(lat float64, profile LunarLimb) []PlanetOccultationInfo {
		events, err := FindPlanetOccultationsWithInstance(start, end, OccultationSaturn, lon, lat, 0, OccultationSearchOptions{LunarLimb: profile}, nil)
		if err != nil {
			t.Fatal(err)
		}
		return events
	}
	if circular := find(edge, nil); len(circular) != 0 {
		t.Fatalf("circular edge = %+v, want no occultation just outside the circular graze line", circular)
	}
	limbed := find(edge, limb.Default())
	if len(limbed) != 1 || !limbed[0].LunarLimbApplied || limbed[0].Type == OccultationGrazing ||
		!limbed[0].ExternalEmersion.After(limbed[0].ExternalImmersion) {
		t.Fatalf("limb at circular graze line = %+v, want a partial occultation with positive duration", limbed)
	}
	if far := find(edge+6/111.2, limb.Default()); len(far) != 0 {
		t.Errorf("6 km outside the limit the limb still reports %d events", len(far))
	}
	constant := &recordingLunarLimb{radiusKM: moonEquatorialRadiusKM}
	for _, dkm := range []float64{-3, -1} {
		circular, withConstant := find(edge+dkm/111.2, nil), find(edge+dkm/111.2, constant)
		if len(circular) != 1 || len(withConstant) != 1 || circular[0].Type != withConstant[0].Type {
			t.Fatalf("%+.0f km: circular %v vs constant limb %v", dkm, circular, withConstant)
		}
		for _, pair := range [][2]time.Time{{circular[0].ExternalImmersion, withConstant[0].ExternalImmersion}, {circular[0].ExternalEmersion, withConstant[0].ExternalEmersion}} {
			if d := math.Abs(pair[0].Sub(pair[1]).Seconds()); d > 0.05 {
				t.Errorf("%+.0f km: constant limb moved a contact by %.3f s", dkm, d)
			}
		}
	}
}

func TestPlanetLimbEphemerisMatchesExact(t *testing.T) {
	config, _ := planetOccultationConfigFor(OccultationSaturn, nil)
	observer := Observer{Longitude: -30.072, Latitude: 16.21}
	greatest := occultationTimeToTT(time.Date(2024, time.August, 21, 2, 45, 0, 0, time.UTC))
	ephemeris, ok := newPlanetLimbEphemeris(greatest, config, limb.Default(), nil)
	if !ok {
		t.Fatal("ephemeris failed")
	}
	target := planetOccultationLimbTarget(config, observer)
	for _, offsetMinutes := range []float64{-200, -33.3, 0, 7.7, 199} {
		tt := greatest + offsetMinutes/1440
		state := planetOccultationStateAt(tt, config, &observer, -1, nil)
		for _, internal := range []bool{false, true} {
			got, ok := ephemeris.circularGap(tt, observer, internal)
			want := state.externalContactMetric
			if internal {
				want = state.internalContactMetric
			}
			if !ok || math.Abs(got-want) > 2e-3 {
				t.Errorf("%+.1f min internal=%v: circular gap %.4f\" want %.4f\"", offsetMinutes, internal, got, want)
			}
			gotLimb, okLimb := ephemeris.limbGap(tt, observer, internal)
			wantLimb, okExact := occultationLimbGap(tt, limb.Default(), observer, target, internal, nil)
			if !okLimb || !okExact || math.Abs(gotLimb-wantLimb) > 3e-3 {
				t.Errorf("%+.1f min internal=%v: limb gap %.4f\" want %.4f\"", offsetMinutes, internal, gotLimb, wantLimb)
			}
		}
	}
}

// occultationLimbGap 逐点精确求站心月缘接触间隙（角秒），作插值表的对拍基准。
func occultationLimbGap(tt float64, limb LunarLimb, observer Observer, target occultationLimbTarget, internal bool, inst *Instance) (float64, bool) {
	geoRA, geoDec := HMoonGeocentricApparentRaDecNWithInstance(tt, -1, inst)
	geoDistance := HMoonAwayNWithInstance(tt, -1, inst)
	topoRA, topoDec := moonTopocentricApparentRaDec(tt, observer, -1, inst)
	ut := TT2UTC(tt)
	topoDistance := topocentricDistanceKM(geoRA, geoDec, geoDistance, observer, ut)
	ra, dec, distance, radius, ok := target(tt)
	if !ok || !finite(topoRA) || !finite(topoDec) || !finite(topoDistance) || topoDistance <= 0 {
		return 0, false
	}
	if distance > 0 {
		ra, dec = TopocentricRaDec(ra, dec, observer.Latitude, observer.Longitude, ut, distance/angularDiameterAstronomicalUnitKM, observer.Height)
		ra = normalizeRA(ra)
	}
	frame := newLunarLimbFrame(limb, tt, lunarEclipseLimbVector(geoRA, geoDec, geoDistance), inst)
	view, ok := frame.viewToward(lunarEclipseLimbVector(topoRA, topoDec, topoDistance), lunarEclipseLimbVector(ra, dec, 1))
	if !ok {
		return 0, false
	}
	contact := lunarLimbContact{
		frame:        &frame,
		view:         view,
		distanceKM:   topoDistance,
		separation:   angularSeparationDegrees(topoRA, topoDec, ra, dec) * rad,
		targetRadius: radius / angularDiameterArcsecPerRadian,
	}
	if internal {
		inner, ok := contact.innerRadius(true)
		if !ok {
			return 0, false
		}
		return (contact.separation - math.Abs(inner-contact.targetRadius)) * angularDiameterArcsecPerRadian, true
	}
	outer, ok := contact.outerRadius()
	if !ok {
		return 0, false
	}
	return (contact.separation - outer - contact.targetRadius) * angularDiameterArcsecPerRadian, true
}

// constantRangedLunarLimb 常数半径且报告零宽范围，起伏带退化成一点。
type constantRangedLunarLimb struct{ radiusKM float64 }

func (c constantRangedLunarLimb) LimbRadiusKM(lon, lat, angle float64) (float64, bool) {
	return c.radiusKM, true
}

func (c constantRangedLunarLimb) RadiusRangeKM() (float64, float64) { return c.radiusKM, c.radiusKM }

func (c constantRangedLunarLimb) AngleStepDeg() float64 { return 0.5 }

func TestStarOccultationLimbProfileZeroWidthRangeKeepsSamplesOrdered(t *testing.T) {
	profileLimb := constantRangedLunarLimb{radiusKM: moonEquatorialRadiusKM}
	circular := limbTestAntaresEvent(t, 41.5907, -57.0, nil)
	if len(circular) != 1 {
		t.Fatalf("circular events = %d, want one", len(circular))
	}
	profile, err := StarOccultationLimbProfileWithInstance(circular[0], limbTestAntares, profileLimb, StarOccultationLimbProfileOptions{}, nil)
	if err != nil {
		t.Fatal(err)
	}
	for index, segment := range profile.Segments {
		for k := 1; k < len(segment); k++ {
			if !segment[k].Time.After(segment[k-1].Time) {
				t.Fatalf("segment %d sample %d: time %v not after %v", index, k, segment[k].Time, segment[k-1].Time)
			}
		}
	}
	if len(profile.Hidden) != 1 || profile.Occultation.Type != OccultationTotal {
		t.Fatalf("zero-width range: %d intervals, type %v; want one total occultation", len(profile.Hidden), profile.Occultation.Type)
	}
	for _, pair := range [][2]time.Time{{profile.Hidden[0].Start, circular[0].Immersion}, {profile.Hidden[0].End, circular[0].Emersion}} {
		if d := math.Abs(pair[0].Sub(pair[1]).Seconds()); d > 0.05 {
			t.Errorf("zero-width range moved a contact by %.3f s", d)
		}
	}
}
