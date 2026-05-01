package orbit

import (
	"encoding/json"
	"math"
	"os"
	"testing"
	"time"

	"github.com/starainrt/astro/basic"
	marspkg "github.com/starainrt/astro/mars"
)

const orbitAngleToleranceDeg = 0.02
const orbitDistanceToleranceAU = 3e-4
const orbitVectorToleranceAU = 3e-4
const orbitAstrometricToleranceDeg = 0.02
const orbitAstrometricDistanceToleranceAU = 3e-4
const shanghaiLon = 121.4737
const shanghaiLat = 31.2304
const shanghaiHeightMeters = 20.0

type baselineElements struct {
	Form    string  `json:"form"`
	EpochJD float64 `json:"epoch_jd"`
	A       float64 `json:"a"`
	E       float64 `json:"e"`
	I       float64 `json:"i"`
	Omega   float64 `json:"omega"`
	W       float64 `json:"w"`
	M0      float64 `json:"m0"`
	Q       float64 `json:"q"`
	TpJD    float64 `json:"tp_jd"`
}

type baselineVector struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	Z float64 `json:"z"`
}

type baselineHeliocentric struct {
	Vector   baselineVector `json:"vector"`
	Lon      float64        `json:"lon"`
	Lat      float64        `json:"lat"`
	Distance float64        `json:"distance"`
}

type baselineGeocentric struct {
	Vector   baselineVector `json:"vector"`
	RA       float64        `json:"ra"`
	Dec      float64        `json:"dec"`
	Distance float64        `json:"distance"`
}

type baselineObservation struct {
	RA       float64 `json:"ra"`
	Dec      float64 `json:"dec"`
	Distance float64 `json:"distance"`
}

type baselineSample struct {
	JDTT                  float64              `json:"jd_tt"`
	Heliocentric          baselineHeliocentric `json:"heliocentric_j2000"`
	Geocentric            baselineGeocentric   `json:"geocentric_equatorial_j2000"`
	AstrometricGeocentric baselineObservation  `json:"astrometric_geocentric_j2000"`
	ApparentGeocentric    baselineObservation  `json:"apparent_geocentric_equatorial"`
	ApparentTopocentric   baselineObservation  `json:"apparent_topocentric_equatorial"`
}

type baselineObject struct {
	Name     string           `json:"name"`
	Elements baselineElements `json:"elements"`
	Samples  []baselineSample `json:"samples"`
}

func TestGeometricOrbitMatchesJPLBaseline(t *testing.T) {
	objects := loadOrbitBaseline(t)

	var maxHelioVectorDiffAU float64
	var maxHelioLonDiffDeg float64
	var maxHelioLatDiffDeg float64
	var maxGeoVectorDiffAU float64
	var maxGeoRADiffDeg float64
	var maxGeoDecDiffDeg float64
	var maxGeoDistanceDiffAU float64

	for _, object := range objects {
		elements := elementsFromBaseline(object)
		basicElements := toBasicElements(elements)

		for _, sample := range object.Samples {
			date := basic.JDE2DateByZone(basic.TD2UT(sample.JDTT, false), time.UTC, false)

			helVector := basic.OrbitHeliocentricXYZJ2000(sample.JDTT, basicElements)
			helVectorDiff := vectorDiffAU(helVector, sample.Heliocentric.Vector)
			if helVectorDiff > maxHelioVectorDiffAU {
				maxHelioVectorDiffAU = helVectorDiff
			}
			if helVectorDiff > orbitVectorToleranceAU {
				t.Fatalf("%s helio vector mismatch at JD %.1f: diff=%.9f AU", object.Name, sample.JDTT, helVectorDiff)
			}

			hel := HeliocentricEclipticJ2000(date, elements)
			lonDiff := angleDiffAbs(hel.Lon, sample.Heliocentric.Lon)
			if lonDiff > maxHelioLonDiffDeg {
				maxHelioLonDiffDeg = lonDiff
			}
			latDiff := math.Abs(hel.Lat - sample.Heliocentric.Lat)
			if latDiff > maxHelioLatDiffDeg {
				maxHelioLatDiffDeg = latDiff
			}
			if lonDiff > orbitAngleToleranceDeg {
				t.Fatalf("%s helio lon mismatch at JD %.1f: got %.9f want %.9f", object.Name, sample.JDTT, hel.Lon, sample.Heliocentric.Lon)
			}
			if latDiff > orbitAngleToleranceDeg {
				t.Fatalf("%s helio lat mismatch at JD %.1f: got %.9f want %.9f", object.Name, sample.JDTT, hel.Lat, sample.Heliocentric.Lat)
			}
			if distanceDiff := math.Abs(hel.Distance - sample.Heliocentric.Distance); distanceDiff > orbitDistanceToleranceAU {
				t.Fatalf("%s helio distance mismatch at JD %.1f: got %.9f want %.9f", object.Name, sample.JDTT, hel.Distance, sample.Heliocentric.Distance)
			}

			geo := GeocentricEquatorialJ2000(date, elements)
			geoVector := equatorialVectorAU(geo)
			geoVectorDiff := baselineVectorDiffAU(geoVector, sample.Geocentric.Vector)
			if geoVectorDiff > maxGeoVectorDiffAU {
				maxGeoVectorDiffAU = geoVectorDiff
			}
			if geoVectorDiff > orbitVectorToleranceAU {
				t.Fatalf("%s geo vector mismatch at JD %.1f: diff=%.9f AU", object.Name, sample.JDTT, geoVectorDiff)
			}

			raDiff := angleDiffAbs(geo.RA, sample.Geocentric.RA)
			if raDiff > maxGeoRADiffDeg {
				maxGeoRADiffDeg = raDiff
			}
			decDiff := math.Abs(geo.Dec - sample.Geocentric.Dec)
			if decDiff > maxGeoDecDiffDeg {
				maxGeoDecDiffDeg = decDiff
			}
			distanceDiff := math.Abs(geo.Distance - sample.Geocentric.Distance)
			if distanceDiff > maxGeoDistanceDiffAU {
				maxGeoDistanceDiffAU = distanceDiff
			}
			if raDiff > orbitAngleToleranceDeg {
				t.Fatalf("%s geo RA mismatch at JD %.1f: got %.9f want %.9f", object.Name, sample.JDTT, geo.RA, sample.Geocentric.RA)
			}
			if decDiff > orbitAngleToleranceDeg {
				t.Fatalf("%s geo Dec mismatch at JD %.1f: got %.9f want %.9f", object.Name, sample.JDTT, geo.Dec, sample.Geocentric.Dec)
			}
			if distanceDiff > orbitDistanceToleranceAU {
				t.Fatalf("%s geo distance mismatch at JD %.1f: got %.9f want %.9f", object.Name, sample.JDTT, geo.Distance, sample.Geocentric.Distance)
			}
		}
	}

	t.Logf("orbit geometric max diff: helVec=%.9fAU helLon=%.6fdeg helLat=%.6fdeg geoVec=%.9fAU geoRA=%.6fdeg geoDec=%.6fdeg geoDist=%.9fAU",
		maxHelioVectorDiffAU, maxHelioLonDiffDeg, maxHelioLatDiffDeg,
		maxGeoVectorDiffAU, maxGeoRADiffDeg, maxGeoDecDiffDeg, maxGeoDistanceDiffAU)
}

func TestAstrometricGeocentricMatchesJPLBaseline(t *testing.T) {
	maxRADiff, maxDecDiff, maxDistanceDiff := runObservationBaseline(
		t,
		"astrometric",
		func(date time.Time, elements Elements) EquatorialPosition {
			return AstrometricGeocentricEquatorialJ2000(date, elements)
		},
		func(sample baselineSample) baselineObservation {
			return sample.AstrometricGeocentric
		},
	)
	t.Logf("orbit astrometric max diff: RA=%.6fdeg Dec=%.6fdeg Dist=%.9fAU", maxRADiff, maxDecDiff, maxDistanceDiff)
}

func TestApparentGeocentricMatchesJPLBaseline(t *testing.T) {
	maxRADiff, maxDecDiff, maxDistanceDiff := runObservationBaseline(
		t,
		"geocentric apparent",
		func(date time.Time, elements Elements) EquatorialPosition {
			return ApparentGeocentricEquatorial(date, elements)
		},
		func(sample baselineSample) baselineObservation {
			return sample.ApparentGeocentric
		},
	)
	t.Logf("orbit geocentric apparent max diff: RA=%.6fdeg Dec=%.6fdeg Dist=%.9fAU", maxRADiff, maxDecDiff, maxDistanceDiff)
}

func TestApparentTopocentricMatchesJPLBaseline(t *testing.T) {
	maxRADiff, maxDecDiff, maxDistanceDiff := runObservationBaseline(
		t,
		"topocentric apparent",
		func(date time.Time, elements Elements) EquatorialPosition {
			return ApparentTopocentricEquatorial(date, elements, shanghaiLon, shanghaiLat, shanghaiHeightMeters)
		},
		func(sample baselineSample) baselineObservation {
			return sample.ApparentTopocentric
		},
	)
	t.Logf("orbit topocentric apparent max diff: RA=%.6fdeg Dec=%.6fdeg Dist=%.9fAU", maxRADiff, maxDecDiff, maxDistanceDiff)
}

func runObservationBaseline(
	t *testing.T,
	label string,
	gotFn func(time.Time, Elements) EquatorialPosition,
	wantFn func(baselineSample) baselineObservation,
) (maxRADiff, maxDecDiff, maxDistanceDiff float64) {
	t.Helper()
	objects := loadOrbitBaseline(t)

	for _, object := range objects {
		elements := elementsFromBaseline(object)
		for _, sample := range object.Samples {
			date := basic.JDE2DateByZone(basic.TD2UT(sample.JDTT, false), time.UTC, false)
			got := gotFn(date, elements)
			want := wantFn(sample)
			raDiff := angleDiffAbs(got.RA, want.RA)
			if raDiff > maxRADiff {
				maxRADiff = raDiff
			}
			decDiff := math.Abs(got.Dec - want.Dec)
			if decDiff > maxDecDiff {
				maxDecDiff = decDiff
			}
			distanceDiff := math.Abs(got.Distance - want.Distance)
			if distanceDiff > maxDistanceDiff {
				maxDistanceDiff = distanceDiff
			}
			if raDiff > orbitAstrometricToleranceDeg {
				t.Fatalf("%s %s RA mismatch at JD %.1f: got %.9f want %.9f", object.Name, label, sample.JDTT, got.RA, want.RA)
			}
			if decDiff > orbitAstrometricToleranceDeg {
				t.Fatalf("%s %s Dec mismatch at JD %.1f: got %.9f want %.9f", object.Name, label, sample.JDTT, got.Dec, want.Dec)
			}
			if distanceDiff > orbitAstrometricDistanceToleranceAU {
				t.Fatalf("%s %s distance mismatch at JD %.1f: got %.9f want %.9f", object.Name, label, sample.JDTT, got.Distance, want.Distance)
			}
		}
	}
	return maxRADiff, maxDecDiff, maxDistanceDiff
}

func TestSecularRatesReduceMarsDriftAgainstVSOP(t *testing.T) {
	withRates := Elements{
		EpochJD:  2451545.0,
		A:        1.52371034,
		E:        0.09339410,
		I:        1.84969142,
		Omega:    49.55953891,
		W:        -23.94362959 - 49.55953891,
		M0:       -4.55343205 - (-23.94362959),
		ADot:     0.00001847 / 36525.0,
		EDot:     0.00007882 / 36525.0,
		IDot:     -0.00813131 / 36525.0,
		OmegaDot: -0.29257343 / 36525.0,
		WDot:     (0.44441088 - (-0.29257343)) / 36525.0,
		MDot:     (19140.30268499 - 0.44441088) / 36525.0,
	}
	static := withRates
	static.ADot, static.EDot, static.IDot, static.OmegaDot, static.WDot, static.MDot = 0, 0, 0, 0, 0, 0

	cases := []time.Time{
		time.Date(1900, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(1950, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2000, 1, 1, 12, 0, 0, 0, time.UTC),
		time.Date(2050, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	for _, date := range cases {
		wantRA, wantDec := marspkg.ApparentRaDec(date)
		dynamic := ApparentGeocentricEquatorial(date, withRates)
		stale := ApparentGeocentricEquatorial(date, static)

		dynamicError := angleDiffAbs(dynamic.RA, wantRA) + math.Abs(dynamic.Dec-wantDec)
		staticError := angleDiffAbs(stale.RA, wantRA) + math.Abs(stale.Dec-wantDec)
		if dynamicError > staticError+1e-9 {
			t.Fatalf("%s dynamic elements should improve Mars drift: dynamic=%.9f static=%.9f", date.Format("2006-01-02"), dynamicError, staticError)
		}
		if date.Equal(time.Date(2000, 1, 1, 12, 0, 0, 0, time.UTC)) {
			continue
		}
		if staticError/dynamicError < 4 {
			t.Fatalf("%s Mars drift improvement too small: dynamic=%.9f static=%.9f", date.Format("2006-01-02"), dynamicError, staticError)
		}
	}
}

func TestApparentTopocentricFiniteAndReasonable(t *testing.T) {
	elements := Elements{
		EpochJD: 2461000.5,
		A:       2.765615651508659,
		E:       0.07957631994408416,
		I:       10.58788658206854,
		Omega:   80.24963090816965,
		W:       73.29975464616518,
		M0:      231.5397330043706,
	}
	date := time.Date(2025, 11, 21, 0, 0, 0, 0, time.UTC)
	geo := ApparentGeocentricEquatorial(date, elements)
	top := ApparentTopocentricEquatorial(date, elements, shanghaiLon, shanghaiLat, shanghaiHeightMeters)
	if math.IsNaN(top.RA) || math.IsNaN(top.Dec) || math.IsNaN(top.Distance) {
		t.Fatalf("topocentric result contains NaN: %+v", top)
	}
	if top.Distance <= 0 {
		t.Fatalf("unexpected topocentric distance: %.12f", top.Distance)
	}
	if math.Abs(top.Distance-geo.Distance) > 5e-5 {
		t.Fatalf("topocentric distance shift unexpectedly large: geo=%.12f top=%.12f", geo.Distance, top.Distance)
	}
	if angleDiffAbs(top.RA, geo.RA) > 1 || math.Abs(top.Dec-geo.Dec) > 1 {
		t.Fatalf("topocentric shift unexpectedly large: geo=%+v top=%+v", geo, top)
	}
	if angleDiffAbs(top.RA, geo.RA) == 0 && math.Abs(top.Dec-geo.Dec) == 0 {
		t.Fatalf("topocentric correction should not be identically zero")
	}
}

func TestMeanMotionAndAnomaliesAreFinite(t *testing.T) {
	elements := Elements{
		EpochJD: 2461000.5,
		A:       2.765615651508659,
		E:       0.07957631994408416,
		I:       10.58788658206854,
		Omega:   80.24963090816965,
		W:       73.29975464616518,
		M0:      231.5397330043706,
	}
	date := time.Date(2025, 11, 21, 0, 0, 0, 0, time.UTC)

	meanMotion := MeanMotion(elements)
	if math.IsNaN(meanMotion) || math.IsInf(meanMotion, 0) || meanMotion <= 0 {
		t.Fatalf("invalid mean motion: %.18f", meanMotion)
	}

	meanAnomaly := MeanAnomaly(date, elements)
	trueAnomaly := TrueAnomaly(date, elements)
	for name, value := range map[string]float64{"mean": meanAnomaly, "true": trueAnomaly} {
		if math.IsNaN(value) || math.IsInf(value, 0) || value < 0 || value >= 360 {
			t.Fatalf("%s anomaly out of range: %.18f", name, value)
		}
	}
}

func TestObservationHelpersMatchTopocentricCoordinates(t *testing.T) {
	elements := sampleObservationElements()
	date := time.Date(2025, 11, 21, 20, 0, 0, 0, time.FixedZone("CST", 8*3600))

	altitude := Altitude(date, elements, shanghaiLon, shanghaiLat, shanghaiHeightMeters)
	zenith := Zenith(date, elements, shanghaiLon, shanghaiLat, shanghaiHeightMeters)
	azimuth := Azimuth(date, elements, shanghaiLon, shanghaiLat, shanghaiHeightMeters)
	hourAngle := HourAngle(date, elements, shanghaiLon, shanghaiLat, shanghaiHeightMeters)
	topocentric := ApparentTopocentricEquatorial(date, elements, shanghaiLon, shanghaiLat, shanghaiHeightMeters)

	for name, value := range map[string]float64{
		"altitude":  altitude,
		"zenith":    zenith,
		"azimuth":   azimuth,
		"hourAngle": hourAngle,
		"ra":        topocentric.RA,
		"dec":       topocentric.Dec,
	} {
		if math.IsNaN(value) || math.IsInf(value, 0) {
			t.Fatalf("%s is not finite: %.18f", name, value)
		}
	}

	jde := basic.Date2JDE(date)
	_, offsetSeconds := date.Zone()
	timezone := float64(offsetSeconds) / 3600.0
	siderealLongitude := normalize360(basic.ApparentSiderealTime(jde-timezone/24.0)*15 + shanghaiLon)
	wantHourAngle := normalize360(siderealLongitude - topocentric.RA)
	if angleDiffAbs(hourAngle, wantHourAngle) > 1e-9 {
		t.Fatalf("hour angle mismatch: got %.12f want %.12f", hourAngle, wantHourAngle)
	}

	wantAltitude := math.Asin(
		math.Sin(shanghaiLat*math.Pi/180)*math.Sin(topocentric.Dec*math.Pi/180)+
			math.Cos(topocentric.Dec*math.Pi/180)*math.Cos(shanghaiLat*math.Pi/180)*math.Cos(wantHourAngle*math.Pi/180),
	) * 180 / math.Pi
	if math.Abs(altitude-wantAltitude) > 1e-9 {
		t.Fatalf("altitude mismatch: got %.12f want %.12f", altitude, wantAltitude)
	}
	wantZenith := 90 - wantAltitude
	if math.Abs(zenith-wantZenith) > 1e-9 {
		t.Fatalf("zenith mismatch: got %.12f want %.12f", zenith, wantZenith)
	}

	wantAzimuth := sphericalAzimuthFromHourAngle(wantHourAngle, topocentric.Dec, shanghaiLat)
	if angleDiffAbs(azimuth, wantAzimuth) > 1e-9 {
		t.Fatalf("azimuth mismatch: got %.12f want %.12f", azimuth, wantAzimuth)
	}
}

func TestCulminationTimeMaximizesAltitude(t *testing.T) {
	elements := sampleObservationElements()
	date := time.Date(2025, 11, 21, 0, 0, 0, 0, time.FixedZone("CST", 8*3600))

	culmination := CulminationTime(date, elements, shanghaiLon, shanghaiLat, shanghaiHeightMeters)
	before := Altitude(culmination.Add(-5*time.Minute), elements, shanghaiLon, shanghaiLat, shanghaiHeightMeters)
	at := Altitude(culmination, elements, shanghaiLon, shanghaiLat, shanghaiHeightMeters)
	after := Altitude(culmination.Add(5*time.Minute), elements, shanghaiLon, shanghaiLat, shanghaiHeightMeters)
	if at < before || at < after {
		t.Fatalf("culmination should maximize altitude: before=%.9f at=%.9f after=%.9f", before, at, after)
	}
	if angleDiffAbs(HourAngle(culmination, elements, shanghaiLon, shanghaiLat, shanghaiHeightMeters), 0) > 0.02 {
		t.Fatalf("culmination hour angle should be near zero")
	}
}

func TestRiseSetTimesReachStandardAltitude(t *testing.T) {
	elements := sampleObservationElements()
	date := time.Date(2025, 11, 21, 0, 0, 0, 0, time.FixedZone("CST", 8*3600))

	rise, err := RiseTime(date, elements, shanghaiLon, shanghaiLat, shanghaiHeightMeters, true)
	if err != nil {
		t.Fatalf("rise time failed: %v", err)
	}
	set, err := SetTime(date, elements, shanghaiLon, shanghaiLat, shanghaiHeightMeters, true)
	if err != nil {
		t.Fatalf("set time failed: %v", err)
	}
	culmination := CulminationTime(date, elements, shanghaiLon, shanghaiLat, shanghaiHeightMeters)
	targetAltitude := basic.StandardAltitudePlanet(1, shanghaiHeightMeters, shanghaiLat)

	riseAltitude := Altitude(rise, elements, shanghaiLon, shanghaiLat, shanghaiHeightMeters)
	setAltitude := Altitude(set, elements, shanghaiLon, shanghaiLat, shanghaiHeightMeters)
	if math.Abs(riseAltitude-targetAltitude) > 0.03 {
		t.Fatalf("rise altitude mismatch: got %.9f want %.9f", riseAltitude, targetAltitude)
	}
	if math.Abs(setAltitude-targetAltitude) > 0.03 {
		t.Fatalf("set altitude mismatch: got %.9f want %.9f", setAltitude, targetAltitude)
	}
	if !rise.Before(culmination) {
		t.Fatalf("rise should precede culmination: rise=%s culmination=%s", rise, culmination)
	}
	if !culmination.Before(set) {
		t.Fatalf("culmination should precede set: culmination=%s set=%s", culmination, set)
	}
}

func loadOrbitBaseline(t *testing.T) []baselineObject {
	t.Helper()
	data, err := os.ReadFile("testdata/orbit_baseline.json")
	if err != nil {
		t.Fatalf("read baseline: %v", err)
	}
	var objects []baselineObject
	if err := json.Unmarshal(data, &objects); err != nil {
		t.Fatalf("decode baseline: %v", err)
	}
	return objects
}

func elementsFromBaseline(object baselineObject) Elements {
	if object.Elements.Form == "perihelion" {
		return Elements{
			E:     object.Elements.E,
			I:     object.Elements.I,
			Omega: object.Elements.Omega,
			W:     object.Elements.W,
			Q:     object.Elements.Q,
			TpJD:  object.Elements.TpJD,
		}
	}
	return Elements{
		EpochJD: object.Elements.EpochJD,
		A:       object.Elements.A,
		E:       object.Elements.E,
		I:       object.Elements.I,
		Omega:   object.Elements.Omega,
		W:       object.Elements.W,
		M0:      object.Elements.M0,
	}
}

func vectorDiffAU(vector basic.Vector3, want baselineVector) float64 {
	dx := vector[0] - want.X
	dy := vector[1] - want.Y
	dz := vector[2] - want.Z
	return math.Sqrt(dx*dx + dy*dy + dz*dz)
}

func baselineVectorDiffAU(got, want baselineVector) float64 {
	dx := got.X - want.X
	dy := got.Y - want.Y
	dz := got.Z - want.Z
	return math.Sqrt(dx*dx + dy*dy + dz*dz)
}

func angleDiffAbs(got, want float64) float64 {
	diff := math.Abs(got - want)
	if diff > 180 {
		diff = 360 - diff
	}
	return diff
}

func equatorialVectorAU(position EquatorialPosition) baselineVector {
	raRad := position.RA * math.Pi / 180
	decRad := position.Dec * math.Pi / 180
	cosDec := math.Cos(decRad)
	return baselineVector{
		X: position.Distance * cosDec * math.Cos(raRad),
		Y: position.Distance * cosDec * math.Sin(raRad),
		Z: position.Distance * math.Sin(decRad),
	}
}

func sampleObservationElements() Elements {
	return Elements{
		EpochJD: 2461000.5,
		A:       2.765615651508659,
		E:       0.07957631994408416,
		I:       10.58788658206854,
		Omega:   80.24963090816965,
		W:       73.29975464616518,
		M0:      231.5397330043706,
	}
}

func normalize360(value float64) float64 {
	value = math.Mod(value, 360)
	if value < 0 {
		value += 360
	}
	return value
}

func sphericalAzimuthFromHourAngle(hourAngle, dec, lat float64) float64 {
	tanAzimuth := math.Sin(hourAngle*math.Pi/180) / (math.Cos(hourAngle*math.Pi/180)*math.Sin(lat*math.Pi/180) - math.Tan(dec*math.Pi/180)*math.Cos(lat*math.Pi/180))
	azimuth := math.Atan(tanAzimuth) * 180 / math.Pi
	if azimuth < 0 {
		if hourAngle/15 < 12 {
			return azimuth + 360
		}
		return azimuth + 180
	}
	if hourAngle/15 < 12 {
		return azimuth + 180
	}
	return azimuth
}
