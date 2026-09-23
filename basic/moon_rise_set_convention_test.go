package basic

import (
	"errors"
	"math"
	"testing"
)

// 本文件钉住月出/月落「缺失事件」的口径：错误名描述缺失的那个现象，不描述被问的事件。

func moonRiseSetErrorMatches(got, want error) bool {
	if want == nil {
		return got == nil
	}
	return errors.Is(got, want)
}

func moonRiseSetDailyGeometry(jd, lon, lat, tz float64) (aboveAll, belowAll bool) {
	dayStart := math.Floor(jd) + 0.5
	aboveAll, belowAll = true, true
	previous := moonRiseSetResidual(dayStart, lon, lat, tz, 1, 0, -1)
	for i := 1; i <= 288; i++ {
		current := moonRiseSetResidual(dayStart+float64(i)/288.0, lon, lat, tz, 1, 0, -1)
		if previous <= 0 || current <= 0 {
			aboveAll = false
		}
		if previous >= 0 || current >= 0 {
			belowAll = false
		}
		previous = current
	}
	return aboveAll, belowAll
}

func TestMoonRiseSetMissingEventConvention(t *testing.T) {
	cases := []struct {
		name            string
		jd              float64
		lon, lat, tz    float64
		riseErr, setErr error
	}{
		{"极昼：全天在地平线上", JDCalc(2023, 6, 21), 0, 85, 0, ErrNeverSet, ErrNeverSet},
		{"极夜：全天在地平线下", JDCalc(2023, 12, 22), 0, -85, 0, ErrNeverRise, ErrNeverRise},
		{"当日无升起但别日有", JDCalc(2024, 2, 29), 0, 60, 0, ErrNotOnThisDate, nil},
		{"正常日两侧都有", JDCalc(2025, 6, 5), 116.4074, 39.9042, 8, nil, nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := GetMoonRiseTime(tc.jd, tc.lon, tc.lat, tc.tz, 1, 0); !moonRiseSetErrorMatches(err, tc.riseErr) {
				t.Fatalf("GetMoonRiseTime error = %v, want %v", err, tc.riseErr)
			}
			if _, err := GetMoonSetTime(tc.jd, tc.lon, tc.lat, tc.tz, 1, 0); !moonRiseSetErrorMatches(err, tc.setErr) {
				t.Fatalf("GetMoonSetTime error = %v, want %v", err, tc.setErr)
			}
		})
	}
}

func TestMoonRiseSetMissingEventMatchesDailyGeometry(t *testing.T) {
	dates := []float64{
		JDCalc(2023, 6, 21), JDCalc(2023, 12, 22), JDCalc(2024, 2, 29),
		JDCalc(2025, 6, 21), JDCalc(2025, 12, 22), JDCalc(2026, 3, 3),
	}
	latitudes := []float64{-89, -85, -75, -66, -60, 60, 66, 75, 85, 89}
	for _, jd := range dates {
		for _, lat := range latitudes {
			aboveAll, belowAll := moonRiseSetDailyGeometry(jd, 0, lat, 0)
			if !aboveAll && !belowAll {
				continue
			}
			_, riseErr := GetMoonRiseTime(jd, 0, lat, 0, 1, 0)
			_, setErr := GetMoonSetTime(jd, 0, lat, 0, 1, 0)
			want := ErrNeverSet
			if belowAll {
				want = ErrNeverRise
			}
			if !moonRiseSetErrorMatches(riseErr, want) {
				t.Fatalf("lat %v jd %.1f aboveAll=%v belowAll=%v: moonrise error = %v, want %v",
					lat, jd, aboveAll, belowAll, riseErr, want)
			}
			if !moonRiseSetErrorMatches(setErr, want) {
				t.Fatalf("lat %v jd %.1f aboveAll=%v belowAll=%v: moonset error = %v, want %v",
					lat, jd, aboveAll, belowAll, setErr, want)
			}
		}
	}
}
