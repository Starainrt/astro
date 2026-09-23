package basic

import (
	"math"
	"testing"
	"time"
)

func TestSolarEclipseLocalEphemerisBoundedError(t *testing.T) {
	events := []struct {
		name string
		seed float64
	}{
		{"2010-01-15", JDCalc(2010, 1, 15)},
		{"2014-04-29", JDCalc(2014, 4, 29)},
		{"2023-04-20", JDCalc(2023, 4, 20)},
		{"2031-05-21", JDCalc(2031, 5, 21)},
		{"2309-06-09", JDCalc(2309, 6, 9)},
	}
	for _, event := range events {
		t.Run(event.name, func(t *testing.T) {
			center := CalcMoonSHByJDE(event.seed, 0)
			ephemeris := newSolarEclipseLocalEphemeris(center)
			for offset := -4.5; offset <= 4.5; offset += 0.25 {
				jd := center + offset/24
				approxSun, approxMoon, ok := ephemeris.equatorialAt(jd)
				if !ok {
					t.Fatalf("interpolator rejected in-window time %.3fh", offset)
				}
				exactSun, exactMoon := solarEclipseSunMoonEquatorial(jd)
				for name, pair := range map[string][2][3]float64{
					"sun":  {solarEclipseLLRToXYZ(approxSun[0], approxSun[1], approxSun[2]), solarEclipseLLRToXYZ(exactSun[0], exactSun[1], exactSun[2])},
					"moon": {solarEclipseLLRToXYZ(approxMoon[0], approxMoon[1], approxMoon[2]), solarEclipseLLRToXYZ(exactMoon[0], exactMoon[1], exactMoon[2])},
				} {
					if errorKM := vectorDifferenceKM(pair[0], pair[1]); errorKM > 1 {
						t.Fatalf("%s interpolation error at %.3fh = %.6f km", name, offset, errorKM)
					}
				}
			}
			if _, _, ok := ephemeris.equatorialAt(center + 6.1/24); ok {
				t.Fatal("interpolator should reject times outside its bounded window")
			}
		})
	}
}

func TestOccultationLocalEphemerisBoundedError(t *testing.T) {
	star := hr4799OccultationCoordinateForTest()
	starTT := occultationTimeToTT(time.Date(2025, 6, 5, 20, 0, 0, 0, time.UTC))
	starEphemeris := newStarOccultationLocalEphemeris(starTT, star)
	for offset := -4.5; offset <= 4.5; offset += 0.25 {
		jd := starTT + offset/24
		got, ok := starEphemeris.stateAt(jd)
		if !ok {
			t.Fatalf("star interpolator rejected in-window time %.3fh", offset)
		}
		want := starOccultationEphemerisStateAt(jd, star)
		if difference := angularSeparationDegrees(got.moonRA, got.moonDec, want.moonRA, want.moonDec); difference > 1e-5 {
			t.Fatalf("moon direction interpolation error at %.3fh = %.9f deg", offset, difference)
		}
		if difference := angularSeparationDegrees(got.starRA, got.starDec, want.starRA, want.starDec); difference > 1e-5 {
			t.Fatalf("star direction interpolation error at %.3fh = %.9f deg", offset, difference)
		}
	}

	config, ok := planetOccultationConfigFor(OccultationSaturn)
	if !ok {
		t.Fatal("Saturn occultation config is unavailable")
	}
	planetTT := occultationTimeToTT(time.Date(2024, 8, 21, 2, 40, 0, 0, time.UTC))
	planetEphemeris := newPlanetOccultationLocalEphemeris(planetTT, config)
	for offset := -4.5; offset <= 4.5; offset += 0.25 {
		jd := planetTT + offset/24
		got, ok := planetEphemeris.stateAt(jd)
		if !ok {
			t.Fatalf("planet interpolator rejected in-window time %.3fh", offset)
		}
		want := planetOccultationEphemerisStateAt(jd, config)
		if difference := angularSeparationDegrees(got.moonRA, got.moonDec, want.moonRA, want.moonDec); difference > 1e-5 {
			t.Fatalf("moon direction interpolation error at %.3fh = %.9f deg", offset, difference)
		}
		if difference := angularSeparationDegrees(got.planetRA, got.planetDec, want.planetRA, want.planetDec); difference > 1e-5 {
			t.Fatalf("planet direction interpolation error at %.3fh = %.9f deg", offset, difference)
		}
		if math.Abs(got.planetDistanceKM-want.planetDistanceKM) > 100 {
			t.Fatalf("planet distance interpolation error at %.3fh = %.6f km", offset, math.Abs(got.planetDistanceKM-want.planetDistanceKM))
		}
	}
}

func TestOccultationLocalEphemerisFullWindowBoundedError(t *testing.T) {
	star := hr4799OccultationCoordinateForTest()
	starTT := occultationTimeToTT(time.Date(2025, 6, 5, 20, 0, 0, 0, time.UTC))
	starEphemeris := newStarOccultationLocalEphemeris(starTT, star)
	config, ok := planetOccultationConfigFor(OccultationSaturn)
	if !ok {
		t.Fatal("Saturn occultation config is unavailable")
	}
	planetTT := occultationTimeToTT(time.Date(2024, 8, 21, 2, 40, 0, 0, time.UTC))
	planetEphemeris := newPlanetOccultationLocalEphemeris(planetTT, config)
	for _, offset := range []float64{-47.5, -40, -24, 24, 40, 47.5} {
		starJD := starTT + offset/24
		gotStar, starOK := starEphemeris.stateAt(starJD)
		wantStar := starOccultationEphemerisStateAt(starJD, star)
		if !starOK {
			t.Fatalf("star interpolator rejected in-window time %.3fh", offset)
		}
		if difference := angularSeparationDegrees(gotStar.moonRA, gotStar.moonDec, wantStar.moonRA, wantStar.moonDec); difference > 1e-5 {
			t.Fatalf("star-event moon direction error at %.3fh = %.9f deg", offset, difference)
		}
		if difference := angularSeparationDegrees(gotStar.starRA, gotStar.starDec, wantStar.starRA, wantStar.starDec); difference > 1e-5 {
			t.Fatalf("star direction error at %.3fh = %.9f deg", offset, difference)
		}

		planetJD := planetTT + offset/24
		gotPlanet, planetOK := planetEphemeris.stateAt(planetJD)
		wantPlanet := planetOccultationEphemerisStateAt(planetJD, config)
		if !planetOK {
			t.Fatalf("planet interpolator rejected in-window time %.3fh", offset)
		}
		if difference := angularSeparationDegrees(gotPlanet.moonRA, gotPlanet.moonDec, wantPlanet.moonRA, wantPlanet.moonDec); difference > 1e-5 {
			t.Fatalf("planet-event moon direction error at %.3fh = %.9f deg", offset, difference)
		}
		if difference := angularSeparationDegrees(gotPlanet.planetRA, gotPlanet.planetDec, wantPlanet.planetRA, wantPlanet.planetDec); difference > 1e-5 {
			t.Fatalf("planet direction error at %.3fh = %.9f deg", offset, difference)
		}
		if difference := math.Abs(gotPlanet.planetDistanceKM - wantPlanet.planetDistanceKM); difference > 100 {
			t.Fatalf("planet distance error at %.3fh = %.6f km", offset, difference)
		}
	}
}

func vectorDifferenceKM(a, b [3]float64) float64 {
	return math.Sqrt((a[0]-b[0])*(a[0]-b[0]) + (a[1]-b[1])*(a[1]-b[1]) + (a[2]-b[2])*(a[2]-b[2]))
}

// 矢量回读的赤经必须落在 [0,360)：atan2 给 (−180,180]，而精确分支
// （LoBoToRaDec、starMeanToApparentRaDec）与 rise/set 的矢量回读都在 [0,360)，
// 不归一会让同一时刻的密集/精确状态相差 360 度，误导比较与日志。
func TestOccultationPathVectorRaDecKeepsNormalizedRA(t *testing.T) {
	for _, want := range []float64{0.5, 90, 189.1975, 270, 359.9} {
		vector := occultationPathRaDecVector(want, -5.8, 2.5)
		got, dec, distance, ok := occultationPathVectorRaDec(vector)
		if !ok {
			t.Fatalf("RA %.4f 的回读失败", want)
		}
		if math.Abs(got-want) > 1e-9 {
			t.Fatalf("RA %.4f 回读为 %.4f，want [0,360) 内的同一读数", want, got)
		}
		if math.Abs(dec+5.8) > 1e-9 || math.Abs(distance-2.5) > 1e-12 {
			t.Fatalf("RA %.4f 的赤纬/距离 = %.6f / %.6f，want -5.8 / 2.5", want, dec, distance)
		}
	}
	if _, _, _, ok := occultationPathVectorRaDec(occultationPathVector{}); ok {
		t.Fatal("零矢量应判为无效")
	}
}
