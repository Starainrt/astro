package tools

import "math"

// DistanceUnit 距离输入单位 / a unit accepted for distance input.
type DistanceUnit uint8

const (
	// DistanceParsec 秒差距 / parsec.
	DistanceParsec DistanceUnit = iota
	// DistanceLightYear 光年 / light-year.
	DistanceLightYear
	// DistanceAU 天文单位 / astronomical unit.
	DistanceAU
)

// 天文单位与光年取 IAU 定义值，秒差距由精确关系 648000/π 天文单位导出。
const (
	AstronomicalUnitKilometers = 149597870.7
	LightYearKilometers        = 9460730472580.8
	parsecAU                   = 648000 / math.Pi
)

// DistanceToParsecs 把给定单位的距离换算为秒差距 / converts a distance in the given unit to parsecs.
//
// 非正数、NaN 与未知单位返回 NaN。
func DistanceToParsecs(value float64, unit DistanceUnit) float64 {
	if math.IsNaN(value) || math.IsInf(value, 0) || value <= 0 {
		return math.NaN()
	}
	switch unit {
	case DistanceParsec:
		return value
	case DistanceLightYear:
		return value * LightYearKilometers / AstronomicalUnitKilometers / parsecAU
	case DistanceAU:
		return value / parsecAU
	}
	return math.NaN()
}
