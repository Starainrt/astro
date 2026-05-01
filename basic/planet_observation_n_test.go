package basic_test

import (
	"math"
	"testing"
	"time"

	"github.com/starainrt/astro/basic"
)

func TestBasicPlanetObservationNFullMatchesDefault(t *testing.T) {
	date := time.Date(2026, 4, 26, 9, 30, 45, 123456789, time.FixedZone("CST", 8*3600))
	ttJD := basic.TD2UT(basic.Date2JDE(date.UTC()), true)
	jde := basic.Date2JDE(date)
	lon := 116.391
	lat := 39.907
	tz := 8.0
	height := 45.0

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
	assertSameErr := func(name string, got, want error) {
		t.Helper()
		if got != want {
			t.Fatalf("%s full-n mismatch", name)
		}
	}

	floatChecks := []struct {
		name string
		got  func() float64
		want func() float64
	}{
		{"MercuryApparentLo", func() float64 { return basic.MercuryApparentLo(ttJD) }, func() float64 { return basic.MercuryApparentLoN(ttJD, -1) }},
		{"MercuryApparentBo", func() float64 { return basic.MercuryApparentBo(ttJD) }, func() float64 { return basic.MercuryApparentBoN(ttJD, -1) }},
		{"MercuryApparentRa", func() float64 { return basic.MercuryApparentRa(ttJD) }, func() float64 { return basic.MercuryApparentRaN(ttJD, -1) }},
		{"MercuryApparentDec", func() float64 { return basic.MercuryApparentDec(ttJD) }, func() float64 { return basic.MercuryApparentDecN(ttJD, -1) }},
		{"EarthMercuryAway", func() float64 { return basic.EarthMercuryAway(ttJD) }, func() float64 { return basic.EarthMercuryAwayN(ttJD, -1) }},
		{"MercuryMag", func() float64 { return basic.MercuryMag(ttJD) }, func() float64 { return basic.MercuryMagN(ttJD, -1) }},
		{"MercuryPhaseAngle", func() float64 { return basic.MercuryPhaseAngle(ttJD) }, func() float64 { return basic.MercuryPhaseAngleN(ttJD, -1) }},
		{"MercuryIlluminatedFraction", func() float64 { return basic.MercuryIlluminatedFraction(ttJD) }, func() float64 { return basic.MercuryIlluminatedFractionN(ttJD, -1) }},
		{"MercuryBrightLimbPositionAngle", func() float64 { return basic.MercuryBrightLimbPositionAngle(ttJD) }, func() float64 { return basic.MercuryBrightLimbPositionAngleN(ttJD, -1) }},
		{"MercuryHeight", func() float64 { return basic.MercuryHeight(jde, lon, lat, tz) }, func() float64 { return basic.MercuryHeightN(jde, lon, lat, tz, -1) }},
		{"MercuryAzimuth", func() float64 { return basic.MercuryAzimuth(jde, lon, lat, tz) }, func() float64 { return basic.MercuryAzimuthN(jde, lon, lat, tz, -1) }},
		{"MercuryHourAngle", func() float64 { return basic.MercuryHourAngle(jde, lon, tz) }, func() float64 { return basic.MercuryHourAngleN(jde, lon, tz, -1) }},
		{"MercuryCulminationTime", func() float64 { return basic.MercuryCulminationTime(jde, lon, tz) }, func() float64 { return basic.MercuryCulminationTimeN(jde, lon, tz, -1) }},
		{"MercuryRiseTime", func() float64 { value, _ := basic.MercuryRiseTime(jde, lon, lat, tz, 1, height); return value }, func() float64 { value, _ := basic.MercuryRiseTimeN(jde, lon, lat, tz, 1, height, -1); return value }},
		{"MercurySetTime", func() float64 { value, _ := basic.MercurySetTime(jde, lon, lat, tz, 1, height); return value }, func() float64 { value, _ := basic.MercurySetTimeN(jde, lon, lat, tz, 1, height, -1); return value }},
		{"VenusApparentLo", func() float64 { return basic.VenusApparentLo(ttJD) }, func() float64 { return basic.VenusApparentLoN(ttJD, -1) }},
		{"VenusApparentBo", func() float64 { return basic.VenusApparentBo(ttJD) }, func() float64 { return basic.VenusApparentBoN(ttJD, -1) }},
		{"VenusApparentRa", func() float64 { return basic.VenusApparentRa(ttJD) }, func() float64 { return basic.VenusApparentRaN(ttJD, -1) }},
		{"VenusApparentDec", func() float64 { return basic.VenusApparentDec(ttJD) }, func() float64 { return basic.VenusApparentDecN(ttJD, -1) }},
		{"EarthVenusAway", func() float64 { return basic.EarthVenusAway(ttJD) }, func() float64 { return basic.EarthVenusAwayN(ttJD, -1) }},
		{"VenusMag", func() float64 { return basic.VenusMag(ttJD) }, func() float64 { return basic.VenusMagN(ttJD, -1) }},
		{"VenusPhaseAngle", func() float64 { return basic.VenusPhaseAngle(ttJD) }, func() float64 { return basic.VenusPhaseAngleN(ttJD, -1) }},
		{"VenusIlluminatedFraction", func() float64 { return basic.VenusIlluminatedFraction(ttJD) }, func() float64 { return basic.VenusIlluminatedFractionN(ttJD, -1) }},
		{"VenusBrightLimbPositionAngle", func() float64 { return basic.VenusBrightLimbPositionAngle(ttJD) }, func() float64 { return basic.VenusBrightLimbPositionAngleN(ttJD, -1) }},
		{"VenusHeight", func() float64 { return basic.VenusHeight(jde, lon, lat, tz) }, func() float64 { return basic.VenusHeightN(jde, lon, lat, tz, -1) }},
		{"VenusAzimuth", func() float64 { return basic.VenusAzimuth(jde, lon, lat, tz) }, func() float64 { return basic.VenusAzimuthN(jde, lon, lat, tz, -1) }},
		{"VenusHourAngle", func() float64 { return basic.VenusHourAngle(jde, lon, tz) }, func() float64 { return basic.VenusHourAngleN(jde, lon, tz, -1) }},
		{"VenusCulminationTime", func() float64 { return basic.VenusCulminationTime(jde, lon, tz) }, func() float64 { return basic.VenusCulminationTimeN(jde, lon, tz, -1) }},
		{"VenusRiseTime", func() float64 { value, _ := basic.VenusRiseTime(jde, lon, lat, tz, 1, height); return value }, func() float64 { value, _ := basic.VenusRiseTimeN(jde, lon, lat, tz, 1, height, -1); return value }},
		{"VenusSetTime", func() float64 { value, _ := basic.VenusSetTime(jde, lon, lat, tz, 1, height); return value }, func() float64 { value, _ := basic.VenusSetTimeN(jde, lon, lat, tz, 1, height, -1); return value }},
		{"MarsApparentLo", func() float64 { return basic.MarsApparentLo(ttJD) }, func() float64 { return basic.MarsApparentLoN(ttJD, -1) }},
		{"MarsApparentBo", func() float64 { return basic.MarsApparentBo(ttJD) }, func() float64 { return basic.MarsApparentBoN(ttJD, -1) }},
		{"MarsApparentRa", func() float64 { return basic.MarsApparentRa(ttJD) }, func() float64 { return basic.MarsApparentRaN(ttJD, -1) }},
		{"MarsApparentDec", func() float64 { return basic.MarsApparentDec(ttJD) }, func() float64 { return basic.MarsApparentDecN(ttJD, -1) }},
		{"EarthMarsAway", func() float64 { return basic.EarthMarsAway(ttJD) }, func() float64 { return basic.EarthMarsAwayN(ttJD, -1) }},
		{"MarsMag", func() float64 { return basic.MarsMag(ttJD) }, func() float64 { return basic.MarsMagN(ttJD, -1) }},
		{"MarsPhaseAngle", func() float64 { return basic.MarsPhaseAngle(ttJD) }, func() float64 { return basic.MarsPhaseAngleN(ttJD, -1) }},
		{"MarsIlluminatedFraction", func() float64 { return basic.MarsIlluminatedFraction(ttJD) }, func() float64 { return basic.MarsIlluminatedFractionN(ttJD, -1) }},
		{"MarsBrightLimbPositionAngle", func() float64 { return basic.MarsBrightLimbPositionAngle(ttJD) }, func() float64 { return basic.MarsBrightLimbPositionAngleN(ttJD, -1) }},
		{"MarsHeight", func() float64 { return basic.MarsHeight(jde, lon, lat, tz) }, func() float64 { return basic.MarsHeightN(jde, lon, lat, tz, -1) }},
		{"MarsAzimuth", func() float64 { return basic.MarsAzimuth(jde, lon, lat, tz) }, func() float64 { return basic.MarsAzimuthN(jde, lon, lat, tz, -1) }},
		{"MarsHourAngle", func() float64 { return basic.MarsHourAngle(jde, lon, tz) }, func() float64 { return basic.MarsHourAngleN(jde, lon, tz, -1) }},
		{"MarsCulminationTime", func() float64 { return basic.MarsCulminationTime(jde, lon, tz) }, func() float64 { return basic.MarsCulminationTimeN(jde, lon, tz, -1) }},
		{"MarsRiseTime", func() float64 { value, _ := basic.MarsRiseTime(jde, lon, lat, tz, 1, height); return value }, func() float64 { value, _ := basic.MarsRiseTimeN(jde, lon, lat, tz, 1, height, -1); return value }},
		{"MarsSetTime", func() float64 { value, _ := basic.MarsSetTime(jde, lon, lat, tz, 1, height); return value }, func() float64 { value, _ := basic.MarsSetTimeN(jde, lon, lat, tz, 1, height, -1); return value }},
		{"JupiterApparentLo", func() float64 { return basic.JupiterApparentLo(ttJD) }, func() float64 { return basic.JupiterApparentLoN(ttJD, -1) }},
		{"JupiterApparentBo", func() float64 { return basic.JupiterApparentBo(ttJD) }, func() float64 { return basic.JupiterApparentBoN(ttJD, -1) }},
		{"JupiterApparentRa", func() float64 { return basic.JupiterApparentRa(ttJD) }, func() float64 { return basic.JupiterApparentRaN(ttJD, -1) }},
		{"JupiterApparentDec", func() float64 { return basic.JupiterApparentDec(ttJD) }, func() float64 { return basic.JupiterApparentDecN(ttJD, -1) }},
		{"EarthJupiterAway", func() float64 { return basic.EarthJupiterAway(ttJD) }, func() float64 { return basic.EarthJupiterAwayN(ttJD, -1) }},
		{"JupiterMag", func() float64 { return basic.JupiterMag(ttJD) }, func() float64 { return basic.JupiterMagN(ttJD, -1) }},
		{"JupiterPhaseAngle", func() float64 { return basic.JupiterPhaseAngle(ttJD) }, func() float64 { return basic.JupiterPhaseAngleN(ttJD, -1) }},
		{"JupiterIlluminatedFraction", func() float64 { return basic.JupiterIlluminatedFraction(ttJD) }, func() float64 { return basic.JupiterIlluminatedFractionN(ttJD, -1) }},
		{"JupiterBrightLimbPositionAngle", func() float64 { return basic.JupiterBrightLimbPositionAngle(ttJD) }, func() float64 { return basic.JupiterBrightLimbPositionAngleN(ttJD, -1) }},
		{"JupiterHeight", func() float64 { return basic.JupiterHeight(jde, lon, lat, tz) }, func() float64 { return basic.JupiterHeightN(jde, lon, lat, tz, -1) }},
		{"JupiterAzimuth", func() float64 { return basic.JupiterAzimuth(jde, lon, lat, tz) }, func() float64 { return basic.JupiterAzimuthN(jde, lon, lat, tz, -1) }},
		{"JupiterHourAngle", func() float64 { return basic.JupiterHourAngle(jde, lon, tz) }, func() float64 { return basic.JupiterHourAngleN(jde, lon, tz, -1) }},
		{"JupiterCulminationTime", func() float64 { return basic.JupiterCulminationTime(jde, lon, tz) }, func() float64 { return basic.JupiterCulminationTimeN(jde, lon, tz, -1) }},
		{"JupiterRiseTime", func() float64 { value, _ := basic.JupiterRiseTime(jde, lon, lat, tz, 1, height); return value }, func() float64 { value, _ := basic.JupiterRiseTimeN(jde, lon, lat, tz, 1, height, -1); return value }},
		{"JupiterSetTime", func() float64 { value, _ := basic.JupiterSetTime(jde, lon, lat, tz, 1, height); return value }, func() float64 { value, _ := basic.JupiterSetTimeN(jde, lon, lat, tz, 1, height, -1); return value }},
		{"SaturnApparentLo", func() float64 { return basic.SaturnApparentLo(ttJD) }, func() float64 { return basic.SaturnApparentLoN(ttJD, -1) }},
		{"SaturnApparentBo", func() float64 { return basic.SaturnApparentBo(ttJD) }, func() float64 { return basic.SaturnApparentBoN(ttJD, -1) }},
		{"SaturnApparentRa", func() float64 { return basic.SaturnApparentRa(ttJD) }, func() float64 { return basic.SaturnApparentRaN(ttJD, -1) }},
		{"SaturnApparentDec", func() float64 { return basic.SaturnApparentDec(ttJD) }, func() float64 { return basic.SaturnApparentDecN(ttJD, -1) }},
		{"EarthSaturnAway", func() float64 { return basic.EarthSaturnAway(ttJD) }, func() float64 { return basic.EarthSaturnAwayN(ttJD, -1) }},
		{"SaturnRingB", func() float64 { return basic.SaturnRingB(ttJD) }, func() float64 { return basic.SaturnRingBN(ttJD, -1) }},
		{"SaturnRingSunB", func() float64 { return basic.SaturnRingSunB(ttJD) }, func() float64 { return basic.SaturnRingSunBN(ttJD, -1) }},
		{"SaturnRingPositionAngle", func() float64 { return basic.SaturnRingPositionAngle(ttJD) }, func() float64 { return basic.SaturnRingPositionAngleN(ttJD, -1) }},
		{"SaturnRingDeltaU", func() float64 { return basic.SaturnRingDeltaU(ttJD) }, func() float64 { return basic.SaturnRingDeltaUN(ttJD, -1) }},
		{"SaturnMag", func() float64 { return basic.SaturnMag(ttJD) }, func() float64 { return basic.SaturnMagN(ttJD, -1) }},
		{"SaturnPhaseAngle", func() float64 { return basic.SaturnPhaseAngle(ttJD) }, func() float64 { return basic.SaturnPhaseAngleN(ttJD, -1) }},
		{"SaturnIlluminatedFraction", func() float64 { return basic.SaturnIlluminatedFraction(ttJD) }, func() float64 { return basic.SaturnIlluminatedFractionN(ttJD, -1) }},
		{"SaturnBrightLimbPositionAngle", func() float64 { return basic.SaturnBrightLimbPositionAngle(ttJD) }, func() float64 { return basic.SaturnBrightLimbPositionAngleN(ttJD, -1) }},
		{"SaturnHeight", func() float64 { return basic.SaturnHeight(jde, lon, lat, tz) }, func() float64 { return basic.SaturnHeightN(jde, lon, lat, tz, -1) }},
		{"SaturnAzimuth", func() float64 { return basic.SaturnAzimuth(jde, lon, lat, tz) }, func() float64 { return basic.SaturnAzimuthN(jde, lon, lat, tz, -1) }},
		{"SaturnHourAngle", func() float64 { return basic.SaturnHourAngle(jde, lon, tz) }, func() float64 { return basic.SaturnHourAngleN(jde, lon, tz, -1) }},
		{"SaturnCulminationTime", func() float64 { return basic.SaturnCulminationTime(jde, lon, tz) }, func() float64 { return basic.SaturnCulminationTimeN(jde, lon, tz, -1) }},
		{"SaturnRiseTime", func() float64 { value, _ := basic.SaturnRiseTime(jde, lon, lat, tz, 1, height); return value }, func() float64 { value, _ := basic.SaturnRiseTimeN(jde, lon, lat, tz, 1, height, -1); return value }},
		{"SaturnSetTime", func() float64 { value, _ := basic.SaturnSetTime(jde, lon, lat, tz, 1, height); return value }, func() float64 { value, _ := basic.SaturnSetTimeN(jde, lon, lat, tz, 1, height, -1); return value }},
		{"UranusApparentLo", func() float64 { return basic.UranusApparentLo(ttJD) }, func() float64 { return basic.UranusApparentLoN(ttJD, -1) }},
		{"UranusApparentBo", func() float64 { return basic.UranusApparentBo(ttJD) }, func() float64 { return basic.UranusApparentBoN(ttJD, -1) }},
		{"UranusApparentRa", func() float64 { return basic.UranusApparentRa(ttJD) }, func() float64 { return basic.UranusApparentRaN(ttJD, -1) }},
		{"UranusApparentDec", func() float64 { return basic.UranusApparentDec(ttJD) }, func() float64 { return basic.UranusApparentDecN(ttJD, -1) }},
		{"EarthUranusAway", func() float64 { return basic.EarthUranusAway(ttJD) }, func() float64 { return basic.EarthUranusAwayN(ttJD, -1) }},
		{"UranusMag", func() float64 { return basic.UranusMag(ttJD) }, func() float64 { return basic.UranusMagN(ttJD, -1) }},
		{"UranusPhaseAngle", func() float64 { return basic.UranusPhaseAngle(ttJD) }, func() float64 { return basic.UranusPhaseAngleN(ttJD, -1) }},
		{"UranusIlluminatedFraction", func() float64 { return basic.UranusIlluminatedFraction(ttJD) }, func() float64 { return basic.UranusIlluminatedFractionN(ttJD, -1) }},
		{"UranusBrightLimbPositionAngle", func() float64 { return basic.UranusBrightLimbPositionAngle(ttJD) }, func() float64 { return basic.UranusBrightLimbPositionAngleN(ttJD, -1) }},
		{"UranusHeight", func() float64 { return basic.UranusHeight(jde, lon, lat, tz) }, func() float64 { return basic.UranusHeightN(jde, lon, lat, tz, -1) }},
		{"UranusAzimuth", func() float64 { return basic.UranusAzimuth(jde, lon, lat, tz) }, func() float64 { return basic.UranusAzimuthN(jde, lon, lat, tz, -1) }},
		{"UranusHourAngle", func() float64 { return basic.UranusHourAngle(jde, lon, tz) }, func() float64 { return basic.UranusHourAngleN(jde, lon, tz, -1) }},
		{"UranusCulminationTime", func() float64 { return basic.UranusCulminationTime(jde, lon, tz) }, func() float64 { return basic.UranusCulminationTimeN(jde, lon, tz, -1) }},
		{"UranusRiseTime", func() float64 { value, _ := basic.UranusRiseTime(jde, lon, lat, tz, 1, height); return value }, func() float64 { value, _ := basic.UranusRiseTimeN(jde, lon, lat, tz, 1, height, -1); return value }},
		{"UranusSetTime", func() float64 { value, _ := basic.UranusSetTime(jde, lon, lat, tz, 1, height); return value }, func() float64 { value, _ := basic.UranusSetTimeN(jde, lon, lat, tz, 1, height, -1); return value }},
		{"NeptuneApparentLo", func() float64 { return basic.NeptuneApparentLo(ttJD) }, func() float64 { return basic.NeptuneApparentLoN(ttJD, -1) }},
		{"NeptuneApparentBo", func() float64 { return basic.NeptuneApparentBo(ttJD) }, func() float64 { return basic.NeptuneApparentBoN(ttJD, -1) }},
		{"NeptuneApparentRa", func() float64 { return basic.NeptuneApparentRa(ttJD) }, func() float64 { return basic.NeptuneApparentRaN(ttJD, -1) }},
		{"NeptuneApparentDec", func() float64 { return basic.NeptuneApparentDec(ttJD) }, func() float64 { return basic.NeptuneApparentDecN(ttJD, -1) }},
		{"EarthNeptuneAway", func() float64 { return basic.EarthNeptuneAway(ttJD) }, func() float64 { return basic.EarthNeptuneAwayN(ttJD, -1) }},
		{"NeptuneMag", func() float64 { return basic.NeptuneMag(ttJD) }, func() float64 { return basic.NeptuneMagN(ttJD, -1) }},
		{"NeptunePhaseAngle", func() float64 { return basic.NeptunePhaseAngle(ttJD) }, func() float64 { return basic.NeptunePhaseAngleN(ttJD, -1) }},
		{"NeptuneIlluminatedFraction", func() float64 { return basic.NeptuneIlluminatedFraction(ttJD) }, func() float64 { return basic.NeptuneIlluminatedFractionN(ttJD, -1) }},
		{"NeptuneBrightLimbPositionAngle", func() float64 { return basic.NeptuneBrightLimbPositionAngle(ttJD) }, func() float64 { return basic.NeptuneBrightLimbPositionAngleN(ttJD, -1) }},
		{"NeptuneHeight", func() float64 { return basic.NeptuneHeight(jde, lon, lat, tz) }, func() float64 { return basic.NeptuneHeightN(jde, lon, lat, tz, -1) }},
		{"NeptuneAzimuth", func() float64 { return basic.NeptuneAzimuth(jde, lon, lat, tz) }, func() float64 { return basic.NeptuneAzimuthN(jde, lon, lat, tz, -1) }},
		{"NeptuneHourAngle", func() float64 { return basic.NeptuneHourAngle(jde, lon, tz) }, func() float64 { return basic.NeptuneHourAngleN(jde, lon, tz, -1) }},
		{"NeptuneCulminationTime", func() float64 { return basic.NeptuneCulminationTime(jde, lon, tz) }, func() float64 { return basic.NeptuneCulminationTimeN(jde, lon, tz, -1) }},
		{"NeptuneRiseTime", func() float64 { value, _ := basic.NeptuneRiseTime(jde, lon, lat, tz, 1, height); return value }, func() float64 { value, _ := basic.NeptuneRiseTimeN(jde, lon, lat, tz, 1, height, -1); return value }},
		{"NeptuneSetTime", func() float64 { value, _ := basic.NeptuneSetTime(jde, lon, lat, tz, 1, height); return value }, func() float64 { value, _ := basic.NeptuneSetTimeN(jde, lon, lat, tz, 1, height, -1); return value }},
	}
	for _, tc := range floatChecks {
		assertSame(tc.name, tc.got(), tc.want())
	}

	errorChecks := []struct {
		name string
		got  func() error
		want func() error
	}{
		{"MercuryRiseTime.err", func() error { _, err := basic.MercuryRiseTime(jde, lon, lat, tz, 1, height); return err }, func() error { _, err := basic.MercuryRiseTimeN(jde, lon, lat, tz, 1, height, -1); return err }},
		{"MercurySetTime.err", func() error { _, err := basic.MercurySetTime(jde, lon, lat, tz, 1, height); return err }, func() error { _, err := basic.MercurySetTimeN(jde, lon, lat, tz, 1, height, -1); return err }},
		{"VenusRiseTime.err", func() error { _, err := basic.VenusRiseTime(jde, lon, lat, tz, 1, height); return err }, func() error { _, err := basic.VenusRiseTimeN(jde, lon, lat, tz, 1, height, -1); return err }},
		{"VenusSetTime.err", func() error { _, err := basic.VenusSetTime(jde, lon, lat, tz, 1, height); return err }, func() error { _, err := basic.VenusSetTimeN(jde, lon, lat, tz, 1, height, -1); return err }},
		{"MarsRiseTime.err", func() error { _, err := basic.MarsRiseTime(jde, lon, lat, tz, 1, height); return err }, func() error { _, err := basic.MarsRiseTimeN(jde, lon, lat, tz, 1, height, -1); return err }},
		{"MarsSetTime.err", func() error { _, err := basic.MarsSetTime(jde, lon, lat, tz, 1, height); return err }, func() error { _, err := basic.MarsSetTimeN(jde, lon, lat, tz, 1, height, -1); return err }},
		{"JupiterRiseTime.err", func() error { _, err := basic.JupiterRiseTime(jde, lon, lat, tz, 1, height); return err }, func() error { _, err := basic.JupiterRiseTimeN(jde, lon, lat, tz, 1, height, -1); return err }},
		{"JupiterSetTime.err", func() error { _, err := basic.JupiterSetTime(jde, lon, lat, tz, 1, height); return err }, func() error { _, err := basic.JupiterSetTimeN(jde, lon, lat, tz, 1, height, -1); return err }},
		{"SaturnRiseTime.err", func() error { _, err := basic.SaturnRiseTime(jde, lon, lat, tz, 1, height); return err }, func() error { _, err := basic.SaturnRiseTimeN(jde, lon, lat, tz, 1, height, -1); return err }},
		{"SaturnSetTime.err", func() error { _, err := basic.SaturnSetTime(jde, lon, lat, tz, 1, height); return err }, func() error { _, err := basic.SaturnSetTimeN(jde, lon, lat, tz, 1, height, -1); return err }},
		{"UranusRiseTime.err", func() error { _, err := basic.UranusRiseTime(jde, lon, lat, tz, 1, height); return err }, func() error { _, err := basic.UranusRiseTimeN(jde, lon, lat, tz, 1, height, -1); return err }},
		{"UranusSetTime.err", func() error { _, err := basic.UranusSetTime(jde, lon, lat, tz, 1, height); return err }, func() error { _, err := basic.UranusSetTimeN(jde, lon, lat, tz, 1, height, -1); return err }},
		{"NeptuneRiseTime.err", func() error { _, err := basic.NeptuneRiseTime(jde, lon, lat, tz, 1, height); return err }, func() error { _, err := basic.NeptuneRiseTimeN(jde, lon, lat, tz, 1, height, -1); return err }},
		{"NeptuneSetTime.err", func() error { _, err := basic.NeptuneSetTime(jde, lon, lat, tz, 1, height); return err }, func() error { _, err := basic.NeptuneSetTimeN(jde, lon, lat, tz, 1, height, -1); return err }},
	}
	for _, tc := range errorChecks {
		assertSameErr(tc.name, tc.got(), tc.want())
	}

	pairChecks := []struct {
		name string
		got  func() (float64, float64)
		want func() (float64, float64)
	}{
		{"MercuryApparentRaDec", func() (float64, float64) { return basic.MercuryApparentRaDec(ttJD) }, func() (float64, float64) { return basic.MercuryApparentRaDecN(ttJD, -1) }},
		{"VenusApparentRaDec", func() (float64, float64) { return basic.VenusApparentRaDec(ttJD) }, func() (float64, float64) { return basic.VenusApparentRaDecN(ttJD, -1) }},
		{"MarsApparentRaDec", func() (float64, float64) { return basic.MarsApparentRaDec(ttJD) }, func() (float64, float64) { return basic.MarsApparentRaDecN(ttJD, -1) }},
		{"JupiterApparentRaDec", func() (float64, float64) { return basic.JupiterApparentRaDec(ttJD) }, func() (float64, float64) { return basic.JupiterApparentRaDecN(ttJD, -1) }},
		{"SaturnApparentRaDec", func() (float64, float64) { return basic.SaturnApparentRaDec(ttJD) }, func() (float64, float64) { return basic.SaturnApparentRaDecN(ttJD, -1) }},
		{"SaturnRingAxis", func() (float64, float64) { return basic.SaturnRingAxis(ttJD) }, func() (float64, float64) { return basic.SaturnRingAxisN(ttJD, -1) }},
		{"UranusApparentRaDec", func() (float64, float64) { return basic.UranusApparentRaDec(ttJD) }, func() (float64, float64) { return basic.UranusApparentRaDecN(ttJD, -1) }},
		{"NeptuneApparentRaDec", func() (float64, float64) { return basic.NeptuneApparentRaDec(ttJD) }, func() (float64, float64) { return basic.NeptuneApparentRaDecN(ttJD, -1) }},
	}
	for _, tc := range pairChecks {
		got1, got2 := tc.got()
		want1, want2 := tc.want()
		assertSamePair(tc.name, got1, got2, want1, want2)
	}
}
