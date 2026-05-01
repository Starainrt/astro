package basic

import (
	"math"

	. "github.com/starainrt/astro/tools"
)

const (
	sunPhysicalInclinationDeg = 7.25
	sunPhysicalKBaseDeg       = 73.6667
	sunPhysicalKRateDeg       = 1.3958333
	sunCarringtonStartJD      = 2398220.0
	sunCarringtonRotationDays = 25.38
	sunPhysicalKEpochJD       = 2396758.0
)

// SunPhysicalInfo 太阳物理观测参数 / physical observing parameters of the Sun.
type SunPhysicalInfo struct {
	// P 太阳北极位置角，单位度 / position angle of the solar north pole in degrees.
	P float64
	// B0 日面中心太阳纬度，单位度 / heliographic latitude of the disk center in degrees.
	B0 float64
	// L0 日面中心卡林顿经度，单位度 / Carrington heliographic longitude of the disk center in degrees.
	L0 float64
}

// SunPhysical 太阳物理观测参数 / physical observing parameters of the Sun.
func SunPhysical(jd float64) SunPhysicalInfo {
	return SunPhysicalN(jd, -1)
}

// SunPhysicalN 太阳物理观测参数（截断版） / truncated physical observing parameters of the Sun.
func SunPhysicalN(jd float64, n int) SunPhysicalInfo {
	lambda := HSunApparentLoN(jd, n)
	epsilon := TrueObliquity(jd)
	k := sunPhysicalKBaseDeg + sunPhysicalKRateDeg*(jd-sunPhysicalKEpochJD)/36525.0
	theta := (jd - sunCarringtonStartJD) * 360.0 / sunCarringtonRotationDays

	x := math.Atan(-Cos(lambda)*Tan(epsilon)) * 180.0 / math.Pi
	y := math.Atan(-Cos(lambda-k)*Tan(sunPhysicalInclinationDeg)) * 180.0 / math.Pi
	p := Limit360(x + y)
	b0 := ArcSin(Sin(lambda-k) * Sin(sunPhysicalInclinationDeg))
	eta := ArcTan2(-Sin(lambda-k)*Cos(sunPhysicalInclinationDeg), -Cos(lambda-k))
	l0 := Limit360(eta - theta)

	return SunPhysicalInfo{
		P:  p,
		B0: b0,
		L0: l0,
	}
}
