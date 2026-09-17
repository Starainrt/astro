// Package svg 生成日食与月食的 SVG 图件。
//
// 可降级的图层用 data-source 标注实际几何来源，取值词表如下；没有降级分支的图层不带该属性。
//
//	partial-band-union           偏食可见域填充：权威边界线材与瞬时足迹求并集
//	sampled-footprint-sweep      偏食可见域填充：逐瞬时足迹扫掠（并集不可用时的降级）
//	partial-band-contours        偏食可见域边界线
//	rise-set-phase-lines         初亏、食甚、复圆的日升日落线
//	magnitude-contours           地方最大食分等值线
//	greatest-time-isochrones     食甚时刻等时线
//	besselian-critical-envelope  中心食带：Bessel 临界包络
//	paired-limit-chords          中心食带：南北限界弦带（包络不可用时的降级）
//	sampled-open-sweep           中心食带：采样扫掠（前两者都不可用时的降级）
//	central-path-limits          中心线与南北限界线
//	penumbral-outlines           瞬时半影轮廓
//	central-shadow-outlines      瞬时本影、反本影轮廓
//	p1-p4-visibility-regions     月食可见性分区
//	p1-p4-horizon-boundaries     月食 P1、P4 地平边界
//
// Package svg renders solar- and lunar-eclipse SVG charts. Degradable layers tag the geometry
// source they actually used in data-source, using exactly the values above.
package svg

// data-source 词表，见包注释。
const (
	eclipseMapSourcePartialBandUnion       = "partial-band-union"
	eclipseMapSourceSampledFootprintSweep  = "sampled-footprint-sweep"
	eclipseMapSourcePartialBandContours    = "partial-band-contours"
	eclipseMapSourceRiseSetPhaseLines      = "rise-set-phase-lines"
	eclipseMapSourceMagnitudeContours      = "magnitude-contours"
	eclipseMapSourceGreatestTimeIsochrones = "greatest-time-isochrones"
	eclipseMapSourceBesselianEnvelope      = "besselian-critical-envelope"
	eclipseMapSourcePairedLimitChords      = "paired-limit-chords"
	eclipseMapSourceSampledOpenSweep       = "sampled-open-sweep"
	eclipseMapSourceCentralPathLimits      = "central-path-limits"
	eclipseMapSourcePenumbralOutlines      = "penumbral-outlines"
	eclipseMapSourceCentralShadowOutlines  = "central-shadow-outlines"
	eclipseMapSourceLunarVisibilityRegions = "p1-p4-visibility-regions"
	eclipseMapSourceLunarHorizonBoundaries = "p1-p4-horizon-boundaries"
)

// eclipseMapDataSources 是全部合法取值，测试用它封闭词表。
var eclipseMapDataSources = []string{
	eclipseMapSourcePartialBandUnion,
	eclipseMapSourceSampledFootprintSweep,
	eclipseMapSourcePartialBandContours,
	eclipseMapSourceRiseSetPhaseLines,
	eclipseMapSourceMagnitudeContours,
	eclipseMapSourceGreatestTimeIsochrones,
	eclipseMapSourceBesselianEnvelope,
	eclipseMapSourcePairedLimitChords,
	eclipseMapSourceSampledOpenSweep,
	eclipseMapSourceCentralPathLimits,
	eclipseMapSourcePenumbralOutlines,
	eclipseMapSourceCentralShadowOutlines,
	eclipseMapSourceLunarVisibilityRegions,
	eclipseMapSourceLunarHorizonBoundaries,
}
