package saturn

import (
	"time"

	"github.com/starainrt/astro/basic"
	"github.com/starainrt/astro/calendar"
)

// PhaseAngle 相位角，单位度 / phase angle in degrees.
func (module *Saturn) PhaseAngle(date time.Time) float64 {
	return module.PhaseAngleN(date, -1)
}

// PhaseAngleN 相位角（n 变体），单位度 / n-parameterised phase angle in degrees.
func (module *Saturn) PhaseAngleN(date time.Time, n int) float64 {
	return basic.SaturnPhaseAngleNWithInstance(phaseJD(date), n, module.inst)
}

// IlluminatedFraction 被照亮比例 / illuminated fraction.
func (module *Saturn) IlluminatedFraction(date time.Time) float64 {
	return module.IlluminatedFractionN(date, -1)
}

// IlluminatedFractionN 被照亮比例（n 变体） / n-parameterised illuminated fraction.
func (module *Saturn) IlluminatedFractionN(date time.Time, n int) float64 {
	return basic.SaturnIlluminatedFractionNWithInstance(phaseJD(date), n, module.inst)
}

// Phase 相位，被照亮比例 / phase, illuminated fraction.
func (module *Saturn) Phase(date time.Time) float64 {
	return module.IlluminatedFraction(date)
}

// PhaseN 相位（n 变体），被照亮比例 / n-parameterised phase, illuminated fraction.
func (module *Saturn) PhaseN(date time.Time, n int) float64 {
	return module.IlluminatedFractionN(date, n)
}

// BrightLimbPositionAngle 亮面中心位置角，单位度 / bright limb position angle in degrees.
func (module *Saturn) BrightLimbPositionAngle(date time.Time) float64 {
	return module.BrightLimbPositionAngleN(date, -1)
}

// BrightLimbPositionAngleN 亮面中心位置角（n 变体），单位度 / n-parameterised bright limb position angle in degrees.
func (module *Saturn) BrightLimbPositionAngleN(date time.Time, n int) float64 {
	return basic.SaturnBrightLimbPositionAngleNWithInstance(phaseJD(date), n, module.inst)
}

func phaseJD(date time.Time) float64 {
	return basic.UTC2TT(calendar.Date2JD(date.UTC()))
}
