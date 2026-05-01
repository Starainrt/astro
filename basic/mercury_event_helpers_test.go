package basic

import (
	"math"
	"time"
)

type mercuryEventFunc func(float64) float64

type mercuryEventCase struct {
	name      string
	tolerance float64
	fn        mercuryEventFunc
}

func mercuryEventSamples() []time.Time {
	start := time.Date(1995, 1, 15, 12, 34, 56, 789000000, time.UTC)
	samples := make([]time.Time, 0, 96)
	for i := 0; i < 48; i++ {
		d := start.AddDate(0, 0, i*137)
		d = d.Add(time.Duration((i%7)*3)*time.Hour + time.Duration((i%11)*7)*time.Minute + time.Duration((i%13)*11)*time.Second)
		samples = append(samples, d)
	}
	extraStart := start.AddDate(0, 0, 41)
	for i := 0; i < 48; i++ {
		d := extraStart.AddDate(0, 0, i*89)
		d = d.Add(time.Duration((i%5)*7)*time.Hour + time.Duration((i%13)*5)*time.Minute + time.Duration((i%17)*19)*time.Second)
		samples = append(samples, d)
	}
	return samples
}

func mercuryEventSampleTTJD(date time.Time) float64 {
	return TD2UT(Date2JDE(date.UTC()), true)
}

func mercuryEventCases() []mercuryEventCase {
	const (
		conjunctionTolerance = 0.00001
		searchTolerance      = 30.0 / 86400.0
	)
	return []mercuryEventCase{
		{name: "LastMercuryConjunction", tolerance: conjunctionTolerance, fn: LastMercuryConjunction},
		{name: "NextMercuryConjunction", tolerance: conjunctionTolerance, fn: NextMercuryConjunction},
		{name: "LastMercuryInferiorConjunction", tolerance: conjunctionTolerance, fn: LastMercuryInferiorConjunction},
		{name: "NextMercuryInferiorConjunction", tolerance: conjunctionTolerance, fn: NextMercuryInferiorConjunction},
		{name: "LastMercurySuperiorConjunction", tolerance: conjunctionTolerance, fn: LastMercurySuperiorConjunction},
		{name: "NextMercurySuperiorConjunction", tolerance: conjunctionTolerance, fn: NextMercurySuperiorConjunction},
		{name: "LastMercuryRetrograde", tolerance: searchTolerance, fn: LastMercuryRetrograde},
		{name: "NextMercuryRetrograde", tolerance: searchTolerance, fn: NextMercuryRetrograde},
		{name: "LastMercuryProgradeToRetrograde", tolerance: searchTolerance, fn: LastMercuryProgradeToRetrograde},
		{name: "NextMercuryProgradeToRetrograde", tolerance: searchTolerance, fn: NextMercuryProgradeToRetrograde},
		{name: "LastMercuryRetrogradeToPrograde", tolerance: searchTolerance, fn: LastMercuryRetrogradeToPrograde},
		{name: "NextMercuryRetrogradeToPrograde", tolerance: searchTolerance, fn: NextMercuryRetrogradeToPrograde},
		{name: "LastMercuryGreatestElongation", tolerance: searchTolerance, fn: LastMercuryGreatestElongation},
		{name: "NextMercuryGreatestElongation", tolerance: searchTolerance, fn: NextMercuryGreatestElongation},
		{name: "LastMercuryGreatestElongationEast", tolerance: searchTolerance, fn: LastMercuryGreatestElongationEast},
		{name: "NextMercuryGreatestElongationEast", tolerance: searchTolerance, fn: NextMercuryGreatestElongationEast},
		{name: "LastMercuryGreatestElongationWest", tolerance: searchTolerance, fn: LastMercuryGreatestElongationWest},
		{name: "NextMercuryGreatestElongationWest", tolerance: searchTolerance, fn: NextMercuryGreatestElongationWest},
	}
}

func almostEqualWithinDays(got, want, tolerance float64) bool {
	return math.Abs(got-want) <= tolerance
}
