package jupiter

import (
	"time"

	"github.com/starainrt/astro/basic"
)

// GalileanSatellitePhenomenon 木星伽利略卫星瞬时现象 / instantaneous Galilean-satellite phenomena.
//
// Transit 为凌日，Occultation 为掩蔽，Eclipse 为进入木星本影，ShadowTransit 为卫星影子落在可见木星盘面上。
// ShadowOffset* 仅在 ShadowTransit 为 true 时有意义。
// Transit means a transit across Jupiter's disk, Occultation means hidden behind Jupiter, Eclipse means inside Jupiter's umbra, and ShadowTransit means the shadow falls on the visible Jovian disk.
// ShadowOffset* are meaningful only when ShadowTransit is true.
type GalileanSatellitePhenomenon struct {
	Transit       bool
	Occultation   bool
	Eclipse       bool
	ShadowTransit bool

	ShadowOffsetXArcsec float64
	ShadowOffsetYArcsec float64

	ShadowOffsetXJupiterR float64
	ShadowOffsetYJupiterR float64
}

// GalileanPhenomenaInfo 四颗伽利略卫星瞬时现象 / instantaneous phenomena of the four Galilean satellites.
type GalileanPhenomenaInfo struct {
	Io       GalileanSatellitePhenomenon
	Europa   GalileanSatellitePhenomenon
	Ganymede GalileanSatellitePhenomenon
	Callisto GalileanSatellitePhenomenon
}

// SatellitePhenomena 木星四颗伽利略卫星瞬时现象 / instantaneous phenomena of Jupiter's four Galilean satellites.
//
// date 表示观测绝对时刻；内部使用该时刻对应的 TT/TDB 历元求值。
// date is the observing instant; internally the corresponding TT/TDB epoch is used.
func SatellitePhenomena(date time.Time) GalileanPhenomenaInfo {
	jde := basic.Date2JDE(date.UTC())
	phenomena := basic.JupiterGalileanSatellitePhenomena(jde)
	return GalileanPhenomenaInfo{
		Io:       galileanPhenomenonFromBasic(phenomena[0]),
		Europa:   galileanPhenomenonFromBasic(phenomena[1]),
		Ganymede: galileanPhenomenonFromBasic(phenomena[2]),
		Callisto: galileanPhenomenonFromBasic(phenomena[3]),
	}
}

func galileanPhenomenonFromBasic(phenomenon basic.JupiterGalileanPhenomenon) GalileanSatellitePhenomenon {
	return GalileanSatellitePhenomenon{
		Transit:               phenomenon.Transit,
		Occultation:           phenomenon.Occultation,
		Eclipse:               phenomenon.Eclipse,
		ShadowTransit:         phenomenon.ShadowTransit,
		ShadowOffsetXArcsec:   phenomenon.ShadowOffsetXArcsec,
		ShadowOffsetYArcsec:   phenomenon.ShadowOffsetYArcsec,
		ShadowOffsetXJupiterR: phenomenon.ShadowOffsetXJupiterRadii,
		ShadowOffsetYJupiterR: phenomenon.ShadowOffsetYJupiterRadii,
	}
}
