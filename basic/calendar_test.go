package basic

import (
	"math"
	"testing"
)

func TestGetJQTime(t *testing.T) {
	originalFunc := func(year, angle int) float64 {
		const iterations = 1

		var day int
		if angle%2 == 0 {
			day = 18
		} else {
			day = 3
		}

		month := 3.0
		if angle%10 != 0 {
			month += float64(angle+15) / 30.0
		} else {
			month += float64(angle) / 30.0
		}
		if month > 12 {
			month -= 12
		}

		jd := JDECalc(year, int(month), float64(day))
		if angle == 0 {
			angle = 360
		}

		for i := 0; i < iterations; i++ {
			for {
				jd0 := jd
				slope := (JQLospec(jd0+0.000005, float64(angle)) - JQLospec(jd0-0.000005, float64(angle))) / 0.00001
				jd = jd0 - (JQLospec(jd0, float64(angle))-float64(angle))/slope
				if math.Abs(jd-jd0) <= 0.00001 {
					break
				}
			}
			jd -= 0.001
		}
		jd += 0.001
		return TD2UT(jd, false)
	}

	testCases := []struct {
		year  int
		angle int
	}{
		{1900, 0}, {1900, 15}, {1900, 30}, {1900, 45}, {1900, 90}, {1900, 180}, {1900, 270}, {1900, 360},
		{1950, 0}, {1950, 30}, {1950, 90}, {1950, 180}, {1950, 270},
		{2000, 0}, {2000, 15}, {2000, 45}, {2000, 90}, {2000, 360},
		{2023, 0}, {2023, 30}, {2023, 90}, {2023, 180}, {2023, 270},
		{2100, 0}, {2100, 15}, {2100, 30}, {2100, 45}, {2100, 90}, {2100, 180}, {2100, 270}, {2100, 360},
		{2200, 0}, {2200, 30}, {2200, 90}, {2200, 180}, {2200, 270},
	}

	for _, tc := range testCases {
		originalResult := originalFunc(tc.year, tc.angle)
		optimizedResult := GetJQTime(tc.year, tc.angle)
		diff := math.Abs(originalResult - optimizedResult)
		if diff > 1e-10 {
			t.Fatalf("GetJQTime mismatch: year=%d angle=%d original=%.15f optimized=%.15f diff=%.15f",
				tc.year, tc.angle, originalResult, optimizedResult, diff)
		}
	}
}
