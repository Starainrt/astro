package svg

import (
	"math"
	"time"

	"github.com/starainrt/astro/basic"
	eclipsecore "github.com/starainrt/astro/eclipse"
)

type LunarEclipseType = eclipsecore.LunarEclipseType

type LunarEclipseContactPoint = eclipsecore.LunarEclipseContactPoint

type LunarEclipseInfo = eclipsecore.LunarEclipseInfo

const (
	LunarEclipseNone      = eclipsecore.LunarEclipseNone
	LunarEclipsePenumbral = eclipsecore.LunarEclipsePenumbral
	LunarEclipsePartial   = eclipsecore.LunarEclipsePartial
	LunarEclipseTotal     = eclipsecore.LunarEclipseTotal
)

func lunarEclipseInfoFromBasic(result basic.LunarEclipseResult, location *time.Location) LunarEclipseInfo {
	return LunarEclipseInfo{
		Type:               mapBasicLunarEclipseType(result.Type),
		PenumbralMagnitude: result.PenumbralMagnitude,
		UmbralMagnitude:    result.Magnitude,
		PenumbralStart:     ttJDEToTime(result.PenumbralStart, location),
		PartialStart:       ttJDEToTime(result.PartialStart, location),
		TotalStart:         ttJDEToTime(result.TotalStart, location),
		Maximum:            ttJDEToTime(result.Maximum, location),
		TotalEnd:           ttJDEToTime(result.TotalEnd, location),
		PartialEnd:         ttJDEToTime(result.PartialEnd, location),
		PenumbralEnd:       ttJDEToTime(result.PenumbralEnd, location),
		ContactPoints:      lunarEclipseContactPointsFromBasic(result, location),
		HasPenumbral:       result.HasPenumbral,
		HasPartial:         result.HasPartial,
		HasTotal:           result.HasTotal,
	}
}

func lunarEclipseContactPointsFromBasic(
	result basic.LunarEclipseResult,
	location *time.Location,
) []LunarEclipseContactPoint {
	if !result.HasPenumbral {
		return nil
	}

	contacts := []LunarEclipseContactPoint{
		lunarEclipseContactPoint("P1", result.PenumbralStart, location, false),
	}
	if result.HasPartial {
		contacts = append(contacts, lunarEclipseContactPoint("U1", result.PartialStart, location, false))
	}
	if result.HasTotal {
		contacts = append(contacts, lunarEclipseContactPoint("U2", result.TotalStart, location, true))
	}
	if result.HasTotal {
		contacts = append(contacts, lunarEclipseContactPoint("U3", result.TotalEnd, location, true))
	}
	if result.HasPartial {
		contacts = append(contacts, lunarEclipseContactPoint("U4", result.PartialEnd, location, false))
	}
	contacts = append(contacts, lunarEclipseContactPoint("P4", result.PenumbralEnd, location, false))
	return contacts
}

func lunarEclipseContactPoint(
	label string,
	ttJDE float64,
	location *time.Location,
	internalContact bool,
) LunarEclipseContactPoint {
	moonCenterPA := lunarEclipseMoonCenterPositionAngle(ttJDE)
	shadowCenterPA := normalizeDegree360(moonCenterPA + 180)
	contactPA := shadowCenterPA
	if internalContact {
		contactPA = moonCenterPA
	}
	return LunarEclipseContactPoint{
		Label:                     label,
		Time:                      ttJDEToTime(ttJDE, location),
		ContactPositionAngle:      contactPA,
		ContactClockwiseAngle:     normalizeDegree360(360 - contactPA),
		MoonCenterPositionAngle:   moonCenterPA,
		ShadowCenterPositionAngle: shadowCenterPA,
	}
}

func lunarEclipseMoonCenterPositionAngle(ttJDE float64) float64 {
	shadowRA, shadowDec := lunarEclipseShadowCenterRaDec(ttJDE)
	moonRA, moonDec := basic.HMoonTrueRaDec(ttJDE)
	return lunarEclipsePositionAngle(shadowRA, shadowDec, moonRA, moonDec)
}

func lunarEclipseShadowCenterRaDec(ttJDE float64) (float64, float64) {
	sunRA, sunDec := basic.HSunApparentRaDec(ttJDE)
	return normalizeDegree360(sunRA + 180), -sunDec
}

func lunarEclipsePositionAngle(fromRA, fromDec, toRA, toDec float64) float64 {
	dRA := (toRA - fromRA) * math.Pi / 180
	fromDecRad := fromDec * math.Pi / 180
	toDecRad := toDec * math.Pi / 180
	angle := math.Atan2(
		math.Sin(dRA),
		math.Cos(fromDecRad)*math.Tan(toDecRad)-math.Sin(fromDecRad)*math.Cos(dRA),
	) * 180 / math.Pi
	return normalizeDegree360(angle)
}

func mapBasicLunarEclipseType(eclipseType basic.LunarEclipseType) LunarEclipseType {
	switch eclipseType {
	case basic.LunarEclipsePenumbral:
		return LunarEclipsePenumbral
	case basic.LunarEclipsePartial:
		return LunarEclipsePartial
	case basic.LunarEclipseTotal:
		return LunarEclipseTotal
	default:
		return LunarEclipseNone
	}
}

func ttJDEToTime(ttJDE float64, location *time.Location) time.Time {
	if ttJDE == 0 {
		return time.Time{}
	}
	utcJDE := basic.TD2UT(ttJDE, false)
	return basic.JDE2DateByZone(utcJDE, location, false)
}

func timeToTTJDE(date time.Time) float64 {
	utcJDE := basic.Date2JDE(date.UTC())
	return basic.TD2UT(utcJDE, true)
}

func normalizeDegree360(angle float64) float64 {
	angle = math.Mod(angle, 360)
	if angle < 0 {
		angle += 360
	}
	return angle
}
