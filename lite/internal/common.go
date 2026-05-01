package internal

import (
	"errors"
	"math"

	. "github.com/starainrt/astro/tools"
)

const (
	SynodicMonthDays = 29.530588853
)

var (
	ErrNeverRise     = errors.New("rise event does not occur on this date")
	ErrNeverSet      = errors.New("set event does not occur on this date")
	ErrNotOnThisDate = errors.New("rise/set event occurs on adjacent date")
)

func MeanObliquity(jd float64) float64 {
	t := (jd - 2451545.0) / 36525.0
	return 23.4392911111 - (46.8150*t+0.00059*t*t-0.001813*t*t*t)/3600.0
}

func MeanSiderealTime(jd float64) float64 {
	t := (jd - 2451545.0) / 36525.0
	return Limit360(280.46061837 + 360.98564736629*(jd-2451545.0) + 0.000387933*t*t - t*t*t/38710000.0)
}

func EclipticToEquatorial(jd, lo, bo float64) (float64, float64) {
	eps := MeanObliquity(jd)
	ra := math.Atan2(Sin(lo)*Cos(eps)-Tan(bo)*Sin(eps), Cos(lo)) * 180.0 / math.Pi
	if ra < 0 {
		ra += 360
	}
	dec := ArcSin(Sin(bo)*Cos(eps) + Cos(bo)*Sin(eps)*Sin(lo))
	return ra, dec
}

func HorizontalCoordinates(ra, dec, jd, lon, lat float64) (float64, float64, float64) {
	lst := Limit360(MeanSiderealTime(jd) + lon)
	hourAngle := Limit360(lst - ra)
	altitude := ArcSin(clampUnit(Sin(lat)*Sin(dec) + Cos(lat)*Cos(dec)*Cos(hourAngle)))

	y := Sin(hourAngle)
	x := Cos(hourAngle)*Sin(lat) - Tan(dec)*Cos(lat)
	azimuth := math.Atan2(y, x) * 180.0 / math.Pi
	if azimuth < 0 {
		if hourAngle < 180 {
			azimuth += 360
		} else {
			azimuth += 180
		}
	} else if hourAngle < 180 {
		azimuth += 180
	}
	return altitude, Limit360(azimuth), hourAngle
}

func TopocentricRaDec(ra, dec, observerLat, observerLon, jd, distanceEarthRadii, heightMeters float64) (float64, float64) {
	u := math.Atan(0.99664719 * Tan(observerLat))
	rhoSin := 0.99664719*math.Sin(u) + heightMeters/6378140.0*Sin(observerLat)
	rhoCos := math.Cos(u) + heightMeters/6378140.0*Cos(observerLat)
	parallax := math.Asin(1.0 / distanceEarthRadii)

	hourAngle := (Limit360(MeanSiderealTime(jd) + observerLon - ra)) * math.Pi / 180.0
	decRad := dec * math.Pi / 180.0

	numerator := -rhoCos * math.Sin(parallax) * math.Sin(hourAngle)
	denominator := math.Cos(decRad) - rhoCos*math.Sin(parallax)*math.Cos(hourAngle)
	deltaRA := math.Atan2(numerator, denominator)

	topRA := Limit360(ra + deltaRA*180.0/math.Pi)
	topDec := math.Atan2((math.Sin(decRad)-rhoSin*math.Sin(parallax))*math.Cos(deltaRA), denominator) * 180.0 / math.Pi
	return topRA, topDec
}

func SearchRiseSet(startJD, targetAltitude, stepMinutes float64, isRise bool, altitudeFn func(float64) float64) (float64, error) {
	step := stepMinutes / 1440.0
	prevJD := startJD
	prevAlt := altitudeFn(prevJD) - targetAltitude
	minAlt := prevAlt
	maxAlt := prevAlt

	for i := 1; i <= int(math.Round(1440.0/stepMinutes)); i++ {
		currentJD := startJD + float64(i)*step
		currentAlt := altitudeFn(currentJD) - targetAltitude
		if currentAlt < minAlt {
			minAlt = currentAlt
		}
		if currentAlt > maxAlt {
			maxAlt = currentAlt
		}
		if crosses(prevAlt, currentAlt, isRise) {
			return bisectEvent(prevJD, currentJD, targetAltitude, altitudeFn), nil
		}
		prevJD = currentJD
		prevAlt = currentAlt
	}

	if maxAlt < 0 {
		return 0, ErrNeverRise
	}
	if minAlt > 0 {
		return 0, ErrNeverSet
	}
	return 0, ErrNotOnThisDate
}

func crosses(prevAlt, currentAlt float64, isRise bool) bool {
	if isRise {
		return prevAlt < 0 && currentAlt >= 0
	}
	return prevAlt > 0 && currentAlt <= 0
}

func bisectEvent(lo, hi, targetAltitude float64, altitudeFn func(float64) float64) float64 {
	loAlt := altitudeFn(lo) - targetAltitude
	for i := 0; i < 40; i++ {
		mid := (lo + hi) / 2.0
		midAlt := altitudeFn(mid) - targetAltitude
		if midAlt == 0 {
			return mid
		}
		if sameSign(loAlt, midAlt) {
			lo = mid
			loAlt = midAlt
		} else {
			hi = mid
		}
	}
	return (lo + hi) / 2.0
}

func sameSign(a, b float64) bool {
	return (a >= 0 && b >= 0) || (a <= 0 && b <= 0)
}

func clampUnit(v float64) float64 {
	if v > 1 {
		return 1
	}
	if v < -1 {
		return -1
	}
	return v
}
