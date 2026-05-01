package basic

import (
	"math"
	"testing"
)

func TestOrbitCircularStateAtEpoch(t *testing.T) {
	elements := OrbitElements{
		EpochJD: orbitReferenceJD,
		A:       2,
		E:       0,
		I:       0,
		Omega:   0,
		W:       0,
		M0:      0,
	}

	vector := OrbitHeliocentricXYZJ2000(orbitReferenceJD, elements)
	assertSameFloat(t, "x", vector[0], 2)
	assertSameFloat(t, "y", vector[1], 0)
	assertSameFloat(t, "z", vector[2], 0)

	lon, lat, distance := OrbitHeliocentricEclipticJ2000(orbitReferenceJD, elements)
	assertSameFloat(t, "lon", lon, 0)
	assertSameFloat(t, "lat", lat, 0)
	assertSameFloat(t, "distance", distance, 2)
}

func TestOrbitQuarterPeriodOnCircularOrbit(t *testing.T) {
	elements := OrbitElements{
		EpochJD: orbitReferenceJD,
		A:       1,
		E:       0,
		I:       0,
		Omega:   0,
		W:       0,
		M0:      0,
	}
	quarterPeriodDays := 90 / OrbitMeanMotion(elements)
	vector := OrbitHeliocentricXYZJ2000(orbitReferenceJD+quarterPeriodDays, elements)

	if math.Abs(vector[0]) > 1e-10 || math.Abs(vector[1]-1) > 1e-10 || math.Abs(vector[2]) > 1e-10 {
		t.Fatalf("quarter-period vector mismatch: got %+v", vector)
	}
}

func TestOrbitMeanAndTrueAnomalyCircularMatch(t *testing.T) {
	elements := OrbitElements{
		EpochJD: orbitReferenceJD,
		A:       1.523679,
		E:       0,
		I:       1.85,
		Omega:   49.5,
		W:       286.5,
		M0:      123.4,
	}
	jd := orbitReferenceJD + 123.456
	meanAnomaly := OrbitMeanAnomaly(jd, elements)
	trueAnomaly := OrbitTrueAnomaly(jd, elements)
	if math.Abs(meanAnomaly-trueAnomaly) > 1e-12 {
		t.Fatalf("circular mean/true anomaly mismatch: mean=%.18f true=%.18f", meanAnomaly, trueAnomaly)
	}
}

func TestPerihelionFormMatchesClassicalEllipticState(t *testing.T) {
	classical := OrbitElements{
		EpochJD: 2457305.5,
		A:       3.462249489765068,
		E:       0.6409081306555051,
		I:       7.040294906760007,
		Omega:   50.13557380441372,
		W:       12.79824973415729,
		M0:      8.859927418758764,
	}
	perihelion := OrbitElements{
		Q:     1.243265641416762,
		E:     classical.E,
		I:     classical.I,
		Omega: classical.Omega,
		W:     classical.W,
		TpJD:  2457247.588657863465,
	}
	jd := 2457308.5
	classicalVector := OrbitHeliocentricXYZJ2000(jd, classical)
	perihelionVector := OrbitHeliocentricXYZJ2000(jd, perihelion)
	for i := range classicalVector {
		if math.Abs(classicalVector[i]-perihelionVector[i]) > 1e-11 {
			t.Fatalf("component %d mismatch: classical=%.18f perihelion=%.18f", i, classicalVector[i], perihelionVector[i])
		}
	}
}

func TestParabolicPerihelionStateAtTp(t *testing.T) {
	elements := OrbitElements{Q: 1, E: 1, I: 0, Omega: 0, W: 0, TpJD: orbitReferenceJD}
	vector := OrbitHeliocentricXYZJ2000(orbitReferenceJD, elements)
	assertSameFloat(t, "x", vector[0], 1)
	assertSameFloat(t, "y", vector[1], 0)
	assertSameFloat(t, "z", vector[2], 0)
	if !math.IsNaN(OrbitMeanAnomaly(orbitReferenceJD, elements)) {
		t.Fatalf("parabolic mean anomaly should be NaN")
	}
}

func TestHyperbolicOrbitProducesFiniteState(t *testing.T) {
	elements := OrbitElements{Q: 0.5, E: 1.2, I: 12, Omega: 30, W: 45, TpJD: orbitReferenceJD}
	vector := OrbitHeliocentricXYZJ2000(orbitReferenceJD+20, elements)
	for i, component := range vector {
		if math.IsNaN(component) || math.IsInf(component, 0) {
			t.Fatalf("component %d not finite: %.18f", i, component)
		}
	}
}

func TestOrbitInvalidEllipticElementsReturnNaN(t *testing.T) {
	elements := OrbitElements{EpochJD: orbitReferenceJD, A: 1, E: 1.1}
	if !math.IsNaN(OrbitMeanMotion(elements)) {
		t.Fatalf("expected NaN mean motion for invalid elements")
	}
	vector := OrbitHeliocentricXYZJ2000(orbitReferenceJD, elements)
	for i, component := range vector {
		if !math.IsNaN(component) {
			t.Fatalf("component %d expected NaN, got %.18f", i, component)
		}
	}
}
