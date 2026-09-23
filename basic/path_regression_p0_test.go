package basic

import (
	"math"
	"testing"
	"time"
)

func TestSolarEclipseSarosFamilyRemainsFiniteAcrossFiveCenturies(t *testing.T) {
	base := JDCalc(2024, 4, 8)
	const sarosDays = 6585.321314
	for familyIndex := -28; familyIndex <= 28; familyIndex++ {
		seed := base + float64(familyIndex)*sarosDays
		result := SolarEclipse(seed)
		for name, value := range map[string]float64{
			"greatest":  result.GreatestEclipse,
			"gamma":     result.Gamma,
			"magnitude": result.Magnitude,
			"longitude": result.GreatestLongitude,
			"latitude":  result.GreatestLatitude,
		} {
			if !finite(value) {
				t.Fatalf("saros family index %d %s=%v", familyIndex, name, value)
			}
		}
		if result.HasPartial && !(result.PartialBeginOnEarth <= result.GreatestEclipse &&
			result.GreatestEclipse <= result.PartialEndOnEarth) {
			t.Fatalf("saros family index %d partial window does not contain greatest: %+v", familyIndex, result)
		}
		if result.HasCentral && !(result.CentralBeginOnEarth <= result.GreatestEclipse &&
			result.GreatestEclipse <= result.CentralEndOnEarth) {
			t.Fatalf("saros family index %d central window does not contain greatest: %+v", familyIndex, result)
		}
	}
}

func TestSolarEclipseRepresentativePathSeriesAreOrderedAndFinite(t *testing.T) {
	cases := []struct {
		name string
		seed float64
	}{
		{name: "2009-07-22", seed: JDCalc(2009, 7, 22)},
		{name: "2010-01-15", seed: JDCalc(2010, 1, 15)},
		{name: "2014-04-29-non-central", seed: JDCalc(2014, 4, 29)},
		{name: "2023-04-20", seed: JDCalc(2023, 4, 20)},
		{name: "2043-10-03", seed: JDCalc(2043, 10, 3)},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			result := SolarEclipsePartialFootprints(test.seed, SolarEclipsePartialFootprintOptions{
				StepDays: 20.0 / 1440.0, BoundaryPoints: 24,
				CentralShadowStepDays: 20.0 / 1440.0,
				DisableRiseSetCurves:  true,
			})
			if !result.Eclipse.HasPartial {
				t.Fatalf("expected partial eclipse, got %+v", result.Eclipse)
			}
			if test.name == "2014-04-29-non-central" &&
				(result.Eclipse.Centrality != SolarEclipseNonCentral || len(result.CentralBandSegments) == 0) {
				t.Fatalf("non-central eclipse lost centrality envelope: centrality=%s segments=%d",
					result.Eclipse.Centrality, len(result.CentralBandSegments))
			}
			assertSolarEclipseFootprintSeriesFinite(t, result.Footprints)
			assertSolarEclipseFootprintSeriesFinite(t, result.CentralShadowFootprints)
			assertSolarEclipseFootprintSeriesFinite(t, result.CentralBandFootprints)
			for index := 1; index < len(result.CentralBandSegments); index++ {
				if len(result.CentralBandSegments[index]) == 0 {
					t.Fatalf("central band segment %d is empty", index)
				}
			}
		})
	}
}

func assertSolarEclipseFootprintSeriesFinite(t *testing.T, footprints []SolarEclipsePartialFootprint) {
	t.Helper()
	for index, footprint := range footprints {
		if !finite(footprint.JDE) {
			t.Fatalf("footprint %d has invalid JDE=%v", index, footprint.JDE)
		}
		for boundaryIndex, boundary := range footprint.Boundaries {
			if len(boundary) < 2 {
				t.Fatalf("footprint %d boundary %d has %d points", index, boundaryIndex, len(boundary))
			}
			for pointIndex, point := range boundary {
				if !finite(point.JDE) || !finite(point.Longitude) || !finite(point.Latitude) || !finite(point.SunAltitude) {
					t.Fatalf("footprint %d boundary %d point %d is invalid: %+v", index, boundaryIndex, pointIndex, point)
				}
			}
		}
		if footprint.Closed {
			totalPoints := 0
			for _, boundary := range footprint.Boundaries {
				totalPoints += len(boundary)
			}
			if totalPoints < 3 {
				t.Fatalf("closed footprint %d has only %d points", index, totalPoints)
			}
			if len(footprint.Boundaries) == 1 {
				boundary := footprint.Boundaries[0]
				if solarEclipsePathDistanceKM(boundary[0], boundary[len(boundary)-1]) > 5 {
					t.Fatalf("closed footprint %d has %.3f km endpoint gap", index,
						solarEclipsePathDistanceKM(boundary[0], boundary[len(boundary)-1]))
				}
			}
		}
	}
}

func TestOccultationFiniteDiskStatesRemainValidAcrossFiveCenturies(t *testing.T) {
	base := occultationTimeToTT(time.Date(2025, time.January, 5, 0, 0, 0, 0, time.UTC))
	for _, yearOffset := range []float64{-500, -250, 0, 250, 500} {
		tt := base + yearOffset*365.2425
		for _, planet := range []OccultationPlanet{
			OccultationMercury, OccultationVenus, OccultationMars,
			OccultationJupiter, OccultationSaturn, OccultationUranus, OccultationNeptune,
		} {
			config, ok := planetOccultationConfigFor(planet)
			if !ok {
				t.Fatalf("%s configuration unavailable", planet)
			}
			state := planetOccultationStateAt(tt, config, nil, -1)
			if !state.valid {
				t.Fatalf("year offset %.0f %s state is invalid: %+v", yearOffset, planet, state)
			}
			if !finite(state.externalContactMetric) || !finite(state.internalContactMetric) {
				t.Fatalf("year offset %.0f %s contact metrics are not finite: %+v", yearOffset, planet, state)
			}
			if state.internalContactMetric < state.externalContactMetric {
				t.Fatalf("year offset %.0f %s inner gap %.9f is below outer gap %.9f",
					yearOffset, planet, state.internalContactMetric, state.externalContactMetric)
			}
		}
		star := starOccultationEphemerisStateAt(tt, hr4799OccultationCoordinateForTest())
		if !star.valid || !finite(star.moonDistanceKM) || !finite(star.starRA) || !finite(star.starDec) {
			t.Fatalf("year offset %.0f star state is invalid: %+v", yearOffset, star)
		}
	}
}

func TestOccultationRepresentativePathSeriesAreOrderedAndFinite(t *testing.T) {
	zone := time.FixedZone("UTC+8", 8*60*60)
	planetPaths, err := FindPlanetOccultationPaths(
		time.Date(2025, time.January, 5, 0, 0, 0, 0, zone),
		time.Date(2025, time.January, 6, 0, 0, 0, 0, zone),
		OccultationSaturn,
		OccultationPathOptions{Step: 20 * time.Minute, TargetSpacingKM: 900, DisableFootprints: true, DisableRiseSet: true},
	)
	if err != nil || len(planetPaths) != 1 {
		t.Fatalf("planet paths=%d err=%v, want one", len(planetPaths), err)
	}
	assertOccultationPointSeriesFinite(t, planetPaths[0].CenterLine)
	assertOccultationPointSeriesFinite(t, planetPaths[0].NorthernLimit)
	assertOccultationPointSeriesFinite(t, planetPaths[0].SouthernLimit)
	if planetPaths[0].HasTotalBand {
		assertOccultationPointSeriesFinite(t, planetPaths[0].NorthernTotalLimit)
		assertOccultationPointSeriesFinite(t, planetPaths[0].SouthernTotalLimit)
	}

	starPaths, err := FindStarOccultationPaths(
		time.Date(2025, time.June, 5, 0, 0, 0, 0, zone),
		time.Date(2025, time.June, 6, 0, 0, 0, 0, zone),
		hr4799OccultationCoordinateForTest(),
		OccultationPathOptions{Step: 20 * time.Minute, TargetSpacingKM: 900, DisableFootprints: true, DisableRiseSet: true},
	)
	if err != nil || len(starPaths) != 1 {
		t.Fatalf("star paths=%d err=%v, want one", len(starPaths), err)
	}
	assertOccultationPointSeriesFinite(t, starPaths[0].CenterLine)
	assertOccultationPointSeriesFinite(t, starPaths[0].NorthernLimit)
	assertOccultationPointSeriesFinite(t, starPaths[0].SouthernLimit)
}

func assertOccultationPointSeriesFinite(t *testing.T, points []OccultationPathPoint) {
	t.Helper()
	if len(points) == 0 {
		t.Fatal("path series is empty")
	}
	for index, point := range points {
		if !finite(point.Longitude) || !finite(point.Latitude) || !finite(point.WidthKM) {
			t.Fatalf("point %d is not finite: %+v", index, point)
		}
		if index > 0 && !point.Time.After(points[index-1].Time) {
			t.Fatalf("path times are not strictly increasing at %d: %v then %v", index, points[index-1].Time, point.Time)
		}
	}
}

func TestSolarEclipseRepresentativePathPointsDoNotContainNaN(t *testing.T) {
	path := SolarEclipseCentralPath(JDCalc(2010, 1, 15), SolarEclipsePathOptions{StepDays: 20.0 / 1440.0})
	for index, point := range append(append(append([]SolarEclipsePathPoint{}, path.CenterLine...), path.NorthernLimit...), path.SouthernLimit...) {
		if math.IsNaN(point.Longitude) || math.IsNaN(point.Latitude) || math.IsNaN(point.JDE) {
			t.Fatalf("path point %d contains NaN: %+v", index, point)
		}
	}
}
