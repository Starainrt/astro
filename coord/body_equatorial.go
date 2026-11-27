package coord

import (
	"time"

	"github.com/starainrt/astro/basic"
)

// BodyEquatorOrigin 行星赤道坐标的赤经零点 / right ascension origin of planetary equatorial coordinates.
type BodyEquatorOrigin = basic.BodyEquatorOrigin

const (
	// BodyEquatorOriginIAUNode 行星赤道与 ICRF 赤道的升交点 / ascending node of the planetary equator on the ICRF equator.
	BodyEquatorOriginIAUNode = basic.BodyEquatorOriginIAUNode
	// BodyEquatorOriginEquinox 行星赤道与瞬时轨道面的春分点 / equinox of the planetary equator and instantaneous orbital plane.
	BodyEquatorOriginEquinox = basic.BodyEquatorOriginEquinox
)

// ToBodyEquatorial ICRF/J2000 赤道方向转行星赤道坐标，不区分框架偏差 / ICRF/J2000 direction to planetary equatorial coordinates, ignoring frame bias.
func ToBodyEquatorial(body basic.SkyBody, date time.Time, ra, dec float64, origin BodyEquatorOrigin) (Equatorial, error) {
	bodyRA, bodyDec, err := basic.ToBodyEquatorial(body, basic.UTC2TT(basic.Date2JD(date.UTC())), ra, dec, origin)
	return Equatorial{RA: bodyRA, Dec: bodyDec}, err
}

// FromBodyEquatorial 行星赤道方向转 ICRF/J2000 坐标，不区分框架偏差 / planetary equatorial direction to ICRF/J2000 coordinates, ignoring frame bias.
func FromBodyEquatorial(body basic.SkyBody, date time.Time, bodyRA, bodyDec float64, origin BodyEquatorOrigin) (Equatorial, error) {
	ra, dec, err := basic.FromBodyEquatorial(body, basic.UTC2TT(basic.Date2JD(date.UTC())), bodyRA, bodyDec, origin)
	return Equatorial{RA: ra, Dec: dec}, err
}
