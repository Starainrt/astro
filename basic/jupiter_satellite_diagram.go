package basic

import "math"

// JupiterGalileanDiagram 地球视点盘面快照，距离单位为木星赤道半径 / Earth-view snapshot in Jupiter equatorial radii.
type JupiterGalileanDiagram struct {
	Observations         [4]JupiterGalileanObservation
	Phenomena            [4]JupiterGalileanPhenomenon
	RadiiJupiterR        [4]float64
	AngularRadiiJupiterR [4]float64
	ShadowUmbraRadii     [4]float64
	MinorRadiusRatio     float64
	PolePositionAngle    float64
}

// JupiterGalileanMutualDiagram 被遮星相对遮挡星的投影，距离单位为木星赤道半径 / passive-relative-to-active projection in Jupiter equatorial radii.
type JupiterGalileanMutualDiagram struct {
	Active, Passive                                    int
	GeometryJDE                                        float64
	XJupiterR, YJupiterR                               float64
	ActiveRadiusJupiterR, PassiveRadiusJupiterR        float64
	UmbraRadiusJupiterR, PenumbraRadiusJupiterR        float64
	MinimumSeparationArcsec, ImpactParameter, Coverage float64
	Annular                                            bool
}

// JupiterGalileanDiagramWithInstance 接收 TT 时刻的系统快照，位置角自天北向东 / system snapshot at reception TT, position angle north through east.
func JupiterGalileanDiagramWithInstance(jde float64, inst *Instance) (JupiterGalileanDiagram, bool) {
	if !isFinite(jde) {
		return JupiterGalileanDiagram{}, false
	}
	context := newJupiterGalileanObservationContext(jde, inst)
	if !isFinite(context.jupiterDistance) || context.jupiterDistance <= 0 {
		return JupiterGalileanDiagram{}, false
	}
	result := JupiterGalileanDiagram{
		MinorRadiusRatio:  context.earthMinorRadius,
		PolePositionAngle: math.Atan2(vectorDot(context.bodyZ, context.east), vectorDot(context.bodyZ, context.north)) * deg,
	}
	for index := range result.Observations {
		result.Observations[index] = context.observationForSatellite(index)
		result.Phenomena[index] = context.phenomenonForSatellite(index)
		result.RadiiJupiterR[index] = jupiterGalileanSatelliteRadiusJupiterRadii(index + 1)
		result.AngularRadiiJupiterR[index] = math.Atan2(result.RadiiJupiterR[index]*jupiterGalileanEquatorialRadiusKM/astronomicalUnitKM, result.Observations[index].Distance) / math.Atan2(jupiterGalileanEquatorialRadiusKM/astronomicalUnitKM, context.jupiterDistance)
		if !isFinite(result.Observations[index].OffsetXJupiterRadii) || !isFinite(result.Observations[index].OffsetYJupiterRadii) {
			return JupiterGalileanDiagram{}, false
		}
		if !result.Phenomena[index].ShadowTransit {
			result.Phenomena[index].ShadowOffsetXArcsec = 0
			result.Phenomena[index].ShadowOffsetYArcsec = 0
			result.Phenomena[index].ShadowOffsetXJupiterRadii = 0
			result.Phenomena[index].ShadowOffsetYJupiterRadii = 0
		} else {
			state := result.Observations[index].State
			result.ShadowUmbraRadii[index], _ = context.shadowSection(Vector3{state.X, state.Y, state.Z}, index+1)
		}
	}
	return result, true
}

// JupiterGalileanMutualDiagramWithInstance 接收 TT 的互掩地球视平面或互食太阳视平面快照，X 东 Y 北 / Earth-plane occultation or Sun-plane eclipse at reception TT, X east and Y north.
func JupiterGalileanMutualDiagramWithInstance(jde float64, first, second int, kind JupiterGalileanMutualPhenomenonType, inst *Instance) (JupiterGalileanMutualDiagram, bool) {
	first, second, ok := normalizeJupiterGalileanMutualPair(first, second)
	if !ok || !isFinite(jde) {
		return JupiterGalileanMutualDiagram{}, false
	}
	if kind == JupiterGalileanMutualOccultation {
		geometry, valid := jupiterGalileanMutualGeometryAt(jde, first, second, inst)
		if !valid {
			return JupiterGalileanMutualDiagram{}, false
		}
		context := newJupiterGalileanObservationContext(jde, inst)
		semidiameter := math.Atan2(jupiterGalileanEquatorialRadiusKM/astronomicalUnitKM, context.jupiterDistance) * deg * 3600
		return JupiterGalileanMutualDiagram{
			Active: geometry.Active, Passive: geometry.Passive, GeometryJDE: jde,
			XJupiterR: geometry.PassiveX - geometry.ActiveX, YJupiterR: geometry.PassiveY - geometry.ActiveY,
			ActiveRadiusJupiterR:    jupiterGalileanSatelliteRadiusJupiterRadii(geometry.Active),
			PassiveRadiusJupiterR:   jupiterGalileanSatelliteRadiusJupiterRadii(geometry.Passive),
			MinimumSeparationArcsec: geometry.Separation * semidiameter,
			ImpactParameter:         geometry.ImpactParameter, Coverage: geometry.Coverage, Annular: geometry.Annular,
		}, true
	}
	if kind != JupiterGalileanMutualEclipse {
		return JupiterGalileanMutualDiagram{}, false
	}
	for _, passive := range []int{first, second} {
		emission := jupiterGalileanEclipseEmission(jde, passive, inst)
		geometry, valid := jupiterGalileanMutualEclipseGeometryAt(emission, first, second, inst)
		if !valid || geometry.Passive != passive {
			continue
		}
		frame := newJupiterGalileanEclipseFrame(emission, inst)
		activeVector := frame.satellite(emission, geometry.Active)
		passiveVector := frame.satellite(emission, passive)
		axis := normalizeVector(activeVector)
		ra, dec := vectorToRaDec(axis)
		east := Vector3{-math.Sin(ra * rad), math.Cos(ra * rad), 0}
		north := Vector3{-math.Cos(ra*rad) * math.Sin(dec*rad), -math.Sin(ra*rad) * math.Sin(dec*rad), math.Cos(dec * rad)}
		delta := Vector3{passiveVector[0] - activeVector[0], passiveVector[1] - activeVector[1], passiveVector[2] - activeVector[2]}
		scale := astronomicalUnitKM / jupiterGalileanEquatorialRadiusKM
		return JupiterGalileanMutualDiagram{
			Active: geometry.Active, Passive: passive, GeometryJDE: emission,
			XJupiterR: vectorDot(delta, east) * scale, YJupiterR: vectorDot(delta, north) * scale,
			ActiveRadiusJupiterR:    jupiterGalileanSatelliteRadiusJupiterRadii(geometry.Active),
			PassiveRadiusJupiterR:   geometry.PassiveRadiusKM / jupiterGalileanEquatorialRadiusKM,
			UmbraRadiusJupiterR:     geometry.UmbraKM / jupiterGalileanEquatorialRadiusKM,
			PenumbraRadiusJupiterR:  geometry.PenumbraKM / jupiterGalileanEquatorialRadiusKM,
			MinimumSeparationArcsec: geometry.MinimumSeparationArcsec,
			ImpactParameter:         geometry.ImpactParameter, Coverage: geometry.Coverage, Annular: geometry.Annular,
		}, true
	}
	return JupiterGalileanMutualDiagram{}, false
}
