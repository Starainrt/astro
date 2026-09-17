package basic

import (
	"math"
	"testing"
)

func TestMoonSeriesNumericalPrecision(t *testing.T) {
	// The existing binary64 coefficients evaluated at 70 decimal digits.
	// These check arithmetic precision, not the lunar theory's physical error.
	for _, sample := range []struct{ jd, longitude, latitude, distance, velocity float64 }{
		{990647.125, 48.68531043735905108, -3.208940888099671943, 365328.8143785083007, 14.53542662716284001},
		{1730647.9333686847, 346.3227699608293767, -2.151500042672265917, 368450.7410464923284, 14.30553896629878483},
		{2451545.0, 223.3189003731153751, 5.170885995272171723, 402448.7431750887777, 12.02142669184144213},
		{2460770.123456789, 99.20516643712787636, 5.156920731462777075, 378303.2056716878878, 13.57356557811076973},
		{2817000.99999, 256.0031142427467021, 3.438015761341575242, 402219.0739759512533, 11.99285384469081927},
		{3191874.3956256355, 178.1121054545695063, 1.37109140818496243, 379219.9689425243825, 13.41568168416747773},
		{3547272.75, 167.6849402044916958, 5.088592207858965847, 374590.3971822900352, 13.85120063917007945},
	} {
		longitude := HMoonTrueLo(sample.jd)
		latitude := HMoonTrueBo(sample.jd)
		distance := HMoonAway(sample.jd)
		if error := math.Abs(math.Remainder(longitude-sample.longitude, 360)); error > 2e-11 {
			t.Errorf("jd=%.12f longitude error=%g degrees", sample.jd, error)
		}
		if error := math.Abs(latitude - sample.latitude); error > 2e-11 {
			t.Errorf("jd=%.12f latitude error=%g degrees", sample.jd, error)
		}
		if error := math.Abs(distance - sample.distance); error > 1e-5 {
			t.Errorf("jd=%.12f distance error=%g km", sample.jd, error)
		}
		const step = 5.0 / 86400
		before, after := sample.jd-step, sample.jd+step
		velocity := math.Remainder(HMoonTrueLo(after)-HMoonTrueLo(before), 360) / (after - before)
		if error := math.Abs(velocity - sample.velocity); error > 2e-7 {
			t.Errorf("jd=%.12f velocity error=%g degrees/day", sample.jd, error)
		}
	}
}

func TestOccultationContactDerivativeNumericalStability(t *testing.T) {
	config, _ := planetOccultationConfigFor(OccultationMercury)
	cache := newPlanetOccultationEventCache(config)
	const tt = 1730647.933368684724
	const longitude = 90.88806942770216
	const latitude = 29.603106445560446
	evaluations := newOccultationRiseSetEvaluationCache(cache.riseSetContextAt)
	derivative := func(jd float64) float64 {
		return evaluations.evaluation(jd).contactDerivative(longitude, latitude)
	}
	center := derivative(tt)
	const trendStep = 0.5 / 86400
	trend := (derivative(tt+trendStep) - derivative(tt-trendStep)) / (2 * trendStep)
	ulp := math.Nextafter(tt, math.Inf(1)) - tt
	for i := -16; i <= 16; i++ {
		offset := float64(i) * ulp
		jitter := derivative(tt+offset) - center - trend*offset
		if math.Abs(jitter) > occultationRiseSetJunctionDerivativeTolerance {
			t.Fatalf("offset=%g seconds: derivative jitter=%g degrees/day", offset*86400, jitter)
		}
	}
	cache.preparePathEphemeris(tt, OccultationPathAlgorithmOptimized)
	if cache.local == nil || !cache.local.dense {
		t.Fatal("smooth exact states must support the checked dense ephemeris")
	}
}

// MoonCalcNew 与 HMoonTrueLo 是同一个物理量的两个入口，必须给出同一个值（此前只有一个入口
// 走线性相位补偿，两者差约 1e-9 度）。
func TestMoonCalcNewMatchesHMoonTrueLo(t *testing.T) {
	for _, jd := range []float64{2451545.0, 2460310.5, 2415020.5, 2299160.5, 2500000.5} {
		direct := math.Mod(MoonCalcNew(0, jd)*180/math.Pi, 360)
		if direct < 0 {
			direct += 360
		}
		high := HMoonTrueLo(jd)
		if diff := math.Abs(direct - high); diff > 1e-12 {
			t.Errorf("jd %.1f: MoonCalcNew=%0.15f HMoonTrueLo=%0.15f diff=%.3g deg", jd, direct, high, diff)
		}
	}
}
