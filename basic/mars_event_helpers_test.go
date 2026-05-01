package basic

import "time"

type marsEventFunc func(float64) float64

type marsEventCase struct {
	name      string
	tolerance float64
	fn        marsEventFunc
}

func marsEventSamples() []time.Time {
	start := time.Date(1992, 8, 17, 9, 21, 45, 678000000, time.UTC)
	samples := make([]time.Time, 0, 96)
	for i := 0; i < 48; i++ {
		d := start.AddDate(0, 0, i*311)
		d = d.Add(time.Duration((i%8)*4)*time.Hour + time.Duration((i%10)*9)*time.Minute + time.Duration((i%12)*17)*time.Second)
		samples = append(samples, d)
	}
	extraStart := start.AddDate(0, 0, 53)
	for i := 0; i < 48; i++ {
		d := extraStart.AddDate(0, 0, i*197)
		d = d.Add(time.Duration((i%7)*5)*time.Hour + time.Duration((i%9)*14)*time.Minute + time.Duration((i%16)*23)*time.Second)
		samples = append(samples, d)
	}
	return samples
}

func marsEventSampleTTJD(date time.Time) float64 {
	return TD2UT(Date2JDE(date.UTC()), true)
}

func marsEventCases() []marsEventCase {
	const (
		conjunctionTolerance = 0.00001
		searchTolerance      = 30.0 / 86400.0
	)
	return []marsEventCase{
		{name: "LastMarsConjunction", tolerance: conjunctionTolerance, fn: LastMarsConjunction},
		{name: "NextMarsConjunction", tolerance: conjunctionTolerance, fn: NextMarsConjunction},
		{name: "LastMarsOpposition", tolerance: conjunctionTolerance, fn: LastMarsOpposition},
		{name: "NextMarsOpposition", tolerance: conjunctionTolerance, fn: NextMarsOpposition},
		{name: "LastMarsEasternQuadrature", tolerance: conjunctionTolerance, fn: LastMarsEasternQuadrature},
		{name: "NextMarsEasternQuadrature", tolerance: conjunctionTolerance, fn: NextMarsEasternQuadrature},
		{name: "LastMarsWesternQuadrature", tolerance: conjunctionTolerance, fn: LastMarsWesternQuadrature},
		{name: "NextMarsWesternQuadrature", tolerance: conjunctionTolerance, fn: NextMarsWesternQuadrature},
		{name: "LastMarsProgradeToRetrograde", tolerance: searchTolerance, fn: LastMarsProgradeToRetrograde},
		{name: "NextMarsProgradeToRetrograde", tolerance: searchTolerance, fn: NextMarsProgradeToRetrograde},
		{name: "LastMarsRetrogradeToPrograde", tolerance: searchTolerance, fn: LastMarsRetrogradeToPrograde},
		{name: "NextMarsRetrogradeToPrograde", tolerance: searchTolerance, fn: NextMarsRetrogradeToPrograde},
	}
}
