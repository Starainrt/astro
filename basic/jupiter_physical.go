package basic

import (
	"math"

	"github.com/starainrt/astro/planet"
	. "github.com/starainrt/astro/tools"
)

type jupiterPhysicalObservationInfo struct {
	DS       float64
	DE       float64
	SystemI  float64
	SystemII float64
}

// JupiterCentralMeridianInfo 木星中央经线 / Jupiter central meridians.
type JupiterCentralMeridianInfo struct {
	// SystemI 木星 System I 照亮盘中央经线，单位度，西经为正。
	SystemI float64
	// SystemII 木星 System II 照亮盘中央经线，单位度，西经为正。
	SystemII float64
	// SystemIII 木星 System III 盘面中央经线，单位度，西经为正。
	SystemIII float64
}

// JupiterCentralMeridians 木星 System I/II/III 中央经线 / Jupiter System I/II/III central meridians.
func JupiterCentralMeridians(jd float64) JupiterCentralMeridianInfo {
	return JupiterCentralMeridiansN(jd, -1)
}

// JupiterCentralMeridiansN 木星 System I/II/III 中央经线（截断版） / truncated Jupiter System I/II/III central meridians.
func JupiterCentralMeridiansN(jd float64, n int) JupiterCentralMeridianInfo {
	observations := jupiterPhysicalObservationsN(jd, n)
	physical := JupiterPhysicalN(jd, n)
	return JupiterCentralMeridianInfo{
		SystemI:   observations.SystemI,
		SystemII:  observations.SystemII,
		SystemIII: physical.SubEarthLongitude,
	}
}

// JupiterDSDE 木星 DS/DE 行星中心赤纬 / Jupiter planetocentric declinations of Sun and Earth.
func JupiterDSDE(jd float64) (ds, de float64) {
	return JupiterDSDEN(jd, -1)
}

// JupiterDSDEN 木星 DS/DE 行星中心赤纬（截断版） / truncated Jupiter planetocentric declinations of Sun and Earth.
func JupiterDSDEN(jd float64, n int) (ds, de float64) {
	observations := jupiterPhysicalObservationsN(jd, n)
	return observations.DS, observations.DE
}

func jupiterPhysicalObservationsN(jd float64, n int) jupiterPhysicalObservationInfo {
	days := jd - 2433282.5
	julianCentury := days / 36525.0

	poleRA := (268.0 + 0.1061*julianCentury) * rad
	poleDec := (64.5 - 0.0164*julianCentury) * rad
	w1 := (17.71 + 877.90003539*days) * rad
	w2 := (16.838 + 870.27003539*days) * rad

	earthLon := planet.WherePlanetN(-1, 0, jd, n)
	earthLat := planet.WherePlanetN(-1, 1, jd, n)
	earthRadius := planet.WherePlanetN(-1, 2, jd, n)

	delta := 4.0
	var jupiterLon float64
	var jupiterLat float64
	var jupiterRadius float64
	var x float64
	var y float64
	var z float64
	for i := 0; i < 2; i++ {
		lightTimeDays := astronomicalUnitLightTimeDays * delta
		jupiterLon = planet.WherePlanetN(4, 0, jd-lightTimeDays, n)
		jupiterLat = planet.WherePlanetN(4, 1, jd-lightTimeDays, n)
		jupiterRadius = planet.WherePlanetN(4, 2, jd-lightTimeDays, n)
		x = jupiterRadius*Cos(jupiterLat)*Cos(jupiterLon) - earthRadius*Cos(earthLat)*Cos(earthLon)
		y = jupiterRadius*Cos(jupiterLat)*Sin(jupiterLon) - earthRadius*Cos(earthLat)*Sin(earthLon)
		z = jupiterRadius*Sin(jupiterLat) - earthRadius*Sin(earthLat)
		delta = math.Sqrt(x*x + y*y + z*z)
	}

	meanObliquity := EclipticObliquity(jd, false)
	sinMeanObliquity, cosMeanObliquity := math.Sincos(meanObliquity)
	sinJupiterLat, cosJupiterLat := math.Sincos(jupiterLat * rad)
	sinJupiterLon, cosJupiterLon := math.Sincos(jupiterLon * rad)
	alphaSun := math.Atan2(cosMeanObliquity*sinJupiterLon-sinMeanObliquity*sinJupiterLat/cosJupiterLat, cosJupiterLon)
	deltaSun := math.Asin(cosMeanObliquity*sinJupiterLat + sinMeanObliquity*cosJupiterLat*sinJupiterLon)
	u := y*Cos(meanObliquity) - z*Sin(meanObliquity)
	v := y*Sin(meanObliquity) + z*Cos(meanObliquity)
	alpha := math.Atan2(u, x)
	deltaEarth := math.Atan2(v, math.Hypot(x, u))

	sinPoleDec, cosPoleDec := math.Sincos(poleDec)
	sinDeltaSun, cosDeltaSun := math.Sincos(deltaSun)
	ds := math.Asin(-sinPoleDec*sinDeltaSun-cosPoleDec*cosDeltaSun*math.Cos(poleRA-alphaSun)) * deg
	sinDeltaEarth, cosDeltaEarth := math.Sincos(deltaEarth)
	sinPoleDeltaRA, cosPoleDeltaRA := math.Sincos(poleRA - alpha)
	zeta := math.Atan2(
		sinPoleDec*cosDeltaEarth*cosPoleDeltaRA-sinDeltaEarth*cosPoleDec,
		cosDeltaEarth*sinPoleDeltaRA,
	)
	de := math.Asin(-sinPoleDec*sinDeltaEarth-cosPoleDec*cosDeltaEarth*math.Cos(poleRA-alpha)) * deg

	systemI := w1 - zeta - 5.07033*rad*delta
	systemII := w2 - zeta - 5.02626*rad*delta
	phaseCorrection := (2*jupiterRadius*delta + earthRadius*earthRadius - jupiterRadius*jupiterRadius - delta*delta) / (4 * jupiterRadius * delta)
	if Sin(jupiterLon-earthLon) < 0 {
		phaseCorrection = -phaseCorrection
	}

	return jupiterPhysicalObservationInfo{
		DS:       ds,
		DE:       de,
		SystemI:  Limit360((systemI + phaseCorrection) * deg),
		SystemII: Limit360((systemII + phaseCorrection) * deg),
	}
}
