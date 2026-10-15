package geojson

import (
	eclipsecore "github.com/starainrt/astro/eclipse"
	"github.com/starainrt/astro/mercury"
	"github.com/starainrt/astro/venus"
)

// MarshalLunarEclipse 将月食 P1/P4 站心月心可见区、几何地平线以及 P1-P4 的时间包络编码为 GeoJSON，不含折射。
// MarshalLunarEclipse encodes the P1/P4 topocentric Moon-center visibility regions, their geometric horizons and the P1-P4 time envelopes, without refraction.
// boundaryPoints 小于等于零时使用 360；其他值限制在 [12, 1440]。
// boundaryPoints values <= 0 use 360; other values are clamped to [12, 1440].
func MarshalLunarEclipse(info eclipsecore.LunarEclipseInfo, boundaryPoints int) ([]byte, error) {
	return defaultModule.MarshalLunarEclipse(info, boundaryPoints)
}

// MarshalLunarEclipseWithOptions 编码月食，输出内容与采样精度由 options 选择。
// MarshalLunarEclipseWithOptions encodes a lunar eclipse with the content and sampling selected by options.
func MarshalLunarEclipseWithOptions(
	info eclipsecore.LunarEclipseInfo,
	boundaryPoints int,
	options LunarEclipseOptions,
) ([]byte, error) {
	return defaultModule.MarshalLunarEclipseWithOptions(info, boundaryPoints, options)
}

// MarshalLunarEclipseWithTimeMarkers 编码月食，并沿半影开始到结束的月下点轨迹追加 Point 要素。
// MarshalLunarEclipseWithTimeMarkers encodes a lunar eclipse and adds Point Features along the sublunar track from penumbral start through end.
// 已有要素保持不变；标记标签使用 options.Location，时间值保持 UTC。
// Existing features are unchanged; marker labels use options.Location while time values stay UTC.
func MarshalLunarEclipseWithTimeMarkers(
	info eclipsecore.LunarEclipseInfo,
	boundaryPoints int,
	options TimeMarkerOptions,
) ([]byte, error) {
	return defaultModule.MarshalLunarEclipseWithTimeMarkers(info, boundaryPoints, options)
}

// MarshalSolarEclipse 将日食半影足迹和可选中心食带编码为 GeoJSON。
// MarshalSolarEclipse encodes penumbral footprints and an optional central path as GeoJSON.
func MarshalSolarEclipse(
	partial eclipsecore.SolarEclipsePartialFootprintsInfo,
	central *eclipsecore.SolarEclipsePath,
) ([]byte, error) {
	return defaultModule.MarshalSolarEclipse(partial, central)
}

// MarshalSolarEclipseShadowInstant 单时刻阴影足迹的最小 GeoJSON：只含该时刻的区域与物理边界 / one instant as the smallest useful GeoJSON.
func MarshalSolarEclipseShadowInstant(
	instant eclipsecore.SolarEclipseShadowInstant,
) ([]byte, error) {
	return defaultModule.MarshalSolarEclipseShadowInstant(instant)
}

// MarshalSolarEclipseWithOptions 编码日食，输出内容由 options 选择 / encodes a solar eclipse with the content selected by options.
func MarshalSolarEclipseWithOptions(
	partial eclipsecore.SolarEclipsePartialFootprintsInfo,
	central *eclipsecore.SolarEclipsePath,
	options SolarEclipseOptions,
) ([]byte, error) {
	return defaultModule.MarshalSolarEclipseWithOptions(partial, central, options)
}

// MarshalSolarEclipseWithTimeMarkers 编码日食，并沿中心线按固定间隔追加 Point 要素；已有要素不变，标记标签使用 options.Location，时间值保持 UTC。
// MarshalSolarEclipseWithTimeMarkers encodes a solar eclipse and adds Point Features at regular intervals along the central line. Existing features are unchanged; marker labels use options.Location while time values stay UTC.
func MarshalSolarEclipseWithTimeMarkers(
	partial eclipsecore.SolarEclipsePartialFootprintsInfo,
	central *eclipsecore.SolarEclipsePath,
	options TimeMarkerOptions,
) ([]byte, error) {
	return defaultModule.MarshalSolarEclipseWithTimeMarkers(partial, central, options)
}

// MarshalMercuryTransit 将水星凌日的全球可见分区、一触/四触视地平线与可选站心接触等时线编码为 GeoJSON。
// MarshalMercuryTransit encodes the global visibility regions, C1/C4 apparent horizons and optional topocentric contact isochrones of a Mercury transit.
func MarshalMercuryTransit(info mercury.TransitInfo, options PlanetTransitOptions) ([]byte, error) {
	return defaultModule.MarshalMercuryTransit(info, options)
}

// MarshalVenusTransit 将金星凌日的全球可见分区、一触/四触视地平线与可选站心接触等时线编码为 GeoJSON。
// MarshalVenusTransit encodes the global visibility regions, C1/C4 apparent horizons and optional topocentric contact isochrones of a Venus transit.
func MarshalVenusTransit(info venus.TransitInfo, options PlanetTransitOptions) ([]byte, error) {
	return defaultModule.MarshalVenusTransit(info, options)
}
