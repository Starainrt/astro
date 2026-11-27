package sun

import (
	"time"

	"github.com/starainrt/astro/basic"
)

func (module *Sun) nextOnMars(after time.Time, eastLongitude, latitude float64, kind basic.SurfaceEventKind) (time.Time, error) {
	observer := basic.BodyObserver{Body: basic.SkyMars, EastLongitude: eastLongitude, Latitude: latitude}
	event, err := basic.NextSurfaceEventWithInstance(basic.SkySun, observer, kind, basic.UTC2TT(basic.Date2JD(after.UTC())), basic.SurfaceOptions{}, module.inst)
	if err != nil {
		return time.Time{}, err
	}
	return basic.JD2DateByZone(basic.TT2UTC(event.JDE), after.Location(), false), nil
}

// NextRiseOnMars 火星表面下一次太阳上缘升起，包含起点 / next solar upper-limb rise on Mars including the starting instant.
func (module *Sun) NextRiseOnMars(after time.Time, eastLongitude, latitude float64) (time.Time, error) {
	return module.nextOnMars(after, eastLongitude, latitude, basic.SurfaceRise)
}

// NextSetOnMars 火星表面下一次太阳上缘落下，包含起点 / next solar upper-limb set on Mars including the starting instant.
func (module *Sun) NextSetOnMars(after time.Time, eastLongitude, latitude float64) (time.Time, error) {
	return module.nextOnMars(after, eastLongitude, latitude, basic.SurfaceSet)
}
