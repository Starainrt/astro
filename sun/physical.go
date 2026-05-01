package sun

import (
	"time"

	"github.com/starainrt/astro/basic"
)

// PhysicalInfo 太阳物理观测参数 / physical observing parameters of the Sun.
type PhysicalInfo struct {
	// P 太阳北极位置角，单位度 / position angle of the solar north pole in degrees.
	P float64
	// B0 日面中心太阳纬度，单位度 / heliographic latitude of the disk center in degrees.
	B0 float64
	// L0 日面中心卡林顿经度，单位度 / Carrington heliographic longitude of the disk center in degrees.
	L0 float64
}

// Physical 太阳物理观测参数 / physical observing parameters of the Sun.
func Physical(date time.Time) PhysicalInfo {
	return PhysicalN(date, -1)
}

// PhysicalN 太阳物理观测参数（截断版） / truncated physical observing parameters of the Sun.
func PhysicalN(date time.Time, n int) PhysicalInfo {
	jde := basic.Date2JDE(date.UTC())
	info := basic.SunPhysicalN(basic.TD2UT(jde, true), n)
	return PhysicalInfo{
		P:  info.P,
		B0: info.B0,
		L0: info.L0,
	}
}
