package basic

import "github.com/starainrt/astro/tools"

// JupiterRotationSystem 木星大气旋转系统 / Jovian atmospheric rotation system.
type JupiterRotationSystem int

const (
	// JupiterSystemI 木星赤道旋转系统 / Jovian equatorial rotation system.
	JupiterSystemI JupiterRotationSystem = iota
	// JupiterSystemII 木星非赤道旋转系统 / Jovian nonequatorial rotation system.
	JupiterSystemII
	// JupiterSystemIII 木星固连旋转系统 / Jovian body-fixed rotation system.
	JupiterSystemIII
)

// JupiterRotationAngleWithInstance 木星指定旋转系统的 ICRF 本初子午线角，时刻为 TT 儒略日 / ICRF prime meridian angle of a Jovian rotation system at a TT Julian day.
func JupiterRotationAngleWithInstance(jdTT float64, system JupiterRotationSystem, _ *Instance) (float64, error) {
	if system < JupiterSystemI || system > JupiterSystemIII {
		return 0, ErrInvalidObservationInput
	}
	orientation, err := BodyOrientationAt(SkyJupiter, jdTT)
	if err != nil {
		return 0, err
	}
	angle := orientation.PrimeMeridianAngle
	switch system {
	case JupiterSystemI:
		angle += 142.153988083590 + (877.90-870.536)*(jdTT-2451545)
	case JupiterSystemII:
		angle += 118.407000999787 + (870.27-870.536)*(jdTT-2451545)
	}
	if !isFiniteFloat(angle) {
		return 0, ErrInvalidObservationInput
	}
	return tools.Limit360(angle), nil
}
