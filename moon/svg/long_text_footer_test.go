package svg

import (
	"strings"
	"testing"

	"github.com/starainrt/astro/internal/svgchart"
)

// 调用方给图题、摘要、地图标题与页脚传超长文本时，产物必须留在画布内并在截断处给出省略号；
// 判定口径与 internal/svgchart 的 EstimatedTextWidth/EstimatedTextExtents 一致。

func footerFitLongText(prefix string) string {
	piece := prefix + "long custom text 自定义长文本，用于验证折行与截断；"
	value := ""
	for len([]rune(value)) < 640 {
		value += piece
	}
	return value
}

func footerFitViolations(t *testing.T, diagram string, width, height int) []string {
	t.Helper()
	violations := []string{}
	for _, text := range occultationSVGTexts(t, diagram) {
		textWidth := svgchart.EstimatedTextWidth(text.value, text.fontSize)
		left := text.x
		switch text.anchor {
		case "middle":
			left = text.x - textWidth/2
		case "end":
			left = text.x - textWidth
		}
		above, below := svgchart.EstimatedTextExtents(text.fontSize)
		if text.x < 0 || text.x > float64(width) || text.y < 0 || text.y > float64(height) {
			violations = append(violations, text.value)
			continue
		}
		if left < 0 || left+textWidth > float64(width) || text.y-above < 0 || text.y+below > float64(height) {
			violations = append(violations, text.value)
		}
	}
	return violations
}

func footerFitAssert(t *testing.T, name string, diagram string, width, height int) {
	t.Helper()
	if violations := footerFitViolations(t, diagram, width, height); len(violations) > 0 {
		t.Fatalf("%s: %d texts outside the %dx%d canvas, first = %q",
			name, len(violations), width, height, violations[0])
	}
	if !strings.Contains(diagram, "…") {
		t.Fatalf("%s: truncated custom text is not marked", name)
	}
}

func footerFitStarTexts(long string) StarOccultationSVGOptions {
	return StarOccultationSVGOptions{
		Title: long, SummaryText: long, GreatestText: long,
		MapTitle: long, ContactsTitle: long, FooterNote: long,
	}
}

func TestStarOccultationPathSVGCustomLongTextStaysInsideCanvas(t *testing.T) {
	long := footerFitLongText("star ")
	if runes := len([]rune(long)); runes < 300 || runes > 800 {
		t.Fatalf("long text runes = %d, want 300..800", runes)
	}
	path := occultationTestStarPath(t)
	for _, size := range [][2]int{{640, 480}, {920, 720}, {1200, 800}} {
		options := footerFitStarTexts(long)
		options.Width, options.Height = size[0], size[1]
		diagram, err := StarOccultationPathSVG(path, options)
		if err != nil {
			t.Fatalf("%dx%d: %v", size[0], size[1], err)
		}
		footerFitAssert(t, "star-global", diagram, size[0], size[1])
	}
}

func TestPlanetOccultationPathSVGCustomLongTextStaysInsideCanvas(t *testing.T) {
	long := footerFitLongText("planet ")
	path := occultationTestSaturnPath(t)
	for _, size := range [][2]int{{640, 480}, {920, 720}, {1200, 800}} {
		options := PlanetOccultationSVGOptions(footerFitStarTexts(long))
		options.Width, options.Height = size[0], size[1]
		diagram, err := PlanetOccultationPathSVG(path, options)
		if err != nil {
			t.Fatalf("%dx%d: %v", size[0], size[1], err)
		}
		footerFitAssert(t, "planet-global", diagram, size[0], size[1])
	}
}

func TestOccultationDetailedSVGCustomLongTextStaysInsideCanvas(t *testing.T) {
	long := footerFitLongText("detailed ")
	star := occultationTestStarPath(t)
	planetPath := occultationTestSaturnPath(t)
	for _, size := range [][2]int{{700, 990}, {1000, 1414}} {
		diagram, err := StarOccultationDetailedSVG(star, hr4799StarCoordinate(), OccultationDetailedSVGOptions{
			Width: size[0], Height: size[1], Title: long, FooterNote: long,
		})
		if err != nil {
			t.Fatalf("star %dx%d: %v", size[0], size[1], err)
		}
		footerFitAssert(t, "star-detailed", diagram, size[0], size[1])
		diagram, err = PlanetOccultationDetailedSVG(planetPath, OccultationDetailedSVGOptions{
			Width: size[0], Height: size[1], Title: long, FooterNote: long,
		})
		if err != nil {
			t.Fatalf("planet %dx%d: %v", size[0], size[1], err)
		}
		footerFitAssert(t, "planet-detailed", diagram, size[0], size[1])
	}
}

func TestLocalStarOccultationSVGCustomLongTextStaysInsideCanvas(t *testing.T) {
	long := footerFitLongText("local ")
	info := localHR4799Occultation(t)
	for _, size := range [][2]int{{640, 520}, {920, 720}} {
		diagram, err := LocalStarOccultationSVG(info, hr4799StarCoordinate(), LocalStarOccultationSVGOptions{
			Width: size[0], Height: size[1], Title: long, SummaryText: long, GreatestText: long,
			OverviewTitle: long, PhasePanelsTitle: long, ContactsTitle: long,
			DirectionText: long, FooterNote: long,
		})
		if err != nil {
			t.Fatalf("%dx%d: %v", size[0], size[1], err)
		}
		footerFitAssert(t, "star-local", diagram, size[0], size[1])
	}
}

func TestLocalPlanetOccultationSVGCustomLongTextStaysInsideCanvas(t *testing.T) {
	long := footerFitLongText("local-planet ")
	info := localSaturnOccultation(t)
	for _, size := range [][2]int{{760, 600}, {920, 720}} {
		diagram, err := LocalPlanetOccultationSVG(info, LocalPlanetOccultationSVGOptions{
			Width: size[0], Height: size[1], Title: long, SummaryText: long, GreatestText: long,
			OverviewTitle: long, PhasePanelsTitle: long, ContactsTitle: long,
			DirectionText: long, FooterNote: long,
		})
		if err != nil {
			t.Fatalf("%dx%d: %v", size[0], size[1], err)
		}
		footerFitAssert(t, "planet-local", diagram, size[0], size[1])
	}
}

// 自动生成的默认文案不得进入截断分支：这些产物在改动前后逐字节一致，出现省略号即说明分流写错。
func TestOccultationSVGDefaultTextIsNotTruncated(t *testing.T) {
	star := occultationTestStarPath(t)
	planetPath := occultationTestSaturnPath(t)
	for _, language := range []string{"zh", "en"} {
		documents := []string{}
		starGlobal, err := StarOccultationPathSVG(star, StarOccultationSVGOptions{Language: language})
		if err != nil {
			t.Fatalf("star global %s: %v", language, err)
		}
		documents = append(documents, starGlobal)
		planetGlobal, err := PlanetOccultationPathSVG(planetPath, PlanetOccultationSVGOptions{Language: language})
		if err != nil {
			t.Fatalf("planet global %s: %v", language, err)
		}
		documents = append(documents, planetGlobal)
		detailed, err := StarOccultationDetailedSVG(star, hr4799StarCoordinate(), OccultationDetailedSVGOptions{Language: language})
		if err != nil {
			t.Fatalf("star detailed %s: %v", language, err)
		}
		documents = append(documents, detailed)
		for _, document := range documents {
			if strings.Contains(document, "…") {
				t.Fatalf("%s: default text must stay unmarked", language)
			}
		}
	}
}
