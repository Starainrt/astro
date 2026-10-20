package basic

import "math"

const (
	planetTransitEarthRadiusKM            = 6378.137
	planetTransitEarthEccentricitySquared = 0.00669437999014
)

func planetTransitShadowGeometry(tt float64, cfg planetTransitConfig, inst *Instance) (float64, float64, Observer, bool) {
	sunRA, sunDec := HSunApparentRaDecNWithInstance(tt, -1, inst)
	planetRA, planetDec := cfg.apparentRaDecN(tt, -1)
	sunDistance := EarthAwayNWithInstance(tt, -1, inst) * angularDiameterAstronomicalUnitKM
	planetDistance := cfg.earthDistanceN(tt, -1) * angularDiameterAstronomicalUnitKM
	sun := eclipticCartesian(sunRA, sunDec, sunDistance)
	body := eclipticCartesian(planetRA, planetDec, planetDistance)
	axis, length := pn(Vector3{body[0] - sun[0], body[1] - sun[1], body[2] - sun[2]})
	projection := -(body[0]*axis[0] + body[1]*axis[1] + body[2]*axis[2])
	impact := Vector3{body[0] + projection*axis[0], body[1] + projection*axis[1], body[2] + projection*axis[2]}
	_, distance := pn(impact)
	planetRadius := venusEquatorialRadiusKM
	if cfg.planetIndex == 1 {
		planetRadius = mercuryEquatorialRadiusKM
	}
	sine := (sunEquatorialRadiusKM + planetRadius) / length
	cosine := math.Sqrt(1 - sine*sine)
	penumbra := (planetRadius + projection*sine) / cosine / planetTransitEarthRadiusKM
	gamma := math.Copysign(distance/planetTransitEarthRadiusKM, impact[2])
	sidereal := ApparentSiderealTime(TT2UT1(tt)) * 15
	seed := Observer{
		Longitude: normalizeLongitude180(math.Atan2(impact[1], impact[0])*180/math.Pi - sidereal),
		Latitude:  math.Atan2(impact[2], math.Hypot(impact[0], impact[1])*(1-planetTransitEarthEccentricitySquared)) * 180 / math.Pi,
	}
	possible := isFiniteFloat(gamma) && isFiniteFloat(penumbra) && projection > 0 && planetDistance < sunDistance && math.Abs(gamma) <= penumbra+1/cosine
	return gamma, penumbra, seed, possible
}

func planetTransitSurfaceOverlap(transit PlanetTransitResult, seed Observer, inst *Instance) bool {
	snapshot, ok := buildPlanetTransitSnapshot(transit, inst)
	if !ok {
		return false
	}
	best := math.Inf(1)
	for step := 4.0; step > 0.0001; step /= 2 {
		center := seed
		for _, latitudeStep := range []float64{-step, 0, step} {
			for _, longitudeStep := range []float64{-step, 0, step} {
				observer := Observer{Longitude: normalizeLongitude180(center.Longitude + longitudeStep), Latitude: math.Max(-90, math.Min(90, center.Latitude+latitudeStep))}
				tt, valid := snapshot.localGreatestTT(UTC2TT(transit.Greatest), observer)
				state, stateOK := snapshot.stateAt(tt, observer)
				if !valid || !stateOK {
					continue
				}
				if state.external < 0 {
					local := snapshot.local(transit, observer)
					if local.Valid && local.Visible {
						return true
					}
				}
				if state.external < best {
					best, seed = state.external, observer
				}
			}
		}
	}
	return false
}
