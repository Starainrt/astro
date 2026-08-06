package basic

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"testing"
	"time"

	"github.com/starainrt/astro/tools"
)

type moonRiseSetExternalEvents struct {
	RiseUTC string `json:"rise_utc"`
	SetUTC  string `json:"set_utc"`
}

type moonRiseSetExternalSample struct {
	Site           string                    `json:"site"`
	DateUTC        string                    `json:"date_utc"`
	Longitude      float64                   `json:"longitude"`
	Latitude       float64                   `json:"latitude"`
	ObserverHeight float64                   `json:"observer_height_m"`
	Horizons       moonRiseSetExternalEvents `json:"jpl_horizons"`
	METNorway      moonRiseSetExternalEvents `json:"met_norway"`
	IMCCEMiriade   moonRiseSetExternalEvents `json:"imcce_miriade"`
}

type moonRiseSetExternalBaseline struct {
	SchemaVersion int `json:"schema_version"`
	Sources       map[string]struct {
		Provider string `json:"provider"`
		Model    string `json:"model"`
	} `json:"sources"`
	Samples []moonRiseSetExternalSample `json:"samples"`
}

type moonRiseSetErrorStats struct {
	Total time.Duration
	Max   time.Duration
	Count int
}

func (stats *moonRiseSetErrorStats) Add(value time.Duration) {
	stats.Total += value
	stats.Count++
	if value > stats.Max {
		stats.Max = value
	}
}

func (stats moonRiseSetErrorStats) Mean() time.Duration {
	if stats.Count == 0 {
		return 0
	}
	return stats.Total / time.Duration(stats.Count)
}

type moonRiseSetExternalTolerances struct {
	Horizons      time.Duration
	METNorway     time.Duration
	IMCCEMiriade  time.Duration
	HorizonsVsMET time.Duration
}

type moonRiseSetComparisonStats struct {
	CurrentHorizons    moonRiseSetErrorStats
	LegacyHorizons     moonRiseSetErrorStats
	CurrentMET         moonRiseSetErrorStats
	LegacyMET          moonRiseSetErrorStats
	CurrentIMCCE       moonRiseSetErrorStats
	LegacyIMCCE        moonRiseSetErrorStats
	HorizonsVsMET      moonRiseSetErrorStats
	HorizonsVsIMCCE    moonRiseSetErrorStats
	CurrentCloserJPL   int
	LegacyCloserJPL    int
	TiesJPL            int
	CurrentCloserMET   int
	LegacyCloserMET    int
	TiesMET            int
	CurrentCloserIMCCE int
	LegacyCloserIMCCE  int
	TiesIMCCE          int
}

func TestMoonRiseSetMatchesExternalBaselines(t *testing.T) {
	previousDeltaT := defDeltaTFn
	SetDeltaTFn(DefaultDeltaTv2)
	defer SetDeltaTFn(previousDeltaT)

	baseline := loadMoonRiseSetExternalBaseline(t)
	if baseline.SchemaVersion != 1 {
		t.Fatalf("unsupported baseline schema version %d", baseline.SchemaVersion)
	}
	if baseline.Sources["jpl_horizons"].Model != "DE441" {
		t.Fatalf("unexpected Horizons model %q", baseline.Sources["jpl_horizons"].Model)
	}
	if len(baseline.Samples) < 7 {
		t.Fatalf("external baseline has only %d samples", len(baseline.Samples))
	}

	tolerances := moonRiseSetExternalTolerances{
		Horizons:      2 * time.Second,
		METNorway:     90 * time.Second,
		IMCCEMiriade:  8 * time.Minute,
		HorizonsVsMET: 90 * time.Second,
	}
	var stats moonRiseSetComparisonStats

	for _, sample := range baseline.Samples {
		day, err := time.Parse("2006-01-02", sample.DateUTC)
		if err != nil {
			t.Fatalf("parse %s date %q: %v", sample.Site, sample.DateUTC, err)
		}
		jd := Date2JDE(day)
		currentRiseJD, err := GetMoonRiseTime(jd, sample.Longitude, sample.Latitude, 0, 1, sample.ObserverHeight)
		if err != nil {
			t.Fatalf("%s current moonrise: %v", sample.Site, err)
		}
		currentSetJD, err := GetMoonSetTime(jd, sample.Longitude, sample.Latitude, 0, 1, sample.ObserverHeight)
		if err != nil {
			t.Fatalf("%s current moonset: %v", sample.Site, err)
		}
		legacyRiseJD, err := legacyMoonRiseSetFromCurrent(currentRiseJD, sample.Longitude, sample.Latitude, 0, 1, sample.ObserverHeight)
		if err != nil {
			t.Fatalf("%s legacy moonrise: %v", sample.Site, err)
		}
		legacySetJD, err := legacyMoonRiseSetFromCurrent(currentSetJD, sample.Longitude, sample.Latitude, 0, 1, sample.ObserverHeight)
		if err != nil {
			t.Fatalf("%s legacy moonset: %v", sample.Site, err)
		}

		compareMoonRiseSetEvent(t, sample.Site+".rise", currentRiseJD, legacyRiseJD,
			sample.Horizons.RiseUTC, sample.METNorway.RiseUTC, sample.IMCCEMiriade.RiseUTC,
			tolerances, &stats)
		compareMoonRiseSetEvent(t, sample.Site+".set", currentSetJD, legacySetJD,
			sample.Horizons.SetUTC, sample.METNorway.SetUTC, sample.IMCCEMiriade.SetUTC,
			tolerances, &stats)
	}

	t.Logf("moon rise/set external baseline: current vs JPL mean=%v max=%v; legacy vs JPL mean=%v max=%v",
		stats.CurrentHorizons.Mean(), stats.CurrentHorizons.Max, stats.LegacyHorizons.Mean(), stats.LegacyHorizons.Max)
	t.Logf("moon rise/set external baseline: current vs MET mean=%v max=%v; legacy vs MET mean=%v max=%v",
		stats.CurrentMET.Mean(), stats.CurrentMET.Max, stats.LegacyMET.Mean(), stats.LegacyMET.Max)
	t.Logf("moon rise/set external baseline: current vs IMCCE mean=%v max=%v; legacy vs IMCCE mean=%v max=%v",
		stats.CurrentIMCCE.Mean(), stats.CurrentIMCCE.Max, stats.LegacyIMCCE.Mean(), stats.LegacyIMCCE.Max)
	t.Logf("moon rise/set external baseline: JPL vs MET mean=%v max=%v; JPL vs IMCCE mean=%v max=%v",
		stats.HorizonsVsMET.Mean(), stats.HorizonsVsMET.Max, stats.HorizonsVsIMCCE.Mean(), stats.HorizonsVsIMCCE.Max)
	t.Logf("moon rise/set external baseline: JPL current closer=%d legacy closer=%d ties=%d",
		stats.CurrentCloserJPL, stats.LegacyCloserJPL, stats.TiesJPL)
	t.Logf("moon rise/set external baseline: MET current closer=%d legacy closer=%d ties=%d",
		stats.CurrentCloserMET, stats.LegacyCloserMET, stats.TiesMET)
	t.Logf("moon rise/set external baseline: IMCCE current closer=%d legacy closer=%d ties=%d",
		stats.CurrentCloserIMCCE, stats.LegacyCloserIMCCE, stats.TiesIMCCE)
}

func TestMoonRiseSetLegacyComparatorMatchesPreFixSnapshot(t *testing.T) {
	previousDeltaT := defDeltaTFn
	SetDeltaTFn(DefaultDeltaTv2)
	defer SetDeltaTFn(previousDeltaT)

	jd := JDECalc(2023, 1, 15)
	currentRise, err := GetMoonRiseTime(jd, 116.4074, 39.9042, 8, 1, 0)
	if err != nil {
		t.Fatalf("current moonrise: %v", err)
	}
	currentSet, err := GetMoonSetTime(jd, 116.4074, 39.9042, 8, 1, 0)
	if err != nil {
		t.Fatalf("current moonset: %v", err)
	}
	legacyRise, err := legacyMoonRiseSetFromCurrent(currentRise, 116.4074, 39.9042, 8, 1, 0)
	if err != nil {
		t.Fatalf("legacy moonrise: %v", err)
	}
	legacySet, err := legacyMoonRiseSetFromCurrent(currentSet, 116.4074, 39.9042, 8, 1, 0)
	if err != nil {
		t.Fatalf("legacy moonset: %v", err)
	}

	const snapshotTolerance = 2.0 / 86400
	if difference := math.Abs(legacyRise - 2459959.509182); difference > snapshotTolerance {
		t.Errorf("legacy moonrise snapshot mismatch: got %.9f want %.9f difference=%.3fs",
			legacyRise, 2459959.509182, difference*86400)
	}
	if difference := math.Abs(legacySet - 2459959.988676); difference > snapshotTolerance {
		t.Errorf("legacy moonset snapshot mismatch: got %.9f want %.9f difference=%.3fs",
			legacySet, 2459959.988676, difference*86400)
	}
}

func loadMoonRiseSetExternalBaseline(t *testing.T) moonRiseSetExternalBaseline {
	t.Helper()
	data, err := os.ReadFile("testdata/moon_rise_set_baseline.json")
	if err != nil {
		t.Fatalf("read moon rise/set baseline: %v", err)
	}
	var baseline moonRiseSetExternalBaseline
	if err := json.Unmarshal(data, &baseline); err != nil {
		t.Fatalf("decode moon rise/set baseline: %v", err)
	}
	return baseline
}

func compareMoonRiseSetEvent(t *testing.T, name string, currentJD, legacyJD float64,
	horizonsUTC, metUTC, imcceUTC string, tolerances moonRiseSetExternalTolerances,
	stats *moonRiseSetComparisonStats) {
	t.Helper()
	current := JDE2DateByZone(currentJD, time.UTC, false)
	legacy := JDE2DateByZone(legacyJD, time.UTC, false)
	horizons := parseMoonRiseSetExternalTime(t, name+".jpl", horizonsUTC)
	met := parseMoonRiseSetExternalTime(t, name+".met", metUTC)
	imcce := parseMoonRiseSetExternalTime(t, name+".imcce", imcceUTC)
	currentHorizonsError := absoluteTimeDifference(current, horizons)
	legacyHorizonsError := absoluteTimeDifference(legacy, horizons)
	currentMETError := absoluteTimeDifference(current, met)
	legacyMETError := absoluteTimeDifference(legacy, met)
	currentIMCCEError := absoluteTimeDifference(current, imcce)
	legacyIMCCEError := absoluteTimeDifference(legacy, imcce)
	stats.CurrentHorizons.Add(currentHorizonsError)
	stats.LegacyHorizons.Add(legacyHorizonsError)
	stats.CurrentMET.Add(currentMETError)
	stats.LegacyMET.Add(legacyMETError)
	stats.CurrentIMCCE.Add(currentIMCCEError)
	stats.LegacyIMCCE.Add(legacyIMCCEError)
	horizonsVsMET := absoluteTimeDifference(horizons, met)
	stats.HorizonsVsMET.Add(horizonsVsMET)
	stats.HorizonsVsIMCCE.Add(absoluteTimeDifference(horizons, imcce))

	if currentHorizonsError > tolerances.Horizons {
		t.Errorf("%s current mismatch against JPL: got %s want %s difference=%v tolerance=%v",
			name, current.Format(time.RFC3339Nano), horizonsUTC, currentHorizonsError, tolerances.Horizons)
	}
	if currentMETError > tolerances.METNorway {
		t.Errorf("%s current mismatch against MET Norway: got %s want %s difference=%v tolerance=%v",
			name, current.Format(time.RFC3339Nano), metUTC, currentMETError, tolerances.METNorway)
	}
	if currentIMCCEError > tolerances.IMCCEMiriade {
		t.Errorf("%s current mismatch against IMCCE Miriade: got %s want %s difference=%v tolerance=%v",
			name, current.Format(time.RFC3339Nano), imcceUTC, currentIMCCEError, tolerances.IMCCEMiriade)
	}
	if horizonsVsMET > tolerances.HorizonsVsMET {
		t.Errorf("%s external sources disagree: JPL=%s MET=%s difference=%v tolerance=%v",
			name, horizonsUTC, metUTC, horizonsVsMET, tolerances.HorizonsVsMET)
	}
	switch {
	case currentHorizonsError < legacyHorizonsError:
		stats.CurrentCloserJPL++
	case legacyHorizonsError < currentHorizonsError:
		stats.LegacyCloserJPL++
	default:
		stats.TiesJPL++
	}
	switch {
	case currentMETError < legacyMETError:
		stats.CurrentCloserMET++
	case legacyMETError < currentMETError:
		stats.LegacyCloserMET++
	default:
		stats.TiesMET++
	}
	switch {
	case currentIMCCEError < legacyIMCCEError:
		stats.CurrentCloserIMCCE++
	case legacyIMCCEError < currentIMCCEError:
		stats.LegacyCloserIMCCE++
	default:
		stats.TiesIMCCE++
	}
	t.Logf("%s current_jpl=%v legacy_jpl=%v current_met=%v legacy_met=%v current_imcce=%v legacy_imcce=%v", name,
		currentHorizonsError, legacyHorizonsError, currentMETError, legacyMETError, currentIMCCEError, legacyIMCCEError)
}

func parseMoonRiseSetExternalTime(t *testing.T, name, value string) time.Time {
	t.Helper()
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		t.Fatalf("parse %s time %q: %v", name, value, err)
	}
	return parsed
}

func absoluteTimeDifference(left, right time.Time) time.Duration {
	difference := left.Sub(right)
	if difference < 0 {
		return -difference
	}
	return difference
}

func legacyMoonRiseSetFromCurrent(currentJD, longitude, latitude, timeZone, zenithShift, height float64) (float64, error) {
	localTimeZone := longitude / 15
	localJD := currentJD + localTimeZone/24 - timeZone/24
	targetAltitude := StandardAltitudeMoon(zenithShift, height, latitude)
	legacyJD := moonRiseSetNewtonRaphsonIteration(localJD, longitude, latitude, localTimeZone,
		targetAltitude, legacyHMoonHeight, 0.00002)
	if math.IsNaN(legacyJD) || math.IsInf(legacyJD, 0) {
		return 0, fmt.Errorf("legacy height iteration did not converge")
	}
	return legacyJD - localTimeZone/24 + timeZone/24, nil
}

func legacyHMoonHeight(jd, longitude, latitude, timeZone float64) float64 {
	calculationJD := TD2UT(jd-timeZone/24, true)
	ra, dec := HMoonTrueRaDecN(calculationJD, -1)
	distanceAU := HMoonAwayN(calculationJD, -1) / 149597870.7
	topocentricRA, topocentricDec := legacyTopocentricRaDec(ra, dec, latitude, longitude, calculationJD, distanceAU, 0)
	siderealTime := tools.Limit360(ApparentSiderealTime(jd-timeZone/24)*15 + longitude)
	hourAngle := tools.Limit360(siderealTime - topocentricRA)
	altitudeSine := tools.Sin(latitude)*tools.Sin(topocentricDec) +
		tools.Cos(topocentricDec)*tools.Cos(latitude)*tools.Cos(hourAngle)
	return tools.ArcSin(altitudeSine)
}

func legacyTopocentricRaDec(ra, dec, latitude, longitude, jd, distanceAU, height float64) (float64, float64) {
	horizontalParallaxSine := tools.Sin(0.0024427777777) / distanceAU
	observerCosine := pcosi(latitude, height)
	observerSine := psini(latitude, height)
	hourAngle := tools.Limit360(TD2UT(ApparentSiderealTime(jd), false)*15 + longitude - ra)
	raCorrection := math.Atan2(-observerCosine*horizontalParallaxSine*tools.Sin(hourAngle),
		tools.Cos(dec)-observerCosine*horizontalParallaxSine*tools.Cos(hourAngle)) * 180 / math.Pi
	correctedDec := math.Atan2((tools.Sin(dec)-observerSine*horizontalParallaxSine)*tools.Cos(raCorrection),
		tools.Cos(dec)-observerCosine*horizontalParallaxSine*tools.Cos(hourAngle)) * 180 / math.Pi
	return ra + raCorrection, correctedDec
}
