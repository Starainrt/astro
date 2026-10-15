package basic

import (
	"github.com/starainrt/astro/internal/ephem"
)

// Instance 显式只读星历实例 / explicit read-only ephemeris instance.
type Instance = ephem.Instance

// EphemerisModel 几何位置所用的星历模型 / ephemeris model behind geometric positions.
type EphemerisModel int

const (
	// EphemerisLegacy VSOP87 + ELP2000 截断级数，astro_legacy 构建下为默认 / truncated VSOP87 and ELP2000 series, default under astro_legacy.
	EphemerisLegacy EphemerisModel = iota
	// EphemerisSemianalytic VSOP2013/TOP2013 + ELP/MPP02 结构、按 DE441 重拟的半解析表 / semi-analytic tables refitted to DE441.
	EphemerisSemianalytic
)

// SetEphemerisModel 切换全局星历模型，未知值返回 false / switches the global ephemeris model; unknown values return false.
func SetEphemerisModel(m EphemerisModel) bool {
	if m != EphemerisLegacy && m != EphemerisSemianalytic {
		return false
	}
	// 依赖星历的缓存都按 ΔT 世代校验，切换时一并递增。
	deltaTFnMu.Lock()
	if ephem.SetActive(ephem.Model(m)) {
		deltaTGeneration++
	}
	deltaTFnMu.Unlock()
	return true
}

// GetEphemerisModel 当前星历模型 / the active ephemeris model.
func GetEphemerisModel() EphemerisModel { return EphemerisModel(ephem.Active()) }
