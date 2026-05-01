package basic

import (
	"time"
)

type outerPlanetEventFunc func(float64) float64

type outerPlanetEventCase struct {
	name      string
	tolerance float64
	fn        outerPlanetEventFunc
}

type outerPlanetEventPlan struct {
	planet       string
	baselineFile string
	samples      func() []time.Time
	phaseCases   []outerPlanetEventCase
	retroCases   []outerPlanetEventCase
}

func (plan outerPlanetEventPlan) allCases() []outerPlanetEventCase {
	cases := make([]outerPlanetEventCase, 0, len(plan.phaseCases)+len(plan.retroCases))
	cases = append(cases, plan.phaseCases...)
	cases = append(cases, plan.retroCases...)
	return cases
}

func outerPlanetEventSampleTTJD(date time.Time) float64 {
	return TD2UT(Date2JDE(date.UTC()), true)
}

func jupiterEventSamples() []time.Time {
	start := time.Date(1991, 3, 11, 5, 17, 23, 456000000, time.UTC)
	samples := make([]time.Time, 0, 96)
	for i := 0; i < 48; i++ {
		d := start.AddDate(0, 0, i*227)
		d = d.Add(time.Duration((i%8)*4)*time.Hour + time.Duration((i%11)*9)*time.Minute + time.Duration((i%13)*17)*time.Second)
		samples = append(samples, d)
	}
	extraStart := start.AddDate(0, 0, 37)
	for i := 0; i < 48; i++ {
		d := extraStart.AddDate(0, 0, i*149)
		d = d.Add(time.Duration((i%6)*7)*time.Hour + time.Duration((i%10)*13)*time.Minute + time.Duration((i%15)*23)*time.Second)
		samples = append(samples, d)
	}
	return samples
}

func saturnEventSamples() []time.Time {
	start := time.Date(1990, 7, 21, 8, 12, 34, 567000000, time.UTC)
	samples := make([]time.Time, 0, 96)
	for i := 0; i < 48; i++ {
		d := start.AddDate(0, 0, i*233)
		d = d.Add(time.Duration((i%9)*5)*time.Hour + time.Duration((i%10)*7)*time.Minute + time.Duration((i%12)*19)*time.Second)
		samples = append(samples, d)
	}
	extraStart := start.AddDate(0, 0, 43)
	for i := 0; i < 48; i++ {
		d := extraStart.AddDate(0, 0, i*157)
		d = d.Add(time.Duration((i%7)*6)*time.Hour + time.Duration((i%9)*11)*time.Minute + time.Duration((i%14)*29)*time.Second)
		samples = append(samples, d)
	}
	return samples
}

func uranusEventSamples() []time.Time {
	start := time.Date(1993, 11, 5, 10, 22, 33, 444000000, time.UTC)
	samples := make([]time.Time, 0, 96)
	for i := 0; i < 48; i++ {
		d := start.AddDate(0, 0, i*239)
		d = d.Add(time.Duration((i%7)*6)*time.Hour + time.Duration((i%9)*13)*time.Minute + time.Duration((i%14)*11)*time.Second)
		samples = append(samples, d)
	}
	extraStart := start.AddDate(0, 0, 59)
	for i := 0; i < 48; i++ {
		d := extraStart.AddDate(0, 0, i*163)
		d = d.Add(time.Duration((i%8)*5)*time.Hour + time.Duration((i%12)*17)*time.Minute + time.Duration((i%13)*31)*time.Second)
		samples = append(samples, d)
	}
	return samples
}

func neptuneEventSamples() []time.Time {
	start := time.Date(1996, 4, 17, 3, 14, 15, 926000000, time.UTC)
	samples := make([]time.Time, 0, 96)
	for i := 0; i < 48; i++ {
		d := start.AddDate(0, 0, i*241)
		d = d.Add(time.Duration((i%10)*3)*time.Hour + time.Duration((i%8)*17)*time.Minute + time.Duration((i%15)*7)*time.Second)
		samples = append(samples, d)
	}
	extraStart := start.AddDate(0, 0, 67)
	for i := 0; i < 48; i++ {
		d := extraStart.AddDate(0, 0, i*167)
		d = d.Add(time.Duration((i%9)*4)*time.Hour + time.Duration((i%11)*19)*time.Minute + time.Duration((i%14)*27)*time.Second)
		samples = append(samples, d)
	}
	return samples
}

func outerPlanetEventPlans() []outerPlanetEventPlan {
	const (
		conjunctionTolerance = 0.00001
		searchTolerance      = 30.0 / 86400.0
	)
	return []outerPlanetEventPlan{
		{
			planet:       "Jupiter",
			baselineFile: "testdata/jupiter_event_baseline.json",
			samples:      jupiterEventSamples,
			phaseCases: []outerPlanetEventCase{
				{name: "LastJupiterConjunction", tolerance: conjunctionTolerance, fn: LastJupiterConjunction},
				{name: "NextJupiterConjunction", tolerance: conjunctionTolerance, fn: NextJupiterConjunction},
				{name: "LastJupiterOpposition", tolerance: conjunctionTolerance, fn: LastJupiterOpposition},
				{name: "NextJupiterOpposition", tolerance: conjunctionTolerance, fn: NextJupiterOpposition},
				{name: "LastJupiterEasternQuadrature", tolerance: conjunctionTolerance, fn: LastJupiterEasternQuadrature},
				{name: "NextJupiterEasternQuadrature", tolerance: conjunctionTolerance, fn: NextJupiterEasternQuadrature},
				{name: "LastJupiterWesternQuadrature", tolerance: conjunctionTolerance, fn: LastJupiterWesternQuadrature},
				{name: "NextJupiterWesternQuadrature", tolerance: conjunctionTolerance, fn: NextJupiterWesternQuadrature},
			},
			retroCases: []outerPlanetEventCase{
				{name: "LastJupiterProgradeToRetrograde", tolerance: searchTolerance, fn: LastJupiterProgradeToRetrograde},
				{name: "NextJupiterProgradeToRetrograde", tolerance: searchTolerance, fn: NextJupiterProgradeToRetrograde},
				{name: "LastJupiterRetrogradeToPrograde", tolerance: searchTolerance, fn: LastJupiterRetrogradeToPrograde},
				{name: "NextJupiterRetrogradeToPrograde", tolerance: searchTolerance, fn: NextJupiterRetrogradeToPrograde},
			},
		},
		{
			planet:       "Saturn",
			baselineFile: "testdata/saturn_event_baseline.json",
			samples:      saturnEventSamples,
			phaseCases: []outerPlanetEventCase{
				{name: "LastSaturnConjunction", tolerance: conjunctionTolerance, fn: LastSaturnConjunction},
				{name: "NextSaturnConjunction", tolerance: conjunctionTolerance, fn: NextSaturnConjunction},
				{name: "LastSaturnOpposition", tolerance: conjunctionTolerance, fn: LastSaturnOpposition},
				{name: "NextSaturnOpposition", tolerance: conjunctionTolerance, fn: NextSaturnOpposition},
				{name: "LastSaturnEasternQuadrature", tolerance: conjunctionTolerance, fn: LastSaturnEasternQuadrature},
				{name: "NextSaturnEasternQuadrature", tolerance: conjunctionTolerance, fn: NextSaturnEasternQuadrature},
				{name: "LastSaturnWesternQuadrature", tolerance: conjunctionTolerance, fn: LastSaturnWesternQuadrature},
				{name: "NextSaturnWesternQuadrature", tolerance: conjunctionTolerance, fn: NextSaturnWesternQuadrature},
			},
			retroCases: []outerPlanetEventCase{
				{name: "LastSaturnProgradeToRetrograde", tolerance: searchTolerance, fn: LastSaturnProgradeToRetrograde},
				{name: "NextSaturnProgradeToRetrograde", tolerance: searchTolerance, fn: NextSaturnProgradeToRetrograde},
				{name: "LastSaturnRetrogradeToPrograde", tolerance: searchTolerance, fn: LastSaturnRetrogradeToPrograde},
				{name: "NextSaturnRetrogradeToPrograde", tolerance: searchTolerance, fn: NextSaturnRetrogradeToPrograde},
			},
		},
		{
			planet:       "Uranus",
			baselineFile: "testdata/uranus_event_baseline.json",
			samples:      uranusEventSamples,
			phaseCases: []outerPlanetEventCase{
				{name: "LastUranusConjunction", tolerance: conjunctionTolerance, fn: LastUranusConjunction},
				{name: "NextUranusConjunction", tolerance: conjunctionTolerance, fn: NextUranusConjunction},
				{name: "LastUranusOpposition", tolerance: conjunctionTolerance, fn: LastUranusOpposition},
				{name: "NextUranusOpposition", tolerance: conjunctionTolerance, fn: NextUranusOpposition},
				{name: "LastUranusEasternQuadrature", tolerance: conjunctionTolerance, fn: LastUranusEasternQuadrature},
				{name: "NextUranusEasternQuadrature", tolerance: conjunctionTolerance, fn: NextUranusEasternQuadrature},
				{name: "LastUranusWesternQuadrature", tolerance: conjunctionTolerance, fn: LastUranusWesternQuadrature},
				{name: "NextUranusWesternQuadrature", tolerance: conjunctionTolerance, fn: NextUranusWesternQuadrature},
			},
			retroCases: []outerPlanetEventCase{
				{name: "LastUranusProgradeToRetrograde", tolerance: searchTolerance, fn: LastUranusProgradeToRetrograde},
				{name: "NextUranusProgradeToRetrograde", tolerance: searchTolerance, fn: NextUranusProgradeToRetrograde},
				{name: "LastUranusRetrogradeToPrograde", tolerance: searchTolerance, fn: LastUranusRetrogradeToPrograde},
				{name: "NextUranusRetrogradeToPrograde", tolerance: searchTolerance, fn: NextUranusRetrogradeToPrograde},
			},
		},
		{
			planet:       "Neptune",
			baselineFile: "testdata/neptune_event_baseline.json",
			samples:      neptuneEventSamples,
			phaseCases: []outerPlanetEventCase{
				{name: "LastNeptuneConjunction", tolerance: conjunctionTolerance, fn: LastNeptuneConjunction},
				{name: "NextNeptuneConjunction", tolerance: conjunctionTolerance, fn: NextNeptuneConjunction},
				{name: "LastNeptuneOpposition", tolerance: conjunctionTolerance, fn: LastNeptuneOpposition},
				{name: "NextNeptuneOpposition", tolerance: conjunctionTolerance, fn: NextNeptuneOpposition},
				{name: "LastNeptuneEasternQuadrature", tolerance: conjunctionTolerance, fn: LastNeptuneEasternQuadrature},
				{name: "NextNeptuneEasternQuadrature", tolerance: conjunctionTolerance, fn: NextNeptuneEasternQuadrature},
				{name: "LastNeptuneWesternQuadrature", tolerance: conjunctionTolerance, fn: LastNeptuneWesternQuadrature},
				{name: "NextNeptuneWesternQuadrature", tolerance: conjunctionTolerance, fn: NextNeptuneWesternQuadrature},
			},
			retroCases: []outerPlanetEventCase{
				{name: "LastNeptuneProgradeToRetrograde", tolerance: searchTolerance, fn: LastNeptuneProgradeToRetrograde},
				{name: "NextNeptuneProgradeToRetrograde", tolerance: searchTolerance, fn: NextNeptuneProgradeToRetrograde},
				{name: "LastNeptuneRetrogradeToPrograde", tolerance: searchTolerance, fn: LastNeptuneRetrogradeToPrograde},
				{name: "NextNeptuneRetrogradeToPrograde", tolerance: searchTolerance, fn: NextNeptuneRetrogradeToPrograde},
			},
		},
	}
}
