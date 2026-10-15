// 被测契约：天狼星在开罗约 8 月初晨始见；金星下合前后依次出现昏末见与晨始见；极限星等随天光单调变化；偕日事件不得越出请求区间。
package basic

import (
	"math"
	"testing"
	"time"
)

func TestSiriusHeliacalRisingCairo(t *testing.T) {
	sirius := StarCoordinate{ID: "Sirius", RA: 101.28715533, Dec: -16.71611586, Epoch: time.Date(2000, 1, 1, 12, 0, 0, 0, time.UTC), Frame: CoordinateFrameJ2000,
		ProperMotionRACosDecMasPerYear: -546.01, ProperMotionDecMasPerYear: -1223.07}
	start := Date2JD(time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC))
	end := Date2JD(time.Date(2025, 9, 30, 0, 0, 0, 0, time.UTC))
	events := FindStarHeliacalEventsWithInstance(sirius, -1.46, start, end, 31.23, 30.04, HeliacalOptions{}, nil)
	var rising *HeliacalEvent
	for i := range events {
		if events[i].Kind == HeliacalMorningFirst {
			rising = &events[i]
			break
		}
	}
	if rising == nil {
		t.Fatalf("no morning first visibility: %+v", events)
	}
	when := JD2Date(rising.JD)
	t.Logf("Sirius morning first %v margin %.2f alt %.2f sun %.2f", when, rising.Margin, rising.TargetAltitude, rising.SunAltitude)
	if when.Month() != time.July && when.Month() != time.August {
		t.Errorf("Sirius heliacal rising %v, expected late July to mid August", when)
	}
}

func TestVenusHeliacalAroundInferiorConjunction(t *testing.T) {
	// 2025-03-23 金星下合。
	start := Date2JD(time.Date(2025, 2, 15, 0, 0, 0, 0, time.UTC))
	end := Date2JD(time.Date(2025, 4, 30, 0, 0, 0, 0, time.UTC))
	events := FindPlanetHeliacalEventsWithInstance(SkyVenus, start, end, 116.4, 39.9, HeliacalOptions{}, nil)
	var eveningLast, morningFirst float64
	for _, e := range events {
		t.Logf("kind %d %v margin %.2f", e.Kind, JD2Date(e.JD), e.Margin)
		if e.Kind == HeliacalEveningLast && eveningLast == 0 {
			eveningLast = e.JD
		}
		if e.Kind == HeliacalMorningFirst && morningFirst == 0 {
			morningFirst = e.JD
		}
	}
	conj := Date2JD(time.Date(2025, 3, 23, 0, 0, 0, 0, time.UTC))
	if eveningLast == 0 || morningFirst == 0 || !(eveningLast < conj && morningFirst > conj) {
		t.Errorf("evening last %.1f, morning first %.1f around conjunction %.1f", eveningLast, morningFirst, conj)
	}
}

func TestLimitingMagnitudeDarkensWithTwilight(t *testing.T) {
	prev := LimitingMagnitude(20, -2, 30, 0.25)
	for _, sun := range []float64{-4, -6, -9, -12, -15, -18} {
		cur := LimitingMagnitude(20, sun, 30, 0.25)
		if !(cur > prev) {
			t.Errorf("sun %.0f: limiting magnitude %.2f not fainter than %.2f", sun, cur, prev)
		}
		prev = cur
	}
	if prev < 5 || prev > 7.5 {
		t.Errorf("dark-sky limiting magnitude at 20° altitude %.2f", prev)
	}
}

func TestMoonlightBrightensSky(t *testing.T) {
	if MoonlightNanolambert(30, -5, 20, 10, 0.25) != 0 {
		t.Error("moon below horizon must contribute nothing")
	}
	full := MoonlightNanolambert(30, 40, 20, 5, 0.25)
	crescent := MoonlightNanolambert(30, 40, 20, 120, 0.25)
	if !(full > crescent && crescent > 0 && full > 1000) {
		t.Errorf("full %.1f crescent %.1f nL", full, crescent)
	}
	start := 2460676.5
	plain := FindPlanetHeliacalEventsWithInstance(SkyVenus, start, start+400, 116.4, 39.9, HeliacalOptions{}, nil)
	lit := FindPlanetHeliacalEventsWithInstance(SkyVenus, start, start+400, 116.4, 39.9, HeliacalOptions{Moonlight: true}, nil)
	if len(plain) == 0 || len(lit) == 0 {
		t.Fatalf("events %d %d", len(plain), len(lit))
	}
}

func TestHeliacalRejectsUnscannableRange(t *testing.T) {
	sirius := StarCoordinate{RA: 101.287, Dec: -16.716}
	for _, r := range [][2]float64{
		{2451545, math.Inf(1)}, {math.Inf(-1), 2451545}, {math.NaN(), 2451545},
		{2451545, math.NaN()}, {2451545, 2451545 + 1e10}, {1e300, math.MaxFloat64},
	} {
		if got := FindStarHeliacalEventsWithInstance(sirius, -1.46, r[0], r[1], 31.23, 30.04, HeliacalOptions{}, nil); got != nil {
			t.Errorf("star %v: %d events", r, len(got))
		}
		if got := FindPlanetHeliacalEventsWithInstance(SkyVenus, r[0], r[1], 116.4, 39.9, HeliacalOptions{}, nil); got != nil {
			t.Errorf("planet %v: %d events", r, len(got))
		}
	}
}

func TestSkyModelRejectsNonPositiveExtinction(t *testing.T) {
	for _, k := range []float64{0, -1, math.NaN(), math.Inf(1)} {
		if v := SkyBrightnessNanolambert(20, -10, 30, k); !math.IsNaN(v) {
			t.Errorf("sky k=%g: %g", k, v)
		}
		if v := LimitingMagnitude(20, -10, 30, k); !math.IsNaN(v) {
			t.Errorf("limit k=%g: %g", k, v)
		}
		if v := MoonlightNanolambert(30, 40, 20, 5, k); !math.IsNaN(v) {
			t.Errorf("moonlight k=%g: %g", k, v)
		}
	}
}

func TestHeliacalEventsStayWithinRequestedRange(t *testing.T) {
	lon, lat := 116.4, 39.9
	narrow := 0
	for _, body := range []SkyBody{SkyMercury, SkyVenus, SkyMars, SkyJupiter, SkySaturn} {
		wide := FindPlanetHeliacalEventsWithInstance(body, 2459000.5, 2460000.5, lon, lat, HeliacalOptions{}, nil)
		for _, event := range wide {
			// 事件时刻是采样日内的细化值；把 endJD 卡在两者之间，正是越界最容易出现的窗口。
			day := math.Floor(event.JD-0.5) + 0.5
			offset := event.JD - day
			if offset <= 0.1 {
				continue
			}
			start, end := day, day+offset-0.05
			narrow++
			for _, got := range FindPlanetHeliacalEventsWithInstance(body, start, end, lon, lat, HeliacalOptions{}, nil) {
				if got.JD < start || got.JD > end {
					t.Fatalf("body %v: event JD %.6f outside [%.6f, %.6f]", body, got.JD, start, end)
				}
			}
		}
	}
	if narrow == 0 {
		t.Fatal("no narrow window was exercised")
	}
}
