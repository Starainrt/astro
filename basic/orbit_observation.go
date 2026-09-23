package basic

import (
	"math"

	. "github.com/starainrt/astro/tools"
)

func orbitTopocentricObservation(localJD, observerLon, observerLat, observerHeight, timezone float64, elements OrbitElements) (ra, dec, distance float64) {
	utcJD := localJD - timezone/24.0
	return OrbitApparentTopocentricEquatorial(UTC2TT(utcJD), observerLon, observerLat, observerHeight, elements)
}

// OrbitHeight 返回轨道目标在观测者所在地的视高度角，单位度。
func OrbitHeight(localJD, observerLon, observerLat, timezone, observerHeight float64, elements OrbitElements) float64 {
	ra, dec, _ := orbitTopocentricObservation(localJD, observerLon, observerLat, observerHeight, timezone, elements)
	st := Limit360(ApparentSiderealTime(UTC2UT1(localJD-timezone/24.0))*15 + observerLon)
	hourAngle := Limit360(st - ra)
	sinHeight := Sin(observerLat)*Sin(dec) + Cos(dec)*Cos(observerLat)*Cos(hourAngle)
	return ArcSin(sinHeight)
}

// OrbitAzimuth 返回轨道目标在观测者所在地的视方位角，按正北为 0°、向东增加。
func OrbitAzimuth(localJD, observerLon, observerLat, timezone, observerHeight float64, elements OrbitElements) float64 {
	ra, dec, _ := orbitTopocentricObservation(localJD, observerLon, observerLat, observerHeight, timezone, elements)
	st := Limit360(ApparentSiderealTime(UTC2UT1(localJD-timezone/24.0))*15 + observerLon)
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
func OrbitHourAngle(localJD, observerLon, observerLat, timezone, observerHeight float64, elements OrbitElements) float64 {
	ra, _, _ := orbitTopocentricObservation(localJD, observerLon, observerLat, observerHeight, timezone, elements)
	st := Limit360(ApparentSiderealTime(UTC2UT1(localJD-timezone/24.0))*15 + observerLon)
	hourAngle := st - ra
	if hourAngle < 0 {
		hourAngle += 360
	}
	return hourAngle
}

// OrbitHourAngleWithTopocentric 返回站心视时角及同一状态的站心视赤经、视赤纬 / hour angle with its topocentric state.
//
// localJD 沿用 OrbitHourAngle 的当地时口径：函数内部先减 timezone/24 得世界时，再按 UT→TT 换算。
// localJD follows the local-time convention of OrbitHourAngle: timezone/24 is subtracted before the UT→TT conversion.
func OrbitHourAngleWithTopocentric(localJD, observerLon, observerLat, timezone, observerHeight float64, elements OrbitElements) (ra, dec, hourAngle float64) {
	ra, dec, _ = orbitTopocentricObservation(localJD, observerLon, observerLat, observerHeight, timezone, elements)
	st := Limit360(ApparentSiderealTime(UTC2UT1(localJD-timezone/24.0))*15 + observerLon)
	hourAngle = st - ra
	if hourAngle < 0 {
		hourAngle += 360
	}
	return ra, dec, hourAngle
}

// OrbitCulminationTime 返回轨道目标的中天时刻，输入输出均沿用本仓库现有观测函数的 JD 语义。
func OrbitCulminationTime(localJD, observerLon, observerLat, timezone, observerHeight float64, elements OrbitElements) float64 {
	if !isFiniteFloat(localJD) || !isFiniteFloat(observerLon) || !isFiniteFloat(observerLat) || !isFiniteFloat(timezone) || !isFiniteFloat(observerHeight) {
		return math.NaN()
	}
	localJD = math.Floor(localJD) + 0.5
	estimateJD := localJD + Limit360(360-OrbitHourAngle(localJD, observerLon, observerLat, timezone, observerHeight, elements))/15.0/24.0*0.99726851851851851851
	normalizedHourAngle := func(localJD float64) float64 {
		currentHourAngle := OrbitHourAngle(localJD, observerLon, observerLat, timezone, observerHeight, elements)
		if currentHourAngle < 180 {
			currentHourAngle += 360
		}
		return currentHourAngle
	}
	var ok bool
	estimateJD, ok = eventNewtonRefine(estimateJD, 0.00001, func(prevJD float64) float64 {
		hourAngleDelta := normalizedHourAngle(prevJD) - 360
		hourAngleSlope := (normalizedHourAngle(prevJD+0.000005) - normalizedHourAngle(prevJD-0.000005)) / 0.00001
		return hourAngleDelta / hourAngleSlope
	})
	if !ok {
		return math.NaN()
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
	if !isFiniteFloat(jde) || !isFiniteFloat(observerLon) || !isFiniteFloat(observerLat) || !isFiniteFloat(timezone) || !isFiniteFloat(aeroCorrection) || !isFiniteFloat(observerHeight) {
		return 0, ErrInvalidObservationInput
	}
	localTimezone := math.Round(observerLon / 15)
	targetAltitude := StandardAltitudePlanet(aeroCorrection, observerHeight, observerLat)

	culminationJD := OrbitCulminationTime(jde, observerLon, observerLat, localTimezone, observerHeight, elements)
	if !isFiniteFloat(culminationJD) {
		return 0, ErrInvalidObservationInput
	}
	culminationHeight := OrbitHeight(culminationJD, observerLon, observerLat, localTimezone, observerHeight, elements)
	previousHeight := OrbitHeight(culminationJD-0.5, observerLon, observerLat, localTimezone, observerHeight, elements)
	if !isFiniteFloat(culminationHeight) || !isFiniteFloat(previousHeight) {
		return 0, ErrInvalidObservationInput
	}
	if culminationHeight < targetAltitude {
		return 0, ErrNeverRise
	}
	if previousHeight > targetAltitude {
		return 0, ErrNeverSet
	}

	_, dec, _ := orbitTopocentricObservation(culminationJD, observerLon, observerLat, observerHeight, localTimezone, elements)
	cosHourAngle := (Sin(targetAltitude) - Sin(dec)*Sin(observerLat)) / (Cos(dec) * Cos(observerLat))
	if !isFiniteFloat(dec) || !isFiniteFloat(cosHourAngle) {
		return 0, ErrInvalidObservationInput
	}

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

	estimateJD, ok := eventNewtonRefine(eventJD, 0.00001, func(prevJD float64) float64 {
		altitudeDelta := OrbitHeight(prevJD, observerLon, observerLat, localTimezone, observerHeight, elements) - targetAltitude
		altitudeSlope := (OrbitHeight(prevJD+0.000005, observerLon, observerLat, localTimezone, observerHeight, elements) - OrbitHeight(prevJD-0.000005, observerLon, observerLat, localTimezone, observerHeight, elements)) / 0.00001
		return altitudeDelta / altitudeSlope
	})
	if !ok {
		return 0, ErrInvalidObservationInput
	}
	return estimateJD - localTimezone/24 + timezone/24, nil
}
