package astro_test

import (
	"encoding/json"
	"math"
	"os"
	"testing"

	"github.com/starainrt/astro/basic"
	"github.com/starainrt/astro/planet"
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

func TestDerivedHighPrecisionTruncationFullMatchesDefault(t *testing.T) {
	jd := 2469808.7654321
	lon := 116.391
	lat := 39.907
	tz := 8.0

	assertSame := func(name string, got, want float64) {
		t.Helper()
		if math.Float64bits(got) != math.Float64bits(want) {
			t.Fatalf("%s full-n mismatch", name)
		}
	}
	assertSamePair := func(name string, got1, got2, want1, want2 float64) {
		t.Helper()
		assertSame(name+".1", got1, want1)
		assertSame(name+".2", got2, want2)
	}

	assertSame("HSunTrueLo", basic.HSunTrueLo(jd), basic.HSunTrueLoN(jd, -1))
	assertSame("HSunTrueBo", basic.HSunTrueBo(jd), basic.HSunTrueBoN(jd, -1))
	assertSame("HSunApparentLo", basic.HSunApparentLo(jd), basic.HSunApparentLoN(jd, -1))
	assertSame("SunLoGXC", basic.SunLoGXC(jd), basic.SunLoGXCN(jd, -1))
	assertSame("EarthAway", basic.EarthAway(jd), basic.EarthAwayN(jd, -1))
	assertSame("HSunApparentRa", basic.HSunApparentRa(jd), basic.HSunApparentRaN(jd, -1))
	assertSame("HSunTrueRa", basic.HSunTrueRa(jd), basic.HSunTrueRaN(jd, -1))
	assertSame("HSunApparentDec", basic.HSunApparentDec(jd), basic.HSunApparentDecN(jd, -1))
	assertSame("HSunTrueDec", basic.HSunTrueDec(jd), basic.HSunTrueDecN(jd, -1))
	ra1, dec1 := basic.HSunApparentRaDec(jd)
	ra2, dec2 := basic.HSunApparentRaDecN(jd, -1)
	assertSamePair("HSunApparentRaDec", ra1, dec1, ra2, dec2)

	assertSame("HMoonTrueLo", basic.HMoonTrueLo(jd), basic.HMoonTrueLoN(jd, -1))
	assertSame("HMoonTrueBo", basic.HMoonTrueBo(jd), basic.HMoonTrueBoN(jd, -1))
	assertSame("HMoonAway", basic.HMoonAway(jd), basic.HMoonAwayN(jd, -1))
	assertSame("HMoonApparentLo", basic.HMoonApparentLo(jd), basic.HMoonApparentLoN(jd, -1))
	assertSame("HMoonTrueRa", basic.HMoonTrueRa(jd), basic.HMoonTrueRaN(jd, -1))
	assertSame("HMoonTrueDec", basic.HMoonTrueDec(jd), basic.HMoonTrueDecN(jd, -1))
	ra1, dec1 = basic.HMoonTrueRaDec(jd)
	ra2, dec2 = basic.HMoonTrueRaDecN(jd, -1)
	assertSamePair("HMoonTrueRaDec", ra1, dec1, ra2, dec2)
	ra1, dec1 = basic.HMoonApparentRaDec(jd, lon, lat, tz)
	ra2, dec2 = basic.HMoonApparentRaDecN(jd, lon, lat, tz, -1)
	assertSamePair("HMoonApparentRaDec", ra1, dec1, ra2, dec2)
	assertSame("HMoonApparentRa", basic.HMoonApparentRa(jd, lon, lat, tz), basic.HMoonApparentRaN(jd, lon, lat, tz, -1))
	assertSame("HMoonApparentDec", basic.HMoonApparentDec(jd, lon, lat, tz), basic.HMoonApparentDecN(jd, lon, lat, tz, -1))
	assertSame("HMoonAzimuth", basic.HMoonAzimuth(jd, lon, lat, tz), basic.HMoonAzimuthN(jd, lon, lat, tz, -1))
	assertSame("HMoonHeight", basic.HMoonHeight(jd, lon, lat, tz), basic.HMoonHeightN(jd, lon, lat, tz, -1))
}
