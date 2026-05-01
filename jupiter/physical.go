package jupiter

import (
	"time"

	"github.com/starainrt/astro/basic"
)

// PhysicalInfo 木星物理观测参数 / physical observing parameters of Jupiter.
type PhysicalInfo struct {
	// SubEarthLongitude 子地经度，单位度；采用 Jupiter 当前 IAU/Horizons 西经为正约定。
	SubEarthLongitude float64
	// SubEarthLatitude 子地纬度，单位度。
	SubEarthLatitude float64
	// SubSolarLongitude 子日经度，单位度；采用 Jupiter 当前 IAU/Horizons 西经为正约定。
	SubSolarLongitude float64
	// SubSolarLatitude 子日纬度，单位度。
	SubSolarLatitude float64
	// NorthPolePositionAngle 木星北极位置角，单位度。
	NorthPolePositionAngle float64
	// DS 太阳相对木星赤道的行星中心赤纬，单位度。
	DS float64
	// DE 地球相对木星赤道的行星中心赤纬，单位度。
	DE float64
	// CentralMeridianSystemI 木星 System I 照亮盘中央经线，单位度，西经为正。
	CentralMeridianSystemI float64
	// CentralMeridianSystemII 木星 System II 照亮盘中央经线，单位度，西经为正。
	CentralMeridianSystemII float64
	// CentralMeridianSystemIII 木星 System III 盘面中央经线，单位度，西经为正。
	CentralMeridianSystemIII float64
}

// CentralMeridianInfo 木星 System I/II/III 中央经线 / Jupiter System I/II/III central meridians.
type CentralMeridianInfo struct {
	// SystemI 木星 System I 照亮盘中央经线，单位度，西经为正。
	SystemI float64
	// SystemII 木星 System II 照亮盘中央经线，单位度，西经为正。
	SystemII float64
	// SystemIII 木星 System III 盘面中央经线，单位度，西经为正。
	SystemIII float64
}

// Physical 木星物理观测参数 / physical observing parameters of Jupiter.
func Physical(date time.Time) PhysicalInfo {
	return PhysicalN(date, -1)
}

// PhysicalN 木星物理观测参数（截断版） / truncated physical observing parameters of Jupiter.
func PhysicalN(date time.Time, n int) PhysicalInfo {
	jde := basic.Date2JDE(date.UTC())
	jd := basic.TD2UT(jde, true)
	info := basic.JupiterPhysicalN(jd, n)
	meridians := basic.JupiterCentralMeridiansN(jd, n)
	ds, de := basic.JupiterDSDEN(jd, n)
	return PhysicalInfo{
		SubEarthLongitude:        info.SubEarthLongitude,
		SubEarthLatitude:         info.SubEarthLatitude,
		SubSolarLongitude:        info.SubSolarLongitude,
		SubSolarLatitude:         info.SubSolarLatitude,
		NorthPolePositionAngle:   info.NorthPolePositionAngle,
		DS:                       ds,
		DE:                       de,
		CentralMeridianSystemI:   meridians.SystemI,
		CentralMeridianSystemII:  meridians.SystemII,
		CentralMeridianSystemIII: meridians.SystemIII,
	}
}

// CentralMeridians 木星 System I/II/III 中央经线 / Jupiter System I/II/III central meridians.
func CentralMeridians(date time.Time) CentralMeridianInfo {
	return CentralMeridiansN(date, -1)
}

// CentralMeridiansN 木星 System I/II/III 中央经线（截断版） / truncated Jupiter System I/II/III central meridians.
func CentralMeridiansN(date time.Time, n int) CentralMeridianInfo {
	jde := basic.Date2JDE(date.UTC())
	info := basic.JupiterCentralMeridiansN(basic.TD2UT(jde, true), n)
	return CentralMeridianInfo{
		SystemI:   info.SystemI,
		SystemII:  info.SystemII,
		SystemIII: info.SystemIII,
	}
}
