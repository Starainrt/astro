package basic

import (
	"math"

	"github.com/starainrt/astro/internal/geodata"
)

// 月缘修正是否生效的统一判据：每种量纲只有一个容差，所有层都经下面的函数判断，避免口径漂移。
const (
	// limbAppliedTimeToleranceDays 接触时刻与时长的容差（0.05 秒）。
	limbAppliedTimeToleranceDays = 0.05 / 86400.0
	// limbAppliedDistanceToleranceKM 地面界线点的平移容差（0.05 km）。
	limbAppliedDistanceToleranceKM = 0.05
	// limbAppliedMagnitudeTolerance 食分与遮蔽率的容差。
	limbAppliedMagnitudeTolerance = 1e-9
)

// limbTimeChanged 报告两个儒略日（或以日为单位的时长）是否相差超过时刻容差；任一非有限时视为未变。
func limbTimeChanged(corrected, original float64) bool {
	return finite(corrected) && finite(original) &&
		math.Abs(corrected-original) > limbAppliedTimeToleranceDays
}

// limbMagnitudeChanged 报告食分或遮蔽率是否相差超过容差；任一非有限时视为未变。
func limbMagnitudeChanged(corrected, original float64) bool {
	return finite(corrected) && finite(original) &&
		math.Abs(corrected-original) > limbAppliedMagnitudeTolerance
}

// limbGroundMoved 报告地面点（度）是否平移超过距离容差。
func limbGroundMoved(originalLon, originalLat, correctedLon, correctedLat float64) bool {
	longitude := signedAngleDifference(correctedLon, originalLon) * math.Cos(originalLat*rad)
	return math.Hypot(longitude, correctedLat-originalLat)*geodata.EarthKMPerDegree > limbAppliedDistanceToleranceKM
}
