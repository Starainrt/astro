package svg

import (
	"errors"
	"fmt"
	"html"
	"math"
	"strings"
	"time"

	"github.com/starainrt/astro/internal/svgmap"
	"github.com/starainrt/astro/moon"
)

// ErrInvalidPlanetOccultationPath 表示有限盘面路径数据格式错误。
// ErrInvalidPlanetOccultationPath reports malformed finite-disk path data.
var ErrInvalidPlanetOccultationPath = errors.New("invalid planetary occultation path")

// ErrInvalidPlanetOccultationSVGOptions 表示行星 SVG 画布选项无效。
// ErrInvalidPlanetOccultationSVGOptions reports invalid planetary SVG canvas options.
var ErrInvalidPlanetOccultationSVGOptions = errors.New("invalid planetary occultation SVG options")

const (
	planetOccultationSVGMinimumWidth  = 640
	planetOccultationSVGMinimumHeight = 480
)

// PlanetOccultationSVGOptions 控制全球行星月掩路径 SVG。
// PlanetOccultationSVGOptions controls a global planetary-occultation path SVG.
type PlanetOccultationSVGOptions = StarOccultationSVGOptions

// FindPlanetOccultationSVGs 搜索时间窗口并渲染每条全球有限盘面行星月掩路径。
// FindPlanetOccultationSVGs searches a time window and renders every global finite-disk planetary occultation path.
func FindPlanetOccultationSVGs(
	start, end time.Time,
	planet moon.OccultationPlanet,
	pathOptions moon.OccultationPathOptions,
	options PlanetOccultationSVGOptions,
) ([]string, error) {
	paths, err := moon.FindPlanetOccultationPaths(start, end, planet, pathOptions)
	if err != nil {
		return nil, err
	}
	diagrams := make([]string, 0, len(paths))
	for _, path := range paths {
		diagram, renderErr := PlanetOccultationPathSVG(path, options)
		if renderErr != nil {
			return nil, renderErr
		}
		diagrams = append(diagrams, diagram)
	}
	return diagrams, nil
}

// PlanetOccultationPathSVG 渲染已计算的有限盘面行星路径。
// PlanetOccultationPathSVG renders a computed finite-disk planetary path.
func PlanetOccultationPathSVG(
	path moon.PlanetOccultationPath,
	options PlanetOccultationSVGOptions,
) (string, error) {
	if err := validatePlanetOccultationPath(path); err != nil {
		return "", err
	}
	targetID := path.TargetID
	if targetID == "" {
		targetID = path.Planet.String()
	}
	if err := validateStarOccultationSVGOptions(options); err != nil {
		return "", fmt.Errorf("%w: %v", ErrInvalidPlanetOccultationSVGOptions, err)
	}
	if options.Width > 0 && options.Width < planetOccultationSVGMinimumWidth {
		return "", fmt.Errorf("%w: width must be zero or at least %d", ErrInvalidPlanetOccultationSVGOptions, planetOccultationSVGMinimumWidth)
	}
	if options.Height > 0 && options.Height < planetOccultationSVGMinimumHeight {
		return "", fmt.Errorf("%w: height must be zero or at least %d", ErrInvalidPlanetOccultationSVGOptions, planetOccultationSVGMinimumHeight)
	}
	options = normalizeStarOccultationSVGOptions(options)
	starShape := planetOccultationStarShape(path, targetID)
	if targetID == path.Planet.String() && options.Language == starOccultationSVGLanguageChinese {
		targetID = planetOccultationChineseName(path.Planet)
		starShape.TargetID = targetID
	}
	options.Projection = MapProjection(resolvePlanetOccultationMapProjection(path, options.Projection))
	options = planetOccultationSVGDefaults(path, options)
	return renderOccultationPathSVG(starShape, &path, options), nil
}

func resolvePlanetOccultationMapProjection(
	path moon.PlanetOccultationPath,
	requested MapProjection,
) svgmap.Projection {
	minimumLatitude := path.Greatest.Latitude
	maximumLatitude := path.Greatest.Latitude
	for _, series := range [][]moon.OccultationPathPoint{
		path.CenterLine, path.NorthernLimit, path.SouthernLimit,
		path.NorthernTotalLimit, path.SouthernTotalLimit,
	} {
		for _, point := range series {
			minimumLatitude = math.Min(minimumLatitude, point.Latitude)
			maximumLatitude = math.Max(maximumLatitude, point.Latitude)
		}
	}
	for _, footprints := range [][]moon.PlanetOccultationFootprint{path.PartialFootprints, path.TotalFootprints} {
		for _, footprint := range footprints {
			for _, polygon := range footprint.Polygons {
				for _, point := range polygon {
					minimumLatitude = math.Min(minimumLatitude, point.Latitude)
					maximumLatitude = math.Max(maximumLatitude, point.Latitude)
				}
			}
		}
	}
	return svgmap.ResolveProjection(internalMapProjection(requested), path.Greatest.Latitude, minimumLatitude, maximumLatitude)
}

func planetOccultationStarShape(path moon.PlanetOccultationPath, targetID string) moon.StarOccultationPath {
	return moon.StarOccultationPath{
		TargetID:        targetID,
		Start:           path.Start,
		Greatest:        path.Greatest,
		End:             path.End,
		Complete:        path.Complete,
		CenterLine:      path.CenterLine,
		NorthernLimit:   path.NorthernLimit,
		SouthernLimit:   path.SouthernLimit,
		Step:            path.Step,
		TargetSpacingKM: path.TargetSpacingKM,
	}
}

func validatePlanetOccultationPath(path moon.PlanetOccultationPath) error {
	if err := path.Planet.Validate(); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidPlanetOccultationPath, err)
	}
	if err := validateStarOccultationPath(planetOccultationStarShape(path, path.TargetID)); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidPlanetOccultationPath, err)
	}
	if !path.HasTotalBand {
		if path.TotalComplete || !path.TotalStart.Time.IsZero() || !path.TotalEnd.Time.IsZero() ||
			len(path.NorthernTotalLimit) != 0 || len(path.SouthernTotalLimit) != 0 ||
			len(path.TotalFootprints) != 0 || path.GreatestTotalWidthKM != 0 {
			return fmt.Errorf("%w: total-band fields require HasTotalBand", ErrInvalidPlanetOccultationPath)
		}
		return validatePlanetOccultationFootprints("partial", path.PartialFootprints, path.Start.Time, path.End.Time)
	}
	if !path.TotalComplete {
		return fmt.Errorf("%w: global total-occultation band is incomplete", ErrInvalidPlanetOccultationPath)
	}
	if err := validateStarOccultationPathPoint("total start", path.TotalStart); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidPlanetOccultationPath, err)
	}
	if err := validateStarOccultationPathPoint("total end", path.TotalEnd); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidPlanetOccultationPath, err)
	}
	if !path.TotalStart.Time.After(path.Start.Time) || !path.TotalStart.Time.Before(path.Greatest.Time) ||
		!path.TotalEnd.Time.After(path.Greatest.Time) || !path.TotalEnd.Time.Before(path.End.Time) {
		return fmt.Errorf("%w: total-band times must be inside outer start, greatest, and end", ErrInvalidPlanetOccultationPath)
	}
	if !starOccultationFinite(path.GreatestTotalWidthKM) || path.GreatestTotalWidthKM <= 0 ||
		path.GreatestTotalWidthKM >= path.Greatest.WidthKM {
		return fmt.Errorf("%w: total-band width must be positive and narrower than the outer band", ErrInvalidPlanetOccultationPath)
	}
	if err := validatePlanetOccultationTotalLimits(path); err != nil {
		return err
	}
	if err := validatePlanetOccultationFootprints("partial", path.PartialFootprints, path.Start.Time, path.End.Time); err != nil {
		return err
	}
	if err := validatePlanetOccultationFootprints("total", path.TotalFootprints, path.TotalStart.Time, path.TotalEnd.Time); err != nil {
		return err
	}
	return nil
}

func validatePlanetOccultationFootprints(
	name string,
	footprints []moon.PlanetOccultationFootprint,
	start, end time.Time,
) error {
	for footprintIndex, footprint := range footprints {
		if footprint.Time.Before(start) || footprint.Time.After(end) {
			return fmt.Errorf("%w: %s footprint[%d] time must be inside its contact interval",
				ErrInvalidPlanetOccultationPath, name, footprintIndex)
		}
		if footprintIndex > 0 && !footprint.Time.After(footprints[footprintIndex-1].Time) {
			return fmt.Errorf("%w: %s footprint times must be strictly increasing",
				ErrInvalidPlanetOccultationPath, name)
		}
		if len(footprint.Polygons) == 0 {
			return fmt.Errorf("%w: %s footprint[%d] must contain a polygon",
				ErrInvalidPlanetOccultationPath, name, footprintIndex)
		}
		for polygonIndex, polygon := range footprint.Polygons {
			if len(polygon) < 3 {
				return fmt.Errorf("%w: %s footprint[%d].polygon[%d] must contain at least three points",
					ErrInvalidPlanetOccultationPath, name, footprintIndex, polygonIndex)
			}
			for pointIndex, point := range polygon {
				pointName := fmt.Sprintf("%s footprint[%d].polygon[%d][%d]", name, footprintIndex, polygonIndex, pointIndex)
				if err := validateStarOccultationPathPoint(pointName, point); err != nil {
					return fmt.Errorf("%w: %v", ErrInvalidPlanetOccultationPath, err)
				}
				if !point.Time.Equal(footprint.Time) {
					return fmt.Errorf("%w: %s time must match its footprint", ErrInvalidPlanetOccultationPath, pointName)
				}
			}
		}
	}
	return nil
}

func validatePlanetOccultationTotalLimits(path moon.PlanetOccultationPath) error {
	if len(path.NorthernTotalLimit) != len(path.SouthernTotalLimit) || len(path.NorthernTotalLimit) < 2 {
		return fmt.Errorf("%w: total northern and southern limits must contain matching samples", ErrInvalidPlanetOccultationPath)
	}
	for index := range path.NorthernTotalLimit {
		for _, item := range []struct {
			name  string
			point moon.OccultationPathPoint
		}{
			{fmt.Sprintf("northern total limit[%d]", index), path.NorthernTotalLimit[index]},
			{fmt.Sprintf("southern total limit[%d]", index), path.SouthernTotalLimit[index]},
		} {
			if err := validateStarOccultationPathPoint(item.name, item.point); err != nil {
				return fmt.Errorf("%w: %v", ErrInvalidPlanetOccultationPath, err)
			}
		}
		if !path.NorthernTotalLimit[index].Time.Equal(path.SouthernTotalLimit[index].Time) {
			return fmt.Errorf("%w: total limit sample %d times must match", ErrInvalidPlanetOccultationPath, index)
		}
		if index > 0 && !path.NorthernTotalLimit[index].Time.After(path.NorthernTotalLimit[index-1].Time) {
			return fmt.Errorf("%w: total limit times must be strictly increasing", ErrInvalidPlanetOccultationPath)
		}
	}
	last := len(path.NorthernTotalLimit) - 1
	if !path.NorthernTotalLimit[0].Time.Equal(path.TotalStart.Time) || !path.SouthernTotalLimit[0].Time.Equal(path.TotalStart.Time) ||
		!path.NorthernTotalLimit[last].Time.Equal(path.TotalEnd.Time) || !path.SouthernTotalLimit[last].Time.Equal(path.TotalEnd.Time) {
		return fmt.Errorf("%w: total limits must span total start through total end", ErrInvalidPlanetOccultationPath)
	}
	return nil
}

func planetOccultationSVGDefaults(
	path moon.PlanetOccultationPath,
	options PlanetOccultationSVGOptions,
) PlanetOccultationSVGOptions {
	if options.SummaryText == "" {
		options.SummaryText = planetOccultationSVGSummaryText(path, options)
	}
	if options.GreatestText == "" {
		options.GreatestText = planetOccultationSVGGreatestText(path, options.Language)
	}
	if options.MapTitle == "" {
		if options.Language == starOccultationSVGLanguageEnglish {
			options.MapTitle = "Global partial- and total-occultation bands"
		} else {
			options.MapTitle = "全球部分掩带与全掩带"
		}
	}
	if options.ContactsTitle == "" {
		if options.Language == starOccultationSVGLanguageEnglish {
			options.ContactsTitle = "Global phases"
		} else {
			options.ContactsTitle = "全球阶段"
		}
	}
	if options.FooterNote == "" {
		options.FooterNote = planetOccultationSVGFooter(path, options)
	}
	return options
}

func planetOccultationSVGSummaryText(path moon.PlanetOccultationPath, options PlanetOccultationSVGOptions) string {
	start := path.Start.Time.In(options.Location)
	greatest := path.Greatest.Time.In(options.Location)
	end := path.End.Time.In(options.Location)
	zone := starOccultationLocationLabel(greatest, options.Location)
	if !path.HasTotalBand {
		if options.Language == starOccultationSVGLanguageEnglish {
			return fmt.Sprintf("Partial begins %s | Greatest %s | Partial ends %s (%s)",
				starOccultationFormatEventTime(start, true),
				starOccultationFormatEventTime(greatest, !starOccultationSameDate(start, greatest)),
				starOccultationFormatEventTime(end, !starOccultationSameDate(start, end)), zone)
		}
		return fmt.Sprintf("外掩始 %s | 掩甚 %s | 外掩终 %s (%s)",
			starOccultationFormatEventTime(start, true),
			starOccultationFormatEventTime(greatest, !starOccultationSameDate(start, greatest)),
			starOccultationFormatEventTime(end, !starOccultationSameDate(start, end)), zone)
	}
	totalStart := path.TotalStart.Time.In(options.Location)
	totalEnd := path.TotalEnd.Time.In(options.Location)
	if options.Language == starOccultationSVGLanguageEnglish {
		return fmt.Sprintf("Partial begins %s | Total begins %s | Greatest %s | Total ends %s | Partial ends %s (%s)",
			starOccultationFormatEventTime(start, true),
			starOccultationFormatEventTime(totalStart, !starOccultationSameDate(start, totalStart)),
			starOccultationFormatEventTime(greatest, !starOccultationSameDate(start, greatest)),
			starOccultationFormatEventTime(totalEnd, !starOccultationSameDate(start, totalEnd)),
			starOccultationFormatEventTime(end, !starOccultationSameDate(start, end)), zone)
	}
	return fmt.Sprintf("外掩始 %s | 全掩始 %s | 掩甚 %s | 全掩终 %s | 外掩终 %s (%s)",
		starOccultationFormatEventTime(start, true),
		starOccultationFormatEventTime(totalStart, !starOccultationSameDate(start, totalStart)),
		starOccultationFormatEventTime(greatest, !starOccultationSameDate(start, greatest)),
		starOccultationFormatEventTime(totalEnd, !starOccultationSameDate(start, totalEnd)),
		starOccultationFormatEventTime(end, !starOccultationSameDate(start, end)), zone)
}

func planetOccultationSVGGreatestText(path moon.PlanetOccultationPath, language string) string {
	coordinates := starOccultationFormatCoordinates(path.Greatest.Longitude, path.Greatest.Latitude)
	if path.HasTotalBand {
		if language == starOccultationSVGLanguageEnglish {
			return fmt.Sprintf("Greatest point %s | partial-band width %.1f km | total-band width %.1f km",
				coordinates, path.Greatest.WidthKM, path.GreatestTotalWidthKM)
		}
		return fmt.Sprintf("掩甚点 %s | 部分掩带宽 %.1f km | 全掩带宽 %.1f km",
			coordinates, path.Greatest.WidthKM, path.GreatestTotalWidthKM)
	}
	if language == starOccultationSVGLanguageEnglish {
		return fmt.Sprintf("Greatest point %s | partial-band width %.1f km", coordinates, path.Greatest.WidthKM)
	}
	return fmt.Sprintf("掩甚点 %s | 部分掩带宽 %.1f km", coordinates, path.Greatest.WidthKM)
}

func planetOccultationSVGFooter(path moon.PlanetOccultationPath, options PlanetOccultationSVGOptions) string {
	projection := starOccultationProjectionLabel(options.Projection, options.Language)
	ringNote := ""
	if path.Planet == moon.OccultationSaturn {
		if options.Language == starOccultationSVGLanguageEnglish {
			ringNote = " Saturn's rings are not included in contact geometry."
		} else {
			ringNote = "土星环不参与接触计算。"
		}
	}
	if options.Language == starOccultationSVGLanguageEnglish {
		return projection + "; Natural Earth 1:50m physical land, no administrative boundaries. " +
			"The outer-contact cone bounds any disk overlap; the inner-contact cone bounds full planet-disk coverage." + ringNote
	}
	return projection + "；Natural Earth 1:50m 物理陆地底图，不含行政边界；" +
		"外切锥面界定任意圆盘重叠，内切锥面界定行星圆盘完全被月球遮住。" + ringNote
}

func writePlanetOccultationMap(
	b *strings.Builder,
	path moon.PlanetOccultationPath,
	starShape moon.StarOccultationPath,
	layout starOccultationSVGLayout,
	options StarOccultationSVGOptions,
) {
	layout.mapFrame().WriteOcean(b)
	writeStarOccultationGraticule(b, layout)
	writeStarOccultationLand(b, layout)
	if len(path.PartialFootprints) > 0 {
		writePlanetOccultationFootprintSweep(b, path.PartialFootprints, layout,
			"partial-occultation-band-layer", "occultation-band", "#e0ae43", 0.34)
	} else {
		writeOccultationBand(b, path.NorthernLimit, path.SouthernLimit, layout,
			"partial-occultation-band-layer", "occultation-band", "#e0ae43", 0.34)
	}
	if path.HasTotalBand {
		if len(path.TotalFootprints) > 0 {
			writePlanetOccultationFootprintSweep(b, path.TotalFootprints, layout,
				"total-occultation-band-layer", "total-occultation-band", "#607d98", 0.72)
		} else {
			writeOccultationBand(b, path.NorthernTotalLimit, path.SouthernTotalLimit, layout,
				"total-occultation-band-layer", "total-occultation-band", "#607d98", 0.72)
		}
	}
	if len(path.PartialFootprints) == 0 {
		writeStarOccultationGeoLine(b, path.NorthernLimit, layout, "northern-limit", "#a66f18", 1.25, "")
		writeStarOccultationGeoLine(b, path.SouthernLimit, layout, "southern-limit", "#a66f18", 1.25, "")
	}
	if path.HasTotalBand && len(path.TotalFootprints) == 0 {
		writeStarOccultationGeoLine(b, path.NorthernTotalLimit, layout, "northern-total-limit", "#355878", 1.3, "")
		writeStarOccultationGeoLine(b, path.SouthernTotalLimit, layout, "southern-total-limit", "#355878", 1.3, "")
	}
	writeStarOccultationGeoLine(b, starShape.CenterLine, layout, "center-line", "#59676b", 1.6, "5 4")
	writeStarOccultationVisibleCenterLine(b, starShape.CenterLine, layout)
	excluded := []time.Time{path.Start.Time, path.End.Time}
	if path.HasTotalBand {
		excluded = append(excluded, path.TotalStart.Time, path.TotalEnd.Time)
	}
	writeOccultationTimeMarkers(b, starShape.CenterLine, layout, options, excluded, path.Greatest.Time)
	writePlanetOccultationEventMarkers(b, path, layout, options.Language)
	layout.mapFrame().WriteFrame(b)
	writePlanetOccultationLegend(b, layout, options.Language, path.HasTotalBand)
}

func writePlanetOccultationFootprintSweep(
	b *strings.Builder,
	footprints []moon.PlanetOccultationFootprint,
	layout starOccultationSVGLayout,
	layerClass, pathClass, color string,
	opacity float64,
) {
	var geometry strings.Builder
	for _, footprint := range footprints {
		for _, polygon := range footprint.Polygons {
			geographic := make([]svgmap.GeoPoint, len(polygon))
			for index, point := range polygon {
				geographic[index] = svgmap.GeoPoint{Longitude: point.Longitude, Latitude: point.Latitude}
			}
			for _, fragment := range svgmap.PolygonFragments(geographic, layout.projection) {
				if len(fragment) < 3 {
					continue
				}
				if planetOccultationProjectedArea(layout, fragment) < 0 {
					for left, right := 0, len(fragment)-1; left < right; left, right = left+1, right-1 {
						fragment[left], fragment[right] = fragment[right], fragment[left]
					}
				}
				for index, point := range fragment {
					x, y := layout.project(point.Longitude, point.Latitude)
					command := "L"
					if index == 0 {
						command = "M"
					}
					fmt.Fprintf(&geometry, "%s %.3f %.3f ", command, x, y)
				}
				geometry.WriteString("Z ")
			}
		}
	}
	if geometry.Len() == 0 {
		return
	}
	fmt.Fprintf(b, `<g class="%s" clip-path="url(#occultation-map-clip)" fill="%s" fill-opacity="%.2f" stroke="none"><path class="%s" fill-rule="nonzero" d="%s"/></g>`,
		layerClass, color, opacity, pathClass, geometry.String())
}

func planetOccultationProjectedArea(layout starOccultationSVGLayout, points []svgmap.GeoPoint) float64 {
	area := 0.0
	for index, point := range points {
		next := points[(index+1)%len(points)]
		x1, y1 := layout.project(point.Longitude, point.Latitude)
		x2, y2 := layout.project(next.Longitude, next.Latitude)
		area += x1*y2 - x2*y1
	}
	return area / 2
}

type planetOccultationProjectedEventMarker struct {
	point     moon.OccultationPathPoint
	label     string
	kind      string
	x, y      float64
	visible   bool
	placement starOccultationEventMarkerPlacement
}

func writePlanetOccultationEventMarkers(
	b *strings.Builder,
	path moon.PlanetOccultationPath,
	layout starOccultationSVGLayout,
	language string,
) {
	markers := []planetOccultationProjectedEventMarker{
		{point: path.Start, label: planetOccultationEventLabel("start", language), kind: "start"},
	}
	if path.HasTotalBand {
		markers = append(markers, planetOccultationProjectedEventMarker{
			point: path.TotalStart, label: planetOccultationEventLabel("total-start", language), kind: "total-start",
		})
	}
	markers = append(markers, planetOccultationProjectedEventMarker{
		point: path.Greatest, label: planetOccultationEventLabel("greatest", language), kind: "greatest",
	})
	if path.HasTotalBand {
		markers = append(markers, planetOccultationProjectedEventMarker{
			point: path.TotalEnd, label: planetOccultationEventLabel("total-end", language), kind: "total-end",
		})
	}
	markers = append(markers, planetOccultationProjectedEventMarker{
		point: path.End, label: planetOccultationEventLabel("end", language), kind: "end",
	})

	for index := range markers {
		markers[index].x, markers[index].y, markers[index].visible = layout.mapFrame().Project(
			markers[index].point.Longitude, markers[index].point.Latitude,
		)
		markers[index].placement = starOccultationDefaultEventMarkerPlacement(
			markers[index].x, markers[index].y, markers[index].kind, layout,
		)
	}
	if path.HasTotalBand {
		separatePlanetOccultationMarkerPair(markers, 0, 1, layout)
		separatePlanetOccultationMarkerPair(markers, 4, 3, layout)
	}
	drawOrder := make([]int, len(markers))
	for index := range drawOrder {
		drawOrder[index] = index
	}
	if path.HasTotalBand {
		// 先绘制接近重合的外接触，再绘制较小的内接触点，使两个真实地理位置都保持可读。
		// Draw the near-coincident outer contacts first, then the smaller inner contacts, so both true geographic positions remain legible.
		drawOrder = []int{0, 4, 1, 3, 2}
	}
	for _, index := range drawOrder {
		marker := markers[index]
		if marker.visible {
			writeStarOccultationProjectedEventMarker(
				b, marker.x, marker.y, marker.label, marker.kind, marker.placement,
			)
		}
	}
}

func separatePlanetOccultationMarkerPair(
	markers []planetOccultationProjectedEventMarker,
	outerIndex, totalIndex int,
	layout starOccultationSVGLayout,
) {
	outer := &markers[outerIndex]
	total := &markers[totalIndex]
	if !outer.visible || !total.visible || math.Hypot(outer.x-total.x, outer.y-total.y) >= 30 {
		return
	}
	upper := math.Min(outer.y, total.y) - 10
	lower := math.Max(outer.y, total.y) + 18
	top := layout.mapY + 12
	bottom := layout.mapY + layout.mapHeight - 5
	if lower > bottom {
		lower = math.Min(outer.y, total.y) - 10
		upper = lower - 16
	}
	if upper < top {
		upper = math.Max(outer.y, total.y) + 16
		lower = upper + 16
	}
	outer.placement.labelY = math.Max(top, math.Min(bottom, upper))
	total.placement.labelY = math.Max(top, math.Min(bottom, lower))
	outer.placement.leader = true
	total.placement.leader = true
}

func writePlanetOccultationLegend(
	b *strings.Builder,
	layout starOccultationSVGLayout,
	language string,
	hasTotal bool,
) {
	labels := []string{"部分掩带", "可见中心线", "几何中心线"}
	if hasTotal {
		labels = []string{"部分掩带", "全掩带", "可见中心线", "几何中心线"}
	}
	if language == starOccultationSVGLanguageEnglish {
		labels = []string{"Partial band", "Visible center", "Geometric center"}
		if hasTotal {
			labels = []string{"Partial band", "Total band", "Visible center", "Geometric center"}
		}
	}
	y := layout.mapY + layout.mapHeight + 30
	itemWidth := layout.mapWidth / float64(len(labels))
	for index, label := range labels {
		x := layout.mapX + float64(index)*itemWidth
		switch {
		case index == 0:
			fmt.Fprintf(b, `<rect x="%.3f" y="%.3f" width="18" height="9" fill="#e0ae43" fill-opacity="0.55"/>`, x, y-10)
		case hasTotal && index == 1:
			fmt.Fprintf(b, `<rect x="%.3f" y="%.3f" width="18" height="9" fill="#607d98" fill-opacity="0.82"/>`, x, y-10)
		default:
			color, dash := "#087f8c", ""
			if index == len(labels)-1 {
				color, dash = "#59676b", "5 4"
			}
			fmt.Fprintf(b, `<line x1="%.3f" y1="%.3f" x2="%.3f" y2="%.3f" stroke="%s" stroke-width="2"`, x, y-5, x+18, y-5, color)
			if dash != "" {
				fmt.Fprintf(b, ` stroke-dasharray="%s"`, dash)
			}
			b.WriteString(`/>`)
		}
		fmt.Fprintf(b, `<text x="%.3f" y="%.3f" fill="#465053" font-family="Arial, sans-serif" font-size="9">%s</text>`,
			x+23, y, html.EscapeString(label))
	}
}

func writePlanetOccultationEventsPanel(
	b *strings.Builder,
	path moon.PlanetOccultationPath,
	layout starOccultationSVGLayout,
	options PlanetOccultationSVGOptions,
) {
	writeOccultationEventsPanel(b, planetOccultationSVGEventRows(path, options.Language), layout, options)
}

func planetOccultationSVGEventRows(path moon.PlanetOccultationPath, language string) []starOccultationEventRow {
	rows := []starOccultationEventRow{
		{name: planetOccultationEventLabel("start", language), point: path.Start},
	}
	if path.HasTotalBand {
		rows = append(rows, starOccultationEventRow{name: planetOccultationEventLabel("total-start", language), point: path.TotalStart})
	}
	rows = append(rows, starOccultationEventRow{name: planetOccultationEventLabel("greatest", language), point: path.Greatest})
	if path.HasTotalBand {
		rows = append(rows, starOccultationEventRow{name: planetOccultationEventLabel("total-end", language), point: path.TotalEnd})
	}
	return append(rows, starOccultationEventRow{name: planetOccultationEventLabel("end", language), point: path.End})
}

func planetOccultationEventLabel(kind, language string) string {
	if language == starOccultationSVGLanguageEnglish {
		switch kind {
		case "start":
			return "Partial begins"
		case "total-start":
			return "Total begins"
		case "greatest":
			return "Greatest"
		case "total-end":
			return "Total ends"
		default:
			return "Partial ends"
		}
	}
	switch kind {
	case "start":
		return "外掩始"
	case "total-start":
		return "全掩始"
	case "greatest":
		return "掩甚"
	case "total-end":
		return "全掩终"
	default:
		return "外掩终"
	}
}

func planetOccultationChineseName(planet moon.OccultationPlanet) string {
	switch planet {
	case moon.OccultationMercury:
		return "水星"
	case moon.OccultationVenus:
		return "金星"
	case moon.OccultationMars:
		return "火星"
	case moon.OccultationJupiter:
		return "木星"
	case moon.OccultationSaturn:
		return "土星"
	case moon.OccultationUranus:
		return "天王星"
	case moon.OccultationNeptune:
		return "海王星"
	default:
		return "行星"
	}
}
