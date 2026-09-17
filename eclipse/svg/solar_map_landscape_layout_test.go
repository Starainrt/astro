package svg

import (
	"regexp"
	"strconv"
	"testing"
	"time"

	eclipsecore "github.com/starainrt/astro/eclipse"
	"github.com/starainrt/astro/internal/svgmap"
)

var solarEclipsePanelRectPattern = regexp.MustCompile(
	`class="solar-detailed-panel"><rect x="([-0-9.]+)" y="([-0-9.]+)" width="([-0-9.]+)" height="([-0-9.]+)"`)

// 横版日食版式必须容下地图框与数据网格：地图框与面板矩形不得相交，面板行距不得塌陷，
// 低于文档下限的画布按文档回落到 960x640。
func TestSolarEclipseMapLandscapeLayoutKeepsPanelsClearOfMap(t *testing.T) {
	date := time.Date(2024, time.April, 8, 0, 0, 0, 0, time.UTC)
	info, ok := eclipsecore.SolarEclipseOnDateNASABulletinSplitK(date)
	if !ok {
		t.Fatal("missing 2024-04-08 eclipse")
	}
	center := svgmap.GeoPoint{Longitude: info.GreatestLongitude, Latitude: info.GreatestLatitude}
	for _, test := range []struct {
		name          string
		width, height int
	}{
		{name: "documented minimum", width: 800, height: 560},
		{name: "minimum landscape", width: 800, height: 600},
		{name: "default", width: 960, height: 640},
		{name: "wide", width: 1200, height: 800},
	} {
		t.Run(test.name, func(t *testing.T) {
			options := SolarEclipseMapSVGOptions{Width: test.width, Height: test.height, Location: time.UTC, PartialStep: 10 * time.Minute}
			normalized := normalizeSolarEclipseMapSVGOptions(date, options)
			if normalized.Width != test.width || normalized.Height != test.height {
				t.Fatalf("canvas %dx%d was resized to %dx%d", test.width, test.height, normalized.Width, normalized.Height)
			}
			layout := solarEclipseMapLayoutFor(normalized, svgmap.ProjectionEquirectangular, center)
			if !layout.landscape {
				t.Fatal("expected the landscape composition")
			}
			frameRight := layout.frame.X + layout.frame.Width
			if frameRight > layout.gridX {
				t.Fatalf("map frame right edge %.3f overlaps the panel grid at x=%.3f by %.3f px",
					frameRight, layout.gridX, frameRight-layout.gridX)
			}
			// 数据块最多六行，行高按 (cellHeight-30)/行数 等分。
			rowSpacing := (layout.gridCellHeight - 30) / 6
			if rowSpacing < 8 {
				t.Fatalf("panel row spacing = %.2f px, want at least 8", rowSpacing)
			}

			rendered, ok := SolarEclipseMapSVG(date, options)
			if !ok {
				t.Fatal("missing map")
			}
			rects := solarEclipsePanelRectPattern.FindAllStringSubmatch(rendered, -1)
			if len(rects) == 0 {
				t.Fatal("landscape map renders no data panels")
			}
			for _, rect := range rects {
				values := make([]float64, 4)
				for index := range values {
					parsed, err := strconv.ParseFloat(rect[index+1], 64)
					if err != nil {
						t.Fatalf("bad panel rect %v: %v", rect, err)
					}
					values[index] = parsed
				}
				x, y, width, height := values[0], values[1], values[2], values[3]
				if x < 0 || y < 0 || x+width > float64(test.width) || y+height > float64(test.height) {
					t.Fatalf("panel rect (%.1f,%.1f,%.1f,%.1f) leaves the %dx%d canvas", x, y, width, height, test.width, test.height)
				}
				if x < frameRight && layout.frame.X < x+width &&
					y < layout.frame.Y+layout.frame.Height && layout.frame.Y < y+height {
					t.Fatalf("panel rect (%.1f,%.1f,%.1f,%.1f) intersects the map frame", x, y, width, height)
				}
			}
		})
	}

	// 低于下限的那一维单独回落到默认值，另一维保持请求值。
	for _, test := range []struct {
		name                  string
		width, height         int
		wantWidth, wantHeight int
	}{
		{name: "legacy minimum", width: 640, height: 420, wantWidth: 960, wantHeight: 640},
		{name: "narrow", width: 700, height: 600, wantWidth: 960, wantHeight: 600},
		{name: "short", width: 1000, height: 500, wantWidth: 1000, wantHeight: 640},
	} {
		t.Run("fallback/"+test.name, func(t *testing.T) {
			normalized := normalizeSolarEclipseMapSVGOptions(date, SolarEclipseMapSVGOptions{Width: test.width, Height: test.height})
			if normalized.Width != test.wantWidth || normalized.Height != test.wantHeight {
				t.Fatalf("%dx%d normalized to %dx%d, want %dx%d",
					test.width, test.height, normalized.Width, normalized.Height,
					test.wantWidth, test.wantHeight)
			}
		})
	}
}

// 画布内闭合：文档下限的横版画布与同宽度的竖版画布都不得让 text/rect 越出画布。
func TestSolarEclipseMapKeepsEveryElementInCanvas(t *testing.T) {
	date := time.Date(2024, time.April, 8, 0, 0, 0, 0, time.UTC)
	for _, size := range [][2]int{{800, 560}, {800, 801}} {
		rendered, ok := SolarEclipseMapSVG(date, SolarEclipseMapSVGOptions{
			Width: size[0], Height: size[1], Location: time.UTC, PartialStep: 10 * time.Minute,
			GreatestTimeStep: 30 * time.Minute, PenumbralOutlineStep: time.Hour, CentralShadowStep: 10 * time.Minute,
		})
		if !ok {
			t.Fatalf("%dx%d: missing map", size[0], size[1])
		}
		assertSVGElementsInCanvas(t, rendered, size[0], size[1])
	}
}
