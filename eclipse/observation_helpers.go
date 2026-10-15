package eclipse

import (
	"time"

	"github.com/starainrt/astro/basic"
	"github.com/starainrt/astro/tools"
)

func moonSunLoDiff(date time.Time) float64 {
	jde := basic.UTC2TT(basic.Date2JD(date.UTC()))
	sunLo := basic.HSunApparentLo(jde)
	moonLo := basic.HMoonApparentLo(jde)
	return tools.Limit360(moonLo - sunLo)
}

func solarAltitude(date time.Time, lon, lat float64, modules ...*Eclipse) float64 {
	module := &defaultModule
	if len(modules) > 0 {
		module = modules[0]
	}

	localJD := basic.Date2JD(date)
	_, loc := date.Zone()
	return basic.SunHeightWithInstance(localJD, lon, lat, float64(loc)/3600.0, module.inst)
}

func solarCulminationTime(date time.Time, lon float64, modules ...*Eclipse) time.Time {
	module := &defaultModule
	if len(modules) > 0 {
		module = modules[0]
	}

	localJD := basic.Date2JD(date.Add(time.Duration(-1*date.Hour())*time.Hour)) + 0.5
	_, loc := date.Zone()
	timezone := float64(loc) / 3600.0
	calcJD := basic.CulminationTimeWithInstance(localJD, lon, timezone, module.inst) - timezone/24.0
	return basic.JD2DateByZone(calcJD, date.Location(), false)
}

func lunarAzimuth(date time.Time, lon, lat float64, modules ...*Eclipse) float64 {
	module := &defaultModule
	if len(modules) > 0 {
		module = modules[0]
	}

	localJD := basic.Date2JD(date)
	_, loc := date.Zone()
	return basic.HMoonAzimuthWithInstance(localJD, lon, lat, float64(loc)/3600.0, module.inst)
}

func lunarAltitude(date time.Time, lon, lat float64, modules ...*Eclipse) float64 {
	module := &defaultModule
	if len(modules) > 0 {
		module = modules[0]
	}

	localJD := basic.Date2JD(date)
	_, loc := date.Zone()
	return basic.HMoonHeightWithInstance(localJD, lon, lat, float64(loc)/3600.0, module.inst)
}
