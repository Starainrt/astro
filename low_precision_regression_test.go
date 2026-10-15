// 低精度日月基线与高精度派生量的 n 截断收敛契约。
package astro_test

import (
	"encoding/json"
	"math"
	"os"
	"testing"

	"github.com/starainrt/astro/basic"
	"github.com/starainrt/astro/planet"

	"github.com/starainrt/astro/internal/ephemtest"
)

type lowPrecisionSunSnapshot struct {
	LoBits          uint64 `json:"lo_bits"`
	MBits           uint64 `json:"m_bits"`
	EccBits         uint64 `json:"ecc_bits"`
	PeriBits        uint64 `json:"peri_bits"`
	MidBits         uint64 `json:"mid_bits"`
	TrueLoBits      uint64 `json:"true_lo_bits"`
	ApparentLoBits  uint64 `json:"apparent_lo_bits"`
	ApparentRaBits  uint64 `json:"apparent_ra_bits"`
	ApparentDecBits uint64 `json:"apparent_dec_bits"`
	TrueRaBits      uint64 `json:"true_ra_bits"`
	TrueDecBits     uint64 `json:"true_dec_bits"`
	DistanceBits    uint64 `json:"distance_bits"`
}

type lowPrecisionMoonSnapshot struct {
	LoBits         uint64 `json:"lo_bits"`
	SunAngleBits   uint64 `json:"sun_angle_bits"`
	MBits          uint64 `json:"m_bits"`
	LonXBits       uint64 `json:"lonx_bits"`
	IBits          uint64 `json:"i_bits"`
	RBits          uint64 `json:"r_bits"`
	BBits          uint64 `json:"b_bits"`
	TrueLoBits     uint64 `json:"true_lo_bits"`
	TrueBoBits     uint64 `json:"true_bo_bits"`
	AwayBits       uint64 `json:"away_bits"`
	ApparentLoBits uint64 `json:"apparent_lo_bits"`
	TrueRaBits     uint64 `json:"true_ra_bits"`
	TrueDecBits    uint64 `json:"true_dec_bits"`
}

type lowPrecisionSample struct {
	UTC  string                   `json:"utc"`
	TTJD float64                  `json:"tt_jd"`
	Sun  lowPrecisionSunSnapshot  `json:"sun"`
	Moon lowPrecisionMoonSnapshot `json:"moon"`
}

func loadLowPrecisionSamples(t *testing.T) []lowPrecisionSample {
	t.Helper()

	data, err := os.ReadFile("testdata/low_precision_sun_moon_baseline.json")
	if err != nil {
		t.Fatal(err)
	}

	var samples []lowPrecisionSample
	if err := json.Unmarshal(data, &samples); err != nil {
		t.Fatal(err)
	}
	if len(samples) == 0 {
		t.Fatal("empty low precision baseline samples")
	}
	return samples
}

func TestLowPrecisionSunMoonBaselineRegression(t *testing.T) {
	samples := loadLowPrecisionSamples(t)

	assertBits := func(t *testing.T, name, utc string, got float64, want uint64) {
		t.Helper()
		if math.Float64bits(got) != want {
			t.Fatalf("%s regression at %s", name, utc)
		}
	}

	for _, sample := range samples {
		jd := sample.TTJD

		assertBits(t, "planet.SunLo", sample.UTC, planet.SunLo(jd), sample.Sun.LoBits)
		assertBits(t, "basic.SunLo", sample.UTC, basic.SunLo(jd), sample.Sun.LoBits)
		assertBits(t, "planet.SunM", sample.UTC, planet.SunM(jd), sample.Sun.MBits)
		assertBits(t, "basic.SunM", sample.UTC, basic.SunM(jd), sample.Sun.MBits)
		assertBits(t, "planet.Earthe", sample.UTC, planet.Earthe(jd), sample.Sun.EccBits)
		assertBits(t, "basic.Earthe", sample.UTC, basic.Earthe(jd), sample.Sun.EccBits)
		assertBits(t, "planet.EarthPI", sample.UTC, planet.EarthPI(jd), sample.Sun.PeriBits)
		assertBits(t, "basic.EarthPI", sample.UTC, basic.EarthPI(jd), sample.Sun.PeriBits)
		assertBits(t, "planet.SunMidFun", sample.UTC, planet.SunMidFun(jd), sample.Sun.MidBits)
		assertBits(t, "basic.SunMidFun", sample.UTC, basic.SunMidFun(jd), sample.Sun.MidBits)
		assertBits(t, "planet.SunTrueLo", sample.UTC, planet.SunTrueLo(jd), sample.Sun.TrueLoBits)
		assertBits(t, "basic.SunTrueLo", sample.UTC, basic.SunTrueLo(jd), sample.Sun.TrueLoBits)
		assertBits(t, "planet.SunApparentLo", sample.UTC, planet.SunApparentLo(jd), sample.Sun.ApparentLoBits)
		assertBits(t, "basic.SunApparentLo", sample.UTC, basic.SunApparentLo(jd), sample.Sun.ApparentLoBits)
		assertBits(t, "basic.SunApparentRa", sample.UTC, basic.SunApparentRa(jd), sample.Sun.ApparentRaBits)
		assertBits(t, "basic.SunApparentDec", sample.UTC, basic.SunApparentDec(jd), sample.Sun.ApparentDecBits)
		assertBits(t, "basic.SunTrueRa", sample.UTC, basic.SunTrueRa(jd), sample.Sun.TrueRaBits)
		assertBits(t, "basic.SunTrueDec", sample.UTC, basic.SunTrueDec(jd), sample.Sun.TrueDecBits)
		assertBits(t, "planet.Distance", sample.UTC, planet.Distance(jd), sample.Sun.DistanceBits)
		assertBits(t, "basic.Distance", sample.UTC, basic.Distance(jd), sample.Sun.DistanceBits)
		assertBits(t, "planet.MoonLo", sample.UTC, planet.MoonLo(jd), sample.Moon.LoBits)
		assertBits(t, "basic.MoonLo", sample.UTC, basic.MoonLo(jd), sample.Moon.LoBits)
		assertBits(t, "planet.SunMoonAngle", sample.UTC, planet.SunMoonAngle(jd), sample.Moon.SunAngleBits)
		assertBits(t, "basic.SunMoonAngle", sample.UTC, basic.SunMoonAngle(jd), sample.Moon.SunAngleBits)
		assertBits(t, "planet.MoonM", sample.UTC, planet.MoonM(jd), sample.Moon.MBits)
		assertBits(t, "basic.MoonM", sample.UTC, basic.MoonM(jd), sample.Moon.MBits)
		assertBits(t, "planet.MoonLonX", sample.UTC, planet.MoonLonX(jd), sample.Moon.LonXBits)
		assertBits(t, "basic.MoonLonX", sample.UTC, basic.MoonLonX(jd), sample.Moon.LonXBits)
		assertBits(t, "planet.MoonI", sample.UTC, planet.MoonI(jd), sample.Moon.IBits)
		assertBits(t, "basic.MoonI", sample.UTC, basic.MoonI(jd), sample.Moon.IBits)
		assertBits(t, "planet.MoonR", sample.UTC, planet.MoonR(jd), sample.Moon.RBits)
		assertBits(t, "basic.MoonR", sample.UTC, basic.MoonR(jd), sample.Moon.RBits)
		assertBits(t, "planet.MoonB", sample.UTC, planet.MoonB(jd), sample.Moon.BBits)
		assertBits(t, "basic.MoonB", sample.UTC, basic.MoonB(jd), sample.Moon.BBits)
		assertBits(t, "planet.MoonTrueLo", sample.UTC, planet.MoonTrueLo(jd), sample.Moon.TrueLoBits)
		assertBits(t, "basic.MoonTrueLo", sample.UTC, basic.MoonTrueLo(jd), sample.Moon.TrueLoBits)
		assertBits(t, "planet.MoonTrueBo", sample.UTC, planet.MoonTrueBo(jd), sample.Moon.TrueBoBits)
		assertBits(t, "basic.MoonTrueBo", sample.UTC, basic.MoonTrueBo(jd), sample.Moon.TrueBoBits)
		assertBits(t, "planet.MoonAway", sample.UTC, planet.MoonAway(jd), sample.Moon.AwayBits)
		assertBits(t, "basic.MoonAway", sample.UTC, basic.MoonAway(jd), sample.Moon.AwayBits)
		assertBits(t, "basic.MoonApparentLo", sample.UTC, basic.MoonApparentLo(jd), sample.Moon.ApparentLoBits)
		assertBits(t, "basic.MoonTrueRa", sample.UTC, basic.MoonTrueRa(jd), sample.Moon.TrueRaBits)
		assertBits(t, "basic.MoonTrueDec", sample.UTC, basic.MoonTrueDec(jd), sample.Moon.TrueDecBits)
	}
}

func derivedTruncationCases(samples []lowPrecisionSample) []truncationCase {
	jdOf := func(index int) float64 { return samples[index].TTJD }
	lon := 116.391
	lat := 39.907
	tz := 8.0
	// withSite 把带观测点经纬度与时区的入口适配成 jd 探针。
	withSite := func(f func(jd, lon, lat, tz float64, n int) float64) func(float64, int) float64 {
		return func(jd float64, n int) float64 { return f(jd, lon, lat, tz, n) }
	}
	withSiteRaDec := func(f func(jd, lon, lat, tz float64, n int) (float64, float64)) func(float64, int) (float64, float64) {
		return func(jd float64, n int) (float64, float64) { return f(jd, lon, lat, tz, n) }
	}

	// 上限按 48 个采样点实测最坏值放宽 5 倍以上给出，随 n 递减。
	return []truncationCase{
		scalarTruncation("HSunTrueLoN", jdOf, basic.HSunTrueLoN, true, []float64{0.15, 0.05, 0.02, 0.006}),
		scalarTruncation("HSunTrueBoN", jdOf, basic.HSunTrueBoN, true, []float64{1e-3, 5e-4, 3e-4, 1.5e-4}),
		scalarTruncation("HSunApparentLoN", jdOf, basic.HSunApparentLoN, true, []float64{0.15, 0.05, 0.02, 0.006}),
		scalarTruncation("SunLoGXCN", jdOf, basic.SunLoGXCN, true, []float64{1e-5, 2e-6, 1e-6, 3e-7}),
		scalarTruncation("EarthAwayN", jdOf, basic.EarthAwayN, false, []float64{1e-3, 3e-4, 1e-4, 3e-5}),
		scalarTruncation("HSunApparentRaN", jdOf, basic.HSunApparentRaN, true, []float64{0.15, 0.05, 0.02, 0.006}),
		scalarTruncation("HSunTrueRaN", jdOf, basic.HSunTrueRaN, true, []float64{0.15, 0.05, 0.02, 0.006}),
		scalarTruncation("HSunApparentDecN", jdOf, basic.HSunApparentDecN, true, []float64{0.05, 0.02, 0.008, 0.003}),
		scalarTruncation("HSunTrueDecN", jdOf, basic.HSunTrueDecN, true, []float64{0.05, 0.02, 0.008, 0.003}),
		pairTruncation("HSunApparentRaDecN", jdOf, basic.HSunApparentRaDecN, []float64{0.15, 0.05, 0.02, 0.006}),
		scalarTruncation("HMoonTrueLoN", jdOf, basic.HMoonTrueLoN, true, []float64{5, 2.5, 1, 0.25}),
		scalarTruncation("HMoonTrueBoN", jdOf, basic.HMoonTrueBoN, true, []float64{3, 0.75, 0.15, 0.06}),
		scalarTruncation("HMoonAwayN", jdOf, basic.HMoonAwayN, false, []float64{40000, 7000, 2500, 400}),
		scalarTruncation("HMoonApparentLoN", jdOf, basic.HMoonApparentLoN, true, []float64{5, 2.5, 1, 0.25}),
		scalarTruncation("HMoonTrueRaN", jdOf, basic.HMoonTrueRaN, true, []float64{5, 2.5, 1, 0.3}),
		scalarTruncation("HMoonTrueDecN", jdOf, basic.HMoonTrueDecN, true, []float64{3, 1, 0.4, 0.1}),
		pairTruncation("HMoonTrueRaDecN", jdOf, basic.HMoonTrueRaDecN, []float64{5, 2.5, 1, 0.3}),
		scalarTruncation("HMoonApparentRaN", jdOf, withSite(basic.HMoonApparentRaN), true, []float64{5, 2.5, 1, 0.3}),
		scalarTruncation("HMoonApparentDecN", jdOf, withSite(basic.HMoonApparentDecN), true, []float64{4, 1, 0.3, 0.1}),
		pairTruncation("HMoonApparentRaDecN", jdOf, withSiteRaDec(basic.HMoonApparentRaDecN), []float64{5, 2.5, 1, 0.3}),
		scalarTruncation("HMoonAzimuthN", jdOf, withSite(basic.HMoonAzimuthN), true, []float64{9, 4, 2, 0.9}),
		scalarTruncation("HMoonHeightN", jdOf, withSite(basic.HMoonHeightN), true, []float64{4, 2, 0.7, 0.15}),
	}
}

func TestDerivedHighPrecisionTruncationConverges(t *testing.T) {
	ephemtest.PinLegacy(t)
	samples := loadLowPrecisionSamples(t)
	assertTruncationConverges(t, derivedTruncationCases(samples), len(samples))
}
