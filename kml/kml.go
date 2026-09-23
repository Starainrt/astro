// Package kml 把 RFC 7946 GeoJSON 转成 KML 2.2，供 Google Earth 等客户端交互查看；除本库根包的
// 时标门面外只依赖标准库。输入可以是本库 geojson 包的输出，也可以是任何第三方 GeoJSON 文件。
// Package kml converts RFC 7946 GeoJSON into KML 2.2 for interactive clients such as Google Earth,
// using only the standard library apart from the root package's time-scale facade. The input may come
// from this library's geojson package or from any third-party GeoJSON file.
//
// 分层与配色都按要素的 role 属性派发：role 作 Folder，样式取下面的默认调色板，可用 Options.Styles 覆盖。
// 默认只有高亮中心带与掩星带填充，其余面只描轮廓，跨半球且互相重叠的重图层可用 Options.SkipRoles 整层丢弃。
// 时间：带 times 的 MultiLineString 拆成逐段 Placemark 并各带 TimeStamp，逐点的 time 属性同理；
// 集合带 time_scale=UT1 时先把标签折回 UTC 再写 <when>（KML 时间轴必须是 UTC），原口径记进 ExtendedData。
// Layers and colors are dispatched by each feature's role property, and times become KML TimeStamps.
package kml

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/starainrt/astro"
)

const namespace = "http://www.opengis.net/kml/2.2"

// 默认调色板，KML 的 aabbggrr 十六进制（与 CSS 的 rrggbb 顺序不同）：日食中心带红、中心线黑、
// 食甚等时线绿、食分线与掩星全掩/半掩带黄、日出日落阶段线与偏食区边界橙，其余中性灰。
const (
	// ColorRed 不透明红，日食中心带 / opaque red for the solar central band.
	ColorRed = "ff0000ff"
	// ColorBlack 不透明黑，中心线 / opaque black for the center line.
	ColorBlack = "ff000000"
	// ColorGreen 不透明绿，日食食甚等时线 / opaque green for solar greatest-time isochrones.
	ColorGreen = "ff008000"
	// ColorYellow 不透明黄，食分线与掩星全掩/半掩带 / opaque yellow for magnitude contours and occultation bands.
	ColorYellow = "ff00ffff"
	// ColorDarkBlue 深蓝，导出常量保留但默认调色板已不用它 / kept exported, unused by the default palette.
	ColorDarkBlue = "ff8b0000"
	// ColorOrange 不透明橙，日出日落阶段线与偏食区边界 / opaque orange for rise/set phase lines and the partial-visibility edge.
	ColorOrange = "ff008cff"
	// ColorGray 灰，未指定家族的中性图层 / gray for unclassified layers.
	ColorGray = "ff808080"

	redFill    = "590000ff"
	yellowFill = "4d00ffff"
	grayFill   = "33808080"
	// 仅见半影的两条带用与 SVG 同一族色相：月出偏蓝、月落偏紫红。
	penumbraMoonriseColor = "ffcc8874"
	penumbraMoonsetColor  = "ffb57fab"
	penumbraMoonriseFill  = "4dcc8874"
	penumbraMoonsetFill   = "4db57fab"
)

const (
	eventSolarEclipse     = "solar-eclipse"
	eventLunarEclipse     = "lunar-eclipse"
	eventLunarOccultation = "lunar-occultation"
)

// Style 是一个 KML 样式；零值字段使用内置调色板。颜色为 KML 的 aabbggrr 十六进制（ff0000ff 是不透明红）。
// Style describes one KML style; zero fields fall back to the built-in palette. Colors use KML's
// aabbggrr hex order (ff0000ff is opaque red).
type Style struct {
	// LineColor 是线色；空值用调色板 / line color, empty for the palette default.
	LineColor string
	// FillColor 是面填充色；空值用调色板（只有中心带、掩星带这类高亮面才有填充），点与线忽略该字段。
	// FillColor is the polygon fill; empty uses the palette, where only highlight bands carry a fill.
	FillColor string
	// NoFill 为 true 时强制不填充，压过调色板与 FillColor / forces no fill, overriding the palette and FillColor.
	NoFill bool
	// LineWidth 是线宽（像素）；<=0 用调色板默认 / line width in pixels.
	LineWidth float64
}

// Options 控制输出 / Options controls the output.
type Options struct {
	// Name 是 Document 名；空值按事件类型与集合里最早的时刻推导。
	// Name is the Document name; empty derives it from the event type and the earliest instant.
	Name string
	// Language 选择图层名语言："zh"（默认）或 "en"；未收录的角色保留原始 role id。
	// Language selects layer names: "zh" (default) or "en"; unknown roles keep their raw id.
	Language string
	// Styles 覆盖默认样式，键为 role，或更具体的 "event/role"（后者优先）。
	// Styles overrides the palette, keyed by role or by the more specific "event/role".
	Styles map[string]Style
	// NoLookAt 为 true 时不输出 Document/LookAt / omits the Document LookAt.
	NoLookAt bool
	// SkipRoles 列出要整层丢弃的 role。日食的 partial-band 与 partial-footprint 覆盖大半个地球，
	// 光栅化很重且互相重叠，移动端客户端容易卡顿，按需丢弃即可 / roles to drop entirely.
	SkipRoles []string
	// FillContext 为 true 时中性面图层也填充；默认只填高亮中心带与掩星带，避免大面积半透明面拖慢客户端。
	// FillContext also fills neutral context polygons; by default only highlight bands are filled.
	FillContext bool
	// NoTimes 为 true 时不输出任何时间元素，得到一份静态叠加；客户端的时间轴停在事件窗口之外时会
	// 隐藏全部带时间的要素（只剩没有时间戳的面），静态查看时用它绕开该行为。
	// NoTimes omits every time primitive, producing a static overlay; clients whose time slider sits
	// outside the event window hide all timed features, and this option sidesteps that behaviour.
	NoTimes bool
}

type document struct {
	Name         string        `xml:"name,omitempty"`
	Description  string        `xml:"description,omitempty"`
	LookAt       *lookAt       `xml:"LookAt,omitempty"`
	TimeSpan     *timeSpan     `xml:"TimeSpan,omitempty"`
	Styles       []kmlStyle    `xml:"Style"`
	ExtendedData *extendedData `xml:"ExtendedData,omitempty"`
	Folders      []folder      `xml:"Folder"`
}

type folder struct {
	Name       string      `xml:"name,omitempty"`
	Placemarks []placemark `xml:"Placemark"`
}

type placemark struct {
	Name         string        `xml:"name,omitempty"`
	TimeStamp    *timeStamp    `xml:"TimeStamp,omitempty"`
	StyleURL     string        `xml:"styleUrl,omitempty"`
	ExtendedData *extendedData `xml:"ExtendedData,omitempty"`
	geometry
}

type timeStamp struct {
	When string `xml:"when"`
}

type timeSpan struct {
	Begin string `xml:"begin"`
	End   string `xml:"end"`
}

// 字段顺序就是 KML 2.2 的 LookAt sequence（heading、tilt、range），顺序错了严格客户端与 XSD 校验会拒绝整个 LookAt。
type lookAt struct {
	Longitude float64 `xml:"longitude"`
	Latitude  float64 `xml:"latitude"`
	Heading   float64 `xml:"heading"`
	Tilt      float64 `xml:"tilt"`
	Range     float64 `xml:"range"`
}

// 字段顺序就是 KML 2.2 的 Style sequence（IconStyle、LabelStyle、LineStyle、PolyStyle），
// 顺序错了严格客户端会整段丢弃样式。
type kmlStyle struct {
	ID        string     `xml:"id,attr"`
	IconStyle *iconStyle `xml:"IconStyle,omitempty"`
	LineStyle *lineStyle `xml:"LineStyle,omitempty"`
	PolyStyle *polyStyle `xml:"PolyStyle,omitempty"`
}

type lineStyle struct {
	Color string  `xml:"color"`
	Width float64 `xml:"width"`
}

type polyStyle struct {
	Color   string `xml:"color"`
	Fill    int    `xml:"fill"`
	Outline int    `xml:"outline"`
}

type iconStyle struct {
	Color string  `xml:"color"`
	Scale float64 `xml:"scale"`
}

type dataEntry struct {
	Name  string `xml:"name,attr"`
	Value string `xml:"value"`
}

type extendedData struct {
	Data []dataEntry `xml:"Data"`
}

type point struct {
	AltitudeMode string `xml:"altitudeMode"`
	Coordinates  string `xml:"coordinates"`
}

type lineString struct {
	Tessellate   int    `xml:"tessellate"`
	AltitudeMode string `xml:"altitudeMode"`
	Coordinates  string `xml:"coordinates"`
}

type linearRing struct {
	Coordinates string `xml:"coordinates"`
}

type boundary struct {
	Ring linearRing `xml:"LinearRing"`
}

type polygon struct {
	Tessellate      int        `xml:"tessellate"`
	AltitudeMode    string     `xml:"altitudeMode"`
	OuterBoundaryIs *boundary  `xml:"outerBoundaryIs,omitempty"`
	InnerBoundaries []boundary `xml:"innerBoundaryIs,omitempty"`
}

type geometry struct {
	Point         *point         `xml:"Point,omitempty"`
	LineString    *lineString    `xml:"LineString,omitempty"`
	Polygon       *polygon       `xml:"Polygon,omitempty"`
	MultiGeometry *multiGeometry `xml:"MultiGeometry,omitempty"`
}

type multiGeometry struct {
	Points   []point      `xml:"Point,omitempty"`
	Lines    []lineString `xml:"LineString,omitempty"`
	Polygons []polygon    `xml:"Polygon,omitempty"`
}

// empty 报告几何在去重与接缝切除之后是否还有可画的部分。
func (g geometry) empty() bool {
	switch {
	case g.Point != nil, g.LineString != nil, g.Polygon != nil:
		return false
	case g.MultiGeometry != nil:
		multi := g.MultiGeometry
		return len(multi.Points) == 0 && len(multi.Lines) == 0 && len(multi.Polygons) == 0
	}
	return true
}

// rawCollection 是解析 RFC 7946 的最小结构；coordinates 按几何类型分别解码。
type rawCollection struct {
	Type      string       `json:"type"`
	TimeScale string       `json:"time_scale"`
	Features  []rawFeature `json:"features"`
}

type rawFeature struct {
	Type       string                 `json:"type"`
	Properties map[string]interface{} `json:"properties"`
	Geometry   rawGeometry            `json:"geometry"`
}

type rawGeometry struct {
	Type        string          `json:"type"`
	Coordinates json.RawMessage `json:"coordinates"`
	Geometries  []rawGeometry   `json:"geometries"`
}

// element 是规范化的几何片段：三种类型互斥，when 是该片段的时间戳（可空）。
type element struct {
	kind  string
	point []float64
	line  [][]float64
	rings [][][]float64
	when  string
}

type kmlRoot struct {
	XMLName  xml.Name `xml:"kml"`
	Xmlns    string   `xml:"xmlns,attr"`
	Document document `xml:"Document"`
}

// FromGeoJSON 把一份 RFC 7946 FeatureCollection 转成 KML 2.2 字节 / converts one RFC 7946 FeatureCollection into KML 2.2 bytes.
func FromGeoJSON(data []byte, options Options) ([]byte, error) {
	var collection rawCollection
	if err := json.Unmarshal(data, &collection); err != nil {
		return nil, fmt.Errorf("kml: decode GeoJSON: %w", err)
	}
	if collection.Type != "FeatureCollection" {
		return nil, fmt.Errorf("kml: GeoJSON type %q is not a FeatureCollection", collection.Type)
	}
	built, err := buildDocument(&collection, options)
	if err != nil {
		return nil, err
	}
	body, err := xml.Marshal(kmlRoot{Xmlns: namespace, Document: *built})
	if err != nil {
		return nil, fmt.Errorf("kml: encode: %w", err)
	}
	return append([]byte(xml.Header), body...), nil
}

type preparedFeature struct {
	event    string
	role     string
	props    map[string]interface{}
	groups   [][]element
	polygon  bool
	timed    bool
	earliest time.Time
	latest   time.Time
}

func buildDocument(collection *rawCollection, options Options) (*document, error) {
	language := options.Language
	if language == "" {
		language = "zh"
	}
	features := make([]preparedFeature, 0, len(collection.Features))
	for index := range collection.Features {
		prepared, err := prepareFeature(collection.Features[index], collection.TimeScale)
		if err != nil {
			return nil, fmt.Errorf("kml: feature %d: %w", index, err)
		}
		features = append(features, *prepared)
	}

	skipRoles := map[string]bool{}
	for _, role := range options.SkipRoles {
		skipRoles[role] = true
	}
	kept := make([]preparedFeature, 0, len(features))
	for _, feature := range features {
		if !skipRoles[feature.role] {
			kept = append(kept, feature)
		}
	}
	if len(collection.Features) > 0 && len(kept) == 0 {
		return nil, fmt.Errorf("kml: every feature was dropped by SkipRoles")
	}
	documentEntries := commonProperties(kept)
	if collection.TimeScale != "" {
		documentEntries["time_scale"] = collection.TimeScale
	}

	// 样式与图层都用 (event, role) 结构体做键：两者都可能含 '-'，拼成字符串后
	// (foo-bar, baz) 与 (foo, bar-baz) 会撞成同一个键，把两个图层和样式并成一个。
	styles := map[styleRef]bool{}
	polygonal := map[styleRef]bool{}
	folderKeys := map[styleRef][]placemark{}
	folderOrder := []styleRef{}
	eventCount := map[string]bool{}
	// 样式 id 必须同时合法且唯一：role 来自任意输入，不同键也可能清洗成同一个 id。
	styleIDs := map[styleRef]string{}
	usedStyleIDs := map[string]bool{}
	styleIDFor := func(ref styleRef) string {
		if id, ok := styleIDs[ref]; ok {
			return id
		}
		base := sanitizeStyleID(ref.id())
		id := base
		for suffix := 2; usedStyleIDs[id]; suffix++ {
			id = base + "-" + strconv.Itoa(suffix)
		}
		usedStyleIDs[id] = true
		styleIDs[ref] = id
		return id
	}
	var lookAtGroups []roleGeometry
	var earliest, latest time.Time
	drawn := 0
	for _, feature := range kept {
		ref := styleRef{event: feature.event, role: feature.role}
		featureStyle := resolveStyle(feature.event, feature.role, feature.polygon, options)
		placemarks, err := placemarksFor(feature, styleIDFor(ref), documentEntries, options.NoTimes, featureStyle.FillColor != "")
		if err != nil {
			return nil, err
		}
		if len(placemarks) == 0 {
			continue
		}
		drawn++
		eventCount[feature.event] = true
		styles[ref] = true
		if feature.polygon {
			polygonal[ref] = true
		}
		if !feature.earliest.IsZero() && (earliest.IsZero() || feature.earliest.Before(earliest)) {
			earliest = feature.earliest
		}
		if !feature.latest.IsZero() && (latest.IsZero() || feature.latest.After(latest)) {
			latest = feature.latest
		}
		if _, ok := folderKeys[ref]; !ok {
			folderOrder = append(folderOrder, ref)
		}
		folderKeys[ref] = append(folderKeys[ref], placemarks...)
		lookAtGroups = append(lookAtGroups, roleGeometry{role: feature.role, groups: feature.groups})
	}
	sort.Slice(folderOrder, func(left, right int) bool { return styleLess(folderOrder[left], folderOrder[right]) })
	if drawn == 0 && len(collection.Features) > 0 {
		return nil, fmt.Errorf("kml: no feature produced drawable geometry")
	}

	built := &document{
		Folders: make([]folder, 0, len(folderOrder)),
	}
	for _, ref := range folderOrder {
		name := localizedRole(ref.event, ref.role, language)
		if len(eventCount) > 1 && ref.event != "" {
			name = localizedEvent(ref.event, language) + " · " + name
		}
		built.Folders = append(built.Folders, folder{Name: name, Placemarks: folderKeys[ref]})
	}

	refs := make([]styleRef, 0, len(styles))
	for ref := range styles {
		refs = append(refs, ref)
	}
	sort.Slice(refs, func(left, right int) bool { return styleLess(refs[left], refs[right]) })
	built.Styles = make([]kmlStyle, 0, len(refs))
	for _, ref := range refs {
		resolved := resolveStyle(ref.event, ref.role, polygonal[ref], options)
		style := kmlStyle{
			ID:        styleIDFor(ref),
			IconStyle: &iconStyle{Color: resolved.LineColor, Scale: 0.8},
			LineStyle: &lineStyle{Color: resolved.LineColor, Width: resolved.LineWidth},
		}
		if resolved.FillColor != "" {
			// 面的描边交给同一 Placemark 里的折线：折线能在反经线接缝处断开，而面必须闭合，
			// 闭合边会在换日线上留下一条染色直线。
			style.PolyStyle = &polyStyle{Color: resolved.FillColor, Fill: 1, Outline: 0}
		} else if polygonal[ref] {
			// 面要素必须显式写 fill=0：缺 PolyStyle 时客户端会套用自己的默认面样式（Google Earth 是白填充），
			// 把整个可见域盖成一片白，底下的中心带与等值线全部看不见。
			style.PolyStyle = &polyStyle{Color: resolved.LineColor, Fill: 0, Outline: 1}
		}
		built.Styles = append(built.Styles, style)
	}

	if len(documentEntries) > 0 {
		built.ExtendedData = &extendedData{Data: sortedDataEntries(documentEntries)}
	}
	built.Name = documentName(options, kept, earliest)
	built.Description = descriptionText(options, language, collection.TimeScale)
	if !options.NoLookAt {
		built.LookAt = lookAtFor(lookAtElements(lookAtGroups))
	}
	if !options.NoTimes && !earliest.IsZero() && !latest.IsZero() {
		// Document 级 TimeSpan 声明整个集合的时间窗口，免得客户端的时间轴默认停在窗口之外把带时间的要素全藏起来。
		built.TimeSpan = &timeSpan{
			Begin: earliest.UTC().Format(time.RFC3339Nano),
			End:   latest.UTC().Format(time.RFC3339Nano),
		}
	}
	return built, nil
}

func prepareFeature(feature rawFeature, timeScale string) (*preparedFeature, error) {
	props := feature.Properties
	if props == nil {
		props = map[string]interface{}{}
	}
	event, _ := props["event"].(string)
	role, _ := props["role"].(string)
	if role == "" {
		role = "unclassified"
	}
	elements, err := decodeGeometry(feature.Geometry)
	if err != nil {
		return nil, err
	}
	polygon := false
	for _, item := range elements {
		if item.kind == "Polygon" {
			polygon = true
		}
	}

	prepared := &preparedFeature{event: event, role: role, props: props, polygon: polygon}
	times, hasTimes, err := decodeTimes(props["times"])
	if err != nil {
		return nil, err
	}
	single, _ := props["time"].(string)
	switch {
	case hasTimes:
		lineOnly := true
		for _, item := range elements {
			if item.kind != "LineString" {
				lineOnly = false
			}
		}
		if !lineOnly || len(times) != len(elements) {
			return nil, fmt.Errorf("times array (%d) does not align with %d line segments", len(times), len(elements))
		}
		for index := range elements {
			when, err := normalizeWhen(times[index][0], timeScale)
			if err != nil {
				return nil, err
			}
			elements[index].when = when
			prepared.timed = true
			prepared.earliest = minTime(prepared.earliest, when)
			prepared.latest = maxTime(prepared.latest, when)
			prepared.groups = append(prepared.groups, []element{elements[index]})
		}
	case single != "":
		when, err := normalizeWhen(single, timeScale)
		if err != nil {
			return nil, err
		}
		// 一个标量 time 覆盖整组几何：它们进同一个 Placemark，时刻写在它的 TimeStamp 上。
		for index := range elements {
			elements[index].when = when
		}
		prepared.timed = true
		prepared.earliest = minTime(prepared.earliest, when)
		prepared.latest = maxTime(prepared.latest, when)
		prepared.groups = append(prepared.groups, elements)
	default:
		prepared.groups = append(prepared.groups, elements)
	}
	return prepared, nil
}

func placemarksFor(feature preparedFeature, styleID string, documentEntries map[string]string, noTimes, filled bool) ([]placemark, error) {
	entries := encodeProperties(feature.props, true)
	// times 是逐段时刻的嵌套数组：它随 TimeStamp 走，NoTimes 下也不该把坐标级数据抄一遍。
	delete(entries, "times")
	if feature.timed && !noTimes {
		// time 已经变成 TimeStamp，不再重复写进 ExtendedData。
		delete(entries, "time")
	}
	placemarks := make([]placemark, 0, len(feature.groups))
	for _, group := range feature.groups {
		value, err := geometryFor(group, filled)
		if err != nil {
			return nil, err
		}
		if value.empty() {
			continue
		}
		current := placemark{StyleURL: "#" + styleID, geometry: value}
		if when := uniformWhen(group); when != "" {
			if !noTimes {
				current.TimeStamp = &timeStamp{When: when}
				current.Name = placemarkName(feature.props, group[0].when)
			} else if label, ok := feature.props["label"].(string); ok && label != "" {
				current.Name = label
			}
		}
		if len(entries) > 0 {
			local := map[string]string{}
			for name, encoded := range entries {
				if shared, ok := documentEntries[name]; ok && shared == encoded {
					continue
				}
				local[name] = encoded
			}
			if len(local) > 0 {
				current.ExtendedData = &extendedData{Data: sortedDataEntries(local)}
			}
		}
		placemarks = append(placemarks, current)
	}
	return placemarks, nil
}

func placemarkName(props map[string]interface{}, when string) string {
	if label, ok := props["label"].(string); ok && label != "" {
		return label
	}
	return compactWhen(when)
}

func compactWhen(when string) string {
	return strings.Replace(truncateToSecond(when), "T", " ", 1)
}

// truncateToSecond 去掉小数秒，仅用于 Placemark 名字；TimeStamp 仍写完整精度。
func truncateToSecond(when string) string {
	dot := strings.IndexByte(when, '.')
	if dot < 0 {
		return when
	}
	end := len(when)
	for index := dot + 1; index < len(when); index++ {
		if when[index] == 'Z' || when[index] == '+' || when[index] == '-' {
			end = index
			break
		}
	}
	return when[:dot] + when[end:]
}

func normalizeWhen(when, timeScale string) (string, error) {
	parsed, err := time.Parse(time.RFC3339Nano, when)
	if err != nil {
		return "", fmt.Errorf("time %q is not RFC 3339: %w", when, err)
	}
	if timeScale == "UT1" {
		// GeoJSON 里的 UT1 标签带 Z 后缀但不等于 UTC，KML 的时间轴必须是 UTC，按同一 ΔT 模型折回。
		parsed = astro.UTCFromUT1(parsed)
	}
	// KML 的 when 是 xsd:dateTime，允许小数秒；截断到整秒会把每个时刻整体提前最多 1 秒。
	return parsed.UTC().Format(time.RFC3339Nano), nil
}

func maxTime(current time.Time, when string) time.Time {
	parsed, err := time.Parse(time.RFC3339, when)
	if err != nil {
		return current
	}
	if current.IsZero() || parsed.After(current) {
		return parsed
	}
	return current
}

func minTime(current time.Time, when string) time.Time {
	parsed, err := time.Parse(time.RFC3339, when)
	if err != nil {
		return current
	}
	if current.IsZero() || parsed.Before(current) {
		return parsed
	}
	return current
}

func decodeTimes(value interface{}) ([][]string, bool, error) {
	if value == nil {
		return nil, false, nil
	}
	raw, ok := value.([]interface{})
	if !ok {
		return nil, false, fmt.Errorf("times property has type %T", value)
	}
	times := make([][]string, 0, len(raw))
	for _, segment := range raw {
		list, ok := segment.([]interface{})
		if !ok {
			return nil, false, fmt.Errorf("times segment has type %T", segment)
		}
		entries := make([]string, 0, len(list))
		for _, entry := range list {
			text, ok := entry.(string)
			if !ok {
				return nil, false, fmt.Errorf("times entry has type %T", entry)
			}
			entries = append(entries, text)
		}
		if len(entries) == 0 {
			return nil, false, fmt.Errorf("times segment is empty")
		}
		times = append(times, entries)
	}
	if len(times) == 0 {
		return nil, false, fmt.Errorf("times array is empty")
	}
	return times, true, nil
}

func decodeGeometry(raw rawGeometry) ([]element, error) {
	switch raw.Type {
	case "Point":
		var position []float64
		if err := json.Unmarshal(raw.Coordinates, &position); err != nil {
			return nil, fmt.Errorf("Point coordinates: %w", err)
		}
		return []element{{kind: "Point", point: position}}, nil
	case "MultiPoint":
		var positions [][]float64
		if err := json.Unmarshal(raw.Coordinates, &positions); err != nil {
			return nil, fmt.Errorf("MultiPoint coordinates: %w", err)
		}
		elements := make([]element, 0, len(positions))
		for _, position := range positions {
			elements = append(elements, element{kind: "Point", point: position})
		}
		return elements, nil
	case "LineString":
		var line [][]float64
		if err := json.Unmarshal(raw.Coordinates, &line); err != nil {
			return nil, fmt.Errorf("LineString coordinates: %w", err)
		}
		return []element{{kind: "LineString", line: line}}, nil
	case "MultiLineString":
		var lines [][][]float64
		if err := json.Unmarshal(raw.Coordinates, &lines); err != nil {
			return nil, fmt.Errorf("MultiLineString coordinates: %w", err)
		}
		elements := make([]element, 0, len(lines))
		for _, line := range lines {
			elements = append(elements, element{kind: "LineString", line: line})
		}
		return elements, nil
	case "Polygon":
		var rings [][][]float64
		if err := json.Unmarshal(raw.Coordinates, &rings); err != nil {
			return nil, fmt.Errorf("Polygon coordinates: %w", err)
		}
		return []element{{kind: "Polygon", rings: rings}}, nil
	case "MultiPolygon":
		var polygons [][][][]float64
		if err := json.Unmarshal(raw.Coordinates, &polygons); err != nil {
			return nil, fmt.Errorf("MultiPolygon coordinates: %w", err)
		}
		elements := make([]element, 0, len(polygons))
		for _, rings := range polygons {
			elements = append(elements, element{kind: "Polygon", rings: rings})
		}
		return elements, nil
	case "GeometryCollection":
		elements := []element{}
		for _, child := range raw.Geometries {
			decoded, err := decodeGeometry(child)
			if err != nil {
				return nil, err
			}
			elements = append(elements, decoded...)
		}
		if len(elements) == 0 {
			return nil, fmt.Errorf("GeometryCollection is empty")
		}
		return elements, nil
	case "":
		return nil, fmt.Errorf("geometry is missing")
	default:
		return nil, fmt.Errorf("unsupported geometry type %q", raw.Type)
	}
}

func geometryFor(elements []element, filled bool) (geometry, error) {
	if len(elements) == 1 {
		return singleGeometry(elements[0], filled)
	}
	multi := &multiGeometry{}
	for _, item := range elements {
		switch item.kind {
		case "Point":
			text, err := pointText(item.point)
			if err != nil {
				return geometry{}, err
			}
			multi.Points = append(multi.Points, point{AltitudeMode: "clampToGround", Coordinates: text})
		case "LineString":
			parts, err := seamCutLine(item.line)
			if err != nil {
				return geometry{}, err
			}
			for _, part := range parts {
				text, err := lineText(part)
				if err != nil {
					return geometry{}, err
				}
				multi.Lines = append(multi.Lines, lineString{Tessellate: 1, AltitudeMode: "clampToGround", Coordinates: text})
			}
		case "Polygon":
			outline, err := outlineGeometry(item.rings)
			if err != nil {
				return geometry{}, err
			}
			switch {
			case outline.LineString != nil:
				multi.Lines = append(multi.Lines, *outline.LineString)
			case outline.MultiGeometry != nil:
				multi.Lines = append(multi.Lines, outline.MultiGeometry.Lines...)
			}
			if !filled {
				continue
			}
			value, err := polygonFor(item.rings)
			if err != nil {
				return geometry{}, err
			}
			multi.Polygons = append(multi.Polygons, value)
		}
	}
	return geometry{MultiGeometry: multi}, nil
}

func singleGeometry(item element, filled bool) (geometry, error) {
	switch item.kind {
	case "Point":
		text, err := pointText(item.point)
		if err != nil {
			return geometry{}, err
		}
		return geometry{Point: &point{AltitudeMode: "clampToGround", Coordinates: text}}, nil
	case "LineString":
		parts, err := seamCutLine(item.line)
		if err != nil {
			return geometry{}, err
		}
		lines := make([]lineString, 0, len(parts))
		for _, part := range parts {
			text, err := lineText(part)
			if err != nil {
				return geometry{}, err
			}
			lines = append(lines, lineString{Tessellate: 1, AltitudeMode: "clampToGround", Coordinates: text})
		}
		switch len(lines) {
		case 0:
			return geometry{}, nil
		case 1:
			return geometry{LineString: &lines[0]}, nil
		}
		return geometry{MultiGeometry: &multiGeometry{Lines: lines}}, nil
	case "Polygon":
		outline, err := outlineGeometry(item.rings)
		if err != nil {
			return geometry{}, err
		}
		if !filled {
			return outline, nil
		}
		value, err := polygonFor(item.rings)
		if err != nil {
			return geometry{}, err
		}
		multi := &multiGeometry{Polygons: []polygon{value}}
		switch {
		case outline.LineString != nil:
			multi.Lines = append(multi.Lines, *outline.LineString)
		case outline.MultiGeometry != nil:
			multi.Lines = append(multi.Lines, outline.MultiGeometry.Lines...)
		}
		return geometry{MultiGeometry: multi}, nil
	}
	return geometry{}, fmt.Errorf("unsupported geometry element %q", item.kind)
}

// outlineGeometry 把只描轮廓的面画成折线。反经线接缝（两端都在 ±180 的经线段）是拆环的产物、不是真实边界，
// 必须在它处断开，否则图上会多出一条沿换日线的直线。
func outlineGeometry(rings [][][]float64) (geometry, error) {
	lines := []lineString{}
	for _, ring := range rings {
		parts, err := outlineLines(ring)
		if err != nil {
			return geometry{}, err
		}
		for _, part := range parts {
			text, err := lineText(part)
			if err != nil {
				return geometry{}, err
			}
			lines = append(lines, lineString{Tessellate: 1, AltitudeMode: "clampToGround", Coordinates: text})
		}
	}
	switch len(lines) {
	case 0:
		return geometry{}, nil
	case 1:
		return geometry{LineString: &lines[0]}, nil
	}
	return geometry{MultiGeometry: &multiGeometry{Lines: lines}}, nil
}

func outlineLines(ring [][]float64) ([][][]float64, error) {
	normalized, err := normalizeRing(ring, true)
	if err != nil {
		return nil, err
	}
	points := normalized
	if len(points) > 1 && equalPosition(points[0], points[len(points)-1]) {
		points = points[:len(points)-1]
	}
	// 从某条接缝边的后一个顶点起绕环：顺序扫描时"环尾回到环首"那一段真实边不会被检查到，
	// 接缝在别处时它就被整段漏掉。旋转后每条真实边都在一圈之内，末尾那条正好是接缝本身。
	start := -1
	for index := range points {
		if isSeamSegment(points[index], points[(index+1)%len(points)]) {
			start = (index + 1) % len(points)
			break
		}
	}
	if start < 0 {
		// 不碰接缝的环按闭合折线画，保持原来的形状。
		closed := append([][]float64{}, points...)
		return [][][]float64{append(closed, append([]float64(nil), points[0]...))}, nil
	}
	ordered := append(append([][]float64{}, points[start:]...), points[:start]...)
	parts := [][][]float64{}
	current := [][]float64{}
	for index := range ordered {
		current = append(current, ordered[index])
		if index == len(ordered)-1 {
			break
		}
		if isSeamSegment(ordered[index], ordered[index+1]) {
			if len(current) > 1 {
				parts = append(parts, current)
			}
			current = nil
		}
	}
	if len(current) > 1 {
		parts = append(parts, current)
	}
	// 整个环都落在 ±180 上时没有可画的真实边，交给调用方按"空几何"跳过。
	return parts, nil
}

// seamCutLine 把折线在反经线接缝处断开：两端都落在 ±180 的段是拆环产物、不是真实边界，
// 描出来就是一条沿换日线的直线（本库 geojson 的 band-outline 正是这种闭合环线）。
// 不碰接缝的折线原样返回一段，保持既有输出。
func seamCutLine(line [][]float64) ([][][]float64, error) {
	points := make([][]float64, 0, len(line))
	for _, position := range line {
		normalized, err := normalizePosition(position)
		if err != nil {
			return nil, err
		}
		if len(points) > 0 && equalPosition(points[len(points)-1], normalized) {
			continue
		}
		points = append(points, normalized)
	}
	if len(points) < 2 {
		// 只剩一个位置的路径片段画不出线（本库 geojson 用两个相同坐标表示它），跳过而不是让整份转换失败。
		return nil, nil
	}
	closed := equalPosition(points[0], points[len(points)-1])
	parts := [][][]float64{}
	current := [][]float64{}
	for index := range points {
		current = append(current, points[index])
		if index == len(points)-1 && !closed {
			break
		}
		if !isSeamSegment(points[index], points[(index+1)%len(points)]) {
			continue
		}
		if len(current) > 1 {
			parts = append(parts, current)
		}
		current = nil
	}
	if len(current) > 1 {
		parts = append(parts, current)
	}
	if len(parts) == 0 {
		// 整条线都落在 ±180 上：那是真实边界（例如极区被反经线裁出的边），不是拆环残留，
		// 原样保留比让整次转换失败更合理。
		return [][][]float64{points}, nil
	}
	return parts, nil
}

func isSeamSegment(from, to []float64) bool {
	const tolerance = 1e-9
	return math.Abs(math.Abs(from[0])-180) < tolerance && math.Abs(math.Abs(to[0])-180) < tolerance
}

func polygonFor(rings [][][]float64) (polygon, error) {
	if len(rings) == 0 {
		return polygon{}, fmt.Errorf("polygon has no rings")
	}
	outer, err := normalizeRing(rings[0], true)
	if err != nil {
		return polygon{}, err
	}
	value := polygon{
		Tessellate:      1,
		AltitudeMode:    "clampToGround",
		OuterBoundaryIs: &boundary{Ring: linearRing{Coordinates: textForPositions(outer)}},
	}
	for _, ring := range rings[1:] {
		hole, err := normalizeRing(ring, false)
		if err != nil {
			return polygon{}, err
		}
		value.InnerBoundaries = append(value.InnerBoundaries, boundary{Ring: linearRing{Coordinates: textForPositions(hole)}})
	}
	return value, nil
}

// normalizeRing 归一化环：位置必须有限、经度回绕到 (−180,180]、纬度在 ±90 内；闭合，并按 KML 要求
// 把外环摆成逆时针、洞摆成顺时针。
func normalizeRing(ring [][]float64, outer bool) ([][]float64, error) {
	positions := make([][]float64, 0, len(ring)+1)
	for _, position := range ring {
		normalized, err := normalizePosition(position)
		if err != nil {
			return nil, err
		}
		if len(positions) > 0 && equalPosition(positions[len(positions)-1], normalized) {
			continue
		}
		positions = append(positions, normalized)
	}
	if len(positions) > 1 && equalPosition(positions[0], positions[len(positions)-1]) {
		positions = positions[:len(positions)-1]
	}
	if len(positions) < 3 {
		return nil, fmt.Errorf("polygon ring requires at least three distinct positions")
	}
	area := signedArea(positions)
	if math.Abs(area) < 1e-12 {
		return nil, fmt.Errorf("polygon ring has zero area")
	}
	if (outer && area < 0) || (!outer && area > 0) {
		for left, right := 0, len(positions)-1; left < right; left, right = left+1, right-1 {
			positions[left], positions[right] = positions[right], positions[left]
		}
	}
	return append(positions, append([]float64(nil), positions[0]...)), nil
}

func normalizePosition(position []float64) ([]float64, error) {
	if len(position) < 2 {
		return nil, fmt.Errorf("position requires longitude and latitude")
	}
	longitude, latitude := position[0], position[1]
	if math.IsNaN(longitude) || math.IsInf(longitude, 0) || math.IsNaN(latitude) || math.IsInf(latitude, 0) {
		return nil, fmt.Errorf("position is not finite")
	}
	if latitude < -90 || latitude > 90 {
		return nil, fmt.Errorf("latitude %.6f is outside ±90", latitude)
	}
	// ±180 必须原样保留：把 +180 折成 −180，会让已经按反经线拆好的环在经度上多出 360° 的边，
	// 客户端会把这种环填成一条环绕全球的假带。
	if longitude > 180 || longitude < -180 {
		longitude = math.Mod(longitude+180, 360)
		if longitude < 0 {
			longitude += 360
		}
		longitude -= 180
	}
	return []float64{longitude, latitude}, nil
}

func equalPosition(a, b []float64) bool {
	return len(a) >= 2 && len(b) >= 2 && a[0] == b[0] && a[1] == b[1]
}

func signedArea(positions [][]float64) float64 {
	area := 0.0
	for index, position := range positions {
		next := positions[(index+1)%len(positions)]
		area += position[0]*next[1] - next[0]*position[1]
	}
	return area / 2
}

func pointText(position []float64) (string, error) {
	normalized, err := normalizePosition(position)
	if err != nil {
		return "", err
	}
	return textForPositions([][]float64{normalized}), nil
}

func lineText(line [][]float64) (string, error) {
	if len(line) < 2 {
		return "", fmt.Errorf("line requires at least two positions")
	}
	positions := make([][]float64, 0, len(line))
	for _, position := range line {
		normalized, err := normalizePosition(position)
		if err != nil {
			return "", err
		}
		positions = append(positions, normalized)
	}
	return textForPositions(positions), nil
}

func textForPositions(positions [][]float64) string {
	var builder strings.Builder
	for index, position := range positions {
		if index > 0 {
			builder.WriteByte(' ')
		}
		builder.WriteString(strconv.FormatFloat(position[0], 'f', -1, 64))
		builder.WriteByte(',')
		builder.WriteString(strconv.FormatFloat(position[1], 'f', -1, 64))
	}
	return builder.String()
}

type roleGeometry struct {
	role   string
	groups [][]element
}

// lookAtPriority 是取景优先的路径类角色：存在这些图层时只按它们取景，免得被跨半球的大区域拉成全球视角。
var lookAtPriority = map[string]bool{
	"center-line":              true,
	"central-band":             true,
	"central-shadow":           true,
	"central-shadow-sweep":     true,
	"central-shadow-footprint": true,
	"north-limit":              true,
	"south-limit":              true,
	"north-total-limit":        true,
	"south-total-limit":        true,
	"band-outline":             true,
	"total-band-outline":       true,
	"occultation-band":         true,
	"total-footprint":          true,
	"sublunar-track":           true,
}

func lookAtElements(groups []roleGeometry) [][]element {
	preferred := [][]element{}
	for _, item := range groups {
		if lookAtPriority[item.role] {
			preferred = append(preferred, item.groups...)
		}
	}
	if len(preferred) > 0 {
		return preferred
	}
	all := [][]element{}
	for _, item := range groups {
		all = append(all, item.groups...)
	}
	return all
}

func lookAtFor(groups [][]element) *lookAt {
	latitudes := []float64{}
	longitudes := []float64{}
	for _, group := range groups {
		for _, item := range group {
			switch item.kind {
			case "Point":
				if normalized, err := normalizePosition(item.point); err == nil {
					latitudes = append(latitudes, normalized[1])
					longitudes = append(longitudes, normalized[0])
				}
			case "LineString":
				for _, position := range item.line {
					if normalized, err := normalizePosition(position); err == nil {
						latitudes = append(latitudes, normalized[1])
						longitudes = append(longitudes, normalized[0])
					}
				}
			case "Polygon":
				for _, ring := range item.rings {
					for _, position := range ring {
						if normalized, err := normalizePosition(position); err == nil {
							latitudes = append(latitudes, normalized[1])
							longitudes = append(longitudes, normalized[0])
						}
					}
				}
			}
		}
	}
	if len(latitudes) == 0 {
		return nil
	}
	sort.Float64s(latitudes)
	sort.Float64s(longitudes)
	latitude := (latitudes[0] + latitudes[len(latitudes)-1]) / 2
	// 经度用"最大间隙"法取跨反经线也正确的包络。
	span := 0.0
	gapStart := 0.0
	for index := range longitudes {
		next := longitudes[(index+1)%len(longitudes)] + 360*float64((index+1)/len(longitudes))
		gap := next - longitudes[index]
		if gap > span {
			span = gap
			gapStart = longitudes[(index+1)%len(longitudes)]
		}
	}
	longitudeSpan := 360 - span
	longitude := gapStart + longitudeSpan/2
	if longitude > 180 {
		longitude -= 360
	}
	rangeKM := math.Max(latitudes[len(latitudes)-1]-latitudes[0], longitudeSpan*math.Cos(latitude*math.Pi/180)) * 111.32
	// KML 的 LookAt/range 单位是米（不是千米），1.8 倍留边、下限 200 km。
	return &lookAt{Longitude: longitude, Latitude: latitude, Range: math.Max(rangeKM*1.8, 200) * 1000}
}

func resolveStyle(event, role string, polygonal bool, options Options) Style {
	resolved := paletteStyle(event, role)
	if polygonal && options.FillContext && resolved.FillColor == "" {
		resolved.FillColor = grayFill
	}
	for _, key := range []string{event + "/" + role, role} {
		custom, ok := options.Styles[key]
		if !ok {
			continue
		}
		if custom.LineColor != "" {
			resolved.LineColor = custom.LineColor
		}
		if custom.FillColor != "" {
			resolved.FillColor = custom.FillColor
		}
		if custom.LineWidth > 0 {
			resolved.LineWidth = custom.LineWidth
		}
		if custom.NoFill {
			resolved.FillColor = ""
		}
		break
	}
	return resolved
}

func paletteStyle(event, role string) Style {
	switch role {
	case "center-line":
		return Style{LineColor: ColorBlack, LineWidth: 3}
	case "central-band", "total-footprint", "central-shadow-footprint", "central-shadow-sweep", "central-shadow":
		if event == eventSolarEclipse {
			return Style{LineColor: ColorRed, FillColor: redFill, LineWidth: 1.5}
		}
	case "greatest-time-line":
		if event == eventSolarEclipse {
			return Style{LineColor: ColorGreen, LineWidth: 2}
		}
	case "magnitude-line", "magnitude-one-envelope":
		return Style{LineColor: ColorYellow, LineWidth: 2}
	case "total-band", "partial-band", "total-band-outline", "occultation-band":
		if event == eventLunarOccultation {
			return Style{LineColor: ColorYellow, FillColor: yellowFill, LineWidth: 2}
		}
		if role == "partial-band" {
			// 日食的 partial-band 是地方最大食分等于零的可见域边界；它有一大段与日出日落线重合，
			// 同色才不会看起来"一半橙一半黄"。
			return Style{LineColor: ColorOrange, LineWidth: 2}
		}
	case "visibility-boundary", "p1-horizon", "p4-horizon",
		"visible-at-p1", "visible-at-p4",
		"visible-during-eclipse", "visible-throughout-eclipse":
		return Style{LineColor: ColorOrange, LineWidth: 2}
	case "penumbra-moonrise":
		return Style{LineColor: penumbraMoonriseColor, FillColor: penumbraMoonriseFill, LineWidth: 2}
	case "penumbra-moonset":
		return Style{LineColor: penumbraMoonsetColor, FillColor: penumbraMoonsetFill, LineWidth: 2}
	}
	return Style{LineColor: ColorGray, LineWidth: 1}
}

// styleRef 是样式与图层的身份：event 与 role 都可能含 '-'，拼成字符串再当键会让
// (foo-bar, baz) 与 (foo, bar-baz) 撞成同一个图层和样式。
type styleRef struct {
	event string
	role  string
}

// id 返回样式 id 的基础串，合法性与唯一性分别由 sanitizeStyleID 与调用方处理。
func (ref styleRef) id() string {
	if ref.event == "" {
		return "style-" + ref.role
	}
	return "style-" + ref.event + "-" + ref.role
}

// styleLess 按 id 字典序排序；id 相同时再比 event 与 role，保证碰撞时输出仍逐字节稳定。
func styleLess(left, right styleRef) bool {
	if left.id() != right.id() {
		return left.id() < right.id()
	}
	if left.event != right.event {
		return left.event < right.event
	}
	return left.role < right.role
}

// commonProperties 取所有要素共有的属性（名字与编码后的值都相同）——只有它们真正属于集合层，
// 首个要素独有的属性（瞬时足迹的 closure、interp_signature 之类）不能冒充集合属性。
func commonProperties(features []preparedFeature) map[string]string {
	if len(features) == 0 {
		return map[string]string{}
	}
	common := encodeProperties(features[0].props, true)
	for _, feature := range features[1:] {
		entries := encodeProperties(feature.props, true)
		for name, value := range common {
			if other, ok := entries[name]; !ok || other != value {
				delete(common, name)
			}
		}
	}
	// 角色与时间只对单个要素有意义，留在要素层。
	delete(common, "role")
	delete(common, "time")
	delete(common, "times")
	return common
}

// sanitizeStyleID 把逻辑键转成合法的 XML ID（NCName）：非法字符换成 '-'，首字符不能是数字、'.' 或 '-'。
// Google Earth 容忍非法 id，但 KML 2.2 的 id 是 xsd:ID，严格客户端与 XSD 校验会直接拒绝。
func sanitizeStyleID(key string) string {
	var builder strings.Builder
	builder.Grow(len(key) + 2)
	for _, symbol := range key {
		switch {
		case symbol >= 'a' && symbol <= 'z', symbol >= 'A' && symbol <= 'Z',
			symbol >= '0' && symbol <= '9', symbol == '_', symbol == '-', symbol == '.':
			builder.WriteRune(symbol)
		default:
			builder.WriteByte('-')
		}
	}
	id := builder.String()
	if id == "" {
		return "style"
	}
	first := id[0]
	if !(first >= 'a' && first <= 'z' || first >= 'A' && first <= 'Z' || first == '_') {
		id = "style-" + id
	}
	return id
}

// uniformWhen 返回整组几何共用的时刻；时刻不一致的组（times 数组那种）返回空串。
func uniformWhen(group []element) string {
	if len(group) == 0 || group[0].when == "" {
		return ""
	}
	for _, item := range group[1:] {
		if item.when != group[0].when {
			return ""
		}
	}
	return group[0].when
}

func encodeProperties(props map[string]interface{}, includeNested bool) map[string]string {
	entries := map[string]string{}
	for name, value := range props {
		switch typed := value.(type) {
		case nil:
			continue
		case string:
			entries[name] = typed
		case bool:
			entries[name] = strconv.FormatBool(typed)
		case float64:
			entries[name] = strconv.FormatFloat(typed, 'g', -1, 64)
		default:
			if !includeNested {
				continue
			}
			encoded, err := json.Marshal(typed)
			if err != nil {
				continue
			}
			entries[name] = string(encoded)
		}
	}
	return entries
}

func sortedDataEntries(entries map[string]string) []dataEntry {
	names := make([]string, 0, len(entries))
	for name := range entries {
		names = append(names, name)
	}
	sort.Strings(names)
	data := make([]dataEntry, 0, len(names))
	for _, name := range names {
		data = append(data, dataEntry{Name: name, Value: entries[name]})
	}
	return data
}

func documentName(options Options, features []preparedFeature, earliest time.Time) string {
	if options.Name != "" {
		return options.Name
	}
	// 事件类型只在整个集合一致时才代表整份文档；混了多种事件就退回通用名，
	// 免得标题只写了首个要素的事件。
	event := ""
	for _, feature := range features {
		if feature.event == "" {
			continue
		}
		if event == "" {
			event = feature.event
			continue
		}
		if feature.event != event {
			event = ""
			break
		}
	}
	name := localizedEvent(event, options.Language)
	if !earliest.IsZero() {
		name = name + " " + earliest.UTC().Format("2006-01-02")
	}
	return name
}

func descriptionText(options Options, language, timeScale string) string {
	if language == "en" {
		text := "Converted from GeoJSON. Layers are grouped by role; red: solar central band; black: center line; green: solar greatest-time isochrone; yellow: magnitude contours and occultation total/partial bands; orange: rise/set phase lines."
		if timeScale != "" {
			text += " Times were written in UTC (original scale: " + timeScale + ")."
		}
		return text
	}
	text := "由 GeoJSON 转换而来。图层按 role 分组；红：日食中心带；黑：中心线；绿：日食食甚等时线；黄：食分线与掩星全掩/半掩带；橙：日出日落阶段线（初亏/食甚/复圆）。"
	if timeScale != "" {
		text += "时间已按 UTC 写入（原口径：" + timeScale + "）。"
	}
	return text
}

var roleNames = map[string][2]string{
	"central-band":                {"中心带", "Central band"},
	"central-shadow":              {"中心影", "Central shadow"},
	"central-shadow-footprint":    {"中心影轮廓", "Central shadow footprint"},
	"central-shadow-sweep":        {"中心影扫掠", "Central shadow sweep"},
	"central-shadow-boundary":     {"中心影边界", "Central shadow boundary"},
	"center-line":                 {"中心线", "Center line"},
	"north-limit":                 {"北限", "Northern limit"},
	"south-limit":                 {"南限", "Southern limit"},
	"north-total-limit":           {"全带北限", "Northern total limit"},
	"south-total-limit":           {"全带南限", "Southern total limit"},
	"band-outline":                {"带边界", "Band outline"},
	"total-band-outline":          {"全带边界", "Total band outline"},
	"partial-footprint":           {"偏食足迹", "Partial footprint"},
	"total-footprint":             {"全食足迹", "Total footprint"},
	"occultation-band":            {"掩带", "Occultation band"},
	"occultation-footprint":       {"掩星足迹", "Occultation footprint"},
	"visible-footprint-sweep":     {"可见足迹扫掠", "Visible footprint sweep"},
	"greatest":                    {"食甚", "Greatest eclipse"},
	"greatest-time-line":          {"食甚等时线", "Greatest-time isochrone"},
	"magnitude-line":              {"食分线", "Magnitude contour"},
	"magnitude-one-envelope":      {"食分 1 包络", "Magnitude-1 envelope"},
	"visibility-boundary":         {"日出日落阶段线", "Rise/set phase lines"},
	"p1-horizon":                  {"P1 地平线", "P1 horizon"},
	"p4-horizon":                  {"P4 地平线", "P4 horizon"},
	"visible-at-p1":               {"P1 可见点", "Visible at P1"},
	"visible-at-p4":               {"P4 可见点", "Visible at P4"},
	"visible-during-eclipse":      {"食内任意时刻可见区", "Visible at some time during eclipse"},
	"visible-throughout-eclipse":  {"食内全程可见区", "Visible throughout eclipse"},
	"penumbra-moonrise":           {"半影月出", "Penumbra moonrise"},
	"penumbra-moonset":            {"半影月落", "Penumbra moonset"},
	"horizon-connector":           {"地平连接线", "Horizon connector"},
	"target-horizon":              {"目标地平线", "Target horizon"},
	"time-marker":                 {"时间标记", "Time markers"},
	"besselian-critical-envelope": {"贝塞尔临界包络", "Besselian critical envelope"},
	"sublunar-track":              {"月下点轨迹", "Sublunar track"},
}

func localizedRole(event, role, language string) string {
	if language == "en" {
		if name, ok := roleNames[role]; ok {
			return name[1]
		}
		return strings.ReplaceAll(role, "-", " ")
	}
	if name, ok := roleNames[role]; ok {
		return name[0]
	}
	if role == "partial-band" && event == eventLunarOccultation {
		return "半掩带"
	}
	if role == "total-band" && event == eventLunarOccultation {
		return "全掩带"
	}
	if role == "partial-band" {
		return "偏食带"
	}
	return role
}

func localizedEvent(event, language string) string {
	switch event {
	case eventSolarEclipse:
		if language == "en" {
			return "Solar eclipse"
		}
		return "日食"
	case eventLunarEclipse:
		if language == "en" {
			return "Lunar eclipse"
		}
		return "月食"
	case eventLunarOccultation:
		if language == "en" {
			return "Lunar occultation"
		}
		return "月掩"
	}
	if event == "" {
		if language == "en" {
			return "GeoJSON"
		}
		return "GeoJSON"
	}
	return event
}
