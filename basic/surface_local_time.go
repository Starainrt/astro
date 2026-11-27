package basic

import (
	"math"

	"github.com/starainrt/astro/tools"
)

// SurfaceLocalTime 行星表面上的恒星时与太阳时，逆行天体恒星时递减 / sidereal and solar times on a planetary reference layer, with decreasing retrograde sidereal time.
type SurfaceLocalTime struct {
	// PrimeMeridianAngle 本初子午线角 W，度 [0,360) / prime meridian angle W in degrees [0,360).
	PrimeMeridianAngle float64
	// SiderealRotationAngle 地点子午线的升交点零点赤经，度 [0,360) / local meridian right ascension from the node in degrees [0,360).
	SiderealRotationAngle float64
	// LocalSiderealTime 地点子午线的春分点零点赤经，度 [0,360) / local meridian right ascension from the equinox in degrees [0,360).
	LocalSiderealTime float64
	// LocalTrueSolarTime 真太阳时，本地小时 [0,24)，水星近日点附近可倒退 / true solar time in local hours [0,24), possibly reversing near Mercury perihelion.
	LocalTrueSolarTime float64
	// LocalMeanSolarTime 平太阳时，本地小时 [0,24) / mean solar time in local hours [0,24).
	LocalMeanSolarTime float64
	// SolarDayLength 平太阳日长度，地球日 / mean solar day length in Earth days.
	SolarDayLength float64
	// SolarLongitude 自春分点沿轨道运动量到几何太阳的经度，度 [0,360) / geometric solar longitude from the equinox along orbital motion in degrees [0,360).
	SolarLongitude float64
	// EquinoxDefined 春分点是否良定义，为 false 时 LocalSiderealTime 与 SolarLongitude 为 0 且不可用 / whether the equinox is defined; otherwise LocalSiderealTime and SolarLongitude are zero and unavailable.
	EquinoxDefined bool
	// OutsideValidity 超出本库认定的姿态有效区间 / outside the library's adopted orientation validity interval.
	OutsideValidity bool
}

func surfaceHours(hours float64) float64 {
	return tools.Limit360(hours*15) / 15
}

func surfaceMeanSolarCoordinates(position, velocity Vector3, frame bodyEquatorialFrame) (meanRA, longitude float64, err error) {
	const mu = 1.32712440018e11 * 86400 * 86400
	radius := vectorMagnitude(position)
	momentum := pxp(position, velocity)
	vh := pxp(velocity, momentum)
	var eccentricity, radial Vector3
	for i := range radial {
		radial[i] = position[i] / radius
		eccentricity[i] = vh[i]/mu - radial[i]
	}
	e := vectorMagnitude(eccentricity)
	if !isFiniteFloat(e) || e >= 1 {
		return 0, 0, ErrInvalidObservationInput
	}
	equationOfCenter := 0.0
	if e > 0 {
		perihelion := normalizeVector(eccentricity)
		nuRad := math.Atan2(vectorDot(pxp(perihelion, radial), frame.normal), vectorDot(perihelion, radial))
		eccentricAnomalyRad := 2 * math.Atan(math.Sqrt((1-e)/(1+e))*math.Tan(nuRad/2))
		meanAnomalyRad := eccentricAnomalyRad - e*math.Sin(eccentricAnomalyRad)
		equationOfCenter = math.Remainder((nuRad-meanAnomalyRad)*180/math.Pi, 360)
	}
	origin, theta := frame.equinox, frame.equinoxAngle
	if !frame.equinoxDefined {
		projection := vectorDot(frame.node, frame.normal)
		for i := range origin {
			origin[i] = frame.node[i] - projection*frame.normal[i]
		}
		origin = normalizeVector(origin)
		theta = math.Atan2(vectorDot(pxp(frame.node, origin), frame.pole), vectorDot(frame.node, origin)) * 180 / math.Pi
	}
	sun := Vector3{-radial[0], -radial[1], -radial[2]}
	longitude = math.Atan2(vectorDot(pxp(origin, sun), frame.normal), vectorDot(origin, sun)) * 180 / math.Pi
	meanRA = tools.Limit360(theta + longitude - equationOfCenter)
	longitude = tools.Limit360(longitude)
	if !isFiniteFloat(meanRA) || !isFiniteFloat(longitude) {
		return 0, 0, ErrInvalidObservationInput
	}
	return meanRA, longitude, nil
}

// SurfaceLocalTimeAtWithInstance 行星表面的本地时，时刻为 TT 儒略日 / local times on a planetary surface at a TT Julian day.
func SurfaceLocalTimeAtWithInstance(observer BodyObserver, jdTDB float64, inst *Instance) (SurfaceLocalTime, error) {
	sun, err := surfaceDirectionAt(SkySun, observer, jdTDB, inst)
	if err != nil {
		return SurfaceLocalTime{}, err
	}
	position, velocity, err := bodyHeliocentricState(observer.Body, jdTDB, inst)
	if err != nil {
		return SurfaceLocalTime{}, err
	}
	orientation := sun.frame.orientation
	frame, err := bodyEquatorialFrameFromState(orientation, position, velocity)
	if err != nil {
		return SurfaceLocalTime{}, err
	}
	meanRA, longitude, err := surfaceMeanSolarCoordinates(position, velocity, frame)
	if err != nil {
		return SurfaceLocalTime{}, err
	}
	sidereal := tools.Limit360(orientation.PrimeMeridianAngle + tools.Limit360(observer.EastLongitude))
	apparentRA := math.Atan2(vectorDot(sun.direction, frame.east), vectorDot(sun.direction, frame.node)) * 180 / math.Pi
	sign := math.Copysign(1, orientation.RotationRate)
	result := SurfaceLocalTime{
		PrimeMeridianAngle:    orientation.PrimeMeridianAngle,
		SiderealRotationAngle: sidereal,
		LocalTrueSolarTime:    surfaceHours(12 + sign*(sidereal-apparentRA)/15),
		LocalMeanSolarTime:    surfaceHours(12 + sign*(sidereal-meanRA)/15),
		SolarDayLength:        1 / math.Abs(orientation.RotationRate/360-1/orbitalPeriodDays[observer.Body]),
		EquinoxDefined:        frame.equinoxDefined,
		OutsideValidity:       orientation.OutsideValidity,
	}
	if frame.equinoxDefined {
		result.LocalSiderealTime = tools.Limit360(sidereal - frame.equinoxAngle)
		result.SolarLongitude = longitude
	}
	return result, nil
}
