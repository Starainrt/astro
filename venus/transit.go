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
func NextTransit(date time.Time) TransitInfo {
	return transitInfoFromBasic(basic.NextVenusTransit(basic.Date2JDE(date.UTC())), date.Location())
}

// LastTransit 上一次地心金星凌日 / previous geocentric Venus transit.
func LastTransit(date time.Time) TransitInfo {
	return transitInfoFromBasic(basic.LastVenusTransit(basic.Date2JDE(date.UTC())), date.Location())
}

// ClosestTransit 最近一次地心金星凌日 / closest geocentric Venus transit.
func ClosestTransit(date time.Time) TransitInfo {
	return transitInfoFromBasic(basic.ClosestVenusTransit(basic.Date2JDE(date.UTC())), date.Location())
}

func transitInfoFromBasic(result basic.PlanetTransitResult, loc *time.Location) TransitInfo {
	if !result.Valid {
		return TransitInfo{}
	}
	start := basic.JDE2DateByZone(result.ExternalIngress, loc, false)
	greatest := basic.JDE2DateByZone(result.Greatest, loc, false)
	end := basic.JDE2DateByZone(result.ExternalEgress, loc, false)
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
		info.InternalStart = basic.JDE2DateByZone(result.InternalIngress, loc, false)
		info.InternalEnd = basic.JDE2DateByZone(result.InternalEgress, loc, false)
		info.InternalDuration = info.InternalEnd.Sub(info.InternalStart)
	}
	return info
}
