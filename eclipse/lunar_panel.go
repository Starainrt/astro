package eclipse

import (
	"time"

	"github.com/starainrt/astro/basic"
)

// LunarEclipseGeocentricPanel 汇总月食图上各面板所需的食甚时刻地心量。
// LunarEclipseGeocentricPanel carries the geocentric quantities that lunar-eclipse panels print.
type LunarEclipseGeocentricPanel struct {
	// FullMoon 望，即地心视黄经相差 180° 的时刻；它与该次月食的食甚不是同一时刻。
	// FullMoon is the instant of full moon, when the geocentric apparent ecliptic longitudes differ by
	// 180 degrees; it is not the same instant as greatest eclipse.
	FullMoon    time.Time
	FullMoonJDE float64
	// DeltaTSeconds 是食甚时刻实际使用的 ΔT。
	// DeltaTSeconds is the ΔT used at greatest eclipse.
	DeltaTSeconds float64

	SunRightAscensionDeg  float64
	SunDeclinationDeg     float64
	SunSemidiameterArcsec float64
	SunParallaxArcsec     float64

	MoonRightAscensionDeg  float64
	MoonDeclinationDeg     float64
	MoonSemidiameterArcsec float64
	MoonParallaxArcsec     float64

	// ShadowModel 是影半径模型，决定影半径与食分。
	// ShadowModel is the shadow-radius model that fixes the radii and magnitudes.
	ShadowModel LunarEclipseShadowModel
	// PenumbralRadiusDegrees 与 UmbralRadiusDegrees 是食甚时的半影、本影半径，单位度。
	// PenumbralRadiusDegrees and UmbralRadiusDegrees are the shadow radii at greatest eclipse, in degrees.
	PenumbralRadiusDegrees float64
	UmbralRadiusDegrees    float64
	// Gamma 是月心到地影轴的最小距离，单位地球赤道半径；AxisDegrees 是同一个量的角度制数值。
	// Gamma is the minimum Moon-centre-to-axis distance in Earth equatorial radii; AxisDegrees is the same in degrees.
	Gamma       float64
	AxisDegrees float64
	// MoonDistanceEarthRadii 是食甚时的地心月距。
	// MoonDistanceEarthRadii is the geocentric Moon distance at greatest eclipse.
	MoonDistanceEarthRadii float64
	// PenumbralMagnitude 与 UmbralMagnitude 与 LunarEclipseInfo 同口径。
	// PenumbralMagnitude and UmbralMagnitude follow the same convention as LunarEclipseInfo.
	PenumbralMagnitude float64
	UmbralMagnitude    float64
	// Ephemeris 是所用影半径模型名称。
	// Ephemeris names the shadow-radius model used.
	Ephemeris string
}

// LunarEclipseGeocentricPanelAt 计算给定日期月食的地心量面板，使用 Danjon 影半径模型 / geocentric lunar eclipse panel with Danjon shadow radii.
func (module *Eclipse) LunarEclipseGeocentricPanelAt(date time.Time) (LunarEclipseGeocentricPanel, bool) {
	return lunarEclipseGeocentricPanelAt(date, func(seedJDE float64) basic.LunarEclipseResult {
		return module.lunarResult(seedJDE, false)
	}, module)
}

// LunarEclipseGeocentricPanelChauvenet 计算给定日期月食的地心量面板，使用 Chauvenet 影半径模型 / geocentric lunar eclipse panel with Chauvenet shadow radii.
func (module *Eclipse) LunarEclipseGeocentricPanelChauvenet(date time.Time) (LunarEclipseGeocentricPanel, bool) {
	return lunarEclipseGeocentricPanelAt(date, func(seedJDE float64) basic.LunarEclipseResult {
		return module.lunarResult(seedJDE, true)
	}, module)
}

func lunarEclipseGeocentricPanelAt(date time.Time, calculator func(float64) basic.LunarEclipseResult, modules ...*Eclipse) (LunarEclipseGeocentricPanel, bool) {
	module := &defaultModule
	if len(modules) > 0 {
		module = modules[0]
	}

	seed := timeToTTJDE(date)
	fullMoonJDE := basic.CalcMoonSHByJDEWithInstance(seed, 1, module.inst)
	result := calculator(fullMoonJDE)
	if result.Type == basic.LunarEclipseNone {
		return LunarEclipseGeocentricPanel{}, false
	}
	geometry := basic.LunarEclipseShadowGeometryAtModelWithInstance(result.Maximum, result.ShadowModel, module.inst)
	ephemeris := "Danjon"
	if result.ShadowModel == basic.LunarEclipseShadowModelChauvenet {
		ephemeris = "Chauvenet"
	}
	panel := LunarEclipseGeocentricPanel{
		FullMoon:               ttJDEToTime(fullMoonJDE, date.Location()),
		FullMoonJDE:            fullMoonJDE,
		DeltaTSeconds:          basic.DeltaT(result.Maximum, true),
		ShadowModel:            LunarEclipseShadowModel(result.ShadowModel),
		PenumbralRadiusDegrees: geometry.PenumbralRadiusDegrees,
		UmbralRadiusDegrees:    geometry.UmbralRadiusDegrees,
		Gamma:                  geometry.Gamma,
		AxisDegrees:            geometry.AxisDegrees,
		MoonDistanceEarthRadii: geometry.MoonDistanceEarthRadii,
		PenumbralMagnitude:     result.PenumbralMagnitude,
		UmbralMagnitude:        result.Magnitude,
		Ephemeris:              ephemeris,
	}
	panel.SunRightAscensionDeg, panel.SunDeclinationDeg = basic.SunApparentRaDecWithInstance(result.Maximum, module.inst)
	panel.MoonRightAscensionDeg, panel.MoonDeclinationDeg = basic.HMoonTrueRaDecWithInstance(result.Maximum, module.inst)
	panel.SunSemidiameterArcsec = basic.SolarEclipseSunSemidiameterWithInstance(result.Maximum, basic.SolarEclipseSunRadiusStandard, module.inst)
	panel.MoonSemidiameterArcsec = basic.MoonSemidiameterWithInstance(result.Maximum, module.inst)
	panel.SunParallaxArcsec = horizontalParallaxArcsec(panel.SunSemidiameterArcsec, horizontalParallaxSunRatio)
	panel.MoonParallaxArcsec = horizontalParallaxArcsec(panel.MoonSemidiameterArcsec, horizontalParallaxMoonRatio)
	return panel, true
}
