package basic

import (
	"math"
	"testing"
)

func TestSaturnRingParametersMeeusExample(t *testing.T) {
	earthLatitude, sunLatitude, positionAngle, deltaU, majorAxis, minorAxis := SaturnRingParameters(2448972.50068)

	assertClose(t, "EarthLatitude", earthLatitude, 16.4415, 0.0001)
	assertClose(t, "SunLatitude", sunLatitude, 14.6782, 0.0001)
	assertClose(t, "PositionAngle", positionAngle, 6.7406, 0.0001)
	assertClose(t, "DeltaU", deltaU, 4.1991, 0.0001)
	assertClose(t, "MajorAxis", majorAxis, 35.8687, 0.0001)
	assertClose(t, "MinorAxis", minorAxis, 10.1521, 0.0001)
}

func TestSaturnRingNFullMatchesDefault(t *testing.T) {
	jd := 2448972.50068

	defaultEarthLatitude, defaultSunLatitude, defaultPositionAngle, defaultDeltaU, defaultMajorAxis, defaultMinorAxis := SaturnRingParameters(jd)
	truncatedEarthLatitude, truncatedSunLatitude, truncatedPositionAngle, truncatedDeltaU, truncatedMajorAxis, truncatedMinorAxis := SaturnRingParametersN(jd, -1)

	assertSameFloat(t, "EarthLatitude", defaultEarthLatitude, truncatedEarthLatitude)
	assertSameFloat(t, "SunLatitude", defaultSunLatitude, truncatedSunLatitude)
	assertSameFloat(t, "PositionAngle", defaultPositionAngle, truncatedPositionAngle)
	assertSameFloat(t, "DeltaU", defaultDeltaU, truncatedDeltaU)
	assertSameFloat(t, "MajorAxis", defaultMajorAxis, truncatedMajorAxis)
	assertSameFloat(t, "MinorAxis", defaultMinorAxis, truncatedMinorAxis)
	assertSameFloat(t, "SaturnRingB", SaturnRingB(jd), SaturnRingBN(jd, -1))
	assertSameFloat(t, "SaturnRingSunB", SaturnRingSunB(jd), SaturnRingSunBN(jd, -1))
	assertSameFloat(t, "SaturnRingPositionAngle", SaturnRingPositionAngle(jd), SaturnRingPositionAngleN(jd, -1))
	assertSameFloat(t, "SaturnRingDeltaU", SaturnRingDeltaU(jd), SaturnRingDeltaUN(jd, -1))
}

func assertClose(t *testing.T, name string, got, want, tolerance float64) {
	t.Helper()
	if math.Abs(got-want) > tolerance {
		t.Fatalf("%s mismatch: got %.12f want %.12f tolerance %.12f", name, got, want, tolerance)
	}
}

func assertSameFloat(t *testing.T, name string, got, want float64) {
	t.Helper()
	if math.Float64bits(got) != math.Float64bits(want) {
		t.Fatalf("%s full-n mismatch: got %.18f want %.18f", name, got, want)
	}
}
