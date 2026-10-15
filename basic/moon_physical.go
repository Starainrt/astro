package basic

import (
	"math"

	. "github.com/starainrt/astro/tools"
)

const moonPhysicalInclinationDeg = 1.54242
const moonPhysicalAstronomicalUnitKM = 149597870.7

// MoonPhysicalInfo 月球物理观测参数 / physical observing parameters of the Moon.
type MoonPhysicalInfo struct {
	// OpticalLongitude 光学经度天平动，单位度 / optical libration in longitude, degrees.
	OpticalLongitude float64
	// OpticalLatitude 光学纬度天平动，单位度 / optical libration in latitude, degrees.
	OpticalLatitude float64
	// PhysicalLongitude 物理经度天平动，单位度 / physical libration in longitude, degrees.
	PhysicalLongitude float64
	// PhysicalLatitude 物理纬度天平动，单位度 / physical libration in latitude, degrees.
	PhysicalLatitude float64
	// LibrationLongitude 总经度天平动，单位度 / total libration in longitude, degrees.
	LibrationLongitude float64
	// LibrationLatitude 总纬度天平动，单位度 / total libration in latitude, degrees.
	LibrationLatitude float64
	// PositionAngle 月球自转轴位置角，单位度 / position angle of the lunar rotation axis, degrees.
	PositionAngle float64
}

// MoonPhysicalWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func MoonPhysicalWithInstance(jde float64, inst *Instance) MoonPhysicalInfo {
	return MoonPhysicalNWithInstance(jde, -1, inst)
}

// MoonPhysicalNWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func MoonPhysicalNWithInstance(jde float64, n int, inst *Instance) MoonPhysicalInfo {
	lo, bo := moonApparentLoBoN(jde, n, inst)
	return moonPhysicalNFromCoordinates(jde, n, lo, bo, HMoonTrueRaNWithInstance(jde, n, inst))
}

// MoonTopocentricPhysical 月球站心物理观测参数 / topocentric physical observing parameters of the Moon.
func MoonTopocentricPhysical(jde, observerLon, observerLat, height float64) MoonPhysicalInfo {
	return MoonTopocentricPhysicalN(jde, observerLon, observerLat, height, -1)
}

// MoonTopocentricPhysicalNWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func MoonTopocentricPhysicalNWithInstance(jde, observerLon, observerLat, height float64, n int, inst *Instance) MoonPhysicalInfo {
	lambda, beta, alpha := moonTopocentricPhysicalCoordinatesN(jde, observerLon, observerLat, height, n, inst)
	return moonPhysicalNFromCoordinates(jde, n, lambda, beta, alpha)
}

func moonPhysicalNFromCoordinates(jde float64, n int, lambda, beta, alpha float64) MoonPhysicalInfo {
	t := (jde - 2451545.0) / 36525.0
	epsilon := TrueObliquity(jde)
	deltaPsi := Nutation2000Bi(jde)

	D := Limit360(SunMoonAngle(jde))
	sunMeanAnomaly := Limit360(SunM(jde))
	moonMeanAnomaly := Limit360(MoonM(jde))
	F := Limit360(MoonLonX(jde))
	omega := moonPhysicalMeanAscendingNode(t)
	E := 1 - 0.002516*t - 0.0000074*t*t
	K1 := 119.75 + 131.849*t
	K2 := 72.56 + 20.186*t

	W := Limit360(lambda - deltaPsi - omega)
	A := ArcTan2(Sin(W)*Cos(beta)*Cos(moonPhysicalInclinationDeg)-Sin(beta)*Sin(moonPhysicalInclinationDeg), Cos(W)*Cos(beta))
	opticalLongitude := wrapSignedAngle180(A - F)
	opticalLatitude := ArcSin(-Sin(W)*Cos(beta)*Sin(moonPhysicalInclinationDeg) - Sin(beta)*Cos(moonPhysicalInclinationDeg))

	rho, sigma, tau := moonPhysicalLibrationSeries(D, sunMeanAnomaly, moonMeanAnomaly, F, omega, E, K1, K2)
	physicalLongitude := -tau + (rho*Cos(A)+sigma*Sin(A))*Tan(opticalLatitude)
	physicalLatitude := sigma*Cos(A) - rho*Sin(A)

	librationLongitude := wrapSignedAngle180(opticalLongitude + physicalLongitude)
	librationLatitude := opticalLatitude + physicalLatitude

	V := Limit360(omega + deltaPsi + sigma/Sin(moonPhysicalInclinationDeg))
	X := Sin(moonPhysicalInclinationDeg+rho) * Sin(V)
	Y := Sin(moonPhysicalInclinationDeg+rho)*Cos(V)*Cos(epsilon) - Cos(moonPhysicalInclinationDeg+rho)*Sin(epsilon)
	littleOmega := ArcTan2(X, Y)
	positionAngle := ArcSin(clampUnit((sqrtXY(X, Y) * Cos(alpha-littleOmega)) / Cos(librationLatitude)))

	return MoonPhysicalInfo{
		OpticalLongitude:   opticalLongitude,
		OpticalLatitude:    opticalLatitude,
		PhysicalLongitude:  physicalLongitude,
		PhysicalLatitude:   physicalLatitude,
		LibrationLongitude: librationLongitude,
		LibrationLatitude:  librationLatitude,
		PositionAngle:      positionAngle,
	}
}

func moonTopocentricPhysicalCoordinatesN(jde, observerLon, observerLat, height float64, n int, instances ...*Instance) (lambda, beta, alpha float64) {
	inst := instanceValue(instances)

	geocentricRA := HMoonTrueRaNWithInstance(jde, n, inst)
	geocentricDec := HMoonTrueDecNWithInstance(jde, n, inst)
	distanceAU := HMoonAwayNWithInstance(jde, n, inst) / moonPhysicalAstronomicalUnitKM
	utcJD := TT2UTC(jde)

	var topocentricDec float64
	alpha, topocentricDec = TopocentricRaDec(geocentricRA, geocentricDec, observerLat, observerLon, utcJD, distanceAU, height)
	lambda, beta = RaDecToLoBo(jde, alpha, topocentricDec)
	return
}

func moonPhysicalLibrationSeries(D, M, MP, F, omega, E, K1, K2 float64) (rho, sigma, tau float64) {
	rho = -0.02752*Cos(MP) -
		0.02245*Sin(F) +
		0.00684*Cos(MP-2*F) -
		0.00293*Cos(2*F) -
		0.00085*Cos(2*F-2*D) -
		0.00054*Cos(MP-2*D) -
		0.00020*Sin(MP+F) -
		0.00020*Cos(MP+2*F) -
		0.00020*Cos(MP-F) +
		0.00014*Cos(MP+2*F-2*D)

	sigma = -0.02816*Sin(MP) +
		0.02244*Cos(F) -
		0.00682*Sin(MP-2*F) -
		0.00279*Sin(2*F) -
		0.00083*Sin(2*F-2*D) +
		0.00069*Sin(MP-2*D) +
		0.00040*Cos(MP+F) -
		0.00025*Sin(2*MP) -
		0.00023*Sin(MP+2*F) +
		0.00020*Cos(MP-F) +
		0.00019*Sin(MP-F) +
		0.00013*Sin(MP+2*F-2*D) -
		0.00010*Cos(MP-3*F)

	tau = 0.02520*E*Sin(M) +
		0.00473*Sin(2*MP-2*F) -
		0.00467*Sin(MP) +
		0.00396*Sin(K1) +
		0.00276*Sin(2*MP-2*D) +
		0.00196*Sin(omega) -
		0.00183*Cos(MP-F) +
		0.00115*Sin(MP-2*D) -
		0.00096*Sin(MP-D) +
		0.00046*Sin(2*F-2*D) -
		0.00039*Sin(MP-F) -
		0.00032*Sin(MP-M-D) +
		0.00027*Sin(2*MP-M-2*D) +
		0.00023*Sin(K2) -
		0.00014*Sin(2*D) +
		0.00014*Cos(2*MP-2*F) -
		0.00012*Sin(MP-2*F) -
		0.00012*Sin(2*MP) +
		0.00011*Sin(2*MP-2*M-2*D)
	return
}

func moonPhysicalMeanAscendingNode(t float64) float64 {
	return Limit360(125.04452222222222 - 1934.136261111111*t + 0.0020708333333333334*t*t + 0.0000022222222222222222*t*t*t)
}

func wrapSignedAngle180(angle float64) float64 {
	angle = Limit360(angle)
	if angle > 180 {
		angle -= 360
	}
	return angle
}

func sqrtXY(x, y float64) float64 {
	return math.Sqrt(x*x + y*y)
}
