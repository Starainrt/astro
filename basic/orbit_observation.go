package basic

import (
	"math"

	. "github.com/starainrt/astro/tools"
)

func orbitTopocentricObservation(jde, observerLon, observerLat, observerHeight, timezone float64, elements OrbitElements) (ra, dec, distance float64) {
	utcJde := jde - timezone/24.0
	return OrbitApparentTopocentricEquatorial(TD2UT(utcJde, true), observerLon, observerLat, observerHeight, elements)
}

// OrbitHeight 返回轨道目标在观测者所在地的视高度角，单位度。
func OrbitHeight(jde, observerLon, observerLat, timezone, observerHeight float64, elements OrbitElements) float64 {
	ra, dec, _ := orbitTopocentricObservation(jde, observerLon, observerLat, observerHeight, timezone, elements)
	st := Limit360(ApparentSiderealTime(jde-timezone/24.0)*15 + observerLon)
	hourAngle := Limit360(st - ra)
	sinHeight := Sin(observerLat)*Sin(dec) + Cos(dec)*Cos(observerLat)*Cos(hourAngle)
	return ArcSin(sinHeight)
}

// OrbitAzimuth 返回轨道目标在观测者所在地的视方位角，按正北为 0°、向东增加。
func OrbitAzimuth(jde, observerLon, observerLat, timezone, observerHeight float64, elements OrbitElements) float64 {
	ra, dec, _ := orbitTopocentricObservation(jde, observerLon, observerLat, observerHeight, timezone, elements)
	st := Limit360(ApparentSiderealTime(jde-timezone/24.0)*15 + observerLon)
	hourAngle := Limit360(st - ra)
	tanAzimuth := Sin(hourAngle) / (Cos(hourAngle)*Sin(observerLat) - Tan(dec)*Cos(observerLat))
	azimuth := ArcTan(tanAzimuth)
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

// OrbitHourAngle 返回轨道目标的站心视时角，单位度。
func OrbitHourAngle(jde, observerLon, observerLat, timezone, observerHeight float64, elements OrbitElements) float64 {
	ra, _, _ := orbitTopocentricObservation(jde, observerLon, observerLat, observerHeight, timezone, elements)
	st := Limit360(ApparentSiderealTime(jde-timezone/24.0)*15 + observerLon)
	hourAngle := st - ra
	if hourAngle < 0 {
		hourAngle += 360
	}
	return hourAngle
}

// OrbitCulminationTime 返回轨道目标的中天时刻，输入输出均沿用本仓库现有观测函数的 JD 语义。
func OrbitCulminationTime(jde, observerLon, observerLat, timezone, observerHeight float64, elements OrbitElements) float64 {
	jde = math.Floor(jde) + 0.5
	estimateJD := jde + Limit360(360-OrbitHourAngle(jde, observerLon, observerLat, timezone, observerHeight, elements))/15.0/24.0*0.99726851851851851851
	normalizedHourAngle := func(jde float64) float64 {
		currentHourAngle := OrbitHourAngle(jde, observerLon, observerLat, timezone, observerHeight, elements)
		if currentHourAngle < 180 {
			currentHourAngle += 360
		}
		return currentHourAngle
	}
	for {
		prevJD := estimateJD
		hourAngleDelta := normalizedHourAngle(prevJD) - 360
		hourAngleSlope := (normalizedHourAngle(prevJD+0.000005) - normalizedHourAngle(prevJD-0.000005)) / 0.00001
		estimateJD = prevJD - hourAngleDelta/hourAngleSlope
		if math.Abs(estimateJD-prevJD) <= 0.00001 {
			break
		}
	}
	return estimateJD
}

// OrbitRiseTime 返回轨道目标在给定当地日期的升起时刻。
func OrbitRiseTime(jde, observerLon, observerLat, timezone, aeroCorrection, observerHeight float64, elements OrbitElements) (float64, error) {
	return orbitRiseDown(jde, observerLon, observerLat, timezone, aeroCorrection, observerHeight, elements, true)
}

// OrbitSetTime 返回轨道目标在给定当地日期的落下时刻。
func OrbitSetTime(jde, observerLon, observerLat, timezone, aeroCorrection, observerHeight float64, elements OrbitElements) (float64, error) {
	return orbitRiseDown(jde, observerLon, observerLat, timezone, aeroCorrection, observerHeight, elements, false)
}

func orbitRiseDown(jde, observerLon, observerLat, timezone, aeroCorrection, observerHeight float64, elements OrbitElements, isRise bool) (float64, error) {
	localTimezone := math.Round(observerLon / 15)
	targetAltitude := StandardAltitudePlanet(aeroCorrection, observerHeight, observerLat)

	culminationJD := OrbitCulminationTime(jde, observerLon, observerLat, localTimezone, observerHeight, elements)
	if OrbitHeight(culminationJD, observerLon, observerLat, localTimezone, observerHeight, elements) < targetAltitude {
		return 0, ErrNeverRise
	}
	if OrbitHeight(culminationJD-0.5, observerLon, observerLat, localTimezone, observerHeight, elements) > targetAltitude {
		return 0, ErrNeverSet
	}

	_, dec, _ := orbitTopocentricObservation(culminationJD, observerLon, observerLat, observerHeight, localTimezone, elements)
	cosHourAngle := (Sin(targetAltitude) - Sin(dec)*Sin(observerLat)) / (Cos(dec) * Cos(observerLat))

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
		for OrbitHeight(eventJD, observerLon, observerLat, localTimezone, observerHeight, elements) > targetAltitude {
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
		altitudeDelta := OrbitHeight(prevJD, observerLon, observerLat, localTimezone, observerHeight, elements) - targetAltitude
		altitudeSlope := (OrbitHeight(prevJD+0.000005, observerLon, observerLat, localTimezone, observerHeight, elements) - OrbitHeight(prevJD-0.000005, observerLon, observerLat, localTimezone, observerHeight, elements)) / 0.00001
		estimateJD = prevJD - altitudeDelta/altitudeSlope
		if math.Abs(estimateJD-prevJD) <= 0.00001 {
			break
		}
	}
	return estimateJD - localTimezone/24 + timezone/24, nil
}
