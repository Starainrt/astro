package basic

// 剖面不报告半径范围时，月掩须在当时视向上扫满一圈求范围，结果与报告范围的同一剖面一致。

import (
	"math"
	"testing"
	"time"

	"github.com/starainrt/astro/moon/limb"
)

// unrangedLunarLimb 只暴露 LimbRadiusKM，迫使月掩走整圈扫描。
type unrangedLunarLimb struct{ profile *limb.Profile }

func (u unrangedLunarLimb) LimbRadiusKM(lon, lat, angle float64) (float64, bool) {
	return u.profile.LimbRadiusKM(lon, lat, angle)
}

// limbBruteRangeKM 以剖面步长的 1/8 细扫一圈，作整圈扫描的对拍基准。
func limbBruteRangeKM(t *testing.T, tt float64, profile LunarLimb, observer Observer, ra, dec float64) (float64, float64) {
	t.Helper()
	geoRA, geoDec := HMoonGeocentricApparentRaDecNWithInstance(tt, -1, nil)
	frame := newLunarLimbFrame(profile, tt, lunarEclipseLimbVector(geoRA, geoDec, HMoonAwayNWithInstance(tt, -1, nil)), nil)
	topoRA, topoDec := moonTopocentricApparentRaDec(tt, observer, -1, nil)
	view, ok := frame.viewToward(lunarEclipseLimbVector(topoRA, topoDec, 1), lunarEclipseLimbVector(ra, dec, 1))
	if !frame.valid || !ok {
		t.Fatal("limb view failed")
	}
	low, high := math.Inf(1), math.Inf(-1)
	for offset := 0.0; offset < 360; offset += frame.stepDeg / 8 {
		radius, ok := frame.radiusAt(view, offset)
		if !ok {
			t.Fatal("limb radius failed")
		}
		low, high = math.Min(low, radius), math.Max(high, radius)
	}
	return low, high
}

func TestOccultationLimbRangeScansFullCircleWithoutProvider(t *testing.T) {
	profile := limb.Default()
	unranged := unrangedLunarLimb{profile: profile}
	observer := Observer{Longitude: 41.5907, Latitude: -55.07036}
	tt := occultationTimeToTT(time.Date(2026, 2, 11, 3, 47, 15, 0, time.UTC))
	ra, dec := starApparentRaDec(tt, limbTestAntares, observer, nil)
	wantLow, wantHigh := limbBruteRangeKM(t, tt, unranged, observer, ra, dec)
	// 剖面网格节点即极值，整圈扫描与细扫只差插值误差；扫描步长放粗会漏掉极值。
	const tolerance = 0.05
	starLow, starHigh, ok := starOccultationLimbRange(tt, limbTestAntares, observer, unranged, nil)
	if !ok {
		t.Fatal("star range failed")
	}
	target := func(float64) (float64, float64, float64, float64, bool) { return ra, dec, 0, 0, true }
	low, high, ok := occultationLimbRangeKM(tt, unranged, observer, target, nil)
	if !ok {
		t.Fatal("occultation range failed")
	}
	for _, got := range []struct {
		name      string
		low, high float64
		pad       float64
	}{{"star", starLow, starHigh, starOccultationLimbUnrangedPadKM}, {"planet", low, high, occultationLimbUnrangedPadKM}} {
		if math.Abs(got.low+got.pad-wantLow) > tolerance || math.Abs(got.high-got.pad-wantHigh) > tolerance {
			t.Errorf("%s: range %.3f–%.3f km, brute force %.3f–%.3f km", got.name, got.low+got.pad, got.high-got.pad, wantLow, wantHigh)
		}
	}
	rangedLow, rangedHigh := profile.RadiusRangeKM()
	if wantLow < rangedLow || wantHigh > rangedHigh || wantHigh-wantLow < 1 {
		t.Errorf("brute force %.3f–%.3f km outside table range %.3f–%.3f km or too flat", wantLow, wantHigh, rangedLow, rangedHigh)
	}
}

func TestStarOccultationUnrangedLimbMatchesRanged(t *testing.T) {
	const lon, lat = 41.5907, -55.07036 + 0.4/111.2
	ranged := limbTestAntaresEvent(t, lon, lat, limb.Default())
	unranged := limbTestAntaresEvent(t, lon, lat, unrangedLunarLimb{profile: limb.Default()})
	if len(ranged) != 1 || len(unranged) != 1 || ranged[0].Type != OccultationGrazing || unranged[0].Type != ranged[0].Type {
		t.Fatalf("ranged %+v vs unranged %+v, want the same limb graze", ranged, unranged)
	}
	for _, pair := range [][2]time.Time{{ranged[0].Immersion, unranged[0].Immersion}, {ranged[0].Emersion, unranged[0].Emersion}} {
		if d := math.Abs(pair[0].Sub(pair[1]).Seconds()); d > 0.05 {
			t.Errorf("unranged limb moved a contact by %.3f s", d)
		}
	}
}
