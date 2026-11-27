package svg

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"io"
	"math"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/starainrt/astro/basic"
	"github.com/starainrt/astro/internal/ephem"
	"github.com/starainrt/astro/internal/svgchart"
)

func systemFixtureDate() time.Time {
	return time.Date(2026, 5, 14, 13, 38, 0, 0, time.UTC)
}

func svgElements(document string) ([]xml.StartElement, []string, error) {
	decoder := xml.NewDecoder(strings.NewReader(document))
	var elements []xml.StartElement
	var texts []string
	depth, roots := 0, 0
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			if depth != 0 || roots != 1 {
				return nil, nil, fmt.Errorf("root/depth: %d/%d", roots, depth)
			}
			return elements, texts, nil
		}
		if err != nil {
			return nil, nil, err
		}
		switch value := token.(type) {
		case xml.StartElement:
			if depth == 0 {
				roots++
				if value.Name.Local != "svg" || value.Name.Space != "http://www.w3.org/2000/svg" {
					return nil, nil, fmt.Errorf("not an SVG root")
				}
			}
			depth++
			if value.Name.Local == "text" {
				var text string
				if err := decoder.DecodeElement(&text, &value); err != nil {
					return nil, nil, err
				}
				elements = append(elements, value)
				texts = append(texts, text)
				depth--
			} else {
				elements = append(elements, value)
				texts = append(texts, "")
			}
		case xml.EndElement:
			depth--
		case xml.CharData:
			if depth == 0 && strings.TrimSpace(string(value)) != "" {
				return nil, nil, fmt.Errorf("text outside root")
			}
		}
	}
}

func svgAttribute(element xml.StartElement, name string) string {
	for _, attribute := range element.Attr {
		if attribute.Name.Local == name {
			return attribute.Value
		}
	}
	return ""
}

func svgNumber(element xml.StartElement, name string) float64 {
	value, err := strconv.ParseFloat(svgAttribute(element, name), 64)
	if err != nil {
		return math.NaN()
	}
	return value
}

func svgTextBounds(document string) error {
	elements, texts, err := svgElements(document)
	if err != nil {
		return err
	}
	var left, top, width, height float64
	if _, err := fmt.Sscan(svgAttribute(elements[0], "viewBox"), &left, &top, &width, &height); err != nil {
		return err
	}
	for index, element := range elements {
		if element.Name.Local != "text" {
			continue
		}
		x, baseline, size := svgNumber(element, "x"), svgNumber(element, "y"), svgNumber(element, "font-size")
		if math.IsNaN(x+baseline+size) || size <= 0 || svgAttribute(element, "transform") != "" {
			return fmt.Errorf("unmeasured text: %q", texts[index])
		}
		box := svgchart.LabelTextBox(x, baseline, size, texts[index], svgAttribute(element, "text-anchor"))
		if box.X < left || box.Y < top || box.X+box.Width > left+width || box.Y+box.Height > top+height {
			return fmt.Errorf("out-of-canvas text %q: %+v", texts[index], box)
		}
	}
	return nil
}

func svgFinite(document string) bool {
	return !strings.Contains(document, "NaN") && !strings.Contains(document, "Inf")
}

func systemDocument(t *testing.T, options GalileanSystemSVGOptions) string {
	t.Helper()
	document, ok := GalileanSystemSVG(systemFixtureDate(), options)
	if !ok {
		t.Fatal("missing system chart")
	}
	return document
}

func TestGalileanSystemSVGSyntheticXML(t *testing.T) {
	document := systemDocument(t, GalileanSystemSVGOptions{Title: "A & B <星>"})
	if _, _, err := svgElements(document); err != nil {
		t.Fatal(err)
	}
	if _, _, err := svgElements(strings.TrimSuffix(document, "</svg>")); err == nil {
		t.Fatal("XML validator accepts an unclosed root")
	}
}

func TestGalileanSystemSVGSyntheticFinite(t *testing.T) {
	if !svgFinite(systemDocument(t, GalileanSystemSVGOptions{})) {
		t.Fatal("nonfinite output")
	}
	for _, invalid := range []string{"NaN", "Inf", "+Inf", "-Inf"} {
		if svgFinite("<svg>" + invalid + "</svg>") {
			t.Fatal("nonfinite validator misses " + invalid)
		}
	}
}

func TestGalileanSystemSVGSyntheticTextBounds(t *testing.T) {
	for _, options := range []GalileanSystemSVGOptions{{}, {Language: "en"}, {Width: 1200, Height: 800, Title: strings.Repeat("很长的标题 long title ", 30)}} {
		if err := svgTextBounds(systemDocument(t, options)); err != nil {
			t.Fatal(err)
		}
	}
	invalid := `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 900 600"><text x="899" y="30" font-size="20">outside</text></svg>`
	if svgTextBounds(invalid) == nil {
		t.Fatal("bounds validator accepts overflowing text")
	}
}

func TestGalileanSystemSVGSyntheticDefaults(t *testing.T) {
	got := systemDocument(t, GalileanSystemSVGOptions{})
	want := systemDocument(t, GalileanSystemSVGOptions{Width: 900, Height: 600, Language: "zh", Location: time.FixedZone("CST", 8*3600)})
	if got != want || got != systemDocument(t, GalileanSystemSVGOptions{Language: "EN"}) || !strings.Contains(got, "2026-05-14 21:38:00 UTC+8") || !strings.Contains(got, "木星伽利略卫星位置图") {
		t.Fatal("default dimensions/language/location differ")
	}
	english := systemDocument(t, GalileanSystemSVGOptions{Language: "en", Location: time.UTC})
	if !strings.Contains(english, "Galilean Satellites of Jupiter") || !strings.Contains(english, "2026-05-14 13:38:00 UTC+0") {
		t.Fatal("explicit English/UTC ignored")
	}
	if !strings.Contains(systemDocument(t, GalileanSystemSVGOptions{Location: time.FixedZone("x", -(3*3600 + 30*60))}), "UTC-3:30") {
		t.Fatal("zone label does not follow the offset")
	}
	if !strings.Contains(got, `fill="#ffffff"`) || strings.Contains(got, "#101b2b") {
		t.Fatal("chart is not on the white card style")
	}
	for _, value := range []interface{}{GalileanSystemSVGOptions{}, GalileanMutualEventSVGOptions{}} {
		typeInfo := reflect.TypeOf(value)
		if typeInfo.NumField() != 6 {
			t.Fatal("options exceed six fields")
		}
		for index, name := range []string{"Width", "Height", "Title", "Language", "Location", "TrueScale"} {
			if typeInfo.Field(index).Name != name {
				t.Fatal("option outside whitelist")
			}
		}
	}
}

func systemCircles(t *testing.T, document string) (map[string]xml.StartElement, []string) {
	t.Helper()
	elements, _, err := svgElements(document)
	if err != nil {
		t.Fatal(err)
	}
	byID := make(map[string]xml.StartElement)
	var order []string
	for _, element := range elements {
		if id := svgAttribute(element, "id"); id != "" {
			byID[id] = element
			order = append(order, id)
		}
	}
	return byID, order
}

func TestGalileanSystemSVGConsistencyTrueScale(t *testing.T) {
	jde := basic.UTC2TT(basic.Date2JD(systemFixtureDate()))
	snapshot, _ := basic.JupiterGalileanDiagramWithInstance(jde, nil)
	observation := basic.JupiterGalileanSatelliteObservation(jde, 1)
	semidiameterArcsec := observation.OffsetXArcsec / observation.OffsetXJupiterRadii
	satelliteArcsec := math.Atan2(1829.4, observation.Distance*149597870.7) * 180 / math.Pi * 3600
	physical := satelliteArcsec / semidiameterArcsec
	for _, trueScale := range []bool{true, false} {
		byID, _ := systemCircles(t, systemDocument(t, GalileanSystemSVGOptions{TrueScale: trueScale}))
		jupiterRadius, satelliteRadius := svgNumber(byID["jupiter-disk"], "rx"), svgNumber(byID["satellite-1"], "r")
		ratio := satelliteRadius / jupiterRadius
		if jupiterRadius <= 0 || satelliteRadius <= 0 || (trueScale && math.Abs(ratio-physical) > 1e-3*physical) || (!trueScale && ratio <= 2*physical) {
			t.Fatalf("trueScale=%v: radius ratio %.9f, independent physical %.9f (snapshot %.9f)", trueScale, ratio, physical, snapshot.AngularRadiiJupiterR[0])
		}
	}
}

func TestGalileanSystemSVGConsistencyGeometry(t *testing.T) {
	jde := basic.UTC2TT(basic.Date2JD(systemFixtureDate()))
	snapshot, ok := basic.JupiterGalileanDiagramWithInstance(jde, nil)
	if !ok {
		t.Fatal("snapshot missing")
	}
	byID, order := systemCircles(t, systemDocument(t, GalileanSystemSVGOptions{}))
	disk := byID["jupiter-disk"]
	scale := svgNumber(disk, "rx")
	if math.Abs(svgNumber(disk, "ry")/scale-snapshot.MinorRadiusRatio) > 1e-3 {
		t.Fatal("projected Jovian flattening differs")
	}
	position := func(id string) int {
		for index, value := range order {
			if value == id {
				return index
			}
		}
		return -1
	}
	centerX, centerY := svgNumber(disk, "cx"), svgNumber(disk, "cy")
	for satellite := 0; satellite < 4; satellite++ {
		id := fmt.Sprintf("satellite-%d", satellite+1)
		element := byID[id]
		observation := basic.JupiterGalileanSatelliteObservation(jde, satellite+1)
		if math.Abs((centerX-svgNumber(element, "cx"))/scale-observation.OffsetXJupiterRadii) > 2e-3 || math.Abs((centerY-svgNumber(element, "cy"))/scale-observation.OffsetYJupiterRadii) > 2e-3 {
			t.Fatalf("%s is not at the east-left / north-up offset", id)
		}
		if (position(id) > position("jupiter-disk")) != snapshot.Observations[satellite].InFrontOfJupiter {
			t.Fatalf("%s has wrong foreground/background order", id)
		}
	}
	document := systemDocument(t, GalileanSystemSVGOptions{})
	if !strings.Contains(document, "被木卫一食") {
		t.Fatal("mutual eclipse state missing from the satellite table")
	}
}

func TestGalileanSystemSVGMutualOccultationPaintOrder(t *testing.T) {
	when := time.Date(2026, 6, 4, 13, 40, 35, 851389169, time.UTC)
	for _, trueScale := range []bool{false, true} {
		document, ok := GalileanSystemSVG(when, GalileanSystemSVGOptions{TrueScale: trueScale})
		if !ok {
			t.Fatal("mutual occultation diagram missing")
		}
		byID, _ := systemCircles(t, document)
		io, europa := byID["satellite-1"], byID["satellite-2"]
		distance := math.Hypot(svgNumber(io, "cx")-svgNumber(europa, "cx"), svgNumber(io, "cy")-svgNumber(europa, "cy"))
		if distance >= svgNumber(io, "r")+svgNumber(europa, "r") {
			t.Fatal("mutual occultation fixture does not overlap")
		}
		if strings.Index(document, `id="satellite-2"`) > strings.Index(document, `id="satellite-1"`) {
			t.Errorf("TrueScale=%t: Europa paints over foreground Io", trueScale)
		}
	}
}

func TestGalileanSystemSVGSatelliteOverlapConsistency(t *testing.T) {
	for _, when := range []time.Time{
		time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 6, 4, 13, 40, 35, 851389169, time.UTC),
	} {
		document, ok := GalileanSystemSVG(when, GalileanSystemSVGOptions{})
		physical, scaleOK := GalileanSystemSVG(when, GalileanSystemSVGOptions{TrueScale: true})
		if !ok || !scaleOK {
			t.Fatal("missing diagram", when)
		}
		drawn, _ := systemCircles(t, document)
		reference, _ := systemCircles(t, physical)
		for first := 1; first <= 4; first++ {
			for second := first + 1; second <= 4; second++ {
				firstID, secondID := fmt.Sprintf("satellite-%d", first), fmt.Sprintf("satellite-%d", second)
				left, right := drawn[firstID], drawn[secondID]
				separation := math.Hypot(svgNumber(left, "cx")-svgNumber(right, "cx"), svgNumber(left, "cy")-svgNumber(right, "cy"))
				physicalSum := svgNumber(reference[firstID], "r") + svgNumber(reference[secondID], "r")
				paintedSum := svgNumber(left, "r") + svgNumber(right, "r") + (svgNumber(left, "stroke-width")+svgNumber(right, "stroke-width"))/2
				if math.IsNaN(separation+paintedSum+physicalSum) || (separation < paintedSum) != (separation < physicalSum) {
					t.Errorf("%v %s/%s: separation %.3f, painted radii %.3f, physical radii %.3f", when, firstID, secondID, separation, paintedSum, physicalSum)
				}
			}
		}
	}
}

func TestGalileanSystemSVGJupiterLimbConsistency(t *testing.T) {
	for _, fixture := range []struct {
		when       time.Time
		id         string
		foreground bool
	}{
		{time.Date(2026, 3, 18, 0, 0, 0, 0, time.UTC), "satellite-4", true},
		{time.Date(2026, 5, 18, 0, 0, 0, 0, time.UTC), "satellite-1", false},
	} {
		physical, ok := GalileanSystemSVG(fixture.when, GalileanSystemSVGOptions{TrueScale: true})
		if !ok {
			t.Fatal("missing reference diagram", fixture.when)
		}
		reference, _ := systemCircles(t, physical)
		for _, trueScale := range []bool{false, true} {
			document, ok := GalileanSystemSVG(fixture.when, GalileanSystemSVGOptions{TrueScale: trueScale})
			if !ok {
				t.Fatal("missing diagram", fixture.when)
			}
			byID, _ := systemCircles(t, document)
			limb, moon := byID["jupiter-disk"], byID[fixture.id]
			distance := math.Hypot(svgNumber(moon, "cx")-svgNumber(limb, "cx"), svgNumber(moon, "cy")-svgNumber(limb, "cy"))
			physicalRadius := svgNumber(reference[fixture.id], "r")
			if !(distance > svgNumber(limb, "rx")+physicalRadius && distance < svgNumber(limb, "ry")+4.5) {
				t.Fatalf("%v: fixture must be clear at true scale and touch at 4.5 chart units", fixture.when)
			}
			if svgNumber(moon, "r") != physicalRadius || svgNumber(moon, "stroke-width") != 0 || svgAttribute(limb, "clip-path") != "url(#jovian-limb)" {
				t.Errorf("%v TrueScale=%t: limb-contact geometry includes enlarged radii or outlines", fixture.when, trueScale)
			}
			foreground := strings.Index(document, `id="`+fixture.id+`"`) > strings.Index(document, `id="jupiter-disk"`)
			if foreground != fixture.foreground {
				t.Errorf("%v TrueScale=%t: incorrect paint order", fixture.when, trueScale)
			}
		}
	}
}

func TestGalileanSystemSVGConsistencyBaseline(t *testing.T) {
	if go116386() {
		t.Skip("Go 1.16/386 floating-point text differs; render-only test covers this build")
	}
	for _, fixture := range []struct {
		language string
		digest   [2]string
		length   [2]int
	}{
		{"zh", [2]string{"a37e9ee433bee1500018128751ccc3c1", "651f4a2a542c0b0052a7ccf671e042c9"}, [2]int{10444, 10444}},
		{"en", [2]string{"5cb6cdb5f14742d89affb5cb034aa45b", "29e476d7c19e9357b9102f8800d37338"}, [2]int{10347, 10347}},
	} {
		for _, model := range []basic.EphemerisModel{basic.EphemerisLegacy, basic.EphemerisSemianalytic} {
			previous := basic.GetEphemerisModel()
			basic.SetEphemerisModel(model)
			document, ok := GalileanSystemSVG(systemFixtureDate(), GalileanSystemSVGOptions{Language: fixture.language})
			basic.SetEphemerisModel(previous)
			sum := md5.Sum([]byte(document))
			digest := hex.EncodeToString(sum[:])
			if !ok || digest != fixture.digest[model] || len(document) != fixture.length[model] {
				t.Errorf("%s model=%d: md5=%s bytes=%d", fixture.language, model, digest, len(document))
			}
		}
	}
}

func TestGalileanSystemSVGSyntheticRendersOnGo116386(t *testing.T) {
	if !go116386() {
		t.Skip("Go 1.16/386 only")
	}
	for _, language := range []string{"zh", "en"} {
		document := systemDocument(t, GalileanSystemSVGOptions{Language: language})
		if _, _, err := svgElements(document); err != nil {
			t.Fatal(err)
		}
	}
}

func TestGalileanSystemSVGConsistencyFiniteModels(t *testing.T) {
	previous := ephem.Active()
	defer basic.SetEphemerisModel(basic.EphemerisModel(previous))
	for _, model := range []basic.EphemerisModel{basic.EphemerisLegacy, basic.EphemerisSemianalytic} {
		basic.SetEphemerisModel(model)
		for _, date := range []time.Time{systemFixtureDate(), time.Date(2026, 6, 4, 13, 41, 0, 0, time.UTC), time.Date(2026, 9, 10, 6, 33, 0, 0, time.UTC)} {
			document, ok := GalileanSystemSVG(date, GalileanSystemSVGOptions{})
			if !ok || !svgFinite(document) || svgTextBounds(document) != nil {
				t.Fatal("invalid chart under model/date variant")
			}
		}
	}
}

// 视运动箭头方向须与 10 分钟后的公开观测位移同向（东左北上），用独立观测差分核对。
func TestGalileanSystemSVGConsistencyMotionArrows(t *testing.T) {
	document := systemDocument(t, GalileanSystemSVGOptions{})
	elements, _, err := svgElements(document)
	if err != nil {
		t.Fatal(err)
	}
	jde := basic.UTC2TT(basic.Date2JD(systemFixtureDate()))
	for satellite := 1; satellite <= 4; satellite++ {
		var line xml.StartElement
		for index, element := range elements {
			if svgAttribute(element, "id") == fmt.Sprintf("motion-%d", satellite) && index+1 < len(elements) {
				line = elements[index+1]
			}
		}
		if line.Name.Local != "line" {
			t.Fatalf("satellite %d motion arrow missing", satellite)
		}
		before := basic.JupiterGalileanSatelliteObservation(jde-5.0/1440, satellite)
		after := basic.JupiterGalileanSatelliteObservation(jde+5.0/1440, satellite)
		wantX, wantY := -(after.OffsetXJupiterRadii - before.OffsetXJupiterRadii), -(after.OffsetYJupiterRadii - before.OffsetYJupiterRadii)
		gotX, gotY := svgNumber(line, "x2")-svgNumber(line, "x1"), svgNumber(line, "y2")-svgNumber(line, "y1")
		cosine := (gotX*wantX + gotY*wantY) / math.Hypot(gotX, gotY) / math.Hypot(wantX, wantY)
		if cosine < 0.999 {
			t.Fatalf("satellite %d arrow direction cosine %.6f", satellite, cosine)
		}
	}
}

func TestGalileanInsetTransitPaintOrder(t *testing.T) {
	var snapshot basic.JupiterGalileanDiagram
	for index := range snapshot.Observations {
		snapshot.Observations[index].OffsetXJupiterRadii = 0.2
		snapshot.Observations[index].OffsetZJupiterRadii = float64(index - 10)
		snapshot.AngularRadiiJupiterR[index] = 0.05
	}
	snapshot.Observations[0].OffsetZJupiterRadii = -20
	snapshot.Observations[1].OffsetZJupiterRadii = -5
	snapshot.Phenomena[0].Transit, snapshot.Phenomena[1].Transit = true, true
	snapshot.MinorRadiusRatio = 0.93
	drawing := &chart{}
	drawJupiterInset(drawing, snapshot)
	document := drawing.String()
	io, europa := strings.Index(document, `id="inset-transit-1"`), strings.Index(document, `id="inset-transit-2"`)
	if io < 0 || europa < 0 || io < europa {
		t.Fatalf("nearer Io must paint after Europa: io=%d europa=%d", io, europa)
	}
}

func TestGalileanSystemSVGTitleSynthetic(t *testing.T) {
	title := "NaN & Information"
	document := systemDocument(t, GalileanSystemSVGOptions{Title: title})
	if !strings.Contains(document, "NaN &amp; Information") {
		t.Fatalf("missing escaped title %q", title)
	}
	drawing := newChart(normalizeOptions(0, 0, title, "en", nil, false), title, "")
	drawing.WriteString(`<circle cx="NaN" cy="0" r="1"/>`)
	if _, ok := drawing.finish(); ok {
		t.Fatal("nonfinite coordinate rendered")
	}
	for _, invalid := range []string{"NaN", "Inf", "+Inf", "-Inf"} {
		drawing := newChart(normalizeOptions(0, 0, "finite title", "en", nil, false), "finite title", "")
		drawing.WriteString(`<text x="10" y="20">` + invalid + `</text>`)
		if _, ok := drawing.finish(); ok {
			t.Errorf("nonfinite generated text %q rendered", invalid)
		}
	}
}
