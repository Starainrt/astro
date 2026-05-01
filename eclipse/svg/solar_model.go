package svg

import (
	"math"
	"time"

	"github.com/starainrt/astro/basic"
	eclipsecore "github.com/starainrt/astro/eclipse"
)

type SolarEclipseRadiusModel = eclipsecore.SolarEclipseRadiusModel

type SolarEclipseType = eclipsecore.SolarEclipseType

type LocalSolarEclipseContactPoint = eclipsecore.LocalSolarEclipseContactPoint

type LocalSolarEclipseInfo = eclipsecore.LocalSolarEclipseInfo

const (
	SolarEclipseModelIAUSingleK         = eclipsecore.SolarEclipseModelIAUSingleK
	SolarEclipseModelNASABulletinSplitK = eclipsecore.SolarEclipseModelNASABulletinSplitK

	SolarEclipseNone    = eclipsecore.SolarEclipseNone
	SolarEclipsePartial = eclipsecore.SolarEclipsePartial
	SolarEclipseAnnular = eclipsecore.SolarEclipseAnnular
	SolarEclipseTotal   = eclipsecore.SolarEclipseTotal
	SolarEclipseHybrid  = eclipsecore.SolarEclipseHybrid
)

func localSolarEclipseInfoFromDiagram(
	diagram basic.LocalSolarEclipseDiagramResult,
	lon, lat, height float64,
	location *time.Location,
) LocalSolarEclipseInfo {
	info := localSolarEclipseInfoFieldsFromBasic(diagram.Eclipse, lon, lat, height, location)
	info.ContactPoints = localSolarEclipseContactPointsFromFrames(diagram.Frames, location)
	return info
}

func localSolarEclipseInfoFieldsFromBasic(
	result basic.LocalSolarEclipseResult,
	lon, lat, height float64,
	location *time.Location,
) LocalSolarEclipseInfo {
	visibleThreshold := localSolarEclipseVisibilityThreshold(height, lat)
	return LocalSolarEclipseInfo{
		Model:             mapBasicSolarEclipseModel(result.Model),
		Type:              mapBasicSolarEclipseType(result.Type),
		Longitude:         lon,
		Latitude:          lat,
		Height:            height,
		GreatestEclipse:   solarEclipseTTJDEToTime(result.GreatestEclipse, location),
		PartialStart:      solarEclipseTTJDEToTime(result.PartialStart, location),
		PartialEnd:        solarEclipseTTJDEToTime(result.PartialEnd, location),
		CentralStart:      solarEclipseTTJDEToTime(result.CentralStart, location),
		CentralEnd:        solarEclipseTTJDEToTime(result.CentralEnd, location),
		Magnitude:         result.Magnitude,
		Obscuration:       result.Obscuration,
		Separation:        result.Separation,
		SunAltitude:       result.SunAltitude,
		SunAzimuth:        result.SunAzimuth,
		VisibleAtGreatest: result.SunAltitude > visibleThreshold,
		HasPartial:        result.HasPartial,
		HasCentral:        result.HasCentral,
		HasAnnular:        result.HasAnnular,
		HasTotal:          result.HasTotal,
	}
}

func localSolarEclipseContactPointsFromFrames(
	frames []basic.LocalSolarEclipseDiagramFrame,
	location *time.Location,
) []LocalSolarEclipseContactPoint {
	contacts := make([]LocalSolarEclipseContactPoint, 0, 4)
	for _, frame := range frames {
		for _, label := range localSolarEclipseFrameLabels(frame) {
			switch label {
			case "C1", "C2", "C3", "C4":
				contactPA := frame.PositionAngle
				if (label == "C2" || label == "C3") && frame.MoonRadius >= frame.SunRadius {
					contactPA = normalizeSolarEclipseDegree360(contactPA + 180)
				}
				contacts = append(contacts, LocalSolarEclipseContactPoint{
					Label:                   label,
					Time:                    solarEclipseTTJDEToTime(frame.JDE, location),
					ContactPositionAngle:    contactPA,
					ContactClockwiseAngle:   normalizeSolarEclipseDegree360(360 - contactPA),
					MoonCenterPositionAngle: frame.PositionAngle,
				})
			}
		}
	}
	return contacts
}

func localSolarEclipseFrameLabels(frame basic.LocalSolarEclipseDiagramFrame) []string {
	if len(frame.Labels) > 0 {
		return frame.Labels
	}
	if frame.Label == "" {
		return nil
	}
	return []string{frame.Label}
}

func mapBasicSolarEclipseModel(model basic.SolarEclipseRadiusModel) SolarEclipseRadiusModel {
	switch model {
	case basic.SolarEclipseModelIAUSingleK:
		return SolarEclipseModelIAUSingleK
	default:
		return SolarEclipseModelNASABulletinSplitK
	}
}

func mapBasicSolarEclipseType(eclipseType basic.SolarEclipseType) SolarEclipseType {
	switch eclipseType {
	case basic.SolarEclipsePartial:
		return SolarEclipsePartial
	case basic.SolarEclipseAnnular:
		return SolarEclipseAnnular
	case basic.SolarEclipseTotal:
		return SolarEclipseTotal
	case basic.SolarEclipseHybrid:
		return SolarEclipseHybrid
	default:
		return SolarEclipseNone
	}
}

func solarEclipseTTJDEToTime(ttJDE float64, location *time.Location) time.Time {
	if ttJDE == 0 {
		return time.Time{}
	}
	utcJDE := basic.TD2UT(ttJDE, false)
	return basic.JDE2DateByZone(utcJDE, location, false)
}

func solarEclipseTimeToTTJDE(date time.Time) float64 {
	utcJDE := basic.Date2JDE(date.UTC())
	return basic.TD2UT(utcJDE, true)
}

func localSolarEclipseVisibilityThreshold(height, latitude float64) float64 {
	if height <= 0 {
		return 0
	}
	return -basic.HeightDegreeByLat(height, latitude)
}

func normalizeSolarEclipseDegree360(angle float64) float64 {
	angle = math.Mod(angle, 360)
	if angle < 0 {
		angle += 360
	}
	return angle
}
