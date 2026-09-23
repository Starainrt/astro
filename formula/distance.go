package formula

import "github.com/starainrt/astro/tools"

// DistanceUnit 距离输入单位，与 tools 共用同一套枚举 / a unit accepted for distance input, shared with tools.
type DistanceUnit = tools.DistanceUnit

const (
	// DistanceParsec 秒差距 / parsec.
	DistanceParsec = tools.DistanceParsec
	// DistanceLightYear 光年 / light-year.
	DistanceLightYear = tools.DistanceLightYear
	// DistanceAU 天文单位 / astronomical unit.
	DistanceAU = tools.DistanceAU
)

// Distance 把给定单位的距离换算为秒差距 / converts a distance in the given unit to parsecs.
//
// 换算基元位于 tools，本函数只是对外转发；非正数、NaN 与未知单位返回 NaN。
func Distance(value float64, unit DistanceUnit) float64 {
	return tools.DistanceToParsecs(value, unit)
}
