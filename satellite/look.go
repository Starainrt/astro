package satellite

import (
	"math"
	"time"

	"github.com/starainrt/astro/basic"
	"github.com/starainrt/astro/internal/ephem"
)

// Observer WGS84观测地点，高度为米 / WGS84 observing site with height in metres.
type Observer = basic.Observer

// ObserveOptions 站点视位置选项，零值采用标准大气折射 / observation options; zero selects standard atmospheric refraction.
type ObserveOptions struct {
	NoRefraction      bool
	PressureHPa       float64 // PressureHPa 为0时采用1010 hPa及10 °C / zero selects 1010 hPa and 10 degrees Celsius.
	TemperatureC      float64
	StandardMagnitude *float64
}

// Look 卫星站心几何视位置与光照 / geometric topocentric satellite position and illumination.
type Look struct {
	Time             time.Time
	Azimuth          float64 // Azimuth 正北起顺时针，度 / degrees clockwise from north.
	Elevation        float64
	RangeKM          float64
	RangeRateKMS     float64
	RA, Dec          float64 // RA,Dec 站心真赤道of date方向，度 / topocentric true equatorial direction of date in degrees.
	Illumination     Illumination
	SunCenterVisible bool
	PhaseAngle       float64 // PhaseAngle 卫星处太阳与观测者的夹角，度 / Sun-observer angle at the satellite in degrees.
	Magnitude        float64 // Magnitude 无标准星等或本影中为NaN / NaN without a standard magnitude or in umbra.
}

// LookAt 计算卫星站心视位置，保留输入时区 / compute a topocentric satellite observation preserving the input time zone.
func (s *Satellite) LookAt(e Elements, obs Observer, t time.Time, opt ObserveOptions) (Look, error) {
	position, err := observerPosition(obs)
	if err != nil || !validObserveOptions(opt) {
		return Look{}, ErrInvalidInput
	}
	orbit, err := newSatelliteOrbit(e)
	if err != nil {
		return Look{}, err
	}
	return s.lookAtOrbit(orbit, obs, position, t, opt)
}

func (s *Satellite) lookAtOrbit(orbit *satelliteOrbit, obs Observer, position [3]float64, t time.Time, opt ObserveOptions) (Look, error) {
	state, err := orbit.stateAt(t)
	if err != nil {
		return Look{}, err
	}
	r, v := earthFixed(state)
	d := subtractVector(r, position)
	distance := vectorNorm(d)
	if distance == 0 || !finite(distance) {
		return Look{}, ErrInvalidInput
	}
	az, alt := horizontalDirection(d, obs)
	if !opt.NoRefraction {
		pressure, temperature := opt.PressureHPa, opt.TemperatureC
		if pressure == 0 {
			pressure, temperature = 1010, 10
		}
		alt += basic.RefractionFromTrueAltitude(alt, pressure, temperature)
	}
	jdUTC := basic.Date2JD(t.UTC())
	jdUT1 := basic.UTC2UT1(jdUTC)
	jdTT := basic.UTC2TT(jdUTC)
	gast := basic.ApparentSiderealTime(jdUT1) * 15 * degreesToRadians
	eq := rotateZ(-gast, d)
	ra := math.Atan2(eq[1], eq[0]) / degreesToRadians
	if ra < 0 {
		ra += 360
	}
	dec := math.Atan2(eq[2], math.Hypot(eq[0], eq[1])) / degreesToRadians
	var inst *ephem.Instance
	if s != nil {
		inst = s.inst
	}
	sun := apparentSunPosition(jdTT, gast, inst)
	kind, visible := illuminationAt(r, sun)
	phase := vectorAngle(subtractVector(sun, r), subtractVector(position, r)) / degreesToRadians
	fraction := 1.0
	if kind == IlluminationUmbra {
		fraction = 0
	} else if kind == IlluminationPenumbra {
		fraction = illuminatedFraction(earthAngularRadius(r), math.Asin(695700/vectorNorm(subtractVector(sun, r))), vectorAngle([3]float64{-r[0], -r[1], -r[2]}, subtractVector(sun, r)))
	}
	result := Look{Time: t, Azimuth: az, Elevation: alt, RangeKM: distance, RangeRateKMS: dotVector(d, v) / distance, RA: ra, Dec: dec, Illumination: kind, SunCenterVisible: visible, PhaseAngle: phase, Magnitude: satelliteMagnitude(distance, phase*degreesToRadians, fraction, opt.StandardMagnitude)}
	for _, value := range []float64{result.Azimuth, result.Elevation, result.RangeKM, result.RangeRateKMS, result.RA, result.Dec, result.PhaseAngle} {
		if !finite(value) {
			return Look{}, ErrInvalidInput
		}
	}
	return result, nil
}

func validObserveOptions(opt ObserveOptions) bool {
	if !finite(opt.PressureHPa) || !finite(opt.TemperatureC) || opt.StandardMagnitude != nil && !finite(*opt.StandardMagnitude) {
		return false
	}
	return opt.NoRefraction || opt.PressureHPa == 0 || finite(basic.RefractionFromTrueAltitude(0, opt.PressureHPa, opt.TemperatureC))
}

func observerPosition(obs Observer) ([3]float64, error) {
	if !finite(obs.Longitude) || !finite(obs.Latitude) || !finite(obs.Height) || math.Abs(obs.Latitude) > 90 {
		return [3]float64{}, ErrInvalidInput
	}
	sinLat, cosLat := math.Sincos(obs.Latitude * degreesToRadians)
	sinLon, cosLon := math.Sincos(obs.Longitude * degreesToRadians)
	e2 := earthFlattening * (2 - earthFlattening)
	n := earthRadiusKM / math.Sqrt(1-e2*sinLat*sinLat)
	h := obs.Height / 1000
	return [3]float64{(n + h) * cosLat * cosLon, (n + h) * cosLat * sinLon, (n*(1-e2) + h) * sinLat}, nil
}

func horizontalDirection(d [3]float64, obs Observer) (az, alt float64) {
	sinLat, cosLat := math.Sincos(obs.Latitude * degreesToRadians)
	sinLon, cosLon := math.Sincos(obs.Longitude * degreesToRadians)
	east := -sinLon*d[0] + cosLon*d[1]
	north := -sinLat*cosLon*d[0] - sinLat*sinLon*d[1] + cosLat*d[2]
	up := cosLat*cosLon*d[0] + cosLat*sinLon*d[1] + sinLat*d[2]
	az = math.Atan2(east, north) / degreesToRadians
	if az < 0 {
		az += 360
	}
	alt = math.Atan2(up, math.Hypot(east, north)) / degreesToRadians
	return
}

func apparentSunPosition(jdTT, gast float64, inst *ephem.Instance) [3]float64 {
	ra, dec := basic.HSunApparentRaDecNWithInstance(jdTT, -1, inst)
	distance := basic.EarthAwayNWithInstance(jdTT, -1, inst) * 149597870.7
	sinRA, cosRA := math.Sincos(ra * degreesToRadians)
	sinDec, cosDec := math.Sincos(dec * degreesToRadians)
	return rotateZ(gast, [3]float64{distance * cosDec * cosRA, distance * cosDec * sinRA, distance * sinDec})
}
