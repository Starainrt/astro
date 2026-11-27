package svg

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"math"
	"strings"
	"testing"
	"time"

	"github.com/starainrt/astro/basic"
	"github.com/starainrt/astro/internal/svgchart"
	"github.com/starainrt/astro/jupiter"
)

// 单星现象图契约：四种类型各一场，接触标记位置对公开观测独立核对，木影食标记在太阳视平面，不可见接触置灰。
type phenomenonFixture struct {
	name      string
	satellite int
	kind      jupiter.GalileanPhenomenonType
	language  string
	digest    [2]string
	length    [2]int
}

func phenomenonFixtures() []phenomenonFixture {
	return []phenomenonFixture{
		{"transit-zh", 1, jupiter.GalileanPhenomenonTransit, "zh", [2]string{"675fbd40ba85e9523682b3c836523e71", "ad535ccaba1fb366e0972493521f57ea"}, [2]int{9282, 9282}},
		{"shadow-en", 1, jupiter.GalileanPhenomenonShadowTransit, "en", [2]string{"61324218a6e55619bf0bc8513cf34af4", "63043b7015b63b3f03fce0ed7a9b7478"}, [2]int{9464, 9464}},
		{"eclipse-zh", 1, jupiter.GalileanPhenomenonEclipse, "zh", [2]string{"4ab28b07dd6251fe56d693670ef9339f", "e6f4ea700de74615d7d72182d9a175f7"}, [2]int{11567, 11567}},
		{"occultation-en", 1, jupiter.GalileanPhenomenonOccultation, "en", [2]string{"fca202a6b7208880bfea4a7d194d3a6c", "e036716ba9f983dd2cc3deda5cd825ea"}, [2]int{9248, 9248}},
	}
}

func phenomenonFixtureEvent(t *testing.T, fixture phenomenonFixture) jupiter.GalileanPhenomenonContactEvent {
	t.Helper()
	event := jupiter.NextGalileanPhenomenonContactEvent(time.Date(2026, 5, 14, 0, 0, 0, 0, time.UTC), fixture.satellite, fixture.kind)
	if !event.Valid {
		t.Fatalf("%s fixture missing", fixture.name)
	}
	return event
}

func phenomenonDocument(t *testing.T, event jupiter.GalileanPhenomenonContactEvent, options GalileanPhenomenonSVGOptions) string {
	t.Helper()
	document, ok := GalileanPhenomenonSVG(event, options)
	if !ok {
		t.Fatal("phenomenon chart missing")
	}
	return document
}

func TestGalileanPhenomenonSVGSyntheticStructure(t *testing.T) {
	for _, fixture := range phenomenonFixtures() {
		event := phenomenonFixtureEvent(t, fixture)
		for _, options := range []GalileanPhenomenonSVGOptions{{}, {Language: "en"}, {Width: 1200, Height: 800, Title: strings.Repeat("很长的标题 long title ", 30)}} {
			document := phenomenonDocument(t, event, options)
			if _, _, err := svgElements(document); err != nil {
				t.Fatal(err)
			}
			if !svgFinite(document) || svgTextBounds(document) != nil {
				t.Fatalf("%s invalid document", fixture.name)
			}
		}
	}
	if document, ok := GalileanPhenomenonSVG(jupiter.GalileanPhenomenonContactEvent{}, GalileanPhenomenonSVGOptions{}); ok || document != "" {
		t.Fatal("invalid event rendered")
	}
}

func TestGalileanPhenomenonSVGConsistencyPositions(t *testing.T) {
	for _, fixture := range phenomenonFixtures() {
		if fixture.kind == jupiter.GalileanPhenomenonShadowTransit || fixture.kind == jupiter.GalileanPhenomenonEclipse {
			continue
		}
		event := phenomenonFixtureEvent(t, fixture)
		byID := mutualByID(t, phenomenonDocument(t, event, GalileanPhenomenonSVGOptions{}))
		markers := [2]string{"phenomenon-position-0", "phenomenon-position-4"}
		dates := [2]time.Time{event.Disappearance.Start, event.Reappearance.End}
		var screen, model [2][2]float64
		for index := range markers {
			marker, ok := byID[markers[index]]
			if !ok {
				t.Fatalf("%s marker missing", fixture.name)
			}
			screen[index] = [2]float64{svgNumber(marker, "cx"), svgNumber(marker, "cy")}
			jde := basic.UTC2TT(basic.Date2JD(dates[index].UTC()))
			observation := basic.JupiterGalileanSatelliteObservation(jde, fixture.satellite)
			model[index] = [2]float64{observation.OffsetXJupiterRadii, observation.OffsetYJupiterRadii}
		}
		dxScreen, dyScreen := screen[1][0]-screen[0][0], screen[1][1]-screen[0][1]
		dxModel, dyModel := -(model[1][0] - model[0][0]), -(model[1][1] - model[0][1])
		cosine := (dxScreen*dxModel + dyScreen*dyModel) / math.Hypot(dxScreen, dyScreen) / math.Hypot(dxModel, dyModel)
		if cosine < 0.999 {
			t.Fatalf("%s C1→C4 direction cosine %.5f against public observations", fixture.name, cosine)
		}
	}
}

func TestGalileanPhenomenonSVGConsistencyEclipseContacts(t *testing.T) {
	for _, satellite := range []int{1, 4} {
		event := jupiter.NextGalileanPhenomenonContactEvent(time.Date(2026, 5, 14, 0, 0, 0, 0, time.UTC), satellite, jupiter.GalileanPhenomenonEclipse)
		document := phenomenonDocument(t, event, GalileanPhenomenonSVGOptions{})
		previous := time.Time{}
		for _, code := range []string{"P1", "U1", "U2", "Mid", "U3", "U4", "P4"} {
			date, ok := mutualContactTime(t, document, code, event.Greatest)
			if code == "Mid" {
				date, ok = mutualContactTime(t, document, "甚", event.Greatest)
			}
			if !ok || (!previous.IsZero() && !date.After(previous)) {
				t.Fatalf("satellite %d missing or unordered %s", satellite, code)
			}
			previous = date
			if code == "Mid" {
				middle := independentJovianEdge(date, satellite, false)
				if middle > independentJovianEdge(date.Add(-time.Minute), satellite, false) || middle > independentJovianEdge(date.Add(time.Minute), satellite, false) {
					t.Fatal("greatest is not an umbra-edge minimum")
				}
				continue
			}
			penumbra := strings.HasPrefix(code, "P")
			radius := [...]float64{1829.4, 1562.6, 2631.2, 2410.3}[satellite-1] / 71492
			if code == "U2" || code == "U3" {
				radius = -radius
			}
			edge := independentJovianEdge(date, satellite, penumbra)
			rounding := math.Abs(independentJovianEdge(date.Add(time.Second), satellite, penumbra)-independentJovianEdge(date.Add(-time.Second), satellite, penumbra)) / 2
			if math.Abs(edge-radius) > rounding+5e-6 {
				t.Fatalf("satellite %d %s independent tangency residual %.9f RJ", satellite, code, edge-radius)
			}
		}
		for _, invalid := range []string{"一触", "二触", "三触", "四触", "开始变暗", "消失", "重现"} {
			if strings.Contains(document, invalid) {
				t.Fatalf("eclipse has obsolete label %q", invalid)
			}
		}
		changed := event
		changed.Disappearance.Start = event.Greatest.Add(-2 * time.Hour)
		changed.Disappearance.End = event.Greatest.Add(-time.Hour)
		changed.Reappearance.Start = event.Greatest.Add(time.Hour)
		changed.Reappearance.End = event.Greatest.Add(2 * time.Hour)
		if document != phenomenonDocument(t, changed, GalileanPhenomenonSVGOptions{}) {
			t.Fatal("geometric eclipse chart depends on D/F windows")
		}
	}
}

func independentJovianEdge(date time.Time, satellite int, penumbra bool) float64 {
	jde := basic.UTC2TT(basic.Date2JD(date.UTC()))
	geometry, _ := basic.JupiterGalileanPhenomenonDiagramWithInstance(jde, satellite, nil)
	observation := basic.JupiterGalileanSatelliteObservation(jde, satellite)
	state := observation.State
	radiusAU := 71492.0 / 149597870.691
	projected := math.Hypot(geometry.SunX, geometry.SunY)
	behind := math.Sqrt(math.Max(0, (state.X*state.X+state.Y*state.Y+state.Z*state.Z)/(radiusAU*radiusAU)-projected*projected))
	solarDistance := basic.JupiterR(jde - observation.Distance*499.004783836/86400)
	slope := -(695700.0 - 71492) / (solarDistance * 149597870.691)
	if penumbra {
		slope = (695700.0 + 71492) / (solarDistance * 149597870.691)
	}
	major := 1 + behind*slope
	minor := major * geometry.SunMinorRadiusRatio
	pointX, pointY := math.Abs(geometry.SunX), math.Abs(geometry.SunY)
	distanceAt := func(angle float64) float64 {
		return math.Hypot(pointX-major*math.Cos(angle), pointY-minor*math.Sin(angle))
	}
	best, step := 0.0, math.Pi/2048
	for index := 1; index <= 1024; index++ {
		angle := float64(index) * step
		if distanceAt(angle) < distanceAt(best) {
			best = angle
		}
	}
	left, right := math.Max(0, best-step), math.Min(math.Pi/2, best+step)
	for iteration := 0; iteration < 48; iteration++ {
		first, second := left+(right-left)/3, right-(right-left)/3
		if distanceAt(first) < distanceAt(second) {
			right = second
		} else {
			left = first
		}
	}
	distance := distanceAt((left + right) / 2)
	if pointX*pointX/(major*major)+pointY*pointY/(minor*minor) < 1 {
		distance = -distance
	}
	return distance
}

func TestGalileanPhenomenonSVGConsistencyEclipseVisibility(t *testing.T) {
	hiddenByJupiter, hiddenBySatellite, visibleInShadow := 0, 0, 0
	for _, fixture := range []struct {
		satellite int
		date      time.Time
	}{
		{1, time.Date(2026, 5, 14, 0, 0, 0, 0, time.UTC)},
		{4, time.Date(2026, 5, 14, 0, 0, 0, 0, time.UTC)},
		{2, time.Date(2026, 5, 25, 0, 0, 0, 0, time.UTC)},
	} {
		event := jupiter.NextGalileanPhenomenonContactEvent(fixture.date, fixture.satellite, jupiter.GalileanPhenomenonEclipse)
		document := phenomenonDocument(t, event, GalileanPhenomenonSVGOptions{})
		byID := mutualByID(t, document)
		elements, texts, err := svgElements(document)
		if err != nil {
			t.Fatal(err)
		}
		hiddenCount := 0
		for index, code := range []string{"P1", "U1", "U2", "甚", "U3", "U4", "P4"} {
			date, ok := mutualContactTime(t, document, code, event.Greatest)
			if !ok {
				t.Fatalf("missing %s", code)
			}
			jde := basic.UTC2TT(basic.Date2JD(date.UTC()))
			state := basic.JupiterGalileanSatellitePhenomenon(jde, fixture.satellite)
			target := basic.JupiterGalileanSatelliteObservation(jde, fixture.satellite)
			radii := [...]float64{1829.4, 1562.6, 2631.2, 2410.3}
			backRadius := math.Atan2(radii[fixture.satellite-1], target.Distance*149597870.691) * 180 / math.Pi * 3600
			fraction := 0.0
			for front := 1; front <= 4; front++ {
				observation := basic.JupiterGalileanSatelliteObservation(jde, front)
				if front == fixture.satellite || observation.OffsetZJupiterRadii >= target.OffsetZJupiterRadii {
					continue
				}
				frontRadius := math.Atan2(radii[front-1], observation.Distance*149597870.691) * 180 / math.Pi * 3600
				distance := math.Hypot(observation.OffsetXArcsec-target.OffsetXArcsec, observation.OffsetYArcsec-target.OffsetYArcsec)
				fraction = math.Max(fraction, independentLens(distance, frontRadius, backRadius)/(math.Pi*backRadius*backRadius))
			}
			hidden := state.Occultation || fraction >= 0.5
			if hidden {
				hiddenCount++
			}
			if state.Occultation {
				hiddenByJupiter++
			}
			if !state.Occultation && fraction >= 0.5 {
				hiddenBySatellite++
			}
			if !hidden && state.Eclipse {
				visibleInShadow++
			}
			if strings.Contains(document, code+" "+date.Format("15:04:05")+"（不可见）") != hidden {
				t.Fatalf("satellite %d %s visibility differs: occultation=%v fraction=%.6f", fixture.satellite, code, state.Occultation, fraction)
			}
			marker := byID[fmt.Sprintf("phenomenon-position-%d", index)]
			if (svgAttribute(marker, "fill") == "#9a9a9a") != hidden {
				t.Fatal("contact marker is not grey exactly when hidden")
			}
			foundRow := false
			for row, text := range texts {
				if strings.HasPrefix(text, code+" ") && svgNumber(elements[row], "x") == panelX {
					foundRow = true
					if (svgAttribute(elements[row], "fill") == "#9a9a9a") != hidden {
						t.Fatal("table contact visibility colour differs")
					}
				}
			}
			if !foundRow {
				t.Fatal("table contact row missing")
			}
		}
		crosses, warning := 0, false
		for index, text := range texts {
			if text == "✕" {
				crosses++
			}
			if strings.Contains(text, "地面不可见") && svgNumber(elements[index], "y") < 160 && svgAttribute(elements[index], "fill") == "#b21f16" {
				warning = true
			}
		}
		if crosses != hiddenCount || warning != (hiddenCount > 0) {
			t.Fatalf("crosses=%d hidden=%d warning=%v", crosses, hiddenCount, warning)
		}
	}
	if hiddenByJupiter == 0 || hiddenBySatellite == 0 || visibleInShadow == 0 {
		t.Fatalf("fixtures do not cover visibility cases: Jupiter=%d satellite=%d visible in shadow=%d", hiddenByJupiter, hiddenBySatellite, visibleInShadow)
	}
}

func TestGalileanPhenomenonSVGConsistencyGrazingEclipse(t *testing.T) {
	seed := time.Date(2025, 6, 14, 12, 0, 0, 0, time.UTC)
	event := jupiter.NextGalileanPhenomenonContactEvent(seed, 4, jupiter.GalileanPhenomenonEclipse)
	if !event.Valid || event.Greatest.Day() != 14 || event.Greatest.Month() != time.June {
		t.Fatal("grazing fixture missing")
	}
	document := phenomenonDocument(t, event, GalileanPhenomenonSVGOptions{})
	minimum := math.Inf(1)
	for minute := -120; minute <= 120; minute++ {
		minimum = math.Min(minimum, independentJovianEdge(event.Greatest.Add(time.Duration(minute)*time.Minute), 4, false))
	}
	if minimum <= -2410.3/71492 || minimum >= 0 {
		t.Fatalf("fixture is not a grazing umbral eclipse: edge %.9f", minimum)
	}
	for _, code := range []string{"U2", "U3"} {
		if _, exists := mutualContactTime(t, document, code, event.Greatest); exists {
			t.Fatal("grazing eclipse has inner contact " + code)
		}
	}
	for _, code := range []string{"P1", "U1", "甚", "U4", "P4"} {
		if _, exists := mutualContactTime(t, document, code, event.Greatest); !exists {
			t.Fatal("grazing eclipse missing " + code)
		}
	}
}

func TestGalileanPhenomenonSVGConsistencyEclipseTextLayout(t *testing.T) {
	for _, fixture := range []struct {
		satellite int
		seed      time.Time
	}{
		{1, time.Date(2026, 5, 14, 0, 0, 0, 0, time.UTC)},
		{4, time.Date(2026, 5, 14, 0, 0, 0, 0, time.UTC)},
		{4, time.Date(2025, 6, 14, 12, 0, 0, 0, time.UTC)},
	} {
		geometry := jupiter.NextGalileanPhenomenonEvent(fixture.seed, fixture.satellite, jupiter.GalileanPhenomenonEclipse)
		event := jupiter.GalileanPhenomenonContactEvent{Valid: geometry.Valid, Satellite: fixture.satellite, Type: geometry.Type, Greatest: geometry.Greatest}
		for _, language := range []string{"zh", "en"} {
			document := phenomenonDocument(t, event, GalileanPhenomenonSVGOptions{Language: language})
			elements, texts, err := svgElements(document)
			if err != nil {
				t.Fatal(err)
			}
			var boxes []svgchart.LabelBox
			var labels []string
			for index, text := range texts {
				if text == "" {
					continue
				}
				box := svgchart.LabelTextBox(svgNumber(elements[index], "x"), svgNumber(elements[index], "y"), svgNumber(elements[index], "font-size"), text, svgAttribute(elements[index], "text-anchor"))
				for previous, other := range boxes {
					if box.X < other.X+other.Width && other.X < box.X+box.Width && box.Y < other.Y+other.Height && other.Y < box.Y+box.Height {
						t.Errorf("satellite %d %s: overlapping text %q and %q", fixture.satellite, language, text, labels[previous])
					}
				}
				boxes = append(boxes, box)
				labels = append(labels, text)
			}
		}
	}
}

func TestGalileanPhenomenonSVGConsistencyShadowIngress(t *testing.T) {
	fixture := phenomenonFixtures()[1]
	event := phenomenonFixtureEvent(t, fixture)
	byID := mutualByID(t, phenomenonDocument(t, event, GalileanPhenomenonSVGOptions{}))
	first, disk := byID["phenomenon-contact-1"], byID["phenomenon-jupiter"]
	if byID["phenomenon-shadow-limb"].Name.Local != "clipPath" || svgAttribute(byID["phenomenon-body"], "clip-path") != "url(#phenomenon-shadow-limb)" || svgAttribute(first, "clip-path") != "" {
		t.Fatal("clip the filled shadow, not the contact outline")
	}
	distance := math.Hypot(svgNumber(first, "cx")-svgNumber(disk, "cx"), svgNumber(first, "cy")-svgNumber(disk, "cy"))
	if ratio := distance / svgNumber(disk, "rx"); ratio < 0.9 || ratio > 1.15 {
		t.Fatalf("first shadow contact is not near the limb: %.6f RJ", ratio)
	}
}

func TestGalileanPhenomenonSVGExternalContacts(t *testing.T) {
	for _, fixture := range []struct {
		day       string
		satellite int
		kind      jupiter.GalileanPhenomenonType
		contacts  map[string]string
		absent    []string
		tolerance time.Duration
	}{
		{"2026-05-15", 1, jupiter.GalileanPhenomenonShadowTransit, map[string]string{"C1": "06:50:34", "C2": "06:54:10", "C3": "09:07:25", "C4": "09:11:01"}, nil, 5 * time.Second},
		{"2025-05-20", 4, jupiter.GalileanPhenomenonShadowTransit, map[string]string{"C1": "18:57:19", "C4": "20:40:39"}, []string{"C2", "C3"}, 75 * time.Second},
		{"2025-06-06", 4, jupiter.GalileanPhenomenonTransit, map[string]string{"C1": "10:12:03", "C4": "11:54:03"}, []string{"C2", "C3"}, 45 * time.Second},
		{"2025-06-14", 4, jupiter.GalileanPhenomenonEclipse, map[string]string{"P1": "20:04:55", "U1": "20:22:56", "U4": "22:01:01", "P4": "22:19:02"}, []string{"U2", "U3"}, 45 * time.Second},
		{"2025-04-08", 4, jupiter.GalileanPhenomenonEclipse, map[string]string{"P1": "20:19:47", "P4": "20:54:41"}, []string{"U1", "U2", "U3", "U4"}, 25 * time.Second},
	} {
		t.Run(fixture.day, func(t *testing.T) {
			seed, err := time.Parse("2006-01-02", fixture.day)
			if err != nil {
				t.Fatal(err)
			}
			event := jupiter.NextGalileanPhenomenonContactEvent(seed, fixture.satellite, fixture.kind)
			document := phenomenonDocument(t, event, GalileanPhenomenonSVGOptions{Language: "en"})
			for code, clock := range fixture.contacts {
				want, err := time.Parse("2006-01-02 15:04:05", fixture.day+" "+clock)
				if err != nil {
					t.Fatal(err)
				}
				got, ok := mutualContactTime(t, document, code, event.Greatest)
				if delta := got.Sub(want); !ok || delta < -fixture.tolerance || delta > fixture.tolerance {
					t.Errorf("%s: got %v, reference %v", code, got.UTC(), want)
				}
			}
			for _, code := range fixture.absent {
				if _, ok := mutualContactTime(t, document, code, event.Greatest); ok {
					t.Errorf("unexpected %s", code)
				}
			}
		})
	}
}

func TestGalileanPhenomenonSVGConsistencyBaseline(t *testing.T) {
	if go116386() {
		t.Skip("Go 1.16/386 floating-point text differs; render-only test covers this build")
	}
	previous := basic.GetEphemerisModel()
	defer basic.SetEphemerisModel(previous)
	for _, fixture := range phenomenonFixtures() {
		for _, model := range []basic.EphemerisModel{basic.EphemerisLegacy, basic.EphemerisSemianalytic} {
			basic.SetEphemerisModel(model)
			document := phenomenonDocument(t, phenomenonFixtureEvent(t, fixture), GalileanPhenomenonSVGOptions{Language: fixture.language})
			sum := md5.Sum([]byte(document))
			digest := hex.EncodeToString(sum[:])
			if digest != fixture.digest[model] || len(document) != fixture.length[model] {
				t.Errorf("%s model=%d md5=%s bytes=%d", fixture.name, model, digest, len(document))
			}
		}
	}
}

func TestGalileanPhenomenonSVGSyntheticRendersOnGo116386(t *testing.T) {
	if !go116386() {
		t.Skip("Go 1.16/386 only")
	}
	for _, fixture := range phenomenonFixtures() {
		document := phenomenonDocument(t, phenomenonFixtureEvent(t, fixture), GalileanPhenomenonSVGOptions{Language: fixture.language})
		if _, _, err := svgElements(document); err != nil || !svgFinite(document) {
			t.Fatal(fmt.Sprint("invalid rendered SVG ", fixture.name))
		}
	}
}

// 木影食：二触/三触时卫星盘与木星本影截面内切（太阳视平面），独立按卫星日心矢量与木星日心矢量求。
func TestGalileanPhenomenonSVGConsistencyEclipseTangency(t *testing.T) {
	for _, satellite := range []int{1, 4} {
		event := jupiter.NextGalileanPhenomenonContactEvent(time.Date(2026, 5, 14, 0, 0, 0, 0, time.UTC), satellite, jupiter.GalileanPhenomenonEclipse)
		byID := mutualByID(t, phenomenonDocument(t, event, GalileanPhenomenonSVGOptions{}))
		umbra := byID["jupiter-umbra"]
		body := byID["phenomenon-body"]
		if svgNumber(umbra, "rx") <= 0 || svgNumber(body, "r") <= 0 {
			t.Fatalf("satellite %d umbra/body missing", satellite)
		}
		ratio := svgNumber(body, "r") / svgNumber(umbra, "rx")
		jde := basic.UTC2TT(basic.Date2JD(event.Greatest.UTC()))
		diagram, _ := basic.JupiterGalileanPhenomenonDiagramWithInstance(jde, satellite, nil)
		radii := [...]float64{1829.4, 1562.6, 2631.2, 2410.3}
		want := radii[satellite-1] / 71492 / diagram.UmbraScale
		if math.Abs(ratio-want) > 2e-3*want {
			t.Fatalf("satellite %d disk/umbra ratio %.5f, independent %.5f", satellite, ratio, want)
		}
	}
}
