package svg

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"math"
	"strings"
	"testing"
	"time"

	"github.com/starainrt/astro/basic"
	"github.com/starainrt/astro/internal/svgchart"
	"github.com/starainrt/astro/jupiter"
)

type mutualFixture struct {
	name, greatestTT, language string
	active, passive            int
	kind                       jupiter.GalileanMutualPhenomenonType
	annular                    bool
	digest                     [2]string
	length                     [2]int
}

func mutualFixtures() []mutualFixture {
	return []mutualFixture{
		{"occultation-zh", "2026-06-04T13:41:43.381", "zh", 1, 2, jupiter.GalileanMutualOccultation, false, [2]string{"e5a53f992335bd31b42003424aa83762", "4cbc59b3e920906d59a67b6636ffca40"}, [2]int{13693, 13693}},
		{"occultation-annular-en", "2026-09-10T06:33:59.172", "en", 2, 1, jupiter.GalileanMutualOccultation, true, [2]string{"9e9975b839453c825977791dec5ffe97", "ceddf1e2ac3a69d1acaad3d1f8a6ce70"}, [2]int{16270, 16270}},
		{"eclipse-zh", "2026-05-14T13:39:37.867", "zh", 1, 2, jupiter.GalileanMutualEclipse, false, [2]string{"6a1bd4afe640462a123aba7e626b4e3b", "bdeb585e5957664e9332985d400cdaf9"}, [2]int{64791, 64791}},
	}
}

func mutualFixtureEvent(t *testing.T, fixture mutualFixture) jupiter.GalileanMutualEvent {
	t.Helper()
	tt, err := time.Parse("2006-01-02T15:04:05.000", fixture.greatestTT)
	if err != nil {
		t.Fatal(err)
	}
	date := basic.JD2DateByZone(basic.TT2UTC(basic.Date2JD(tt)), time.UTC, false)
	event := jupiter.ClosestGalileanMutualEvent(date, fixture.active, fixture.passive, fixture.kind)
	if !event.Valid || event.Active != fixture.active || event.Passive != fixture.passive || event.Type != fixture.kind || event.Annular != fixture.annular || math.Abs(event.Greatest.Sub(date).Seconds()) > 30 {
		t.Fatalf("fixture %s invalid: %+v", fixture.name, event)
	}
	return event
}

func mutualDocument(t *testing.T, event jupiter.GalileanMutualEvent, options GalileanMutualEventSVGOptions) string {
	t.Helper()
	document, ok := GalileanMutualEventSVG(event, options)
	if !ok {
		t.Fatal("mutual chart missing")
	}
	return document
}

func TestGalileanMutualEventSVGSyntheticXML(t *testing.T) {
	for _, fixture := range mutualFixtures() {
		document := mutualDocument(t, mutualFixtureEvent(t, fixture), GalileanMutualEventSVGOptions{Language: fixture.language})
		if _, _, err := svgElements(document); err != nil {
			t.Fatal(err)
		}
	}
}

func TestGalileanMutualEventSVGSyntheticFinite(t *testing.T) {
	for _, fixture := range mutualFixtures() {
		if !svgFinite(mutualDocument(t, mutualFixtureEvent(t, fixture), GalileanMutualEventSVGOptions{})) {
			t.Fatal("nonfinite mutual chart")
		}
	}
}

func TestGalileanMutualEventSVGSyntheticTextBounds(t *testing.T) {
	for _, fixture := range mutualFixtures() {
		event := mutualFixtureEvent(t, fixture)
		for _, options := range []GalileanMutualEventSVGOptions{{}, {Language: "en"}, {Width: 1200, Height: 800, Title: strings.Repeat("很长的事件标题 long title ", 30)}} {
			if err := svgTextBounds(mutualDocument(t, event, options)); err != nil {
				t.Fatal(err)
			}
		}
	}
}

func TestGalileanMutualEventSVGSyntheticDefaults(t *testing.T) {
	event := mutualFixtureEvent(t, mutualFixtures()[0])
	got := mutualDocument(t, event, GalileanMutualEventSVGOptions{})
	want := mutualDocument(t, event, GalileanMutualEventSVGOptions{Width: 900, Height: 600, Language: "zh", Location: time.FixedZone("CST", 8*3600)})
	if got != want || got != mutualDocument(t, event, GalileanMutualEventSVGOptions{Language: "EN"}) || !strings.Contains(got, event.Greatest.In(time.FixedZone("CST", 8*3600)).Format("2006-01-02 15:04:05")+" UTC+8") {
		t.Fatal("default dimensions/language/location differ")
	}
	english := mutualDocument(t, event, GalileanMutualEventSVGOptions{Language: "en", Location: time.UTC})
	if !strings.Contains(english, "Io occults Europa") || !strings.Contains(english, event.Start.UTC().Format("15:04:05")) {
		t.Fatal("explicit language/location ignored")
	}
	if !strings.Contains(got, `fill="#ffffff"`) || strings.Contains(got, "#17283b") {
		t.Fatal("chart is not on the white card style")
	}
}

func mutualByID(t *testing.T, document string) map[string]xml.StartElement {
	t.Helper()
	elements, _, err := svgElements(document)
	if err != nil {
		t.Fatal(err)
	}
	byID := make(map[string]xml.StartElement)
	for _, element := range elements {
		if id := svgAttribute(element, "id"); id != "" {
			byID[id] = element
		}
	}
	return byID
}

func TestGalileanMutualEventSVGConsistencyFieldsAndGeometry(t *testing.T) {
	zone := time.FixedZone("UTC+8", 8*3600)
	for _, fixture := range mutualFixtures() {
		event := mutualFixtureEvent(t, fixture)
		document := mutualDocument(t, event, GalileanMutualEventSVGOptions{Language: "en"})
		for _, label := range []string{fmt.Sprintf("Coverage %.1f%%", 100*event.Coverage), fmt.Sprintf("minimum separation %.4f″", event.MinimumSeparationArcsec), fmt.Sprintf("impact b=%.4f", event.ImpactParameter), event.Start.In(zone).Format("15:04:05"), event.Greatest.In(zone).Format("15:04:05"), event.End.In(zone).Format("15:04:05")} {
			if !strings.Contains(document, label) {
				t.Fatalf("%s missing label %q", fixture.name, label)
			}
		}
		if fixture.annular != strings.Contains(document, "(annular)") {
			t.Fatalf("%s annular class not shown", fixture.name)
		}
		byID := mutualByID(t, document)
		outer := [2]string{"C1", "C4"}
		if fixture.kind == jupiter.GalileanMutualEclipse {
			outer = [2]string{"P1", "P4"}
		}
		for _, id := range []string{"mutual-passive", "mutual-path", "mutual-position-" + outer[0], "mutual-position-" + outer[1]} {
			if _, exists := byID[id]; !exists {
				t.Fatalf("%s missing %s", fixture.name, id)
			}
		}
		centerX, centerY := planeCenterX, layoutFor(event).centerY
		scale := svgNumber(byID["mutual-passive"], "r") / mutualPassiveRadius(event)
		for index, date := range []time.Time{event.Start, event.End} {
			x, y := mutualIndependentOffset(t, event, date)
			marker := byID["mutual-position-"+outer[index]]
			gotX := (centerX - svgNumber(marker, "cx")) / scale
			gotY := (centerY - svgNumber(marker, "cy")) / scale
			if math.Abs(gotX-x) > 2e-3*mutualPassiveRadius(event) || math.Abs(gotY-y) > 2e-3*mutualPassiveRadius(event) {
				t.Fatalf("%s marker %s at (%.6f, %.6f), independent (%.6f, %.6f)", fixture.name, outer[index], gotX, gotY, x, y)
			}
		}
		_, hasUmbra := byID["mutual-umbra"]
		_, hasPenumbra := byID["mutual-penumbra"]
		_, hasActive := byID["mutual-active"]
		if (fixture.kind == jupiter.GalileanMutualEclipse) != (hasUmbra && hasPenumbra && !hasActive) {
			t.Fatal("incorrect shadow-cone section")
		}
	}
}

func mutualPassiveRadius(event jupiter.GalileanMutualEvent) float64 {
	return [...]float64{1829.4, 1562.6, 2631.2, 2410.3}[event.Passive-1] / 71492
}

// mutualIndependentOffset 直接由公开观测量求掩视平面相对位置；互食只核对图心对称与接触半径。
func mutualIndependentOffset(t *testing.T, event jupiter.GalileanMutualEvent, date time.Time) (float64, float64) {
	t.Helper()
	jde := basic.UTC2TT(basic.Date2JD(date.UTC()))
	if event.Type == jupiter.GalileanMutualOccultation {
		active := basic.JupiterGalileanSatelliteObservation(jde, event.Active)
		passive := basic.JupiterGalileanSatelliteObservation(jde, event.Passive)
		return active.OffsetXJupiterRadii - passive.OffsetXJupiterRadii, active.OffsetYJupiterRadii - passive.OffsetYJupiterRadii
	}
	geometry, ok := basic.JupiterGalileanMutualDiagramWithInstance(jde, event.Active, event.Passive, basic.JupiterGalileanMutualEclipse, nil)
	if !ok {
		t.Fatal("eclipse geometry missing")
	}
	return geometry.XJupiterR, geometry.YJupiterR
}

// 每个接触点的两盘圆心距须等于对应半径和或差：外切 = 和，内切 = 差；互食用半影、本影截面。
func TestGalileanMutualEventSVGConsistencyContactsTouch(t *testing.T) {
	for _, fixture := range mutualFixtures() {
		event := mutualFixtureEvent(t, fixture)
		document := mutualDocument(t, event, GalileanMutualEventSVGOptions{})
		byID := mutualByID(t, document)
		checked := 0
		for id, element := range byID {
			if !strings.HasPrefix(id, "mutual-position-") || strings.HasPrefix(id, "mutual-position-g") {
				continue
			}
			code := strings.TrimPrefix(id, "mutual-position-")
			date, ok := mutualContactTime(t, document, code, event.Greatest)
			if !ok {
				t.Fatalf("%s contact %s time missing", fixture.name, code)
			}
			jde := basic.UTC2TT(basic.Date2JD(date.UTC()))
			geometry, valid := basic.JupiterGalileanMutualDiagramWithInstance(jde, event.Active, event.Passive, basic.JupiterGalileanMutualPhenomenonType(event.Type), nil)
			if !valid {
				t.Fatal("geometry missing")
			}
			distance := math.Hypot(geometry.XJupiterR, geometry.YJupiterR)
			var want float64
			switch code {
			case "C1", "C4":
				want = geometry.ActiveRadiusJupiterR + geometry.PassiveRadiusJupiterR
			case "C2", "C3":
				want = math.Abs(geometry.ActiveRadiusJupiterR - geometry.PassiveRadiusJupiterR)
			case "P1", "P4":
				want = geometry.PenumbraRadiusJupiterR + geometry.PassiveRadiusJupiterR
			case "U1", "U4":
				want = geometry.UmbraRadiusJupiterR + geometry.PassiveRadiusJupiterR
			case "U2", "U3":
				want = math.Abs(geometry.UmbraRadiusJupiterR - geometry.PassiveRadiusJupiterR)
			default:
				t.Fatalf("unknown contact %s", code)
			}
			if math.Abs(distance-want) > 0.02*geometry.PassiveRadiusJupiterR {
				t.Fatalf("%s contact %s: separation %.6f, radius combination %.6f", fixture.name, code, distance, want)
			}
			_ = element
			checked++
		}
		minimum := 2
		if fixture.kind == jupiter.GalileanMutualEclipse && event.Coverage > 0 {
			minimum = 4
		}
		if fixture.annular {
			minimum += 2
		}
		if checked < minimum {
			t.Fatalf("%s only %d contacts drawn, want at least %d", fixture.name, checked, minimum)
		}
	}
}

// mutualContactTime 从右侧接触时刻表读出某代号所在行的时刻（显示时区 UTC+8）。
func mutualContactTime(t *testing.T, document, code string, greatest time.Time) (time.Time, bool) {
	t.Helper()
	_, texts, err := svgElements(document)
	if err != nil {
		t.Fatal(err)
	}
	for index, text := range texts {
		if !strings.HasPrefix(text, code+" ") || strings.Count(text, ":") > 0 || index+1 >= len(texts) {
			continue
		}
		clock, err := time.Parse("15:04:05", texts[index+1])
		if err != nil {
			continue
		}
		zone := time.FixedZone("UTC+8", 8*3600)
		local := greatest.In(zone)
		candidate := time.Date(local.Year(), local.Month(), local.Day(), clock.Hour(), clock.Minute(), clock.Second(), 0, zone)
		for _, shift := range []int{-1, 1} {
			if other := candidate.AddDate(0, 0, shift); math.Abs(other.Sub(greatest).Hours()) < math.Abs(candidate.Sub(greatest).Hours()) {
				candidate = other
			}
		}
		return candidate, true
	}
	return time.Time{}, false
}

func TestGalileanMutualEventSVGConsistencyTrueScale(t *testing.T) {
	event := mutualFixtureEvent(t, mutualFixtures()[0])
	plain := mutualDocument(t, event, GalileanMutualEventSVGOptions{})
	if plain != mutualDocument(t, event, GalileanMutualEventSVGOptions{TrueScale: true}) {
		t.Fatal("TrueScale changes a chart that is always drawn to relative scale")
	}
	byID := mutualByID(t, plain)
	got := svgNumber(byID["mutual-active"], "r") / svgNumber(byID["mutual-passive"], "r")
	want := [...]float64{1829.4, 1562.6, 2631.2, 2410.3}[event.Active-1] / [...]float64{1829.4, 1562.6, 2631.2, 2410.3}[event.Passive-1]
	if math.Abs(got-want) > 1e-3 {
		t.Fatalf("active/passive radius ratio %.6f, physical %.6f", got, want)
	}
}

func TestGalileanMutualEventSVGConsistencyInvisibleWarning(t *testing.T) {
	visible := mutualDocument(t, mutualFixtureEvent(t, mutualFixtures()[0]), GalileanMutualEventSVGOptions{})
	hidden := jupiter.NextGalileanMutualEvent(time.Date(2026, 12, 17, 3, 0, 0, 0, time.UTC), 2, 3, jupiter.GalileanMutualEclipse)
	if !hidden.Valid || !hidden.GreatestState[hidden.Passive-1].Eclipse {
		t.Fatal("invisible fixture missing")
	}
	if strings.Contains(visible, "地面不可见") || !strings.Contains(mutualDocument(t, hidden, GalileanMutualEventSVGOptions{}), "地面不可见") {
		t.Fatal("visibility warning wrong")
	}
}

func TestGalileanMutualEventSVGSyntheticInvalid(t *testing.T) {
	event := mutualFixtureEvent(t, mutualFixtures()[0])
	invalids := []jupiter.GalileanMutualEvent{{}}
	for _, mutate := range []func(*jupiter.GalileanMutualEvent){
		func(value *jupiter.GalileanMutualEvent) { value.Active = 0 },
		func(value *jupiter.GalileanMutualEvent) { value.Passive = value.Active },
		func(value *jupiter.GalileanMutualEvent) { value.Type = "bad" },
		func(value *jupiter.GalileanMutualEvent) { value.Coverage = math.NaN() },
		func(value *jupiter.GalileanMutualEvent) { value.ImpactParameter = math.Inf(1) },
		func(value *jupiter.GalileanMutualEvent) { value.End = value.Start.Add(-time.Second) },
	} {
		value := event
		mutate(&value)
		invalids = append(invalids, value)
	}
	for _, invalid := range invalids {
		if document, ok := GalileanMutualEventSVG(invalid, GalileanMutualEventSVGOptions{}); ok || document != "" {
			t.Fatal("invalid event rendered")
		}
	}
}

func TestGalileanMutualEventSVGConsistencyBaseline(t *testing.T) {
	if go116386() {
		t.Skip("Go 1.16/386 floating-point text differs; render-only test covers this build")
	}
	previous := basic.GetEphemerisModel()
	defer basic.SetEphemerisModel(previous)
	for _, fixture := range mutualFixtures() {
		for _, model := range []basic.EphemerisModel{basic.EphemerisLegacy, basic.EphemerisSemianalytic} {
			basic.SetEphemerisModel(model)
			event := mutualFixtureEvent(t, fixture)
			document := mutualDocument(t, event, GalileanMutualEventSVGOptions{Language: fixture.language})
			sum := md5.Sum([]byte(document))
			digest := hex.EncodeToString(sum[:])
			if digest != fixture.digest[model] || len(document) != fixture.length[model] {
				t.Errorf("%s model=%d md5=%s bytes=%d", fixture.name, model, digest, len(document))
			}
		}
	}
}

func TestGalileanMutualEventSVGSyntheticRendersOnGo116386(t *testing.T) {
	if !go116386() {
		t.Skip("Go 1.16/386 only")
	}
	for _, fixture := range mutualFixtures() {
		document := mutualDocument(t, mutualFixtureEvent(t, fixture), GalileanMutualEventSVGOptions{Language: fixture.language})
		if _, _, err := svgElements(document); err != nil || !svgFinite(document) || svgTextBounds(document) != nil {
			t.Fatal("invalid rendered SVG")
		}
	}
}

func TestGalileanMutualFinderOccultationPaintOrder(t *testing.T) {
	at := time.Date(2026, 6, 4, 12, 40, 0, 0, time.UTC)
	event := jupiter.NextGalileanMutualEvent(at, 1, 2, jupiter.GalileanMutualOccultation)
	if !event.Valid || event.Active != 1 || event.Passive != 2 || event.Coverage < 0.9 {
		t.Fatalf("missing Io occultation of Europa: %+v", event)
	}
	document := mutualDocument(t, event, GalileanMutualEventSVGOptions{})
	front := strings.Index(document, `id="finder-satellite-1"`)
	back := strings.Index(document, `id="finder-satellite-2"`)
	if back < 0 || front <= back {
		t.Fatalf("Io drawn at %d, Europa at %d", front, back)
	}
}

// 寻星小图的两星角距与相位角须与公开观测独立计算一致；临边距在中心距减卫星角半径与木星半径之间。
func TestGalileanMutualEventSVGConsistencyFinder(t *testing.T) {
	for _, fixture := range mutualFixtures() {
		event := mutualFixtureEvent(t, fixture)
		document := mutualDocument(t, event, GalileanMutualEventSVGOptions{Language: "en"})
		jde := basic.UTC2TT(basic.Date2JD(event.Greatest.UTC()))
		active := basic.JupiterGalileanSatelliteObservation(jde, event.Active)
		passive := basic.JupiterGalileanSatelliteObservation(jde, event.Passive)
		pair := math.Hypot(active.OffsetXArcsec-passive.OffsetXArcsec, active.OffsetYArcsec-passive.OffsetYArcsec)
		if !strings.Contains(document, fmt.Sprintf("%.2f″", pair)) {
			t.Fatalf("%s pair separation %.1f″ missing", fixture.name, pair)
		}
		if !strings.Contains(document, fmt.Sprintf("Jupiter phase angle %.1f°", jupiter.PhaseAngle(event.Greatest))) {
			t.Fatalf("%s phase angle missing", fixture.name)
		}
		semidiameter := math.Hypot(passive.OffsetXArcsec, passive.OffsetYArcsec) / math.Hypot(passive.OffsetXJupiterRadii, passive.OffsetYJupiterRadii)
		centre := math.Hypot(passive.OffsetXArcsec, passive.OffsetYArcsec)
		_, texts, _ := svgElements(document)
		found := false
		for _, text := range texts {
			var name string
			var limb float64
			if n, _ := fmt.Sscanf(strings.TrimPrefix(text, "● "), "%s to Jupiter limb %f″", &name, &limb); n == 2 {
				found = true
				if limb > centre-0.9*semidiameter || limb < centre-1.0*semidiameter-1 {
					t.Fatalf("%s limb gap %.1f″ outside [%.1f, %.1f]", fixture.name, limb, centre-semidiameter-1, centre-0.9*semidiameter)
				}
			}
		}
		if !found {
			t.Fatalf("%s limb gap missing", fixture.name)
		}
		byID := mutualByID(t, document)
		for _, id := range []string{"finder-frame", "finder-jupiter", "finder-satellite-1", "finder-satellite-4", "finder-pair-passive", "finder-pair-active"} {
			if _, ok := byID[id]; !ok {
				t.Fatalf("%s finder missing %s", fixture.name, id)
			}
		}
	}
}

// 2021-08-22 木卫三食木卫二：U4、P4 时木卫二已被木卫三掩住，图中须标为不可见；用公开观测独立求遮挡比例核对。
func TestGalileanMutualEventSVGConsistencyHiddenContacts(t *testing.T) {
	zone := time.FixedZone("UTC+8", 8*3600)
	event := jupiter.NextGalileanMutualEvent(time.Date(2021, 8, 22, 21, 0, 0, 0, zone), 3, 2, jupiter.GalileanMutualEclipse)
	if !event.Valid || event.Active != 3 || event.Passive != 2 {
		t.Fatalf("fixture missing: %+v", event)
	}
	document := mutualDocument(t, event, GalileanMutualEventSVGOptions{})
	for code, wantHidden := range map[string]bool{"P1": false, "U1": false, "U2": false, "U3": false, "U4": true, "P4": true} {
		date, ok := mutualContactTime(t, document, code, event.Greatest)
		if !ok {
			t.Fatalf("contact %s missing", code)
		}
		jde := basic.UTC2TT(basic.Date2JD(date.UTC()))
		front := basic.JupiterGalileanSatelliteObservation(jde, 3)
		back := basic.JupiterGalileanSatelliteObservation(jde, 2)
		backRadius := math.Atan2(1562.6, back.Distance*149597870.7) * 180 / math.Pi * 3600
		frontRadius := math.Atan2(2631.2, front.Distance*149597870.7) * 180 / math.Pi * 3600
		separation := math.Hypot(front.OffsetXArcsec-back.OffsetXArcsec, front.OffsetYArcsec-back.OffsetYArcsec)
		fraction := 0.0
		if front.OffsetZJupiterRadii < back.OffsetZJupiterRadii {
			fraction = independentLens(separation, frontRadius, backRadius) / (math.Pi * backRadius * backRadius)
		}
		if (fraction >= 0.5) != wantHidden {
			t.Fatalf("contact %s: independent occulted fraction %.3f, want hidden=%v", code, fraction, wantHidden)
		}
		if strings.Contains(document, code+" "+date.Format("15:04:05")+"（不可见）") != wantHidden {
			t.Fatalf("contact %s hidden mark wrong", code)
		}
	}
	if !strings.Contains(document, "U4、P4 时被食星被木卫三掩") {
		t.Fatal("hidden-contact warning missing")
	}
	plain := mutualDocument(t, mutualFixtureEvent(t, mutualFixtures()[2]), GalileanMutualEventSVGOptions{})
	if strings.Contains(plain, "（不可见）") || strings.Contains(plain, "地面看不到") {
		t.Fatal("visible fixture marked hidden")
	}
}

func independentLens(distance, first, second float64) float64 {
	if distance >= first+second {
		return 0
	}
	if distance <= math.Abs(first-second) {
		smaller := math.Min(first, second)
		return math.Pi * smaller * smaller
	}
	a := math.Acos((distance*distance + first*first - second*second) / (2 * distance * first))
	b := math.Acos((distance*distance + second*second - first*first) / (2 * distance * second))
	return first*first*a + second*second*b - 0.5*math.Sqrt((-distance+first+second)*(distance+first-second)*(distance-first+second)*(distance+first+second))
}

// 地球视点相位条：每个接触点一个小盘；缘上三角指向该时刻影斑的质心方向（按图中扫描段面积独立求质心核对）。
func TestGalileanMutualEventSVGConsistencyEarthStrip(t *testing.T) {
	event := jupiter.NextGalileanMutualEvent(time.Date(2026, 5, 31, 0, 0, 0, 0, time.UTC), 1, 2, jupiter.GalileanMutualEclipse)
	document := mutualDocument(t, event, GalileanMutualEventSVGOptions{})
	elements, _, err := svgElements(document)
	if err != nil {
		t.Fatal(err)
	}
	disks := 0
	for index, element := range elements {
		id := svgAttribute(element, "id")
		if !strings.HasPrefix(id, "earth-disk-") {
			continue
		}
		disks++
		code := strings.TrimPrefix(id, "earth-disk-")
		circle := elements[index+1]
		cx, cy := svgNumber(circle, "cx"), svgNumber(circle, "cy")
		var sx, sy, area float64
		for next := index + 2; next < len(elements) && elements[next].Name.Local == "rect"; next++ {
			fill := svgAttribute(elements[next], "fill")
			if fill != "#c4b9a6" && fill != "#b21f16" && fill != "#e08a2c" {
				continue
			}
			w, h := svgNumber(elements[next], "width"), svgNumber(elements[next], "height")
			sx += w * h * (svgNumber(elements[next], "x") + w/2 - cx)
			sy += w * h * (svgNumber(elements[next], "y") + h/2 - cy)
			area += w * h
		}
		var tick xml.StartElement
		for _, other := range elements {
			if svgAttribute(other, "id") == "earth-tick-"+code {
				tick = other
			}
		}
		if code == "U1" || code == "U4" {
			if area == 0 || tick.Name.Local == "" {
				t.Fatalf("%s disk lacks shadow or tick", code)
			}
			var tx, ty float64
			fmt.Sscanf(svgAttribute(tick, "d"), "M %f %f", &tx, &ty)
			tx, ty = tx-cx, ty-cy
			cosine := (sx*tx + sy*ty) / math.Hypot(sx, sy) / math.Hypot(tx, ty)
			if cosine < 0.98 {
				t.Fatalf("%s tick direction cosine %.4f", code, cosine)
			}
		}
	}
	if disks != 5 {
		t.Fatalf("earth strip disks %d, want P1 U1 G U4 P4", disks)
	}
	if !strings.Contains(document, "地球视角本影/伪本影") || !strings.Contains(document, "木星相位角 7.6°") {
		t.Fatal("earth-view coverage or phase angle missing")
	}
	occultation := mutualDocument(t, mutualFixtureEvent(t, mutualFixtures()[0]), GalileanMutualEventSVGOptions{})
	if strings.Contains(occultation, "earth-disk-") {
		t.Fatal("occultation chart must not carry the eclipse earth strip")
	}
}

// 慢事件（2026-06-01 木卫一食木卫二，接触点挤在一起）的接触标签不得互相重叠，且每条引线都从对应接触点出发。
func TestGalileanMutualEventSVGConsistencyCalloutsSeparate(t *testing.T) {
	for _, event := range []jupiter.GalileanMutualEvent{
		jupiter.NextGalileanMutualEvent(time.Date(2026, 5, 31, 0, 0, 0, 0, time.UTC), 1, 2, jupiter.GalileanMutualEclipse),
		jupiter.NextGalileanMutualEvent(time.Date(2021, 8, 22, 13, 0, 0, 0, time.UTC), 3, 2, jupiter.GalileanMutualEclipse),
	} {
		document := mutualDocument(t, event, GalileanMutualEventSVGOptions{})
		elements, texts, err := svgElements(document)
		if err != nil {
			t.Fatal(err)
		}
		var boxes []svgchart.LabelBox
		for index, element := range elements {
			if element.Name.Local != "text" || svgAttribute(element, "fill") != "#2554c7" && svgAttribute(element, "fill") != "#9a9a9a" {
				continue
			}
			if !strings.Contains(texts[index], ":") {
				continue
			}
			boxes = append(boxes, svgchart.LabelTextBox(svgNumber(element, "x"), svgNumber(element, "y"), svgNumber(element, "font-size"), texts[index], svgAttribute(element, "text-anchor")))
		}
		for i := range boxes {
			for j := i + 1; j < len(boxes); j++ {
				a, b := boxes[i], boxes[j]
				if a.X < b.X+b.Width && b.X < a.X+a.Width && a.Y < b.Y+b.Height && b.Y < a.Y+a.Height {
					t.Fatalf("contact labels %d and %d overlap", i, j)
				}
			}
		}
		markers := map[[2]float64]bool{}
		for _, element := range elements {
			if strings.HasPrefix(svgAttribute(element, "id"), "mutual-position-") {
				markers[[2]float64{svgNumber(element, "cx"), svgNumber(element, "cy")}] = true
			}
		}
		callouts := 0
		for _, element := range elements {
			if svgAttribute(element, "class") != "callout" {
				continue
			}
			callouts++
			var x, y float64
			fmt.Sscanf(svgAttribute(element, "d"), "M %f %f", &x, &y)
			if !markers[[2]float64{x, y}] {
				t.Fatalf("callout starts at (%.3f, %.3f), not a contact marker", x, y)
			}
		}
		if callouts != len(markers) || len(boxes) < callouts {
			t.Fatalf("callouts %d, markers %d, labels %d", callouts, len(markers), len(boxes))
		}
	}
}

// 掩的地球视点条：极大处小盘上红圆与被掩盘的交面积比须等于事件 Coverage（独立圆交公式），各接触点都有一个小盘。
func TestGalileanMutualEventSVGConsistencyOccultationStrip(t *testing.T) {
	for _, fixture := range mutualFixtures()[:2] {
		event := mutualFixtureEvent(t, fixture)
		document := mutualDocument(t, event, GalileanMutualEventSVGOptions{})
		elements, _, err := svgElements(document)
		if err != nil {
			t.Fatal(err)
		}
		disks := 0
		for index, element := range elements {
			id := svgAttribute(element, "id")
			if !strings.HasPrefix(id, "occultation-disk-") {
				continue
			}
			disks++
			if !strings.HasPrefix(id, "occultation-disk-g") {
				continue
			}
			var disk, cover xml.StartElement
			for next := index + 1; next < len(elements) && next < index+6; next++ {
				if elements[next].Name.Local == "circle" && disk.Name.Local == "" && svgAttribute(elements[next], "fill") != "#b21f16" && svgAttribute(elements[next], "fill") != "none" && elements[next-1].Name.Local != "clipPath" {
					disk = elements[next]
				}
				if svgAttribute(elements[next], "class") == "occultation-cover" {
					cover = elements[next]
				}
			}
			r := svgNumber(disk, "r")
			distance := math.Hypot(svgNumber(cover, "cx")-svgNumber(disk, "cx"), svgNumber(cover, "cy")-svgNumber(disk, "cy"))
			got := independentLens(distance, svgNumber(cover, "r"), r) / (math.Pi * r * r)
			if math.Abs(got-event.Coverage) > 2e-3 {
				t.Fatalf("%s strip coverage %.4f, event %.4f", fixture.name, got, event.Coverage)
			}
		}
		minimum := 3
		if fixture.annular {
			minimum = 5
		}
		if disks != minimum {
			t.Fatalf("%s strip disks %d, want %d", fixture.name, disks, minimum)
		}
	}
}

// 影核：本影 r_u>0 用红、伪本影 r_u<0 用橙；半径按独立锥公式核对，过小时放大到可见并在图例注明。
func TestGalileanMutualEventSVGConsistencyCoreShadow(t *testing.T) {
	cases := []struct {
		name string
		date time.Time
		a, b int
	}{
		{"2026-11-16 tiny umbra", time.Date(2026, 11, 15, 0, 0, 0, 0, time.UTC), 1, 4},
		{"2021-08-22 total", time.Date(2021, 8, 22, 13, 0, 0, 0, time.UTC), 3, 2},
	}
	for _, test := range cases {
		event := jupiter.NextGalileanMutualEvent(test.date, test.a, test.b, jupiter.GalileanMutualEclipse)
		document := mutualDocument(t, event, GalileanMutualEventSVGOptions{})
		jde := basic.UTC2TT(basic.Date2JD(event.Greatest.UTC()))
		geometry, ok := basic.JupiterGalileanMutualDiagramWithInstance(jde, event.Active, event.Passive, basic.JupiterGalileanMutualEclipse, nil)
		if !ok {
			t.Fatal("geometry missing")
		}
		radii := [...]float64{1829.4, 1562.6, 2631.2, 2410.3}
		activeKM := radii[event.Active-1]
		axialKM := (geometry.PenumbraRadiusJupiterR*71492 - activeKM) / ((695700 + activeKM) / 1)
		umbraKM := activeKM - axialKM*(695700-activeKM)
		if math.Abs(umbraKM-geometry.UmbraRadiusJupiterR*71492) > 0.5 {
			t.Fatalf("%s umbra %.2f km vs independent %.2f km", test.name, geometry.UmbraRadiusJupiterR*71492, umbraKM)
		}
		byID := mutualByID(t, document)
		want, fill := "mutual-umbra", "#b21f16"
		if umbraKM < 0 {
			want, fill = "mutual-antumbra", "#e08a2c"
		}
		core, exists := byID[want]
		if !exists || svgAttribute(core, "fill") != fill {
			t.Fatalf("%s core shadow %s missing", test.name, want)
		}
		if svgNumber(core, "r") < 2.5 {
			t.Fatalf("%s core shadow not visible: r=%.3f", test.name, svgNumber(core, "r"))
		}
		label := fmt.Sprintf("r=%.0f km", math.Abs(umbraKM))
		if !strings.Contains(document, label) {
			t.Fatalf("%s legend lacks %q", test.name, label)
		}
		if (math.Abs(umbraKM) < 120) != strings.Contains(document, "（放大）") {
			t.Fatalf("%s enlargement note wrong", test.name)
		}
	}
}

func TestGalileanMutualCoreEnlargementConsistency(t *testing.T) {
	seed := time.Date(2026, 10, 12, 0, 0, 0, 0, time.UTC)
	event := jupiter.NextGalileanMutualEvent(seed, 1, 4, jupiter.GalileanMutualEclipse)
	if !event.Valid || event.Active != 1 || event.Passive != 4 || event.Greatest.UTC().Day() != 13 {
		t.Fatalf("missing Io–Callisto eclipse: %+v", event)
	}
	for _, test := range []struct{ language, legend string }{{"en", "r=123 km (enlarged)"}, {"zh", "r=123 km（放大）"}} {
		document := mutualDocument(t, event, GalileanMutualEventSVGOptions{Language: test.language})
		core := mutualByID(t, document)["mutual-umbra"]
		if svgNumber(core, "r") != minimumCoreRadius || !strings.Contains(document, test.legend) {
			t.Fatalf("%s: missing enlargement annotation %q", test.language, test.legend)
		}
	}
}

func TestGalileanMutualEventSVGEnglishHiddenAndDiskOverlap(t *testing.T) {
	hidden := jupiter.NextGalileanMutualEvent(time.Date(2026, 12, 17, 3, 0, 0, 0, time.UTC), 2, 3, jupiter.GalileanMutualEclipse)
	if !hidden.Valid {
		t.Fatal("missing hidden eclipse")
	}
	document := mutualDocument(t, hidden, GalileanMutualEventSVGOptions{Language: "en"})
	if !strings.Contains(document, "✕ not observable") || !svgFinite(document) || svgTextBounds(document) != nil {
		t.Fatal("invalid English hidden eclipse SVG")
	}
	if math.IsNaN(diskOverlap(math.Nextafter(1.0, 2.0), 2.0, 1.0)) {
		t.Fatal("near-tangent diskOverlap returned NaN")
	}
}
