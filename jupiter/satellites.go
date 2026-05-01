package jupiter

import (
	"time"

	"github.com/starainrt/astro/basic"
)

// GalileanSatellitePosition 木星伽利略卫星视位置 / apparent position of a Galilean satellite.
//
// Jovicentric* 为木心 J2000 平赤道直角坐标与速度，单位 AU / AU-day。
// ApparentRA/Dec 为该卫星的地心天体测量赤经赤纬，单位度。
// OffsetX/OffsetY 以木星中心为原点，分别沿天球东、北方向为正。
// Jovicentric* are Jovicentric J2000 mean-equatorial coordinates and velocities in AU / AU-day.
// ApparentRA/Dec are geocentric astrometric right ascension and declination in degrees.
// OffsetX/OffsetY are measured from Jupiter's center, positive to celestial east and north.
type GalileanSatellitePosition struct {
	JovicentricX  float64
	JovicentricY  float64
	JovicentricZ  float64
	JovicentricVX float64
	JovicentricVY float64
	JovicentricVZ float64

	ApparentRA       float64
	ApparentDec      float64
	EarthDistance    float64
	OffsetXArcsec    float64
	OffsetYArcsec    float64
	OffsetXJupiterR  float64
	OffsetYJupiterR  float64
	OffsetZJupiterR  float64
	InFrontOfJupiter bool
}

// GalileanSatellitesInfo 四颗伽利略卫星视位置 / apparent positions of the four Galilean satellites.
type GalileanSatellitesInfo struct {
	Io       GalileanSatellitePosition
	Europa   GalileanSatellitePosition
	Ganymede GalileanSatellitePosition
	Callisto GalileanSatellitePosition
}

// Satellites 木星四颗伽利略卫星视位置 / apparent positions of Jupiter's four Galilean satellites.
//
// date 表示观测绝对时刻；内部使用该时刻对应的 TT/TDB 历元做 L1 星历求值。
// date is the observing instant; internally the corresponding TT/TDB epoch is used for the L1 ephemeris evaluation.
func Satellites(date time.Time) GalileanSatellitesInfo {
	jde := basic.Date2JDE(date.UTC())
	observations := basic.JupiterGalileanSatelliteObservations(jde)
	return GalileanSatellitesInfo{
		Io:       galileanSatellitePositionFromBasic(observations[0]),
		Europa:   galileanSatellitePositionFromBasic(observations[1]),
		Ganymede: galileanSatellitePositionFromBasic(observations[2]),
		Callisto: galileanSatellitePositionFromBasic(observations[3]),
	}
}

func galileanSatellitePositionFromBasic(observation basic.JupiterGalileanObservation) GalileanSatellitePosition {
	return GalileanSatellitePosition{
		JovicentricX:  observation.State.X,
		JovicentricY:  observation.State.Y,
		JovicentricZ:  observation.State.Z,
		JovicentricVX: observation.State.VX,
		JovicentricVY: observation.State.VY,
		JovicentricVZ: observation.State.VZ,

		ApparentRA:       observation.RA,
		ApparentDec:      observation.Dec,
		EarthDistance:    observation.Distance,
		OffsetXArcsec:    observation.OffsetXArcsec,
		OffsetYArcsec:    observation.OffsetYArcsec,
		OffsetXJupiterR:  observation.OffsetXJupiterRadii,
		OffsetYJupiterR:  observation.OffsetYJupiterRadii,
		OffsetZJupiterR:  observation.OffsetZJupiterRadii,
		InFrontOfJupiter: observation.InFrontOfJupiter,
	}
}
