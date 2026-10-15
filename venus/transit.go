package venus

import (
	"time"

	"github.com/starainrt/astro/basic"
)

// TransitInfo 地心金星凌日信息 / geocentric Venus transit information.
//
// Start、Greatest、End、InternalStart、InternalEnd 都保持调用者输入的时区。
// 内切接触不存在时 InternalStart / InternalEnd 为零值。
// Start, Greatest, End, InternalStart, and InternalEnd preserve the caller's timezone.
// InternalStart and InternalEnd are zero values when internal contacts do not exist.
type TransitInfo struct {
	Valid bool

	Start         time.Time
	InternalStart time.Time
	Greatest      time.Time
	InternalEnd   time.Time
	End           time.Time

	Duration         time.Duration
	InternalDuration time.Duration

	MinimumSeparationArcsec  float64
	SunSemidiameterArcsec    float64
	PlanetSemidiameterArcsec float64

	HasInternal bool
}

// NextTransit 下一次地心金星凌日 / next geocentric Venus transit.
func (module *Venus) NextTransit(date time.Time) TransitInfo {
	return transitInfoFromBasic(basic.NextVenusTransitWithInstance(basic.Date2JD(date.UTC()), module.inst), date.Location())
}

// LastTransit 上一次地心金星凌日 / previous geocentric Venus transit.
func (module *Venus) LastTransit(date time.Time) TransitInfo {
	return transitInfoFromBasic(basic.LastVenusTransitWithInstance(basic.Date2JD(date.UTC()), module.inst), date.Location())
}

// ClosestTransit 最近一次地心金星凌日 / closest geocentric Venus transit.
func (module *Venus) ClosestTransit(date time.Time) TransitInfo {
	return transitInfoFromBasic(basic.ClosestVenusTransitWithInstance(basic.Date2JD(date.UTC()), module.inst), date.Location())
}

func transitInfoFromBasic(result basic.PlanetTransitResult, loc *time.Location) TransitInfo {
	if !result.Valid {
		return TransitInfo{}
	}
	start := basic.JD2DateByZone(result.ExternalIngress, loc, false)
	greatest := basic.JD2DateByZone(result.Greatest, loc, false)
	end := basic.JD2DateByZone(result.ExternalEgress, loc, false)
	info := TransitInfo{
		Valid:                    true,
		Start:                    start,
		Greatest:                 greatest,
		End:                      end,
		Duration:                 end.Sub(start),
		MinimumSeparationArcsec:  result.MinimumSeparationArcsec,
		SunSemidiameterArcsec:    result.SunSemidiameterArcsec,
		PlanetSemidiameterArcsec: result.PlanetSemidiameterArcsec,
		HasInternal:              result.HasInternal,
	}
	if result.HasInternal {
		info.InternalStart = basic.JD2DateByZone(result.InternalIngress, loc, false)
		info.InternalEnd = basic.JD2DateByZone(result.InternalEgress, loc, false)
		info.InternalDuration = info.InternalEnd.Sub(info.InternalStart)
	}
	return info
}

// LocalTransitInfo 站心金星凌日信息 / topocentric Venus transit information.
//
// 时刻保持调用者传入的 TransitInfo 的时区；太阳高度为几何中心高度，不含折射。
// Times keep the timezone of the supplied TransitInfo; Sun altitudes are geometric centre altitudes without refraction.
type LocalTransitInfo = basic.PlanetTransitLocalInfo

// LocalTransitContact 站心凌日接触 / topocentric transit contact.
type LocalTransitContact = basic.PlanetTransitLocalContactInfo

// LocalTransit 地心凌日 transit 在观测地的站心接触与凌甚 / topocentric contacts of transit seen from the site.
//
// lon、lat 为度，东经、北纬为正；height 为海拔米。
// lon and lat are in degrees, east and north positive; height is in meters.
func (module *Venus) LocalTransit(transit TransitInfo, lon, lat, height float64) LocalTransitInfo {
	if !transit.Valid {
		return LocalTransitInfo{}
	}
	result := basic.PlanetTransitResultFromTimes(2,
		transit.Start, transit.InternalStart, transit.Greatest, transit.InternalEnd, transit.End)
	local := basic.PlanetTransitLocalWithInstance(result, basic.Observer{Longitude: lon, Latitude: lat, Height: height}, module.inst)
	return basic.PlanetTransitLocalInfoFromResult(local, transit.Greatest.Location())
}
