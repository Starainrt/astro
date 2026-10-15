package basic

import (
	"math"

	. "github.com/starainrt/astro/tools"
)

// Exported N variants below keep the same jd semantics as the non-N APIs; n < 0 means full series.
type planetDeclinationFuncN func(float64, int) float64

func planetXYZN(planetIndex int, jde float64, n int, instances ...*Instance) (float64, float64, float64) {
	inst := instanceValue(instances)

	l := PlanetCoordinateNWithInstance(planetIndex, 0, jde, n, inst)
	b := PlanetCoordinateNWithInstance(planetIndex, 1, jde, n, inst)
	r := PlanetCoordinateNWithInstance(planetIndex, 2, jde, n, inst)
	el := PlanetCoordinateNWithInstance(-1, 0, jde, n, inst)
	eb := PlanetCoordinateNWithInstance(-1, 1, jde, n, inst)
	er := PlanetCoordinateNWithInstance(-1, 2, jde, n, inst)
	x := r*Cos(b)*Cos(l) - er*Cos(eb)*Cos(el)
	y := r*Cos(b)*Sin(l) - er*Cos(eb)*Sin(el)
	z := r*Sin(b) - er*Sin(eb)
	return x, y, z
}

func planetApparentLoBoN(planetIndex int, jd float64, n int) (float64, float64) {
	geo, _ := planetApparentGeocentricPositionN(planetIndex, jd, n)
	return geo.lo, geo.bo
}

func planetApparentRaManualN(planetIndex int, jde float64, n int) float64 {
	lo, bo := planetApparentLoBoN(planetIndex, jde, n)
	eps := TrueObliquity(jde)
	ra := math.Atan2((Sin(lo)*Cos(eps) - Tan(bo)*Sin(eps)), Cos(lo))
	return Limit360(ra * 180 / math.Pi)
}

func planetApparentDecManualN(planetIndex int, jde float64, n int) float64 {
	lo, bo := planetApparentLoBoN(planetIndex, jde, n)
	eps := TrueObliquity(jde)
	return ArcSin(Sin(bo)*Cos(eps) + Cos(bo)*Sin(eps)*Sin(lo))
}

func planetApparentRaDecManualN(planetIndex int, jde float64, n int) (float64, float64) {
	lo, bo := planetApparentLoBoN(planetIndex, jde, n)
	return planetApparentRaDecFromLoBo(jde, lo, bo)
}

func planetApparentRaDecFromLoBo(jde, lo, bo float64) (float64, float64) {
	eps := TrueObliquity(jde)
	ra := math.Atan2((Sin(lo)*Cos(eps) - Tan(bo)*Sin(eps)), Cos(lo))
	ra = ra * 180 / math.Pi
	dec := ArcSin(Sin(bo)*Cos(eps) + Cos(bo)*Sin(eps)*Sin(lo))
	return Limit360(ra), dec
}

func planetEarthAwayN(planetIndex int, jd float64, n int) float64 {
	return planetEarthAwayExplicitN(planetIndex, jd, n)
}

func planetHourAngleN(jd, lon, timezone float64, n int, apparentRa func(float64, int) float64) float64 {
	siderealLongitude := Limit360(ApparentSiderealTime(UTC2UT1(jd-timezone/24))*15 + lon)
	hourAngle := siderealLongitude - apparentRa(UTC2TT(jd-timezone/24.0), n)
	if hourAngle < 0 {
		hourAngle += 360
	}
	return hourAngle
}

func planetCulminationTimeN(localJD, lon, timezone float64, n int, hourAngle func(float64, float64, float64, int) float64) float64 {
	if !isFiniteFloat(localJD) || !isFiniteFloat(lon) || !isFiniteFloat(timezone) {
		return math.NaN()
	}
	localJD = math.Floor(localJD) + 0.5
	estimateJD := localJD + Limit360(360-hourAngle(localJD, lon, timezone, n))/15.0/24.0*0.99726851851851851851
	normalizedHourAngle := func(localJD, lon, timezone float64) float64 {
		currentHourAngle := hourAngle(localJD, lon, timezone, n)
		if currentHourAngle < 180 {
			currentHourAngle += 360
		}
		return currentHourAngle
	}
	var ok bool
	estimateJD, ok = eventNewtonRefine(estimateJD, 0.00001, func(prevJD float64) float64 {
		hourAngleDelta := normalizedHourAngle(prevJD, lon, timezone) - 360
		hourAngleSlope := (normalizedHourAngle(prevJD+0.000005, lon, timezone) - normalizedHourAngle(prevJD-0.000005, lon, timezone)) / 0.00001
		return hourAngleDelta / hourAngleSlope
	})
	if !ok {
		return math.NaN()
	}
	return estimateJD
}

func planetRiseDownN(jd, lon, lat, timezone, aeroCorrection, observerHeight float64, isRise bool, n int, culmination func(float64, float64, float64, int) float64, height func(float64, float64, float64, float64, int) float64, declination planetDeclinationFuncN) (float64, error) {
	if !isFiniteFloat(jd) || !isFiniteFloat(lon) || !isFiniteFloat(lat) || !isFiniteFloat(timezone) || !isFiniteFloat(aeroCorrection) || !isFiniteFloat(observerHeight) {
		return 0, ErrInvalidObservationInput
	}
	jd = math.Floor(jd) + 0.5
	localTimezone := math.Round(lon / 15)
	targetAltitude := StandardAltitudePlanet(aeroCorrection, observerHeight, lat)
	culminationJD := culmination(jd, lon, localTimezone, n)
	if !isFiniteFloat(culminationJD) {
		return 0, ErrInvalidObservationInput
	}
	culminationHeight := height(culminationJD, lon, lat, localTimezone, n)
	previousHeight := height(culminationJD-0.5, lon, lat, localTimezone, n)
	if !isFiniteFloat(culminationHeight) || !isFiniteFloat(previousHeight) {
		return 0, ErrInvalidObservationInput
	}
	if culminationHeight < targetAltitude {
		return 0, ErrNeverRise
	}
	if previousHeight > targetAltitude {
		return 0, ErrNeverSet
	}
	dec := declination(UTC2TT(culminationJD-localTimezone/24), n)
	cosHourAngle := (Sin(targetAltitude) - Sin(dec)*Sin(lat)) / (Cos(dec) * Cos(lat))
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
		for height(eventJD, lon, lat, localTimezone, n) > targetAltitude {
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
		altitudeDelta := height(prevJD, lon, lat, localTimezone, n) - targetAltitude
		altitudeSlope := (height(prevJD+0.000005, lon, lat, localTimezone, n) - height(prevJD-0.000005, lon, lat, localTimezone, n)) / 0.00001
		return altitudeDelta / altitudeSlope
	})
	if !ok {
		return 0, ErrInvalidObservationInput
	}
	return estimateJD - localTimezone/24 + timezone/24, nil
}

// MercuryApparentLoN 水星视黄经（n 变体） / n-parameterised apparent ecliptic longitude of Mercury.
func MercuryApparentLoN(jd float64, n int) float64 {
	lo, _ := planetApparentLoBoN(1, jd, n)
	return lo
}

// MercuryApparentBoN 水星视黄纬（n 变体） / n-parameterised apparent ecliptic latitude of Mercury.
func MercuryApparentBoN(jd float64, n int) float64 {
	_, bo := planetApparentLoBoN(1, jd, n)
	return bo
}

// MercuryApparentLoBoN 水星视黄经黄纬（n 变体） / n-parameterised apparent ecliptic longitude and latitude of Mercury.
func MercuryApparentLoBoN(jd float64, n int) (float64, float64) {
	return planetApparentLoBoN(1, jd, n)
}

// MercuryApparentRaN 水星视赤经（n 变体） / n-parameterised apparent right ascension of Mercury.
func MercuryApparentRaN(jde float64, n int) float64 {
	lo, bo := MercuryApparentLoBoN(jde, n)
	return LoToRa(jde, lo, bo)
}

// MercuryApparentDecN 水星视赤纬（n 变体） / n-parameterised apparent declination of Mercury.
func MercuryApparentDecN(jde float64, n int) float64 {
	lo, bo := MercuryApparentLoBoN(jde, n)
	eps := TrueObliquity(jde)
	return ArcSin(Sin(bo)*Cos(eps) + Cos(bo)*Sin(eps)*Sin(lo))
}

// MercuryApparentRaDecN 水星视赤经赤纬（n 变体） / n-parameterised apparent right ascension and declination of Mercury.
func MercuryApparentRaDecN(jde float64, n int) (float64, float64) {
	lo, bo := MercuryApparentLoBoN(jde, n)
	return LoBoToRaDec(jde, lo, bo)
}

// EarthMercuryAwayN 地水距离（n 变体） / n-parameterised Earth-Mercury distance.
func EarthMercuryAwayN(jd float64, n int) float64 {
	return planetEarthAwayN(1, jd, n)
}

// MercuryMagNWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func MercuryMagNWithInstance(jde float64, n int, inst *Instance) float64 {
	awaySun := PlanetCoordinateNWithInstance(1, 2, jde, n, inst)
	awayEarth := EarthMercuryAwayNWithInstance(jde, n, inst)
	away := PlanetCoordinateNWithInstance(-1, 2, jde, n, inst)
	i := (awaySun*awaySun + awayEarth*awayEarth - away*away) / (2 * awaySun * awayEarth)
	i = ArcCos(i)
	mag := -0.42 + 5*math.Log10(awaySun*awayEarth) + 0.0380*i - 0.000273*i*i + 0.000002*i*i*i
	return FloatRound(mag, 2)
}

// MercuryHeightNWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func MercuryHeightNWithInstance(localJD, lon, lat, timezone float64, n int, inst *Instance) float64 {
	return planetHeightN(1, localJD, lon, lat, timezone, n, inst)
}

// MercuryAzimuthNWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func MercuryAzimuthNWithInstance(localJD, lon, lat, timezone float64, n int, inst *Instance) float64 {
	return planetAzimuthN(1, localJD, lon, lat, timezone, n, inst)
}

// MercuryHourAngleNWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func MercuryHourAngleNWithInstance(jd, lon, timezone float64, n int, inst *Instance) float64 {
	return planetHourAngleN(jd, lon, timezone, n, func(jde float64, n int) float64 { return MercuryApparentRaNWithInstance(jde, n, inst) })
}

// MercuryCulminationTimeNWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func MercuryCulminationTimeNWithInstance(localJD, lon, timezone float64, n int, inst *Instance) float64 {
	return planetCulminationTimeN(localJD, lon, timezone, n, func(jd, lon, timezone float64, n int) float64 {
		return MercuryHourAngleNWithInstance(jd, lon, timezone, n, inst)
	})
}

// MercuryRiseTimeNWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func MercuryRiseTimeNWithInstance(jd, lon, lat, timezone, aeroCorrection, observerHeight float64, n int, inst *Instance) (float64, error) {
	return planetRiseDownN(jd, lon, lat, timezone, aeroCorrection, observerHeight, true, n, func(localJD, lon, timezone float64, n int) float64 {
		return MercuryCulminationTimeNWithInstance(localJD, lon, timezone, n, inst)
	}, func(localJD, lon, lat, timezone float64, n int) float64 {
		return MercuryHeightNWithInstance(localJD, lon, lat, timezone, n, inst)
	}, func(jde float64, n int) float64 { return MercuryApparentDecNWithInstance(jde, n, inst) })
}

// MercurySetTimeNWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func MercurySetTimeNWithInstance(jd, lon, lat, timezone, aeroCorrection, observerHeight float64, n int, inst *Instance) (float64, error) {
	return planetRiseDownN(jd, lon, lat, timezone, aeroCorrection, observerHeight, false, n, func(localJD, lon, timezone float64, n int) float64 {
		return MercuryCulminationTimeNWithInstance(localJD, lon, timezone, n, inst)
	}, func(localJD, lon, lat, timezone float64, n int) float64 {
		return MercuryHeightNWithInstance(localJD, lon, lat, timezone, n, inst)
	}, func(jde float64, n int) float64 { return MercuryApparentDecNWithInstance(jde, n, inst) })
}

// VenusApparentLoN 金星视黄经（n 变体） / n-parameterised apparent ecliptic longitude of Venus.
func VenusApparentLoN(jd float64, n int) float64 {
	lo, _ := planetApparentLoBoN(2, jd, n)
	return lo
}

// VenusApparentBoN 金星视黄纬（n 变体） / n-parameterised apparent ecliptic latitude of Venus.
func VenusApparentBoN(jd float64, n int) float64 {
	_, bo := planetApparentLoBoN(2, jd, n)
	return bo
}

// VenusApparentLoBoN 金星视黄经黄纬（n 变体） / n-parameterised apparent ecliptic longitude and latitude of Venus.
func VenusApparentLoBoN(jd float64, n int) (float64, float64) {
	return planetApparentLoBoN(2, jd, n)
}

// VenusApparentRaN 金星视赤经（n 变体） / n-parameterised apparent right ascension of Venus.
func VenusApparentRaN(jd float64, n int) float64 {
	return planetApparentRaManualN(2, jd, n)
}

// VenusApparentDecN 金星视赤纬（n 变体） / n-parameterised apparent declination of Venus.
func VenusApparentDecN(jd float64, n int) float64 {
	return planetApparentDecManualN(2, jd, n)
}

// VenusApparentRaDecN 金星视赤经赤纬（n 变体） / n-parameterised apparent right ascension and declination of Venus.
func VenusApparentRaDecN(jd float64, n int) (float64, float64) {
	return planetApparentRaDecManualN(2, jd, n)
}

// EarthVenusAwayN 地金距离（n 变体） / n-parameterised Earth-Venus distance.
func EarthVenusAwayN(jd float64, n int) float64 {
	return planetEarthAwayN(2, jd, n)
}

// VenusMagNWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func VenusMagNWithInstance(jde float64, n int, inst *Instance) float64 {
	awaySun := PlanetCoordinateNWithInstance(2, 2, jde, n, inst)
	awayEarth := EarthVenusAwayNWithInstance(jde, n, inst)
	away := PlanetCoordinateNWithInstance(-1, 2, jde, n, inst)
	i := (awaySun*awaySun + awayEarth*awayEarth - away*away) / (2 * awaySun * awayEarth)
	i = ArcCos(i)
	mag := -4.40 + 5*math.Log10(awaySun*awayEarth) + 0.0009*i + 0.000239*i*i - 0.00000065*i*i*i
	return FloatRound(mag, 2)
}

// VenusHeightNWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func VenusHeightNWithInstance(localJD, lon, lat, timezone float64, n int, inst *Instance) float64 {
	return planetHeightN(2, localJD, lon, lat, timezone, n, inst)
}

// VenusAzimuthNWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func VenusAzimuthNWithInstance(localJD, lon, lat, timezone float64, n int, inst *Instance) float64 {
	return planetAzimuthN(2, localJD, lon, lat, timezone, n, inst)
}

// VenusHourAngleNWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func VenusHourAngleNWithInstance(jd, lon, timezone float64, n int, inst *Instance) float64 {
	return planetHourAngleN(jd, lon, timezone, n, func(jd float64, n int) float64 { return VenusApparentRaNWithInstance(jd, n, inst) })
}

// VenusCulminationTimeNWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func VenusCulminationTimeNWithInstance(localJD, lon, timezone float64, n int, inst *Instance) float64 {
	return planetCulminationTimeN(localJD, lon, timezone, n, func(jd, lon, timezone float64, n int) float64 {
		return VenusHourAngleNWithInstance(jd, lon, timezone, n, inst)
	})
}

// VenusRiseTimeNWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func VenusRiseTimeNWithInstance(jd, lon, lat, timezone, aeroCorrection, observerHeight float64, n int, inst *Instance) (float64, error) {
	return planetRiseDownN(jd, lon, lat, timezone, aeroCorrection, observerHeight, true, n, func(localJD, lon, timezone float64, n int) float64 {
		return VenusCulminationTimeNWithInstance(localJD, lon, timezone, n, inst)
	}, func(localJD, lon, lat, timezone float64, n int) float64 {
		return VenusHeightNWithInstance(localJD, lon, lat, timezone, n, inst)
	}, func(jd float64, n int) float64 { return VenusApparentDecNWithInstance(jd, n, inst) })
}

// VenusSetTimeNWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func VenusSetTimeNWithInstance(jd, lon, lat, timezone, aeroCorrection, observerHeight float64, n int, inst *Instance) (float64, error) {
	return planetRiseDownN(jd, lon, lat, timezone, aeroCorrection, observerHeight, false, n, func(localJD, lon, timezone float64, n int) float64 {
		return VenusCulminationTimeNWithInstance(localJD, lon, timezone, n, inst)
	}, func(localJD, lon, lat, timezone float64, n int) float64 {
		return VenusHeightNWithInstance(localJD, lon, lat, timezone, n, inst)
	}, func(jd float64, n int) float64 { return VenusApparentDecNWithInstance(jd, n, inst) })
}

// MarsApparentLoN 火星视黄经（n 变体） / n-parameterised apparent ecliptic longitude of Mars.
func MarsApparentLoN(jd float64, n int) float64 {
	lo, _ := planetApparentLoBoN(3, jd, n)
	return lo
}

// MarsApparentBoN 火星视黄纬（n 变体） / n-parameterised apparent ecliptic latitude of Mars.
func MarsApparentBoN(jd float64, n int) float64 {
	_, bo := planetApparentLoBoN(3, jd, n)
	return bo
}

// MarsApparentLoBoN 火星视黄经黄纬（n 变体） / n-parameterised apparent ecliptic longitude and latitude of Mars.
func MarsApparentLoBoN(jd float64, n int) (float64, float64) {
	return planetApparentLoBoN(3, jd, n)
}

// MarsApparentRaN 火星视赤经（n 变体） / n-parameterised apparent right ascension of Mars.
func MarsApparentRaN(jd float64, n int) float64 {
	return planetApparentRaManualN(3, jd, n)
}

// MarsApparentDecN 火星视赤纬（n 变体） / n-parameterised apparent declination of Mars.
func MarsApparentDecN(jd float64, n int) float64 {
	return planetApparentDecManualN(3, jd, n)
}

// MarsApparentRaDecN 火星视赤经赤纬（n 变体） / n-parameterised apparent right ascension and declination of Mars.
func MarsApparentRaDecN(jd float64, n int) (float64, float64) {
	return planetApparentRaDecManualN(3, jd, n)
}

// EarthMarsAwayN 地火距离（n 变体） / n-parameterised Earth-Mars distance.
func EarthMarsAwayN(jd float64, n int) float64 {
	return planetEarthAwayN(3, jd, n)
}

// MarsMagNWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func MarsMagNWithInstance(jde float64, n int, inst *Instance) float64 {
	awaySun := PlanetCoordinateNWithInstance(3, 2, jde, n, inst)
	awayEarth := EarthMarsAwayNWithInstance(jde, n, inst)
	away := PlanetCoordinateNWithInstance(-1, 2, jde, n, inst)
	i := (awaySun*awaySun + awayEarth*awayEarth - away*away) / (2 * awaySun * awayEarth)
	i = ArcCos(i)
	mag := -1.52 + 5*math.Log10(awaySun*awayEarth) + 0.016*i
	return FloatRound(mag, 2)
}

// MarsHeightNWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func MarsHeightNWithInstance(localJD, lon, lat, timezone float64, n int, inst *Instance) float64 {
	return planetHeightN(3, localJD, lon, lat, timezone, n, inst)
}

// MarsAzimuthNWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func MarsAzimuthNWithInstance(localJD, lon, lat, timezone float64, n int, inst *Instance) float64 {
	return planetAzimuthN(3, localJD, lon, lat, timezone, n, inst)
}

// MarsHourAngleNWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func MarsHourAngleNWithInstance(jd, lon, timezone float64, n int, inst *Instance) float64 {
	return planetHourAngleN(jd, lon, timezone, n, func(jd float64, n int) float64 { return MarsApparentRaNWithInstance(jd, n, inst) })
}

// MarsCulminationTimeNWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func MarsCulminationTimeNWithInstance(localJD, lon, timezone float64, n int, inst *Instance) float64 {
	return planetCulminationTimeN(localJD, lon, timezone, n, func(jd, lon, timezone float64, n int) float64 {
		return MarsHourAngleNWithInstance(jd, lon, timezone, n, inst)
	})
}

// MarsRiseTimeNWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func MarsRiseTimeNWithInstance(jd, lon, lat, timezone, aeroCorrection, observerHeight float64, n int, inst *Instance) (float64, error) {
	return planetRiseDownN(jd, lon, lat, timezone, aeroCorrection, observerHeight, true, n, func(localJD, lon, timezone float64, n int) float64 {
		return MarsCulminationTimeNWithInstance(localJD, lon, timezone, n, inst)
	}, func(localJD, lon, lat, timezone float64, n int) float64 {
		return MarsHeightNWithInstance(localJD, lon, lat, timezone, n, inst)
	}, func(jd float64, n int) float64 { return MarsApparentDecNWithInstance(jd, n, inst) })
}

// MarsSetTimeNWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func MarsSetTimeNWithInstance(jd, lon, lat, timezone, aeroCorrection, observerHeight float64, n int, inst *Instance) (float64, error) {
	return planetRiseDownN(jd, lon, lat, timezone, aeroCorrection, observerHeight, false, n, func(localJD, lon, timezone float64, n int) float64 {
		return MarsCulminationTimeNWithInstance(localJD, lon, timezone, n, inst)
	}, func(localJD, lon, lat, timezone float64, n int) float64 {
		return MarsHeightNWithInstance(localJD, lon, lat, timezone, n, inst)
	}, func(jd float64, n int) float64 { return MarsApparentDecNWithInstance(jd, n, inst) })
}

// JupiterApparentLoN 木星视黄经（n 变体） / n-parameterised apparent ecliptic longitude of Jupiter.
func JupiterApparentLoN(jd float64, n int) float64 {
	lo, _ := planetApparentLoBoN(4, jd, n)
	return lo
}

// JupiterApparentBoN 木星视黄纬（n 变体） / n-parameterised apparent ecliptic latitude of Jupiter.
func JupiterApparentBoN(jd float64, n int) float64 {
	_, bo := planetApparentLoBoN(4, jd, n)
	return bo
}

// JupiterApparentLoBoN 木星视黄经黄纬（n 变体） / n-parameterised apparent ecliptic longitude and latitude of Jupiter.
func JupiterApparentLoBoN(jd float64, n int) (float64, float64) {
	return planetApparentLoBoN(4, jd, n)
}

// JupiterApparentRaN 木星视赤经（n 变体） / n-parameterised apparent right ascension of Jupiter.
func JupiterApparentRaN(jd float64, n int) float64 {
	return planetApparentRaManualN(4, jd, n)
}

// JupiterApparentDecN 木星视赤纬（n 变体） / n-parameterised apparent declination of Jupiter.
func JupiterApparentDecN(jd float64, n int) float64 {
	return planetApparentDecManualN(4, jd, n)
}

// JupiterApparentRaDecN 木星视赤经赤纬（n 变体） / n-parameterised apparent right ascension and declination of Jupiter.
func JupiterApparentRaDecN(jd float64, n int) (float64, float64) {
	return planetApparentRaDecManualN(4, jd, n)
}

// EarthJupiterAwayN 地木距离（n 变体） / n-parameterised Earth-Jupiter distance.
func EarthJupiterAwayN(jd float64, n int) float64 {
	return planetEarthAwayN(4, jd, n)
}

// JupiterMagNWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func JupiterMagNWithInstance(jde float64, n int, inst *Instance) float64 {
	awaySun := PlanetCoordinateNWithInstance(4, 2, jde, n, inst)
	awayEarth := EarthJupiterAwayNWithInstance(jde, n, inst)
	away := PlanetCoordinateNWithInstance(-1, 2, jde, n, inst)
	i := (awaySun*awaySun + awayEarth*awayEarth - away*away) / (2 * awaySun * awayEarth)
	i = ArcCos(i)
	mag := -9.40 + 5*math.Log10(awaySun*awayEarth) + 0.0005*i
	return FloatRound(mag, 2)
}

// JupiterHeightNWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func JupiterHeightNWithInstance(localJD, lon, lat, timezone float64, n int, inst *Instance) float64 {
	return planetHeightN(4, localJD, lon, lat, timezone, n, inst)
}

// JupiterAzimuthNWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func JupiterAzimuthNWithInstance(localJD, lon, lat, timezone float64, n int, inst *Instance) float64 {
	return planetAzimuthN(4, localJD, lon, lat, timezone, n, inst)
}

// JupiterHourAngleNWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func JupiterHourAngleNWithInstance(jd, lon, timezone float64, n int, inst *Instance) float64 {
	return planetHourAngleN(jd, lon, timezone, n, func(jd float64, n int) float64 { return JupiterApparentRaNWithInstance(jd, n, inst) })
}

// JupiterCulminationTimeNWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func JupiterCulminationTimeNWithInstance(localJD, lon, timezone float64, n int, inst *Instance) float64 {
	return planetCulminationTimeN(localJD, lon, timezone, n, func(jd, lon, timezone float64, n int) float64 {
		return JupiterHourAngleNWithInstance(jd, lon, timezone, n, inst)
	})
}

// JupiterRiseTimeNWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func JupiterRiseTimeNWithInstance(jd, lon, lat, timezone, aeroCorrection, observerHeight float64, n int, inst *Instance) (float64, error) {
	return planetRiseDownN(jd, lon, lat, timezone, aeroCorrection, observerHeight, true, n, func(localJD, lon, timezone float64, n int) float64 {
		return JupiterCulminationTimeNWithInstance(localJD, lon, timezone, n, inst)
	}, func(localJD, lon, lat, timezone float64, n int) float64 {
		return JupiterHeightNWithInstance(localJD, lon, lat, timezone, n, inst)
	}, func(jd float64, n int) float64 { return JupiterApparentDecNWithInstance(jd, n, inst) })
}

// JupiterSetTimeNWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func JupiterSetTimeNWithInstance(jd, lon, lat, timezone, aeroCorrection, observerHeight float64, n int, inst *Instance) (float64, error) {
	return planetRiseDownN(jd, lon, lat, timezone, aeroCorrection, observerHeight, false, n, func(localJD, lon, timezone float64, n int) float64 {
		return JupiterCulminationTimeNWithInstance(localJD, lon, timezone, n, inst)
	}, func(localJD, lon, lat, timezone float64, n int) float64 {
		return JupiterHeightNWithInstance(localJD, lon, lat, timezone, n, inst)
	}, func(jd float64, n int) float64 { return JupiterApparentDecNWithInstance(jd, n, inst) })
}

// SaturnApparentLoN 土星视黄经（n 变体） / n-parameterised apparent ecliptic longitude of Saturn.
func SaturnApparentLoN(jd float64, n int) float64 {
	lo, _ := planetApparentLoBoN(5, jd, n)
	return lo
}

// SaturnApparentBoN 土星视黄纬（n 变体） / n-parameterised apparent ecliptic latitude of Saturn.
func SaturnApparentBoN(jd float64, n int) float64 {
	_, bo := planetApparentLoBoN(5, jd, n)
	return bo
}

// SaturnApparentLoBoN 土星视黄经黄纬（n 变体） / n-parameterised apparent ecliptic longitude and latitude of Saturn.
func SaturnApparentLoBoN(jd float64, n int) (float64, float64) {
	return planetApparentLoBoN(5, jd, n)
}

// SaturnApparentRaN 土星视赤经（n 变体） / n-parameterised apparent right ascension of Saturn.
func SaturnApparentRaN(jd float64, n int) float64 {
	return planetApparentRaManualN(5, jd, n)
}

// SaturnApparentDecN 土星视赤纬（n 变体） / n-parameterised apparent declination of Saturn.
func SaturnApparentDecN(jd float64, n int) float64 {
	return planetApparentDecManualN(5, jd, n)
}

// SaturnApparentRaDecN 土星视赤经赤纬（n 变体） / n-parameterised apparent right ascension and declination of Saturn.
func SaturnApparentRaDecN(jd float64, n int) (float64, float64) {
	return planetApparentRaDecManualN(5, jd, n)
}

// EarthSaturnAwayN 地土距离（n 变体） / n-parameterised Earth-Saturn distance.
func EarthSaturnAwayN(jd float64, n int) float64 {
	return planetEarthAwayN(5, jd, n)
}

// SaturnMagNWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func SaturnMagNWithInstance(jde float64, n int, inst *Instance) float64 {
	awaySun := PlanetCoordinateNWithInstance(5, 2, jde, n, inst)
	awayEarth := EarthSaturnAwayNWithInstance(jde, n, inst)
	ringB, _, _, deltaU, _, _ := SaturnRingParametersNWithInstance(jde, n, inst)
	ringB = math.Abs(ringB)
	mag := -8.68 + 5*math.Log10(awaySun*awayEarth) + 0.044*deltaU - 2.6*Sin(ringB) + 1.25*Sin(ringB)*Sin(ringB)
	return FloatRound(mag, 2)
}

// SaturnHeightNWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func SaturnHeightNWithInstance(localJD, lon, lat, timezone float64, n int, inst *Instance) float64 {
	return planetHeightN(5, localJD, lon, lat, timezone, n, inst)
}

// SaturnAzimuthNWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func SaturnAzimuthNWithInstance(localJD, lon, lat, timezone float64, n int, inst *Instance) float64 {
	return planetAzimuthN(5, localJD, lon, lat, timezone, n, inst)
}

// SaturnHourAngleNWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func SaturnHourAngleNWithInstance(jd, lon, timezone float64, n int, inst *Instance) float64 {
	return planetHourAngleN(jd, lon, timezone, n, func(jd float64, n int) float64 { return SaturnApparentRaNWithInstance(jd, n, inst) })
}

// SaturnCulminationTimeNWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func SaturnCulminationTimeNWithInstance(localJD, lon, timezone float64, n int, inst *Instance) float64 {
	return planetCulminationTimeN(localJD, lon, timezone, n, func(jd, lon, timezone float64, n int) float64 {
		return SaturnHourAngleNWithInstance(jd, lon, timezone, n, inst)
	})
}

// SaturnRiseTimeNWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func SaturnRiseTimeNWithInstance(jd, lon, lat, timezone, aeroCorrection, observerHeight float64, n int, inst *Instance) (float64, error) {
	return planetRiseDownN(jd, lon, lat, timezone, aeroCorrection, observerHeight, true, n, func(localJD, lon, timezone float64, n int) float64 {
		return SaturnCulminationTimeNWithInstance(localJD, lon, timezone, n, inst)
	}, func(localJD, lon, lat, timezone float64, n int) float64 {
		return SaturnHeightNWithInstance(localJD, lon, lat, timezone, n, inst)
	}, func(jd float64, n int) float64 { return SaturnApparentDecNWithInstance(jd, n, inst) })
}

// SaturnSetTimeNWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func SaturnSetTimeNWithInstance(jd, lon, lat, timezone, aeroCorrection, observerHeight float64, n int, inst *Instance) (float64, error) {
	return planetRiseDownN(jd, lon, lat, timezone, aeroCorrection, observerHeight, false, n, func(localJD, lon, timezone float64, n int) float64 {
		return SaturnCulminationTimeNWithInstance(localJD, lon, timezone, n, inst)
	}, func(localJD, lon, lat, timezone float64, n int) float64 {
		return SaturnHeightNWithInstance(localJD, lon, lat, timezone, n, inst)
	}, func(jd float64, n int) float64 { return SaturnApparentDecNWithInstance(jd, n, inst) })
}

// UranusApparentLoN 天王星视黄经（n 变体） / n-parameterised apparent ecliptic longitude of Uranus.
func UranusApparentLoN(jd float64, n int) float64 {
	lo, _ := planetApparentLoBoN(6, jd, n)
	return lo
}

// UranusApparentBoN 天王星视黄纬（n 变体） / n-parameterised apparent ecliptic latitude of Uranus.
func UranusApparentBoN(jd float64, n int) float64 {
	_, bo := planetApparentLoBoN(6, jd, n)
	return bo
}

// UranusApparentLoBoN 天王星视黄经黄纬（n 变体） / n-parameterised apparent ecliptic longitude and latitude of Uranus.
func UranusApparentLoBoN(jd float64, n int) (float64, float64) {
	return planetApparentLoBoN(6, jd, n)
}

// UranusApparentRaN 天王星视赤经（n 变体） / n-parameterised apparent right ascension of Uranus.
func UranusApparentRaN(jd float64, n int) float64 {
	return planetApparentRaManualN(6, jd, n)
}

// UranusApparentDecN 天王星视赤纬（n 变体） / n-parameterised apparent declination of Uranus.
func UranusApparentDecN(jd float64, n int) float64 {
	return planetApparentDecManualN(6, jd, n)
}

// UranusApparentRaDecN 天王星视赤经赤纬（n 变体） / n-parameterised apparent right ascension and declination of Uranus.
func UranusApparentRaDecN(jd float64, n int) (float64, float64) {
	return planetApparentRaDecManualN(6, jd, n)
}

// EarthUranusAwayN 地天距离（n 变体） / n-parameterised Earth-Uranus distance.
func EarthUranusAwayN(jd float64, n int) float64 {
	return planetEarthAwayN(6, jd, n)
}

// UranusMagNWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func UranusMagNWithInstance(jde float64, n int, inst *Instance) float64 {
	awaySun := PlanetCoordinateNWithInstance(6, 2, jde, n, inst)
	awayEarth := EarthUranusAwayNWithInstance(jde, n, inst)
	away := PlanetCoordinateNWithInstance(-1, 2, jde, n, inst)
	i := (awaySun*awaySun + awayEarth*awayEarth - away*away) / (2 * awaySun * awayEarth)
	i = ArcCos(i)
	mag := -7.19 + 5*math.Log10(awaySun*awayEarth) + 0.016*i
	return FloatRound(mag, 2)
}

// UranusHeightNWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func UranusHeightNWithInstance(localJD, lon, lat, timezone float64, n int, inst *Instance) float64 {
	return planetHeightN(6, localJD, lon, lat, timezone, n, inst)
}

// UranusAzimuthNWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func UranusAzimuthNWithInstance(localJD, lon, lat, timezone float64, n int, inst *Instance) float64 {
	return planetAzimuthN(6, localJD, lon, lat, timezone, n, inst)
}

// UranusHourAngleNWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func UranusHourAngleNWithInstance(jd, lon, timezone float64, n int, inst *Instance) float64 {
	return planetHourAngleN(jd, lon, timezone, n, func(jd float64, n int) float64 { return UranusApparentRaNWithInstance(jd, n, inst) })
}

// UranusCulminationTimeNWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func UranusCulminationTimeNWithInstance(localJD, lon, timezone float64, n int, inst *Instance) float64 {
	return planetCulminationTimeN(localJD, lon, timezone, n, func(jd, lon, timezone float64, n int) float64 {
		return UranusHourAngleNWithInstance(jd, lon, timezone, n, inst)
	})
}

// UranusRiseTimeNWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func UranusRiseTimeNWithInstance(jd, lon, lat, timezone, aeroCorrection, observerHeight float64, n int, inst *Instance) (float64, error) {
	return planetRiseDownN(jd, lon, lat, timezone, aeroCorrection, observerHeight, true, n, func(localJD, lon, timezone float64, n int) float64 {
		return UranusCulminationTimeNWithInstance(localJD, lon, timezone, n, inst)
	}, func(localJD, lon, lat, timezone float64, n int) float64 {
		return UranusHeightNWithInstance(localJD, lon, lat, timezone, n, inst)
	}, func(jd float64, n int) float64 { return UranusApparentDecNWithInstance(jd, n, inst) })
}

// UranusSetTimeNWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func UranusSetTimeNWithInstance(jd, lon, lat, timezone, aeroCorrection, observerHeight float64, n int, inst *Instance) (float64, error) {
	return planetRiseDownN(jd, lon, lat, timezone, aeroCorrection, observerHeight, false, n, func(localJD, lon, timezone float64, n int) float64 {
		return UranusCulminationTimeNWithInstance(localJD, lon, timezone, n, inst)
	}, func(localJD, lon, lat, timezone float64, n int) float64 {
		return UranusHeightNWithInstance(localJD, lon, lat, timezone, n, inst)
	}, func(jd float64, n int) float64 { return UranusApparentDecNWithInstance(jd, n, inst) })
}

// NeptuneApparentLoN 海王星视黄经（n 变体） / n-parameterised apparent ecliptic longitude of Neptune.
func NeptuneApparentLoN(jd float64, n int) float64 {
	lo, _ := planetApparentLoBoN(7, jd, n)
	return lo
}

// NeptuneApparentBoN 海王星视黄纬（n 变体） / n-parameterised apparent ecliptic latitude of Neptune.
func NeptuneApparentBoN(jd float64, n int) float64 {
	_, bo := planetApparentLoBoN(7, jd, n)
	return bo
}

// NeptuneApparentLoBoN 海王星视黄经黄纬（n 变体） / n-parameterised apparent ecliptic longitude and latitude of Neptune.
func NeptuneApparentLoBoN(jd float64, n int) (float64, float64) {
	return planetApparentLoBoN(7, jd, n)
}

// NeptuneApparentRaN 海王星视赤经（n 变体） / n-parameterised apparent right ascension of Neptune.
func NeptuneApparentRaN(jd float64, n int) float64 {
	return planetApparentRaManualN(7, jd, n)
}

// NeptuneApparentDecN 海王星视赤纬（n 变体） / n-parameterised apparent declination of Neptune.
func NeptuneApparentDecN(jd float64, n int) float64 {
	return planetApparentDecManualN(7, jd, n)
}

// NeptuneApparentRaDecN 海王星视赤经赤纬（n 变体） / n-parameterised apparent right ascension and declination of Neptune.
func NeptuneApparentRaDecN(jd float64, n int) (float64, float64) {
	return planetApparentRaDecManualN(7, jd, n)
}

// EarthNeptuneAwayN 地海距离（n 变体） / n-parameterised Earth-Neptune distance.
func EarthNeptuneAwayN(jd float64, n int) float64 {
	return planetEarthAwayN(7, jd, n)
}

// NeptuneMagNWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func NeptuneMagNWithInstance(jde float64, n int, inst *Instance) float64 {
	awaySun := PlanetCoordinateNWithInstance(7, 2, jde, n, inst)
	awayEarth := EarthNeptuneAwayNWithInstance(jde, n, inst)
	mag := -6.87 + 5*math.Log10(awaySun*awayEarth)
	return FloatRound(mag, 2)
}

// NeptuneHeightNWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func NeptuneHeightNWithInstance(localJD, lon, lat, timezone float64, n int, inst *Instance) float64 {
	return planetHeightN(7, localJD, lon, lat, timezone, n, inst)
}

// NeptuneAzimuthNWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func NeptuneAzimuthNWithInstance(localJD, lon, lat, timezone float64, n int, inst *Instance) float64 {
	return planetAzimuthN(7, localJD, lon, lat, timezone, n, inst)
}

// NeptuneHourAngleNWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func NeptuneHourAngleNWithInstance(jd, lon, timezone float64, n int, inst *Instance) float64 {
	return planetHourAngleN(jd, lon, timezone, n, func(jd float64, n int) float64 { return NeptuneApparentRaNWithInstance(jd, n, inst) })
}

// NeptuneCulminationTimeNWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func NeptuneCulminationTimeNWithInstance(localJD, lon, timezone float64, n int, inst *Instance) float64 {
	return planetCulminationTimeN(localJD, lon, timezone, n, func(jd, lon, timezone float64, n int) float64 {
		return NeptuneHourAngleNWithInstance(jd, lon, timezone, n, inst)
	})
}

// NeptuneRiseTimeNWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func NeptuneRiseTimeNWithInstance(jd, lon, lat, timezone, aeroCorrection, observerHeight float64, n int, inst *Instance) (float64, error) {
	return planetRiseDownN(jd, lon, lat, timezone, aeroCorrection, observerHeight, true, n, func(localJD, lon, timezone float64, n int) float64 {
		return NeptuneCulminationTimeNWithInstance(localJD, lon, timezone, n, inst)
	}, func(localJD, lon, lat, timezone float64, n int) float64 {
		return NeptuneHeightNWithInstance(localJD, lon, lat, timezone, n, inst)
	}, func(jd float64, n int) float64 { return NeptuneApparentDecNWithInstance(jd, n, inst) })
}

// NeptuneSetTimeNWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func NeptuneSetTimeNWithInstance(jd, lon, lat, timezone, aeroCorrection, observerHeight float64, n int, inst *Instance) (float64, error) {
	return planetRiseDownN(jd, lon, lat, timezone, aeroCorrection, observerHeight, false, n, func(localJD, lon, timezone float64, n int) float64 {
		return NeptuneCulminationTimeNWithInstance(localJD, lon, timezone, n, inst)
	}, func(localJD, lon, lat, timezone float64, n int) float64 {
		return NeptuneHeightNWithInstance(localJD, lon, lat, timezone, n, inst)
	}, func(jd float64, n int) float64 { return NeptuneApparentDecNWithInstance(jd, n, inst) })
}
