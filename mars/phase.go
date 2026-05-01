package mars

import (
	"time"

	"github.com/starainrt/astro/basic"
	"github.com/starainrt/astro/calendar"
)

// PhaseAngle 相位角，单位度 / phase angle in degrees.
func PhaseAngle(date time.Time) float64 {
	return PhaseAngleN(date, -1)
}

// PhaseAngleN 相位角（截断版），单位度 / truncated phase angle in degrees.
func PhaseAngleN(date time.Time, n int) float64 {
	return basic.MarsPhaseAngleN(phaseJD(date), n)
}

// IlluminatedFraction 被照亮比例 / illuminated fraction.
func IlluminatedFraction(date time.Time) float64 {
	return IlluminatedFractionN(date, -1)
}

// IlluminatedFractionN 被照亮比例（截断版） / truncated illuminated fraction.
func IlluminatedFractionN(date time.Time, n int) float64 {
	return basic.MarsIlluminatedFractionN(phaseJD(date), n)
}

// Phase 相位，被照亮比例 / phase, illuminated fraction.
func Phase(date time.Time) float64 {
	return IlluminatedFraction(date)
}

// PhaseN 相位（截断版），被照亮比例 / truncated phase, illuminated fraction.
func PhaseN(date time.Time, n int) float64 {
	return IlluminatedFractionN(date, n)
}

// BrightLimbPositionAngle 亮面中心位置角，单位度 / bright limb position angle in degrees.
func BrightLimbPositionAngle(date time.Time) float64 {
	return BrightLimbPositionAngleN(date, -1)
}

// BrightLimbPositionAngleN 亮面中心位置角（截断版），单位度 / truncated bright limb position angle in degrees.
func BrightLimbPositionAngleN(date time.Time, n int) float64 {
	return basic.MarsBrightLimbPositionAngleN(phaseJD(date), n)
}

func phaseJD(date time.Time) float64 {
	return basic.TD2UT(calendar.Date2JDE(date.UTC()), true)
}
