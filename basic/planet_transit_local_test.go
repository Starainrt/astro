package basic

import (
	"math"
	"testing"
	"time"
)

func TestPlanetTransitLocalContactsNearGeocentric(t *testing.T) {
	cases := []struct {
		name     string
		transit  PlanetTransitResult
		maxShift float64 // 站心与地心接触的最大偏移，单位分钟
	}{
		{"venus-2012", NextVenusTransit(JDCalc(2012, 1, 1)), 9},
		{"mercury-2016", NextMercuryTransit(JDCalc(2016, 1, 1)), 2},
	}
	observers := []Observer{
		{Longitude: 151.2, Latitude: -33.9},
		{Longitude: 116.4, Latitude: 39.9},
		{Longitude: -157.8, Latitude: 21.3},
		{Longitude: -0.1, Latitude: 51.5, Height: 100},
	}
	for _, tc := range cases {
		if !tc.transit.Valid {
			t.Fatalf("%s: geocentric transit not found", tc.name)
		}
		snapshot, ok := newPlanetTransitSnapshot(tc.transit, nil)
		if !ok {
			t.Fatalf("%s: snapshot failed", tc.name)
		}
		for _, observer := range observers {
			local := PlanetTransitLocal(tc.transit, observer)
			if !local.Valid || !local.HasInternal {
				t.Fatalf("%s %+v: local transit invalid", tc.name, observer)
			}
			pairs := []struct {
				local PlanetTransitLocalContact
				geo   float64
				kind  PlanetTransitContactKind
			}{
				{local.ExternalIngress, tc.transit.ExternalIngress, PlanetTransitC1},
				{local.InternalIngress, tc.transit.InternalIngress, PlanetTransitC2},
				{local.InternalEgress, tc.transit.InternalEgress, PlanetTransitC3},
				{local.ExternalEgress, tc.transit.ExternalEgress, PlanetTransitC4},
			}
			for _, pair := range pairs {
				shift := math.Abs(pair.local.JD-pair.geo) * 1440
				if shift > tc.maxShift {
					t.Fatalf("%s %+v C%d: shift %.2f min", tc.name, observer, pair.kind, shift)
				}
				state, _ := snapshot.stateAt(UTC2TT(pair.local.JD), observer)
				metric := state.external
				if pair.kind == PlanetTransitC2 || pair.kind == PlanetTransitC3 {
					metric = state.internal
				}
				if math.Abs(metric) > 0.01 {
					t.Fatalf("%s %+v C%d: residual %.4f arcsec", tc.name, observer, pair.kind, metric)
				}
			}
			// 直接星历求值核对插值缓存：凌甚时刻的站心角距与缓存结果一致。
			tt := UTC2TT(local.Greatest.JD)
			cfg := venusTransitConfig()
			if tc.transit.PlanetIndex == 1 {
				cfg = mercuryTransitConfig()
			}
			sidereal := ApparentSiderealTime(TT2UT1(tt)) * 15
			ra, dec := HSunApparentRaDecN(tt, -1)
			sunRA, sunDec := topocentricRaDecWithSidereal(ra, dec, observer.Latitude, observer.Longitude, sidereal, EarthAwayN(tt, -1), observer.Height)
			ra, dec = cfg.apparentRaDecN(tt, -1)
			planetRA, planetDec := topocentricRaDecWithSidereal(ra, dec, observer.Latitude, observer.Longitude, sidereal, cfg.earthDistanceN(tt, -1), observer.Height)
			direct := StarAngularSeparation(sunRA, sunDec, planetRA, planetDec) * 3600
			if diff := math.Abs(direct - local.MinimumSeparationArcsec); diff > 0.01 {
				t.Fatalf("%s %+v: cached separation off by %.4f arcsec", tc.name, observer, diff)
			}
		}
	}
}

func TestPlanetTransitContactIsochronesMatchLocalTimes(t *testing.T) {
	transit := NextVenusTransit(JDCalc(2012, 1, 1))
	for _, kind := range []PlanetTransitContactKind{PlanetTransitC1, PlanetTransitC4} {
		lines := PlanetTransitContactIsochrones(transit, kind, 2*time.Minute)
		if len(lines) == 0 {
			t.Fatalf("C%d: no isochrones", kind)
		}
		checked := 0
		for _, line := range lines {
			for _, segment := range line.Segments {
				for index := 0; index < len(segment); index += 7 {
					point := segment[index]
					local := PlanetTransitLocal(transit, Observer{Longitude: point.Longitude, Latitude: point.Latitude})
					contact := local.ExternalIngress
					if kind == PlanetTransitC4 {
						contact = local.ExternalEgress
					}
					if diff := math.Abs(contact.JD-line.JD) * 86400; diff > 3 {
						t.Fatalf("C%d %.5f at (%.2f,%.2f): local contact off by %.1f s", kind, line.JD, point.Longitude, point.Latitude, diff)
					}
					if contact.SunAltitudeDeg < planetTransitSunHorizonAltitudeDeg-0.01 {
						t.Fatalf("C%d at (%.2f,%.2f): Sun below horizon %.2f", kind, point.Longitude, point.Latitude, contact.SunAltitudeDeg)
					}
					checked++
				}
			}
		}
		if checked < 20 {
			t.Fatalf("C%d: only %d points checked", kind, checked)
		}
	}
}

func TestPlanetTransitLocalFramesLimitAndDedup(t *testing.T) {
	cases := []struct {
		name    string
		transit PlanetTransitResult
	}{
		{"venus-2012", NextVenusTransit(JDCalc(2012, 1, 1))},
		{"mercury-2016", NextMercuryTransit(JDCalc(2016, 1, 1))},
	}
	observer := Observer{Longitude: 116.4, Latitude: 39.9}
	for _, tc := range cases {
		if !tc.transit.Valid {
			t.Fatalf("%s: geocentric transit not found", tc.name)
		}
		local := PlanetTransitLocal(tc.transit, observer)
		if !local.Valid || !local.HasInternal {
			t.Fatalf("%s: local transit invalid", tc.name)
		}
		frames := PlanetTransitLocalFrames(tc.transit, local, 1e-9)
		if len(frames) == 0 || len(frames) > planetTransitLocalMaxFrames {
			t.Fatalf("%s: got %d frames, want 1..%d", tc.name, len(frames), planetTransitLocalMaxFrames)
		}
		for index := 1; index < len(frames); index++ {
			if frames[index].JD <= frames[index-1].JD {
				t.Fatalf("%s: duplicate or unsorted frame at %d: %.12f <= %.12f", tc.name, index, frames[index].JD, frames[index-1].JD)
			}
		}
	}
}

// 1999-11-15 水星掠凌：地心最小角距只比内切限小约 1.9″，南半球站心视差把它推出内切限，
// 这些地点只剩一触、四触；北半球仍有二触、三触。内切等时线只能落在仍有内切接触的地点。
func TestPlanetTransitLocalGrazingMercury1999(t *testing.T) {
	transit := NextMercuryTransit(JDCalc(1999, 1, 1))
	if !transit.Valid || !transit.HasInternal {
		t.Fatalf("1999 geocentric transit must exist with internal contacts: %+v", transit)
	}
	limit := transit.SunSemidiameterArcsec - transit.PlanetSemidiameterArcsec
	if margin := limit - transit.MinimumSeparationArcsec; margin <= 0 || margin > 5 {
		t.Fatalf("geocentric margin %.2f arcsec, want a near-grazing transit", margin)
	}
	grazing := PlanetTransitLocal(transit, Observer{Longitude: 0, Latitude: -60})
	if !grazing.Valid || grazing.HasInternal || grazing.InternalIngress.Valid || grazing.InternalEgress.Valid {
		t.Fatalf("southern site must see a grazing transit: %+v", grazing)
	}
	if grazing.MinimumSeparationArcsec <= grazing.SunSemidiameterArcsec-grazing.PlanetSemidiameterArcsec {
		t.Fatalf("grazing site separation %.2f inside the internal limit", grazing.MinimumSeparationArcsec)
	}
	if !(grazing.ExternalIngress.JD < grazing.Greatest.JD && grazing.Greatest.JD < grazing.ExternalEgress.JD) {
		t.Fatalf("grazing site contacts out of order: %+v", grazing)
	}
	internal := PlanetTransitLocal(transit, Observer{Longitude: 0, Latitude: 40})
	if !internal.Valid || !internal.HasInternal ||
		!(internal.ExternalIngress.JD < internal.InternalIngress.JD && internal.InternalIngress.JD < internal.Greatest.JD &&
			internal.Greatest.JD < internal.InternalEgress.JD && internal.InternalEgress.JD < internal.ExternalEgress.JD) {
		t.Fatalf("northern site must see ordered internal contacts: %+v", internal)
	}
	lines := PlanetTransitContactIsochrones(transit, PlanetTransitC2, 20*time.Second)
	if len(lines) == 0 {
		t.Fatal("C2 isochrones missing")
	}
	checked := 0
	for _, line := range lines {
		for _, segment := range line.Segments {
			for index := 0; index < len(segment); index += 5 {
				point := segment[index]
				local := PlanetTransitLocal(transit, Observer{Longitude: point.Longitude, Latitude: point.Latitude})
				if !local.HasInternal {
					t.Fatalf("C2 isochrone point (%.2f,%.2f) has no internal contact", point.Longitude, point.Latitude)
				}
				// 掠凌边缘内切时刻随余量按平方根变化，梯度发散，网格插值在余量 <0.5″ 处可差到 20 s。
				if local.SunSemidiameterArcsec-local.PlanetSemidiameterArcsec-local.MinimumSeparationArcsec < 0.5 {
					continue
				}
				if diff := math.Abs(local.InternalIngress.JD-line.JD) * 86400; diff > 3 {
					t.Fatalf("C2 at (%.2f,%.2f): local contact off by %.1f s", point.Longitude, point.Latitude, diff)
				}
				checked++
			}
		}
	}
	if checked < 10 {
		t.Fatalf("only %d C2 points checked", checked)
	}
}

// 步长过细时等时线不应被截断：条数不超上限，且仍覆盖整个接触时段。
func TestPlanetTransitContactIsochronesFineStepCoversSpan(t *testing.T) {
	transit := NextVenusTransit(JDCalc(2012, 1, 1))
	coarse := PlanetTransitContactIsochrones(transit, PlanetTransitC1, 2*time.Minute)
	fine := PlanetTransitContactIsochrones(transit, PlanetTransitC1, 5*time.Second)
	if len(coarse) == 0 || len(fine) == 0 {
		t.Fatal("isochrones missing")
	}
	if len(fine) > planetTransitIsochroneMaxLevels {
		t.Fatalf("%d lines exceed the cap", len(fine))
	}
	coarseSpan := (coarse[len(coarse)-1].JD - coarse[0].JD) * 1440
	fineSpan := (fine[len(fine)-1].JD - fine[0].JD) * 1440
	if fineSpan < coarseSpan-4 {
		t.Fatalf("fine step covers %.1f min, coarse %.1f min", fineSpan, coarseSpan)
	}
}

func TestPlanetTransitSnapshotRejectsBadInput(t *testing.T) {
	transit := NextVenusTransit(JDCalc(2012, 1, 1))
	cases := map[string]func(*PlanetTransitResult){
		"no external":      func(r *PlanetTransitResult) { r.HasExternal = false },
		"huge span":        func(r *PlanetTransitResult) { r.ExternalEgress = r.ExternalIngress + 1e6 },
		"greatest outside": func(r *PlanetTransitResult) { r.Greatest = r.ExternalEgress + 0.1 },
		"greatest NaN":     func(r *PlanetTransitResult) { r.Greatest = math.NaN() },
	}
	for name, mutate := range cases {
		bad := transit
		mutate(&bad)
		if _, ok := newPlanetTransitSnapshot(bad, nil); ok {
			t.Fatalf("%s: snapshot must be rejected", name)
		}
		if local := PlanetTransitLocal(bad, Observer{}); local.Valid {
			t.Fatalf("%s: local result must be invalid", name)
		}
	}
	if _, ok := newPlanetTransitSnapshot(transit, nil); !ok {
		t.Fatal("valid transit rejected")
	}
}
