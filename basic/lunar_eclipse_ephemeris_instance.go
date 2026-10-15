package basic

import ()

// LunarEclipseWithInstance 指定星历实例的月食（TT，Danjon），计算失败时食甚为 NaN / lunar eclipse with an explicit instance; TT maximum is NaN on failure.
func LunarEclipseWithInstance(seedJDE float64, inst *Instance) LunarEclipseResult {
	return LunarEclipseDanjonWithInstance(seedJDE, inst)
}

// LunarEclipseDanjonWithInstance 指定星历实例的 Danjon 月食 / Danjon lunar eclipse with an explicit ephemeris instance.
func LunarEclipseDanjonWithInstance(seedJDE float64, inst *Instance) LunarEclipseResult {
	return lunarEclipseWithInstance(seedJDE, lunarEclipseShadowDanjon, inst)
}

// LunarEclipseChauvenetWithInstance 指定星历实例的 Chauvenet 月食 / Chauvenet lunar eclipse with an explicit ephemeris instance.
func LunarEclipseChauvenetWithInstance(seedJDE float64, inst *Instance) LunarEclipseResult {
	return lunarEclipseWithInstance(seedJDE, lunarEclipseShadowChauvenet, inst)
}

func lunarEclipseWithInstance(seedJDE float64, shadowModel lunarEclipseShadowModel, inst *Instance) LunarEclipseResult {
	if inst == nil {
		return lunarEclipse(seedJDE, shadowModel)
	}
	fullMoonJDE := CalcMoonSHByJDEWithInstance(seedJDE, 1, inst)
	return lunarEclipseFromFullMoon(fullMoonJDE, shadowModel, func(jde float64) lunarShadowState {
		return computeLunarShadowStateWithInstance(jde, shadowModel, inst)
	})
}

// LunarEclipseShadowGeometryAtWithInstance 指定星历实例的 Danjon 地影几何 / Danjon shadow geometry with an explicit ephemeris instance.
func LunarEclipseShadowGeometryAtWithInstance(maximumJDE float64, inst *Instance) LunarEclipseShadowGeometry {
	return lunarEclipseShadowGeometryAtWithInstance(maximumJDE, lunarEclipseShadowDanjon, inst)
}

// LunarEclipseShadowGeometryChauvenetAtWithInstance 指定星历实例的 Chauvenet 地影几何 / Chauvenet shadow geometry with an explicit ephemeris instance.
func LunarEclipseShadowGeometryChauvenetAtWithInstance(maximumJDE float64, inst *Instance) LunarEclipseShadowGeometry {
	return lunarEclipseShadowGeometryAtWithInstance(maximumJDE, lunarEclipseShadowChauvenet, inst)
}

// LunarEclipseShadowGeometryAtModelWithInstance 指定星历实例及影半径模型的地影几何 / shadow geometry with an explicit instance and shadow-radius model.
func LunarEclipseShadowGeometryAtModelWithInstance(maximumJDE float64, model LunarEclipseShadowModel, inst *Instance) LunarEclipseShadowGeometry {
	if model == LunarEclipseShadowModelChauvenet {
		return LunarEclipseShadowGeometryChauvenetAtWithInstance(maximumJDE, inst)
	}
	return LunarEclipseShadowGeometryAtWithInstance(maximumJDE, inst)
}

// LunarEclipseDiagramWithInstance 指定星历实例的 Danjon 月食穿影图 / Danjon lunar eclipse diagram with an explicit ephemeris instance.
func LunarEclipseDiagramWithInstance(seedJDE float64, options LunarEclipseDiagramOptions, inst *Instance) LunarEclipseDiagramResult {
	return LunarEclipseDiagramDanjonWithInstance(seedJDE, options, inst)
}

// LunarEclipseDiagramDanjonWithInstance 指定星历实例的 Danjon 月食穿影图 / Danjon lunar eclipse diagram with an explicit ephemeris instance.
func LunarEclipseDiagramDanjonWithInstance(seedJDE float64, options LunarEclipseDiagramOptions, inst *Instance) LunarEclipseDiagramResult {
	return lunarEclipseDiagramWithInstance(seedJDE, lunarEclipseShadowDanjon, options, inst)
}

// LunarEclipseDiagramChauvenetWithInstance 指定星历实例的 Chauvenet 月食穿影图 / Chauvenet lunar eclipse diagram with an explicit ephemeris instance.
func LunarEclipseDiagramChauvenetWithInstance(seedJDE float64, options LunarEclipseDiagramOptions, inst *Instance) LunarEclipseDiagramResult {
	return lunarEclipseDiagramWithInstance(seedJDE, lunarEclipseShadowChauvenet, options, inst)
}
