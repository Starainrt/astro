package mercury

import (
	"time"

	"github.com/starainrt/astro/basic"
)

// TransitInfo 全球水星凌日与地心接触 / global Mercury transit and geocentric contacts.
//
// Start、Greatest、End、InternalStart、InternalEnd 都保持调用者输入的时区。
// 地心接触不存在时对应时刻为零值；WindowStart、WindowEnd 只是站心求解窗口。
// Start, Greatest, End, InternalStart, and InternalEnd preserve the caller's timezone.
// Absent geocentric contacts are zero; WindowStart and WindowEnd are only a local search window.
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

	HasInternal     bool
	HasExternal     bool
	WindowStart     time.Time
	WindowEnd       time.Time
	Gamma           float64
	PenumbralRadius float64
	Series          basic.PlanetTransitSeries
}

// NextTransit 下一次全球水星凌日 / next global Mercury transit.
func (module *Mercury) NextTransit(date time.Time) TransitInfo {
	return transitInfoFromBasic(basic.NextMercuryTransitWithInstance(basic.Date2JD(date.UTC()), module.inst), date.Location())
}

// LastTransit 上一次全球水星凌日 / previous global Mercury transit.
func (module *Mercury) LastTransit(date time.Time) TransitInfo {
	return transitInfoFromBasic(basic.LastMercuryTransitWithInstance(basic.Date2JD(date.UTC()), module.inst), date.Location())
}

// ClosestTransit 最近一次全球水星凌日 / closest global Mercury transit.
func (module *Mercury) ClosestTransit(date time.Time) TransitInfo {
	return transitInfoFromBasic(basic.ClosestMercuryTransitWithInstance(basic.Date2JD(date.UTC()), module.inst), date.Location())
}

func transitInfoFromBasic(result basic.PlanetTransitResult, loc *time.Location) TransitInfo {
	if !result.Valid {
		return TransitInfo{}
	}
	greatest := basic.JD2DateByZone(result.Greatest, loc, false)
	info := TransitInfo{
		Valid:                    true,
		Greatest:                 greatest,
		MinimumSeparationArcsec:  result.MinimumSeparationArcsec,
		SunSemidiameterArcsec:    result.SunSemidiameterArcsec,
		PlanetSemidiameterArcsec: result.PlanetSemidiameterArcsec,
		HasInternal:              result.HasInternal,
		HasExternal:              result.HasExternal,
		WindowStart:              basic.JD2DateByZone(result.WindowStart, loc, false),
		WindowEnd:                basic.JD2DateByZone(result.WindowEnd, loc, false),
		Gamma:                    result.Gamma,
		PenumbralRadius:          result.PenumbralRadius,
		Series:                   result.Series,
	}
	if result.HasExternal {
		info.Start = basic.JD2DateByZone(result.ExternalIngress, loc, false)
		info.End = basic.JD2DateByZone(result.ExternalEgress, loc, false)
		info.Duration = info.End.Sub(info.Start)
	}
	if result.HasInternal {
		info.InternalStart = basic.JD2DateByZone(result.InternalIngress, loc, false)
		info.InternalEnd = basic.JD2DateByZone(result.InternalEgress, loc, false)
		info.InternalDuration = info.InternalEnd.Sub(info.InternalStart)
	}
	return info
}

// LocalTransitInfo 站心水星凌日信息 / topocentric Mercury transit information.
//
// 时刻保持调用者传入的 TransitInfo 的时区；太阳高度为几何中心高度，不含折射。
// Times keep the timezone of the supplied TransitInfo; Sun altitudes are geometric centre altitudes without refraction.
type LocalTransitInfo = basic.PlanetTransitLocalInfo

// LocalTransitContact 站心凌日接触 / topocentric transit contact.
type LocalTransitContact = basic.PlanetTransitLocalContactInfo

// LocalTransit 凌日 transit 在观测地的站心接触与凌甚 / topocentric contacts of transit seen from the site.
//
// lon、lat 为度，东经、北纬为正；height 为海拔米。
// lon and lat are in degrees, east and north positive; height is in meters.
func (module *Mercury) LocalTransit(transit TransitInfo, lon, lat, height float64) LocalTransitInfo {
	if !transit.Valid {
		return LocalTransitInfo{}
	}
	local := basic.PlanetTransitLocalWithInstance(transitResultFromInfo(transit), basic.Observer{Longitude: lon, Latitude: lat, Height: height}, module.inst)
	return basic.PlanetTransitLocalInfoFromResult(local, transit.Greatest.Location())
}

// transitResultFromInfo 按一触、四触是否为零选择地心接触或求解窗口还原 basic 结果。
func transitResultFromInfo(transit TransitInfo) basic.PlanetTransitResult {
	if transit.Start.IsZero() && transit.End.IsZero() {
		return basic.PlanetTransitResultFromWindow(1, transit.WindowStart, transit.Greatest, transit.WindowEnd)
	}
	return basic.PlanetTransitResultFromTimes(1,
		transit.Start, transit.InternalStart, transit.Greatest, transit.InternalEnd, transit.End)
}
