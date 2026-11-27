package basic

import (
	"math"

	"github.com/starainrt/astro/tools"
)

// BodyOrientation 行星观测时刻的 ICRF 姿态 / planetary ICRF orientation at the observation epoch.
type BodyOrientation struct {
	// PoleRA 北极 ICRF 赤经，度 / ICRF pole right ascension in degrees.
	PoleRA float64
	// PoleDec 北极 ICRF 赤纬，度 / ICRF pole declination in degrees.
	PoleDec float64
	// PrimeMeridianAngle 本初子午线角 W，度 [0,360) / prime meridian angle W in degrees [0,360).
	PrimeMeridianAngle float64
	// RotationRate W 的线性速率，度/日，逆行为负 / linear W rate in degrees/day, negative for retrograde rotation.
	RotationRate float64
	// OutsideValidity 超出本库认定的姿态有效区间 / outside the library's adopted orientation validity interval.
	OutsideValidity bool
}

// BodyOrientationAt 在 TT/TDB 儒略日求行星姿态，区间外仍计算 / planetary orientation at a TT/TDB Julian day, including outside validity.
func BodyOrientationAt(body SkyBody, jdTDB float64) (BodyOrientation, error) {
	model, rate, fromYear, toYear, ok := bodyOrientationModel(body)
	if !ok || !isFiniteFloat(jdTDB) {
		return BodyOrientation{}, ErrInvalidObservationInput
	}
	ra, dec, w := model.poleRotation(jdTDB)
	if !isFiniteFloat(ra) || !isFiniteFloat(dec) || !isFiniteFloat(w) {
		return BodyOrientation{}, ErrInvalidObservationInput
	}
	// 有效年代以 J2000 起算的 TT 儒略年计，包含两端。
	fromJD := 2451545 + (fromYear-2000)*365.25
	toJD := 2451545 + (toYear-2000)*365.25
	return BodyOrientation{
		PoleRA:             ra,
		PoleDec:            dec,
		PrimeMeridianAngle: tools.Limit360(w),
		RotationRate:       rate,
		OutsideValidity:    jdTDB < fromJD || jdTDB > toJD,
	}, nil
}

func bodyOrientationModel(body SkyBody) (planetPhysicalModel, float64, float64, float64, bool) {
	switch body {
	case SkyMercury:
		return mercuryPhysicalModel, 6.1385108, 1800, 2200, true
	case SkyVenus:
		return venusPhysicalModel, -1.4813688, 1800, 2200, true
	case SkyMars:
		return marsPhysicalModel, 350.891982443297, 1800, 2200, true
	case SkyJupiter:
		return jupiterPhysicalModel, 870.536, 1900, 2100, true
	case SkySaturn:
		return saturnPhysicalModel, 810.7939024, 1950, 2050, true
	case SkyUranus:
		return uranusPhysicalModel, -501.1600928, 1950, 2050, true
	case SkyNeptune:
		return neptunePhysicalModel, 541.1397757, 1950, 2050, true
	default:
		return planetPhysicalModel{}, 0, 0, 0, false
	}
}

func bodyPoleNode(orientation BodyOrientation) (pole, node Vector3) {
	pole = raDecToVector(orientation.PoleRA, orientation.PoleDec)
	node = normalizeVector(Vector3{-pole[1], pole[0], 0})
	return
}

func bodyFixedToICRF(orientation BodyOrientation) Matrix3 {
	pole, node := bodyPoleNode(orientation)
	east := pxp(pole, node)
	wRad := orientation.PrimeMeridianAngle * math.Pi / 180
	sinW, cosW := math.Sincos(wRad)
	prime := Vector3{
		node[0]*cosW + east[0]*sinW,
		node[1]*cosW + east[1]*sinW,
		node[2]*cosW + east[2]*sinW,
	}
	y := pxp(pole, prime)
	return Matrix3{
		{prime[0], y[0], pole[0]},
		{prime[1], y[1], pole[1]},
		{prime[2], y[2], pole[2]},
	}
}
