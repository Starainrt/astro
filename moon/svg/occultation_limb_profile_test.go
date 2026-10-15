package svg

// 精密月掩剖面图：每段遮挡都要出现在接触表里，无样本或非法选项时返回错误。

import (
	"errors"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/starainrt/astro"
	"github.com/starainrt/astro/internal/svgchart"
	"github.com/starainrt/astro/moon"
	"github.com/starainrt/astro/moon/limb"
)

var limbProfileTestStar = moon.StarCoordinate{
	ID: "Antares", RA: 247.3516666666667, Dec: -26.431944444444444,
	Epoch: time.Date(2000, time.January, 1, 12, 0, 0, 0, time.UTC), Frame: moon.CoordinateFrameJ2000,
	ProperMotionRACosDecMasPerYear: -10, ProperMotionDecMasPerYear: -20, ParallaxMas: 24,
}

func limbProfileTestGraze(t *testing.T) moon.StarOccultationLimbProfileResult {
	t.Helper()
	start := time.Date(2026, time.February, 10, 12, 0, 0, 0, time.UTC)
	events, err := moon.FindStarOccultations(start, start.Add(36*time.Hour), limbProfileTestStar,
		41.5907, -55.07036, 0, moon.OccultationSearchOptions{LunarLimb: limb.Default()})
	if err != nil || len(events) != 1 {
		t.Fatalf("FindStarOccultations() = %d events, %v; want one", len(events), err)
	}
	profile, err := moon.StarOccultationLimbProfile(events[0], limbProfileTestStar, limb.Default(), moon.StarOccultationLimbProfileOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(profile.Hidden) < 2 {
		t.Fatalf("hidden intervals = %d, want a multi-contact graze", len(profile.Hidden))
	}
	return profile
}

func TestStarOccultationLimbProfileSVGListsEveryContact(t *testing.T) {
	profile := limbProfileTestGraze(t)
	for _, language := range []string{"", "en"} {
		diagram, err := StarOccultationLimbProfileSVG(profile, StarOccultationLimbProfileSVGOptions{Language: language, Location: time.UTC})
		if err != nil {
			t.Fatal(err)
		}
		if !strings.HasPrefix(diagram, "<svg") || !strings.HasSuffix(diagram, "</svg>") {
			t.Fatal("output is not a complete SVG document")
		}
		for _, interval := range profile.Hidden {
			for _, instant := range []time.Time{interval.Start, interval.End} {
				if label := instant.UTC().Format("15:04:05.00"); !strings.Contains(diagram, label) {
					t.Errorf("language %q: contact %s missing", language, label)
				}
			}
		}
		if got := strings.Count(diagram, `fill-opacity="0.16"`); got < len(profile.Hidden) {
			t.Errorf("language %q: %d shaded rects, want at least %d", language, got, len(profile.Hidden))
		}
	}
}

func TestStarOccultationLimbProfileSVGDeclaresUT1(t *testing.T) {
	profile := limbProfileTestGraze(t)
	diagram, err := StarOccultationLimbProfileSVG(profile, StarOccultationLimbProfileSVGOptions{TimeScale: astro.TimeScaleUT1})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(diagram, "UT1") {
		t.Error("UT1 chart does not declare the time scale")
	}
	if _, err := StarOccultationLimbProfileSVG(profile, StarOccultationLimbProfileSVGOptions{TimeScale: astro.TimeScaleUT1, Location: time.FixedZone("x", 3600)}); err == nil {
		t.Error("UT1 with a non-UTC location accepted")
	}
}

func TestStarOccultationLimbProfileSVGRejectsBadInput(t *testing.T) {
	if _, err := StarOccultationLimbProfileSVG(moon.StarOccultationLimbProfileResult{}, StarOccultationLimbProfileSVGOptions{}); err == nil {
		t.Error("empty profile accepted")
	}
	profile := limbProfileTestGraze(t)
	if _, err := StarOccultationLimbProfileSVG(profile, StarOccultationLimbProfileSVGOptions{Width: 100}); err == nil {
		t.Error("tiny width accepted")
	}
}

func TestStarOccultationLimbProfileSVGTextStaysInsideFrame(t *testing.T) {
	profile := limbProfileTestGraze(t)
	text := regexp.MustCompile(`<text x="([0-9.]+)"[^>]*font-size="([0-9]+)"([^>]*)>([^<]*)</text>`)
	for _, language := range []string{"zh", "en"} {
		diagram, err := StarOccultationLimbProfileSVG(profile, StarOccultationLimbProfileSVGOptions{Language: language, Location: time.UTC})
		if err != nil {
			t.Fatal(err)
		}
		for _, match := range text.FindAllStringSubmatch(diagram, -1) {
			x, _ := strconv.ParseFloat(match[1], 64)
			size, _ := strconv.ParseFloat(match[2], 64)
			width := svgchart.EstimatedTextWidth(match[4], size)
			low, high := x, x+width
			if strings.Contains(match[3], "middle") {
				low, high = x-width/2, x+width/2
			} else if strings.Contains(match[3], `"end"`) {
				low, high = x-width, x
			}
			if low < 22 || high > limbProfileSVGDefaultWidth-22 {
				t.Errorf("language %q: text [%.0f, %.0f] leaves the frame: %s", language, low, high, match[4])
			}
		}
	}
}

func TestStarOccultationLimbProfileSVGDrawsPhasedMoonAndNumberedContacts(t *testing.T) {
	profile := limbProfileTestGraze(t)
	diagram, err := StarOccultationLimbProfileSVG(profile, StarOccultationLimbProfileSVGOptions{Location: time.UTC})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(diagram, `href="#le-moon"`) || !strings.Contains(diagram, `class="limb-profile-moon"`) {
		t.Fatal("inset does not use the textured, phased Moon")
	}
	if strings.Contains(diagram, `data-illumination="1.000000"`) {
		t.Error("inset ignores the lunar phase")
	}
	for number := 1; number <= 2*len(profile.Hidden); number++ {
		if !strings.Contains(diagram, `font-weight="700" text-anchor="middle">`+strconv.Itoa(number)+`</text>`) {
			t.Errorf("contact %d is not numbered on the profile", number)
		}
	}
}

func limbProfileTestAt(t *testing.T, latitude float64) moon.StarOccultationLimbProfileResult {
	t.Helper()
	info := moon.StarOccultationInfo{Greatest: time.Date(2026, time.February, 11, 3, 47, 15, 0, time.UTC),
		Observer: moon.Observer{Longitude: 41.5907, Latitude: latitude}}
	if latitude < -55.5 {
		start := time.Date(2026, time.February, 10, 12, 0, 0, 0, time.UTC)
		events, err := moon.FindStarOccultations(start, start.Add(36*time.Hour), limbProfileTestStar, 41.5907, latitude, 0,
			moon.OccultationSearchOptions{LunarLimb: limb.Default()})
		if err != nil || len(events) != 1 {
			t.Fatalf("FindStarOccultations() = %d events, %v", len(events), err)
		}
		info = events[0]
	}
	profile, err := moon.StarOccultationLimbProfile(info, limbProfileTestStar, limb.Default(), moon.StarOccultationLimbProfileOptions{})
	if err != nil {
		t.Fatal(err)
	}
	return profile
}

type limbProfileTestText struct {
	x, y, width float64
	text        string
}

func limbProfileTestTexts(diagram string) []limbProfileTestText {
	pattern := regexp.MustCompile(`<text x="([0-9.]+)" y="([0-9.]+)"[^>]*font-size="([0-9]+)"([^>]*)>([^<]*)</text>`)
	var texts []limbProfileTestText
	for _, match := range pattern.FindAllStringSubmatch(diagram, -1) {
		x, _ := strconv.ParseFloat(match[1], 64)
		y, _ := strconv.ParseFloat(match[2], 64)
		size, _ := strconv.ParseFloat(match[3], 64)
		width := svgchart.EstimatedTextWidth(match[5], size)
		if strings.Contains(match[4], "middle") {
			x -= width / 2
		} else if strings.Contains(match[4], `"end"`) {
			x -= width
		}
		texts = append(texts, limbProfileTestText{x, y, width, match[5]})
	}
	return texts
}

func TestStarOccultationLimbProfileSVGDeepCase(t *testing.T) {
	profile := limbProfileTestAt(t, -57.0)
	if len(profile.Segments) != 2 || len(profile.Hidden) != 1 || profile.Occultation.Type != moon.OccultationTotal {
		t.Fatalf("deep profile: %d segments, %d intervals, type %v", len(profile.Segments), len(profile.Hidden), profile.Occultation.Type)
	}
	diagram, err := StarOccultationLimbProfileSVG(profile, StarOccultationLimbProfileSVGOptions{Location: time.UTC})
	if err != nil {
		t.Fatal(err)
	}
	texts := limbProfileTestTexts(diagram)
	for i, a := range texts {
		for _, b := range texts[i+1:] {
			if math.Abs(a.y-b.y) < 9 && a.x < b.x+b.width && b.x < a.x+a.width {
				t.Errorf("texts overlap: %q at (%.0f,%.0f) and %q at (%.0f,%.0f)", a.text, a.x, a.y, b.text, b.x, b.y)
			}
		}
	}
	immersion := profile.Hidden[0].Start
	want := ""
	for _, sample := range profile.Segments[0] {
		if !sample.Time.Before(immersion) {
			want = fmt.Sprintf("PA %.1f°", sample.PositionAngleDeg)
			break
		}
	}
	if !strings.Contains(diagram, want) {
		t.Errorf("direction text does not use the immersion position angle %s", want)
	}
	if !strings.Contains(diagram, "横向放大约") {
		t.Error("compressed vertical scale is not labelled")
	}
}

func TestStarOccultationLimbProfileSVGNotOcculted(t *testing.T) {
	profile := limbProfileTestAt(t, -55.07036+0.8/111.2)
	if len(profile.Hidden) != 0 || profile.Occultation.Type != moon.OccultationNone {
		t.Fatalf("near-miss profile: %d intervals, type %v", len(profile.Hidden), profile.Occultation.Type)
	}
	diagram, err := StarOccultationLimbProfileSVG(profile, StarOccultationLimbProfileSVGOptions{Location: time.UTC})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(diagram, "未发生遮挡") {
		t.Error("not-occulted chart does not say so")
	}
	if _, err := StarOccultationLimbProfileSVG(moon.StarOccultationLimbProfileResult{}, StarOccultationLimbProfileSVGOptions{}); !errors.Is(err, ErrEmptyStarOccultationLimbProfile) {
		t.Errorf("empty profile error = %v, want ErrEmptyStarOccultationLimbProfile", err)
	}
}

func TestLimbProfileHelpers(t *testing.T) {
	kathmandu := time.FixedZone("UTC+5:45", 5*3600+45*60)
	start := time.Date(2026, 2, 11, 3, 3, 7, 0, time.UTC)
	for _, step := range []time.Duration{10 * time.Minute, time.Hour} {
		ticks := limbProfileTicks(start, start.Add(3*time.Hour), step, kathmandu)
		if len(ticks) == 0 || ticks[0].Before(start) {
			t.Fatalf("step %v: ticks %v", step, ticks)
		}
		for _, tick := range ticks {
			local := tick.In(kathmandu)
			if local.Sub(time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, kathmandu))%step != 0 {
				t.Errorf("step %v: tick %v not aligned in the display zone", step, local)
			}
		}
	}
	// 豪勋爵岛夏令时只拨 30 分钟：切换后 20 分钟刻度仍须落在当地整 20 分上。
	lordHowe, err := time.LoadLocation("Australia/Lord_Howe")
	if err == nil {
		switchStart := time.Date(2026, 4, 4, 14, 0, 0, 0, time.UTC)
		for _, tick := range limbProfileTicks(switchStart, switchStart.Add(2*time.Hour), 20*time.Minute, lordHowe) {
			if local := tick.In(lordHowe); local.Minute()%20 != 0 || local.Second() != 0 {
				t.Errorf("Lord Howe tick %v is not on a local 20-minute mark", local)
			}
		}
	}
	if got := limbProfileFormatAngle(359.96); got != "0.0°" {
		t.Errorf("359.96 formats as %s", got)
	}
}

func TestStarOccultationLimbProfileSVGRejectsMalformedProfiles(t *testing.T) {
	good := limbProfileTestGraze(t)
	clone := func() moon.StarOccultationLimbProfileResult {
		profile := good
		profile.Segments = make([][]moon.StarOccultationLimbSample, len(good.Segments))
		for index, segment := range good.Segments {
			profile.Segments[index] = append([]moon.StarOccultationLimbSample(nil), segment...)
		}
		profile.Hidden = append([]moon.OccultationInterval(nil), good.Hidden...)
		return profile
	}
	cases := map[string]func(*moon.StarOccultationLimbProfileResult){
		"NaN position angle": func(p *moon.StarOccultationLimbProfileResult) { p.Segments[0][3].PositionAngleDeg = math.NaN() },
		"NaN mean radius":    func(p *moon.StarOccultationLimbProfileResult) { p.Segments[0][3].MeanRadiusArcsec = math.NaN() },
		"Inf separation":     func(p *moon.StarOccultationLimbProfileResult) { p.Segments[0][3].SeparationArcsec = math.Inf(1) },
		"five segments": func(p *moon.StarOccultationLimbProfileResult) {
			for len(p.Segments) < 5 {
				p.Segments = append(p.Segments, p.Segments[0])
			}
		},
		"time going back": func(p *moon.StarOccultationLimbProfileResult) {
			p.Segments[0][4].Time = p.Segments[0][2].Time
		},
		"reversed interval": func(p *moon.StarOccultationLimbProfileResult) {
			p.Hidden[0].Start, p.Hidden[0].End = p.Hidden[0].End, p.Hidden[0].Start
		},
	}
	for name, mutate := range cases {
		profile := clone()
		mutate(&profile)
		diagram, err := StarOccultationLimbProfileSVG(profile, StarOccultationLimbProfileSVGOptions{Location: time.UTC})
		if !errors.Is(err, ErrInvalidStarOccultationLimbProfile) {
			t.Errorf("%s: err = %v, want ErrInvalidStarOccultationLimbProfile", name, err)
		}
		if diagram != "" {
			t.Errorf("%s: produced output", name)
		}
	}
}

func TestStarOccultationLimbProfileSVGLabelsNeverOverlap(t *testing.T) {
	base := limbProfileTestGraze(t)
	start := base.Segments[0][0].Time
	end := base.Segments[0][len(base.Segments[0])-1].Time
	// 8 段各 8 s 的密集遮挡，加上不同起点相位，覆盖编号与首刻度的碰撞。
	for phase := 0; phase < 40; phase++ {
		profile := base
		shift := time.Duration(phase) * 1700 * time.Millisecond
		profile.Segments = [][]moon.StarOccultationLimbSample{append([]moon.StarOccultationLimbSample(nil), base.Segments[0]...)}
		for index := range profile.Segments[0] {
			profile.Segments[0][index].Time = profile.Segments[0][index].Time.Add(shift)
		}
		profile.Hidden = nil
		for k := 0; k < 8; k++ {
			at := start.Add(shift).Add(end.Sub(start)/3 + time.Duration(k)*12*time.Second)
			profile.Hidden = append(profile.Hidden, moon.OccultationInterval{Start: at, End: at.Add(8 * time.Second)})
		}
		profile.Occultation.Greatest = profile.Occultation.Greatest.Add(shift)
		diagram, err := StarOccultationLimbProfileSVG(profile, StarOccultationLimbProfileSVGOptions{Location: time.UTC})
		if err != nil {
			t.Fatal(err)
		}
		texts := limbProfileTestTexts(diagram)
		for i, a := range texts {
			for _, b := range texts[i+1:] {
				if math.Abs(a.y-b.y) < 8 && a.x < b.x+b.width-0.5 && b.x < a.x+a.width-0.5 {
					t.Fatalf("phase %d: %q at (%.1f,%.0f) overlaps %q at (%.1f,%.0f)", phase, a.text, a.x, a.y, b.text, b.x, b.y)
				}
			}
		}
	}
}

func TestStarOccultationLimbProfileSVGContactNumbersStayInsideFrame(t *testing.T) {
	base := limbProfileTestGraze(t)
	start := base.Segments[0][0].Time
	end := base.Segments[0][len(base.Segments[0])-1].Time
	const frameLeft, frameRight = 70.0, limbProfileSVGDefaultWidth - 260.0 - 48
	for _, count := range []int{3, 6, 8, 12} {
		for _, gap := range []time.Duration{2 * time.Second, 8 * time.Second} {
			profile := base
			profile.Hidden = nil
			for k := count - 1; k >= 0; k-- {
				at := end.Add(-time.Duration(k+1) * 2 * gap)
				profile.Hidden = append(profile.Hidden, moon.OccultationInterval{Start: at, End: at.Add(gap)})
			}
			if profile.Hidden[0].Start.Before(start) {
				t.Fatalf("%d intervals do not fit the profile span", count)
			}
			diagram, err := StarOccultationLimbProfileSVG(profile, StarOccultationLimbProfileSVGOptions{Location: time.UTC})
			if err != nil {
				t.Fatal(err)
			}
			var numbers []limbProfileTestText
			for _, text := range limbProfileTestTexts(diagram) {
				if _, err := strconv.Atoi(text.text); err == nil && text.y > 110 && text.y < 136 {
					numbers = append(numbers, text)
				}
			}
			if len(numbers) != 2*count {
				t.Fatalf("%d intervals, gap %v: %d contact numbers, want %d", count, gap, len(numbers), 2*count)
			}
			for i, a := range numbers {
				if a.x < frameLeft-0.5 || a.x+a.width > frameRight+0.5 {
					t.Errorf("%d intervals, gap %v: number %s spans %.1f–%.1f, frame %.0f–%.0f", count, gap, a.text, a.x, a.x+a.width, frameLeft, frameRight)
				}
				for _, b := range numbers[i+1:] {
					if a.y == b.y && a.x < b.x+b.width-0.5 && b.x < a.x+a.width-0.5 {
						t.Errorf("%d intervals, gap %v: numbers %s and %s overlap", count, gap, a.text, b.text)
					}
				}
			}
		}
	}
}
