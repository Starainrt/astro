package phenomena

import (
	"time"

	"github.com/starainrt/astro/basic"
)

// AltitudeLimb 日月取高度的位置 / limb used for Sun and Moon altitude.
type AltitudeLimb = basic.AltitudeLimb

const (
	LimbUpper  = basic.AltitudeLimbUpper
	LimbCenter = basic.AltitudeLimbCenter
	LimbLower  = basic.AltitudeLimbLower
)

// AltitudeOptions 任意高度升落选项；零值为上缘、标准气象、计入折射 / altitude crossing options.
type AltitudeOptions = basic.AltitudeCrossingOptions

// AltitudeState 一个民用日内的穿越状态 / crossing state over one civil day.
type AltitudeState = basic.AltitudeCrossingState

const (
	AltitudeCrosses     = basic.AltitudeCrosses
	AltitudeAlwaysAbove = basic.AltitudeAlwaysAbove
	AltitudeAlwaysBelow = basic.AltitudeAlwaysBelow
	AltitudeTangent     = basic.AltitudeTangent
)

// AltitudeCrossingResult date 所在民用日内首次升、落（无则为零值，对应 HasRise/HasSet 为 false）与残差范围（度）/ first rise and set on the civil day.
type AltitudeCrossingResult = basic.AltitudeCrossingResult

// AltitudeCrossing 日、月或行星在 date 所在民用日内穿过视高度 altitude（度，相对天文地平）的时刻；height 只影响月球视差 /
// Sun, Moon or planet crossings of an apparent altitude.
func (module *Phenomena) AltitudeCrossing(body Body, date time.Time, lon, lat, height, altitude float64, options AltitudeOptions) (AltitudeCrossingResult, error) {
	return basic.SkyAltitudeCrossingOnDate(body, date, lon, lat, height, altitude, options, module.inst)
}

// StarAltitudeCrossing 恒星（视赤经、视赤纬，度）穿过视高度 altitude 的时刻 / star crossings of an apparent altitude.
func StarAltitudeCrossing(date time.Time, ra, dec, lon, lat, altitude float64, options AltitudeOptions) (AltitudeCrossingResult, error) {
	return basic.StarAltitudeCrossingOnDate(date, ra, dec, lon, lat, altitude, options)
}
