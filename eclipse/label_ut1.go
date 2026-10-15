package eclipse

import (
	"time"

	"github.com/starainrt/astro/basic"
)

// 以下函数把结构里的民用时刻换成同一物理时刻的 UT1 时刻，供出图与导出共用。
// 零值 time.Time 表示该点或阶段不存在，一律原样保留；切片返回新副本，不改写调用方数据。

func ut1Label(t time.Time) time.Time {
	if t.IsZero() {
		return t
	}
	return basic.LabelIn(basic.TimeScaleUT1, t)
}

// TimeLabelsInUT1 返回换成 UT1 时刻的新切片 / returns a copy in UT1.
func TimeLabelsInUT1(times []time.Time) []time.Time {
	return ut1Labels(times)
}

func ut1Labels(times []time.Time) []time.Time {
	out := make([]time.Time, len(times))
	for i := range times {
		out[i] = ut1Label(times[i])
	}
	return out
}

func ut1PathPoint(p SolarEclipsePathPoint) SolarEclipsePathPoint {
	p.Time = ut1Label(p.Time)
	return p
}

func ut1PathPoints(points []SolarEclipsePathPoint) []SolarEclipsePathPoint {
	out := make([]SolarEclipsePathPoint, len(points))
	for i := range points {
		out[i] = ut1PathPoint(points[i])
	}
	return out
}

func ut1PathPointGrid(grid [][]SolarEclipsePathPoint) [][]SolarEclipsePathPoint {
	out := make([][]SolarEclipsePathPoint, len(grid))
	for i := range grid {
		out[i] = ut1PathPoints(grid[i])
	}
	return out
}

func ut1PartialFootprints(list []SolarEclipsePartialFootprint) []SolarEclipsePartialFootprint {
	out := make([]SolarEclipsePartialFootprint, len(list))
	for i := range list {
		out[i] = list[i]
		out[i].Time = ut1Label(list[i].Time)
		out[i].Boundaries = ut1PathPointGrid(list[i].Boundaries)
		out[i].HorizonEnds = ut1PathPoints(list[i].HorizonEnds)
	}
	return out
}

// SolarEclipseInfoInUT1 返回各阶段时刻换成 UT1 时刻的副本 / returns a copy in UT1.
func SolarEclipseInfoInUT1(info SolarEclipseInfo) SolarEclipseInfo {
	info.GreatestEclipse = ut1Label(info.GreatestEclipse)
	info.PartialBeginOnEarth = ut1Label(info.PartialBeginOnEarth)
	info.PartialEndOnEarth = ut1Label(info.PartialEndOnEarth)
	info.CentralBeginOnEarth = ut1Label(info.CentralBeginOnEarth)
	info.CentralEndOnEarth = ut1Label(info.CentralEndOnEarth)
	return info
}

// SolarEclipsePartialFootprintsInUT1 返回足迹、轮廓与接触点都换成 UT1 时刻的副本 / copy in UT1.
func SolarEclipsePartialFootprintsInUT1(info SolarEclipsePartialFootprintsInfo) SolarEclipsePartialFootprintsInfo {
	info.Eclipse = SolarEclipseInfoInUT1(info.Eclipse)
	info.Footprints = ut1PartialFootprints(info.Footprints)
	info.CentralShadowFootprints = ut1PartialFootprints(info.CentralShadowFootprints)
	info.CentralBandFootprints = ut1PartialFootprints(info.CentralBandFootprints)
	info.CentralBandSegments = ut1PathPointGrid(info.CentralBandSegments)
	info.CentralBandHorizonClosures = ut1PathPointGrid(info.CentralBandHorizonClosures)
	info.PartialBandContours = ut1PathPointGrid(info.PartialBandContours)
	contours := make([]SolarEclipseMagnitudeContour, len(info.MagnitudeContours))
	for i := range info.MagnitudeContours {
		contours[i] = info.MagnitudeContours[i]
		contours[i].Segments = ut1PathPointGrid(info.MagnitudeContours[i].Segments)
		contours[i].NorthernLimit = ut1PathPoints(info.MagnitudeContours[i].NorthernLimit)
		contours[i].SouthernLimit = ut1PathPoints(info.MagnitudeContours[i].SouthernLimit)
	}
	info.MagnitudeContours = contours
	greatest := make([]SolarEclipseGreatestTimeContour, len(info.GreatestTimeContours))
	for i := range info.GreatestTimeContours {
		greatest[i] = info.GreatestTimeContours[i]
		greatest[i].Time = ut1Label(info.GreatestTimeContours[i].Time)
		greatest[i].Segments = ut1PathPointGrid(info.GreatestTimeContours[i].Segments)
	}
	info.GreatestTimeContours = greatest
	curves := make([]SolarEclipseRiseSetCurve, len(info.RiseSetCurves))
	for i := range info.RiseSetCurves {
		curves[i] = info.RiseSetCurves[i]
		curves[i].Segments = ut1PathPointGrid(info.RiseSetCurves[i].Segments)
	}
	info.RiseSetCurves = curves
	info.P1 = ut1PathPoint(info.P1)
	info.P2 = ut1PathPoint(info.P2)
	info.P3 = ut1PathPoint(info.P3)
	info.P4 = ut1PathPoint(info.P4)
	info.U1 = ut1PathPoint(info.U1)
	info.U2 = ut1PathPoint(info.U2)
	info.U3 = ut1PathPoint(info.U3)
	info.U4 = ut1PathPoint(info.U4)
	return info
}

// SolarEclipsePathInUT1 返回中心线与限界都换成 UT1 时刻的副本 / copy in UT1.
func SolarEclipsePathInUT1(path SolarEclipsePath) SolarEclipsePath {
	path.Eclipse = SolarEclipseInfoInUT1(path.Eclipse)
	path.Greatest = ut1PathPoint(path.Greatest)
	path.CenterLine = ut1PathPoints(path.CenterLine)
	path.NorthernLimit = ut1PathPoints(path.NorthernLimit)
	path.SouthernLimit = ut1PathPoints(path.SouthernLimit)
	path.CentralBandSegments = ut1PathPointGrid(path.CentralBandSegments)
	return path
}

// LunarEclipseInfoInUT1 返回各阶段与接触点换成 UT1 时刻的副本 / copy in UT1.
func LunarEclipseInfoInUT1(info LunarEclipseInfo) LunarEclipseInfo {
	info.PenumbralStart = ut1Label(info.PenumbralStart)
	info.PartialStart = ut1Label(info.PartialStart)
	info.TotalStart = ut1Label(info.TotalStart)
	info.Maximum = ut1Label(info.Maximum)
	info.TotalEnd = ut1Label(info.TotalEnd)
	info.PartialEnd = ut1Label(info.PartialEnd)
	info.PenumbralEnd = ut1Label(info.PenumbralEnd)
	contacts := make([]LunarEclipseContactPoint, len(info.ContactPoints))
	for i := range info.ContactPoints {
		contacts[i] = info.ContactPoints[i]
		contacts[i].Time = ut1Label(info.ContactPoints[i].Time)
	}
	info.ContactPoints = contacts
	return info
}

// LocalSolarEclipseInfoInUT1 返回各阶段与接触点换成 UT1 时刻的副本 / copy in UT1.
func LocalSolarEclipseInfoInUT1(info LocalSolarEclipseInfo) LocalSolarEclipseInfo {
	info.GreatestEclipse = ut1Label(info.GreatestEclipse)
	info.PartialStart = ut1Label(info.PartialStart)
	info.PartialEnd = ut1Label(info.PartialEnd)
	info.CentralStart = ut1Label(info.CentralStart)
	info.CentralEnd = ut1Label(info.CentralEnd)
	contacts := make([]LocalSolarEclipseContactPoint, len(info.ContactPoints))
	for i := range info.ContactPoints {
		contacts[i] = info.ContactPoints[i]
		contacts[i].Time = ut1Label(info.ContactPoints[i].Time)
	}
	info.ContactPoints = contacts
	return info
}

// SolarEclipseGeocentricPanelInUT1 返回两个朔时刻换成 UT1 时刻的副本 / copy in UT1.
func SolarEclipseGeocentricPanelInUT1(panel SolarEclipseGeocentricPanel) SolarEclipseGeocentricPanel {
	panel.Conjunction = ut1Label(panel.Conjunction)
	panel.RightAscensionConjunction = ut1Label(panel.RightAscensionConjunction)
	return panel
}
