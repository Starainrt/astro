package eclipse

import (
	"time"

	"github.com/starainrt/astro/basic"
	"github.com/starainrt/astro/tools"
)

func moonSunLoDiff(date time.Time) float64 {
	jde := basic.TD2UT(basic.Date2JDE(date.UTC()), true)
	sunLo := basic.HSunApparentLo(jde)
	moonLo := basic.HMoonApparentLo(jde)
	return tools.Limit360(moonLo - sunLo)
}

func solarAltitude(date time.Time, lon, lat float64) float64 {
	jde := basic.Date2JDE(date)
	_, loc := date.Zone()
	return basic.SunHeight(jde, lon, lat, float64(loc)/3600.0)
}

func solarCulminationTime(date time.Time, lon float64) time.Time {
	jde := basic.Date2JDE(date.Add(time.Duration(-1*date.Hour())*time.Hour)) + 0.5
	_, loc := date.Zone()
	timezone := float64(loc) / 3600.0
	calcJde := basic.CulminationTime(jde, lon, timezone) - timezone/24.0
	return basic.JDE2DateByZone(calcJde, date.Location(), false)
}

func lunarAzimuth(date time.Time, lon, lat float64) float64 {
	jde := basic.Date2JDE(date)
	_, loc := date.Zone()
	return basic.HMoonAzimuth(jde, lon, lat, float64(loc)/3600.0)
}

func lunarAltitude(date time.Time, lon, lat float64) float64 {
	jde := basic.Date2JDE(date)
	_, loc := date.Zone()
	return basic.HMoonHeight(jde, lon, lat, float64(loc)/3600.0)
}

func lunarCulminationTime(date time.Time, lon, lat float64) time.Time {
	if date.Hour() > 12 {
		date = date.Add(-12 * time.Hour)
	}
	jde := basic.Date2JDE(date)
	_, loc := date.Zone()
	return basic.JDE2DateByZone(basic.MoonCulminationTime(jde, lon, lat, float64(loc)/3600.0), date.Location(), true)
}
