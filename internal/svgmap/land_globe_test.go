package svgmap

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/starainrt/astro/internal/geodata"
)

var landGlobeCoordinatePattern = regexp.MustCompile(`[ML](-?[0-9.]+) (-?[0-9.]+)`)

// renderGlobeLand 渲染球面底图并取回陆地的路径命令与画布。
func renderGlobeLand(t *testing.T, center geodata.GeoPoint) (Frame, string) {
	t.Helper()
	frame := Frame{
		X: 30, Y: 30, Width: 640, Height: 640, Projection: ProjectionOrthographic,
		CenterLongitude: center.Longitude, CenterLatitude: center.Latitude,
	}
	var builder strings.Builder
	fmt.Fprintf(&builder, `<svg xmlns="http://www.w3.org/2000/svg" width="700" height="700"><defs>%s</defs>`,
		frame.ClipDefinition("clip"))
	frame.WriteOcean(&builder)
	frame.WriteGraticule(&builder, "clip")
	frame.WriteLand(&builder, "clip")
	builder.WriteString(`</svg>`)
	rendered := builder.String()
	const marker = `<path class="land" d="`
	rest := rendered[strings.Index(rendered, marker)+len(marker):]
	return frame, rest[:strings.Index(rest, `"`)]
}

// 球面底图的输出路径必须只由短边组成，且全部落在圆盘内。
// 这条断言针对的是一个真实缺陷：把已经裁好的环再交给 PolylineSegments 二次裁剪时，
// 视界上的点深度恰为 0 会被判成不可见，环被拆成多段，每段再用 Z 直线闭合，
// 于是圆盘上出现一条横穿的弦——亚洲大陆看起来就"少了一块"。
func TestOrthographicLandPathHasNoChords(t *testing.T) {
	for _, center := range []geodata.GeoPoint{
		{Longitude: 144.1, Latitude: 24.2},
		{Longitude: -35.9, Latitude: -24.2},
		{Longitude: 0, Latitude: 80},
	} {
		frame, path := renderGlobeLand(t, center)
		subpaths := make([][][2]float64, 0, 2048)
		for _, match := range landGlobeCoordinatePattern.FindAllStringSubmatch(path, -1) {
			x, errX := strconv.ParseFloat(match[1], 64)
			y, errY := strconv.ParseFloat(match[2], 64)
			if errX != nil || errY != nil {
				t.Fatalf("unparsable coordinate %q %q", match[1], match[2])
			}
			if len(subpaths) == 0 || match[0][0] == 'M' {
				subpaths = append(subpaths, nil)
			}
			subpaths[len(subpaths)-1] = append(subpaths[len(subpaths)-1], [2]float64{x, y})
		}
		if len(subpaths) < 100 {
			t.Fatalf("%v: only %d subpaths rendered", center, len(subpaths))
		}
		centerX := frame.X + frame.Width/2
		centerY := frame.Y + frame.Height/2
		radius := math.Min(frame.Width, frame.Height) / 2
		// 视界闭合弧按 180 段加密，整圈弦长约 2°；留 4 倍余量即可判定"这条边不可能是海岸线"。
		limit := radius * 0.09
		vertices := 0
		for _, subpath := range subpaths {
			if len(subpath) < 3 {
				t.Fatalf("%v: subpath with %d points", center, len(subpath))
			}
			vertices += len(subpath)
			for index, point := range subpath {
				if math.Hypot(point[0]-centerX, point[1]-centerY) > radius+0.6 {
					t.Fatalf("%v: vertex %.1f,%.1f falls outside the disk", center, point[0], point[1])
				}
				next := subpath[(index+1)%len(subpath)]
				if gap := math.Hypot(next[0]-point[0], next[1]-point[1]); gap > limit {
					t.Fatalf("%v: subpath edge of %.1f px (limit %.1f) between %.1f,%.1f and %.1f,%.1f",
						center, gap, limit, point[0], point[1], next[0], next[1])
				}
			}
		}
		t.Logf("%v: subpaths=%d vertices=%d", center, len(subpaths), vertices)
	}
}
