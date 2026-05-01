package internal

import (
	"math"

	. "github.com/starainrt/astro/tools"
)

type MoonState struct {
	Longitude          float64
	Latitude           float64
	DistanceEarthRadii float64
	RightAscension     float64
	Declination        float64
}

func MoonGeocentric(jd float64) MoonState {
	d := jd - 2451543.5

	node := Limit360(125.1228 - 0.0529538083*d)
	inclination := 5.1454
	perigee := Limit360(318.0634 + 0.1643573223*d)
	semiMajorAxis := 60.2666
	eccentricity := 0.054900
	meanAnomaly := Limit360(115.3654 + 13.0649929509*d)
	meanLongitude := Limit360(node + perigee + meanAnomaly)
	argumentLatitude := Limit360(meanLongitude - node)

	sunPerigee := Limit360(282.9404 + 0.0000470935*d)
	sunEccentricity := 0.016709 - 0.000000001151*d
	sunMeanAnomaly := Limit360(356.0470 + 0.9856002585*d)
	sunTrueLongitude, _ := orbitalLongitudeDistance(sunPerigee, sunEccentricity, 1, sunMeanAnomaly)

	longitude, latitude, distance := orbitalLongitudeLatitudeDistance(node, inclination, perigee, semiMajorAxis, eccentricity, meanAnomaly)
	elongation := Limit360(meanLongitude - sunTrueLongitude)

	longitude += -1.274 * Sin(meanAnomaly-2*elongation)
	longitude += 0.658 * Sin(2*elongation)
	longitude += -0.186 * Sin(sunMeanAnomaly)
	longitude += -0.059 * Sin(2*meanAnomaly-2*elongation)
	longitude += -0.057 * Sin(meanAnomaly-2*elongation+sunMeanAnomaly)
	longitude += 0.053 * Sin(meanAnomaly+2*elongation)
	longitude += 0.046 * Sin(2*elongation-sunMeanAnomaly)
	longitude += 0.041 * Sin(meanAnomaly-sunMeanAnomaly)
	longitude += -0.035 * Sin(elongation)
	longitude += -0.031 * Sin(meanAnomaly+sunMeanAnomaly)
	longitude += -0.015 * Sin(2*argumentLatitude-2*elongation)
	longitude += 0.011 * Sin(meanAnomaly-4*elongation)

	latitude += -0.173 * Sin(argumentLatitude-2*elongation)
	latitude += -0.055 * Sin(meanAnomaly-argumentLatitude-2*elongation)
	latitude += -0.046 * Sin(meanAnomaly+argumentLatitude-2*elongation)
	latitude += 0.033 * Sin(argumentLatitude+2*elongation)
	latitude += 0.017 * Sin(2*meanAnomaly+argumentLatitude)

	distance += -0.58 * Cos(meanAnomaly-2*elongation)
	distance += -0.46 * Cos(2*elongation)

	longitude = Limit360(longitude)
	ra, dec := EclipticToEquatorial(jd, longitude, latitude)
	return MoonState{
		Longitude:          longitude,
		Latitude:           latitude,
		DistanceEarthRadii: distance,
		RightAscension:     ra,
		Declination:        dec,
	}
}

func MoonTopocentric(jd, observerLon, observerLat, heightMeters float64) MoonState {
	geo := MoonGeocentric(jd)
	ra, dec := TopocentricRaDec(geo.RightAscension, geoDeclinationClamp(geo.Declination), observerLat, observerLon, jd, geo.DistanceEarthRadii, heightMeters)
	geo.RightAscension = ra
	geo.Declination = dec
	return geo
}

func orbitalLongitudeLatitudeDistance(node, inclination, perigee, axis, eccentricity, meanAnomaly float64) (float64, float64, float64) {
	meanAnomalyRad := meanAnomaly * math.Pi / 180.0
	eccentricAnomaly := meanAnomalyRad + eccentricity*math.Sin(meanAnomalyRad)*(1+eccentricity*math.Cos(meanAnomalyRad))
	for i := 0; i < 5; i++ {
		eccentricAnomaly -= (eccentricAnomaly - eccentricity*math.Sin(eccentricAnomaly) - meanAnomalyRad) / (1 - eccentricity*math.Cos(eccentricAnomaly))
	}

	xv := axis * (math.Cos(eccentricAnomaly) - eccentricity)
	yv := axis * math.Sqrt(1-eccentricity*eccentricity) * math.Sin(eccentricAnomaly)
	trueAnomaly := math.Atan2(yv, xv)
	radius := math.Hypot(xv, yv)

	nodeRad := node * math.Pi / 180.0
	inclinationRad := inclination * math.Pi / 180.0
	perigeeRad := perigee * math.Pi / 180.0

	xh := radius * (math.Cos(nodeRad)*math.Cos(trueAnomaly+perigeeRad) - math.Sin(nodeRad)*math.Sin(trueAnomaly+perigeeRad)*math.Cos(inclinationRad))
	yh := radius * (math.Sin(nodeRad)*math.Cos(trueAnomaly+perigeeRad) + math.Cos(nodeRad)*math.Sin(trueAnomaly+perigeeRad)*math.Cos(inclinationRad))
	zh := radius * math.Sin(trueAnomaly+perigeeRad) * math.Sin(inclinationRad)

	longitude := math.Atan2(yh, xh) * 180.0 / math.Pi
	latitude := math.Atan2(zh, math.Hypot(xh, yh)) * 180.0 / math.Pi
	return Limit360(longitude), latitude, radius
}

func orbitalLongitudeDistance(perigee, eccentricity, axis, meanAnomaly float64) (float64, float64) {
	meanAnomalyRad := meanAnomaly * math.Pi / 180.0
	eccentricAnomaly := meanAnomalyRad + eccentricity*math.Sin(meanAnomalyRad)*(1+eccentricity*math.Cos(meanAnomalyRad))
	for i := 0; i < 5; i++ {
		eccentricAnomaly -= (eccentricAnomaly - eccentricity*math.Sin(eccentricAnomaly) - meanAnomalyRad) / (1 - eccentricity*math.Cos(eccentricAnomaly))
	}

	xv := axis * (math.Cos(eccentricAnomaly) - eccentricity)
	yv := axis * math.Sqrt(1-eccentricity*eccentricity) * math.Sin(eccentricAnomaly)
	trueAnomaly := math.Atan2(yv, xv) * 180.0 / math.Pi
	radius := math.Hypot(xv, yv)
	return Limit360(trueAnomaly + perigee), radius
}

func geoDeclinationClamp(dec float64) float64 {
	if dec > 90 {
		return 90
	}
	if dec < -90 {
		return -90
	}
	return dec
}
