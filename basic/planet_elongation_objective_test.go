package basic

import (
	"math"
	"testing"
)

// 大距口径契约：返回时刻必须是公开“视距角”函数（MercurySunElongation/VenusSunElongation）的极大，
// 而不是忽略光行差与视位置修正的真距角极大。

type elongationObjectiveCase struct {
	name     string
	elongate func(float64) float64
	next     func(float64) float64
	last     func(float64) float64
}

var elongationObjectiveCases = []elongationObjectiveCase{
	{"Mercury", MercurySunElongation, NextMercuryGreatestElongation, LastMercuryGreatestElongation},
	{"Venus", VenusSunElongation, NextVenusGreatestElongation, LastVenusGreatestElongation},
}

func elongationObjectiveSeeds() []float64 {
	seeds := make([]float64, 0, 24)
	for year := 2024; year <= 2027; year++ {
		for month := 1; month <= 12; month += 2 {
			seeds = append(seeds, JDCalc(year, month, 1))
		}
	}
	return seeds
}

func elongationObjectiveTruth(elongate func(float64) float64, eventUT float64) float64 {
	left, right := eventUT-30.0/1440.0, eventUT+30.0/1440.0
	for i := 0; i < 200; i++ {
		third := (right - left) / 3
		if elongate(UTC2TT(left+third)) <= elongate(UTC2TT(right-third)) {
			left += third
			continue
		}
		right -= third
	}
	return (left + right) / 2
}

func TestGreatestElongationMatchesPublicElongationExtremum(t *testing.T) {
	const toleranceMinutes = 2.0
	for _, tc := range elongationObjectiveCases {
		for _, seed := range elongationObjectiveSeeds() {
			for _, direction := range []struct {
				name string
				fn   func(float64) float64
			}{{"next", tc.next}, {"last", tc.last}} {
				eventUT := direction.fn(seed)
				if math.IsNaN(eventUT) {
					t.Fatalf("%s %s at %.1f returned NaN", tc.name, direction.name, seed)
				}
				truth := elongationObjectiveTruth(tc.elongate, eventUT)
				if deviation := math.Abs(eventUT-truth) * 1440; deviation > toleranceMinutes {
					t.Fatalf("%s %s at %.1f = %.9f deviates %.2f min from the public elongation maximum",
						tc.name, direction.name, seed, eventUT, deviation)
				}
			}
		}
	}
}

// 金星大距的侧向判定必须与水星同口径（0.1 s 事件容差），不能沿用站事件的 0.5 s。
//
// 这些入口的查询参数是力学时，返回时刻是世界时：
// 查询“事件后 0.3 秒”要把世界时事件先换成力学时再加偏移。
func TestGreatestElongationLateralToleranceIsEventTolerance(t *testing.T) {
	const offsetSeconds = 0.3
	cases := []struct {
		name string
		next func(float64) float64
		last func(float64) float64
	}{
		{"MercuryEast", NextMercuryGreatestElongationEast, LastMercuryGreatestElongationEast},
		{"MercuryWest", NextMercuryGreatestElongationWest, LastMercuryGreatestElongationWest},
		{"VenusEast", NextVenusGreatestElongationEast, LastVenusGreatestElongationEast},
		{"VenusWest", NextVenusGreatestElongationWest, LastVenusGreatestElongationWest},
	}
	seeds := []float64{JDCalc(2024, 3, 1), JDCalc(2025, 7, 1), JDCalc(2026, 11, 1)}
	tolerance := exactQueryTTToleranceUT
	for _, tc := range cases {
		for _, seed := range seeds {
			event := tc.next(seed)
			if math.IsNaN(event) {
				t.Fatalf("%s at %.1f returned NaN", tc.name, seed)
			}
			eventTT := UTC2TT(event)
			afterUT := event + offsetSeconds/86400.0
			if got := tc.next(eventTT + offsetSeconds/86400.0); got < afterUT-tolerance {
				t.Fatalf("%s: Next at event+%.1f s = %.9f returned an event %.3f s before the query",
					tc.name, offsetSeconds, got, (afterUT-got)*86400)
			}
			beforeUT := event - offsetSeconds/86400.0
			if got := tc.last(eventTT - offsetSeconds/86400.0); got > beforeUT+tolerance {
				t.Fatalf("%s: Last at event-%.1f s = %.9f returned an event %.3f s after the query",
					tc.name, offsetSeconds, got, (got-beforeUT)*86400)
			}
		}
	}
}

// 无东西侧参数的大距只算查询所在窗口那一侧；这里用「两侧都算再取极值」作差分对照，输出必须逐位一致。
func TestGreatestElongationAnySideKeepsBothSideExtremum(t *testing.T) {
	cases := []struct {
		name               string
		next               func(float64) float64
		last               func(float64) float64
		eastNext, westNext func(float64) float64
		eastLast, westLast func(float64) float64
	}{
		{"Mercury", NextMercuryGreatestElongation, LastMercuryGreatestElongation,
			NextMercuryGreatestElongationEast, NextMercuryGreatestElongationWest,
			LastMercuryGreatestElongationEast, LastMercuryGreatestElongationWest},
		{"Venus", NextVenusGreatestElongation, LastVenusGreatestElongation,
			NextVenusGreatestElongationEast, NextVenusGreatestElongationWest,
			LastVenusGreatestElongationEast, LastVenusGreatestElongationWest},
	}
	for _, tc := range cases {
		for jd := JDCalc(2024, 1, 1); jd <= JDCalc(2027, 1, 1); jd += 14 {
			east, west := tc.eastNext(jd), tc.westNext(jd)
			want := east
			if !sameEventJD(east, west) {
				want = earliestFiniteEventUT(east, west)
			}
			if got := tc.next(jd); got != want {
				t.Fatalf("%s Next(%.1f) = %.12f, want %.12f (both-side reference)", tc.name, jd, got, want)
			}
			east, west = tc.eastLast(jd), tc.westLast(jd)
			want = east
			if !sameEventJD(east, west) {
				want = latestFiniteEventUT(east, west)
			}
			if got := tc.last(jd); got != want {
				t.Fatalf("%s Last(%.1f) = %.12f, want %.12f (both-side reference)", tc.name, jd, got, want)
			}
		}
	}
}

func TestEventAggregationKeepsFiniteCandidate(t *testing.T) {
	const jd = 2460310.5
	if got := earliestFiniteEventUT(math.NaN(), jd); got != jd {
		t.Fatalf("earliestFiniteEventUT(NaN, jd) = %v, want %v", got, jd)
	}
	if got := latestFiniteEventUT(jd, math.NaN()); got != jd {
		t.Fatalf("latestFiniteEventUT(jd, NaN) = %v, want %v", got, jd)
	}
	if got := earliestFiniteEventUT(math.NaN(), math.Inf(1)); !math.IsNaN(got) {
		t.Fatalf("earliestFiniteEventUT(NaN, +Inf) = %v, want NaN", got)
	}
}
