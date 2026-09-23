package svg

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/starainrt/astro/internal/svgchart"
)

// 页脚排版契约：任何文字的下沿都要离画布底边留出余量，避免末行贴边或被裁掉。
const occultationChartBottomMarginMinimum = 6.0

func occultationBottomMarginViolations(t *testing.T, diagram string, height int) []string {
	t.Helper()
	violations := []string{}
	for _, text := range occultationSVGTexts(t, diagram) {
		if text.value == "" {
			continue
		}
		_, below := svgchart.EstimatedTextExtents(text.fontSize)
		if margin := float64(height) - (text.y + below); margin < occultationChartBottomMarginMinimum {
			violations = append(violations, fmt.Sprintf("%q 距底边 %.2f px（下限 %.1f）", text.value, margin, occultationChartBottomMarginMinimum))
		}
	}
	return violations
}

func TestOccultationChartsKeepBottomMargin(t *testing.T) {
	long := footerFitLongText("margin ")
	starPath := occultationTestStarPath(t)
	planetPath := occultationTestSaturnPath(t)
	cases := []struct {
		name     string
		height   int
		document string
	}{}
	add := func(name string, height int, document string, err error) {
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		cases = append(cases, struct {
			name     string
			height   int
			document string
		}{name, height, document})
	}
	for _, size := range [][2]int{{640, 480}, {920, 720}} {
		options := footerFitStarTexts(long)
		options.Width, options.Height = size[0], size[1]
		diagram, err := StarOccultationPathSVG(starPath, options)
		add(fmt.Sprintf("恒星全球 %dx%d", size[0], size[1]), size[1], diagram, err)
		planetOptions := PlanetOccultationSVGOptions(options)
		planetDiagram, err := PlanetOccultationPathSVG(planetPath, planetOptions)
		add(fmt.Sprintf("行星全球 %dx%d", size[0], size[1]), size[1], planetDiagram, err)
	}
	local, err := LocalStarOccultationSVG(localHR4799Occultation(t), hr4799StarCoordinate(), LocalStarOccultationSVGOptions{
		Width: 640, Height: 520, Title: long, DirectionText: long, FooterNote: long,
	})
	add("站心恒星", 520, local, err)
	localPlanet, err := LocalPlanetOccultationSVG(localSaturnOccultation(t), LocalPlanetOccultationSVGOptions{
		Width: 760, Height: 600, Title: long, DirectionText: long, FooterNote: long,
	})
	add("站心行星", 600, localPlanet, err)
	detailed, err := StarOccultationDetailedSVG(starPath, hr4799StarCoordinate(), OccultationDetailedSVGOptions{
		Width: 700, Height: 990, Title: long, FooterNote: long,
	})
	add("恒星详图", 990, detailed, err)

	for _, tc := range cases {
		if violations := occultationBottomMarginViolations(t, tc.document, tc.height); len(violations) > 0 {
			t.Errorf("%s：%d 行文字贴到画布底边，首条 %s", tc.name, len(violations), violations[0])
		}
	}
}

// 页脚时标声明契约：超长自定义说明可以吃掉说明文字，但不能吃掉时标声明。
func TestOccultationChartsKeepTimeScaleDeclaration(t *testing.T) {
	long := footerFitLongText("scale ")
	cst := time.FixedZone("CST", 8*3600)
	starPath := occultationTestStarPath(t)
	planetPath := occultationTestSaturnPath(t)
	options := footerFitStarTexts(long)
	options.Width, options.Height = 920, 720
	options.Location = cst
	global, err := StarOccultationPathSVG(starPath, options)
	if err != nil {
		t.Fatalf("恒星全球：%v", err)
	}
	planetOptions := PlanetOccultationSVGOptions(options)
	planet, err := PlanetOccultationPathSVG(planetPath, planetOptions)
	if err != nil {
		t.Fatalf("行星全球：%v", err)
	}
	local, err := LocalStarOccultationSVG(localHR4799Occultation(t), hr4799StarCoordinate(), LocalStarOccultationSVGOptions{
		Width: 920, Height: 720, Location: cst, FooterNote: long,
	})
	if err != nil {
		t.Fatalf("站心恒星：%v", err)
	}
	for name, document := range map[string]string{"恒星全球": global, "行星全球": planet, "站心恒星": local} {
		if !strings.Contains(document, "显示时区") && !strings.Contains(document, "图中时刻为 UTC") {
			t.Errorf("%s：超长说明把时标声明挤掉了", name)
		}
	}
}
