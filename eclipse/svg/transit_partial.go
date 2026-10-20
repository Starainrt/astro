package svg

import (
	"fmt"
	"math"
	"strings"

	"github.com/starainrt/astro/basic"
	"github.com/starainrt/astro/internal/svgmap"
)

func transitMapOrthographicCenter(transit basic.PlanetTransitResult) svgmap.GeoPoint {
	if !transit.HasExternal {
		if center, ok := transitMapVisibilityCenter(basic.PlanetTransitVisibilityColumns(transit)); ok {
			return center
		}
	}
	return solarEclipseSubsolarPoint(transitMapContactsOf(transit).greatest)
}

func transitMapVisibilityCenter(columns []basic.PlanetTransitVisibilityColumn) (svgmap.GeoPoint, bool) {
	var momentX, momentY, momentZ float64
	for index, column := range columns {
		weight := 0.0
		if index > 0 {
			weight += (column.Longitude - columns[index-1].Longitude) / 2
		}
		if index+1 < len(columns) {
			weight += (columns[index+1].Longitude - column.Longitude) / 2
		}
		longitude := column.Longitude * math.Pi / 180
		for _, interval := range column.Intervals {
			lower, upper := interval[0]*math.Pi/180, interval[1]*math.Pi/180
			horizontalMoment := (upper-lower)/2 + (math.Sin(2*upper)-math.Sin(2*lower))/4
			verticalMoment := (math.Sin(upper)*math.Sin(upper) - math.Sin(lower)*math.Sin(lower)) / 2
			momentX += weight * horizontalMoment * math.Cos(longitude)
			momentY += weight * horizontalMoment * math.Sin(longitude)
			momentZ += weight * verticalMoment
		}
	}
	if math.Sqrt(momentX*momentX+momentY*momentY+momentZ*momentZ) < 1e-12 {
		return svgmap.GeoPoint{}, false
	}
	return svgmap.GeoPoint{
		Longitude: math.Atan2(momentY, momentX) * 180 / math.Pi,
		Latitude:  math.Atan2(momentZ, math.Hypot(momentX, momentY)) * 180 / math.Pi,
	}, true
}

func transitMapProjectionLabel(projection svgmap.Projection, hasExternal bool, language string) string {
	if projection != svgmap.ProjectionOrthographic {
		return eclipseMapProjectionLabel(projection, language)
	}
	if language == "en" {
		center := "the subsolar point at greatest"
		if !hasExternal {
			center = "the topocentric visible footprint"
		}
		return "Orthographic globe projection, centred on " + center + "; one hemisphere only"
	}
	center := "凌甚太阳直射点"
	if !hasExternal {
		center = "站心可见区中心"
	}
	return "正射球面投影，视点取" + center + "，只画朝向视点的半个地球"
}

func writeTransitMapPartialRegions(builder *strings.Builder, frame svgmap.Frame, transit basic.PlanetTransitResult) {
	fmt.Fprintf(builder, `<g class="transit-visibility-regions" data-source="%s" clip-path="url(#transit-map-clip)">`, eclipseMapSourceTransitPartialVisibility)
	var path strings.Builder
	for _, polygon := range basic.PlanetTransitVisibilityPolygons(transit) {
		points := make([]svgmap.GeoPoint, len(polygon))
		for index, point := range polygon {
			points[index] = svgmap.GeoPoint{Longitude: point.Longitude, Latitude: point.Latitude}
		}
		path.WriteString(lunarEclipseGeoPolygonPath(points, frame))
	}
	if path.Len() > 0 {
		fmt.Fprintf(builder, `<path class="transit-partial-region" d="%s" fill="%s" fill-opacity="%.2f" fill-rule="nonzero"/>`, path.String(), transitMapColorPartial, transitMapPartialFillOpacity)
	}
	builder.WriteString(`</g>`)
}

func transitMapGeometryNote(transit basic.PlanetTransitResult, language string) string {
	text := fmt.Sprintf("γ %.4f", transit.Gamma)
	if transit.Series.Number != 0 {
		if language == "en" {
			text += fmt.Sprintf(" | Series %d (%d yr)", transit.Series.Number, transit.Series.PeriodYears)
		} else {
			text += fmt.Sprintf(" | 周期族 %d（%d 年）", transit.Series.Number, transit.Series.PeriodYears)
		}
	}
	return text
}
