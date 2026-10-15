package kml

import (
	"encoding/xml"
	"math"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/starainrt/astro"
	"github.com/starainrt/astro/eclipse"
	"github.com/starainrt/astro/geojson"
	"github.com/starainrt/astro/moon"
)

// 契约：调色板按 (event, role) 派发、role 分层、时间变 TimeStamp、环闭合与朝向、UT1 折回 UTC、LookAt 跨反经线。

func convert(t *testing.T, input string, options Options) kmlRoot {
	t.Helper()
	out, err := FromGeoJSON([]byte(input), options)
	if err != nil {
		t.Fatalf("FromGeoJSON: %v", err)
	}
	if !strings.HasPrefix(string(out), "<?xml") {
		t.Fatalf("输出缺少 XML 声明：%q", string(out[:20]))
	}
	var root kmlRoot
	if err := xml.Unmarshal(out, &root); err != nil {
		t.Fatalf("解析输出的 KML 失败：%v\n%s", err, out)
	}
	return root
}

func styleByID(t *testing.T, root kmlRoot, id string) kmlStyle {
	t.Helper()
	for _, style := range root.Document.Styles {
		if style.ID == id {
			return style
		}
	}
	t.Fatalf("找不到样式 %q", id)
	return kmlStyle{}
}

func placemarksOfRole(root kmlRoot, role string) []placemark {
	var out []placemark
	for _, folder := range root.Document.Folders {
		for _, item := range folder.Placemarks {
			if strings.HasSuffix(item.StyleURL, "-"+role) {
				out = append(out, item)
			}
		}
	}
	return out
}

func feature(event, role, geometry string, properties string) string {
	props := `{"event":"` + event + `","role":"` + role + `"`
	if properties != "" {
		props += "," + properties
	}
	props += "}"
	return `{"type":"Feature","properties":` + props + `,"geometry":` + geometry + `}`
}

func collection(features ...string) string {
	return `{"type":"FeatureCollection","features":[` + strings.Join(features, ",") + `]}`
}

const lineGeometry = `{"type":"LineString","coordinates":[[120,30],[121,31]]}`

func TestPaletteFollowsRoleAndEvent(t *testing.T) {
	input := collection(
		feature("solar-eclipse", "central-band", `{"type":"Polygon","coordinates":[[[100,20],[101,21],[102,20],[100,20]]]}`, ""),
		feature("solar-eclipse", "center-line", lineGeometry, ""),
		feature("solar-eclipse", "greatest-time-line", lineGeometry, ""),
		feature("solar-eclipse", "magnitude-line", lineGeometry, ""),
		feature("solar-eclipse", "visibility-boundary", lineGeometry, `"phase":"start","horizon":"rise"`),
		feature("lunar-occultation", "partial-band", `{"type":"Polygon","coordinates":[[[100,20],[101,21],[102,20],[100,20]]]}`, ""),
		feature("solar-eclipse", "partial-band", `{"type":"Polygon","coordinates":[[[100,20],[101,21],[102,20],[100,20]]]}`, ""),
		feature("lunar-occultation", "total-band", `{"type":"Polygon","coordinates":[[[100,20],[101,21],[102,20],[100,20]]]}`, ""),
		feature("lunar-occultation", "visibility-boundary", lineGeometry, ""),
		feature("lunar-eclipse", "visible-during-eclipse", `{"type":"Polygon","coordinates":[[[100,20],[101,21],[102,20],[100,20]]]}`, ""),
		feature("lunar-eclipse", "visible-throughout-eclipse", `{"type":"Polygon","coordinates":[[[100,20],[101,21],[102,20],[100,20]]]}`, ""),
		feature("lunar-eclipse", "penumbra-moonrise", `{"type":"Polygon","coordinates":[[[100,20],[101,21],[102,20],[100,20]]]}`, ""),
		feature("lunar-eclipse", "penumbra-moonset", `{"type":"Polygon","coordinates":[[[100,20],[101,21],[102,20],[100,20]]]}`, ""),
	)
	root := convert(t, input, Options{})

	cases := []struct {
		id      string
		line    string
		fill    string
		hasFill bool
		polygon bool
	}{
		{"style-solar-eclipse-central-band", ColorRed, redFill, true, true},
		{"style-solar-eclipse-center-line", ColorBlack, "", false, false},
		{"style-solar-eclipse-greatest-time-line", ColorGreen, "", false, false},
		{"style-solar-eclipse-magnitude-line", ColorYellow, "", false, false},
		{"style-solar-eclipse-visibility-boundary", ColorOrange, "", false, false},
		{"style-solar-eclipse-partial-band", ColorOrange, "", false, true},
		{"style-lunar-occultation-partial-band", ColorYellow, yellowFill, true, true},
		{"style-lunar-occultation-total-band", ColorYellow, yellowFill, true, true},
		{"style-lunar-occultation-visibility-boundary", ColorOrange, "", false, false},
		{"style-lunar-eclipse-visible-during-eclipse", ColorOrange, "", false, true},
		{"style-lunar-eclipse-visible-throughout-eclipse", ColorOrange, "", false, true},
		{"style-lunar-eclipse-penumbra-moonrise", penumbraMoonriseColor, penumbraMoonriseFill, true, true},
		{"style-lunar-eclipse-penumbra-moonset", penumbraMoonsetColor, penumbraMoonsetFill, true, true},
	}
	for _, tc := range cases {
		style := styleByID(t, root, tc.id)
		if style.LineStyle == nil || style.LineStyle.Color != tc.line {
			t.Errorf("%s 线色 = %+v, want %s", tc.id, style.LineStyle, tc.line)
		}
		switch {
		case tc.hasFill:
			if style.PolyStyle == nil || style.PolyStyle.Color != tc.fill || style.PolyStyle.Fill != 1 {
				t.Errorf("%s 填充色 = %+v, want %s", tc.id, style.PolyStyle, tc.fill)
			}
		case tc.polygon:
			// 面要素：显式 fill=0，否则客户端会套用默认白填充。
			if style.PolyStyle == nil || style.PolyStyle.Fill != 0 {
				t.Errorf("%s 应有 fill=0 的 PolyStyle，got %+v", tc.id, style.PolyStyle)
			}
		default:
			if style.PolyStyle != nil {
				t.Errorf("%s 不该有 PolyStyle，got %+v", tc.id, style.PolyStyle)
			}
		}
	}
}

// 契约：只描轮廓的面画成折线（不再有面），有填充的面保留 Polygon 且样式必须显式声明 PolyStyle。
func TestOutlineOnlyPolygonBecomesLine(t *testing.T) {
	input := collection(
		feature("solar-eclipse", "partial-band", `{"type":"Polygon","coordinates":[[[100,20],[101,21],[102,20],[100,20]]]}`, ""),
		feature("solar-eclipse", "central-band", `{"type":"Polygon","coordinates":[[[110,20],[111,21],[112,20],[110,20]]]}`, ""),
		feature("solar-eclipse", "center-line", lineGeometry, ""),
	)
	root := convert(t, input, Options{})
	outline := placemarksOfRole(root, "partial-band")
	if len(outline) != 1 || outline[0].Polygon != nil || outline[0].LineString == nil {
		t.Fatalf("中性面应画成折线：%+v", outline)
	}
	band := placemarksOfRole(root, "central-band")
	if len(band) != 1 || band[0].MultiGeometry == nil {
		t.Fatalf("有填充的面应带 Polygon：%+v", band)
	}
	if len(band[0].MultiGeometry.Polygons) != 1 || len(band[0].MultiGeometry.Lines) != 1 {
		t.Fatalf("有填充的面 = %d 面 / %d 线，want 1/1", len(band[0].MultiGeometry.Polygons), len(band[0].MultiGeometry.Lines))
	}
	style := styleByID(t, root, "style-solar-eclipse-central-band")
	if style.PolyStyle == nil || style.PolyStyle.Fill != 1 || style.PolyStyle.Outline != 0 {
		t.Fatalf("有填充的面样式应写 fill=1 且 outline=0（描边交给折线）：%+v", style.PolyStyle)
	}
	if style := styleByID(t, root, "style-solar-eclipse-center-line"); style.PolyStyle != nil {
		t.Fatalf("线图层不该有 PolyStyle：%+v", style.PolyStyle)
	}
}

// 契约：反经线接缝是拆环的产物，不能在图上画成一条沿换日线的直线；不碰接缝的环仍按闭合折线画。
func TestAntimeridianSeamIsNotDrawn(t *testing.T) {
	input := collection(
		feature("solar-eclipse", "partial-band",
			`{"type":"Polygon","coordinates":[[[170,10],[180,10],[180,-10],[170,-10],[170,10]]]}`, ""),
		feature("solar-eclipse", "partial-band",
			`{"type":"Polygon","coordinates":[[[-180,-10],[-170,-10],[-170,10],[-180,10],[-180,-10]]]}`, ""),
	)
	root := convert(t, input, Options{})
	items := placemarksOfRole(root, "partial-band")
	if len(items) != 2 {
		t.Fatalf("应有两个折线 Placemark：%d", len(items))
	}
	for _, item := range items {
		texts := coordinatesOf(item.geometry)
		if len(texts) == 0 {
			t.Fatal("缺少坐标")
		}
		for _, text := range texts {
			previous := math.NaN()
			for _, token := range strings.Fields(text) {
				longitude, err := strconv.ParseFloat(strings.Split(token, ",")[0], 64)
				if err != nil {
					t.Fatalf("解析经度 %q：%v", token, err)
				}
				if !math.IsNaN(previous) && math.Abs(math.Abs(previous)-180) < 1e-9 && math.Abs(math.Abs(longitude)-180) < 1e-9 {
					t.Fatalf("接缝边被画出来了：%s", text)
				}
				previous = longitude
			}
		}
	}
	plain := convert(t, collection(feature("solar-eclipse", "partial-band",
		`{"type":"Polygon","coordinates":[[[100,20],[101,21],[102,20],[100,20]]]}`, "")), Options{})
	item := placemarksOfRole(plain, "partial-band")[0]
	if item.LineString == nil {
		t.Fatalf("普通环应是一条折线：%+v", item.geometry)
	}
	if first := strings.Split(item.LineString.Coordinates, " ")[0]; !strings.HasSuffix(item.LineString.Coordinates, first) {
		t.Fatalf("不碰接缝的环应闭合：%s", item.LineString.Coordinates)
	}
}

func TestStyleOverrideWins(t *testing.T) {
	input := collection(feature("solar-eclipse", "center-line", lineGeometry, ""))
	root := convert(t, input, Options{Styles: map[string]Style{
		"center-line": {LineColor: "ffffffff", LineWidth: 7},
	}})
	style := styleByID(t, root, "style-solar-eclipse-center-line")
	if style.LineStyle.Color != "ffffffff" || style.LineStyle.Width != 7 {
		t.Fatalf("覆盖未生效：%+v", style.LineStyle)
	}
	// "event/role" 比 "role" 更具体。
	root = convert(t, input, Options{Styles: map[string]Style{
		"center-line":               {LineColor: "ffffffff"},
		"solar-eclipse/center-line": {LineColor: "ff112233"},
	}})
	if got := styleByID(t, root, "style-solar-eclipse-center-line").LineStyle.Color; got != "ff112233" {
		t.Fatalf("event/role 覆盖未优先：%s", got)
	}
}

func TestFoldersAndExtendedData(t *testing.T) {
	input := collection(
		feature("solar-eclipse", "center-line", lineGeometry, `"model":"nasa_bulletin_split_k"`),
		feature("solar-eclipse", "north-limit", lineGeometry, `"model":"nasa_bulletin_split_k"`),
	)
	root := convert(t, input, Options{Name: "2024 日全食"})
	if root.Document.Name != "2024 日全食" {
		t.Fatalf("Document 名 = %q", root.Document.Name)
	}
	if len(root.Document.Folders) != 2 {
		t.Fatalf("图层数 = %d, want 2", len(root.Document.Folders))
	}
	if root.Document.Folders[0].Name != "中心线" || root.Document.Folders[1].Name != "北限" {
		t.Fatalf("图层名 = %q / %q", root.Document.Folders[0].Name, root.Document.Folders[1].Name)
	}
	// 集合级 ExtendedData 收下公共属性，要素级不重复。
	if root.Document.ExtendedData == nil || len(root.Document.ExtendedData.Data) == 0 {
		t.Fatal("缺少 Document ExtendedData")
	}
	for _, folder := range root.Document.Folders {
		for _, item := range folder.Placemarks {
			if item.ExtendedData != nil {
				for _, entry := range item.ExtendedData.Data {
					if entry.Name == "model" {
						t.Fatalf("要素级重复了集合级属性 model")
					}
				}
			}
			if item.LineString == nil || item.LineString.Tessellate != 1 {
				t.Fatalf("线要素缺少 LineString/tessellate：%+v", item.geometry)
			}
		}
	}
	// 英文图层名。
	rootEN := convert(t, input, Options{Language: "en"})
	if rootEN.Document.Folders[0].Name != "Center line" {
		t.Fatalf("英文图层名 = %q", rootEN.Document.Folders[0].Name)
	}
}

func TestTimedSegmentsBecomePlacemarksWithTimeStamps(t *testing.T) {
	input := collection(feature("solar-eclipse", "visible-footprint-sweep",
		`{"type":"MultiLineString","coordinates":[[[100,20],[101,20]],[[102,21],[103,21]]]}`,
		`"times":[["2024-04-08T18:00:00Z","2024-04-08T18:05:00Z"],["2024-04-08T18:10:00Z","2024-04-08T18:15:00Z"]]`))
	root := convert(t, input, Options{})
	items := placemarksOfRole(root, "visible-footprint-sweep")
	if len(items) != 2 {
		t.Fatalf("逐段 Placemark 数 = %d, want 2", len(items))
	}
	if items[0].TimeStamp == nil || items[0].TimeStamp.When != "2024-04-08T18:00:00Z" {
		t.Fatalf("首段 TimeStamp = %+v", items[0].TimeStamp)
	}
	if items[1].TimeStamp == nil || items[1].TimeStamp.When != "2024-04-08T18:10:00Z" {
		t.Fatalf("次段 TimeStamp = %+v", items[1].TimeStamp)
	}
	if items[0].Name != "2024-04-08 18:00:00Z" {
		t.Fatalf("段名 = %q", items[0].Name)
	}
	// 单点要素用 time 属性，标签优先作名字。
	input = collection(feature("solar-eclipse", "time-marker",
		`{"type":"Point","coordinates":[121.5,31.2]}`,
		`"time":"2024-04-08T18:17:20Z","label":"02:17"`))
	root = convert(t, input, Options{})
	items = placemarksOfRole(root, "time-marker")
	if len(items) != 1 || items[0].TimeStamp == nil || items[0].TimeStamp.When != "2024-04-08T18:17:20Z" {
		t.Fatalf("时间标记 = %+v", items)
	}
	if items[0].Name != "02:17" {
		t.Fatalf("时间标记名 = %q, want 标签", items[0].Name)
	}
	if items[0].Point == nil || items[0].Point.AltitudeMode != "clampToGround" {
		t.Fatalf("点几何 = %+v", items[0].geometry)
	}
}

func TestPolygonRingsAreClosedAndOriented(t *testing.T) {
	// 外环顺时针且未闭合，洞逆时针，且带一个超出 ±180 的经度。
	input := collection(feature("solar-eclipse", "central-band",
		`{"type":"Polygon","coordinates":[[[100,20],[100,22],[102,22],[102,20]],[[100.5,20.5],[101.5,20.5],[101.5,21.5],[100.5,21.5]]]}`,
		`"lon_extra":185.5`))
	root := convert(t, input, Options{})
	items := placemarksOfRole(root, "central-band")
	if len(items) != 1 || items[0].MultiGeometry == nil || len(items[0].MultiGeometry.Polygons) != 1 {
		t.Fatalf("有填充的面应是 MultiGeometry 里的一个 Polygon：%+v", items)
	}
	value := items[0].MultiGeometry.Polygons[0]
	if value.OuterBoundaryIs == nil || len(value.InnerBoundaries) != 1 {
		t.Fatalf("外环/洞缺失：%+v", value)
	}
	outer := value.OuterBoundaryIs.Ring.Coordinates
	if !strings.HasPrefix(outer, strings.Split(outer, " ")[0]) {
		t.Fatalf("外环未闭合：%s", outer)
	}
	if !ringIsCounterClockwise(t, outer) {
		t.Fatalf("外环应为逆时针：%s", outer)
	}
	if ringIsCounterClockwise(t, value.InnerBoundaries[0].Ring.Coordinates) {
		t.Fatalf("洞应为顺时针：%s", value.InnerBoundaries[0].Ring.Coordinates)
	}
	if strings.Contains(outer, "185.5") || strings.Contains(outer, "-174.5") {
		// 经度回绕只对坐标做；185.5 只是无关属性，仍应出现在 ExtendedData 里。
		t.Fatalf("无关属性不应混进坐标：%s", outer)
	}
}

func ringIsCounterClockwise(t *testing.T, text string) bool {
	t.Helper()
	area := 0.0
	points := [][]float64{}
	for _, tuple := range strings.Fields(text) {
		parts := strings.Split(tuple, ",")
		if len(parts) != 2 {
			t.Fatalf("坐标元组 %q 非法", tuple)
		}
		lon, err := strconv.ParseFloat(parts[0], 64)
		if err != nil {
			t.Fatal(err)
		}
		lat, err := strconv.ParseFloat(parts[1], 64)
		if err != nil {
			t.Fatal(err)
		}
		points = append(points, []float64{lon, lat})
	}
	for index, point := range points {
		next := points[(index+1)%len(points)]
		area += point[0]*next[1] - next[0]*point[1]
	}
	return area > 0
}

// 契约：集合的时间窗口写成 Document/TimeSpan；NoTimes 去掉全部时间元素，得到不受客户端时间轴影响的静态叠加。
func TestDocumentTimeSpanAndNoTimes(t *testing.T) {
	sweep := feature("solar-eclipse", "visible-footprint-sweep",
		`{"type":"MultiLineString","coordinates":[[[100,20],[101,20]],[[102,21],[103,21]]]}`,
		`"times":[["2024-04-08T18:00:00Z","2024-04-08T18:05:00Z"],["2024-04-08T18:10:00Z","2024-04-08T18:15:00Z"]]`)
	marker := feature("solar-eclipse", "time-marker",
		`{"type":"Point","coordinates":[121.5,31.2]}`, `"time":"2024-04-08T18:17:20Z","label":"02:17"`)
	input := collection(sweep, marker)

	root := convert(t, input, Options{})
	if root.Document.TimeSpan == nil {
		t.Fatal("缺少 Document/TimeSpan")
	}
	if root.Document.TimeSpan.Begin != "2024-04-08T18:00:00Z" || root.Document.TimeSpan.End != "2024-04-08T18:17:20Z" {
		t.Fatalf("TimeSpan = %+v", root.Document.TimeSpan)
	}
	if len(placemarksOfRole(root, "visible-footprint-sweep")) != 2 {
		t.Fatal("带时间时应逐段拆 Placemark")
	}

	staticOut, err := FromGeoJSON([]byte(input), Options{NoTimes: true})
	if err != nil {
		t.Fatalf("FromGeoJSON: %v", err)
	}
	if text := string(staticOut); strings.Contains(text, "<when>") || strings.Contains(text, "<TimeStamp>") || strings.Contains(text, "<TimeSpan>") {
		t.Fatalf("NoTimes 仍写出时间元素：%s", text)
	}
	var static kmlRoot
	if err := xml.Unmarshal(staticOut, &static); err != nil {
		t.Fatalf("解析输出的 KML 失败：%v", err)
	}
	if static.Document.TimeSpan != nil {
		t.Fatalf("NoTimes 不该有 TimeSpan：%+v", static.Document.TimeSpan)
	}
	items := placemarksOfRole(static, "visible-footprint-sweep")
	if len(items) != 2 {
		t.Fatalf("NoTimes 仍应保留几何：%d", len(items))
	}
	if items[0].Name != "" {
		t.Fatalf("NoTimes 下没有 label 的段不该再拿时刻当名字：%q", items[0].Name)
	}
	// 集合里没有时间时不写 TimeSpan。
	plain := convert(t, collection(feature("solar-eclipse", "center-line", lineGeometry, "")), Options{})
	if plain.Document.TimeSpan != nil {
		t.Fatalf("无时间集合不该有 TimeSpan：%+v", plain.Document.TimeSpan)
	}
}

func TestUT1LabelsAreConvertedToUTC(t *testing.T) {
	label := time.Date(2024, 4, 8, 18, 17, 20, 0, time.UTC)
	// 2024 年的 DUT1 只有十几毫秒：秒以下精度必须保留，否则这次折算在输出里完全看不见。
	want := astro.UTCFromUT1(label).Format(time.RFC3339Nano)
	if !strings.Contains(want, ".") {
		t.Fatalf("样例应带小数秒以覆盖精度保留：%s", want)
	}
	input := `{"type":"FeatureCollection","time_scale":"UT1","features":[` +
		feature("solar-eclipse", "time-marker", `{"type":"Point","coordinates":[121.5,31.2]}`,
			`"time":"`+label.Format(time.RFC3339)+`"`) + `]}`
	root := convert(t, input, Options{})
	items := placemarksOfRole(root, "time-marker")
	if len(items) != 1 || items[0].TimeStamp == nil || items[0].TimeStamp.When != want {
		t.Fatalf("UT1 折回 UTC 后 = %+v, want %s", items, want)
	}
	if !strings.Contains(root.Document.Description, "UT1") {
		t.Fatalf("说明里应记录原口径：%q", root.Document.Description)
	}
	found := false
	for _, entry := range root.Document.ExtendedData.Data {
		if entry.Name == "time_scale" && entry.Value == "UT1" {
			found = true
		}
	}
	if !found {
		t.Fatal("ExtendedData 缺少 time_scale")
	}
}

// 契约：中性面图层默认只出轮廓，FillContext 才填充，NoFill 压过两者。
func TestFillPolicy(t *testing.T) {
	band := feature("solar-eclipse", "partial-band", `{"type":"Polygon","coordinates":[[[100,20],[101,21],[102,20],[100,20]]]}`, "")
	input := collection(band, feature("solar-eclipse", "central-band", `{"type":"Polygon","coordinates":[[[100,20],[101,21],[102,20],[100,20]]]}`, ""))

	plain := convert(t, input, Options{})
	if got := styleByID(t, plain, "style-solar-eclipse-partial-band").PolyStyle; got == nil || got.Fill != 0 {
		t.Fatalf("中性面默认必须显式 fill=0：%+v", got)
	}
	if got := styleByID(t, plain, "style-solar-eclipse-central-band").PolyStyle; got == nil || got.Color != redFill {
		t.Fatalf("中心带应有红色填充：%+v", got)
	}

	filled := convert(t, input, Options{FillContext: true})
	if got := styleByID(t, filled, "style-solar-eclipse-partial-band").PolyStyle; got == nil || got.Color != grayFill {
		t.Fatalf("FillContext 应给中性面灰色填充：%+v", got)
	}

	custom := convert(t, input, Options{
		FillContext: true,
		Styles: map[string]Style{
			"partial-band": {FillColor: "8000ff00"},
			"central-band": {NoFill: true},
		},
	})
	if got := styleByID(t, custom, "style-solar-eclipse-partial-band").PolyStyle; got == nil || got.Color != "8000ff00" {
		t.Fatalf("自定义填充未生效：%+v", got)
	}
	if got := styleByID(t, custom, "style-solar-eclipse-central-band").PolyStyle; got == nil || got.Fill != 0 {
		t.Fatalf("NoFill 应压成 fill=0：%+v", got)
	}
}

// 契约：SkipRoles 整层丢弃，图层、样式与 LookAt 都不再包含该 role。
func TestSkipRoles(t *testing.T) {
	input := collection(
		feature("solar-eclipse", "center-line", lineGeometry, ""),
		feature("solar-eclipse", "partial-footprint", `{"type":"Polygon","coordinates":[[[100,20],[101,21],[102,20],[100,20]]]}`, ""),
	)
	root := convert(t, input, Options{SkipRoles: []string{"partial-footprint"}})
	for _, folder := range root.Document.Folders {
		if strings.Contains(folder.Name, "偏食足迹") {
			t.Fatalf("被跳过的 role 仍出图层：%q", folder.Name)
		}
	}
	for _, style := range root.Document.Styles {
		if strings.HasSuffix(style.ID, "partial-footprint") {
			t.Fatalf("被跳过的 role 仍出样式：%s", style.ID)
		}
	}
	if got := len(placemarksOfRole(root, "center-line")); got != 1 {
		t.Fatalf("保留图层被误删：%d", got)
	}
}

// 契约：存在中心线等路径图层时，LookAt 只按路径取景，不被跨半球的大区域拉成全球视角。
func TestLookAtPrefersPathRoles(t *testing.T) {
	globe := feature("solar-eclipse", "partial-band",
		`{"type":"Polygon","coordinates":[[[-170,-40],[170,-40],[170,40],[-170,40],[-170,-40]]]}`, "")
	path := feature("solar-eclipse", "center-line", `{"type":"LineString","coordinates":[[120,30],[140,40]]}`, "")
	root := convert(t, collection(globe, path), Options{})
	if root.Document.LookAt == nil {
		t.Fatal("缺少 LookAt")
	}
	if lat := root.Document.LookAt.Latitude; lat < 34 || lat > 36 {
		t.Fatalf("纬度 = %v，应按中心线取景", lat)
	}
	if lon := root.Document.LookAt.Longitude; lon < 129 || lon > 131 {
		t.Fatalf("经度 = %v，应按中心线取景", lon)
	}
	single := feature("solar-eclipse", "magnitude-line", `{"type":"LineString","coordinates":[[100,10],[140,20]]}`, "")
	only := convert(t, collection(single), Options{})
	if only.Document.LookAt == nil || math.Abs(only.Document.LookAt.Longitude-120) > 1 || math.Abs(only.Document.LookAt.Latitude-15) > 1 {
		t.Fatalf("无路径图层时应退回全部要素：%+v", only.Document.LookAt)
	}
}

func TestLookAtHandlesAntimeridian(t *testing.T) {
	input := collection(feature("lunar-occultation", "occultation-band",
		`{"type":"Polygon","coordinates":[[[179,10],[-179,10],[-179,12],[179,12],[179,10]]]}`, ""))
	root := convert(t, input, Options{})
	if root.Document.LookAt == nil {
		t.Fatal("缺少 LookAt")
	}
	if math.Abs(root.Document.LookAt.Longitude) < 179 || math.Abs(root.Document.LookAt.Longitude) > 180 {
		t.Fatalf("跨反经线的中心经度 = %v，应贴近 ±180", root.Document.LookAt.Longitude)
	}
	if root.Document.LookAt.Latitude < 10 || root.Document.LookAt.Latitude > 12 {
		t.Fatalf("中心纬度 = %v", root.Document.LookAt.Latitude)
	}
	plain := convert(t, input, Options{NoLookAt: true})
	if plain.Document.LookAt != nil {
		t.Fatal("NoLookAt 时不应输出 LookAt")
	}
}

func TestFromGeoJSONRejectsBadInput(t *testing.T) {
	cases := []struct {
		name  string
		input string
	}{
		{"非 JSON", `{`},
		{"非 FeatureCollection", `{"type":"Feature","geometry":null,"properties":{}}`},
		{"缺少几何", collection(`{"type":"Feature","properties":{"role":"center-line"}}`)},
		{"未知几何", collection(`{"type":"Feature","properties":{"role":"center-line"},"geometry":{"type":"CircularString","coordinates":[]}}`)},
		{"times 与段数不齐", collection(feature("solar-eclipse", "visible-footprint-sweep",
			`{"type":"MultiLineString","coordinates":[[[100,20],[101,20]],[[102,21],[103,21]]]}`,
			`"times":[["2024-04-08T18:00:00Z","2024-04-08T18:05:00Z"]]`))},
		{"纬度越界", collection(feature("solar-eclipse", "center-line", `{"type":"LineString","coordinates":[[100,95],[101,20]]}`, ""))},
		{"非 UTC 时间", collection(feature("solar-eclipse", "time-marker", `{"type":"Point","coordinates":[100,20]}`, `"time":"not-a-time"`))},
	}
	for _, tc := range cases {
		if _, err := FromGeoJSON([]byte(tc.input), Options{}); err == nil {
			t.Errorf("%s：应当报错", tc.name)
		}
	}
}

func TestGeometryCollectionBecomesMultiGeometry(t *testing.T) {
	input := collection(feature("lunar-occultation", "occultation-band",
		`{"type":"GeometryCollection","geometries":[`+
			`{"type":"MultiPolygon","coordinates":[[[[100,20],[101,20],[101,21],[100,20]]]]},`+
			`{"type":"MultiLineString","coordinates":[[[102,22],[103,22]]]}]}`, ""))
	root := convert(t, input, Options{})
	items := placemarksOfRole(root, "occultation-band")
	if len(items) != 1 || items[0].MultiGeometry == nil {
		t.Fatalf("GeometryCollection 应转成 MultiGeometry：%+v", items)
	}
	multi := items[0].MultiGeometry
	// 面自己另带一条描边折线（面只填充、不描边），加上集合里原有的线。
	if len(multi.Polygons) != 1 || len(multi.Lines) != 2 {
		t.Fatalf("MultiGeometry 内容 = %d 面 / %d 线", len(multi.Polygons), len(multi.Lines))
	}
}

// ringLongitudes 取某个 role 的所有外环经度序列。
func ringLongitudes(t *testing.T, root kmlRoot, role string) [][]float64 {
	t.Helper()
	rings := [][]float64{}
	for _, item := range placemarksOfRole(root, role) {
		polygons := []polygon{}
		if item.Polygon != nil {
			polygons = append(polygons, *item.Polygon)
		}
		if item.MultiGeometry != nil {
			polygons = append(polygons, item.MultiGeometry.Polygons...)
		}
		for _, shape := range polygons {
			if shape.OuterBoundaryIs == nil {
				continue
			}
			ring := []float64{}
			for _, token := range strings.Fields(shape.OuterBoundaryIs.Ring.Coordinates) {
				value, err := strconv.ParseFloat(strings.Split(token, ",")[0], 64)
				if err != nil {
					t.Fatalf("解析经度 %q：%v", token, err)
				}
				ring = append(ring, value)
			}
			rings = append(rings, ring)
		}
	}
	if len(rings) == 0 {
		t.Fatalf("%s 没有外环", role)
	}
	return rings
}

// 契约：本库 geojson 的输出已经按反经线拆好，转出的每条线与每个环都不得出现跨 180° 的经度边。
func TestLibraryOutputHasNoGlobeSpanningEdge(t *testing.T) {
	date := time.Date(2009, time.July, 22, 0, 0, 0, 0, time.UTC)
	partial, ok := eclipse.SolarEclipsePartialFootprints(date, eclipse.SolarEclipsePartialFootprintOptions{
		Step: 20 * time.Minute, BoundaryPoints: 24, GreatestTimeStep: time.Hour,
	})
	if !ok {
		t.Fatal("expected the 2009-07-22 partial phase")
	}
	data, err := geojson.MarshalSolarEclipse(partial, nil)
	if err != nil {
		t.Fatalf("MarshalSolarEclipse: %v", err)
	}
	out, err := FromGeoJSON(data, Options{})
	if err != nil {
		t.Fatalf("转换本库 GeoJSON 失败：%v", err)
	}
	var root kmlRoot
	if err := xml.Unmarshal(out, &root); err != nil {
		t.Fatalf("输出不是合法 KML：%v", err)
	}
	for _, folder := range root.Document.Folders {
		for _, item := range folder.Placemarks {
			for _, coordinates := range coordinatesOf(item.geometry) {
				previous := math.NaN()
				for _, token := range strings.Fields(coordinates) {
					longitude, parseErr := strconv.ParseFloat(strings.Split(token, ",")[0], 64)
					if parseErr != nil {
						t.Fatalf("解析经度 %q：%v", token, parseErr)
					}
					if !math.IsNaN(previous) && math.Abs(longitude-previous) > 180 {
						t.Fatalf("%s 出现跨度 %.1f° 的边（%v → %v）", folder.Name, math.Abs(longitude-previous), previous, longitude)
					}
					previous = longitude
				}
			}
		}
	}
}

func coordinatesOf(shape geometry) []string {
	result := []string{}
	switch {
	case shape.Point != nil:
		result = append(result, shape.Point.Coordinates)
	case shape.LineString != nil:
		result = append(result, shape.LineString.Coordinates)
	case shape.Polygon != nil:
		result = append(result, ringCoordinates(*shape.Polygon)...)
	case shape.MultiGeometry != nil:
		for _, item := range shape.MultiGeometry.Points {
			result = append(result, item.Coordinates)
		}
		for _, item := range shape.MultiGeometry.Lines {
			result = append(result, item.Coordinates)
		}
		for _, item := range shape.MultiGeometry.Polygons {
			result = append(result, ringCoordinates(item)...)
		}
	}
	return result
}

// lineCoordinatesOf 只取折线坐标：面的环允许含 ±180 接缝段（那是填充边界），
// 而面的样式写了 outline=0，客户端不会描它。
func lineCoordinatesOf(shape geometry) []string {
	result := []string{}
	switch {
	case shape.LineString != nil:
		result = append(result, shape.LineString.Coordinates)
	case shape.MultiGeometry != nil:
		for _, item := range shape.MultiGeometry.Lines {
			result = append(result, item.Coordinates)
		}
	}
	return result
}

func ringCoordinates(shape polygon) []string {
	result := []string{}
	if shape.OuterBoundaryIs != nil {
		result = append(result, shape.OuterBoundaryIs.Ring.Coordinates)
	}
	for _, inner := range shape.InnerBoundaries {
		result = append(result, inner.Ring.Coordinates)
	}
	return result
}

func TestFromGeoJSONAcceptsLibraryOutput(t *testing.T) {
	info := eclipse.ClosestLunarEclipse(time.Date(2029, time.January, 1, 0, 0, 0, 0, time.UTC))
	data, err := geojson.MarshalLunarEclipse(info, 48)
	if err != nil {
		t.Fatalf("MarshalLunarEclipse: %v", err)
	}
	out, err := FromGeoJSON(data, Options{})
	if err != nil {
		t.Fatalf("转换本库 GeoJSON 失败：%v", err)
	}
	var root kmlRoot
	if err := xml.Unmarshal(out, &root); err != nil {
		t.Fatalf("输出不是合法 KML：%v", err)
	}
	if len(root.Document.Folders) == 0 {
		t.Fatal("本库月食 GeoJSON 应至少产出一个图层")
	}
	if root.Document.LookAt == nil {
		t.Fatal("应输出 LookAt")
	}
}

// 契约补充（review 修复）：LookAt.range 用米；Style 子元素按 KML sequence；样式 id 合法；
// when 保留小数秒；SkipRoles 裁光全部要素时报错。
func TestLookAtRangeIsMeters(t *testing.T) {
	wide := feature("solar-eclipse", "center-line",
		`{"type":"LineString","coordinates":[[-60,10],[60,20]]}`, "")
	root := convert(t, collection(wide), Options{})
	if root.Document.LookAt == nil {
		t.Fatal("缺少 LookAt")
	}
	// 120° 经度跨度约 1.2e7 m；若仍按千米写入会是 1e4 量级。
	if rangeMeters := root.Document.LookAt.Range; rangeMeters < 1e7 {
		t.Fatalf("LookAt range = %v, 应使用米（≥1e7）", rangeMeters)
	}
	narrow := feature("solar-eclipse", "time-marker",
		`{"type":"Point","coordinates":[121.5,31.2]}`, "")
	small := convert(t, collection(narrow), Options{})
	if small.Document.LookAt == nil || small.Document.LookAt.Range < 2e5 {
		t.Fatalf("小要素的 LookAt range 下限应为 200 km：%+v", small.Document.LookAt)
	}
}

func TestStyleElementsFollowKMLSequence(t *testing.T) {
	input := collection(feature("solar-eclipse", "central-band",
		`{"type":"Polygon","coordinates":[[[100,20],[101,20],[101,21],[100,20]]]}`, ""))
	out, err := FromGeoJSON([]byte(input), Options{})
	if err != nil {
		t.Fatal(err)
	}
	text := string(out)
	icon := strings.Index(text, "<IconStyle>")
	line := strings.Index(text, "<LineStyle>")
	poly := strings.Index(text, "<PolyStyle>")
	if icon < 0 || line < 0 || poly < 0 || !(icon < line && line < poly) {
		t.Fatalf("Style 子元素顺序应为 IconStyle<LineStyle<PolyStyle：%d %d %d", icon, line, poly)
	}
}

func TestStyleIDsAreValidAndUnique(t *testing.T) {
	input := collection(
		feature("", "third-party & <role>", `{"type":"LineString","coordinates":[[0,0],[1,1]]}`, ""),
		feature("", "third-party-_-role-", `{"type":"LineString","coordinates":[[0,0],[1,1]]}`, ""),
	)
	out, err := FromGeoJSON([]byte(input), Options{})
	if err != nil {
		t.Fatal(err)
	}
	var root kmlRoot
	if err := xml.Unmarshal(out, &root); err != nil {
		t.Fatal(err)
	}
	valid := regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_.-]*$`)
	seen := map[string]bool{}
	for _, style := range root.Document.Styles {
		if !valid.MatchString(style.ID) {
			t.Errorf("样式 id 不是合法 NCName：%q", style.ID)
		}
		if seen[style.ID] {
			t.Errorf("样式 id 重复：%q", style.ID)
		}
		seen[style.ID] = true
	}
	if len(seen) != 2 {
		t.Fatalf("样式数 = %d, want 2", len(seen))
	}
	if !strings.Contains(string(out), "#"+firstStyleID(t, root)) {
		t.Fatal("styleUrl 未引用清洗后的 id")
	}
	// 图层名仍应是原始 role，而不是清洗后的 id。
	foundRaw := false
	for _, folder := range root.Document.Folders {
		if folder.Name == "third-party & <role>" {
			foundRaw = true
		}
	}
	if !foundRaw {
		t.Fatalf("图层名未保留原始 role：%+v", root.Document.Folders)
	}
}

func firstStyleID(t *testing.T, root kmlRoot) string {
	t.Helper()
	if len(root.Document.Styles) == 0 {
		t.Fatal("缺少样式")
	}
	return root.Document.Styles[0].ID
}

func TestTimeStampsKeepSubSecondPrecision(t *testing.T) {
	input := collection(feature("solar-eclipse", "time-marker",
		`{"type":"Point","coordinates":[121.5,31.2]}`,
		`"time":"2024-04-08T18:17:20.123456789Z","label":"02:17"`))
	root := convert(t, input, Options{})
	items := placemarksOfRole(root, "time-marker")
	if len(items) != 1 || items[0].TimeStamp == nil {
		t.Fatalf("时间标记 = %+v", items)
	}
	if items[0].TimeStamp.When != "2024-04-08T18:17:20.123456789Z" {
		t.Fatalf("when = %q, 应保留小数秒", items[0].TimeStamp.When)
	}
}

func TestSkipRolesRejectsDroppingEverything(t *testing.T) {
	input := collection(feature("solar-eclipse", "partial-footprint",
		`{"type":"Polygon","coordinates":[[[100,20],[101,21],[102,20],[100,20]]]}`, ""))
	if _, err := FromGeoJSON([]byte(input), Options{SkipRoles: []string{"partial-footprint"}}); err == nil {
		t.Fatal("裁光全部要素时应报错，而不是输出空 Document")
	}
}

func TestMixedEventsPrefixFolderNames(t *testing.T) {
	input := collection(
		feature("solar-eclipse", "center-line", lineGeometry, ""),
		feature("lunar-occultation", "center-line", lineGeometry, ""),
	)
	root := convert(t, input, Options{})
	if len(root.Document.Folders) != 2 {
		t.Fatalf("图层数 = %d, want 2", len(root.Document.Folders))
	}
	for _, folder := range root.Document.Folders {
		if !strings.Contains(folder.Name, " · ") {
			t.Fatalf("多事件集合的图层名应带事件前缀：%q", folder.Name)
		}
	}
	if root.Document.Name != "GeoJSON" {
		t.Fatalf("多事件集合的文档名不该只写首个事件：%q", root.Document.Name)
	}
	single := convert(t, collection(feature("solar-eclipse", "center-line", lineGeometry, "")), Options{})
	if single.Document.Folders[0].Name != "中心线" {
		t.Fatalf("单事件集合不应加前缀：%q", single.Document.Folders[0].Name)
	}
	if single.Document.Name != "日食" {
		t.Fatalf("单事件集合的文档名应按事件类型：%q", single.Document.Name)
	}
}

func TestEmptyCollectionProducesDocumentWithoutLayers(t *testing.T) {
	out, err := FromGeoJSON([]byte(`{"type":"FeatureCollection","features":[]}`), Options{})
	if err != nil {
		t.Fatalf("空集合应合法：%v", err)
	}
	var root kmlRoot
	if err := xml.Unmarshal(out, &root); err != nil {
		t.Fatal(err)
	}
	if len(root.Document.Folders) != 0 || len(root.Document.Styles) != 0 {
		t.Fatalf("空集合不该产生图层或样式：%+v", root.Document)
	}
}

// 契约：折线同样要在反经线接缝处断开。本库 geojson 的 band-outline 是"闭合环线"，
// 环上带 ±180 接缝段；只处理面的轮廓会让这条折线在换日线上画出一条直线。
func TestClosedLineSeamIsNotDrawn(t *testing.T) {
	input := collection(feature("lunar-occultation", "band-outline",
		`{"type":"MultiLineString","coordinates":[[[170,10],[180,10],[180,-10],[170,-10],[170,10]],[[-180,-10],[-170,-10],[-170,10],[-180,10],[-180,-10]]]}`, ""))
	root := convert(t, input, Options{})
	items := placemarksOfRole(root, "band-outline")
	if len(items) != 1 {
		t.Fatalf("Placemark 数 = %d, want 1", len(items))
	}
	texts := coordinatesOf(items[0].geometry)
	// 东侧环的接缝边在环中段，切成 2 段；西侧环的接缝边在收尾处，只切成 1 段：共 3 段。
	if len(texts) != 3 {
		t.Fatalf("接缝切分后应有 3 段折线，得到 %d：%v", len(texts), texts)
	}
	for _, text := range texts {
		previous := math.NaN()
		previousLatitude := math.NaN()
		for _, token := range strings.Fields(text) {
			parts := strings.Split(token, ",")
			longitude, err := strconv.ParseFloat(parts[0], 64)
			if err != nil {
				t.Fatalf("解析经度 %q：%v", token, err)
			}
			latitude, err := strconv.ParseFloat(parts[1], 64)
			if err != nil {
				t.Fatalf("解析纬度 %q：%v", token, err)
			}
			if !math.IsNaN(previous) && math.Abs(math.Abs(previous)-180) < 1e-9 &&
				math.Abs(math.Abs(longitude)-180) < 1e-9 && math.Abs(latitude-previousLatitude) > 1e-12 {
				t.Fatalf("折线上仍残留日界线段：%s", text)
			}
			previous, previousLatitude = longitude, latitude
		}
	}
}

// 端到端：真实星掩的 band-outline 跨反经线，转换后的 KML 不得出现沿 ±180 的直线段。
func TestLibraryOccultationOutputDrawsNoSeamLine(t *testing.T) {
	antares := moon.StarCoordinate{
		ID: "Antares", RA: 247.3516666666667, Dec: -26.431944444444444,
		Epoch:                          time.Date(2000, 1, 1, 12, 0, 0, 0, time.UTC),
		Frame:                          moon.CoordinateFrameJ2000,
		ProperMotionRACosDecMasPerYear: -10,
		ProperMotionDecMasPerYear:      -20,
		ParallaxMas:                    24,
	}
	paths, err := moon.FindStarOccultationPaths(
		time.Date(2026, time.February, 11, 0, 0, 0, 0, time.UTC),
		time.Date(2026, time.February, 12, 0, 0, 0, 0, time.UTC), antares,
		moon.OccultationPathOptions{Step: 10 * time.Minute, TargetSpacingKM: 200, DisableFootprints: true, DisableRiseSet: true})
	if err != nil || len(paths) == 0 {
		t.Fatalf("FindStarOccultationPaths: paths=%d err=%v", len(paths), err)
	}
	data, err := geojson.MarshalStarOccultation(paths[0])
	if err != nil {
		t.Fatalf("MarshalStarOccultation: %v", err)
	}
	out, err := FromGeoJSON(data, Options{})
	if err != nil {
		t.Fatalf("FromGeoJSON: %v", err)
	}
	var root kmlRoot
	if err := xml.Unmarshal(out, &root); err != nil {
		t.Fatalf("输出不是合法 KML：%v", err)
	}
	seams := 0
	for _, folder := range root.Document.Folders {
		for _, item := range folder.Placemarks {
			for _, coordinates := range lineCoordinatesOf(item.geometry) {
				previousLongitude, previousLatitude := math.NaN(), math.NaN()
				for _, token := range strings.Fields(coordinates) {
					parts := strings.Split(token, ",")
					longitude, parseErr := strconv.ParseFloat(parts[0], 64)
					if parseErr != nil {
						t.Fatalf("解析经度 %q：%v", token, parseErr)
					}
					latitude, parseErr := strconv.ParseFloat(parts[1], 64)
					if parseErr != nil {
						t.Fatalf("解析纬度 %q：%v", token, parseErr)
					}
					if !math.IsNaN(previousLongitude) && math.Abs(math.Abs(previousLongitude)-180) < 1e-9 &&
						math.Abs(math.Abs(longitude)-180) < 1e-9 && math.Abs(latitude-previousLatitude) > 1e-12 {
						seams++
					}
					previousLongitude, previousLatitude = longitude, latitude
				}
			}
		}
	}
	if seams != 0 {
		t.Fatalf("KML 里仍画出 %d 条日界线直线段", seams)
	}
}

// 契约：整条都落在 ±180 上的线是真实边界（例如极区被反经线裁出的边），原样保留而不是报错。
func TestLineOnTheSeamIsPreserved(t *testing.T) {
	input := collection(feature("lunar-occultation", "visibility-boundary",
		`{"type":"LineString","coordinates":[[180,-80],[180,-60],[180,-40]]}`, ""))
	root := convert(t, input, Options{})
	items := placemarksOfRole(root, "visibility-boundary")
	if len(items) != 1 {
		t.Fatalf("Placemark 数 = %d, want 1", len(items))
	}
	texts := coordinatesOf(items[0].geometry)
	if len(texts) != 1 {
		t.Fatalf("接缝上的线应原样保留 1 段，得到 %d", len(texts))
	}
	if !strings.Contains(texts[0], "180,-80") || !strings.Contains(texts[0], "180,-40") {
		t.Fatalf("坐标未保留：%s", texts[0])
	}
}

// 契约补充（review 修复）：LookAt 子元素按 KML 2.2 sequence；单点路径片段跳过而不是让整份转换失败；
// 多几何要素的标量 time 也变成 TimeStamp；环上除接缝外的真实边一条不少；任意 event 都支持样式覆盖。

func splitPositions(text string) [][2]float64 {
	result := [][2]float64{}
	for _, chunk := range strings.Fields(text) {
		parts := strings.Split(chunk, ",")
		if len(parts) < 2 {
			continue
		}
		longitude, firstErr := strconv.ParseFloat(parts[0], 64)
		latitude, secondErr := strconv.ParseFloat(parts[1], 64)
		if firstErr != nil || secondErr != nil {
			continue
		}
		result = append(result, [2]float64{longitude, latitude})
	}
	return result
}

func TestLookAtElementsFollowKMLSequence(t *testing.T) {
	out, err := FromGeoJSON([]byte(collection(feature("solar-eclipse", "center-line", lineGeometry, ""))), Options{})
	if err != nil {
		t.Fatal(err)
	}
	text := string(out)
	start := strings.Index(text, "<LookAt>")
	end := strings.Index(text, "</LookAt>")
	if start < 0 || end < start {
		t.Fatalf("缺少 LookAt：%s", text)
	}
	block := text[start:end]
	order := []int{}
	for _, name := range []string{"longitude", "latitude", "heading", "tilt", "range"} {
		order = append(order, strings.Index(block, "<"+name+">"))
	}
	for index, position := range order {
		if position < 0 {
			t.Fatalf("LookAt 缺少第 %d 个子元素：%s", index, block)
		}
		if index > 0 && position < order[index-1] {
			t.Fatalf("LookAt 子元素顺序应为 longitude<latitude<heading<tilt<range（KML 2.2 sequence）：%s", block)
		}
	}
}

func TestDegenerateLineFragmentIsSkipped(t *testing.T) {
	degenerate := feature("solar-eclipse", "horizon-connector",
		`{"type":"LineString","coordinates":[[120,30],[120,30]]}`, "")
	valid := feature("solar-eclipse", "center-line", lineGeometry, "")
	root := convert(t, collection(degenerate, valid), Options{})
	if len(placemarksOfRole(root, "horizon-connector")) != 0 {
		t.Fatal("两个相同坐标的单点片段不应画出折线")
	}
	if len(placemarksOfRole(root, "center-line")) != 1 {
		t.Fatal("同一集合里的有效折线应保留")
	}
	if len(root.Document.Folders) != 1 {
		t.Fatalf("退化的要素不应留下空图层：%+v", root.Document.Folders)
	}
	if _, err := FromGeoJSON([]byte(collection(degenerate)), Options{}); err == nil ||
		!strings.Contains(err.Error(), "drawable geometry") {
		t.Fatalf("全部几何都不可画时应报错，实际：%v", err)
	}
}

func TestMultiGeometryScalarTimeBecomesTimeStamp(t *testing.T) {
	input := collection(feature("solar-eclipse", "time-marker",
		`{"type":"GeometryCollection","geometries":[{"type":"Point","coordinates":[1,2]},{"type":"Point","coordinates":[3,4]}]}`,
		`"time":"2024-04-08T18:00:00Z"`))
	root := convert(t, input, Options{})
	placemarks := placemarksOfRole(root, "time-marker")
	if len(placemarks) != 1 {
		t.Fatalf("多几何要素应合成一个 Placemark：%d", len(placemarks))
	}
	if placemarks[0].TimeStamp == nil || placemarks[0].TimeStamp.When != "2024-04-08T18:00:00Z" {
		t.Fatalf("多几何要素丢失 TimeStamp：%+v", placemarks[0])
	}
	if placemarks[0].ExtendedData != nil {
		for _, data := range placemarks[0].ExtendedData.Data {
			if data.Name == "time" {
				t.Fatal("time 已变成 TimeStamp，不应再留在 ExtendedData")
			}
		}
	}
}

func TestSeamCutRingKeepsRealClosingEdge(t *testing.T) {
	input := collection(feature("solar-eclipse", "band-outline",
		`{"type":"Polygon","coordinates":[[[170,10],[180,10],[180,-10],[170,-10],[170,10]]]}`, ""))
	root := convert(t, input, Options{})
	placemarks := placemarksOfRole(root, "band-outline")
	if len(placemarks) != 1 {
		t.Fatalf("带边界应是一个 Placemark：%d", len(placemarks))
	}
	lines := lineCoordinatesOf(placemarks[0].geometry)
	if len(lines) != 1 {
		t.Fatalf("环上只有一条接缝边，应合成一条折线：%v", lines)
	}
	points := splitPositions(lines[0])
	if len(points) != 4 {
		t.Fatalf("三条真实边应有四个顶点：%v", lines)
	}
	meridian, seam := 0, 0
	for index := 0; index+1 < len(points); index++ {
		from, to := points[index], points[index+1]
		if math.Abs(from[0]-170) < 1e-9 && math.Abs(to[0]-170) < 1e-9 {
			meridian++
		}
		if math.Abs(math.Abs(from[0])-180) < 1e-9 && math.Abs(math.Abs(to[0])-180) < 1e-9 {
			seam++
		}
	}
	if meridian != 1 {
		t.Fatalf("经度 170° 的真实闭合边被漏掉：%v", lines)
	}
	if seam != 0 {
		t.Fatalf("折线里不应画出接缝：%v", lines)
	}
}

func TestCustomEventStyleOverride(t *testing.T) {
	input := collection(feature("custom-event", "route", lineGeometry, ""))
	for _, key := range []string{"route", "custom-event/route"} {
		root := convert(t, input, Options{Styles: map[string]Style{key: {LineColor: "ff00ff00", LineWidth: 7}}})
		style := styleByID(t, root, firstStyleID(t, root))
		if style.LineStyle == nil || style.LineStyle.Color != "ff00ff00" || style.LineStyle.Width != 7 {
			t.Fatalf("样式键 %q 未生效：%+v", key, style.LineStyle)
		}
		if names := folderNames(root); len(names) != 1 || names[0] != "route" {
			t.Fatalf("单个事件时图层名应只用 role：%v", names)
		}
	}
}

func folderNames(root kmlRoot) []string {
	names := []string{}
	for _, folder := range root.Document.Folders {
		names = append(names, folder.Name)
	}
	return names
}

func TestStyleKeysDoNotCollideAcrossEventAndRole(t *testing.T) {
	// (event=foo-bar, role=baz) 与 (event=foo, role=bar-baz) 拼出的 id 相同，
	// 但它们是两个图层、两个样式，不能并成一个。
	input := collection(
		feature("foo-bar", "baz", lineGeometry, ""),
		feature("foo", "bar-baz", `{"type":"LineString","coordinates":[[5,6],[7,8]]}`, ""),
	)
	root := convert(t, input, Options{})
	if len(root.Document.Folders) != 2 {
		t.Fatalf("两个 (event, role) 应各成一个图层：%+v", folderNames(root))
	}
	styles := map[string]bool{}
	for _, style := range root.Document.Styles {
		if styles[style.ID] {
			t.Fatalf("样式 id 重复：%q", style.ID)
		}
		styles[style.ID] = true
	}
	if len(styles) != 2 {
		t.Fatalf("两个 (event, role) 应各有自己的样式：%d", len(styles))
	}
	for _, folder := range root.Document.Folders {
		if len(folder.Placemarks) != 1 {
			t.Fatalf("图层 %q 的 Placemark 数应为 1：%d", folder.Name, len(folder.Placemarks))
		}
		if !styles[strings.TrimPrefix(folder.Placemarks[0].StyleURL, "#")] {
			t.Fatalf("图层 %q 引用了不存在的样式 %q", folder.Name, folder.Placemarks[0].StyleURL)
		}
	}
}

func TestDocumentExtendedDataIsCommonIntersection(t *testing.T) {
	first := feature("solar-eclipse", "central-shadow-footprint",
		`{"type":"Polygon","coordinates":[[[100,20],[101,20],[101,21],[100,20]]]}`,
		`"model":"nasa_bulletin_split_k","eclipse_type":"total","central_duration_seconds":123.4,"closure":{"kind":"horizon","exact":true}`)
	second := feature("solar-eclipse", "center-line", lineGeometry, `"model":"nasa_bulletin_split_k","eclipse_type":"total"`)
	root := convert(t, collection(first, second), Options{})
	shared := map[string]string{}
	if root.Document.ExtendedData != nil {
		for _, data := range root.Document.ExtendedData.Data {
			shared[data.Name] = data.Value
		}
	}
	for _, name := range []string{"model", "eclipse_type", "event"} {
		if _, ok := shared[name]; !ok {
			t.Fatalf("全体要素共有的属性 %q 应写进 Document：%v", name, shared)
		}
	}
	for _, name := range []string{"central_duration_seconds", "closure"} {
		if _, ok := shared[name]; ok {
			t.Fatalf("只在部分要素上出现的属性 %q 不该冒充集合属性", name)
		}
	}
	placemarks := placemarksOfRole(root, "central-shadow-footprint")
	if len(placemarks) != 1 || placemarks[0].ExtendedData == nil {
		t.Fatalf("足迹要素应带自己的 ExtendedData：%+v", placemarks)
	}
	local := map[string]string{}
	for _, data := range placemarks[0].ExtendedData.Data {
		local[data.Name] = data.Value
	}
	if local["central_duration_seconds"] != "123.4" {
		t.Fatalf("要素独有属性应留在 Placemark：%v", local)
	}
	if !strings.Contains(local["closure"], `"kind":"horizon"`) {
		t.Fatalf("嵌套属性应写进 Placemark：%v", local)
	}
	for _, name := range []string{"model", "eclipse_type"} {
		if _, ok := local[name]; ok {
			t.Fatalf("已提升到集合层的属性 %q 不该在 Placemark 里重复", name)
		}
	}
}

// 仅见半影的两条带从中英图层名到配色都要能对号：中文「半影月出/半影月落」，英文 "Penumbra moonrise/moonset"。
func TestLunarPenumbraBandFeaturesCarryNames(t *testing.T) {
	date := time.Date(2029, 1, 1, 0, 0, 0, 0, time.FixedZone("CST", 8*3600))
	info, ok := eclipse.LunarEclipseOnDate(date)
	if !ok {
		t.Fatal("missing lunar eclipse")
	}
	data, err := geojson.MarshalLunarEclipse(info, 360)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		language string
		names    []string
	}{
		{"", []string{"半影月出", "半影月落"}},
		{"en", []string{"Penumbra moonrise", "Penumbra moonset"}},
	} {
		out, convertErr := FromGeoJSON(data, Options{Language: tc.language})
		if convertErr != nil {
			t.Fatalf("FromGeoJSON(%q): %v", tc.language, convertErr)
		}
		text := string(out)
		for _, name := range tc.names {
			if !strings.Contains(text, "<name>"+name+"</name>") {
				t.Fatalf("language %q: KML 缺少图层名 %q", tc.language, name)
			}
		}
	}
}

// 契约：第三方输入里真正跨过反经线的面在 ±180 处切成两块，填充环里不得出现跨 180° 的边。
func TestCrossingPolygonIsSplitAtAntimeridian(t *testing.T) {
	inputs := []string{
		`{"type":"Polygon","coordinates":[[[179,10],[-179,10],[-179,12],[179,12],[179,10]]]}`,
		`{"type":"Polygon","coordinates":[[[-179,10],[179,10],[179,12],[-179,12],[-179,10]]]}`,
		`{"type":"MultiPolygon","coordinates":[[[[170,0],[-170,0],[-170,20],[170,20],[170,0]],[[175,5],[175,15],[-175,15],[-175,5],[175,5]]]]}`,
	}
	for _, input := range inputs {
		root := convert(t, collection(feature("solar-eclipse", "central-band", input, "")), Options{FillContext: true})
		polygons := 0
		for _, item := range placemarksOfRole(root, "central-band") {
			for _, coordinates := range coordinatesOf(item.geometry) {
				positions := splitPositions(coordinates)
				for index := 1; index < len(positions); index++ {
					if math.Abs(positions[index][0]-positions[index-1][0]) > 180 {
						t.Fatalf("%s: edge %v → %v spans the globe", input, positions[index-1], positions[index])
					}
				}
			}
			if item.MultiGeometry != nil {
				polygons += len(item.MultiGeometry.Polygons)
			}
		}
		if polygons != 2 {
			t.Fatalf("%s: %d polygons, want 2", input, polygons)
		}
	}
}
