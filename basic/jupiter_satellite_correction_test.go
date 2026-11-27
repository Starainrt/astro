package basic

import (
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"
)

type jupiterGalileanJUP365Window struct {
	Label     string                        `json:"label"`
	Satellite int                           `json:"satellite"`
	Source    string                        `json:"source"`
	QueryURLs []string                      `json:"queryURLs"`
	Samples   []jupiterGalileanJUP365Vector `json:"samples"`
}

type jupiterGalileanJUP365Vector struct {
	JD float64 `json:"jd"`
	X  float64 `json:"x"`
	Y  float64 `json:"y"`
	Z  float64 `json:"z"`
	VX float64 `json:"vx"`
	VY float64 `json:"vy"`
	VZ float64 `json:"vz"`
}

var jupiterGalileanCorrectionPerformanceSink JupiterGalileanState

func TestJupiterGalileanCorrectionExternalHoldout(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "jupiter", "testdata", "galilean_jup365_holdout.json"))
	if err != nil {
		t.Fatal(err)
	}
	var curves []jupiterGalileanFitCurve
	if err := json.Unmarshal(data, &curves); err != nil {
		t.Fatal(err)
	}
	var modernSquares, modernMaximum [4]float64
	var modernCounts [4]int
	type centuryMetrics struct {
		pureSquares      float64
		correctedSquares float64
		count            int
	}
	centuries := [4]map[int]centuryMetrics{}
	fitSpanMetrics := [4]centuryMetrics{}
	for satellite := range centuries {
		centuries[satellite] = make(map[int]centuryMetrics)
	}
	for _, curve := range curves {
		if curve.Label != "holdout-modern" && curve.Label != "holdout-centuries" {
			continue
		}
		if curve.Satellite < 1 || curve.Satellite > 4 {
			t.Fatalf("invalid holdout satellite %d", curve.Satellite)
		}
		index := curve.Satellite - 1
		for _, sample := range curve.Samples {
			et := sample.JD - jupiterGalileanReferenceJD
			includeSolar := jupiterGalileanUseSolarLongPeriod(sample.JD)
			pure := jupiterGalileanSatelliteStateAtET(et, index, includeSolar)
			corrected := jupiterGalileanSatelliteCorrectedStateAtET(et, index, includeSolar)
			pureError := jupiterGalileanPositionErrorKM(pure, sample)
			correctedError := jupiterGalileanPositionErrorKM(corrected, sample)
			if curve.Label == "holdout-modern" {
				modernSquares[index] += correctedError * correctedError
				modernMaximum[index] = math.Max(modernMaximum[index], correctedError)
				modernCounts[index]++
				continue
			}
			year := 2000 + (sample.JD-2451545)/365.25
			century := int(math.Floor(year/100)) * 100
			metric := centuries[index][century]
			metric.pureSquares += pureError * pureError
			metric.correctedSquares += correctedError * correctedError
			metric.count++
			centuries[index][century] = metric
			if century >= 1900 && century < 2100 {
				fitSpanMetrics[index].pureSquares += pureError * pureError
				fitSpanMetrics[index].correctedSquares += correctedError * correctedError
				fitSpanMetrics[index].count++
			}
		}
	}
	for index := range modernCounts {
		if modernCounts[index] == 0 {
			t.Fatalf("satellite %d has no modern holdout samples", index+1)
		}
		rms := math.Sqrt(modernSquares[index] / float64(modernCounts[index]))
		limits := [4][2]float64{{20, 45}, {20, 45}, {35, 70}, {20, 45}}
		t.Logf("satellite %d modern holdout RMS=%.6f km maximum=%.6f km", index+1, rms, modernMaximum[index])
		if rms > limits[index][0] || modernMaximum[index] > limits[index][1] {
			t.Fatalf("satellite %d modern holdout exceeds limits: RMS %.6f/%.0f km maximum %.6f/%.0f km", index+1, rms, limits[index][0], modernMaximum[index], limits[index][1])
		}
	}
	for index := range centuries {
		for century, metric := range centuries[index] {
			pureRMS := math.Sqrt(metric.pureSquares / float64(metric.count))
			correctedRMS := math.Sqrt(metric.correctedSquares / float64(metric.count))
			t.Logf("satellite %d century %d-%d RMS pure=%.6f km corrected=%.6f km", index+1, century, century+100, pureRMS, correctedRMS)
			if (century < 1900 || century >= 2100) && correctedRMS > pureRMS+1e-9 {
				t.Fatalf("satellite %d century %d-%d correction worsens RMS: pure %.6f km corrected %.6f km", index+1, century, century+100, pureRMS, correctedRMS)
			}
		}
	}
	fitSpanLimits := [4]float64{40, 40, 60, 40}
	for index, metric := range fitSpanMetrics {
		if metric.count == 0 {
			t.Fatalf("satellite %d has no 1900-2100 holdout samples", index+1)
		}
		pureRMS := math.Sqrt(metric.pureSquares / float64(metric.count))
		correctedRMS := math.Sqrt(metric.correctedSquares / float64(metric.count))
		t.Logf("satellite %d 1900-2100 RMS pure=%.6f km corrected=%.6f km", index+1, pureRMS, correctedRMS)
		if correctedRMS > fitSpanLimits[index] {
			t.Fatalf("satellite %d 1900-2100 RMS %.6f km exceeds %.0f km", index+1, correctedRMS, fitSpanLimits[index])
		}
	}

	const iterations = 100000
	const rounds = 7
	const et = 2451545.0 - jupiterGalileanReferenceJD
	measure := func(index int, corrected bool) time.Duration {
		started := time.Now()
		for iteration := 0; iteration < iterations; iteration++ {
			if corrected {
				jupiterGalileanCorrectionPerformanceSink = jupiterGalileanSatelliteCorrectedStateAtET(et, index, true)
			} else {
				jupiterGalileanCorrectionPerformanceSink = jupiterGalileanSatelliteStateAtET(et, index, true)
			}
		}
		return time.Since(started)
	}
	for index := range jupiterGalileanCorrectionCounts {
		performanceIncreases := make([]float64, rounds)
		for round := range performanceIncreases {
			var pureElapsed, correctedElapsed time.Duration
			if round%2 == 0 {
				pureElapsed, correctedElapsed = measure(index, false), measure(index, true)
			} else {
				correctedElapsed, pureElapsed = measure(index, true), measure(index, false)
			}
			performanceIncreases[round] = float64(correctedElapsed-pureElapsed) / float64(pureElapsed)
		}
		sort.Float64s(performanceIncreases)
		performanceIncrease := performanceIncreases[len(performanceIncreases)/2]
		t.Logf("satellite %d corrected state performance: median increase=%.2f%% range=%.2f%%..%.2f%% rounds=%d iterations=%d", index+1, performanceIncrease*100, performanceIncreases[0]*100, performanceIncreases[len(performanceIncreases)-1]*100, rounds, iterations)
		if performanceIncrease > 0.2 {
			t.Errorf("satellite %d correction increases state evaluation time by %.2f%%, limit is 20%%", index+1, performanceIncrease*100)
		}
	}
}

func TestJupiterGalileanCorrectionExternalGrazingWindow(t *testing.T) {
	window := loadJupiterGalileanJUP365Window(t)
	minimumMetric := math.Inf(1)
	minimumJD := 0.0
	for jd := 2463144.0; jd <= 2463144.05; jd += 10.0 / 86400 {
		metric := jupiterGalileanJUP365TransitMetricAt(jd, window.Samples)
		if !isFinite(metric) {
			t.Fatalf("invalid JUP365 transit metric at JD %.12f: %v", jd, metric)
		}
		if metric < minimumMetric {
			minimumMetric, minimumJD = metric, jd
		}
	}
	t.Logf("JUP365 Callisto transit metric minimum=%.12f at JD %.12f", minimumMetric, minimumJD)
	if minimumMetric <= 1+1e-12 {
		t.Fatalf("JUP365 geometry contains a Callisto transit at legacy grazing window: metric=%.12f", minimumMetric)
	}
}

func TestJupiterGalileanCorrectionConsistencyStatePaths(t *testing.T) {
	for _, year := range []int{1400, 1700, 1950, 2000, 2026, 2100, 2300, 2800} {
		jd := JDCalc(year, 1, 1)
		states := JupiterGalileanSatelliteStates(jd)
		for index := range states {
			want := jupiterGalileanFitPureState(jd-jupiterGalileanReferenceJD, index, jupiterGalileanUseSolarLongPeriod(jd))
			if states[index] != want || JupiterGalileanSatelliteState(jd, index+1) != want {
				t.Fatalf("satellite %d year %d public state does not preserve pure L1", index+1, year)
			}
		}
	}
}

func TestJupiterGalileanCorrectionConsistencyEvaluation(t *testing.T) {
	maximumPositionKM, maximumVelocityKMPerDay := 0.0, 0.0
	for index, span := range jupiterGalileanCorrectionSpans {
		jupiterGalileanFoldedElementsAtET(0, index, true)
		foldedCount, expectedCount := 0, 0
		for _, blocks := range jupiterGalileanFoldedCorrections[index].series {
			for _, terms := range blocks {
				foldedCount += len(terms)
			}
		}
		for _, element := range []int{0, 1, 3} {
			terms := jupiterGalileanCorrections[index][element].Terms
			expectedCount += len(terms)
			if element == 0 {
				continue
			}
			partner := jupiterGalileanCorrections[index][element+1].Terms
			if len(terms) != len(partner) {
				t.Fatalf("satellite %d element %d paired period count differs", index+1, element)
			}
			for term, model := range terms {
				if model.Period != partner[term].Period {
					t.Fatalf("satellite %d element %d term %d paired period differs", index+1, element, term)
				}
			}
		}
		if foldedCount != expectedCount {
			t.Fatalf("satellite %d folded %d periods, want %d", index+1, foldedCount, expectedCount)
		}
		dates := []float64{span.StartJD - 36525, span.StartJD, span.EndJD, span.EndJD + 36525}
		for _, year := range []int{1400, 1600, 1800, 1900, 1950, 2000, 2026, 2100, 2300, 2600, 2800} {
			for sample := 0; sample < 257; sample++ {
				dates = append(dates, JDCalc(year, 1, 1)+float64(sample)*1.375)
			}
		}
		for _, jd := range dates {
			et := jd - jupiterGalileanReferenceJD
			for _, includeSolar := range []bool{false, true} {
				folded := jupiterGalileanSatelliteCorrectedStateAtET(et, index, includeSolar)
				direct := jupiterGalileanCorrectionDirectStateForTest(et, index, includeSolar)
				dx, dy, dz := folded.X-direct.X, folded.Y-direct.Y, folded.Z-direct.Z
				dvx, dvy, dvz := folded.VX-direct.VX, folded.VY-direct.VY, folded.VZ-direct.VZ
				positionKM := math.Sqrt(dx*dx+dy*dy+dz*dz) * astronomicalUnitKM
				velocityKMPerDay := math.Sqrt(dvx*dvx+dvy*dvy+dvz*dvz) * astronomicalUnitKM
				maximumPositionKM = math.Max(maximumPositionKM, positionKM)
				maximumVelocityKMPerDay = math.Max(maximumVelocityKMPerDay, velocityKMPerDay)
				if !isFinite(positionKM) || !isFinite(velocityKMPerDay) || positionKM > 0.001 || velocityKMPerDay > 0.001 {
					t.Fatalf("satellite %d at JD %.12f solar=%v folded evaluation differs: position %.12g km velocity %.12g km/day", index+1, jd, includeSolar, positionKM, velocityKMPerDay)
				}
			}
		}
	}
	t.Logf("folded versus direct maximum position=%.12g km velocity=%.12g km/day", maximumPositionKM, maximumVelocityKMPerDay)
}

func TestJupiterGalileanCorrectionConsistencyFoldedWindow(t *testing.T) {
	const positionToleranceAU = 1e-12
	const velocityToleranceAUPerDay = 1e-12
	maximumPositionAU, maximumVelocityAUPerDay := 0.0, 0.0
	for index, span := range jupiterGalileanCorrectionSpans {
		for _, boundary := range []struct {
			jd     float64
			weight float64
		}{
			{span.StartJD - 36525, 0},
			{span.StartJD, 1},
			{span.EndJD, 1},
			{span.EndJD + 36525, 0},
		} {
			for _, includeSolar := range []bool{false, true} {
				var actual, natural [2]JupiterGalileanState
				for side, et := range [2]float64{
					boundary.jd - jupiterGalileanReferenceJD - 0.001/86400,
					boundary.jd - jupiterGalileanReferenceJD + 0.001/86400,
				} {
					actual[side] = jupiterGalileanSatelliteCorrectedStateAtET(et, index, includeSolar)
					correction := jupiterGalileanCorrectionUnwindowedForTest(et, index)
					for element := range correction {
						correction[element] *= boundary.weight
					}
					natural[side] = jupiterGalileanCorrectionStateWithDeltaForTest(et, index, includeSolar, correction)
				}
				position := Vector3{
					(actual[1].X - actual[0].X) - (natural[1].X - natural[0].X),
					(actual[1].Y - actual[0].Y) - (natural[1].Y - natural[0].Y),
					(actual[1].Z - actual[0].Z) - (natural[1].Z - natural[0].Z),
				}
				velocity := Vector3{
					(actual[1].VX - actual[0].VX) - (natural[1].VX - natural[0].VX),
					(actual[1].VY - actual[0].VY) - (natural[1].VY - natural[0].VY),
					(actual[1].VZ - actual[0].VZ) - (natural[1].VZ - natural[0].VZ),
				}
				positionAU, velocityAUPerDay := vectorMagnitude(position), vectorMagnitude(velocity)
				maximumPositionAU = math.Max(maximumPositionAU, positionAU)
				maximumVelocityAUPerDay = math.Max(maximumVelocityAUPerDay, velocityAUPerDay)
				if !isFinite(positionAU) || !isFinite(velocityAUPerDay) || positionAU > positionToleranceAU || velocityAUPerDay > velocityToleranceAUPerDay {
					t.Errorf("satellite %d boundary %.1f solar=%v folded splice residual position=%.12g AU velocity=%.12g AU/day", index+1, boundary.jd, includeSolar, positionAU, velocityAUPerDay)
				}
			}
		}
	}
	t.Logf("folded window maximum splice residual position=%.12g km velocity=%.12g km/day", maximumPositionAU*astronomicalUnitKM, maximumVelocityAUPerDay*astronomicalUnitKM)
}

func jupiterGalileanCorrectionDirectStateForTest(et float64, index int, includeSolar bool) JupiterGalileanState {
	return jupiterGalileanCorrectionStateWithDeltaForTest(et, index, includeSolar, jupiterGalileanCorrectionDirectForTest(et, index))
}

func jupiterGalileanCorrectionStateWithDeltaForTest(et float64, index int, includeSolar bool, correction [5]float64) JupiterGalileanState {
	longPeriod := jupiterGalileanEvaluateSeries(jupiterGalileanL1LongPeriodTerms[index], et+jupiterGalileanLongPeriodShift, et, includeSolar, index)
	crossPeriod := jupiterGalileanEvaluateSeries(jupiterGalileanL1CrossPeriodTerms[index], et, et, false, index)
	elements := jupiterGalileanElements{
		A: longPeriod.A + crossPeriod.A,
		L: longPeriod.L + crossPeriod.L + jupiterGalileanBaseMeanLongitudes[index]*et,
		K: longPeriod.K + crossPeriod.K, H: longPeriod.H + crossPeriod.H,
		Q: longPeriod.Q + crossPeriod.Q, P: longPeriod.P + crossPeriod.P,
	}
	jupiterGalileanFitApply(&elements, correction)
	elements.L = math.Atan2(math.Sin(elements.L), math.Cos(elements.L))
	if elements.L < 0 {
		elements.L += 2 * math.Pi
	}
	return jupiterGalileanStateFromElements(elements, index)
}

func jupiterGalileanCorrectionDirectForTest(et float64, index int) [5]float64 {
	var correction [5]float64
	span := jupiterGalileanCorrectionSpans[index]
	weight := span.window(et)
	if weight == 0 {
		return correction
	}
	tau := span.normalizedTime(et)
	for element, model := range jupiterGalileanCorrections[index] {
		value := model.Polynomial[0] + model.Polynomial[1]*tau + model.Polynomial[2]*(2*tau*tau-1)
		for _, term := range model.Terms {
			sine, cosine := math.Sincos(2 * math.Pi * et / term.Period)
			value += term.Sin*sine + term.Cos*cosine
		}
		correction[element] = weight * value
	}
	return correction
}

var jupiterGalileanCorrectionKernelSink [5]float64

func BenchmarkJupiterGalileanCorrection(b *testing.B) {
	for index, name := range []string{"Io", "Europa", "Ganymede", "Callisto"} {
		b.Run(name, func(b *testing.B) {
			for _, mode := range []string{"PureState", "DirectCorrectedState", "CorrectedState", "DirectKernel"} {
				b.Run(mode, func(b *testing.B) {
					b.ReportAllocs()
					for iteration := 0; iteration < b.N; iteration++ {
						et := 2451545 - jupiterGalileanReferenceJD + float64(iteration%8192)/64
						switch mode {
						case "PureState":
							jupiterGalileanCorrectionPerformanceSink = jupiterGalileanSatelliteStateAtET(et, index, true)
						case "CorrectedState":
							jupiterGalileanCorrectionPerformanceSink = jupiterGalileanSatelliteCorrectedStateAtET(et, index, true)
						case "DirectCorrectedState":
							jupiterGalileanCorrectionPerformanceSink = jupiterGalileanCorrectionDirectStateForTest(et, index, true)
						case "DirectKernel":
							jupiterGalileanCorrectionKernelSink = jupiterGalileanCorrectionDirectForTest(et, index)
						}
					}
				})
			}
		})
	}
}

func loadJupiterGalileanJUP365Window(t *testing.T) jupiterGalileanJUP365Window {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", "galilean_callisto_jup365_2031.json"))
	if err != nil {
		t.Fatal(err)
	}
	var window jupiterGalileanJUP365Window
	if err := json.Unmarshal(data, &window); err != nil {
		t.Fatal(err)
	}
	if window.Satellite != 4 || len(window.Samples) < 2 || !strings.Contains(window.Source, "jup365_merged") || len(window.QueryURLs) != 1 {
		t.Fatalf("invalid JUP365 event window: satellite=%d samples=%d source=%q urls=%d", window.Satellite, len(window.Samples), window.Source, len(window.QueryURLs))
	}
	for index, sample := range window.Samples {
		if index > 0 && sample.JD <= window.Samples[index-1].JD {
			t.Fatalf("JUP365 samples are not ordered at row %d", index)
		}
	}
	return window
}

func jupiterGalileanJUP365TransitMetricAt(jde float64, samples []jupiterGalileanJUP365Vector) float64 {
	context := newJupiterGalileanObservationContext(jde)
	lightTime := context.jupiterLightTime
	geocentric := Vector3{}
	for iteration := 0; iteration < 8; iteration++ {
		targetJDE := jde - lightTime
		satellite := jupiterGalileanInterpolateJUP365(samples, targetJDE)
		jupiter := rotateEclipticToEquatorial(jupiterHeliocentricVectorJ2000(targetJDE, context.inst), orbitJ2000Obliquity)
		geocentric = Vector3{
			jupiter[0] + satellite[0] - context.earthHelioJ2000[0],
			jupiter[1] + satellite[1] - context.earthHelioJ2000[1],
			jupiter[2] + satellite[2] - context.earthHelioJ2000[2],
		}
		nextLightTime := lightTimeDaysPerAU * vectorMagnitude(geocentric)
		if math.Abs(nextLightTime-lightTime) < 1e-12 {
			break
		}
		lightTime = nextLightTime
	}
	relative := Vector3{
		geocentric[0] - context.jupiterGeoJ2000[0],
		geocentric[1] - context.jupiterGeoJ2000[1],
		geocentric[2] - context.jupiterGeoJ2000[2],
	}
	if vectorDot(relative, context.lineOfSight) >= 0 {
		return math.Inf(1)
	}
	direction := normalizeVector(geocentric)
	offsetX, offsetY := tangentPlaneOffsetAngles(direction, context.lineOfSight, context.east, context.north)
	semidiameter := math.Atan2(jupiterGalileanEquatorialRadiusKM/astronomicalUnitKM, context.jupiterDistance) * deg * 3600
	x := offsetX * deg * 3600 / semidiameter
	y := offsetY * deg * 3600 / semidiameter
	return ellipseMetric(x, y, 1, context.earthMinorRadius)
}

func jupiterGalileanInterpolateJUP365(samples []jupiterGalileanJUP365Vector, jd float64) Vector3 {
	for index := 0; index+1 < len(samples); index++ {
		left, right := samples[index], samples[index+1]
		if jd < left.JD || jd > right.JD {
			continue
		}
		days := right.JD - left.JD
		fraction := (jd - left.JD) / days
		fraction2, fraction3 := fraction*fraction, fraction*fraction*fraction
		h00 := 2*fraction3 - 3*fraction2 + 1
		h10 := fraction3 - 2*fraction2 + fraction
		h01 := -2*fraction3 + 3*fraction2
		h11 := fraction3 - fraction2
		leftPosition := Vector3{left.X / astronomicalUnitKM, left.Y / astronomicalUnitKM, left.Z / astronomicalUnitKM}
		rightPosition := Vector3{right.X / astronomicalUnitKM, right.Y / astronomicalUnitKM, right.Z / astronomicalUnitKM}
		leftVelocity := Vector3{left.VX * 86400 / astronomicalUnitKM, left.VY * 86400 / astronomicalUnitKM, left.VZ * 86400 / astronomicalUnitKM}
		rightVelocity := Vector3{right.VX * 86400 / astronomicalUnitKM, right.VY * 86400 / astronomicalUnitKM, right.VZ * 86400 / astronomicalUnitKM}
		return Vector3{
			h00*leftPosition[0] + h10*days*leftVelocity[0] + h01*rightPosition[0] + h11*days*rightVelocity[0],
			h00*leftPosition[1] + h10*days*leftVelocity[1] + h01*rightPosition[1] + h11*days*rightVelocity[1],
			h00*leftPosition[2] + h10*days*leftVelocity[2] + h01*rightPosition[2] + h11*days*rightVelocity[2],
		}
	}
	return Vector3{math.NaN(), math.NaN(), math.NaN()}
}

func jupiterGalileanPositionErrorKM(state JupiterGalileanState, sample struct{ JD, X, Y, Z float64 }) float64 {
	model := jupiterGalileanPositionInFitFrame(state)
	target := jupiterGalileanFitTarget([3]float64{sample.X, sample.Y, sample.Z})
	dx := (model[0] - target[0]) * astronomicalUnitKM
	dy := (model[1] - target[1]) * astronomicalUnitKM
	dz := (model[2] - target[2]) * astronomicalUnitKM
	return math.Sqrt(dx*dx + dy*dy + dz*dz)
}

func jupiterGalileanPositionInFitFrame(state JupiterGalileanState) [3]float64 {
	cosNode, sinNode := math.Cos(jupiterGalileanFrameNode), math.Sin(jupiterGalileanFrameNode)
	cosTilt, sinTilt := math.Cos(jupiterGalileanFrameTilt), math.Sin(jupiterGalileanFrameTilt)
	x := state.X*cosNode + state.Y*sinNode
	yRotated := -state.X*sinNode + state.Y*cosNode
	y := yRotated*cosTilt + state.Z*sinTilt
	z := -yRotated*sinTilt + state.Z*cosTilt
	return [3]float64{x, y, z}
}
