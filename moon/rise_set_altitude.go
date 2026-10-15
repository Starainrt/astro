package moon

import (
	"time"

	"github.com/starainrt/astro/basic"
)

// AltitudeLimb 取高度的位置（上缘、中心、下缘）/ limb used for the altitude.
type AltitudeLimb = basic.AltitudeLimb

const (
	LimbUpper  = basic.AltitudeLimbUpper
	LimbCenter = basic.AltitudeLimbCenter
	LimbLower  = basic.AltitudeLimbLower
)

// AltitudeOptions 升落选项；零值为上缘、标准气象、计入折射 / altitude crossing options.
type AltitudeOptions = basic.AltitudeCrossingOptions

// AltitudeState 一个民用日内的穿越状态 / crossing state over one civil day.
type AltitudeState = basic.AltitudeCrossingState

const (
	AltitudeCrosses     = basic.AltitudeCrosses
	AltitudeAlwaysAbove = basic.AltitudeAlwaysAbove
	AltitudeAlwaysBelow = basic.AltitudeAlwaysBelow
	AltitudeTangent     = basic.AltitudeTangent
)

// AltitudeCrossingResult date 所在民用日内首次升、落与残差范围（度）/ first rise and set on the civil day.
type AltitudeCrossingResult = basic.AltitudeCrossingResult

// RiseSet 月球在 date 所在民用日内穿过观测者地平线（含 height 带来的地平俯角）的升落；options 控制月缘与大气 / lunar rise and set on the civil day.
func (module *Moon) RiseSet(date time.Time, lon, lat, height float64, options AltitudeOptions) (AltitudeCrossingResult, error) {
	return basic.SkyAltitudeCrossingOnDate(basic.SkyMoon, date, lon, lat, height, -basic.HeightDegreeByLat(height, lat), options, module.inst)
}

// RiseSet 月球在 date 所在民用日内穿过观测者地平线（含 height 带来的地平俯角）的升落；options 控制月缘与大气 / lunar rise and set on the civil day.
func RiseSet(date time.Time, lon, lat, height float64, options AltitudeOptions) (AltitudeCrossingResult, error) {
	return defaultModule.RiseSet(date, lon, lat, height, options)
}
