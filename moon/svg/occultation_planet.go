package svg

import (
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/starainrt/astro/internal/geodata"
	"github.com/starainrt/astro/internal/occultationgeo"
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
	// 默认文案会先写回 options，必须在填充之前记下调用方真正给出的字段。
	flags := starOccultationSVGTextFlags(options)
	options = planetOccultationSVGDefaults(path, options)
	diagram, err := renderOccultationPathSVG(starShape, &path, options, flags)
	if err != nil {
		// 几何引擎失败不是数据格式错误，不能裹成 ErrInvalidPlanetOccultationPath 哨兵。
		return "", fmt.Errorf("planetary occultation SVG band geometry: %w", err)
	}
	return diagram, nil
}

func resolvePlanetOccultationMapProjection(
	path moon.PlanetOccultationPath,
	requested MapProjection,
) svgmap.Projection {
	minimumLatitude, maximumLatitude := planetOccultationBandLatitudeRange(path)
	// 掩带包住极点时等经纬会把它拉成横贯上下边缘的长条，必须改用极区图；
	// 判据要看带子本身的纬度范围，掩甚点在中纬也可能绕极。
	if requested == MapProjectionAuto {
		if maximumLatitude >= occultationPolarBandReach {
			return svgmap.ProjectionNorthPolar
		}
		if minimumLatitude <= -occultationPolarBandReach {
			return svgmap.ProjectionSouthPolar
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
		RiseSetCurves:   path.RiseSetCurves,
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
	if err := validateOccultationGreatestTimeContours(ErrInvalidPlanetOccultationPath, path.GreatestTimeContours); err != nil {
		return err
	}
	if !path.HasTotalBand {
		if path.TotalComplete || !path.TotalStart.Time.IsZero() || !path.TotalEnd.Time.IsZero() ||
			len(path.NorthernTotalLimit) != 0 || len(path.SouthernTotalLimit) != 0 ||
			len(path.TotalFootprints) != 0 || len(path.TotalBandFootprints) != 0 ||
			len(path.TotalBandContours) != 0 || len(path.TotalRiseSetCurves) != 0 ||
			path.GreatestTotalWidthKM != 0 {
			return fmt.Errorf("%w: total-band fields require HasTotalBand", ErrInvalidPlanetOccultationPath)
		}
		if err := validatePlanetOccultationFootprints("partial", path.PartialFootprints, path.Start.Time, path.End.Time); err != nil {
			return err
		}
		if err := validatePlanetOccultationContours(
			"partial band contours", path.PartialBandContours, path.Start.Time, path.End.Time,
		); err != nil {
			return err
		}
		if err := validatePlanetOccultationContours(
			"partial visibility contours", path.PartialVisibilityContours, path.Start.Time, path.End.Time,
		); err != nil {
			return err
		}
		return validatePlanetOccultationFootprints(
			"partial compact band", path.PartialBandFootprints, path.Start.Time, path.End.Time,
		)
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
	if err := occultationgeo.ValidateRiseSetCurves(path.TotalRiseSetCurves, path.TotalStart.Time, path.TotalEnd.Time); err != nil {
		return fmt.Errorf("%w: invalid total rise/set curves: %v", ErrInvalidPlanetOccultationPath, err)
	}
	if err := validatePlanetOccultationContours(
		"partial band contours", path.PartialBandContours, path.Start.Time, path.End.Time,
	); err != nil {
		return err
	}
	if err := validatePlanetOccultationContours(
		"partial visibility contours", path.PartialVisibilityContours, path.Start.Time, path.End.Time,
	); err != nil {
		return err
	}
	if err := validatePlanetOccultationContours(
		"total band contours", path.TotalBandContours, path.TotalStart.Time, path.TotalEnd.Time,
	); err != nil {
		return err
	}
	if err := validatePlanetOccultationContours(
		"total visibility contours", path.TotalVisibilityContours, path.TotalStart.Time, path.TotalEnd.Time,
	); err != nil {
		return err
	}
	if err := validatePlanetOccultationFootprints("partial", path.PartialFootprints, path.Start.Time, path.End.Time); err != nil {
		return err
	}
	if err := validatePlanetOccultationFootprints(
		"partial compact band", path.PartialBandFootprints, path.Start.Time, path.End.Time,
	); err != nil {
		return err
	}
	if err := validatePlanetOccultationFootprints("total", path.TotalFootprints, path.TotalStart.Time, path.TotalEnd.Time); err != nil {
		return err
	}
	return validatePlanetOccultationFootprints(
		"total compact band", path.TotalBandFootprints, path.TotalStart.Time, path.TotalEnd.Time,
	)
}

func validatePlanetOccultationContours(
	name string,
	contours [][]moon.OccultationPathPoint,
	start, end time.Time,
) error {
	return validateOccultationContours(ErrInvalidPlanetOccultationPath, name, contours, start, end)
}

func validatePlanetOccultationFootprints(
	name string,
	footprints []moon.PlanetOccultationFootprint,
	start, end time.Time,
) error {
	return validateOccultationFootprints(ErrInvalidPlanetOccultationPath, name, footprints, start, end)
}

func validateOccultationFootprints(
	baseError error,
	name string,
	footprints []moon.OccultationFootprint,
	start, end time.Time,
) error {
	if err := occultationgeo.ValidateFootprints(footprints, start, end); err != nil {
		return fmt.Errorf("%w: %s %v", baseError, name, err)
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
	// 带宽用南北限的地面间距。Greatest.WidthKM 是另一种构造（接触锥可见弧上横向偏移的极差），
	// 数值可以远大于实际带宽，两者不可互换。
	partialWidth := path.GreatestLimitSeparationKM
	if partialWidth <= 0 {
		partialWidth = path.Greatest.WidthKM
	}
	if path.HasTotalBand {
		if language == starOccultationSVGLanguageEnglish {
			return fmt.Sprintf("Greatest point %s | partial-band width %.1f km | total-band width %.1f km",
				coordinates, partialWidth, path.GreatestTotalWidthKM)
		}
		return fmt.Sprintf("掩甚点 %s | 部分掩带宽 %.1f km | 全掩带宽 %.1f km",
			coordinates, partialWidth, path.GreatestTotalWidthKM)
	}
	if language == starOccultationSVGLanguageEnglish {
		return fmt.Sprintf("Greatest point %s | partial-band width %.1f km", coordinates, partialWidth)
	}
	return fmt.Sprintf("掩甚点 %s | 部分掩带宽 %.1f km", coordinates, partialWidth)
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
) error {
	layout.mapFrame().WriteOcean(b)
	writeStarOccultationGraticule(b, layout)
	writeStarOccultationLand(b, layout)
	partial, err := writePlanetOccultationPartialBand(b, path, layout)
	if err != nil {
		return err
	}
	total := planetOccultationBand{}
	if path.HasTotalBand {
		total, err = writePlanetOccultationTotalBand(b, path, layout)
		if err != nil {
			return err
		}
	}
	if !partial.draw.compact && len(path.PartialFootprints) == 0 {
		writeStarOccultationGeoLine(b, path.NorthernLimit, layout,
			occultationLineStyle{className: "northern-limit", color: "#a66f18", strokeWidth: 1.25, boundary: true})
		writeStarOccultationGeoLine(b, path.SouthernLimit, layout,
			occultationLineStyle{className: "southern-limit", color: "#a66f18", strokeWidth: 1.25, boundary: true})
	}
	if path.HasTotalBand && !total.draw.compact && len(path.TotalFootprints) == 0 {
		writeStarOccultationGeoLine(b, path.NorthernTotalLimit, layout,
			occultationLineStyle{className: "northern-total-limit", color: "#355878", strokeWidth: 1.3, boundary: true})
		writeStarOccultationGeoLine(b, path.SouthernTotalLimit, layout,
			occultationLineStyle{className: "southern-total-limit", color: "#355878", strokeWidth: 1.3, boundary: true})
	}
	// Keep the physical phase curves above both translucent footprint bands;
	// do not turn them into a closed static-band outline.
	writeOccultationRiseSetCurves(b, path.RiseSetCurves, layout, "occultation")
	writeOccultationRiseSetCurves(b, path.TotalRiseSetCurves, layout, "occultation-total")
	writeOccultationHorizonConnectors(
		b, partial.footprints, path.NorthernLimit, path.SouthernLimit, path.RiseSetCurves, layout,
		false,
	)
	// 事件标记先占位、后绘制：三族标签共用一张占位表，后一族必须绕开先放置的标签。
	placer := &occultationLabelPlacer{}
	eventMarkers := planetOccultationEventMarkerDraws(path, layout, options.Language, placer)
	writeOccultationGreatestTimeContours(b, path.GreatestTimeContours, layout, options, placer)
	writeStarOccultationGeoLine(b, starShape.CenterLine, layout,
		occultationLineStyle{className: "center-line", color: "#59676b", strokeWidth: 1.6, dash: "5 4"})
	writeStarOccultationVisibleCenterLine(b, starShape.CenterLine, layout)
	excluded := []time.Time{path.Start.Time, path.End.Time}
	if path.HasTotalBand {
		excluded = append(excluded, path.TotalStart.Time, path.TotalEnd.Time)
	}
	writeOccultationTimeMarkers(b, starShape.CenterLine, layout, options, excluded, path.Greatest.Time, placer)
	writePlanetOccultationEventMarkers(b, path, eventMarkers)
	layout.mapFrame().WriteFrame(b)
	writePlanetOccultationLegend(
		b, layout, options.Language, path.HasTotalBand, len(path.RiseSetCurves) > 0,
		len(path.GreatestTimeContours) > 0,
	)
	return nil
}

// planetOccultationBand 是一次行星掩带填充的结果及它采用的足迹。
type planetOccultationBand struct {
	draw       occultationBandDraw
	footprints []moon.PlanetOccultationFootprint
}

// writePlanetOccultationPartialBand 绘制偏掩带；没有足迹时退回静态限带。
func writePlanetOccultationPartialBand(
	b *strings.Builder,
	path moon.PlanetOccultationPath,
	layout starOccultationSVGLayout,
) (planetOccultationBand, error) {
	footprints := path.PartialBandFootprints
	compact := len(footprints) > 0
	if len(footprints) == 0 {
		footprints = path.PartialFootprints
	}
	if len(footprints) == 0 {
		writeOccultationBand(b, path.NorthernLimit, path.SouthernLimit, layout,
			"partial-occultation-band-layer", "occultation-band", "#e0ae43", 0.34)
		return planetOccultationBand{}, nil
	}
	draw, err := (occultationFootprintSweep{
		footprints:         footprints,
		contours:           path.PartialBandContours,
		visibilityContours: path.PartialVisibilityContours,
		northern:           path.NorthernLimit,
		southern:           path.SouthernLimit,
		curves:             path.RiseSetCurves,
		layout:             layout,
		layerClass:         "partial-occultation-band-layer",
		pathClass:          "occultation-band",
		color:              "#e0ae43",
		opacity:            0.34,
		compactBand:        compact,
		kind:               partialOccultationFootprintSweep,
	}).write(b)
	if err != nil {
		return planetOccultationBand{}, err
	}
	return planetOccultationBand{draw: draw, footprints: footprints}, nil
}

// writePlanetOccultationTotalBand 绘制全掩带；没有足迹时退回静态内接触限带。
func writePlanetOccultationTotalBand(
	b *strings.Builder,
	path moon.PlanetOccultationPath,
	layout starOccultationSVGLayout,
) (planetOccultationBand, error) {
	footprints := path.TotalBandFootprints
	compact := len(footprints) > 0
	if len(footprints) == 0 {
		footprints = path.TotalFootprints
	}
	if len(footprints) == 0 {
		writeOccultationBand(b, path.NorthernTotalLimit, path.SouthernTotalLimit, layout,
			"total-occultation-band-layer", "total-occultation-band", "#607d98", 0.72)
		return planetOccultationBand{}, nil
	}
	draw, err := (occultationFootprintSweep{
		footprints:         footprints,
		contours:           path.TotalBandContours,
		visibilityContours: path.TotalVisibilityContours,
		northern:           path.NorthernTotalLimit,
		southern:           path.SouthernTotalLimit,
		curves:             path.TotalRiseSetCurves,
		layout:             layout,
		layerClass:         "total-occultation-band-layer",
		pathClass:          "total-occultation-band",
		color:              "#607d98",
		opacity:            0.72,
		compactBand:        compact,
		kind:               totalOccultationFootprintSweep,
	}).write(b)
	if err != nil {
		return planetOccultationBand{}, err
	}
	return planetOccultationBand{draw: draw, footprints: footprints}, nil
}

type occultationFootprintSweepKind uint8

const (
	starOccultationFootprintSweep occultationFootprintSweepKind = iota
	partialOccultationFootprintSweep
	totalOccultationFootprintSweep
)

// occultationBandSource 说明填充掩带几何的来源。
type occultationBandSource uint8

const (
	occultationBandAbsent occultationBandSource = iota
	occultationBandFootprints
	occultationBandEnvelope
	occultationBandSweep
)

// occultationBandDraw 报告一次掩带填充的结果。
type occultationBandDraw struct {
	source occultationBandSource
	// compact 表示填充由合并后的紧凑带构造，调用方不再叠加限线。
	compact bool
}

func (draw occultationBandDraw) drawn() bool { return draw.source != occultationBandAbsent }

// claimBoundary 报告图例能否声明掩带边界：解析包络与直接绘制的瞬时足迹都可以，
// 只有足迹扫掠合并回退不保证边界就是解析边界。
func (draw occultationBandDraw) claimBoundary() bool {
	return draw.source == occultationBandEnvelope || draw.source == occultationBandFootprints
}

// occultationFootprintSweep 是一次掩带填充需要的几何输入与画布样式。
type occultationFootprintSweep struct {
	footprints         []moon.PlanetOccultationFootprint
	contours           [][]moon.OccultationPathPoint
	visibilityContours [][]moon.OccultationPathPoint
	northern, southern []moon.OccultationPathPoint
	curves             []moon.OccultationRiseSetCurve
	layout             starOccultationSVGLayout
	layerClass         string
	pathClass          string
	color              string
	opacity            float64
	compactBand        bool
	kind               occultationFootprintSweepKind
}

// write 写出一个复合填充路径；几何引擎报错时返回错误，不退回原始瞬时足迹。
func (s occultationFootprintSweep) write(b *strings.Builder) (occultationBandDraw, error) {
	rings, source, err := s.resolveRings()
	if err != nil {
		return occultationBandDraw{}, err
	}
	draw := occultationBandDraw{compact: s.compactBand && len(rings) > 0}
	if len(rings) == 0 {
		return draw, nil
	}
	draw.source = source
	stroke, strokeWidth := s.outlineStyle(source)
	fmt.Fprintf(b, `<g class="%s" clip-path="url(#occultation-map-clip)" fill="%s" fill-opacity="%.2f" stroke="%s" stroke-opacity="0.88" stroke-width="%.2f"><path class="%s" fill-rule="nonzero" d="`,
		s.layerClass, s.color, s.opacity, stroke, strokeWidth, s.pathClass)
	for _, ring := range rings {
		writeOccultationPathCommands(b, ring)
		b.WriteString("Z ")
	}
	b.WriteString(`"/></g>`)
	return draw, nil
}

// resolveRings 返回裁剪到画布、定向并简化后的填充环。
func (s occultationFootprintSweep) resolveRings() ([][]occultationScreenPoint, occultationBandSource, error) {
	source := occultationBandFootprints
	var polygons [][]geodata.GeoPoint
	if s.compactBand {
		merged, authoritative, err := s.mergedBandPolygons()
		if err != nil {
			return nil, source, err
		}
		if len(merged) == 0 {
			return nil, source, errors.New("occultation band geometry produced no polygon")
		}
		polygons = merged
		source = occultationBandSweep
		if authoritative {
			source = occultationBandEnvelope
		}
	} else {
		polygons = planetOccultationFootprintPolygons(s.footprints)
	}
	rings := make([][]occultationScreenPoint, 0, len(polygons))
	for _, polygon := range polygons {
		geographic := make([]svgmap.GeoPoint, len(polygon))
		for index, point := range polygon {
			geographic[index] = svgmap.GeoPoint{Longitude: point.Longitude, Latitude: point.Latitude}
		}
		for _, fragment := range svgmap.PolygonFragments(geographic, s.layout.mapFrame().Clip()) {
			if len(fragment) < 3 {
				continue
			}
			if planetOccultationProjectedArea(s.layout, fragment) < 0 {
				for left, right := 0, len(fragment)-1; left < right; left, right = left+1, right-1 {
					fragment[left], fragment[right] = fragment[right], fragment[left]
				}
			}
			rings = append(rings, s.layout.projectOccultationGeoRing(fragment))
		}
	}
	return rings, source, nil
}

// mergedBandPolygons 命中缓存时直接复用只读的合并结果：紧凑带的合并远贵于逐足迹绘制。
func (s occultationFootprintSweep) mergedBandPolygons() ([][]geodata.GeoPoint, bool, error) {
	key := occultationBandCacheKey{kind: s.kind, digest: occultationBandDigest(s)}
	if cached, ok := lookupOccultationBandPolygons(key); ok {
		return cached.polygons, cached.authoritative, nil
	}
	polygons, authoritative, err := s.mergeBandPolygons()
	if err != nil {
		return nil, false, err
	}
	storeOccultationBandPolygons(key, occultationBandPolygons{polygons: polygons, authoritative: authoritative})
	return polygons, authoritative, nil
}

func (s occultationFootprintSweep) mergeBandPolygons() ([][]geodata.GeoPoint, bool, error) {
	if len(s.contours) == 0 {
		if s.kind == totalOccultationFootprintSweep {
			return occultationgeo.VisibleTotalBandPolygons(s.footprints, s.northern, s.southern, s.curves)
		}
		return occultationgeo.VisibleBandPolygons(s.footprints, s.northern, s.southern, s.curves)
	}
	if s.kind == totalOccultationFootprintSweep {
		return occultationgeo.VisibleTotalBandPolygonsFromAnalyticContours(
			s.footprints, s.contours, s.visibilityContours, s.northern, s.southern, s.curves,
		)
	}
	if s.kind == starOccultationFootprintSweep {
		return occultationgeo.VisibleStarBandPolygonsFromAnalyticContours(
			s.footprints, s.contours, s.visibilityContours, s.northern, s.southern, s.curves,
		)
	}
	return occultationgeo.VisibleBandPolygonsFromAnalyticContours(
		s.footprints, s.contours, s.visibilityContours, s.northern, s.southern, s.curves,
	)
}

// outlineStyle 只给解析轮廓描边；总掩带保持纯填充外观。
func (s occultationFootprintSweep) outlineStyle(source occultationBandSource) (string, float64) {
	if source != occultationBandEnvelope || s.pathClass == "total-occultation-band" {
		return "none", 0
	}
	return s.color, 1.3
}

func planetOccultationFootprintPolygons(footprints []moon.PlanetOccultationFootprint) [][]geodata.GeoPoint {
	polygons := make([][]geodata.GeoPoint, 0, len(footprints))
	for _, footprint := range footprints {
		for _, polygon := range footprint.Polygons {
			geographic := make([]geodata.GeoPoint, len(polygon))
			for index, point := range polygon {
				geographic[index] = geodata.GeoPoint{Longitude: point.Longitude, Latitude: point.Latitude}
			}
			polygons = append(polygons, geographic)
		}
	}
	return polygons
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

// planetOccultationEventMarkerDraws 计算全部事件标记；近似重合的外接触与内接触先分开标签再登记占位。
func planetOccultationEventMarkerDraws(
	path moon.PlanetOccultationPath,
	layout starOccultationSVGLayout,
	language string,
	placer *occultationLabelPlacer,
) []planetOccultationProjectedEventMarker {
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
	visible := make([]starOccultationEventMarkerDraw, 0, len(markers))
	for index := range markers {
		if !markers[index].visible {
			continue
		}
		visible = append(visible, starOccultationEventMarkerDraw{
			x: markers[index].x, y: markers[index].y,
			label: markers[index].label, kind: markers[index].kind,
			placement: markers[index].placement,
		})
	}
	reserveOccultationEventMarkerLabels(visible, layout, placer)
	offset := 0
	for index := range markers {
		if !markers[index].visible {
			continue
		}
		markers[index].placement = visible[offset].placement
		offset++
	}
	return markers
}

func writePlanetOccultationEventMarkers(
	b *strings.Builder,
	path moon.PlanetOccultationPath,
	markers []planetOccultationProjectedEventMarker,
) {
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
	hasRiseSet bool,
	hasIsochrones bool,
) {
	type legendItem struct {
		label string
		kind  string
	}
	items := []legendItem{{"部分掩带", "partial"}}
	if hasTotal {
		items = append(items, legendItem{"全掩带", "total"})
	}
	items = append(items, legendItem{"可见中心线", "visible"}, legendItem{"几何中心线", "geometric"})
	if hasRiseSet {
		items = append(items, legendItem{"初掩/掩甚/终掩月升月落线", "rise-set"})
	}
	if hasIsochrones {
		items = append(items, legendItem{"掩甚时刻等时线", "isochrone"})
	}
	if language == starOccultationSVGLanguageEnglish {
		for index := range items {
			switch items[index].kind {
			case "partial":
				items[index].label = "Partial band"
			case "total":
				items[index].label = "Total band"
			case "visible":
				items[index].label = "Visible center"
			case "geometric":
				items[index].label = "Geometric center"
			case "rise-set":
				items[index].label = "Rise/set phase lines"
			case "isochrone":
				items[index].label = "Greatest-time isochrones"
			}
		}
	}
	y := layout.mapY + layout.mapHeight + 30
	legend := make([]occultationLegendItem, 0, len(items))
	for _, item := range items {
		current := occultationLegendItem{label: item.label, markerWidth: 18}
		switch item.kind {
		case "partial":
			current.fill, current.fillOpacity = "#e0ae43", 0.55
		case "total":
			current.fill, current.fillOpacity = "#607d98", 0.82
		case "geometric":
			current.color, current.dash = "#59676b", "5 4"
		case "rise-set":
			current.color = "#d97706"
		case "isochrone":
			current.color = "#1f6fb2"
		default:
			current.color = "#087f8c"
		}
		legend = append(legend, current)
	}
	writeOccultationLegendItems(b, y, layout, legend, 9)
}

func writeOccultationHorizonConnectors(
	b *strings.Builder,
	footprints []moon.OccultationFootprint,
	northern, southern []moon.OccultationPathPoint,
	curves []moon.OccultationRiseSetCurve,
	layout starOccultationSVGLayout,
	pointSource bool,
) {
	connectors := occultationgeo.HorizonConnectorSegments(footprints, curves, northern, southern)
	if pointSource {
		connectors = occultationgeo.StarHorizonConnectorSegments(footprints, curves, northern, southern)
	}
	for _, connector := range connectors {
		if len(connector.Points) < 2 {
			continue
		}
		points := occultationgeo.DensifyOccultationPathPoints(connector.Points, 35)
		className := fmt.Sprintf("occultation-rise-set-boundary occultation-horizon-%s", connector.Direction)
		writeStarOccultationGeoLine(
			b, points, layout,
			occultationLineStyle{className: className, color: "#d97706", strokeWidth: 1.35},
		)
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

// occultationPolarBandReach 是判定掩带已经包住极点的纬度阈值。
const occultationPolarBandReach = 89.0

// planetOccultationBandLatitudeRange 返回掩带（中心线、南北限与足印）的纬度范围。
func planetOccultationBandLatitudeRange(path moon.PlanetOccultationPath) (float64, float64) {
	minimumLatitude, maximumLatitude := path.Greatest.Latitude, path.Greatest.Latitude
	for _, series := range [][]moon.OccultationPathPoint{
		path.CenterLine, path.NorthernLimit, path.SouthernLimit,
		path.NorthernTotalLimit, path.SouthernTotalLimit,
	} {
		for _, point := range series {
			minimumLatitude = math.Min(minimumLatitude, point.Latitude)
			maximumLatitude = math.Max(maximumLatitude, point.Latitude)
		}
	}
	for _, footprints := range [][]moon.PlanetOccultationFootprint{
		path.PartialFootprints, path.PartialBandFootprints,
		path.TotalFootprints, path.TotalBandFootprints,
	} {
		for _, footprint := range footprints {
			for _, polygon := range footprint.Polygons {
				for _, point := range polygon {
					minimumLatitude = math.Min(minimumLatitude, point.Latitude)
					maximumLatitude = math.Max(maximumLatitude, point.Latitude)
				}
			}
		}
	}
	return minimumLatitude, maximumLatitude
}

// planetOccultationOrthographicCentre 取掩甚点作为正射球面的视点。
func planetOccultationOrthographicCentre(path moon.PlanetOccultationPath) svgmap.GeoPoint {
	return svgmap.GeoPoint{Longitude: path.Greatest.Longitude, Latitude: path.Greatest.Latitude}
}
