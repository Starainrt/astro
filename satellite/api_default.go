package satellite

import (
	"github.com/starainrt/astro/basic"
	"time"
)

// CoverageAt 计算单时刻的几何与目视可见范围 / compute instantaneous geometric and visual coverage.
func CoverageAt(e Elements, t time.Time, opt CoverageOptions) (Coverage, error) {
	return defaultModule.CoverageAt(e, t, opt)
}

// CoverageBetween 采样闭区间并自动加密内部扫掠网格，至多24小时 / sample a closed interval up to 24 hours with internal swept-grid refinement.
func CoverageBetween(e Elements, start, end time.Time, step time.Duration, opt CoverageOptions) (CoverageSweep, error) {
	return defaultModule.CoverageBetween(e, start, end, step, opt)
}

// LookAt 计算卫星站心视位置，保留输入时区 / compute a topocentric satellite observation preserving the input time zone.
func LookAt(e Elements, obs Observer, t time.Time, opt ObserveOptions) (Look, error) {
	return defaultModule.LookAt(e, obs, t, opt)
}

// NextPass 搜索此刻起至多30天的下一过境 / find the next pass within 30 days.
func NextPass(e Elements, obs Observer, after time.Time, opt PassOptions) (Pass, error) {
	return defaultModule.NextPass(e, obs, after, opt)
}

// Passes 搜索闭区间内的过境 / search passes in a closed interval.
func Passes(e Elements, obs Observer, start, end time.Time, opt PassOptions) ([]Pass, basic.AltitudeCrossingState, error) {
	return defaultModule.Passes(e, obs, start, end, opt)
}

// Transits 搜索站点凌日凌月与近掠 / search site solar and lunar transits and near misses.
func Transits(e Elements, obs Observer, start, end time.Time, opt TransitOptions) ([]Transit, error) {
	return defaultModule.Transits(e, obs, start, end, opt)
}

// TransitEvents 搜索全球凌日凌月带 / search global solar and lunar transit bands.
func TransitEvents(e Elements, start, end time.Time, opt TransitSearchOptions) ([]TransitEvent, error) {
	return defaultModule.TransitEvents(e, start, end, opt)
}

// ClosestTransitEvent 搜索此刻最近的凌现象带 / find the transit band closest to this time.
func ClosestTransitEvent(e Elements, target TransitTarget, t time.Time, opt TransitSearchOptions) (TransitEvent, error) {
	return defaultModule.ClosestTransitEvent(e, target, t, opt)
}

// TransitEventPath 采样凌现象带的中心线、左右限和高度边界 / sample a transit band's centre line, limits, and altitude boundaries.
func TransitEventPath(e Elements, event TransitEvent, step time.Duration) (TransitPath, error) {
	return defaultModule.TransitEventPath(e, event, step)
}

// NearbyTransits 搜索站点附近的凌现象带与站点事件 / search nearby transit bands and site events.
func NearbyTransits(elements []Elements, obs Observer, start, end time.Time, opt NearbyTransitOptions) ([]NearbyTransit, error) {
	return defaultModule.NearbyTransits(elements, obs, start, end, opt)
}

// TransitFrames 采样站心凌现象轨迹，保留接触帧与输入时区 / sample a local transit, preserving contact frames and the input time zone.
func TransitFrames(e Elements, obs Observer, transit Transit, step time.Duration) (TransitTrack, error) {
	return defaultModule.TransitFrames(e, obs, transit, step)
}

// TransitShadowAt 计算任意时刻的凌现象快照，边缘未全部落地时无效 / compute a transit snapshot, invalid unless the entire limb reaches Earth.
func TransitShadowAt(e Elements, target TransitTarget, t time.Time, opt TransitSearchOptions) (TransitShadow, error) {
	return defaultModule.TransitShadowAt(e, target, t, opt)
}
