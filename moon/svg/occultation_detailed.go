package svg

import (
	"github.com/starainrt/astro/internal/timenote"
	"errors"
	"fmt"
	"html"
	"math"
	"strings"
	"time"

	"github.com/starainrt/astro"
	"github.com/starainrt/astro/basic"
	"github.com/starainrt/astro/internal/svgchart"
	"github.com/starainrt/astro/internal/svgmap"
	"github.com/starainrt/astro/moon"
)

// ErrInvalidOccultationDetailedSVGOptions 表示详细版式画布装不下地图、数据块与页脚。
// ErrInvalidOccultationDetailedSVGOptions reports a detailed-layout canvas that cannot hold the map, data blocks, and footer.
var ErrInvalidOccultationDetailedSVGOptions = errors.New("invalid occultation detailed SVG options")

const (
	occultationDetailedMinimumWidth  = 480
	occultationDetailedMinimumHeight = 320
	// occultationDetailedMinimumMap 与 occultationDetailedMinimumCell 是球面与数据格仍可辨认的最小尺寸。
	occultationDetailedMinimumMap  = 120
	occultationDetailedMinimumCell = 120
)

// OccultationDetailedSVGOptions 控制详细版式的月掩星组合图。
// OccultationDetailedSVGOptions configures the detailed-layout occultation chart.
type OccultationDetailedSVGOptions struct {
	// Width 与 Height 是画布尺寸；<=0 使用 1000x1414 的整页开本。
	// 版式必须容得下地图与数据块，否则返回 ErrInvalidOccultationDetailedSVGOptions。
	// Width and Height are the canvas size; <=0 uses the 1000x1414 whole-page format.
	// The layout must fit the map and the data block, otherwise the renderer returns ErrInvalidOccultationDetailedSVGOptions.
	Width  int
	Height int
	// Language 为 "en" 时使用英文，其他值使用中文。
	// Language renders English for "en" and Chinese for any other value.
	Language string
	// Location 控制显示时刻的时区；nil 使用 UTC+8。
	// Location sets the zone of displayed instants; nil uses UTC+8.
	Location *time.Location
	// TimeScale 选择图中时刻的时标：零值 UTC；TimeScaleUT1 改用 UT1 时刻，此时 Location 必须是
	// nil 或 UTC（否则渲染器返回错误），并在图注里声明尺度。
	// TimeScale selects the label scale: the zero value is UTC; TimeScaleUT1 uses UT1 labels, requires
	// Location to be nil or UTC (otherwise the renderer returns an error) and declares the scale.
	TimeScale astro.TimeScale
	// TimeLabelStep 控制中心线上的时刻标签；零值用 30 分钟，负值禁用。
	// TimeLabelStep controls the time labels along the center line; zero uses 30 minutes and a negative value disables them.
	TimeLabelStep time.Duration
	// Title 与 FooterNote 为空时自动生成。
	// Title and FooterNote are generated automatically when left empty.
	Title      string
	FooterNote string
}

// occultationDetailedView 汇总详细版式各区块的位置。
type occultationDetailedView struct {
	width, height float64
	margin        float64
	mapX, mapY    float64
	mapSize       float64
	gridX, gridY  float64
	cellWidth     float64
	cellHeight    float64
	gapX, gapY    float64
	footerY       float64
	landscape     bool
}

func normalizeOccultationDetailedSVGOptions(options OccultationDetailedSVGOptions) OccultationDetailedSVGOptions {
	if options.Width <= 0 {
		options.Width = 1000
	}
	if options.Height <= 0 {
		options.Height = 1414
	}
	if options.Location == nil {
		options.Location = time.FixedZone("UTC+8", starOccultationSVGDefaultZone)
	}
	if strings.EqualFold(options.Language, starOccultationSVGLanguageEnglish) {
		options.Language = starOccultationSVGLanguageEnglish
	} else {
		options.Language = starOccultationSVGLanguageChinese
	}
	if options.TimeLabelStep == 0 {
		options.TimeLabelStep = 30 * time.Minute
	}
	return options
}

func validateOccultationDetailedSVGOptions(options OccultationDetailedSVGOptions) error {
	if options.Width > 0 && options.Width < occultationDetailedMinimumWidth {
		return fmt.Errorf("%w: width must be zero or at least %d",
			ErrInvalidOccultationDetailedSVGOptions, occultationDetailedMinimumWidth)
	}
	if options.Height > 0 && options.Height < occultationDetailedMinimumHeight {
		return fmt.Errorf("%w: height must be zero or at least %d",
			ErrInvalidOccultationDetailedSVGOptions, occultationDetailedMinimumHeight)
	}
	return nil
}

func occultationDetailedMargin(width float64) float64 {
	return math.Max(30, math.Min(48, width*0.044))
}

func occultationDetailedViewFor(options OccultationDetailedSVGOptions, summaryLines int) (occultationDetailedView, error) {
	width, height := float64(options.Width), float64(options.Height)
	margin := occultationDetailedMargin(width)
	// 摘要折行数随画布宽度变化，标题区必须跟着长，否则窄画布上摘压到地图或数据块。
	headerBottom := math.Max(150, math.Max(74+float64(summaryLines)*20+16, height*0.115))
	footerHeight := math.Max(56, height*0.05)
	legendHeight := math.Max(42, height*0.04)
	gap := math.Max(14, height*0.014)
	view := occultationDetailedView{
		width: width, height: height, margin: margin,
		gapX: math.Max(14, width*0.014), gapY: math.Max(12, height*0.013),
	}
	legendY := height - footerHeight - legendHeight
	view.footerY = height - 42
	blockHeight := math.Max(74, height*0.055)
	panelHeight := math.Max(96, height*0.10)

	if width >= height {
		// 横版：地图在左，数据块两栏三行在右。
		view.landscape = true
		gridWidth := 2*230 + view.gapX
		view.gridX = width - margin - gridWidth
		view.cellWidth = 230
		bandTop := headerBottom + gap
		bandBottom := legendY - gap
		view.gridY = bandTop
		view.cellHeight = (bandBottom - bandTop - 2*view.gapY) / 3
		view.mapX = margin
		view.mapY = bandTop
		view.mapSize = math.Min(view.gridX-gap-margin, bandBottom-bandTop)
		if view.mapSize < occultationDetailedMinimumMap || view.cellHeight <= 0 || view.cellWidth <= 0 {
			return view, detailedCanvasError(options)
		}
		return view, nil
	}
	// 竖版：日月块在地图上方两侧，地图居中，两行三栏面板在地图下方。
	view.mapY = headerBottom + blockHeight + gap
	// 球面下方先留图例的位置，再排两行面板；地图边长取剩余高度，不足则拒绝而不是按宽度放大。
	const legendSpace = 56.0
	view.mapSize = math.Min(width-2*margin,
		height-view.mapY-legendSpace-2*panelHeight-3*gap-legendHeight-footerHeight)
	view.cellWidth = (width - 2*margin - 2*view.gapX) / 3
	if view.mapSize < occultationDetailedMinimumMap || view.cellWidth < occultationDetailedMinimumCell {
		return view, detailedCanvasError(options)
	}
	view.mapX = (width - view.mapSize) / 2
	panelY := view.mapY + view.mapSize + legendSpace
	view.gridX = margin
	view.gridY = panelY
	view.cellHeight = panelHeight
	return view, nil
}

func detailedCanvasError(options OccultationDetailedSVGOptions) error {
	return fmt.Errorf("%w: canvas %dx%d cannot hold the map, data blocks, and footer",
		ErrInvalidOccultationDetailedSVGOptions, options.Width, options.Height)
}

func (view occultationDetailedView) mapFrame(projection svgmap.Projection, center svgmap.GeoPoint) svgmap.Frame {
	return svgmap.Frame{
		X: view.mapX, Y: view.mapY, Width: view.mapSize, Height: view.mapSize,
		Projection: projection, CenterLongitude: center.Longitude, CenterLatitude: center.Latitude,
	}
}

// validateOccultationDetailedBlocks 逐块检查等分行高下文字是否压叠；太矮的画布必须拒绝而不是画出压叠的面板。
func validateOccultationDetailedBlocks(view occultationDetailedView, blocks []svgchart.PanelBlock) error {
	columns := 3
	if view.landscape {
		columns = 2
	}
	for index, block := range blocks {
		if len(block.Rows) == 0 {
			continue
		}
		column, row := index%columns, index/columns
		x := view.gridX + float64(column)*(view.cellWidth+view.gapX)
		y := view.gridY + float64(row)*(view.cellHeight+view.gapY)
		if svgchart.PanelBoxOverlaps(x, y, view.cellWidth, view.cellHeight, block.Title, block.Rows) {
			return fmt.Errorf("%w: canvas %gx%g cannot keep data block %q rows apart",
				ErrInvalidOccultationDetailedSVGOptions, view.width, view.height, block.Title)
		}
	}
	return nil
}

// writeOccultationDetailedGrid 按两栏三行（横版）或三栏两行（竖版）画数据块。
func writeOccultationDetailedGrid(b *strings.Builder, view occultationDetailedView, blocks []svgchart.PanelBlock) {
	columns := 3
	if view.landscape {
		columns = 2
	}
	for index, block := range blocks {
		if len(block.Rows) == 0 {
			continue
		}
		column := index % columns
		row := index / columns
		x := view.gridX + float64(column)*(view.cellWidth+view.gapX)
		y := view.gridY + float64(row)*(view.cellHeight+view.gapY)
		svgchart.WritePanelBox(b, "occultation-detailed-panel", x, y, view.cellWidth, view.cellHeight, block.Title, block.Rows)
	}
}

// occultationDetailedMoonRows 月亮的地心坐标；S.D./H.P. 由地心月距换算，与日食月食同口径。
func occultationDetailedMoonRows(at time.Time, language string) []svgchart.PanelRow {
	jde := basic.UTC2TT(basic.Date2JD(at.UTC()))
	ra, dec := basic.HMoonTrueRaDec(jde)
	sd := basic.MoonSemidiameter(jde)
	hp := math.Asin(math.Sin(sd/3600*math.Pi/180)*6378.137/1737.4) * 180 / math.Pi * 3600
	rows := []svgchart.PanelRow{
		{Label: "赤经 R.A.", Value: occultationDetailedRA(ra)},
		{Label: "赤纬 Dec.", Value: occultationDetailedDec(dec)},
		{Label: "视半径 S.D.", Value: occultationDetailedArcsec(sd)},
		{Label: "地平视差 H.P.", Value: occultationDetailedArcsec(hp)},
	}
	if language == starOccultationSVGLanguageEnglish {
		rows[0].Label, rows[1].Label = "R.A.", "Dec."
		rows[2].Label, rows[3].Label = "S.D.", "H.P."
	}
	return rows
}

func occultationDetailedRA(degrees float64) string {
	total := math.Mod(degrees, 360) / 15
	hours := math.Floor(total)
	minutes := math.Floor((total - hours) * 60)
	return fmt.Sprintf("%02.0fh%02.0fm%04.1fs", hours, minutes, ((total-hours)*60-minutes)*60)
}

func occultationDetailedDec(degrees float64) string {
	sign := "+"
	if degrees < 0 {
		sign, degrees = "-", -degrees
	}
	whole := math.Floor(degrees)
	minutes := math.Floor((degrees - whole) * 60)
	return fmt.Sprintf("%s%02.0f°%02.0f'%04.1f\"", sign, whole, minutes, ((degrees-whole)*60-minutes)*60)
}

func occultationDetailedArcsec(arcsec float64) string {
	whole := math.Floor(arcsec / 3600)
	minutes := math.Floor((arcsec/3600 - whole) * 60)
	return fmt.Sprintf("%02.0f°%02.0f'%04.1f\"", whole, minutes, ((arcsec/3600-whole)*60-minutes)*60)
}

// StarOccultationDetailedSVG 生成详细版式的恒星月掩星图。
// StarOccultationDetailedSVG renders a detailed stellar-occultation diagram.
func StarOccultationDetailedSVG(
	path moon.StarOccultationPath,
	star moon.StarCoordinate,
	options OccultationDetailedSVGOptions,
) (string, error) {
	if options.TimeScale == astro.TimeScaleUT1 && options.Location != nil && options.Location != time.UTC {
		return "", fmt.Errorf("occultation SVG: UT1 labels do not take a non-UTC location")
	}
	if err := validateStarOccultationPath(path); err != nil {
		return "", err
	}
	civilGreatest := path.Greatest.Time
	if options.TimeScale == astro.TimeScaleUT1 {
		options.Location = time.UTC
		path = moon.StarOccultationPathInUT1(path)
	}
	return renderOccultationDetailedSVG(path, star, nil, civilGreatest, options)
}

// PlanetOccultationDetailedSVG 生成详细版式的行星月掩星图。
// PlanetOccultationDetailedSVG renders a detailed planetary-occultation diagram.
func PlanetOccultationDetailedSVG(
	path moon.PlanetOccultationPath,
	options OccultationDetailedSVGOptions,
) (string, error) {
	if options.TimeScale == astro.TimeScaleUT1 && options.Location != nil && options.Location != time.UTC {
		return "", fmt.Errorf("occultation SVG: UT1 labels do not take a non-UTC location")
	}
	if err := validatePlanetOccultationPath(path); err != nil {
		return "", err
	}
	civilGreatest := path.Greatest.Time
	if options.TimeScale == astro.TimeScaleUT1 {
		options.Location = time.UTC
		path = moon.PlanetOccultationPathInUT1(path)
	}

	// 中文标题与目标块要写"木星"这类中文行星名，不能直接用 TargetID 里的英文。
	shape := planetOccultationStarShape(path, path.TargetID)
	if options.Language == starOccultationSVGLanguageChinese {
		shape.TargetID = planetOccultationChineseName(path.Planet)
	}
	return renderOccultationDetailedSVG(shape, moon.StarCoordinate{}, &path, civilGreatest, options)
}

func renderOccultationDetailedSVG(
	path moon.StarOccultationPath,
	star moon.StarCoordinate,
	planetPath *moon.PlanetOccultationPath,
	civilGreatest time.Time,
	options OccultationDetailedSVGOptions,
) (string, error) {
	if err := validateOccultationDetailedSVGOptions(options); err != nil {
		return "", err
	}
	options = normalizeOccultationDetailedSVGOptions(options)
	summaryLines := occultationDetailedSummaryLines(path, planetPath, options)
	view, err := occultationDetailedViewFor(options, len(summaryLines))
	if err != nil {
		return "", err
	}
	center := svgmap.GeoPoint{Longitude: path.Greatest.Longitude, Latitude: path.Greatest.Latitude}
	layout := starOccultationSVGLayout{
		width: view.width, height: view.height, margin: view.margin,
		mapX: view.mapX, mapY: view.mapY, mapWidth: view.mapSize, mapHeight: view.mapSize,
		panelX: view.gridX, panelY: view.gridY, panelWidth: view.cellWidth,
		footerY: view.footerY, projection: svgmap.ProjectionOrthographic, center: center,
	}
	svgOptions := StarOccultationSVGOptions{
		Width: options.Width, Height: options.Height, Language: options.Language,
		Location: options.Location, TimeLabelStep: options.TimeLabelStep,
		Projection: MapProjectionOrthographic, Title: options.Title,
	}
	title := starOccultationSVGTitle(path, svgOptions)

	var b strings.Builder
	fmt.Fprintf(&b, `<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" viewBox="0 0 %d %d" role="img" aria-label="%s">`,
		options.Width, options.Height, options.Width, options.Height, html.EscapeString(title))
	b.WriteString(`<defs>`)
	b.WriteString(layout.mapFrame().ClipDefinition("occultation-map-clip"))
	b.WriteString(`</defs>`)
	b.WriteString(`<rect width="100%" height="100%" fill="#efefed"/>`)
	fmt.Fprintf(&b, `<rect x="22" y="18" width="%.3f" height="%.3f" fill="#ffffff" stroke="#c9c9c6" stroke-width="1.2"/>`,
		view.width-44, view.height-36)
	titleText := title
	if options.Title != "" {
		titleText = svgchart.EllipsizeText(title, view.width-80, 25)
	}
	fmt.Fprintf(&b, `<text x="%.3f" y="46" fill="#111111" font-family="Georgia, 'Times New Roman', serif" font-size="25" font-weight="700" text-anchor="middle">%s</text>`,
		view.width/2, html.EscapeString(titleText))
	writeOccultationDetailedSummary(&b, view, summaryLines)

	// 几何引擎失败不是数据格式错误，不能裹成 ErrInvalid…Path 哨兵。
	if planetPath == nil {
		if err := writeStarOccultationMap(&b, path, layout, svgOptions); err != nil {
			return "", fmt.Errorf("stellar occultation detailed band geometry: %w", err)
		}
	} else {
		if err := writePlanetOccultationMap(&b, *planetPath, path, layout, svgOptions); err != nil {
			return "", fmt.Errorf("planetary occultation detailed band geometry: %w", err)
		}
	}
	blocks := occultationDetailedBlocks(path, star, planetPath, civilGreatest, options)
	if err := validateOccultationDetailedBlocks(view, blocks); err != nil {
		return "", err
	}
	writeOccultationDetailedGrid(&b, view, blocks)
	writeOccultationDetailedFooter(&b, view, options, occultationScaleInstant(path, planetPath))
	b.WriteString(`</svg>`)
	return b.String(), nil
}

// occultationDetailedSummaryLines 汇总标题下的摘要行，并按画布可用宽度折行。
func occultationDetailedSummaryLines(
	path moon.StarOccultationPath,
	planetPath *moon.PlanetOccultationPath,
	options OccultationDetailedSVGOptions,
) []string {
	available := float64(options.Width) - 2*occultationDetailedMargin(float64(options.Width))
	lines := []string{}
	for _, line := range occultationDetailedSummaryText(path, planetPath, options) {
		lines = append(lines, starOccultationWrapText(line, available, 12.5)...)
	}
	return lines
}

func occultationDetailedSummaryText(
	path moon.StarOccultationPath,
	planetPath *moon.PlanetOccultationPath,
	options OccultationDetailedSVGOptions,
) []string {
	zone := starOccultationLocationLabel(path.Greatest.Time.In(options.Location), options.Location)
	english := options.Language == starOccultationSVGLanguageEnglish
	lines := []string{}
	format := func(value time.Time) string {
		return value.In(options.Location).Format("15:04:05")
	}
	name := func(zh, en string) string {
		if english {
			return en
		}
		return zh
	}
	if planetPath != nil && planetPath.HasTotalBand {
		if english {
			lines = append(lines, fmt.Sprintf("Partial begins %s | Total begins %s | Greatest %s | Total ends %s | Partial ends %s (%s)",
				format(path.Start.Time), format(planetPath.TotalStart.Time), format(path.Greatest.Time),
				format(planetPath.TotalEnd.Time), format(path.End.Time), zone))
		} else {
			lines = append(lines, fmt.Sprintf("外掩始 %s | 全掩始 %s | 掩甚 %s | 全掩终 %s | 外掩终 %s（%s）",
				format(path.Start.Time), format(planetPath.TotalStart.Time), format(path.Greatest.Time),
				format(planetPath.TotalEnd.Time), format(path.End.Time), zone))
		}
	} else if english {
		lines = append(lines, fmt.Sprintf("Occultation begins %s | Greatest %s | Occultation ends %s (%s)",
			format(path.Start.Time), format(path.Greatest.Time), format(path.End.Time), zone))
	} else {
		lines = append(lines, fmt.Sprintf("掩始 %s | 掩甚 %s | 掩终 %s（%s）",
			format(path.Start.Time), format(path.Greatest.Time), format(path.End.Time), zone))
	}
	lines = append(lines, fmt.Sprintf("%s %s | %s", name("掩甚点", "Greatest point"),
		starOccultationFormatCoordinates(path.Greatest.Longitude, path.Greatest.Latitude),
		name("月球高度", "Moon altitude")+" "+starOccultationFormatSignedDegree(path.Greatest.MoonAltitude)))
	return lines
}

func writeOccultationDetailedSummary(
	b *strings.Builder,
	view occultationDetailedView,
	lines []string,
) {
	for index, line := range lines {
		fmt.Fprintf(b, `<text x="%.3f" y="%.3f" fill="#293235" font-family="Arial, sans-serif" font-size="12.5" text-anchor="middle">%s</text>`,
			view.width/2, 74+float64(index)*20, html.EscapeString(line))
	}
}

func writeOccultationDetailedFooter(b *strings.Builder, view occultationDetailedView, options OccultationDetailedSVGOptions, instant time.Time) {
	text := options.FooterNote
	if text == "" {
		if options.Language == starOccultationSVGLanguageEnglish {
			text = "Orthographic globe centred on the greatest occultation; Natural Earth 1:50m physical land, no administrative boundaries."
		} else {
			text = "正射球面投影，视点取掩甚点；Natural Earth 1:50m 物理陆地底图，不含行政边界。"
		}
	}
	maxWidth := view.width - 2*view.margin
	// 默认说明本来就是单行，只有调用方文本才折行截断；时标声明单独占一行。
	lines := []string{text}
	if options.FooterNote != "" {
		lines = svgchart.WrapText(text, maxWidth, 11)
	}
	lines = svgchart.FooterLinesWithScale(lines, timenote.Scale(options.TimeScale, instant, options.Location, options.Language),
		maxWidth, 11,
		svgchart.BaselineLineLimit(11, svgchart.FooterLineHeight(11), view.footerY,
			view.height-svgchart.FooterBottomPadding(11)))
	for index, line := range lines {
		fmt.Fprintf(b, `<text x="%.3f" y="%.3f" fill="#596164" font-family="Georgia, 'Times New Roman', serif" font-size="11">%s</text>`,
			view.margin, view.footerY+float64(index)*svgchart.FooterLineHeight(11), html.EscapeString(line))
	}
}

// occultationDetailedTargetRows 目标天体的坐标与视半径；恒星没有视圆面，相关行留空。
func occultationDetailedTargetRows(
	star moon.StarCoordinate,
	planetPath *moon.PlanetOccultationPath,
	at time.Time,
	language string,
) []svgchart.PanelRow {
	english := language == starOccultationSVGLanguageEnglish
	jde := basic.UTC2TT(basic.Date2JD(at.UTC()))
	rows := []svgchart.PanelRow{{Label: "目标", Value: star.ID}}
	if english {
		rows[0].Label = "Target"
	}
	if planetPath != nil {
		rows[0].Value = planetPath.TargetID
		if language == starOccultationSVGLanguageChinese {
			rows[0].Value = planetOccultationChineseName(planetPath.Planet)
		}
		ra, dec := occultationDetailedPlanetRaDec(planetPath.Planet, jde)
		if math.IsNaN(ra) {
			rows = append(rows, svgchart.PanelRow{Label: "赤经 R.A.", Value: "—"},
				svgchart.PanelRow{Label: "赤纬 Dec.", Value: "—"})
		} else {
			rows = append(rows, svgchart.PanelRow{Label: "赤经 R.A.", Value: occultationDetailedRA(ra)},
				svgchart.PanelRow{Label: "赤纬 Dec.", Value: occultationDetailedDec(dec)})
		}
		rows = append(rows, svgchart.PanelRow{Label: "视半径 S.D.",
			Value: occultationDetailedArcsec(occultationDetailedPlanetSemidiameter(planetPath.Planet, jde))})
		if english {
			rows[1].Label, rows[2].Label, rows[3].Label = "R.A.", "Dec.", "S.D."
		}
		return rows
	}
	rows = append(rows,
		svgchart.PanelRow{Label: "赤经 R.A.", Value: occultationDetailedRA(star.RA)},
		svgchart.PanelRow{Label: "赤纬 Dec.", Value: occultationDetailedDec(star.Dec)},
		svgchart.PanelRow{Label: "视半径 S.D.", Value: "—"})
	if english {
		rows[1].Label, rows[2].Label, rows[3].Label = "R.A.", "Dec.", "S.D."
	}
	return rows
}

// occultationDetailedPlanetRaDec 空值表示该行星没有可用的赤经入口。
func occultationDetailedPlanetRaDec(planet moon.OccultationPlanet, jde float64) (float64, float64) {
	switch planet {
	case moon.OccultationMercury:
		return basic.MercuryApparentRaDec(jde)
	case moon.OccultationVenus:
		return basic.VenusApparentRaDec(jde)
	case moon.OccultationMars:
		return basic.MarsApparentRaDec(jde)
	case moon.OccultationJupiter:
		return basic.JupiterApparentRaDec(jde)
	case moon.OccultationSaturn:
		return basic.SaturnApparentRaDec(jde)
	case moon.OccultationUranus:
		return basic.UranusApparentRaDec(jde)
	case moon.OccultationNeptune:
		return basic.NeptuneApparentRaDec(jde)
	}
	return math.NaN(), math.NaN()
}

func occultationDetailedPlanetSemidiameter(planet moon.OccultationPlanet, jde float64) float64 {
	switch planet {
	case moon.OccultationMercury:
		return basic.MercurySemidiameter(jde)
	case moon.OccultationVenus:
		return basic.VenusSemidiameter(jde)
	case moon.OccultationMars:
		return basic.MarsSemidiameter(jde)
	case moon.OccultationJupiter:
		return basic.JupiterSemidiameter(jde)
	case moon.OccultationSaturn:
		return basic.SaturnSemidiameter(jde)
	case moon.OccultationUranus:
		return basic.UranusSemidiameter(jde)
	case moon.OccultationNeptune:
		return basic.NeptuneSemidiameter(jde)
	}
	return math.NaN()
}

func occultationDetailedBlocks(
	path moon.StarOccultationPath,
	star moon.StarCoordinate,
	planetPath *moon.PlanetOccultationPath,
	civilGreatest time.Time,
	options OccultationDetailedSVGOptions,
) []svgchart.PanelBlock {
	english := options.Language == starOccultationSVGLanguageEnglish
	name := func(zh, en string) string {
		if english {
			return en
		}
		return zh
	}
	format := func(value time.Time) string {
		if value.IsZero() {
			return "—"
		}
		return value.In(options.Location).Format("15:04:05")
	}
	point := func(label string, value moon.OccultationPathPoint) svgchart.PanelRow {
		if value.Time.IsZero() {
			return svgchart.PanelRow{}
		}
		return svgchart.PanelRow{Label: label, Value: format(value.Time),
			Detail: starOccultationFormatCoordinates(value.Longitude, value.Latitude)}
	}
	pathRows := []svgchart.PanelRow{
		point(name("外掩始", "Partial begins"), path.Start),
		point(name("掩甚", "Greatest"), path.Greatest),
		point(name("外掩终", "Partial ends"), path.End),
	}
	// 行星图传入的 path 是由行星路径合成的恒星壳，带宽口径要从原始行星路径取。
	widthSource := path
	if planetPath != nil {
		widthSource = moon.StarOccultationPath{Greatest: planetPath.Greatest,
			GreatestLimitSeparationKM: planetPath.GreatestLimitSeparationKM}
	}
	partialWidth := widthSource.GreatestLimitSeparationKM
	if partialWidth <= 0 {
		partialWidth = widthSource.Greatest.WidthKM
	}
	ephemeris := []svgchart.PanelRow{
		{Label: name("投影", "Projection"), Value: name("正射球面", "Orthographic")},
		{Label: "ΔT", Value: fmt.Sprintf("%.1f s", basic.DeltaT(basic.UTC2TT(basic.Date2JD(civilGreatest.UTC())), true))},
		{Label: name("月距", "Moon distance"), Value: fmt.Sprintf("%.0f km", basic.HMoonAway(basic.UTC2TT(basic.Date2JD(civilGreatest.UTC()))))},
		{Label: name("部分掩带宽", "Partial-band width"), Value: fmt.Sprintf("%.1f km", partialWidth)},
	}
	if planetPath != nil && planetPath.HasTotalBand {
		ephemeris = append(ephemeris, svgchart.PanelRow{Label: name("全掩带宽", "Total-band width"),
			Value: fmt.Sprintf("%.1f km", planetPath.GreatestTotalWidthKM)})
	}
	jde := basic.UTC2TT(basic.Date2JD(civilGreatest.UTC()))
	physical := basic.MoonPhysical(jde)
	libration := []svgchart.PanelRow{
		{Label: name("经天平动 l", "Libration l"), Value: fmt.Sprintf("%+.2f°", physical.LibrationLongitude)},
		{Label: name("纬天平动 b", "Libration b"), Value: fmt.Sprintf("%+.2f°", physical.LibrationLatitude)},
		{Label: name("自转轴位置角 c", "Axis position angle c"), Value: fmt.Sprintf("%.2f°", physical.PositionAngle)},
	}
	contacts := []svgchart.PanelRow{
		{Label: name("外掩始", "Partial begins"), Value: format(path.Start.Time)},
	}
	if planetPath != nil && planetPath.HasTotalBand {
		contacts = append(contacts,
			svgchart.PanelRow{Label: name("全掩始", "Total begins"), Value: format(planetPath.TotalStart.Time)})
	}
	contacts = append(contacts, svgchart.PanelRow{Label: name("掩甚", "Greatest"), Value: format(path.Greatest.Time)})
	if planetPath != nil && planetPath.HasTotalBand {
		contacts = append(contacts,
			svgchart.PanelRow{Label: name("全掩终", "Total ends"), Value: format(planetPath.TotalEnd.Time)})
	}
	contacts = append(contacts, svgchart.PanelRow{Label: name("外掩终", "Partial ends"), Value: format(path.End.Time)})

	return []svgchart.PanelBlock{
		{Title: name("月亮（地心坐标）", "Moon (geocentric)"), Rows: occultationDetailedMoonRows(civilGreatest, options.Language)},
		{Title: name("目标天体", "Target body"), Rows: occultationDetailedTargetRows(star, planetPath, civilGreatest, options.Language)},
		{Title: name("掩带路径点", "Band path points"), Rows: pathRows},
		{Title: name("接触时刻", "Contacts"), Rows: contacts},
		{Title: name("历表与常数", "Ephemeris and constants"), Rows: ephemeris},
		{Title: name("天平动", "Libration"), Rows: libration},
	}
}
