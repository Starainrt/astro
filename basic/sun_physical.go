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
func SunPhysical(jde float64) SunPhysicalInfo {
	return SunPhysicalN(jde, -1)
}

// SunPhysicalNWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func SunPhysicalNWithInstance(jde float64, n int, inst *Instance) SunPhysicalInfo {
	lambda := HSunApparentLoNWithInstance(jde, n, inst)
	epsilon := TrueObliquity(jde)
	k := sunPhysicalKBaseDeg + sunPhysicalKRateDeg*(jde-sunPhysicalKEpochJD)/36525.0
	theta := (jde - sunCarringtonStartJD) * 360.0 / sunCarringtonRotationDays

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
