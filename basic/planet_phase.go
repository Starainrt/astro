package basic

import (
	"github.com/starainrt/astro/planet"
	. "github.com/starainrt/astro/tools"
)

type planetRaDecNFunc func(jd float64, n int) (float64, float64)

// MercuryPhaseAngle 水星相位角 / phase angle of Mercury.
func MercuryPhaseAngle(jd float64) float64 {
	return MercuryPhaseAngleN(jd, -1)
}

// MercuryPhaseAngleN 水星相位角（截断版） / truncated phase angle of Mercury.
func MercuryPhaseAngleN(jd float64, n int) float64 {
	return planetPhaseAngleN(1, jd, n)
}

// MercuryIlluminatedFraction 水星被照亮比例 / illuminated fraction of Mercury.
func MercuryIlluminatedFraction(jd float64) float64 {
	return MercuryIlluminatedFractionN(jd, -1)
}

// MercuryIlluminatedFractionN 水星被照亮比例（截断版） / truncated illuminated fraction of Mercury.
func MercuryIlluminatedFractionN(jd float64, n int) float64 {
	return planetIlluminatedFractionN(1, jd, n)
}

// MercuryBrightLimbPositionAngle 水星亮面中心位置角 / position angle of Mercury bright limb.
func MercuryBrightLimbPositionAngle(jd float64) float64 {
	return MercuryBrightLimbPositionAngleN(jd, -1)
}

// MercuryBrightLimbPositionAngleN 水星亮面中心位置角（截断版） / truncated position angle of Mercury bright limb.
func MercuryBrightLimbPositionAngleN(jd float64, n int) float64 {
	return planetBrightLimbPositionAngleN(jd, n, MercuryApparentRaDecN)
}

// VenusPhaseAngle 金星相位角 / phase angle of Venus.
func VenusPhaseAngle(jd float64) float64 {
	return VenusPhaseAngleN(jd, -1)
}

// VenusPhaseAngleN 金星相位角（截断版） / truncated phase angle of Venus.
func VenusPhaseAngleN(jd float64, n int) float64 {
	return planetPhaseAngleN(2, jd, n)
}

// VenusIlluminatedFraction 金星被照亮比例 / illuminated fraction of Venus.
func VenusIlluminatedFraction(jd float64) float64 {
	return VenusIlluminatedFractionN(jd, -1)
}

// VenusIlluminatedFractionN 金星被照亮比例（截断版） / truncated illuminated fraction of Venus.
func VenusIlluminatedFractionN(jd float64, n int) float64 {
	return planetIlluminatedFractionN(2, jd, n)
}

// VenusBrightLimbPositionAngle 金星亮面中心位置角 / position angle of Venus bright limb.
func VenusBrightLimbPositionAngle(jd float64) float64 {
	return VenusBrightLimbPositionAngleN(jd, -1)
}

// VenusBrightLimbPositionAngleN 金星亮面中心位置角（截断版） / truncated position angle of Venus bright limb.
func VenusBrightLimbPositionAngleN(jd float64, n int) float64 {
	return planetBrightLimbPositionAngleN(jd, n, VenusApparentRaDecN)
}

// MarsPhaseAngle 火星相位角 / phase angle of Mars.
func MarsPhaseAngle(jd float64) float64 {
	return MarsPhaseAngleN(jd, -1)
}

// MarsPhaseAngleN 火星相位角（截断版） / truncated phase angle of Mars.
func MarsPhaseAngleN(jd float64, n int) float64 {
	return planetPhaseAngleN(3, jd, n)
}

// MarsIlluminatedFraction 火星被照亮比例 / illuminated fraction of Mars.
func MarsIlluminatedFraction(jd float64) float64 {
	return MarsIlluminatedFractionN(jd, -1)
}

// MarsIlluminatedFractionN 火星被照亮比例（截断版） / truncated illuminated fraction of Mars.
func MarsIlluminatedFractionN(jd float64, n int) float64 {
	return planetIlluminatedFractionN(3, jd, n)
}

// MarsBrightLimbPositionAngle 火星亮面中心位置角 / position angle of Mars bright limb.
func MarsBrightLimbPositionAngle(jd float64) float64 {
	return MarsBrightLimbPositionAngleN(jd, -1)
}

// MarsBrightLimbPositionAngleN 火星亮面中心位置角（截断版） / truncated position angle of Mars bright limb.
func MarsBrightLimbPositionAngleN(jd float64, n int) float64 {
	return planetBrightLimbPositionAngleN(jd, n, MarsApparentRaDecN)
}

// JupiterPhaseAngle 木星相位角 / phase angle of Jupiter.
func JupiterPhaseAngle(jd float64) float64 {
	return JupiterPhaseAngleN(jd, -1)
}

// JupiterPhaseAngleN 木星相位角（截断版） / truncated phase angle of Jupiter.
func JupiterPhaseAngleN(jd float64, n int) float64 {
	return planetPhaseAngleN(4, jd, n)
}

// JupiterIlluminatedFraction 木星被照亮比例 / illuminated fraction of Jupiter.
func JupiterIlluminatedFraction(jd float64) float64 {
	return JupiterIlluminatedFractionN(jd, -1)
}

// JupiterIlluminatedFractionN 木星被照亮比例（截断版） / truncated illuminated fraction of Jupiter.
func JupiterIlluminatedFractionN(jd float64, n int) float64 {
	return planetIlluminatedFractionN(4, jd, n)
}

// JupiterBrightLimbPositionAngle 木星亮面中心位置角 / position angle of Jupiter bright limb.
func JupiterBrightLimbPositionAngle(jd float64) float64 {
	return JupiterBrightLimbPositionAngleN(jd, -1)
}

// JupiterBrightLimbPositionAngleN 木星亮面中心位置角（截断版） / truncated position angle of Jupiter bright limb.
func JupiterBrightLimbPositionAngleN(jd float64, n int) float64 {
	return planetBrightLimbPositionAngleN(jd, n, JupiterApparentRaDecN)
}

// SaturnPhaseAngle 土星相位角 / phase angle of Saturn.
func SaturnPhaseAngle(jd float64) float64 {
	return SaturnPhaseAngleN(jd, -1)
}

// SaturnPhaseAngleN 土星相位角（截断版） / truncated phase angle of Saturn.
func SaturnPhaseAngleN(jd float64, n int) float64 {
	return planetPhaseAngleN(5, jd, n)
}

// SaturnIlluminatedFraction 土星被照亮比例 / illuminated fraction of Saturn.
func SaturnIlluminatedFraction(jd float64) float64 {
	return SaturnIlluminatedFractionN(jd, -1)
}

// SaturnIlluminatedFractionN 土星被照亮比例（截断版） / truncated illuminated fraction of Saturn.
func SaturnIlluminatedFractionN(jd float64, n int) float64 {
	return planetIlluminatedFractionN(5, jd, n)
}

// SaturnBrightLimbPositionAngle 土星亮面中心位置角 / position angle of Saturn bright limb.
func SaturnBrightLimbPositionAngle(jd float64) float64 {
	return SaturnBrightLimbPositionAngleN(jd, -1)
}

// SaturnBrightLimbPositionAngleN 土星亮面中心位置角（截断版） / truncated position angle of Saturn bright limb.
func SaturnBrightLimbPositionAngleN(jd float64, n int) float64 {
	return planetBrightLimbPositionAngleN(jd, n, SaturnApparentRaDecN)
}

// UranusPhaseAngle 天王星相位角 / phase angle of Uranus.
func UranusPhaseAngle(jd float64) float64 {
	return UranusPhaseAngleN(jd, -1)
}

// UranusPhaseAngleN 天王星相位角（截断版） / truncated phase angle of Uranus.
func UranusPhaseAngleN(jd float64, n int) float64 {
	return planetPhaseAngleN(6, jd, n)
}

// UranusIlluminatedFraction 天王星被照亮比例 / illuminated fraction of Uranus.
func UranusIlluminatedFraction(jd float64) float64 {
	return UranusIlluminatedFractionN(jd, -1)
}

// UranusIlluminatedFractionN 天王星被照亮比例（截断版） / truncated illuminated fraction of Uranus.
func UranusIlluminatedFractionN(jd float64, n int) float64 {
	return planetIlluminatedFractionN(6, jd, n)
}

// UranusBrightLimbPositionAngle 天王星亮面中心位置角 / position angle of Uranus bright limb.
func UranusBrightLimbPositionAngle(jd float64) float64 {
	return UranusBrightLimbPositionAngleN(jd, -1)
}

// UranusBrightLimbPositionAngleN 天王星亮面中心位置角（截断版） / truncated position angle of Uranus bright limb.
func UranusBrightLimbPositionAngleN(jd float64, n int) float64 {
	return planetBrightLimbPositionAngleN(jd, n, UranusApparentRaDecN)
}

// NeptunePhaseAngle 海王星相位角 / phase angle of Neptune.
func NeptunePhaseAngle(jd float64) float64 {
	return NeptunePhaseAngleN(jd, -1)
}

// NeptunePhaseAngleN 海王星相位角（截断版） / truncated phase angle of Neptune.
func NeptunePhaseAngleN(jd float64, n int) float64 {
	return planetPhaseAngleN(7, jd, n)
}

// NeptuneIlluminatedFraction 海王星被照亮比例 / illuminated fraction of Neptune.
func NeptuneIlluminatedFraction(jd float64) float64 {
	return NeptuneIlluminatedFractionN(jd, -1)
}

// NeptuneIlluminatedFractionN 海王星被照亮比例（截断版） / truncated illuminated fraction of Neptune.
func NeptuneIlluminatedFractionN(jd float64, n int) float64 {
	return planetIlluminatedFractionN(7, jd, n)
}

// NeptuneBrightLimbPositionAngle 海王星亮面中心位置角 / position angle of Neptune bright limb.
func NeptuneBrightLimbPositionAngle(jd float64) float64 {
	return NeptuneBrightLimbPositionAngleN(jd, -1)
}

// NeptuneBrightLimbPositionAngleN 海王星亮面中心位置角（截断版） / truncated position angle of Neptune bright limb.
func NeptuneBrightLimbPositionAngleN(jd float64, n int) float64 {
	return planetBrightLimbPositionAngleN(jd, n, NeptuneApparentRaDecN)
}

func planetPhaseAngleN(planetIndex int, jd float64, n int) float64 {
	return ArcCos(planetPhaseCosineN(planetIndex, jd, n))
}

func planetIlluminatedFractionN(planetIndex int, jd float64, n int) float64 {
	return (1 + planetPhaseCosineN(planetIndex, jd, n)) / 2
}

func planetPhaseCosineN(planetIndex int, jd float64, n int) float64 {
	planetSunDistance := planet.WherePlanetN(planetIndex, 2, jd, n)
	planetEarthDistance := planetEarthAwayN(planetIndex, jd, n)
	earthSunDistance := EarthAwayN(jd, n)
	cosine := (planetSunDistance*planetSunDistance + planetEarthDistance*planetEarthDistance - earthSunDistance*earthSunDistance) / (2 * planetSunDistance * planetEarthDistance)
	return clampUnit(cosine)
}

func planetBrightLimbPositionAngleN(jd float64, n int, apparentRaDec planetRaDecNFunc) float64 {
	sunRa, sunDec := HSunApparentRaDecN(jd, n)
	planetRa, planetDec := apparentRaDec(jd, n)
	y := Cos(sunDec) * Sin(sunRa-planetRa)
	x := Sin(sunDec)*Cos(planetDec) - Cos(sunDec)*Sin(planetDec)*Cos(sunRa-planetRa)
	return ArcTan2(y, x)
}

func clampUnit(value float64) float64 {
	if value > 1 {
		return 1
	}
	if value < -1 {
		return -1
	}
	return value
}
