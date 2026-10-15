package basic

import (
	"math"

	"github.com/starainrt/astro/internal/geodata"
)

// LunarEclipseShadowGeometry 是食甚时刻的地影几何。
// 注意单位口径：Gamma 用地球赤道半径，而两个影半径用度——后者是 NASA 月食图上 P./U. Radius 的口径，
// 换成地球赤道半径要乘以月球处的地球视差（弧度）。
// LunarEclipseShadowGeometry is the terrestrial-shadow geometry at maximum eclipse. Gamma is in Earth
// equatorial radii while the two shadow radii are in degrees, matching the P./U. Radius convention of
// NASA lunar-eclipse charts; multiply a radius by the Earth's parallax at the Moon to get Earth radii.
type LunarEclipseShadowGeometry struct {
	// Gamma 是月心到地影轴的最小距离，单位地球赤道半径。
	// Gamma is the least distance from the Moon's centre to the shadow axis, in Earth equatorial radii.
	Gamma float64
	// PenumbralRadiusDegrees 与 UmbralRadiusDegrees 是半影、本影在地影轴垂直面上的角半径，单位度。
	// PenumbralRadiusDegrees and UmbralRadiusDegrees are the penumbral and umbral angular radii on the
	// plane perpendicular to the shadow axis, in degrees.
	PenumbralRadiusDegrees float64
	UmbralRadiusDegrees    float64
	// MoonDistanceEarthRadii 是食甚时的地心月距，单位地球赤道半径。
	// MoonDistanceEarthRadii is the geocentric lunar distance at greatest eclipse, in Earth equatorial radii.
	MoonDistanceEarthRadii float64
	// AxisDegrees 是食甚时月心到地影轴的角距，即 NASA 月食图上那列 Axis。
	// 它与 Gamma 是同一个量的两种刻度：Gamma 除以月球处的地球视差就是它。
	// AxisDegrees is the angular distance from the Moon's centre to the shadow axis at greatest eclipse,
	// the column NASA lunar-eclipse charts print as Axis. It is the same quantity as Gamma on a different
	// scale: divide Gamma by the Earth's parallax at the Moon.
	AxisDegrees float64
}

// LunarEclipseShadowGeometryAt 用 Danjon 影半径模型计算食甚时刻的地影几何。
// LunarEclipseShadowGeometryAt computes the shadow geometry at maximum eclipse with the Danjon shadow model.
func LunarEclipseShadowGeometryAt(maximumJDE float64) LunarEclipseShadowGeometry {
	return lunarEclipseShadowGeometryAt(maximumJDE, lunarEclipseShadowDanjon)
}

// LunarEclipseShadowGeometryChauvenetAt 用 Chauvenet 影半径模型计算食甚时刻的地影几何。
// LunarEclipseShadowGeometryChauvenetAt computes the shadow geometry with the Chauvenet shadow model.
func LunarEclipseShadowGeometryChauvenetAt(maximumJDE float64) LunarEclipseShadowGeometry {
	return lunarEclipseShadowGeometryAt(maximumJDE, lunarEclipseShadowChauvenet)
}

// LunarEclipseShadowGeometryAtModel 按结果里记录的影半径模型取地影几何。
// LunarEclipseShadowGeometryAtModel picks the shadow geometry by the model recorded in a result.
func LunarEclipseShadowGeometryAtModel(maximumJDE float64, model LunarEclipseShadowModel) LunarEclipseShadowGeometry {
	if model == LunarEclipseShadowModelChauvenet {
		return LunarEclipseShadowGeometryChauvenetAt(maximumJDE)
	}
	return LunarEclipseShadowGeometryAt(maximumJDE)
}

func lunarEclipseShadowGeometryAt(maximumJDE float64, shadowModel lunarEclipseShadowModel) LunarEclipseShadowGeometry {
	return lunarEclipseShadowGeometryAtWithInstance(maximumJDE, shadowModel, nil)
}

func lunarEclipseShadowGeometryAtWithInstance(maximumJDE float64, shadowModel lunarEclipseShadowModel, inst *Instance) LunarEclipseShadowGeometry {
	evaluate := func(jde float64) lunarShadowState {
		return computeLunarShadowStateWithInstance(jde, shadowModel, inst)
	}
	state := evaluate(maximumJDE)
	moonDistanceKM := HMoonAwayWithInstance(maximumJDE, inst)
	// 影半径是以地心为顶点的角量，除以月球处的地球视差就换成地球赤道半径。
	earthParallax := geodata.EarthEquatorialRadiusKM / moonDistanceKM
	_, _, _, _, minimumDistance := refineLunarEclipseMaximumWithEvaluator(maximumJDE, evaluate)
	if earthParallax <= 0 {
		return LunarEclipseShadowGeometry{}
	}
	return LunarEclipseShadowGeometry{
		Gamma:                  minimumDistance / earthParallax,
		PenumbralRadiusDegrees: state.penumbraRadiusRad * 180 / math.Pi,
		UmbralRadiusDegrees:    state.umbraRadiusRad * 180 / math.Pi,
		MoonDistanceEarthRadii: moonDistanceKM / geodata.EarthEquatorialRadiusKM,
		// 平面 x 向东（黄经差）、y 向北（黄纬和），方位角自北向东量。
		AxisDegrees: minimumDistance * 180 / math.Pi,
	}
}
