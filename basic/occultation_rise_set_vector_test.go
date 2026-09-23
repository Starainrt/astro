package basic

import (
	"math"
	"testing"
	"time"
)

func TestOccultationRiseSetVectorStateMatchesLegacyTopocentricState(t *testing.T) {
	planetConfig, ok := planetOccultationConfigFor(OccultationSaturn)
	if !ok {
		t.Fatal("Saturn occultation config is unavailable")
	}
	planetTT := occultationTimeToTT(time.Date(2024, time.August, 21, 2, 41, 36, 0, time.UTC))
	planetState := planetOccultationEphemerisStateAt(planetTT, planetConfig)

	star := StarCoordinate{
		ID: "parallax-test", RA: 247.3516666666667, Dec: -26.431944444444444,
		Epoch: time.Date(2000, 1, 1, 12, 0, 0, 0, time.UTC), Frame: CoordinateFrameJ2000,
		ProperMotionRACosDecMasPerYear: -10, ProperMotionDecMasPerYear: -20, ParallaxMas: 24,
	}
	starTT := occultationTimeToTT(time.Date(2026, time.February, 11, 12, 0, 0, 0, time.UTC))
	starState := starOccultationEphemerisStateAt(starTT, star)

	contexts := []struct {
		name                                  string
		tt                                    float64
		moonRA, moonDec, moonDistanceKM       float64
		targetRA, targetDec, targetDistanceKM float64
		targetRadiusKM                        float64
	}{
		{
			name: "planet", tt: planetTT,
			moonRA: planetState.moonRA, moonDec: planetState.moonDec, moonDistanceKM: planetState.moonDistanceKM,
			targetRA: planetState.planetRA, targetDec: planetState.planetDec, targetDistanceKM: planetState.planetDistanceKM,
			targetRadiusKM: planetConfig.equatorialRadiusKM,
		},
		{
			name: "finite-distance-star", tt: starTT,
			moonRA: starState.moonRA, moonDec: starState.moonDec, moonDistanceKM: starState.moonDistanceKM,
			targetRA: starState.starRA, targetDec: starState.starDec, targetDistanceKM: starState.starDistanceKM,
		},
		{
			name: "infinite-distance-star", tt: starTT,
			moonRA: starState.moonRA, moonDec: starState.moonDec, moonDistanceKM: starState.moonDistanceKM,
			targetRA: starState.starRA, targetDec: starState.starDec,
		},
	}
	locations := []Observer{
		{Longitude: 0, Latitude: 0},
		{Longitude: 115.4, Latitude: 32.9},
		{Longitude: -73.9857, Latitude: 40.7484},
		{Longitude: 179.9, Latitude: 80},
		{Longitude: -120, Latitude: -70},
	}

	for _, test := range contexts {
		t.Run(test.name, func(t *testing.T) {
			context := newOccultationRiseSetContext(
				test.tt, test.moonRA, test.moonDec, test.moonDistanceKM,
				test.targetRA, test.targetDec, test.targetDistanceKM, test.targetRadiusKM,
			)
			for _, observer := range locations {
				got := context.stateAt(observer.Longitude, observer.Latitude)
				want := legacyOccultationRiseSetStateAt(
					test.tt, test.moonRA, test.moonDec, test.moonDistanceKM,
					test.targetRA, test.targetDec, test.targetDistanceKM, test.targetRadiusKM,
					observer.Longitude, observer.Latitude,
				)
				if got.valid != want.valid {
					t.Fatalf("observer %.4f %.4f valid=%t, want %t", observer.Longitude, observer.Latitude, got.valid, want.valid)
				}
				if !got.valid {
					continue
				}
				assertOccultationRiseSetStateClose(t, "contact metric", got.contactMetric, want.contactMetric, 2e-10)
				assertOccultationRiseSetStateClose(t, "separation squared", got.separationSquared, want.separationSquared, 2e-13)
				assertOccultationRiseSetStateClose(t, "moon altitude", got.moonAltitude, want.moonAltitude, 2e-10)
			}
		})
	}
}

func TestOccultationRiseSetContextFromVectorsMatchesRADecContext(t *testing.T) {
	config, ok := planetOccultationConfigFor(OccultationSaturn)
	if !ok {
		t.Fatal("Saturn occultation config is unavailable")
	}
	tt := occultationTimeToTT(time.Date(2025, time.January, 5, 17, 0, 0, 0, time.UTC))
	state := planetOccultationEphemerisStateAt(tt, config)
	moon := occultationPathRaDecVector(state.moonRA, state.moonDec, state.moonDistanceKM)
	target := occultationPathRaDecVector(state.planetRA, state.planetDec, state.planetDistanceKM)
	vectorContext := newOccultationRiseSetContextFromVectors(
		tt,
		[3]float64{moon.x, moon.y, moon.z},
		[3]float64{target.x, target.y, target.z},
		true,
		config.equatorialRadiusKM,
	)
	raDecContext := newOccultationRiseSetContext(
		tt, state.moonRA, state.moonDec, state.moonDistanceKM,
		state.planetRA, state.planetDec, state.planetDistanceKM, config.equatorialRadiusKM,
	)
	for _, observer := range []Observer{
		{Longitude: 0, Latitude: 0},
		{Longitude: 115.4, Latitude: 32.9},
		{Longitude: -45, Latitude: 82},
		{Longitude: 170, Latitude: -70},
	} {
		got := vectorContext.stateAt(observer.Longitude, observer.Latitude)
		want := raDecContext.stateAt(observer.Longitude, observer.Latitude)
		if got.valid != want.valid {
			t.Fatalf("observer %.3f %.3f valid=%t, want %t", observer.Longitude, observer.Latitude, got.valid, want.valid)
		}
		if !got.valid {
			continue
		}
		assertOccultationRiseSetStateClose(t, "contact metric", got.contactMetric, want.contactMetric, 1e-12)
		assertOccultationRiseSetStateClose(t, "separation squared", got.separationSquared, want.separationSquared, 1e-15)
		assertOccultationRiseSetStateClose(t, "moon altitude", got.moonAltitude, want.moonAltitude, 1e-12)
	}
}

func TestOccultationRiseSetMoonHorizonResidualMatchesVectorAltitude(t *testing.T) {
	config, ok := planetOccultationConfigFor(OccultationSaturn)
	if !ok {
		t.Fatal("Saturn occultation config is unavailable")
	}
	tt := occultationTimeToTT(time.Date(2025, time.January, 5, 17, 0, 0, 0, time.UTC))
	state := planetOccultationEphemerisStateAt(tt, config)
	context := newOccultationRiseSetContext(
		tt, state.moonRA, state.moonDec, state.moonDistanceKM,
		state.planetRA, state.planetDec, state.planetDistanceKM, config.equatorialRadiusKM,
	)
	for _, observer := range []Observer{
		{Longitude: 0, Latitude: 0},
		{Longitude: 115.4, Latitude: 32.9},
		{Longitude: -45, Latitude: 82},
		{Longitude: 170, Latitude: -70},
	} {
		got, valid := context.moonHorizonResidual(observer.Longitude, observer.Latitude)
		observerParallax, _, zenith := occultationRiseSetObserverVectors(
			context.siderealDegrees, observer.Longitude, observer.Latitude,
		)
		topocentricMoon := occultationPathSub(context.moon.positionKM, observerParallax)
		want := occultationPathDot(topocentricMoon, zenith)
		if !valid || math.Abs(got-want) > 1e-8 {
			t.Fatalf("observer %.3f %.3f horizon residual=%.15g valid=%t, want %.15g",
				observer.Longitude, observer.Latitude, got, valid, want)
		}
		altitude := context.stateAt(observer.Longitude, observer.Latitude).moonAltitude * rad
		if difference := math.Abs(got/occultationPathNorm(topocentricMoon) - math.Sin(altitude)); difference > 1e-14 {
			t.Fatalf("observer %.3f %.3f normalized horizon residual differs by %.3g",
				observer.Longitude, observer.Latitude, difference)
		}
	}
}

func TestOccultationRiseSetEvaluationCacheSeparatesCandidateAndExactContexts(t *testing.T) {
	var exactCalls, candidateCalls int
	contextAt := func(tt float64) occultationRiseSetContext {
		return newOccultationRiseSetContext(tt, 10, 5, 384000, 10.5, 5.25, 1e9, 0)
	}
	cache := newOccultationRiseSetEvaluationCacheWithCandidate(
		func(tt float64) occultationRiseSetContext {
			exactCalls++
			return contextAt(tt)
		},
		func(tt float64) occultationRiseSetContext {
			candidateCalls++
			return contextAt(tt)
		},
	)
	tt := occultationTimeToTT(time.Date(2025, time.January, 5, 17, 0, 0, 0, time.UTC))
	cache.candidateEvaluation(tt)
	cache.candidateEvaluation(tt)
	if exactCalls != 0 || candidateCalls != 3 {
		t.Fatalf("candidate evaluation calls exact=%d candidate=%d, want 0/3", exactCalls, candidateCalls)
	}
	cache.evaluation(tt)
	cache.evaluation(tt)
	if exactCalls != 3 || candidateCalls != 3 {
		t.Fatalf("exact evaluation calls exact=%d candidate=%d, want 3/3", exactCalls, candidateCalls)
	}
}

func legacyOccultationRiseSetStateAt(
	tt, moonRA, moonDec, moonDistanceKM,
	targetRA, targetDec, targetDistanceKM, targetRadiusKM,
	longitude, latitude float64,
) occultationRiseSetState {
	siderealDegrees := ApparentSiderealTime(TT2UT1(tt)) * 15
	observer := Observer{Longitude: longitude, Latitude: latitude}
	moonTopocentricRA, moonTopocentricDec := topocentricRaDecWithSidereal(
		moonRA, moonDec, latitude, longitude, siderealDegrees,
		moonDistanceKM/occultationPathAstronomicalUnitKM, 0,
	)
	moonTopocentricRA = normalizeRA(moonTopocentricRA)
	targetTopocentricRA, targetTopocentricDec := targetRA, targetDec
	if targetDistanceKM > 0 {
		targetTopocentricRA, targetTopocentricDec = topocentricRaDecWithSidereal(
			targetRA, targetDec, latitude, longitude, siderealDegrees,
			targetDistanceKM/occultationPathAstronomicalUnitKM, 0,
		)
		targetTopocentricRA = normalizeRA(targetTopocentricRA)
	}
	moonTopocentricDistanceKM := topocentricDistanceKMWithSidereal(
		moonRA, moonDec, moonDistanceKM, observer, siderealDegrees,
	)
	if !finite(moonTopocentricDistanceKM) || moonTopocentricDistanceKM <= moonEquatorialRadiusKM {
		return occultationRiseSetState{}
	}
	moonRadius := angularSemidiameterArcsec(moonEquatorialRadiusKM, moonTopocentricDistanceKM) / 3600
	targetRadius := 0.0
	if targetRadiusKM > 0 {
		targetTopocentricDistanceKM := topocentricDistanceKMWithSidereal(
			targetRA, targetDec, targetDistanceKM, observer, siderealDegrees,
		)
		if !finite(targetTopocentricDistanceKM) || targetTopocentricDistanceKM <= targetRadiusKM {
			return occultationRiseSetState{}
		}
		targetRadius = angularSemidiameterArcsec(targetRadiusKM, targetTopocentricDistanceKM) / 3600
	}
	separation := angularSeparationDegrees(
		moonTopocentricRA, moonTopocentricDec, targetTopocentricRA, targetTopocentricDec,
	)
	separationRad := separation * rad
	moonAltitude := occultationAltitudeWithSidereal(
		siderealDegrees, observer, moonTopocentricRA, moonTopocentricDec,
	)
	return occultationRiseSetState{
		contactMetric:     separation - moonRadius - targetRadius,
		separationSquared: 2 - 2*math.Cos(separationRad),
		moonAltitude:      moonAltitude,
		valid:             finite(separation) && finite(moonRadius) && finite(targetRadius) && finite(moonAltitude),
	}
}

func assertOccultationRiseSetStateClose(t *testing.T, name string, got, want, tolerance float64) {
	t.Helper()
	if difference := math.Abs(got - want); difference > tolerance {
		t.Fatalf("%s=%.15g, want %.15g (difference %.3g, tolerance %.3g)", name, got, want, difference, tolerance)
	}
}
