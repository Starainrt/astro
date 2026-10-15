package basic

import "math"

// lunarEclipseLimbMaxShiftDays 是月缘精修允许偏离圆月面接触的上限（10 分钟），超出视为不收敛并保留原值。
const lunarEclipseLimbMaxShiftDays = 600.0 / 86400.0

// lunarEclipseLimbStepDays 是割线法的初始步长（1 秒）。
const lunarEclipseLimbStepDays = 1.0 / 86400.0

// LunarEclipseWithLunarLimb 用月缘剖面精修月食接触时刻，食甚、食分与食型仍按圆月面 / refine lunar eclipse contacts with a lunar limb profile.
func LunarEclipseWithLunarLimb(result LunarEclipseResult, limb LunarLimb, inst *Instance) LunarEclipseResult {
	if limb == nil {
		return result
	}
	original := result
	shadowModel := lunarEclipseShadowDanjon
	if result.ShadowModel == LunarEclipseShadowModelChauvenet {
		shadowModel = lunarEclipseShadowChauvenet
	}
	evaluate := func(jde float64) lunarShadowState {
		return computeLunarShadowStateWithInstance(jde, shadowModel, inst)
	}
	refine := func(seed float64, kind lunarEclipseLimbContact) float64 {
		if !finite(seed) {
			return seed
		}
		return refineLunarEclipseLimbContact(seed, kind, limb, evaluate, inst)
	}
	if result.HasPenumbral {
		result.PenumbralStart = refine(result.PenumbralStart, lunarEclipseLimbPenumbral)
		result.PenumbralEnd = refine(result.PenumbralEnd, lunarEclipseLimbPenumbral)
	}
	if result.HasPartial {
		result.PartialStart = refine(result.PartialStart, lunarEclipseLimbUmbral)
		result.PartialEnd = refine(result.PartialEnd, lunarEclipseLimbUmbral)
	}
	if result.HasTotal {
		result.TotalStart = refine(result.TotalStart, lunarEclipseLimbTotal)
		result.TotalEnd = refine(result.TotalEnd, lunarEclipseLimbTotal)
	}
	result.LunarLimbApplied = lunarEclipseLimbChanged(result, original)
	return result
}

func lunarEclipseLimbChanged(result, original LunarEclipseResult) bool {
	for _, pair := range [][2]float64{
		{result.PenumbralStart, original.PenumbralStart},
		{result.PenumbralEnd, original.PenumbralEnd},
		{result.PartialStart, original.PartialStart},
		{result.PartialEnd, original.PartialEnd},
		{result.TotalStart, original.TotalStart},
		{result.TotalEnd, original.TotalEnd},
	} {
		if limbTimeChanged(pair[0], pair[1]) {
			return true
		}
	}
	return false
}

type lunarEclipseLimbContact int

const (
	lunarEclipseLimbPenumbral lunarEclipseLimbContact = iota
	lunarEclipseLimbUmbral
	lunarEclipseLimbTotal
)

// refineLunarEclipseLimbContact 以圆月面接触为初值用割线法求月缘间隙的零点，不收敛时返回初值。
func refineLunarEclipseLimbContact(seed float64, kind lunarEclipseLimbContact, limb LunarLimb, evaluate lunarShadowEvaluator, inst *Instance) float64 {
	t0, t1 := seed, seed+lunarEclipseLimbStepDays
	g0, ok0 := lunarEclipseLimbGap(t0, kind, limb, evaluate, inst)
	g1, ok1 := lunarEclipseLimbGap(t1, kind, limb, evaluate, inst)
	if !ok0 || !ok1 {
		return seed
	}
	for i := 0; i < 12; i++ {
		if g1 == g0 {
			return seed
		}
		t2 := t1 - g1*(t1-t0)/(g1-g0)
		if !finite(t2) || math.Abs(t2-seed) > lunarEclipseLimbMaxShiftDays {
			return seed
		}
		if math.Abs(t2-t1) < 1e-9 {
			return t2
		}
		g2, ok := lunarEclipseLimbGap(t2, kind, limb, evaluate, inst)
		if !ok {
			return seed
		}
		t0, g0, t1, g1 = t1, g1, t2, g2
	}
	return seed
}

// lunarEclipseLimbGap 是月缘版接触间隙（弧度），零点即接触。
func lunarEclipseLimbGap(jde float64, kind lunarEclipseLimbContact, limb LunarLimb, evaluate lunarShadowEvaluator, inst *Instance) (float64, bool) {
	state := evaluate(jde)
	moonRA, moonDec := HMoonGeocentricApparentRaDecWithInstance(jde, inst)
	sunRA, sunDec := HSunApparentRaDecWithInstance(jde, inst)
	moon := lunarEclipseLimbVector(moonRA, moonDec, HMoonAwayWithInstance(jde, inst))
	shadow := lunarEclipseLimbVector(sunRA+180, -sunDec, 1)

	if !finite(state.moonRadiusRad) || state.moonRadiusRad <= 0 {
		return 0, false
	}
	frame := newLunarLimbFrame(limb, jde, moon, inst)
	view, ok := frame.viewToward(moon, shadow)
	if !ok {
		return 0, false
	}
	separation := math.Hypot(state.x, state.y)
	contact := lunarLimbContact{
		frame:      &frame,
		view:       view,
		separation: separation,
		anglePerKM: state.moonRadiusRad / lunarMoonRadiusScale,
	}
	switch kind {
	case lunarEclipseLimbPenumbral:
		contact.targetRadius = state.penumbraRadiusRad
	default:
		contact.targetRadius = state.umbraRadiusRad
	}
	// 半影、本影外切取全方向最外的月缘；全食内切取全方向最靠外、最后进入本影的月缘。
	var limbRad float64
	if kind == lunarEclipseLimbTotal {
		limbRad, ok = contact.innerRadius(false)
	} else {
		limbRad, ok = contact.outerRadius()
	}
	if !ok {
		return 0, false
	}
	if kind == lunarEclipseLimbTotal {
		return separation - math.Abs(limbRad-contact.targetRadius), true
	}
	return separation - limbRad - contact.targetRadius, true
}

func lunarEclipseLimbVector(raDeg, decDeg, distance float64) [3]float64 {
	sinRA, cosRA := math.Sincos(raDeg * rad)
	sinDec, cosDec := math.Sincos(decDeg * rad)
	return [3]float64{distance * cosDec * cosRA, distance * cosDec * sinRA, distance * sinDec}
}
