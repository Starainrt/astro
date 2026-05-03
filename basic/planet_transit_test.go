package basic

import (
	"math"
	"testing"
	"time"
)

func TestKnownMercuryTransits(t *testing.T) {
	tests := []struct {
		name     string
		query    time.Time
		greatest time.Time
	}{
		{
			name:     "2016 May",
			query:    time.Date(2016, 1, 1, 0, 0, 0, 0, time.UTC),
			greatest: time.Date(2016, 5, 9, 14, 57, 0, 0, time.UTC),
		},
		{
			name:     "2019 Nov",
			query:    time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC),
			greatest: time.Date(2019, 11, 11, 15, 20, 0, 0, time.UTC),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := NextMercuryTransit(Date2JDE(tc.query))
			if !result.Valid {
				t.Fatal("expected valid transit")
			}
			got := JDE2DateByZone(result.Greatest, time.UTC, false)
			t.Logf("start=%s greatest=%s end=%s min=%.3f sun=%.3f planet=%.3f",
				JDE2DateByZone(result.ExternalIngress, time.UTC, false),
				got,
				JDE2DateByZone(result.ExternalEgress, time.UTC, false),
				result.MinimumSeparationArcsec,
				result.SunSemidiameterArcsec,
				result.PlanetSemidiameterArcsec,
			)
			if math.Abs(got.Sub(tc.greatest).Seconds()) > 20*60 {
				t.Fatalf("greatest mismatch: got %s want near %s", got, tc.greatest)
			}
			if !result.HasInternal {
				t.Fatalf("expected internal contacts")
			}
			if !(result.ExternalIngress < result.InternalIngress &&
				result.InternalIngress < result.Greatest &&
				result.Greatest < result.InternalEgress &&
				result.InternalEgress < result.ExternalEgress) {
				t.Fatalf("contacts out of order: %+v", result)
			}
		})
	}
}

func TestKnownVenusTransits(t *testing.T) {
	tests := []struct {
		name     string
		query    time.Time
		greatest time.Time
	}{
		{
			name:     "2004 Jun",
			query:    time.Date(2004, 1, 1, 0, 0, 0, 0, time.UTC),
			greatest: time.Date(2004, 6, 8, 8, 20, 0, 0, time.UTC),
		},
		{
			name:     "2012 Jun",
			query:    time.Date(2012, 1, 1, 0, 0, 0, 0, time.UTC),
			greatest: time.Date(2012, 6, 6, 1, 29, 0, 0, time.UTC),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := NextVenusTransit(Date2JDE(tc.query))
			if !result.Valid {
				t.Fatal("expected valid transit")
			}
			got := JDE2DateByZone(result.Greatest, time.UTC, false)
			t.Logf("start=%s greatest=%s end=%s min=%.3f sun=%.3f planet=%.3f",
				JDE2DateByZone(result.ExternalIngress, time.UTC, false),
				got,
				JDE2DateByZone(result.ExternalEgress, time.UTC, false),
				result.MinimumSeparationArcsec,
				result.SunSemidiameterArcsec,
				result.PlanetSemidiameterArcsec,
			)
			if math.Abs(got.Sub(tc.greatest).Seconds()) > 20*60 {
				t.Fatalf("greatest mismatch: got %s want near %s", got, tc.greatest)
			}
			if !result.HasInternal {
				t.Fatalf("expected internal contacts")
			}
			if !(result.ExternalIngress < result.InternalIngress &&
				result.InternalIngress < result.Greatest &&
				result.Greatest < result.InternalEgress &&
				result.InternalEgress < result.ExternalEgress) {
				t.Fatalf("contacts out of order: %+v", result)
			}
		})
	}
}

func TestTransitSearchSkipsSparseEvents(t *testing.T) {
	mercuryResult := NextMercuryTransit(Date2JDE(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)))
	if !mercuryResult.Valid {
		t.Fatal("expected Mercury transit")
	}
	mercuryGreatest := JDE2DateByZone(mercuryResult.Greatest, time.UTC, false)
	if mercuryGreatest.Year() != 2032 || mercuryGreatest.Month() != time.November {
		t.Fatalf("unexpected next Mercury transit: %s", mercuryGreatest)
	}

	venusResult := NextVenusTransit(Date2JDE(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)))
	if !venusResult.Valid {
		t.Fatal("expected Venus transit")
	}
	venusGreatest := JDE2DateByZone(venusResult.Greatest, time.UTC, false)
	if venusGreatest.Year() != 2117 || venusGreatest.Month() != time.December {
		t.Fatalf("unexpected next Venus transit: %s", venusGreatest)
	}
}

func BenchmarkNextMercuryTransitFrom2026(b *testing.B) {
	jd := Date2JDE(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))
	for i := 0; i < b.N; i++ {
		result := NextMercuryTransit(jd)
		if !result.Valid {
			b.Fatal("expected valid transit")
		}
	}
}

func BenchmarkNextVenusTransitFrom2026(b *testing.B) {
	jd := Date2JDE(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))
	for i := 0; i < b.N; i++ {
		result := NextVenusTransit(jd)
		if !result.Valid {
			b.Fatal("expected valid transit")
		}
	}
}
