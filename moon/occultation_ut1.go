package moon

import (
	"time"

	"github.com/starainrt/astro"
	"github.com/starainrt/astro/basic"
)

// 月掩各载体的 UT1 时刻换算，供月掩 SVG 使用。两条铁律与日食侧一致：
// 零值 time.Time 表示该点或阶段不存在，一律原样保留；返回新副本，不改写调用方传进来的 path/info。

func occultationUT1Label(t time.Time) time.Time {
	if t.IsZero() {
		return t
	}
	return astro.LabelIn(astro.TimeScaleUT1, t)
}

func occultationUT1PathPoint(p basic.OccultationPathPoint) basic.OccultationPathPoint {
	p.Time = occultationUT1Label(p.Time)
	return p
}

func occultationUT1PathPoints(points []basic.OccultationPathPoint) []basic.OccultationPathPoint {
	out := make([]basic.OccultationPathPoint, len(points))
	for i := range points {
		out[i] = occultationUT1PathPoint(points[i])
	}
	return out
}

func occultationUT1PathGrid(grid [][]basic.OccultationPathPoint) [][]basic.OccultationPathPoint {
	out := make([][]basic.OccultationPathPoint, len(grid))
	for i := range grid {
		out[i] = occultationUT1PathPoints(grid[i])
	}
	return out
}

func occultationUT1Footprints(list []basic.OccultationFootprint) []basic.OccultationFootprint {
	out := make([]basic.OccultationFootprint, len(list))
	for i := range list {
		out[i] = list[i]
		out[i].Time = occultationUT1Label(list[i].Time)
		out[i].Polygons = occultationUT1PathGrid(list[i].Polygons)
		out[i].InteriorPolygons = occultationUT1PathGrid(list[i].InteriorPolygons)
		out[i].Boundaries = occultationUT1PathGrid(list[i].Boundaries)
	}
	return out
}

func occultationUT1RiseSetCurves(list []basic.OccultationRiseSetCurve) []basic.OccultationRiseSetCurve {
	out := make([]basic.OccultationRiseSetCurve, len(list))
	for i := range list {
		out[i] = list[i]
		out[i].Segments = occultationUT1PathGrid(list[i].Segments)
	}
	return out
}

func occultationUT1GreatestTimeContours(list []basic.OccultationGreatestTimeContour) []basic.OccultationGreatestTimeContour {
	out := make([]basic.OccultationGreatestTimeContour, len(list))
	for i := range list {
		out[i] = list[i]
		out[i].Time = occultationUT1Label(list[i].Time)
		out[i].Segments = occultationUT1PathGrid(list[i].Segments)
	}
	return out
}

// StarOccultationPathInUT1 返回恒星月掩全球路径的 UT1 时刻副本 / copy in UT1 of a stellar path.
func StarOccultationPathInUT1(path basic.StarOccultationPath) basic.StarOccultationPath {
	path.Start = occultationUT1PathPoint(path.Start)
	path.Greatest = occultationUT1PathPoint(path.Greatest)
	path.End = occultationUT1PathPoint(path.End)
	path.CenterLine = occultationUT1PathPoints(path.CenterLine)
	path.NorthernLimit = occultationUT1PathPoints(path.NorthernLimit)
	path.SouthernLimit = occultationUT1PathPoints(path.SouthernLimit)
	path.BandContours = occultationUT1PathGrid(path.BandContours)
	path.VisibilityContours = occultationUT1PathGrid(path.VisibilityContours)
	path.Footprints = occultationUT1Footprints(path.Footprints)
	path.BandFootprints = occultationUT1Footprints(path.BandFootprints)
	path.RiseSetCurves = occultationUT1RiseSetCurves(path.RiseSetCurves)
	path.GreatestTimeContours = occultationUT1GreatestTimeContours(path.GreatestTimeContours)
	return path
}

// PlanetOccultationPathInUT1 返回行星月掩全球掩带的 UT1 时刻副本 / copy in UT1 of a planetary path.
func PlanetOccultationPathInUT1(path basic.PlanetOccultationPath) basic.PlanetOccultationPath {
	path.Start = occultationUT1PathPoint(path.Start)
	path.Greatest = occultationUT1PathPoint(path.Greatest)
	path.End = occultationUT1PathPoint(path.End)
	path.CenterLine = occultationUT1PathPoints(path.CenterLine)
	path.NorthernLimit = occultationUT1PathPoints(path.NorthernLimit)
	path.SouthernLimit = occultationUT1PathPoints(path.SouthernLimit)
	path.PartialBandContours = occultationUT1PathGrid(path.PartialBandContours)
	path.PartialVisibilityContours = occultationUT1PathGrid(path.PartialVisibilityContours)
	path.PartialFootprints = occultationUT1Footprints(path.PartialFootprints)
	path.PartialBandFootprints = occultationUT1Footprints(path.PartialBandFootprints)
	path.RiseSetCurves = occultationUT1RiseSetCurves(path.RiseSetCurves)
	path.TotalStart = occultationUT1PathPoint(path.TotalStart)
	path.TotalEnd = occultationUT1PathPoint(path.TotalEnd)
	path.NorthernTotalLimit = occultationUT1PathPoints(path.NorthernTotalLimit)
	path.SouthernTotalLimit = occultationUT1PathPoints(path.SouthernTotalLimit)
	path.TotalBandContours = occultationUT1PathGrid(path.TotalBandContours)
	path.TotalVisibilityContours = occultationUT1PathGrid(path.TotalVisibilityContours)
	path.TotalFootprints = occultationUT1Footprints(path.TotalFootprints)
	path.TotalBandFootprints = occultationUT1Footprints(path.TotalBandFootprints)
	path.TotalRiseSetCurves = occultationUT1RiseSetCurves(path.TotalRiseSetCurves)
	path.GreatestTimeContours = occultationUT1GreatestTimeContours(path.GreatestTimeContours)
	return path
}

// StarOccultationInfoInUT1 返回固定地点恒星月掩的 UT1 时刻副本 / copy in UT1 of a local stellar result.
func StarOccultationInfoInUT1(info basic.StarOccultationInfo) basic.StarOccultationInfo {
	info.Immersion = occultationUT1Label(info.Immersion)
	info.Greatest = occultationUT1Label(info.Greatest)
	info.Emersion = occultationUT1Label(info.Emersion)
	return info
}

// PlanetOccultationInfoInUT1 返回固定地点行星月掩的 UT1 时刻副本 / copy in UT1 of a local planetary result.
func PlanetOccultationInfoInUT1(info basic.PlanetOccultationInfo) basic.PlanetOccultationInfo {
	info.ExternalImmersion = occultationUT1Label(info.ExternalImmersion)
	info.InternalImmersion = occultationUT1Label(info.InternalImmersion)
	info.Greatest = occultationUT1Label(info.Greatest)
	info.InternalEmersion = occultationUT1Label(info.InternalEmersion)
	info.ExternalEmersion = occultationUT1Label(info.ExternalEmersion)
	return info
}
