package svg

import (
	"github.com/starainrt/astro/internal/timenote"
	"fmt"
	"html"
	"math"
	"strings"
	"time"

	"github.com/starainrt/astro"
	"github.com/starainrt/astro/basic"
	eclipsecore "github.com/starainrt/astro/eclipse"
	"github.com/starainrt/astro/internal/geodata"
	"github.com/starainrt/astro/internal/solarclosure"
	"github.com/starainrt/astro/internal/svgchart"
	"github.com/starainrt/astro/internal/svgmap"
)

const (
	solarEclipseMapDefaultWidth  = 960
	solarEclipseMapDefaultHeight = 640
	// 真实可用下限：更窄的画布上横版的地图框与右栏数据网格必然水平重叠，面板行距也会压到 1 px 以下。
	solarEclipseMapMinWidth  = 800
	solarEclipseMapMinHeight = 560
	// 等时线时刻取值上限，避免极长事件请求出上万条曲线。
	solarEclipseMapGreatestTimeMaxLevels = 64
	// NASA 全球图的等时线间隔，属于推荐取值而不是默认值：不请求就不画。
	solarEclipseMapDefaultGreatestTimeStep = 30 * time.Minute
	// 偏食足迹扫描的最密步长：并集需要一整条自洽的扫描序列，更密的请求只会成倍放大成本。
	solarEclipseMapPartialSweepMinStep = 2 * time.Minute
	// 中心线偏离包络不得超过 2 km；南北限界端部允许偏离食带宽的 1/4。
	solarEclipseCentralEnvelopeAxisMissToleranceKM = 2.0
	solarEclipseCentralEnvelopeLimitMissFraction   = 0.25
)

// SolarEclipseMapSVGOptions 控制无国界全球日食地图。
// SolarEclipseMapSVGOptions controls a border-free global solar-eclipse map.
type SolarEclipseMapSVGOptions struct {
	// Width 和 Height 是 SVG 画布尺寸；宽度小于 800 或高度小于 560 时使用 960x640 默认值，
	// 这两个下限是横版地图框与两栏数据网格不重叠、面板行距不塌陷的最小值。
	// 图例行数随可选图层增加，因此更窄或更矮的画布、以及开着全部可选图层的下限画布会返回 false。
	// Width and Height are SVG canvas dimensions in user units. Width values below 800 and height
	// values below 560 use the 960x640 defaults; those minima keep the landscape map frame and the
	// two-column data grid apart with a non-collapsed panel row spacing. The legend gains rows with
	// the optional layers, so canvases too small for that stack make the renderer return false.
	Width  int
	Height int
	// Language 为 "en"（不区分大小写）时使用英文，否则使用中文。
	// Language uses English for "en" (case-insensitive) and Chinese otherwise.
	Language string
	// Location 控制显示的事件时刻；nil 使用 date.Location()。
	// Location controls displayed event times. Nil uses date.Location().
	Location *time.Location
	// TimeScale 选择图中时刻的时标：零值 UTC；TimeScaleUT1 改用 UT1 时刻，此时 Location 必须是
	// nil 或 UTC（否则返回 false），并在图注里声明尺度。
	// TimeScale selects the label scale: the zero value is UTC; TimeScaleUT1 uses UT1 labels, requires
	// Location to be nil or UTC (otherwise the renderer returns false) and declares the scale in the footer.
	TimeScale astro.TimeScale
	// Projection 选择地图投影；零值从事件几何中自动选择，不支持的值使渲染器返回 false。
	// Projection selects the map projection. The zero value selects one from the event geometry; unsupported values make the renderer return false.
	Projection EclipseMapProjection
	// 空文本字段使用本地化的自动标签。
	// Empty text fields use localized automatic labels.
	Title    string
	MapTitle string
	// EventsTitle 是“全球阶段”数据块的标题，该块给出食甚经纬度与地球范围的中心食始/终；
	// 为空时使用该块的本地化默认标题。
	// EventsTitle titles the global-phases block that carries the greatest-eclipse coordinates
	// and the earth-wide central begin/end; empty uses that block's localized default title.
	EventsTitle string
	FooterNote  string

	// PartialStep 是半影足迹请求的时间步长；非正值与小于两分钟的正值都使用两分钟。
	// 偏食区填充是瞬时足迹的并集，成本随采样数成倍增长，而并集必须由一整条自洽的扫描序列生成，
	// 更密的请求不改变产物（长事件还会为 30000 个采样点的上限进一步放大步长）。
	// PartialStep is the requested partial-footprint time step. Values <= 0 and positive values below
	// two minutes both use two minutes: the partial region is a union of instantaneous footprints whose
	// cost grows with the sample count, and that union needs one self-consistent sweep, so a denser
	// request does not change the product (long events enlarge the step further to stay within 30000 samples).
	PartialStep time.Duration
	// BoundaryPoints 是每个瞬时偏食足迹的角向采样数；非正值使用 180，正值限制在 12..1440。
	// BoundaryPoints is the angular sample count for each instantaneous partial footprint. Values <= 0 use 180; positive values are clamped to 12..1440.
	BoundaryPoints int
	// PenumbralOutlineStep 请求瞬时半影边界轮廓并给出采样间隔；默认不画（零值或负值都不画），
	// 正值小于一分钟时使用一分钟。NASA 全球图没有这族线，打开会明显遮挡地球。
	// PenumbralOutlineStep requests sampled instantaneous penumbral outlines and gives their
	// interval. They are off by default (zero or negative draws none), and positive values below
	// one minute use one minute. NASA world maps omit this family, which otherwise hides the globe.
	PenumbralOutlineStep time.Duration
	// CentralShadowStep 请求瞬时本影/反本影轮廓并给出采样间隔；默认不画（零值或负值都不画），
	// 正值小于一分钟时使用一分钟。它同时决定本影足迹的采样密度。
	// CentralShadowStep requests sampled instantaneous umbral/antumbral outlines and gives their
	// interval. They are off by default (zero or negative draws none), and positive values below
	// one minute use one minute. It also sets the umbral footprint sampling density.
	CentralShadowStep time.Duration
	// CentralStep 是中心路径请求的时间步长；非正值使用两分钟，正值小于一秒时使用一秒。长事件可能增大实际步长，以保持基础路径不超过 30000 个采样点。
	// CentralStep is the requested central-path time step. Values <= 0 use two minutes; positive values below one second use one second. Long events may use a larger effective step to keep the base path within 30000 samples.
	CentralStep time.Duration
	// TargetSpacingKM 是中心线地面间距上限，单位为千米；非正值（含 -Inf）使用 150 km，NaN 与 +Inf 禁用加密。
	// TargetSpacingKM is the requested maximum center-line ground spacing in kilometers. Values <= 0 (including -Inf) use 150 km; NaN and +Inf disable refinement.
	TargetSpacingKM float64
	// TimeLabelStep 控制中心线上的 HH:MM 标签；零值使用 30 分钟，负值禁用标签，正值小于一分钟时使用一分钟。
	// TimeLabelStep controls HH:MM labels along the central line. Zero uses 30 minutes, negative values disable labels, and positive values below one minute use one minute.
	TimeLabelStep time.Duration
	// GreatestTimeStep 控制食甚时刻等时线的时间间隔；默认不画（零值或负值都不画），
	// 正值小于一分钟时使用一分钟；NASA 全球图用 solarEclipseMapDefaultGreatestTimeStep 的间隔。
	// 等时线是固定时刻残差的零集延拓，成本正比于曲线长度而不是可见域面积。
	// GreatestTimeStep controls the spacing of greatest-eclipse time isolines. They are off by
	// default (zero or negative draws none), positive values below one minute use one minute,
	// and solarEclipseMapDefaultGreatestTimeStep is the spacing of NASA world maps. Each isochrone
	// continues the zero set of a fixed-instant residual, so its cost scales with curve length
	// rather than with the visible area.
	GreatestTimeStep time.Duration
	// MagnitudeValues 是要绘制的地方最大食分等值线电平；nil 使用 NASA 全球图常用的
	// 0.2/0.4/0.6/0.8，显式空切片关闭，非空切片按给定电平绘制。
	// MagnitudeValues lists the local maximum-magnitude contour levels to draw. Nil uses the
	// 0.2/0.4/0.6/0.8 set common to NASA world maps, an explicitly empty slice disables them,
	// and a non-empty slice draws exactly those levels.
	MagnitudeValues []float64
}

type solarEclipseMapCalculators struct {
	global  func(time.Time) (eclipsecore.SolarEclipseInfo, bool)
	panel   func(time.Time) (eclipsecore.SolarEclipseGeocentricPanel, bool)
	partial func(time.Time, eclipsecore.SolarEclipsePartialFootprintOptions) (eclipsecore.SolarEclipsePartialFootprintsInfo, bool)
	central func(time.Time, eclipsecore.SolarEclipsePathOptions) (eclipsecore.SolarEclipsePath, bool)
	local   func(time.Time, float64, float64, float64) (eclipsecore.LocalSolarEclipseInfo, bool)
}

// SolarEclipseMapSVG 使用 NASA bulletin Split-K 绘制完整偏食可见范围，并在存在时绘制全食或环食中心线。
// SolarEclipseMapSVG renders the full partial-visibility sweep and, when present, the total or annular central path using NASA bulletin Split-K.
func SolarEclipseMapSVG(date time.Time, options SolarEclipseMapSVGOptions) (string, bool) {
	return SolarEclipseMapSVGNASABulletinSplitK(date, options)
}

// SolarEclipseMapSVGNASABulletinSplitK 使用 NASA bulletin Split-K 绘制地图。
// SolarEclipseMapSVGNASABulletinSplitK renders a NASA bulletin Split-K map.
func SolarEclipseMapSVGNASABulletinSplitK(date time.Time, options SolarEclipseMapSVGOptions) (string, bool) {
	return solarEclipseMapSVG(date, options, solarEclipseMapCalculators{
		global:  eclipsecore.SolarEclipseOnDateNASABulletinSplitK,
		panel:   eclipsecore.SolarEclipseGeocentricPanelAt,
		partial: eclipsecore.SolarEclipsePartialFootprintsNASABulletinSplitK,
		central: eclipsecore.SolarEclipseCentralPathNASABulletinSplitK,
		local:   eclipsecore.GeometricLocalSolarEclipseOnDateNASABulletinSplitK,
	})
}

// SolarEclipseMapSVGIAUSingleK 使用 IAU Single-K 模型绘制地图。
// SolarEclipseMapSVGIAUSingleK renders an IAU Single-K map.
func SolarEclipseMapSVGIAUSingleK(date time.Time, options SolarEclipseMapSVGOptions) (string, bool) {
	return solarEclipseMapSVG(date, options, solarEclipseMapCalculators{
		global:  eclipsecore.SolarEclipseOnDateIAUSingleK,
		panel:   eclipsecore.SolarEclipseGeocentricPanelIAUSingleK,
		partial: eclipsecore.SolarEclipsePartialFootprintsIAUSingleK,
		central: eclipsecore.SolarEclipseCentralPathIAUSingleK,
		local:   eclipsecore.GeometricLocalSolarEclipseOnDateIAUSingleK,
	})
}

func solarEclipseMapSVG(
	date time.Time,
	options SolarEclipseMapSVGOptions,
	calculators solarEclipseMapCalculators,
) (string, bool) {
	if !validEclipseMapProjection(options.Projection) {
		return "", false
	}
	if options.TimeScale == astro.TimeScaleUT1 {
		if options.Location != nil && options.Location != time.UTC {
			return "", false
		}
		options.Location = time.UTC
	}
	options = normalizeSolarEclipseMapSVGOptions(date, options)
	// 日期门与核心一致：当天没有日食就不出图。偏食足迹是“取最近一次”语义，
	// 缺这道门会把邻近日期的图当成当天的图交出去。
	global, ok := calculators.global(date)
	if !ok {
		return "", false
	}
	greatestTimes := solarEclipseGreatestTimeLevels(global, options)
	partial, ok := calculators.partial(date, eclipsecore.SolarEclipsePartialFootprintOptions{
		Step:               options.PartialStep,
		BoundaryPoints:     options.BoundaryPoints,
		CentralShadowStep:  options.CentralShadowStep,
		GreatestTimeValues: greatestTimes,
		MagnitudeValues:    options.MagnitudeValues,
	})
	if !ok {
		return "", false
	}
	central, hasCentral := calculators.central(date, eclipsecore.SolarEclipsePathOptions{
		Step:            options.CentralStep,
		TargetSpacingKM: options.TargetSpacingKM,
		SkipCentralBand: true,
	})
	local, hasLocal := calculators.local(
		partial.Eclipse.GreatestEclipse,
		partial.Eclipse.GreatestLongitude,
		partial.Eclipse.GreatestLatitude,
		0,
	)
	projection := resolveSolarEclipseMapProjection(partial, central, hasCentral, options.Projection)
	// 正射图的视点固定取食甚点，让整条中心食带尽量落在可见半球内。
	center := svgmap.GeoPoint{
		Longitude: partial.Eclipse.GreatestLongitude,
		Latitude:  partial.Eclipse.GreatestLatitude,
	}
	geocentric, hasGeocentric := eclipsecore.SolarEclipseGeocentricPanel{}, false
	if calculators.panel != nil {
		geocentric, hasGeocentric = calculators.panel(date)
	}
	// 几何（直射点、闭合弧、等时线取值）必须用民用时刻算；只有写出的文字换时标。
	// 时刻标记按输出时标的整点取点（那里本就是另一个物理时刻），位置随整点标注移动。
	civilPartial := partial
	if options.TimeScale == astro.TimeScaleUT1 {
		global = eclipsecore.SolarEclipseInfoInUT1(global)
		greatestTimes = eclipsecore.TimeLabelsInUT1(greatestTimes)
		partial = eclipsecore.SolarEclipsePartialFootprintsInUT1(partial)
		central = eclipsecore.SolarEclipsePathInUT1(central)
		local = eclipsecore.LocalSolarEclipseInfoInUT1(local)
		geocentric = eclipsecore.SolarEclipseGeocentricPanelInUT1(geocentric)
	}
	plan := solarEclipseMapPlanFor(partial, local, hasLocal, geocentric, hasGeocentric,
		options, projection, center, hasCentral, solarEclipseMapHasCentralBand(partial))
	// 数据块行距或图例带放不下时拒绝该画布，而不是把压叠的文字画出来。
	if !plan.fits() {
		return "", false
	}
	return plan.render(civilPartial, partial, central, hasCentral,
		local, hasLocal, geocentric, hasGeocentric, options, projection), true
}

// solarEclipseMapHasCentralBand 报告该事件是否有中心食带（包络、限界或瞬时足迹任一存在）。
func solarEclipseMapHasCentralBand(partial eclipsecore.SolarEclipsePartialFootprintsInfo) bool {
	return len(partial.CentralBandFootprints) > 0 || len(partial.CentralShadowFootprints) > 0 ||
		len(partial.CentralBandSegments) > 0
}

// solarEclipseMapPlan 是渲染前的版面：数据块、图例分行与版式几何，由一处分块逻辑同时供占位与绘制。
type solarEclipseMapPlan struct {
	cells        []solarEclipsePanelCell
	layout       solarEclipseMapLayout
	legend       [][]solarEclipseMapLegendItem
	legendBottom float64
	footerTop    float64
}

func solarEclipseMapPlanFor(
	partial eclipsecore.SolarEclipsePartialFootprintsInfo,
	local eclipsecore.LocalSolarEclipseInfo,
	hasLocal bool,
	geocentric eclipsecore.SolarEclipseGeocentricPanel,
	hasGeocentric bool,
	options SolarEclipseMapSVGOptions,
	projection svgmap.Projection,
	center svgmap.GeoPoint,
	hasCentral bool,
	hasCentralBand bool,
) solarEclipseMapPlan {
	blocks := solarEclipseDetailedBlocks(partial, local, hasLocal, geocentric, hasGeocentric, options)
	legend := solarEclipseMapLegendRows(
		solarEclipseMapLegendItems(partial, hasCentral, hasCentralBand, options),
		solarEclipseMapLegendAvailableWidth(options))
	layout := solarEclipseMapLayoutForBlocks(options, projection, center, len(legend), solarEclipseLandscapeRowFields(blocks))
	_, legendBelow := svgchart.EstimatedTextExtents(solarEclipseLegendFontSize)
	footerAbove, _ := svgchart.EstimatedTextExtents(solarEclipseMapFooterFontSize)
	return solarEclipseMapPlan{
		cells:        solarEclipsePanelCells(layout, blocks),
		layout:       layout,
		legend:       legend,
		legendBottom: layout.legendY + float64(len(legend)-1)*solarEclipseLegendLineStep + legendBelow,
		footerTop:    float64(options.Height) - solarEclipseMapFooterBaselineInset - footerAbove,
	}
}

// fits 报告版面上的数据块与图例带是否都放得下。
func (plan solarEclipseMapPlan) fits() bool {
	if solarEclipsePanelsOverlap(plan.cells) {
		return false
	}
	if len(plan.legend) == 0 {
		return true
	}
	return plan.legendBottom < plan.footerTop
}

// reserveFixed 先占位不动的元素：数据块、图例、比例尺槽、页脚与圆盘四向标记。
func (plan solarEclipseMapPlan) reserveFixed(labels *svgchart.LabelTable, options SolarEclipseMapSVGOptions, projection svgmap.Projection) {
	for _, cell := range plan.cells {
		labels.Reserve(cell.box.X, cell.box.Y, cell.box.Width, cell.box.Height)
	}
	legendX := plan.layout.frame.X
	if plan.layout.nasa {
		legendX = plan.layout.panelX
	}
	for row, items := range plan.legend {
		y := plan.layout.legendY + float64(row)*solarEclipseLegendLineStep
		x := legendX
		for _, item := range items {
			width := solarEclipseLegendIconWidth + svgchart.EstimatedTextWidth(item.label, solarEclipseLegendFontSize)
			labels.Reserve(x, y-9, width, 12)
			x += width + solarEclipseLegendColumnGap
		}
	}
	labels.Reserve(plan.layout.scaleSlot.X, plan.layout.scaleSlot.Y, plan.layout.scaleSlot.Width, plan.layout.scaleSlot.Height)
	labels.ReserveText(plan.layout.margin, float64(options.Height)-solarEclipseMapFooterBaselineInset,
		solarEclipseMapFooterFontSize, solarEclipseMapFooterText(options, projection), "start")
	for _, cardinal := range solarEclipseCardinalLabels(plan.layout) {
		labels.ReserveText(cardinal.x, cardinal.y, 13, cardinal.text, "middle")
	}
}

func normalizeSolarEclipseMapSVGOptions(date time.Time, options SolarEclipseMapSVGOptions) SolarEclipseMapSVGOptions {
	if options.Width < solarEclipseMapMinWidth {
		options.Width = solarEclipseMapDefaultWidth
	}
	if options.Height < solarEclipseMapMinHeight {
		options.Height = solarEclipseMapDefaultHeight
	}
	if strings.EqualFold(options.Language, "en") {
		options.Language = "en"
	} else {
		options.Language = "zh"
	}
	if options.Location == nil {
		options.Location = date.Location()
	}
	// 非正值与小于两分钟的正值都取两分钟：并集需要一整条自洽的扫描序列。
	if options.PartialStep < solarEclipseMapPartialSweepMinStep {
		options.PartialStep = solarEclipseMapPartialSweepMinStep
	}
	if options.BoundaryPoints <= 0 {
		options.BoundaryPoints = 180
	}
	// 瞬时半影与本影轮廓默认不画：它们会把地球盖住，NASA 的全球图也没有这两族。
	if options.PenumbralOutlineStep > 0 && options.PenumbralOutlineStep < time.Minute {
		options.PenumbralOutlineStep = time.Minute
	} else if options.PenumbralOutlineStep < 0 {
		options.PenumbralOutlineStep = 0
	}
	if options.CentralShadowStep > 0 && options.CentralShadowStep < time.Minute {
		options.CentralShadowStep = time.Minute
	} else if options.CentralShadowStep < 0 {
		options.CentralShadowStep = 0
	}
	if options.MagnitudeValues == nil {
		options.MagnitudeValues = []float64{0.2, 0.4, 0.6, 0.8}
	}
	if options.CentralStep <= 0 {
		options.CentralStep = 2 * time.Minute
	}
	if options.TargetSpacingKM <= 0 {
		options.TargetSpacingKM = 150
	}
	if options.TimeLabelStep < 0 {
		options.TimeLabelStep = 0
	} else if options.TimeLabelStep == 0 {
		options.TimeLabelStep = 30 * time.Minute
	} else if options.TimeLabelStep < time.Minute {
		options.TimeLabelStep = time.Minute
	}
	// 零值与负值都表示不请求等时线：这是可选图层，与核心层和 moon/svg 一样必须显式请求。
	if options.GreatestTimeStep < 0 {
		options.GreatestTimeStep = 0
	} else if options.GreatestTimeStep > 0 && options.GreatestTimeStep < time.Minute {
		options.GreatestTimeStep = time.Minute
	}
	return options
}

// solarEclipseGreatestTimeLevels 把等时线时刻取值对齐到步长网格上，覆盖地球范围偏食窗口。
func solarEclipseGreatestTimeLevels(
	info eclipsecore.SolarEclipseInfo,
	options SolarEclipseMapSVGOptions,
) []time.Time {
	if options.GreatestTimeStep <= 0 {
		return nil
	}
	if info.PartialBeginOnEarth.IsZero() || info.PartialEndOnEarth.IsZero() {
		return nil
	}
	current := solarEclipseMapAlignedTime(info.PartialBeginOnEarth, options.GreatestTimeStep, options.Location)
	if current.Before(info.PartialBeginOnEarth) {
		current = current.Add(options.GreatestTimeStep)
	}
	levels := make([]time.Time, 0, solarEclipseMapGreatestTimeMaxLevels)
	for !current.After(info.PartialEndOnEarth) && len(levels) < solarEclipseMapGreatestTimeMaxLevels {
		levels = append(levels, current)
		current = current.Add(options.GreatestTimeStep)
	}
	return levels
}

func resolveSolarEclipseMapProjection(
	partial eclipsecore.SolarEclipsePartialFootprintsInfo,
	central eclipsecore.SolarEclipsePath,
	hasCentral bool,
	requested EclipseMapProjection,
) svgmap.Projection {
	if requested != EclipseMapProjectionAuto {
		return internalEclipseMapProjection(requested)
	}
	focus := partial.Eclipse.GreatestLatitude
	minimum, maximum := focus, focus
	for _, footprint := range partial.Footprints {
		for _, boundary := range footprint.Boundaries {
			for _, point := range boundary {
				minimum = math.Min(minimum, point.Latitude)
				maximum = math.Max(maximum, point.Latitude)
			}
		}
	}
	if hasCentral {
		for _, series := range [][]eclipsecore.SolarEclipsePathPoint{central.CenterLine, central.NorthernLimit, central.SouthernLimit} {
			for _, point := range series {
				minimum = math.Min(minimum, point.Latitude)
				maximum = math.Max(maximum, point.Latitude)
			}
		}
	}
	// 高纬但跨越赤道的事件必须保留等距圆柱视图；极区视图只由半球守卫决定。
	return svgmap.ResolveProjection("", focus, minimum, maximum)
}

func renderSolarEclipseMapSVG(
	partial eclipsecore.SolarEclipsePartialFootprintsInfo,
	central eclipsecore.SolarEclipsePath,
	hasCentral bool,
	local eclipsecore.LocalSolarEclipseInfo,
	hasLocal bool,
	geocentric eclipsecore.SolarEclipseGeocentricPanel,
	hasGeocentric bool,
	options SolarEclipseMapSVGOptions,
	projection svgmap.Projection,
	center svgmap.GeoPoint,
) string {
	plan := solarEclipseMapPlanFor(partial, local, hasLocal, geocentric, hasGeocentric,
		options, projection, center, hasCentral, solarEclipseMapHasCentralBand(partial))
	return plan.render(partial, partial, central, hasCentral, local, hasLocal, geocentric, hasGeocentric, options, projection)
}

func (plan solarEclipseMapPlan) render(
	civilPartial eclipsecore.SolarEclipsePartialFootprintsInfo,
	partial eclipsecore.SolarEclipsePartialFootprintsInfo,
	central eclipsecore.SolarEclipsePath,
	hasCentral bool,
	local eclipsecore.LocalSolarEclipseInfo,
	hasLocal bool,
	geocentric eclipsecore.SolarEclipseGeocentricPanel,
	hasGeocentric bool,
	options SolarEclipseMapSVGOptions,
	projection svgmap.Projection,
) string {
	layout := plan.layout
	frame := layout.frame
	title := solarEclipseMapTitle(partial.Eclipse, options)
	// 图题居中，两侧各留出白色页框 22 px 与一点空隙。
	titleText := title
	if options.Title != "" {
		titleText = svgchart.EllipsizeText(title, float64(options.Width)-52, 24)
	}
	partialPath, partialSource := solarEclipsePartialSweepPath(civilPartial, frame)

	// 一张已放置矩形表：固定元素先占位，地图标注再按候选偏移避让。
	labels := &svgchart.LabelTable{}
	plan.reserveFixed(labels, options, projection)
	labels.ReserveText(float64(options.Width)/2, 47, 24, titleText, "middle")

	var builder strings.Builder
	fmt.Fprintf(&builder, `<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" viewBox="0 0 %d %d" role="img" aria-label="%s">`,
		options.Width, options.Height, options.Width, options.Height, html.EscapeString(title))
	builder.WriteString(`<defs>`)
	builder.WriteString(frame.ClipDefinition("solar-map-clip"))
	builder.WriteString(`</defs>`)
	builder.WriteString(`<rect width="100%" height="100%" fill="#efefed"/>`)
	fmt.Fprintf(&builder, `<rect x="22" y="18" width="%d" height="%d" fill="#ffffff" stroke="#c9c9c6" stroke-width="1.2"/>`,
		options.Width-44, options.Height-36)
	fmt.Fprintf(&builder, `<text x="%.3f" y="47" fill="#111111" font-family="Georgia, 'Times New Roman', serif" font-size="24" font-weight="700" text-anchor="middle">%s</text>`,
		float64(options.Width)/2, html.EscapeString(titleText))
	writeSolarEclipseMapSummary(&builder, labels, partial.Eclipse, local, hasLocal, geocentric, hasGeocentric, options)
	writeSolarEclipseMapSectionTitle(&builder, labels, layout, options, hasCentral)

	frame.WriteOcean(&builder)
	frame.WriteGraticule(&builder, "solar-map-clip")
	frame.WriteLand(&builder, "solar-map-clip")
	fmt.Fprintf(&builder, `<path class="partial-eclipse-region" data-source="%s" d="%s" clip-path="url(#solar-map-clip)" fill="#f7dd93" fill-opacity="0.16" fill-rule="nonzero"/>`, partialSource, partialPath)
	writeSolarEclipsePartialBoundary(&builder, partial.PartialBandContours, frame)
	writeSolarEclipseRiseSetCurves(&builder, partial.RiseSetCurves, frame)
	writeSolarEclipsePenumbralOutlines(&builder, partial, frame, options, labels)
	// 点标注比等值线标注重要：先把食甚、接触、直射点与中心线时刻占到位置，等值线标注再让开。
	pointLabels := solarEclipseMapPlacePointLabels(civilPartial, partial, central, hasCentral, frame, options, labels)
	writeSolarEclipseGreatestTimeContours(&builder, partial.GreatestTimeContours, frame, options, labels)
	magnitudeAxis, magnitudeNormal, hasMagnitudeAxis := solarEclipseMagnitudeLabelAxis(central)
	if !hasMagnitudeAxis {
		magnitudeAxis, magnitudeNormal = [3]float64{}, [3]float64{}
	}
	writeSolarEclipseMagnitudeContours(&builder, partial.MagnitudeContours, frame, labels,
		magnitudeAxis, magnitudeNormal)
	centralBandFootprints := partial.CentralBandFootprints
	if len(centralBandFootprints) == 0 {
		centralBandFootprints = partial.CentralShadowFootprints
	}
	if hasCentral {
		writeSolarEclipseCentralPath(
			&builder, central, centralBandFootprints, partial.CentralBandSegments, frame, partial.Eclipse.Type,
		)
	} else if !writeSolarEclipseCentralBandEnvelope(&builder, partial.CentralBandSegments, frame, partial.Eclipse.Type, false) {
		writeSolarEclipseCentralShadowSweep(&builder, centralBandFootprints, frame, partial.Eclipse.Type)
	}
	// Non-central events are represented by the continuous envelope band. The
	// instantaneous open shadow arcs are diagnostic samples, not an additional
	// boundary; drawing them here creates interior lines that visibly diverge
	// from the annular-band edge near the grazing end.
	if options.CentralShadowStep > 0 && partial.Eclipse.Centrality != eclipsecore.SolarEclipseNonCentral {
		writeSolarEclipseCentralShadowOutlines(&builder, partial.CentralShadowFootprints, frame)
	}
	if hasCentral {
		writeSolarEclipseTimeMarkers(&builder, pointLabels.times)
		writeSolarEclipseAxisMarkers(&builder, central, frame, options)
	}
	writeSolarEclipseContactMarkers(&builder, pointLabels.contacts)
	writeSolarEclipseGreatestMarker(&builder, pointLabels.greatest)
	writeSolarEclipseSubsolarMarker(&builder, pointLabels.subsolar)
	frame.WriteFrame(&builder)
	writeSolarEclipseMapLegend(&builder, layout, plan.legend)
	switch {
	case layout.landscape:
		writeSolarEclipseLandscapePanels(&builder, plan.cells, layout, options)
	case layout.nasa:
		writeSolarEclipseCardinalMarkers(&builder, layout)
		writeSolarEclipseScaleBar(&builder, layout, options)
		writeSolarEclipseDetailedPanels(&builder, plan.cells, layout)
		if hasGeocentric {
			writeSolarEclipseGeocentricBlocks(&builder, geocentric, layout, options)
			writeSolarEclipseEphemerisPanels(&builder, plan.cells, layout)
		}
	}
	writeSolarEclipseMapFooter(&builder, layout, options, projection)
	builder.WriteString(`</svg>`)
	return builder.String()
}

// solarEclipsePartialSweepPath 返回偏食可见域填充路径及其几何来源。
func solarEclipsePartialSweepPath(info eclipsecore.SolarEclipsePartialFootprintsInfo, frame svgmap.Frame) (string, string) {
	if polygons, ok := solarEclipsePartialBandPolygons(info); ok {
		var builder strings.Builder
		for _, polygon := range polygons {
			points := make([]svgmap.GeoPoint, len(polygon))
			for index, point := range polygon {
				points[index] = svgmap.GeoPoint{Longitude: point.Longitude, Latitude: point.Latitude}
			}
			for _, fragment := range svgmap.PolygonFragments(points, frame.Clip()) {
				appendEclipseMapPolygonPathConsistent(&builder, frame, fragment)
			}
		}
		return builder.String(), eclipseMapSourcePartialBandUnion
	}
	return solarEclipsePartialFootprintSweepPath(info, frame), eclipseMapSourceSampledFootprintSweep
}

func solarEclipsePartialBandPolygons(
	info eclipsecore.SolarEclipsePartialFootprintsInfo,
) ([][]geodata.GeoPoint, bool) {
	// 缺少精确擦地点时并集失去权威性：退回逐足迹扫掠，data-source 落到 sampled-footprint-sweep。
	if len(info.PartialBandContours) == 0 || len(info.RiseSetCurves) == 0 ||
		!solarEclipsePartialClosureExact(info.Footprints) {
		return nil, false
	}
	contours := make([][]geodata.GeoPoint, 0, len(info.PartialBandContours))
	for _, contour := range info.PartialBandContours {
		contours = append(contours, solarClosureGeoPoints(contour))
	}
	phaseLines := make([][]geodata.GeoPoint, 0, len(info.RiseSetCurves)*2)
	for _, curve := range info.RiseSetCurves {
		for _, segment := range curve.Segments {
			phaseLines = append(phaseLines, solarClosureGeoPoints(segment))
		}
	}
	footprints := make([]solarclosure.Footprint, 0, len(info.Footprints))
	for _, footprint := range info.Footprints {
		footprints = append(footprints, solarClosureFootprint(footprint))
	}
	return solarclosure.BandPolygons(
		contours, phaseLines, footprints, true, solarclosure.SnapDistanceKM,
	)
}

// solarEclipsePartialClosureExact 报告每个开放足迹是否都有两个精确擦地点。
func solarEclipsePartialClosureExact(footprints []eclipsecore.SolarEclipsePartialFootprint) bool {
	for _, footprint := range footprints {
		// 自身闭合的足迹不需要擦地点。
		if footprint.Closed {
			continue
		}
		if len(footprint.HorizonEnds) != 2 {
			return false
		}
	}
	return true
}

func solarClosureGeoPoints(points []eclipsecore.SolarEclipsePathPoint) []geodata.GeoPoint {
	result := make([]geodata.GeoPoint, len(points))
	for index, point := range points {
		result[index] = geodata.GeoPoint{Longitude: point.Longitude, Latitude: point.Latitude}
	}
	return result
}

// solarClosureFootprint 转换核心层瞬时足迹；擦地点点数不是 2 时按缺失处理。
func solarClosureFootprint(footprint eclipsecore.SolarEclipsePartialFootprint) solarclosure.Footprint {
	segments := make([][]geodata.GeoPoint, len(footprint.Boundaries))
	for index, source := range footprint.Boundaries {
		segments[index] = solarClosureGeoPoints(source)
	}
	ends := make([]geodata.GeoPoint, 0, 2)
	if len(footprint.HorizonEnds) == 2 {
		for _, end := range footprint.HorizonEnds {
			ends = append(ends, geodata.GeoPoint{Longitude: end.Longitude, Latitude: end.Latitude})
		}
	}
	subsolar := solarEclipseSubsolarPoint(footprint.Time)
	return solarclosure.Footprint{
		Boundaries:  segments,
		HorizonEnds: ends,
		Subsolar:    geodata.GeoPoint{Longitude: subsolar.Longitude, Latitude: subsolar.Latitude},
		Closed:      footprint.Closed,
	}
}

func solarEclipsePartialFootprintSweepPath(info eclipsecore.SolarEclipsePartialFootprintsInfo, frame svgmap.Frame) string {
	var builder strings.Builder
	for _, footprint := range info.Footprints {
		polygon := solarEclipsePartialFootprintPolygon(footprint, frame.Projection)
		for _, fragment := range svgmap.PolygonFragments(polygon, frame.Clip()) {
			appendEclipseMapPolygonPathConsistent(&builder, frame, fragment)
		}
	}
	return builder.String()
}

// solarEclipsePartialFootprintPolygon 返回瞬时足迹的填充多边形：开放边界有精确擦地点时用它闭合，
// 否则退回 subsolar 地平圈的近似弧。
func solarEclipsePartialFootprintPolygon(
	footprint eclipsecore.SolarEclipsePartialFootprint,
	projection svgmap.Projection,
) []svgmap.GeoPoint {
	polygon, _, ok := solarclosure.Ring(solarClosureFootprint(footprint), true)
	if !ok {
		return nil
	}
	if interior, ok := solarEclipseSphericalBoundaryCentroid(polygon); ok && projection == svgmap.ProjectionEquirectangular {
		polygon = solarEclipseAppendEquirectangularPoleRim(polygon, interior)
	}
	return polygon
}

func solarEclipseSphericalBoundaryCentroid(points []svgmap.GeoPoint) (svgmap.GeoPoint, bool) {
	var sum solarEclipseMapVector
	for _, point := range points {
		value := solarEclipseMapUnitVector(point)
		sum[0] += value[0]
		sum[1] += value[1]
		sum[2] += value[2]
	}
	length := math.Sqrt(solarEclipseMapDot(sum, sum))
	if length < 1e-12 {
		return svgmap.GeoPoint{}, false
	}
	for index := range sum {
		sum[index] /= length
	}
	return svgmap.GeoPoint{
		Longitude: math.Atan2(sum[1], sum[0]) * 180 / math.Pi,
		Latitude:  math.Asin(math.Max(-1, math.Min(1, sum[2]))) * 180 / math.Pi,
	}, true
}

func solarEclipseAppendEquirectangularPoleRim(
	polygon []svgmap.GeoPoint,
	interior svgmap.GeoPoint,
) []svgmap.GeoPoint {
	interiorAngle := solarEclipseSphericalPolygonAngle(polygon, interior)
	if math.Abs(interiorAngle) < math.Pi {
		return polygon
	}
	poleLatitude := 0.0
	for _, candidate := range []float64{90, -90} {
		angle := solarEclipseSphericalPolygonAngle(polygon, svgmap.GeoPoint{Latitude: candidate})
		if math.Abs(angle) >= math.Pi && math.Signbit(angle) == math.Signbit(interiorAngle) {
			poleLatitude = candidate
			break
		}
	}
	if poleLatitude == 0 {
		return polygon
	}

	firstLongitude, lastLongitude := solarEclipseUnwrappedLongitudeEndpoints(polygon)
	delta := firstLongitude - lastLongitude
	if math.Abs(delta) < 180 {
		return polygon
	}
	steps := int(math.Ceil(math.Abs(delta) / 90))
	result := append([]svgmap.GeoPoint(nil), polygon...)
	for step := 1; step <= steps; step++ {
		longitude := lastLongitude + delta*float64(step)/float64(steps)
		result = append(result, svgmap.GeoPoint{
			Longitude: normalizeDegree180(longitude),
			Latitude:  poleLatitude,
		})
	}
	return result
}

func solarEclipseUnwrappedLongitudeEndpoints(points []svgmap.GeoPoint) (float64, float64) {
	first := points[0].Longitude
	previous := first
	for _, point := range points[1:] {
		longitude := point.Longitude
		for longitude-previous > 180 {
			longitude -= 360
		}
		for longitude-previous < -180 {
			longitude += 360
		}
		previous = longitude
	}
	return first, previous
}

type solarEclipseMapVector [3]float64

func solarEclipseSphericalPolygonAngle(points []svgmap.GeoPoint, target svgmap.GeoPoint) float64 {
	if len(points) < 3 {
		return 0
	}
	reference := solarEclipseMapUnitVector(target)
	angle := 0.0
	for index, point := range points {
		current, currentOK := solarEclipseMapTangentDirection(reference, solarEclipseMapUnitVector(point))
		next, nextOK := solarEclipseMapTangentDirection(reference, solarEclipseMapUnitVector(points[(index+1)%len(points)]))
		if !currentOK || !nextOK {
			return math.Copysign(2*math.Pi, angle)
		}
		angle += math.Atan2(
			solarEclipseMapDot(reference, solarEclipseMapCross(current, next)),
			solarEclipseMapDot(current, next),
		)
	}
	return angle
}

func solarEclipseMapUnitVector(point svgmap.GeoPoint) solarEclipseMapVector {
	longitude := point.Longitude * math.Pi / 180
	latitude := point.Latitude * math.Pi / 180
	return solarEclipseMapVector{
		math.Cos(latitude) * math.Cos(longitude),
		math.Cos(latitude) * math.Sin(longitude),
		math.Sin(latitude),
	}
}

func solarEclipseMapTangentDirection(
	reference, value solarEclipseMapVector,
) (solarEclipseMapVector, bool) {
	projection := solarEclipseMapDot(reference, value)
	result := solarEclipseMapVector{
		value[0] - projection*reference[0],
		value[1] - projection*reference[1],
		value[2] - projection*reference[2],
	}
	length := math.Sqrt(solarEclipseMapDot(result, result))
	if length < 1e-12 {
		return solarEclipseMapVector{}, false
	}
	return solarEclipseMapVector{result[0] / length, result[1] / length, result[2] / length}, true
}

func solarEclipseMapDot(a, b solarEclipseMapVector) float64 {
	return a[0]*b[0] + a[1]*b[1] + a[2]*b[2]
}

func solarEclipseMapCross(a, b solarEclipseMapVector) solarEclipseMapVector {
	return solarEclipseMapVector{
		a[1]*b[2] - a[2]*b[1],
		a[2]*b[0] - a[0]*b[2],
		a[0]*b[1] - a[1]*b[0],
	}
}

func solarEclipseSubsolarPoint(value time.Time) svgmap.GeoPoint {
	ttJDE := solarEclipseTimeToTTJDE(value)
	ra, dec := basic.HSunApparentRaDec(ttJDE)
	ut1JDE := basic.TT2UT1(ttJDE)
	longitude := normalizeDegree180(ra - basic.ApparentSiderealTime(ut1JDE)*15)
	return svgmap.GeoPoint{Longitude: longitude, Latitude: dec}
}

func appendEclipseMapPolygonPathConsistent(builder *strings.Builder, frame svgmap.Frame, points []svgmap.GeoPoint) {
	if projectedPolygonArea(frame, points) < 0 {
		points = append([]svgmap.GeoPoint(nil), points...)
		reverseGeoPoints(points)
	}
	appendEclipseMapPolygonPath(builder, frame, points)
}

func projectedPolygonArea(frame svgmap.Frame, points []svgmap.GeoPoint) float64 {
	area := 0.0
	for index, point := range points {
		next := points[(index+1)%len(points)]
		x1, y1, ok1 := frame.Project(point.Longitude, point.Latitude)
		x2, y2, ok2 := frame.Project(next.Longitude, next.Latitude)
		if ok1 && ok2 {
			area += x1*y2 - x2*y1
		}
	}
	return area / 2
}

func writeSolarEclipseCentralPath(
	builder *strings.Builder,
	path eclipsecore.SolarEclipsePath,
	bandFootprints []eclipsecore.SolarEclipsePartialFootprint,
	bandSegments [][]eclipsecore.SolarEclipsePathPoint,
	frame svgmap.Frame,
	eclipseType eclipsecore.SolarEclipseType,
) {
	northern := path.NorthernLimit
	southern := path.SouthernLimit
	physicalEndpointSweeps := false
	if path.Eclipse.Centrality == eclipsecore.SolarEclipseCentralTwoLimits {
		if north, south, ok := solarEclipseTwoLimitPresentationLimits(path); ok {
			northern, southern = north, south
			physicalEndpointSweeps = true
		}
	}
	count := len(northern)
	paired := count == len(southern) && count >= 2
	for index := 0; paired && index < count; index++ {
		paired = !northern[index].Time.IsZero() && northern[index].Time.Equal(southern[index].Time)
	}
	coveragePath := path
	coveragePath.NorthernLimit = northern
	coveragePath.SouthernLimit = southern
	if len(bandSegments) > 0 && !solarEclipseCentralEnvelopeCoversPath(bandSegments, coveragePath) {
		bandSegments = nil
	}
	if !writeSolarEclipseCentralBandEnvelope(builder, bandSegments, frame, eclipseType, true) && paired {
		capacity := 2 * count
		if physicalEndpointSweeps {
			capacity += 2
		}
		polygon := make([]svgmap.GeoPoint, 0, capacity)
		if physicalEndpointSweeps {
			point := path.CenterLine[0]
			polygon = append(polygon, svgmap.GeoPoint{Longitude: point.Longitude, Latitude: point.Latitude})
		}
		for _, point := range northern[:count] {
			polygon = append(polygon, svgmap.GeoPoint{Longitude: point.Longitude, Latitude: point.Latitude})
		}
		if physicalEndpointSweeps {
			point := path.CenterLine[len(path.CenterLine)-1]
			polygon = append(polygon, svgmap.GeoPoint{Longitude: point.Longitude, Latitude: point.Latitude})
		}
		for index := count - 1; index >= 0; index-- {
			point := southern[index]
			polygon = append(polygon, svgmap.GeoPoint{Longitude: point.Longitude, Latitude: point.Latitude})
		}
		polygons := [][]svgmap.GeoPoint{polygon}
		if physicalEndpointSweeps {
			if merged, ok := solarEclipseTwoLimitBandPolygons(
				path, northern, southern, bandFootprints,
			); ok {
				polygons = merged
			}
		}
		color := solarEclipseCentralPathColor(eclipseType)
		fmt.Fprintf(builder, `<g class="central-eclipse-band" data-source="%s" clip-path="url(#solar-map-clip)" fill="%s" fill-opacity="0.30">`, eclipseMapSourcePairedLimitChords, color)
		for _, merged := range polygons {
			for _, fragment := range svgmap.PolygonFragments(merged, frame.Clip()) {
				builder.WriteString(`<path d="`)
				appendEclipseMapPolygonPathConsistent(builder, frame, fragment)
				builder.WriteString(`"/>`)
			}
		}
		builder.WriteString(`</g>`)
	}
	writeSolarPathLine(builder, frame, northern, "northern-central-limit", "#7c2f28", 1.2, "")
	writeSolarPathLine(builder, frame, southern, "southern-central-limit", "#7c2f28", 1.2, "")
	writeSolarPathLine(builder, frame, path.CenterLine, "solar-center-line", "#263f58", 1.8, "5 3")
}

func solarEclipseCentralEnvelopeCoversPath(
	segments [][]eclipsecore.SolarEclipsePathPoint,
	path eclipsecore.SolarEclipsePath,
) bool {
	if len(segments) == 0 {
		return false
	}
	polygons := make([][]geodata.GeoPoint, 0, len(segments))
	for _, segment := range segments {
		if len(segment) < 3 {
			return false
		}
		polygon := make([]geodata.GeoPoint, len(segment))
		for index, point := range segment {
			polygon[index] = geodata.GeoPoint{Longitude: point.Longitude, Latitude: point.Latitude}
		}
		polygons = append(polygons, polygon)
	}
	// 限制线与中心线都要落在包络内：真正绘制的就是这三条线。限界两端含掠地端帽采样，
	// 允许端部偏差，但整条限界远离包络说明包络并未覆盖将要绘制的中心带。
	if len(path.CenterLine) < 2 {
		return false
	}
	if solarEclipseCentralEnvelopeMissDistanceKM(polygons, path.CenterLine) >
		solarEclipseCentralEnvelopeAxisMissToleranceKM {
		return false
	}
	limitToleranceKM := solarEclipseCentralEnvelopeAxisMissToleranceKM
	if width := path.Eclipse.PathWidthKM; path.Eclipse.PathWidthDefined && width > 0 && !math.IsInf(width, 1) {
		limitToleranceKM = math.Max(limitToleranceKM, solarEclipseCentralEnvelopeLimitMissFraction*width)
	}
	for _, series := range [][]eclipsecore.SolarEclipsePathPoint{path.NorthernLimit, path.SouthernLimit} {
		if solarEclipseCentralEnvelopeMissDistanceKM(polygons, series) > limitToleranceKM {
			return false
		}
	}
	return true
}

func solarEclipseCentralEnvelopeMissDistanceKM(
	polygons [][]geodata.GeoPoint,
	points []eclipsecore.SolarEclipsePathPoint,
) float64 {
	if len(points) == 0 {
		return 0
	}
	line := make([]geodata.GeoPoint, len(points))
	for index, point := range points {
		line[index] = geodata.GeoPoint{Longitude: point.Longitude, Latitude: point.Latitude}
	}
	return geodata.SphericalPolygonsPathMissDistanceKM(polygons, [][]geodata.GeoPoint{line}, false)
}

func solarEclipseTwoLimitBandPolygons(
	path eclipsecore.SolarEclipsePath,
	northern, southern []eclipsecore.SolarEclipsePathPoint,
	footprints []eclipsecore.SolarEclipsePartialFootprint,
) ([][]svgmap.GeoPoint, bool) {
	if len(northern) < 2 || len(northern) != len(southern) || len(footprints) == 0 {
		return nil, false
	}
	middle := make([]geodata.GeoPoint, 0, len(northern)+len(southern))
	for _, point := range northern {
		middle = append(middle, geodata.GeoPoint{Longitude: point.Longitude, Latitude: point.Latitude})
	}
	for index := len(southern) - 1; index >= 0; index-- {
		point := southern[index]
		middle = append(middle, geodata.GeoPoint{Longitude: point.Longitude, Latitude: point.Latitude})
	}
	endSweeps, err := solarEclipseMonotoneCentralShadowSweepPolygons(footprints)
	if err != nil {
		return nil, false
	}
	inputs := make([][]geodata.GeoPoint, 0, 1+len(endSweeps)+2)
	inputs = append(inputs, middle)
	inputs = append(inputs, endSweeps...)
	inputs = append(inputs, solarEclipseCentralBandInnerTransitionCaps(footprints)...)
	inputs = append(inputs, solarEclipseCentralBandContactCaps(
		footprints, path.NorthernLimit[0], path.NorthernLimit[len(path.NorthernLimit)-1],
	)...)
	merged, err := geodata.UnionPolygons(inputs)
	if err != nil || len(merged) != 1 {
		return nil, false
	}
	polygons := make([][]svgmap.GeoPoint, len(merged))
	for polygonIndex, source := range merged {
		polygon := make([]svgmap.GeoPoint, len(source))
		for pointIndex, point := range source {
			polygon[pointIndex] = svgmap.GeoPoint{Longitude: point.Longitude, Latitude: point.Latitude}
		}
		polygons[polygonIndex] = polygon
	}
	return polygons, true
}

func solarEclipseCentralBandContactCaps(
	footprints []eclipsecore.SolarEclipsePartialFootprint,
	startContact, endContact eclipsecore.SolarEclipsePathPoint,
) [][]geodata.GeoPoint {
	if len(footprints) == 0 {
		return nil
	}
	caps := make([][]geodata.GeoPoint, 0, 2)
	appendCap := func(contact eclipsecore.SolarEclipsePathPoint, footprint eclipsecore.SolarEclipsePartialFootprint) {
		segments := make([][]geodata.GeoPoint, 0, len(footprint.Boundaries))
		for _, source := range footprint.Boundaries {
			segment := make([]geodata.GeoPoint, len(source))
			for index, point := range source {
				segment[index] = geodata.GeoPoint{Longitude: point.Longitude, Latitude: point.Latitude}
			}
			segments = append(segments, segment)
		}
		boundary := geodata.JoinPolylineSegments(segments)
		if len(boundary) > 1 && geodata.SameGeoPoint(boundary[0], boundary[len(boundary)-1]) {
			boundary = boundary[:len(boundary)-1]
		}
		if len(boundary) < 2 {
			return
		}
		caps = append(caps, []geodata.GeoPoint{
			{Longitude: contact.Longitude, Latitude: contact.Latitude},
			boundary[0], boundary[len(boundary)-1],
		})
	}
	appendCap(startContact, footprints[0])
	appendCap(endContact, footprints[len(footprints)-1])
	return caps
}

// The path API retains U1/U4 compatibility samples. Rendered side lines use
// only the paired interval; the band fill adds the sampled physical endpoint
// sweeps separately so the duplicate outer samples cannot form long wedges.
func solarEclipseTwoLimitPresentationLimits(
	path eclipsecore.SolarEclipsePath,
) ([]eclipsecore.SolarEclipsePathPoint, []eclipsecore.SolarEclipsePathPoint, bool) {
	northern, southern, centerLine := path.NorthernLimit, path.SouthernLimit, path.CenterLine
	if len(northern) != len(southern) || len(northern) < 4 || len(centerLine) < 2 {
		return nil, nil, false
	}
	start := centerLine[0].Time
	end := centerLine[len(centerLine)-1].Time
	if start.IsZero() || end.IsZero() || !start.Before(end) ||
		!northern[0].Time.Before(start) || !northern[len(northern)-1].Time.After(end) ||
		!solarEclipseSamePathPoint(northern[0], southern[0]) ||
		!solarEclipseSamePathPoint(northern[len(northern)-1], southern[len(southern)-1]) {
		return nil, nil, false
	}
	first := 0
	for first < len(northern) && !northern[first].Time.After(start) {
		first++
	}
	last := first
	for last < len(northern) && northern[last].Time.Before(end) {
		last++
	}
	if first == 0 || last >= len(northern) || last-first < 2 {
		return nil, nil, false
	}
	for index := first; index < last; index++ {
		if northern[index].Time.IsZero() || !northern[index].Time.Equal(southern[index].Time) {
			return nil, nil, false
		}
	}
	return northern[first:last], southern[first:last], true
}

func solarEclipseSamePathPoint(first, second eclipsecore.SolarEclipsePathPoint) bool {
	return math.Abs(first.Longitude-second.Longitude) <= 1e-9 &&
		math.Abs(first.Latitude-second.Latitude) <= 1e-9
}

func writeSolarPathLine(
	builder *strings.Builder,
	frame svgmap.Frame,
	points []eclipsecore.SolarEclipsePathPoint,
	className, color string,
	width float64,
	dash string,
) {
	geographic := make([]svgmap.GeoPoint, len(points))
	for index, point := range points {
		geographic[index] = svgmap.GeoPoint{Longitude: point.Longitude, Latitude: point.Latitude}
	}
	writeEclipseMapGeoLine(builder, frame, geographic, className, color, width, dash, "solar-map-clip", eclipseMapSourceCentralPathLimits)
}

func writeSolarEclipseGreatestMarker(builder *strings.Builder, label solarEclipseMapPointLabel) {
	if label.text == "" {
		return
	}
	fmt.Fprintf(builder, `<g class="solar-greatest-marker"><circle cx="%.3f" cy="%.3f" r="5" fill="#c44336" stroke="#fff" stroke-width="1.3"/>`,
		label.x, label.y)
	if label.ok {
		fmt.Fprintf(builder, `<text x="%.3f" y="%.3f" fill="#182124" stroke="#fff" stroke-width="3" paint-order="stroke" font-family="Arial, sans-serif" font-size="11" font-weight="700" text-anchor="%s">%s</text>`,
			label.placed.X, label.placed.Y, label.placed.Anchor, html.EscapeString(label.text))
	}
	builder.WriteString(`</g>`)
}

func solarEclipseMapTitle(info eclipsecore.SolarEclipseInfo, options SolarEclipseMapSVGOptions) string {
	if options.Title != "" {
		return options.Title
	}
	date := info.GreatestEclipse.In(options.Location).Format("2006-01-02")
	if options.Language == "en" {
		return fmt.Sprintf("%s %s Global Visibility", date, solarEclipseMapTypeName(info.Type, "en"))
	}
	return fmt.Sprintf("%s %s全球见食图", date, solarEclipseMapTypeName(info.Type, "zh"))
}

func writeSolarEclipseMapSummary(
	builder *strings.Builder,
	labels *svgchart.LabelTable,
	info eclipsecore.SolarEclipseInfo,
	local eclipsecore.LocalSolarEclipseInfo,
	hasLocal bool,
	geocentric eclipsecore.SolarEclipseGeocentricPanel,
	hasGeocentric bool,
	options SolarEclipseMapSVGOptions,
) {
	start := info.PartialBeginOnEarth.In(options.Location)
	maximum := info.GreatestEclipse.In(options.Location)
	end := info.PartialEndOnEarth.In(options.Location)
	zone, _ := maximum.Zone()
	if zone == "" {
		zone = "UTC"
	}
	text := fmt.Sprintf("偏食始 %s | 食甚 %s | 偏食终 %s (%s) | 食分 %.3f | Gamma %.4f",
		start.Format("15:04:05"), maximum.Format("15:04:05"), end.Format("15:04:05"), zone, info.Magnitude, info.Gamma)
	if options.Language == "en" {
		text = fmt.Sprintf("Partial begins %s | Greatest %s | Partial ends %s (%s) | magnitude %.3f | Gamma %.4f",
			start.Format("15:04:05"), maximum.Format("15:04:05"), end.Format("15:04:05"), zone, info.Magnitude, info.Gamma)
	}
	summaryX := float64(options.Width) / 2
	labels.ReserveText(summaryX, 82, 13, text, "middle")
	fmt.Fprintf(builder, `<text x="%.3f" y="82" fill="#293235" font-family="Arial, sans-serif" font-size="13" text-anchor="middle">%s</text>`,
		summaryX, html.EscapeString(text))

	details := make([]string, 0, 4)
	if info.PathWidthDefined {
		if options.Language == "en" {
			details = append(details, fmt.Sprintf("central path width %.1f km", info.PathWidthKM))
		} else {
			details = append(details, fmt.Sprintf("中心食带宽 %.1f km", info.PathWidthKM))
		}
	}
	if info.HasSaros {
		if options.Language == "en" {
			label := fmt.Sprintf("Saros series %d, member %d/%d", info.Saros.Series, info.Saros.Member, info.Saros.Count)
			if !info.Saros.Verified {
				label += " (provisional)"
			}
			details = append(details, label)
		} else {
			label := fmt.Sprintf("沙罗序列 %d，第 %d/%d 个成员", info.Saros.Series, info.Saros.Member, info.Saros.Count)
			if !info.Saros.Verified {
				label += "（推算）"
			}
			details = append(details, label)
		}
	}
	if hasLocal {
		if options.Language == "en" {
			details = append(details, fmt.Sprintf("Sun alt %.1f° az %.1f°", local.SunAltitude, local.SunAzimuth))
		} else {
			details = append(details, fmt.Sprintf("食甚点太阳高度 %.1f° 方位 %.1f°", local.SunAltitude, local.SunAzimuth))
		}
		if local.HasCentral && !local.CentralStart.IsZero() && !local.CentralEnd.IsZero() {
			duration := local.CentralEnd.Sub(local.CentralStart)
			if duration > 0 {
				if options.Language == "en" {
					details = append(details, "central duration "+formatSolarEclipseMapDuration(duration))
				} else {
					details = append(details, "中心食持续 "+formatSolarEclipseMapDuration(duration))
				}
			}
		}
	}
	if len(details) > 0 {
		detailText := strings.Join(details, " | ")
		labels.ReserveText(summaryX, 106, 11, detailText, "middle")
		fmt.Fprintf(builder, `<text x="%.3f" y="106" fill="#596164" font-family="Arial, sans-serif" font-size="11" text-anchor="middle">%s</text>`,
			summaryX, html.EscapeString(detailText))
	}
	conjunctionY := 0.0
	if hasGeocentric {
		conjunctionY = 128
		conjunction := geocentric.RightAscensionConjunction
		label := "地心合（视赤经相等）"
		if options.Language == "en" {
			label = "Geocentric conjunction (equal apparent right ascension)"
		}
		conjunctionText := fmt.Sprintf("%s = %s UT | J.D. = %.6f",
			label, conjunction.UTC().Format("15:04:05.0"), geocentric.RightAscensionConjunctionJD)
		labels.ReserveText(summaryX, conjunctionY, 11, conjunctionText, "middle")
		fmt.Fprintf(builder, `<text x="%.3f" y="%.3f" fill="#596164" font-family="Arial, sans-serif" font-size="11" text-anchor="middle">%s</text>`,
			summaryX, conjunctionY, html.EscapeString(conjunctionText))
	}
	// 图上所有时刻都按展示时区，这里明确写出它与 UT 的偏差，避免被当成 UT 读。
	zoneNote := timenote.Scale(options.TimeScale, maximum, options.Location, options.Language)
	labels.ReserveText(summaryX, conjunctionY+20, 11, zoneNote, "middle")
	fmt.Fprintf(builder, `<text x="%.3f" y="%.3f" fill="#7b8794" font-family="Arial, sans-serif" font-size="11" text-anchor="middle">%s</text>`,
		summaryX, conjunctionY+20, html.EscapeString(zoneNote))
}

func formatSolarEclipseMapDuration(value time.Duration) string {
	seconds := int(math.Round(value.Seconds()))
	if seconds < 0 {
		seconds = -seconds
	}
	return fmt.Sprintf("%02d:%02d", seconds/60, seconds%60)
}

func writeSolarEclipseMapLegend(builder *strings.Builder, layout solarEclipseMapLayout, rows [][]solarEclipseMapLegendItem) {
	builder.WriteString(`<g class="solar-map-legend">`)
	for row, items := range rows {
		y := layout.legendY + float64(row)*solarEclipseLegendLineStep
		x := layout.frame.X
		if layout.nasa {
			x = layout.panelX
		}
		for _, item := range items {
			switch item.kind {
			case "line":
				fmt.Fprintf(builder, `<line x1="%.3f" y1="%.3f" x2="%.3f" y2="%.3f" stroke="%s" stroke-width="1.6" stroke-dasharray="%s"/>`,
					x, y-4, x+20, y-4, item.color, item.dash)
			case "contact":
				fmt.Fprintf(builder, `<circle cx="%.3f" cy="%.3f" r="3" fill="%s" stroke="#ffffff" stroke-width="0.8"/>`,
					x+8, y-4, item.color)
			default:
				fmt.Fprintf(builder, `<rect x="%.3f" y="%.3f" width="18" height="9" fill="%s" fill-opacity="0.62"/>`,
					x, y-8, item.color)
			}
			fmt.Fprintf(builder, `<text x="%.3f" y="%.3f" fill="#465053" font-family="Arial, sans-serif" font-size="%.0f">%s</text>`,
				x+solarEclipseLegendIconWidth, y, solarEclipseLegendFontSize, html.EscapeString(item.label))
			x += solarEclipseLegendIconWidth + svgchart.EstimatedTextWidth(item.label, solarEclipseLegendFontSize) +
				solarEclipseLegendColumnGap
		}
	}
	builder.WriteString(`</g>`)
}

func solarEclipseOutlineLegendLabel(kind string, eclipseType eclipsecore.SolarEclipseType, step time.Duration, language string) string {
	interval := solarEclipseMapStepLabel(step, language)
	if kind == "penumbra" {
		if language == "en" {
			return "Penumbral outlines (" + interval + ")"
		}
		return "半影时刻线（" + interval + "）"
	}
	name := "本影轮廓"
	if eclipseType == eclipsecore.SolarEclipseAnnular {
		name = "反本影轮廓"
	} else if eclipseType == eclipsecore.SolarEclipseHybrid {
		name = "本影/反本影轮廓"
	}
	if language == "en" {
		name = "Umbral outlines"
		if eclipseType == eclipsecore.SolarEclipseAnnular {
			name = "Antumbral outlines"
		} else if eclipseType == eclipsecore.SolarEclipseHybrid {
			name = "Umbral/antumbral outlines"
		}
		return name + " (" + interval + ")"
	}
	return name + "（" + interval + "）"
}

func solarEclipseMapStepLabel(step time.Duration, language string) string {
	if step%time.Minute != 0 {
		return step.String()
	}
	minutes := int(step / time.Minute)
	if language == "en" {
		return fmt.Sprintf("%d min", minutes)
	}
	return fmt.Sprintf("%d 分钟", minutes)
}

const (
	solarEclipseMapFooterFontSize      = 11.0
	solarEclipseMapFooterBaselineInset = 38.0
)

func solarEclipseMapFooterText(options SolarEclipseMapSVGOptions, projection svgmap.Projection) string {
	return solarEclipseMapFooterTextBase(options, projection)
}

func solarEclipseMapFooterTextBase(options SolarEclipseMapSVGOptions, projection svgmap.Projection) string {
	if options.FooterNote != "" {
		return options.FooterNote
	}
	if options.Language == "en" {
		return eclipseMapProjectionLabel(projection, "en") + "; sampled penumbral sweep and central path; Natural Earth 1:50m physical land, no administrative boundaries."
	}
	return eclipseMapProjectionLabel(projection, "zh") + "；偏食区为半影足迹时间扫掠，叠加中心食带；Natural Earth 1:50m 物理陆地底图，不含行政边界。"
}

func writeSolarEclipseMapFooter(
	builder *strings.Builder,
	layout solarEclipseMapLayout,
	options SolarEclipseMapSVGOptions,
	projection svgmap.Projection,
) {
	// 详细版式的球面居中，页脚要从左边距起排，否则长句子会被图框裁掉。
	footerX := layout.frame.X
	if layout.nasa {
		footerX = layout.margin
	}
	baseline := float64(options.Height) - solarEclipseMapFooterBaselineInset
	// 默认说明是单行；调用方文本按图框宽度折行，行数按画布底边截断，首行位置不变。
	lines := []string{solarEclipseMapFooterText(options, projection)}
	if options.FooterNote != "" {
		maxWidth := float64(options.Width) - footerX - layout.margin
		lines = svgchart.TruncateTextLines(svgchart.WrapText(options.FooterNote, maxWidth, solarEclipseMapFooterFontSize),
			maxWidth, solarEclipseMapFooterFontSize,
			svgchart.BaselineLineLimit(solarEclipseMapFooterFontSize, svgchart.FooterLineHeight(solarEclipseMapFooterFontSize),
				baseline, float64(options.Height)-svgchart.FooterBottomPadding(solarEclipseMapFooterFontSize)))
	}
	for index, line := range lines {
		fmt.Fprintf(builder, `<text x="%.3f" y="%.3f" fill="#596164" font-family="Georgia, 'Times New Roman', serif" font-size="%.0f">%s</text>`,
			footerX, baseline+float64(index)*svgchart.FooterLineHeight(solarEclipseMapFooterFontSize),
			solarEclipseMapFooterFontSize, html.EscapeString(line))
	}
}

func solarEclipseMapTypeName(value eclipsecore.SolarEclipseType, language string) string {
	if language == "en" {
		switch value {
		case eclipsecore.SolarEclipseTotal:
			return "Total Solar Eclipse"
		case eclipsecore.SolarEclipseAnnular:
			return "Annular Solar Eclipse"
		case eclipsecore.SolarEclipseHybrid:
			return "Hybrid Solar Eclipse"
		default:
			return "Partial Solar Eclipse"
		}
	}
	switch value {
	case eclipsecore.SolarEclipseTotal:
		return "日全食"
	case eclipsecore.SolarEclipseAnnular:
		return "日环食"
	case eclipsecore.SolarEclipseHybrid:
		return "全环食"
	default:
		return "日偏食"
	}
}

func solarEclipseCentralPathLabel(value eclipsecore.SolarEclipseType, language string) string {
	if language == "en" {
		switch value {
		case eclipsecore.SolarEclipseAnnular:
			return "Annular path"
		case eclipsecore.SolarEclipseHybrid:
			return "Hybrid central path"
		default:
			return "Path of totality"
		}
	}
	switch value {
	case eclipsecore.SolarEclipseAnnular:
		return "环食带"
	case eclipsecore.SolarEclipseHybrid:
		return "全环食中心带"
	default:
		return "全食带"
	}
}

func solarEclipseCentralPathColor(value eclipsecore.SolarEclipseType) string {
	if value == eclipsecore.SolarEclipseAnnular {
		return "#a94f3f"
	}
	if value == eclipsecore.SolarEclipseHybrid {
		return "#76506f"
	}
	return "#38516d"
}

func reverseGeoPoints(points []svgmap.GeoPoint) {
	for left, right := 0, len(points)-1; left < right; left, right = left+1, right-1 {
		points[left], points[right] = points[right], points[left]
	}
}
