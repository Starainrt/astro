// 被测契约：标准气象下与原升落一致；高压低温让日出提前、日落推迟；非法气象报错。
package basic

import (
	"math"
	"testing"
)

func TestRiseSetWithAtmosphere(t *testing.T) {
	jd, lon, lat, tz := 2460676.5, 116.4, 39.9, 8.0
	cases := []struct {
		name     string
		standard func() (float64, error)
		weather  func(p, c float64) (float64, error)
		rise     bool
	}{
		{"sunrise", func() (float64, error) { return GetSunRiseTimeWithInstance(jd, lon, lat, tz, 1, 0, nil) },
			func(p, c float64) (float64, error) {
				return GetSunRiseTimeWithAtmosphere(jd, lon, lat, tz, 0, p, c, nil)
			}, true},
		{"sunset", func() (float64, error) { return GetSunSetTimeWithInstance(jd, lon, lat, tz, 1, 0, nil) },
			func(p, c float64) (float64, error) {
				return GetSunSetTimeWithAtmosphere(jd, lon, lat, tz, 0, p, c, nil)
			}, false},
		{"moonrise", func() (float64, error) { return GetMoonRiseTimeWithInstance(jd, lon, lat, tz, 1, 0, nil) },
			func(p, c float64) (float64, error) {
				return GetMoonRiseTimeWithAtmosphere(jd, lon, lat, tz, 0, p, c, nil)
			}, true},
		{"moonset", func() (float64, error) { return GetMoonSetTimeWithInstance(jd, lon, lat, tz, 1, 0, nil) },
			func(p, c float64) (float64, error) {
				return GetMoonSetTimeWithAtmosphere(jd, lon, lat, tz, 0, p, c, nil)
			}, false},
	}
	for _, c := range cases {
		base, err := c.standard()
		if err != nil {
			t.Fatalf("%s: %v", c.name, err)
		}
		same, err := c.weather(refractionStandardPressureHPa, refractionStandardTemperatureC)
		if err != nil || math.Abs(same-base)*86400 > 0.01 {
			t.Errorf("%s: standard weather %.6f vs %.6f (%v)", c.name, same, base, err)
		}
		dense, err := c.weather(1040, -30)
		if err != nil {
			t.Fatalf("%s dense: %v", c.name, err)
		}
		shift := (dense - base) * 86400
		t.Logf("%s: 1040 hPa / -30 °C shifts %+.1f s", c.name, shift)
		if c.rise && shift >= 0 || !c.rise && shift <= 0 || math.Abs(shift) > 120 {
			t.Errorf("%s: implausible shift %.1f s", c.name, shift)
		}
	}
	if _, err := GetSunRiseTimeWithAtmosphere(jd, lon, lat, tz, 0, -1, 10, nil); err != ErrInvalidObservationInput {
		t.Errorf("negative pressure: %v", err)
	}
	if _, err := GetSunRiseTimeWithAtmosphere(jd, lon, lat, tz, 0, 1010, -300, nil); err != ErrInvalidObservationInput {
		t.Errorf("below absolute zero: %v", err)
	}
}

func TestPointRiseSetWithAtmosphere(t *testing.T) {
	jd, lon, lat, tz := 2460676.5, 116.4, 39.9, 8.0
	type pair struct {
		name         string
		plain, atmos func(p, temp float64) (float64, error)
	}
	cases := []pair{
		{"MarsRise", func(float64, float64) (float64, error) { return MarsRiseTimeWithInstance(jd, lon, lat, tz, 1, 0, nil) },
			func(p, temp float64) (float64, error) {
				return MarsRiseTimeWithAtmosphere(jd, lon, lat, tz, 0, p, temp, nil)
			}},
		{"JupiterSet", func(float64, float64) (float64, error) {
			return JupiterSetTimeWithInstance(jd, lon, lat, tz, 1, 0, nil)
		},
			func(p, temp float64) (float64, error) {
				return JupiterSetTimeWithAtmosphere(jd, lon, lat, tz, 0, p, temp, nil)
			}},
		{"SiriusRise", func(float64, float64) (float64, error) {
			return StarRiseTime(jd, 101.29, -16.72, lon, lat, 0, tz, true)
		},
			func(p, temp float64) (float64, error) {
				return StarRiseTimeWithAtmosphere(jd, 101.29, -16.72, lon, lat, tz, 0, p, temp)
			}},
	}
	for _, c := range cases {
		plain, err := c.plain(0, 0)
		if err != nil {
			t.Fatalf("%s plain: %v", c.name, err)
		}
		std, err := c.atmos(refractionStandardPressureHPa, refractionStandardTemperatureC)
		if err != nil {
			t.Fatalf("%s standard: %v", c.name, err)
		}
		// 原接口用固定 34′ 地平折射，这里按公式在实际高度求折射，标准气象下差约 2–3 s。
		if d := math.Abs(std-plain) * 86400; d > 5 {
			t.Errorf("%s: standard weather differs from plain by %.2f s", c.name, d)
		}
		dense, err := c.atmos(1040, -30)
		if err != nil {
			t.Fatalf("%s dense: %v", c.name, err)
		}
		shift := (dense - std) * 86400
		t.Logf("%s: 1040 hPa / -30 °C shifts %+.1f s", c.name, shift)
		rising := c.name != "JupiterSet"
		if (rising && shift >= 0) || (!rising && shift <= 0) || math.Abs(shift) > 300 {
			t.Errorf("%s: implausible shift %.1f s", c.name, shift)
		}
	}
	if _, err := MarsRiseTimeWithAtmosphere(jd, lon, lat, tz, 0, -1, 10, nil); err == nil {
		t.Error("invalid pressure must fail")
	}
}

func TestRiseSetWeatherRejectsOutOfRange(t *testing.T) {
	jd, lon, lat, tz := 2460831.0, 116.4, 39.9, 8.0
	for _, c := range [][2]float64{{1.01e9, 10}, {1e6, 10}, {1201, 10}, {1010, -273.14}, {1010, -101}, {1010, 81}, {math.NaN(), 10}, {1010, math.Inf(1)}} {
		if _, err := GetSunRiseTimeWithAtmosphere(jd, lon, lat, tz, 0, c[0], c[1], nil); err != ErrInvalidObservationInput {
			t.Errorf("sun p=%g T=%g: err=%v", c[0], c[1], err)
		}
		if _, err := MercuryRiseTimeWithAtmosphere(jd, lon, lat, tz, 0, c[0], c[1], nil); err != ErrInvalidObservationInput {
			t.Errorf("mercury p=%g T=%g: err=%v", c[0], c[1], err)
		}
	}
	for _, c := range [][2]float64{{1200, 10}, {1010, -100}, {1010, 80}, {300, 10}} {
		if _, err := GetSunRiseTimeWithAtmosphere(jd, lon, lat, tz, 0, c[0], c[1], nil); err != nil {
			t.Errorf("sun p=%g T=%g must be accepted: %v", c[0], c[1], err)
		}
	}
}
