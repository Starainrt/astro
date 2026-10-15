// 被测契约：上缘、标准气象下的 0° 穿越与现有日出日落一致；高纬状态与相切判定正确。
package basic

import (
	"math"
	"testing"
	"time"
)

func beijingDate(y int, m time.Month, d int) time.Time {
	return time.Date(y, m, d, 0, 0, 0, 0, time.FixedZone("CST", 8*3600))
}

func TestSkyAltitudeCrossingMatchesSunriseAtHorizon(t *testing.T) {
	jd := Date2JD(beijingDate(2025, 3, 20))
	rise, err := GetSunRiseTimeWithAtmosphere(jd, 116.4, 39.9, 8, 0, 1010, 10, nil)
	if err != nil {
		t.Fatal(err)
	}
	set, _ := GetSunSetTimeWithAtmosphere(jd, 116.4, 39.9, 8, 0, 1010, 10, nil)
	got, err := SkyAltitudeCrossingWithInstance(SkySun, jd, 116.4, 39.9, 8, 0, 0, AltitudeCrossingOptions{}, nil)
	if err != nil || got.State != AltitudeCrosses {
		t.Fatalf("state %v err %v", got.State, err)
	}
	if d := math.Abs(got.Rise-rise) * 86400; d > 2 {
		t.Errorf("rise differs by %.2f s", d)
	}
	if d := math.Abs(got.Set-set) * 86400; d > 2 {
		t.Errorf("set differs by %.2f s", d)
	}
	civil, _ := SkyAltitudeCrossingWithInstance(SkySun, jd, 116.4, 39.9, 8, 0, -6, AltitudeCrossingOptions{Limb: AltitudeLimbCenter, NoRefraction: true}, nil)
	if !(civil.Rise < got.Rise && civil.Set > got.Set) {
		t.Error("civil twilight must bracket sunrise and sunset")
	}
}

func TestSkyAltitudeCrossingStates(t *testing.T) {
	summer := Date2JD(beijingDate(2025, 6, 21))
	winter := Date2JD(beijingDate(2025, 12, 21))
	above, _ := SkyAltitudeCrossingWithInstance(SkySun, summer, 20, 78, 0, 0, 0, AltitudeCrossingOptions{}, nil)
	below, _ := SkyAltitudeCrossingWithInstance(SkySun, winter, 20, 78, 0, 0, 0, AltitudeCrossingOptions{}, nil)
	if above.State != AltitudeAlwaysAbove || below.State != AltitudeAlwaysBelow {
		t.Errorf("polar states %v %v", above.State, below.State)
	}
	peak := 90 - 39.9 + 23.4366
	tangent, _ := SkyAltitudeCrossingWithInstance(SkySun, summer, 116.4, 39.9, 8, 0, peak, AltitudeCrossingOptions{Limb: AltitudeLimbCenter, NoRefraction: true}, nil)
	if tangent.State == AltitudeCrosses && !math.IsNaN(tangent.Rise) && !math.IsNaN(tangent.Set) && (tangent.Set-tangent.Rise)*1440 > 30 {
		t.Errorf("near-peak altitude should graze, got %+v", tangent)
	}
	if _, err := SkyAltitudeCrossingWithInstance(SkySun, summer, 0, 95, 0, 0, 0, AltitudeCrossingOptions{}, nil); err == nil {
		t.Error("invalid latitude must fail")
	}
}

func TestStarAndPlanetAltitudeCrossing(t *testing.T) {
	jd := Date2JD(beijingDate(2025, 3, 20))
	star, err := StarAltitudeCrossing(jd, 101.287, -16.716, 116.4, 39.9, 8, 10, AltitudeCrossingOptions{})
	if err != nil || star.State != AltitudeCrosses || math.IsNaN(star.Rise) || math.IsNaN(star.Set) {
		t.Fatalf("Sirius at 10°: %+v %v", star, err)
	}
	mars, err := SkyAltitudeCrossingWithInstance(SkyMars, jd, 116.4, 39.9, 8, 0, 0, AltitudeCrossingOptions{}, nil)
	if err != nil || mars.State != AltitudeCrosses {
		t.Fatalf("mars: %+v %v", mars, err)
	}
}

func TestAltitudeCrossingFindsNearPeakCrossings(t *testing.T) {
	jd := Date2JD(beijingDate(2025, 3, 20))
	options := AltitudeCrossingOptions{NoRefraction: true}
	for ra := 0.0; ra < 360; ra += 15 {
		got, err := StarAltitudeCrossing(jd, ra, 19.9, 116.4, 39.9, 8, 69.999, options)
		if err != nil || got.State != AltitudeCrosses || got.MaxResidual <= 0 {
			t.Fatalf("ra %v: %+v %v", ra, got, err)
		}
		if !math.IsNaN(got.Rise) && !math.IsNaN(got.Set) && !(got.Rise < got.Set && (got.Set-got.Rise)*1440 < 10) {
			t.Errorf("ra %v: rise %v set %v", ra, got.Rise, got.Set)
		}
		if math.IsNaN(got.Rise) && math.IsNaN(got.Set) {
			t.Errorf("ra %v: no crossing", ra)
		}
	}
	missed, _ := StarAltitudeCrossing(jd, 90, 19.9, 116.4, 39.9, 8, 70.01, options)
	if missed.State != AltitudeAlwaysBelow {
		t.Errorf("above peak: %+v", missed)
	}
}
