package jupiter

import (
	"math"
	"time"

	"github.com/starainrt/astro/basic"
)

// RotationSystem 木星旋转系统 / Jovian rotation system.
type RotationSystem = basic.JupiterRotationSystem

const (
	// SystemI 木星赤道旋转系统 / Jovian equatorial rotation system.
	SystemI = basic.JupiterSystemI
	// SystemII 木星非赤道旋转系统 / Jovian nonequatorial rotation system.
	SystemII = basic.JupiterSystemII
	// SystemIII 木星固连旋转系统 / Jovian body-fixed rotation system.
	SystemIII = basic.JupiterSystemIII
)

// RotationAngle 木星 ICRF 本初子午线旋转角，输入无效时为 NaN / Jovian ICRF prime meridian rotation angle, NaN for invalid input.
func (module *Jupiter) RotationAngle(date time.Time, system RotationSystem) float64 {
	angle, err := basic.JupiterRotationAngleWithInstance(basic.UTC2TT(basic.Date2JD(date.UTC())), system, module.inst)
	if err != nil {
		return math.NaN()
	}
	return angle
}
