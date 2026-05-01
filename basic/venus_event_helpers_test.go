package basic

import (
	"time"
)

type venusEventFunc func(float64) float64

type venusEventCase struct {
	name      string
	tolerance float64
	fn        venusEventFunc
}

func venusEventSamples() []time.Time {
	start := time.Date(1994, 2, 3, 4, 56, 12, 345000000, time.UTC)
	samples := make([]time.Time, 0, 96)
	for i := 0; i < 48; i++ {
		d := start.AddDate(0, 0, i*173)
		d = d.Add(time.Duration((i%9)*5)*time.Hour + time.Duration((i%7)*11)*time.Minute + time.Duration((i%13)*13)*time.Second)
		samples = append(samples, d)
	}
	extraStart := start.AddDate(0, 0, 29)
	for i := 0; i < 48; i++ {
		d := extraStart.AddDate(0, 0, i*127)
		d = d.Add(time.Duration((i%8)*6)*time.Hour + time.Duration((i%10)*13)*time.Minute + time.Duration((i%15)*17)*time.Second)
		samples = append(samples, d)
	}
	return samples
}

func venusEventSampleTTJD(date time.Time) float64 {
	return TD2UT(Date2JDE(date.UTC()), true)
}

func venusEventCases() []venusEventCase {
	const (
		conjunctionTolerance = 0.00001
		searchTolerance      = 30.0 / 86400.0
	)
	return []venusEventCase{
		{name: "LastVenusConjunction", tolerance: conjunctionTolerance, fn: LastVenusConjunction},
		{name: "NextVenusConjunction", tolerance: conjunctionTolerance, fn: NextVenusConjunction},
		{name: "LastVenusInferiorConjunction", tolerance: conjunctionTolerance, fn: LastVenusInferiorConjunction},
		{name: "NextVenusInferiorConjunction", tolerance: conjunctionTolerance, fn: NextVenusInferiorConjunction},
		{name: "LastVenusSuperiorConjunction", tolerance: conjunctionTolerance, fn: LastVenusSuperiorConjunction},
		{name: "NextVenusSuperiorConjunction", tolerance: conjunctionTolerance, fn: NextVenusSuperiorConjunction},
		{name: "LastVenusRetrograde", tolerance: searchTolerance, fn: LastVenusRetrograde},
		{name: "NextVenusRetrograde", tolerance: searchTolerance, fn: NextVenusRetrograde},
		{name: "LastVenusProgradeToRetrograde", tolerance: searchTolerance, fn: LastVenusProgradeToRetrograde},
		{name: "NextVenusProgradeToRetrograde", tolerance: searchTolerance, fn: NextVenusProgradeToRetrograde},
		{name: "LastVenusRetrogradeToPrograde", tolerance: searchTolerance, fn: LastVenusRetrogradeToPrograde},
		{name: "NextVenusRetrogradeToPrograde", tolerance: searchTolerance, fn: NextVenusRetrogradeToPrograde},
		{name: "LastVenusGreatestElongation", tolerance: searchTolerance, fn: LastVenusGreatestElongation},
		{name: "NextVenusGreatestElongation", tolerance: searchTolerance, fn: NextVenusGreatestElongation},
		{name: "LastVenusGreatestElongationEast", tolerance: searchTolerance, fn: LastVenusGreatestElongationEast},
		{name: "NextVenusGreatestElongationEast", tolerance: searchTolerance, fn: NextVenusGreatestElongationEast},
		{name: "LastVenusGreatestElongationWest", tolerance: searchTolerance, fn: LastVenusGreatestElongationWest},
		{name: "NextVenusGreatestElongationWest", tolerance: searchTolerance, fn: NextVenusGreatestElongationWest},
	}
}
