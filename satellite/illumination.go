package satellite

import "math"

// Illumination 卫星处的太阳遮蔽状态 / solar illumination at the satellite.
type Illumination int

const (
	// IlluminationSunlit 太阳圆盘未被地球遮挡 / unobscured solar disc.
	IlluminationSunlit Illumination = iota
	// IlluminationPenumbra 太阳圆盘被地球部分遮挡 / partially obscured solar disc.
	IlluminationPenumbra
	// IlluminationUmbra 太阳圆盘被地球完全遮挡 / fully obscured solar disc.
	IlluminationUmbra
)

func earthAngularRadius(r [3]float64) float64 {
	return math.Asin(math.Min(1, earthRadiusKM/vectorNorm(r)))
}

func illuminationAt(r, sun [3]float64) (Illumination, bool) {
	toEarth := [3]float64{-r[0], -r[1], -r[2]}
	toSun := subtractVector(sun, r)
	earthAngle := earthAngularRadius(r)
	sunAngle := math.Asin(695700 / vectorNorm(toSun))
	separation := vectorAngle(toEarth, toSun)
	visible := separation > earthAngle
	if separation >= earthAngle+sunAngle {
		return IlluminationSunlit, visible
	}
	if separation <= earthAngle-sunAngle {
		return IlluminationUmbra, visible
	}
	return IlluminationPenumbra, visible
}

func subtractVector(a, b [3]float64) [3]float64 {
	return [3]float64{a[0] - b[0], a[1] - b[1], a[2] - b[2]}
}

func dotVector(a, b [3]float64) float64 { return a[0]*b[0] + a[1]*b[1] + a[2]*b[2] }

func vectorAngle(a, b [3]float64) float64 {
	cross := [3]float64{a[1]*b[2] - a[2]*b[1], a[2]*b[0] - a[0]*b[2], a[0]*b[1] - a[1]*b[0]}
	return math.Atan2(vectorNorm(cross), dotVector(a, b))
}

func illuminatedFraction(earthRad, sunRad, separationRad float64) float64 {
	if separationRad >= earthRad+sunRad {
		return 1
	}
	if separationRad <= math.Abs(earthRad-sunRad) {
		if earthRad >= sunRad {
			return 0
		}
		return 1 - earthRad*earthRad/(sunRad*sunRad)
	}
	x := (separationRad*separationRad + sunRad*sunRad - earthRad*earthRad) / (2 * separationRad)
	y := math.Sqrt(math.Max(0, sunRad*sunRad-x*x))
	sunSegmentRad := math.Atan2(y, x)
	earthSegmentRad := math.Atan2(y, separationRad-x)
	segment := func(thetaRad float64) float64 {
		if thetaRad < .01 {
			q := thetaRad * thetaRad
			return thetaRad * q * (2.0/3 - q*(2.0/15-q*(4.0/315-q*2.0/2835)))
		}
		return thetaRad - math.Sin(thetaRad)*math.Cos(thetaRad)
	}
	overlap := sunRad*sunRad*segment(sunSegmentRad) + earthRad*earthRad*segment(earthSegmentRad)
	return math.Max(0, math.Min(1, 1-overlap/(math.Pi*sunRad*sunRad)))
}

func satelliteMagnitude(distance, phaseRad, fraction float64, standard *float64) float64 {
	if standard == nil || fraction <= 0 {
		return math.NaN()
	}
	flux := math.Sin(phaseRad) + (math.Pi-phaseRad)*math.Cos(phaseRad)
	return *standard + 5*math.Log10(distance/1000) - 2.5*math.Log10(flux) - 2.5*math.Log10(fraction)
}
