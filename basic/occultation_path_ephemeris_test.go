package basic

import (
	"errors"
	"math"
	"reflect"
	"testing"
	"time"
)

func TestOccultationPathAlgorithmOptions(t *testing.T) {
	for _, algorithm := range []OccultationPathAlgorithm{"", OccultationPathAlgorithmOptimized, OccultationPathAlgorithmExact} {
		options := OccultationPathOptions{Algorithm: algorithm}
		if err := options.Validate(); err != nil {
			t.Fatal(err)
		}
		want := algorithm
		if want == "" {
			want = OccultationPathAlgorithmOptimized
		}
		if got := normalizeOccultationPathOptions(options).Algorithm; got != want {
			t.Fatalf("algorithm=%q, want %q", got, want)
		}
	}
	if err := (OccultationPathOptions{Algorithm: "unknown"}).Validate(); !errors.Is(err, ErrInvalidOccultationInput) {
		t.Fatalf("invalid algorithm: %v", err)
	}
}

func TestOccultationPathEphemerisBranches(t *testing.T) {
	center := occultationTimeToTT(time.Date(2025, 1, 5, 12, 0, 0, 0, time.UTC))
	config, _ := planetOccultationConfigFor(OccultationSaturn)
	star := hr4799OccultationCoordinateForTest()
	for _, algorithm := range []OccultationPathAlgorithm{OccultationPathAlgorithmExact, OccultationPathAlgorithmOptimized} {
		t.Run(string(algorithm), func(t *testing.T) {
			planetCache := newPlanetOccultationEventCache(config)
			starCache := newStarOccultationEventCache(star)
			planetCache.preparePathEphemeris(center, algorithm)
			starCache.preparePathEphemeris(center, algorithm)
			dense := algorithm == OccultationPathAlgorithmOptimized
			if planetCache.local.dense != dense || starCache.local.dense != dense {
				t.Fatalf("unexpected selected tables: planet=%v star=%v", planetCache.local.dense, starCache.local.dense)
			}
			if !dense {
				if !reflect.DeepEqual(planetCache.local, newPlanetOccultationLocalEphemeris(center, config)) ||
					!reflect.DeepEqual(starCache.local, newStarOccultationLocalEphemeris(center, star)) {
					t.Fatal("exact branch changed the legacy predictor")
				}
			}
			for _, tt := range []float64{center + .03123, center + 3} {
				planetState := planetCache.stateAt(tt)
				starState := starCache.stateAt(tt)
				if dense && tt < center+2 {
					wantPlanet, _ := planetCache.local.stateAt(tt)
					wantStar, _ := starCache.local.stateAt(tt)
					if planetState != wantPlanet || starState != wantStar {
						t.Fatal("optimized cache did not use its dense table")
					}
				} else if planetState != planetOccultationEphemerisStateAt(tt, config) || starState != starOccultationEphemerisStateAt(tt, star) {
					t.Fatal("exact branch or out-of-window fallback changed")
				}
			}
		})
	}
	// The independent instant API constructs these caches without a path table.
	if newPlanetOccultationEventCache(config).local != nil || newStarOccultationEventCache(star).local != nil {
		t.Fatal("instant caches must remain exact")
	}
}

func TestDenseOccultationTableRejectsInvalidAndUnresolvedData(t *testing.T) {
	center := 2451545.0
	polynomial := func(tt float64) ([3]float64, [3]float64) {
		u := tt - center
		return [3]float64{1 + u*u, 2 + u, 3}, [3]float64{7, 8 + u*u*u, 9}
	}
	if nodes := newDenseOccultationEphemerisNodes(center, polynomial); len(nodes) != occultationDenseEphemerisNodeCount {
		t.Fatal("smooth table rejected")
	}
	for _, sample := range []func(float64) ([3]float64, [3]float64){
		func(float64) ([3]float64, [3]float64) { return [3]float64{}, [3]float64{} },
		func(float64) ([3]float64, [3]float64) { return [3]float64{math.NaN()}, [3]float64{1} },
		func(tt float64) ([3]float64, [3]float64) {
			return [3]float64{2 + math.Sin((tt-center)*800), 1, 1}, [3]float64{1, 1, 1}
		},
	} {
		if nodes := newDenseOccultationEphemerisNodes(center, sample); nodes != nil {
			t.Fatal("invalid or under-resolved table accepted")
		}
	}
	if nodes := newDenseOccultationEphemerisNodes(math.NaN(), polynomial); nodes != nil {
		t.Fatal("invalid center accepted")
	}
}

func TestDenseOccultationPlanetVectorsAgainstFullEphemerides(t *testing.T) {
	center := occultationTimeToTT(time.Date(2025, 1, 5, 12, 0, 0, 0, time.UTC))
	for _, planet := range []OccultationPlanet{OccultationMercury, OccultationVenus, OccultationMars, OccultationJupiter, OccultationSaturn, OccultationUranus, OccultationNeptune} {
		t.Run(planet.String(), func(t *testing.T) {
			config, _ := planetOccultationConfigFor(planet)
			cache := newPlanetOccultationEventCache(config)
			cache.preparePathEphemeris(center, OccultationPathAlgorithmOptimized)
			if !cache.local.dense {
				t.Fatal("reference event did not accept a dense table")
			}
			// Probe off-node, off-midpoint times including both window ends.
			for _, offset := range []float64{-1.999, -1.9731, -.314159, .03712, 1.9713, 1.999} {
				tt := center + offset
				moon, target, ok := cache.local.vectorsAt(tt)
				exact := planetOccultationEphemerisStateAt(tt, config)
				em := occultationPathRaDecVector(exact.moonRA, exact.moonDec, exact.moonDistanceKM)
				et := occultationPathRaDecVector(exact.planetRA, exact.planetDec, exact.planetDistanceKM)
				if !ok || !denseOccultationVectorAccurate(moon, [3]float64{em.x, em.y, em.z}) ||
					!denseOccultationVectorAccurate(target, [3]float64{et.x, et.y, et.z}) {
					t.Fatalf("interpolation exceeds the vector tolerance at offset %g", offset)
				}
			}
		})
	}
}

func TestDenseOccultationInterpolationUsesActualNodeTimes(t *testing.T) {
	nodes := make([]localEphemerisVectorNode, 9)
	for i := range nodes {
		tt := 2451545 + float64(i)*occultationDenseEphemerisStepDays
		u := tt - 2451545
		nodes[i] = localEphemerisVectorNode{tt: tt, first: [3]float64{1, u, u * u}, next: [3]float64{3, u * u * u, 7}}
	}
	for _, tt := range []float64{nodes[0].tt, nodes[8].tt, nodes[4].tt + .00321} {
		first, next, ok := interpolateDenseOccultationVectors(nodes, tt)
		u := tt - 2451545
		if !ok || math.Abs(first[1]-u) > 1e-14 || math.Abs(first[2]-u*u) > 1e-14 || math.Abs(next[1]-u*u*u) > 1e-14 {
			t.Fatalf("interpolation at %.10f: %v %v %v", tt, first, next, ok)
		}
	}
	for _, tt := range []float64{math.NaN(), nodes[0].tt - .1, nodes[8].tt + .1} {
		if _, _, ok := interpolateDenseOccultationVectors(nodes, tt); ok {
			t.Fatal("out-of-window interpolation accepted")
		}
	}
}

func TestOccultationPathDefaultMatchesOptimized(t *testing.T) {
	start := time.Date(2025, 7, 29, 0, 0, 0, 0, time.FixedZone("UTC+8", 8*3600))
	options := OccultationPathOptions{Step: 20 * time.Minute, TargetSpacingKM: 900, RiseSetStep: time.Minute, DisableFootprints: true}
	want, err := FindPlanetOccultationPaths(start, start.Add(24*time.Hour), OccultationMars, options)
	if err != nil || len(want) != 1 {
		t.Fatalf("default paths=%d err=%v", len(want), err)
	}
	options.Algorithm = OccultationPathAlgorithmOptimized
	got, err := FindPlanetOccultationPaths(start, start.Add(24*time.Hour), OccultationMars, options)
	if err != nil || !reflect.DeepEqual(got, want) {
		t.Fatalf("explicit optimized differs from default: %v", err)
	}
}
