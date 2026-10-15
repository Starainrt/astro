package svg

import (
	"errors"
	"fmt"
	"html"
	"math"
	"strings"
	"time"

	"github.com/starainrt/astro"
	"github.com/starainrt/astro/internal/svgasset"
	"github.com/starainrt/astro/internal/svgchart"
	"github.com/starainrt/astro/internal/timenote"
	"github.com/starainrt/astro/moon"
)

const (
	limbProfileSVGDefaultWidth  = 920
	limbProfileSVGDefaultHeight = 620
	limbProfileSVGMinimumWidth  = 640
	limbProfileSVGMinimumHeight = 480
	limbProfileSVGMaxRows       = 12
	// limbProfilePanelGap 栏间距须容下上栏的时刻、位置角两行刻度与下栏上沿的两行接触编号。
	limbProfilePanelGap = 56.0
	// limbProfileInsetMinHeight 缩略图（含标题）的最小高度，接触表按它限行。
	limbProfileInsetMinHeight = 120.0
)

// ErrInvalidStarOccultationLimbProfileSVGOptions 表示精密月掩剖面图选项无效 / invalid precise occultation profile chart options.
var ErrInvalidStarOccultationLimbProfileSVGOptions = errors.New("invalid stellar occultation limb profile SVG options")

// ErrEmptyStarOccultationLimbProfile 表示剖面没有采样，恒星全程在月缘起伏带外 / the profile has no samples; the star stays outside the relief band.
var ErrEmptyStarOccultationLimbProfile = errors.New("stellar occultation limb profile has no samples")

// ErrInvalidStarOccultationLimbProfile 表示剖面数据不是 StarOccultationLimbProfile 能给出的形态 / malformed limb profile data.
var ErrInvalidStarOccultationLimbProfile = errors.New("invalid stellar occultation limb profile")

// StarOccultationLimbProfileSVGOptions 控制精密恒星月掩的月缘剖面图 / options for the precise stellar occultation profile chart.
type StarOccultationLimbProfileSVGOptions struct {
	// Width、Height 为画布尺寸，非正值用默认 / canvas size; non-positive uses defaults.
	Width  int
	Height int
	// Title 为空时自动生成 / empty uses an automatic title.
	Title string
	// Language 为 "en" 时英文，否则中文 / "en" for English, otherwise Chinese.
	Language string
	// Location 控制显示时刻，nil 为 UTC+8 / display zone; nil uses UTC+8.
	Location *time.Location
	// TimeScale 为 TimeScaleUT1 时改用 UT1 时刻，此时 Location 须为 nil 或 UTC / label time scale.
	TimeScale astro.TimeScale
}

// StarOccultationLimbProfileSVG 渲染精密恒星月掩剖面：横轴时刻、纵轴相对参考圆月面高度（km），给出遮挡区间与全部掩入复现 / precise stellar occultation limb profile chart.
func StarOccultationLimbProfileSVG(profile moon.StarOccultationLimbProfileResult, options StarOccultationLimbProfileSVGOptions) (string, error) {
	if options.Width > 0 && options.Width < limbProfileSVGMinimumWidth {
		return "", fmt.Errorf("%w: width must be zero or at least %d", ErrInvalidStarOccultationLimbProfileSVGOptions, limbProfileSVGMinimumWidth)
	}
	if options.Height > 0 && options.Height < limbProfileSVGMinimumHeight {
		return "", fmt.Errorf("%w: height must be zero or at least %d", ErrInvalidStarOccultationLimbProfileSVGOptions, limbProfileSVGMinimumHeight)
	}
	if options.TimeScale == astro.TimeScaleUT1 && options.Location != nil && options.Location != time.UTC {
		return "", fmt.Errorf("occultation SVG: UT1 labels do not take a non-UTC location")
	}
	samples := 0
	for _, segment := range profile.Segments {
		samples += len(segment)
	}
	if samples < 2 || profile.Occultation.Greatest.IsZero() {
		return "", ErrEmptyStarOccultationLimbProfile
	}
	if err := validateStarOccultationLimbProfile(profile); err != nil {
		return "", err
	}
	if options.Width <= 0 {
		options.Width = limbProfileSVGDefaultWidth
	}
	if options.Height <= 0 {
		options.Height = limbProfileSVGDefaultHeight
	}
	if options.TimeScale == astro.TimeScaleUT1 {
		options.Location = time.UTC
		profile = moon.StarOccultationLimbProfileInUT1(profile)
	}
	if options.Location == nil {
		options.Location = time.FixedZone("UTC+8", starOccultationSVGDefaultZone)
	}
	if strings.EqualFold(options.Language, starOccultationSVGLanguageEnglish) {
		options.Language = starOccultationSVGLanguageEnglish
	} else {
		options.Language = starOccultationSVGLanguageChinese
	}
	return renderStarOccultationLimbProfileSVG(profile, options), nil
}

type limbProfilePoint struct {
	t              time.Time
	limbKM, starKM float64
	positionAngle  float64
	// skyX、skyY 为恒星相对月心的天空平面坐标（km，向东、向北为正）。
	skyX, skyY float64
}

// limbProfileHeights 把角量换算成月面上相对参考球的高度（km），恒星与月缘用同一比例。
func limbProfileHeights(segment []moon.StarOccultationLimbSample, referenceKM float64) []limbProfilePoint {
	points := make([]limbProfilePoint, 0, len(segment))
	for _, sample := range segment {
		scale := referenceKM / sample.MeanRadiusArcsec
		points = append(points, limbProfilePoint{
			t:             sample.Time,
			limbKM:        (sample.LimbRadiusArcsec - sample.MeanRadiusArcsec) * scale,
			starKM:        (sample.SeparationArcsec - sample.MeanRadiusArcsec) * scale,
			positionAngle: sample.PositionAngleDeg,
			skyX:          sample.SeparationArcsec * scale * math.Sin(sample.PositionAngleDeg*math.Pi/180),
			skyY:          sample.SeparationArcsec * scale * math.Cos(sample.PositionAngleDeg*math.Pi/180),
		})
	}
	return points
}

func limbProfileText(language, zh, en string) string {
	if language == starOccultationSVGLanguageEnglish {
		return en
	}
	return zh
}

func renderStarOccultationLimbProfileSVG(profile moon.StarOccultationLimbProfileResult, options StarOccultationLimbProfileSVGOptions) string {
	info := profile.Occultation
	width, height := float64(options.Width), float64(options.Height)
	language := options.Language
	segments := make([][]limbProfilePoint, 0, len(profile.Segments))
	minKM, maxKM := math.Inf(1), math.Inf(-1)
	for _, segment := range profile.Segments {
		points := limbProfileHeights(segment, limbProfileReferenceKM(profile))
		if len(points) < 2 {
			continue
		}
		for _, point := range points {
			minKM = math.Min(minKM, math.Min(point.limbKM, point.starKM))
			maxKM = math.Max(maxKM, math.Max(point.limbKM, point.starKM))
		}
		segments = append(segments, points)
	}
	if !(maxKM > minKM) {
		minKM, maxKM = -1, 1
	}
	padKM := (maxKM - minKM) * 0.08
	minKM, maxKM = minKM-padKM, maxKM+padKM

	title := options.Title
	if title == "" {
		target := info.TargetID
		if target == "" {
			target = limbProfileText(language, "恒星", "star")
		}
		date := info.Greatest.In(options.Location).Format("2006-01-02")
		title = date + limbProfileText(language, " 精密月掩月缘剖面 · "+target, " Limb Profile · Occultation of "+target)
	}
	site := starOccultationFormatCoordinates(info.Observer.Longitude, info.Observer.Latitude)
	summary := fmt.Sprintf(limbProfileText(language, "观测点 %s | 椭球高 %.0f 米 | 遮挡 %d 段 | 采样 %s", "Site %s | elevation %.0f m | %d hidden interval(s) | step %s"),
		site, info.Observer.Height, len(profile.Hidden), profile.Step)
	direction := limbProfileDirectionText(segments, profile.Hidden, language)

	const top = 136.0
	tableWidth := 260.0
	left, right := 70.0, width-tableWidth-48
	chartTop, chartBottom := top, height-122
	panels := len(segments)
	if panels < 1 {
		panels = 1
	}
	panelHeight := (chartBottom - chartTop - limbProfilePanelGap*float64(panels-1)) / float64(panels)

	var b strings.Builder
	fmt.Fprintf(&b, `<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" viewBox="0 0 %d %d" role="img" aria-label="%s">`,
		options.Width, options.Height, options.Width, options.Height, html.EscapeString(title))
	b.WriteString(`<defs>`)
	b.WriteString(svgasset.MoonFaceSymbol())
	b.WriteString(`<marker id="limb-profile-arrow" viewBox="0 0 10 10" refX="8" refY="5" markerWidth="6" markerHeight="6" orient="auto-start-reverse"><path d="M 0 0 L 10 5 L 0 10 z" fill="#d0473b"/></marker>`)
	b.WriteString(`</defs>`)
	b.WriteString(`<rect width="100%" height="100%" fill="#efefed"/>`)
	fmt.Fprintf(&b, `<rect x="22" y="18" width="%.3f" height="%.3f" fill="#ffffff" stroke="#c9c9c6" stroke-width="1.2"/>`, width-44, height-36)
	fmt.Fprintf(&b, `<text x="%.3f" y="46" fill="#111111" font-family="Georgia, 'Times New Roman', serif" font-size="%d" font-weight="700" text-anchor="middle">%s</text>`,
		width/2, starOccultationTitleFontSize(title, width), html.EscapeString(svgchart.EllipsizeText(title, width-80, float64(starOccultationTitleFontSize(title, width)))))
	fmt.Fprintf(&b, `<text x="%.3f" y="76" fill="#222222" font-family="Georgia, 'Times New Roman', serif" font-size="13" text-anchor="middle">%s</text>`,
		width/2, html.EscapeString(svgchart.EllipsizeText(summary, width-80, 13)))
	fmt.Fprintf(&b, `<text x="%.3f" y="98" fill="#3b4143" font-family="Georgia, 'Times New Roman', serif" font-size="12" text-anchor="middle">%s</text>`,
		width/2, html.EscapeString(svgchart.EllipsizeText(direction, width-80, 12)))

	for index, points := range segments {
		y0 := chartTop + float64(index)*(panelHeight+limbProfilePanelGap)
		writeLimbProfilePanel(&b, points, profile.Hidden, left, right, y0, y0+panelHeight, minKM, maxKM, options)
	}
	// 接触表与图例让出缩略图的最小高度：图例占 84 px，表头 28 px，每行 19 px。
	maxRows := int((chartBottom + 30 - limbProfileInsetMinHeight - 18 - top - 28 - 84) / 19)
	if maxRows < 2 {
		maxRows = 2
	}
	if maxRows > limbProfileSVGMaxRows {
		maxRows = limbProfileSVGMaxRows
	}
	legendBottom := writeLimbProfileContacts(&b, profile, right+28, top, width-48, maxRows, options)
	writeLimbProfileInset(&b, segments, info, limbProfileReferenceKM(profile), right+28, legendBottom+18, width-48, chartBottom+30, language)
	writeLimbProfileFooter(&b, info, limbProfileReferenceKM(profile), left, height-66, width-left-40, options)
	b.WriteString(`</svg>`)
	return b.String()
}

func writeLimbProfilePanel(b *strings.Builder, points []limbProfilePoint, hidden []moon.OccultationInterval,
	left, right, top, bottom, minKM, maxKM float64, options StarOccultationLimbProfileSVGOptions) {
	start, end := points[0].t, points[len(points)-1].t
	span := end.Sub(start).Seconds()
	if span <= 0 {
		span = 1
	}
	x := func(t time.Time) float64 { return left + (right-left)*t.Sub(start).Seconds()/span }
	y := func(km float64) float64 { return bottom - (bottom-top)*(km-minKM)/(maxKM-minKM) }
	fmt.Fprintf(b, `<rect x="%.3f" y="%.3f" width="%.3f" height="%.3f" fill="#fafaf8" stroke="#9a9a96" stroke-width="0.8"/>`, left, top, right-left, bottom-top)
	for _, interval := range hidden {
		if interval.End.Before(start) || interval.Start.After(end) {
			continue
		}
		x0, x1 := x(laterTime(interval.Start, start)), x(earlierTime(interval.End, end))
		fmt.Fprintf(b, `<rect x="%.3f" y="%.3f" width="%.3f" height="%.3f" fill="#3b4a5a" fill-opacity="0.16"/>`, x0, top, math.Max(x1-x0, 0.6), bottom-top)
	}
	if minKM < 0 && maxKM > 0 {
		fmt.Fprintf(b, `<line x1="%.3f" y1="%.3f" x2="%.3f" y2="%.3f" stroke="#b8b8b4" stroke-dasharray="4 3" stroke-width="0.8"/>`, left, y(0), right, y(0))
	}
	var limb, star strings.Builder
	for index, point := range points {
		command := "L"
		if index == 0 {
			command = "M"
		}
		fmt.Fprintf(&limb, "%s%.2f %.2f ", command, x(point.t), y(point.limbKM))
		fmt.Fprintf(&star, "%s%.2f %.2f ", command, x(point.t), y(point.starKM))
	}
	fmt.Fprintf(b, `<path d="%s" fill="none" stroke="#6f5a3a" stroke-width="1.3"/>`, strings.TrimSpace(limb.String()))
	fmt.Fprintf(b, `<path d="%s" fill="none" stroke="#9b3f36" stroke-width="1.1" stroke-dasharray="5 2"/>`, strings.TrimSpace(star.String()))
	font := `font-family="Georgia, 'Times New Roman', serif"`
	// 横轴按整刻度标时刻，下方配该时刻的位置角；刻度线贯穿剖面框作网格。
	step := limbProfileTickStep(span, right-left)
	majors := limbProfileTicks(start, end, step, options.Location)
	for _, tick := range majors {
		tx := x(tick)
		fmt.Fprintf(b, `<line x1="%.3f" y1="%.3f" x2="%.3f" y2="%.3f" stroke="#e2e2de" stroke-width="0.6"/>`, tx, top, tx, bottom)
		fmt.Fprintf(b, `<line x1="%.3f" y1="%.3f" x2="%.3f" y2="%.3f" stroke="#9a9a96" stroke-width="0.8"/>`, tx, bottom, tx, bottom+3)
		// 标签夹进剖面框，靠近两端时不越出框线，也不压住框外的 PA 轴标题。
		label := tick.In(options.Location).Format(limbProfileTickLayout(step))
		labelX := limbProfileClampLabel(tx, svgchart.EstimatedTextWidth(label, 10), left, right)
		fmt.Fprintf(b, `<text x="%.3f" y="%.3f" fill="#4f595b" %s font-size="10" text-anchor="middle">%s</text>`, labelX, bottom+14, font, label)
		angle := limbProfileFormatAngle(limbProfilePositionAngleAt(points, tick))
		angleX := limbProfileClampLabel(tx, svgchart.EstimatedTextWidth(angle, 9), left, right)
		fmt.Fprintf(b, `<text x="%.3f" y="%.3f" fill="#6f5a3a" %s font-size="9" text-anchor="middle">%s</text>`, angleX, bottom+26, font, angle)
	}
	fmt.Fprintf(b, `<text x="%.3f" y="%.3f" fill="#6f5a3a" %s font-size="9" text-anchor="end">PA</text>`, left-4, bottom+26, font)
	minor := step / 6
	if step%time.Minute != 0 {
		minor = step / 5
	}
	major := make(map[int64]bool, len(majors))
	for _, tick := range majors {
		major[tick.UnixNano()] = true
	}
	for _, tick := range limbProfileTicks(start, end, minor, options.Location) {
		if major[tick.UnixNano()] {
			continue
		}
		fmt.Fprintf(b, `<line x1="%.3f" y1="%.3f" x2="%.3f" y2="%.3f" stroke="#b8b8b4" stroke-width="0.6"/>`, x(tick), bottom, x(tick), bottom+2)
	}
	for _, label := range limbProfileContactLabels(hidden, start, end, x, left, right) {
		labelY := top - 4 - float64(label.row)*11
		fmt.Fprintf(b, `<line x1="%.3f" y1="%.3f" x2="%.3f" y2="%.3f" stroke="#3b4a5a" stroke-width="0.8"/>`, label.x, labelY+2, label.contactX, top+6)
		fmt.Fprintf(b, `<text x="%.3f" y="%.3f" fill="#3b4a5a" %s font-size="10" font-weight="700" text-anchor="middle">%d</text>`, label.x, labelY, font, label.number)
	}
	if ratio := limbProfileExaggeration(points, right-left, bottom-top, maxKM-minKM); ratio > 1.5 {
		fmt.Fprintf(b, `<text x="%.3f" y="%.3f" fill="#747c7d" %s font-size="10" text-anchor="end">%s</text>`, right-6, top+14, font,
			html.EscapeString(fmt.Sprintf(limbProfileText(options.Language, "纵向放大约 %.0f 倍", "vertical ×%.0f"), ratio)))
	} else if ratio > 0 && ratio < 1/1.5 {
		fmt.Fprintf(b, `<text x="%.3f" y="%.3f" fill="#747c7d" %s font-size="10" text-anchor="end">%s</text>`, right-6, top+14, font,
			html.EscapeString(fmt.Sprintf(limbProfileText(options.Language, "横向放大约 %.0f 倍", "horizontal ×%.0f"), 1/ratio)))
	}
	fmt.Fprintf(b, `<text x="%.3f" y="%.3f" fill="#4f595b" %s font-size="10" text-anchor="end">%+.1f</text>`, left-4, top+9, font, maxKM)
	fmt.Fprintf(b, `<text x="%.3f" y="%.3f" fill="#4f595b" %s font-size="10" text-anchor="end">%+.1f</text>`, left-4, bottom, font, minKM)
}

const limbProfileContactLabelGap = 14.0

type limbProfileContactLabel struct {
	number      int
	row         int
	x, contactX float64
}

// limbProfileContactLabels 接触编号两行排位后逐行自右向左回推，使其落在 [left, right] 内且同行间距不小于 14 px。
func limbProfileContactLabels(hidden []moon.OccultationInterval, start, end time.Time, x func(time.Time) float64, left, right float64) []limbProfileContactLabel {
	var labels []limbProfileContactLabel
	rows := [2][]int{}
	rowEnd := [2]float64{math.Inf(-1), math.Inf(-1)}
	number := 0
	for _, interval := range hidden {
		for _, contact := range []time.Time{interval.Start, interval.End} {
			number++
			if contact.Before(start) || contact.After(end) {
				continue
			}
			cx := x(contact)
			row := 0
			if rowEnd[1] < rowEnd[0] {
				row = 1
			}
			if cx-rowEnd[0] >= limbProfileContactLabelGap {
				row = 0
			}
			labelX := math.Max(cx, rowEnd[row]+limbProfileContactLabelGap)
			rowEnd[row] = labelX
			rows[row] = append(rows[row], len(labels))
			labels = append(labels, limbProfileContactLabel{number: number, row: row, x: labelX, contactX: cx})
		}
	}
	halfWidth := func(label limbProfileContactLabel) float64 {
		return svgchart.EstimatedTextWidth(fmt.Sprint(label.number), 10) / 2
	}
	for _, indices := range rows {
		if len(indices) == 0 {
			continue
		}
		limit := math.Inf(1)
		for k := len(indices) - 1; k >= 0; k-- {
			label := &labels[indices[k]]
			label.x = math.Min(label.x, math.Min(limit, right-halfWidth(*label)))
			limit = label.x - limbProfileContactLabelGap
		}
		// 一行塞不下时均分整行，间距小于 14 px 也不越框。
		first, last := labels[indices[0]], labels[indices[len(indices)-1]]
		if low, high := left+halfWidth(first), right-halfWidth(last); first.x < low && len(indices) > 1 {
			for k, index := range indices {
				labels[index].x = low + (high-low)*float64(k)/float64(len(indices)-1)
			}
		} else if first.x < low {
			labels[indices[0]].x = low
		}
	}
	return labels
}

func laterTime(a, b time.Time) time.Time {
	if a.After(b) {
		return a
	}
	return b
}

func earlierTime(a, b time.Time) time.Time {
	if a.Before(b) {
		return a
	}
	return b
}

// writeLimbProfileContacts 返回图例底边，供缩略图接在下方。
func writeLimbProfileContacts(b *strings.Builder, profile moon.StarOccultationLimbProfileResult, left, top, right float64, maxRows int, options StarOccultationLimbProfileSVGOptions) float64 {
	language := options.Language
	font := `font-family="Georgia, 'Times New Roman', serif"`
	fmt.Fprintf(b, `<text x="%.3f" y="%.3f" fill="#111111" %s font-size="13" font-weight="700">%s</text>`, left, top+4, font,
		html.EscapeString(limbProfileText(language, "掩入 / 复现", "Disappearances / reappearances")))
	rows := make([][2]string, 0, 2*len(profile.Hidden))
	for _, interval := range profile.Hidden {
		rows = append(rows,
			[2]string{fmt.Sprintf("%d  %s", len(rows)+1, limbProfileText(language, "掩入", "Disappears")), interval.Start.In(options.Location).Format("15:04:05.00")},
			[2]string{fmt.Sprintf("%d  %s", len(rows)+2, limbProfileText(language, "复现", "Reappears")), interval.End.In(options.Location).Format("15:04:05.00")})
	}
	if len(rows) == 0 {
		rows = append(rows, [2]string{limbProfileText(language, "未发生遮挡", "Not occulted"), ""})
	}
	overflow := 0
	if len(rows) > maxRows {
		overflow = len(rows) - (maxRows - 1)
		rows = rows[:maxRows-1]
	}
	for index, row := range rows {
		baseline := top + 28 + float64(index)*19
		fmt.Fprintf(b, `<text x="%.3f" y="%.3f" fill="#3b4143" %s font-size="12">%s</text>`, left, baseline, font, html.EscapeString(row[0]))
		fmt.Fprintf(b, `<text x="%.3f" y="%.3f" fill="#222222" %s font-size="12" text-anchor="end">%s</text>`, right, baseline, font, row[1])
	}
	if overflow > 0 {
		fmt.Fprintf(b, `<text x="%.3f" y="%.3f" fill="#747c7d" %s font-size="11">%s</text>`, left, top+28+float64(len(rows))*19, font,
			html.EscapeString(fmt.Sprintf(limbProfileText(language, "另有 %d 个接触未列出", "%d more contacts not listed"), overflow)))
	}
	listed := len(rows)
	if overflow > 0 {
		listed++
	}
	legendTop := top + 40 + float64(listed)*19
	fmt.Fprintf(b, `<line x1="%.3f" y1="%.3f" x2="%.3f" y2="%.3f" stroke="#6f5a3a" stroke-width="1.3"/>`, left, legendTop, left+22, legendTop)
	fmt.Fprintf(b, `<text x="%.3f" y="%.3f" fill="#3b4143" %s font-size="11">%s</text>`, left+28, legendTop+4, font,
		html.EscapeString(limbProfileText(language, "恒星方向月缘高度", "Limb height toward the star")))
	fmt.Fprintf(b, `<line x1="%.3f" y1="%.3f" x2="%.3f" y2="%.3f" stroke="#9b3f36" stroke-width="1.1" stroke-dasharray="5 2"/>`, left, legendTop+18, left+22, legendTop+18)
	fmt.Fprintf(b, `<text x="%.3f" y="%.3f" fill="#3b4143" %s font-size="11">%s</text>`, left+28, legendTop+22, font,
		html.EscapeString(limbProfileText(language, "恒星距参考圆月面（km）", "Star above the mean limb (km)")))
	fmt.Fprintf(b, `<rect x="%.3f" y="%.3f" width="22" height="10" fill="#3b4a5a" fill-opacity="0.16"/>`, left, legendTop+31)
	fmt.Fprintf(b, `<text x="%.3f" y="%.3f" fill="#3b4143" %s font-size="11">%s</text>`, left+28, legendTop+40, font,
		html.EscapeString(limbProfileText(language, "恒星被遮挡", "Star hidden")))
	return legendTop + 44
}

func writeLimbProfileFooter(b *strings.Builder, info moon.StarOccultationInfo, referenceKM, left, baseline, maxWidth float64, options StarOccultationLimbProfileSVGOptions) {
	note := limbProfileText(options.Language,
		"纵轴为月面上相对参考球 %.1f km 的高度；月缘按剖面表插值，短于采样步长的遮挡可能漏检。",
		"Heights are relative to the %.1f km reference sphere; the limb is interpolated from the profile table and events shorter than the step may be missed.")
	note = fmt.Sprintf(note, referenceKM)
	lines := svgchart.TruncateTextLines(svgchart.WrapTextBalanced(note, maxWidth, 10), maxWidth, 10, 2)
	lines = append(lines, svgchart.EllipsizeText(timenote.Scale(options.TimeScale, info.Greatest, options.Location, options.Language), maxWidth, 10))
	font := `font-family="Georgia, 'Times New Roman', serif"`
	for index, line := range lines {
		fmt.Fprintf(b, `<text x="%.3f" y="%.3f" fill="#747c7d" %s font-size="10">%s</text>`, left, baseline+float64(index)*14, font, html.EscapeString(line))
	}
}

// limbProfileTrackKM 返回恒星在天空平面上走过的弦长（km，按月面尺度）。
func limbProfileTrackKM(points []limbProfilePoint) float64 {
	first, last := points[0], points[len(points)-1]
	return math.Hypot(last.skyX-first.skyX, last.skyY-first.skyY)
}

// limbProfileExaggeration 纵向放大倍数：横向每像素代表的月面距离与纵向之比。
func limbProfileExaggeration(points []limbProfilePoint, widthPX, heightPX, spanKM float64) float64 {
	if widthPX <= 0 || heightPX <= 0 || spanKM <= 0 {
		return 0
	}
	return (limbProfileTrackKM(points) / widthPX) / (spanKM / heightPX)
}

// limbProfileCompass 把位置角换成八方位名。
func limbProfileCompass(positionAngle float64, language string) string {
	zh := []string{"北", "东北", "东", "东南", "南", "西南", "西", "西北"}
	en := []string{"N", "NE", "E", "SE", "S", "SW", "W", "NW"}
	index := int(math.Floor(math.Mod(positionAngle+22.5+720, 360) / 45))
	if language == starOccultationSVGLanguageEnglish {
		return en[index]
	}
	return zh[index]
}

func limbProfileDirectionText(segments [][]limbProfilePoint, hidden []moon.OccultationInterval, language string) string {
	if len(segments) == 0 {
		return ""
	}
	first := segments[0][0]
	lastSegment := segments[len(segments)-1]
	last := lastSegment[len(lastSegment)-1]
	if len(segments) == 1 {
		return fmt.Sprintf(limbProfileText(language,
			"恒星沿月缘自 PA %s（%s）移向 PA %s（%s），横向约 %.0f km；PA 自天北向东量",
			"The star runs along the limb from PA %s (%s) to PA %s (%s), about %.0f km; PA measured from north through east"),
			limbProfileFormatAngle(first.positionAngle), limbProfileCompass(first.positionAngle, language),
			limbProfileFormatAngle(last.positionAngle), limbProfileCompass(last.positionAngle, language), limbProfileTrackKM(segments[0]))
	}
	// 深掩时取首次掩入与最后复现时刻的位置角，而非两栏端点。
	immersion, emersion := first.positionAngle, last.positionAngle
	if len(hidden) > 0 {
		immersion = limbProfilePositionAngleAt(segments[0], hidden[0].Start)
		emersion = limbProfilePositionAngleAt(lastSegment, hidden[len(hidden)-1].End)
	}
	return fmt.Sprintf(limbProfileText(language,
		"恒星于 PA %s（%s）掩入、PA %s（%s）复现；上下两栏分别为掩入侧与复现侧，PA 自天北向东量",
		"The star disappears at PA %s (%s) and reappears at PA %s (%s); panels show the two sides; PA measured from north through east"),
		limbProfileFormatAngle(immersion), limbProfileCompass(immersion, language),
		limbProfileFormatAngle(emersion), limbProfileCompass(emersion, language))
}

// limbProfileFormatAngle 按一位小数格式化位置角，舍入到 360.0° 时归一为 0.0°。
func limbProfileFormatAngle(positionAngle float64) string {
	rounded := math.Round(positionAngle*10) / 10
	if rounded >= 360 {
		rounded -= 360
	}
	return fmt.Sprintf("%.1f°", rounded)
}

// writeLimbProfileInset 画掩甚时刻的站心月相缩略图（上北左东），加粗放大的月缘段并用箭头标出恒星运动方向。
func writeLimbProfileInset(b *strings.Builder, segments [][]limbProfilePoint, info moon.StarOccultationInfo, referenceKM, left, top, right, bottom float64, language string) {
	size := math.Min(right-left, bottom-top-26)
	if size < 60 || len(segments) == 0 {
		return
	}
	font := `font-family="Georgia, 'Times New Roman', serif"`
	fmt.Fprintf(b, `<text x="%.3f" y="%.3f" fill="#111111" %s font-size="12" font-weight="700">%s</text>`, left, top+10, font,
		html.EscapeString(svgchart.EllipsizeText(limbProfileText(language, "月缘位置（上北左东）", "On the limb (N up, E left)"), right-left, 12)))
	radius := size/2 - 16
	cx, cy := left+(right-left)/2, top+26+size/2
	writeLocalOccultationMoon(b, cx, cy, radius, "limb-profile-moon",
		localOccultationMoonAppearanceAt(info.Greatest, info.Observer), "limb-profile-moon-phase")
	fmt.Fprintf(b, `<text x="%.3f" y="%.3f" fill="#4f595b" %s font-size="10" text-anchor="middle">%s</text>`, cx, cy-radius-4, font, limbProfileText(language, "北", "N"))
	fmt.Fprintf(b, `<text x="%.3f" y="%.3f" fill="#4f595b" %s font-size="10" text-anchor="end">%s</text>`, cx-radius-4, cy+4, font, limbProfileText(language, "东", "E"))
	limbPoint := func(positionAngle, scale float64) (float64, float64) {
		angle := positionAngle * math.Pi / 180
		return cx - scale*math.Sin(angle), cy - scale*math.Cos(angle)
	}
	for _, points := range segments {
		first, last := points[0], points[len(points)-1]
		x0, y0 := limbPoint(first.positionAngle, radius)
		x1, y1 := limbPoint(last.positionAngle, radius)
		sweep := 0
		if signedPositionAngleStep(first.positionAngle, last.positionAngle) < 0 {
			sweep = 1
		}
		fmt.Fprintf(b, `<path d="M%.2f %.2f A%.2f %.2f 0 0 %d %.2f %.2f" fill="none" stroke="#e0a32e" stroke-width="3.6"/>`, x0, y0, radius, radius, sweep, x1, y1)
	}
	// 恒星轨迹按真实比例画在月盘坐标里，深掩时两段之间连成穿过月面的弦。
	kmToPX := radius / referenceKM
	firstPoint := segments[0][0]
	lastSegment := segments[len(segments)-1]
	lastPoint := lastSegment[len(lastSegment)-1]
	fmt.Fprintf(b, `<line x1="%.2f" y1="%.2f" x2="%.2f" y2="%.2f" stroke="#d0473b" stroke-width="1.6" marker-end="url(#limb-profile-arrow)"/>`,
		cx-firstPoint.skyX*kmToPX, cy-firstPoint.skyY*kmToPX, cx-lastPoint.skyX*kmToPX, cy-lastPoint.skyY*kmToPX)
}

// signedPositionAngleStep 返回自 from 到 to 的位置角变化（度），取 (−180, 180]；负值为顺时针（向西）。
func signedPositionAngleStep(from, to float64) float64 {
	delta := math.Mod(to-from+540, 360) - 180
	return delta
}

// limbProfileTickStep 选使相邻时刻标签不重叠的最小整刻度步长。
func limbProfileTickStep(spanSeconds, widthPX float64) time.Duration {
	for _, step := range []time.Duration{5 * time.Second, 10 * time.Second, 15 * time.Second, 20 * time.Second, 30 * time.Second,
		time.Minute, 2 * time.Minute, 5 * time.Minute, 10 * time.Minute, 15 * time.Minute, 30 * time.Minute, time.Hour} {
		if widthPX*step.Seconds()/spanSeconds >= 56 {
			return step
		}
	}
	return 2 * time.Hour
}

func limbProfileTickLayout(step time.Duration) string {
	if step%time.Minute == 0 {
		return "15:04"
	}
	return "15:04:05"
}

// limbProfilePositionAngleAt 在相邻样本间线性插值位置角，跨 0° 时按最短方向。
func limbProfilePositionAngleAt(points []limbProfilePoint, at time.Time) float64 {
	for index := 1; index < len(points); index++ {
		if points[index].t.Before(at) {
			continue
		}
		previous, next := points[index-1], points[index]
		span := next.t.Sub(previous.t).Seconds()
		if span <= 0 {
			return next.positionAngle
		}
		fraction := at.Sub(previous.t).Seconds() / span
		return math.Mod(previous.positionAngle+fraction*signedPositionAngleStep(previous.positionAngle, next.positionAngle)+360, 360)
	}
	return points[len(points)-1].positionAngle
}

// limbProfileTicks 返回 [start, end] 内按显示时区本地时钟对齐 step 的全部刻度；每个刻度按它自身所在时刻的时区偏移判断，跨夏令时切换也落在当地整刻度上。
func limbProfileTicks(start, end time.Time, step time.Duration, location *time.Location) []time.Time {
	// ceil 返回不早于 at、在当地时钟上是 step 整数倍的第一个时刻；结果若落到另一偏移下，按结果的偏移重算一次。
	ceil := func(at time.Time) time.Time {
		withOffset := func(at time.Time, offset int) time.Time {
			shift := time.Duration(offset) * time.Second
			tick := at.Add(shift).Truncate(step).Add(-shift)
			if tick.Before(at) {
				tick = tick.Add(step)
			}
			return tick
		}
		_, offset := at.In(location).Zone()
		tick := withOffset(at, offset)
		if _, actual := tick.In(location).Zone(); actual != offset {
			tick = withOffset(at, actual)
		}
		return tick
	}
	var ticks []time.Time
	for tick := ceil(start); !tick.After(end) && len(ticks) < 1000; tick = ceil(tick.Add(step)) {
		ticks = append(ticks, tick)
	}
	return ticks
}

// limbProfileClampLabel 返回中锚标签的 x，使宽 width 的标签整体落在 [left, right] 内。
func limbProfileClampLabel(x, width, left, right float64) float64 {
	return math.Max(left+width/2, math.Min(right-width/2, x))
}

// limbProfileReferenceKM 剖面结果声明的参考月球半径，旧结果未填时用 1737.4 km。
func limbProfileReferenceKM(profile moon.StarOccultationLimbProfileResult) float64 {
	if profile.ReferenceRadiusKM > 0 && !math.IsInf(profile.ReferenceRadiusKM, 0) {
		return profile.ReferenceRadiusKM
	}
	return 1737.4
}

// validateStarOccultationLimbProfile 拒绝 StarOccultationLimbProfile 给不出的数据：至多两栏、每栏至少两个样本且时间递增、角量有限、遮挡区间有序。
func validateStarOccultationLimbProfile(profile moon.StarOccultationLimbProfileResult) error {
	if len(profile.Segments) > 2 {
		return fmt.Errorf("%w: at most two segments, got %d", ErrInvalidStarOccultationLimbProfile, len(profile.Segments))
	}
	finiteValue := func(value float64) bool { return !math.IsNaN(value) && !math.IsInf(value, 0) }
	if !finiteValue(profile.ReferenceRadiusKM) || profile.ReferenceRadiusKM < 0 {
		return fmt.Errorf("%w: reference radius must be finite and non-negative", ErrInvalidStarOccultationLimbProfile)
	}
	for index, segment := range profile.Segments {
		if len(segment) < 2 {
			return fmt.Errorf("%w: segment %d has fewer than two samples", ErrInvalidStarOccultationLimbProfile, index)
		}
		for sampleIndex, sample := range segment {
			if !finiteValue(sample.SeparationArcsec) || !finiteValue(sample.LimbRadiusArcsec) ||
				!finiteValue(sample.MeanRadiusArcsec) || sample.MeanRadiusArcsec <= 0 || !finiteValue(sample.PositionAngleDeg) {
				return fmt.Errorf("%w: segment %d sample %d has non-finite geometry", ErrInvalidStarOccultationLimbProfile, index, sampleIndex)
			}
			if sampleIndex > 0 && !sample.Time.After(segment[sampleIndex-1].Time) {
				return fmt.Errorf("%w: segment %d samples must increase in time", ErrInvalidStarOccultationLimbProfile, index)
			}
		}
	}
	for index, interval := range profile.Hidden {
		if !interval.End.After(interval.Start) || (index > 0 && interval.Start.Before(profile.Hidden[index-1].End)) {
			return fmt.Errorf("%w: hidden intervals must be ordered", ErrInvalidStarOccultationLimbProfile)
		}
	}
	return nil
}
