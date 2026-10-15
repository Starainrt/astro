package basic

import (
	"errors"
	"math"

	"github.com/starainrt/astro/internal/ephem"
	"github.com/starainrt/astro/planet"
)

func instanceValue(instances []*Instance) *Instance {
	if len(instances) == 0 {
		return nil
	}
	return instances[0]
}

func instanceModel(inst *Instance) ephem.Model {
	if inst == nil {
		return ephem.Active()
	}
	return inst.Model()
}

// PlanetCoordinateWithInstance 指定星历实例的行星日心经纬或距离 / heliocentric planetary coordinate or distance with an explicit instance.
func PlanetCoordinateWithInstance(index, coordinate int, jde float64, inst *Instance) float64 {
	return PlanetCoordinateNWithInstance(index, coordinate, jde, -1, inst)
}

// PlanetCoordinateNWithInstance 指定星历实例的行星日心经纬或距离（n 变体） / n-parameterised heliocentric planetary coordinate or distance with an explicit instance.
func PlanetCoordinateNWithInstance(index, coordinate int, jde float64, n int, inst *Instance) float64 {
	if inst == nil {
		return planet.WherePlanetN(index, coordinate, jde, n)
	}
	if inst.Model() != ephem.Custom {
		return planet.WherePlanetModelN(inst.Model(), index, coordinate, jde, n)
	}
	value, ok := planetHeliocentricCoordinateWithInstance(index, coordinate, jde, inst)
	if !ok {
		return math.NaN()
	}
	return value
}

func boundVectorApparent(target ephem.Body, jde float64, n int, inst *Instance) (lo, bo, tau, distAU float64, ok bool) {
	lo, bo, tau, distAU, ok, err := vectorApparentWithInstance(target, jde, n, inst)
	if err != nil {
		return math.NaN(), math.NaN(), math.NaN(), math.NaN(), true
	}
	return lo, bo, tau, distAU, ok
}

func instanceStateEcliptic(target ephem.Body, jde float64, inst *Instance) (lo, bo, distAU float64, ok bool) {
	return instanceStateEclipticRelative(target, ephem.Earth, jde, jde, inst)
}

func instanceStateEclipticRelative(target, center ephem.Body, stateJDE, eclipticJDE float64, inst *Instance) (lo, bo, distAU float64, ok bool) {
	// 几何状态捷径只给自定义引擎，内置模型必须走与包级相同的级数与视位置管线。
	if inst == nil || inst.Model() != ephem.Custom {
		return 0, 0, 0, false
	}
	state, _, err := inst.State(target, center, stateJDE)
	if err != nil {
		return 0, 0, 0, false
	}
	distanceKM := math.Sqrt(state[0]*state[0] + state[1]*state[1] + state[2]*state[2])
	if distanceKM == 0 || math.IsNaN(distanceKM) || math.IsInf(distanceKM, 0) {
		return 0, 0, 0, false
	}
	direction := [3]float64{state[0] / distanceKM, state[1] / distanceKM, state[2] / distanceKM}
	lo, bo = icrfToMeanEcliptic(direction, eclipticJDE)
	return lo, bo, distanceKM / astronomicalUnitKM, true
}

func icrfToMeanEcliptic(direction [3]float64, jde float64) (lo, bo float64) {
	matrix := ltpPBMAT(2000 + (jde-2451545)/365.25)
	var equatorial [3]float64
	for axis := range equatorial {
		equatorial[axis] = matrix[axis][0]*direction[0] + matrix[axis][1]*direction[1] + matrix[axis][2]*direction[2]
	}
	sine, cosine := math.Sincos(Obliquity1980(jde) * math.Pi / 180)
	eclipticY := equatorial[1]*cosine + equatorial[2]*sine
	eclipticZ := -equatorial[1]*sine + equatorial[2]*cosine
	lo = math.Atan2(eclipticY, equatorial[0]) * 180 / math.Pi
	bo = math.Atan2(eclipticZ, math.Hypot(equatorial[0], eclipticY)) * 180 / math.Pi
	return math.Mod(lo+720, 360), bo
}

func instanceApparent(target ephem.Body, jde float64, n int, inst *Instance) (lo, bo, tau, distAU float64, ok bool, err error) {
	if inst == nil || n >= 0 && inst.Model() != ephem.Custom {
		return 0, 0, 0, 0, false, nil
	}
	return apparentState(target, jde, inst)
}

// apparentState 求视位置并校验；引擎不支持时 ok=false 且 err 为 nil，状态非法时返回 err。
func apparentState(target ephem.Body, jde float64, inst *Instance) (lo, bo, tau, distAU float64, ok bool, err error) {
	direction, distanceKM, tau, err := inst.Apparent(target, jde)
	if errors.Is(err, ephem.ErrUnsupportedApparent) || inst.Model() != ephem.Custom && errors.Is(err, ephem.ErrNoEngine) {
		return 0, 0, 0, 0, false, nil
	}
	if err != nil {
		return 0, 0, 0, 0, false, err
	}
	norm := math.Hypot(math.Hypot(direction[0], direction[1]), direction[2])
	if !isFiniteFloat(distanceKM) || distanceKM <= 0 || !isFiniteFloat(norm) || norm <= 0 || !isFiniteFloat(tau) || tau < 0 {
		return 0, 0, 0, 0, false, errors.New("ephem: invalid apparent state")
	}
	lo, bo = icrfToTrueEcliptic(direction, jde)
	// 天文单位须与 vectorApparent 一致，否则实例与包级距离不逐位相等。
	return lo, bo, tau, distanceKM / 149597870.7, true, nil
}

func vectorApparentWithInstance(target ephem.Body, jde float64, n int, inst *Instance) (lo, bo, tau, distAU float64, ok bool, err error) {
	if inst != nil {
		return instanceApparent(target, jde, n, inst)
	}
	lo, bo, tau, distAU, ok = vectorApparent(target, jde, n)
	return lo, bo, tau, distAU, ok, nil
}
