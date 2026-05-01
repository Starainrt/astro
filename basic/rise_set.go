package basic

import (
	"errors"
	"math"

	. "github.com/starainrt/astro/tools"
)

var (
	ErrNeverRise     = errors.New("rise event does not occur on this date")
	ErrNeverSet      = errors.New("set event does not occur on this date")
	ErrNotOnThisDate = errors.New("rise/set event occurs on adjacent date")
)

func StandardAltitudeStar(aero bool, observerHeight, lat float64) float64 {
	targetAltitude := 0.0
	if aero {
		targetAltitude = -0.566667
	}
	return targetAltitude - HeightDegreeByLat(observerHeight, lat)
}

func StandardAltitudeSun(zenithShift, observerHeight, lat float64) float64 {
	targetAltitude := 0.0
	if zenithShift != 0 {
		targetAltitude = -0.8333
	}
	return targetAltitude - HeightDegreeByLat(observerHeight, lat)
}

func StandardAltitudePlanet(aeroCorrection, observerHeight, lat float64) float64 {
	targetAltitude := 0.0
	if aeroCorrection != 0 {
		targetAltitude = -0.566667
	}
	return targetAltitude - HeightDegreeByLat(observerHeight, lat)
}

func StandardAltitudeMoon(zenithShift, observerHeight, lat float64) float64 {
	targetAltitude := 0.0
	if zenithShift != 0 {
		targetAltitude = -0.83333
	}
	return targetAltitude - HeightDegreeByLat(observerHeight, lat)
}

type planetCulminationFunc func(float64, float64, float64) float64
type planetHeightFunc func(float64, float64, float64, float64) float64
type planetDeclinationFunc func(float64) float64

func planetRiseDown(jd, lon, lat, timezone, aeroCorrection, observerHeight float64, isRise bool, culmination planetCulminationFunc, height planetHeightFunc, declination planetDeclinationFunc) (float64, error) {
	jd = math.Floor(jd) + 0.5
	localTimezone := math.Round(lon / 15)
	targetAltitude := StandardAltitudePlanet(aeroCorrection, observerHeight, lat)

	culminationJD := culmination(jd, lon, localTimezone)
	if height(culminationJD, lon, lat, localTimezone) < targetAltitude {
		return 0, ErrNeverRise
	}
	if height(culminationJD-0.5, lon, lat, localTimezone) > targetAltitude {
		return 0, ErrNeverSet
	}

	dec := declination(TD2UT(culminationJD-localTimezone/24, true))
	cosHourAngle := (Sin(targetAltitude) - Sin(dec)*Sin(lat)) / (Cos(dec) * Cos(lat))

	var eventJD float64
	if math.Abs(cosHourAngle) <= 1 {
		hourOffset := ArcCos(cosHourAngle) / 15
		if isRise {
			eventJD = culminationJD - hourOffset/24 - 25.0/24.0/60.0
		} else {
			eventJD = culminationJD + hourOffset/24 - 25.0/24.0/60.0
		}
	} else {
		eventJD = culminationJD
		steps := 0
		for height(eventJD, lon, lat, localTimezone) > targetAltitude {
			steps++
			if isRise {
				eventJD -= 15.0 / 60.0 / 24.0
			} else {
				eventJD += 15.0 / 60.0 / 24.0
			}
			if steps > 48 {
				break
			}
		}
	}

	estimateJD := eventJD
	for {
		prevJD := estimateJD
		altitudeDelta := height(prevJD, lon, lat, localTimezone) - targetAltitude
		altitudeSlope := (height(prevJD+0.000005, lon, lat, localTimezone) - height(prevJD-0.000005, lon, lat, localTimezone)) / 0.00001
		estimateJD = prevJD - altitudeDelta/altitudeSlope
		if math.Abs(estimateJD-prevJD) <= 0.00001 {
			break
		}
	}
	return estimateJD - localTimezone/24 + timezone/24, nil
}
