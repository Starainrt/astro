package svgmap

import (
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/starainrt/astro/internal/geodata"
)

var (
	landRingsOnce  sync.Once
	landRingsCache [][]geodata.GeoPoint
)

// landRings 把烘焙的等经纬陆地路径还原成经纬度环。
// 路径编码在 5760x2880 的整数像素网格上，即 1/16° 量化；在球面图上不足 0.2 像素。
// 正射图的视点随事件变化，无法预烘焙，只能运行时投影。
func landRings() [][]geodata.GeoPoint {
	landRingsOnce.Do(func() {
		landRingsCache = decodeLandPath(equirectangularLandPath, worldLandWidth, worldLandHeight)
	})
	return landRingsCache
}

// scanLandNumber 读取一个整数坐标；SVG 允许负号紧接前一个数字，因此不能按空白切分。
func scanLandNumber(text string, position int) (int, int, bool) {
	// 坐标对之间有空白，调用方可能在命令字母之后直接进入，所以这里要自己跳过分隔符。
	for position < len(text) && (text[position] == ' ' || text[position] == ',') {
		position++
	}
	start := position
	if position < len(text) && (text[position] == '-' || text[position] == '+') {
		position++
	}
	digits := position
	for position < len(text) && text[position] >= '0' && text[position] <= '9' {
		position++
	}
	if position == digits {
		return 0, start, false
	}
	value, err := strconv.Atoi(text[start:position])
	if err != nil {
		return 0, start, false
	}
	return value, position, true
}

// decodeLandPath 解析内嵌底图的紧凑路径，只用到 M/l/h/v/z 五种命令。
func decodeLandPath(text string, width, height int) [][]geodata.GeoPoint {
	rings := make([][]geodata.GeoPoint, 0, 2048)
	var ring []geodata.GeoPoint
	x, y := 0, 0
	position := 0
	var command byte
	point := func(px, py int) geodata.GeoPoint {
		return geodata.GeoPoint{
			Longitude: float64(px)/float64(width)*360 - 180,
			Latitude:  90 - float64(py)/float64(height)*180,
		}
	}
	flush := func() {
		if len(ring) >= 3 {
			rings = append(rings, ring)
		}
		ring = nil
	}
	for position < len(text) {
		for position < len(text) && (text[position] == ' ' || text[position] == ',') {
			position++
		}
		if position >= len(text) {
			break
		}
		if character := text[position]; (character >= 'A' && character <= 'Z') || (character >= 'a' && character <= 'z') {
			command = character
			position++
			if command == 'z' || command == 'Z' {
				flush()
				command = 0
			}
			continue
		}
		value, next, ok := scanLandNumber(text, position)
		if !ok {
			break
		}
		position = next
		switch command {
		case 'M':
			second, afterSecond, secondOK := scanLandNumber(text, position)
			if !secondOK {
				return rings
			}
			position = afterSecond
			flush()
			x, y = value, second
			ring = append(ring, point(x, y))
			// M 之后的裸坐标对按规范是隐式相对直线段。
			command = 'l'
		case 'l':
			second, afterSecond, secondOK := scanLandNumber(text, position)
			if !secondOK {
				return rings
			}
			position = afterSecond
			x += value
			y += second
			ring = append(ring, point(x, y))
		case 'h':
			x += value
			ring = append(ring, point(x, y))
		case 'v':
			y += value
			ring = append(ring, point(x, y))
		}
	}
	flush()
	return rings
}

// writeOrthographicPath 先按视界裁剪再投影一段地理折线，写成 SVG 路径命令。
func (frame Frame) writeOrthographicPath(builder *strings.Builder, view geodata.ClipView, points []geodata.GeoPoint, close bool) {
	for _, segment := range geodata.PolylineSegments(points, view) {
		frame.writeOrthographicRing(builder, segment, close)
	}
}

// writeOrthographicRing 投影一个已经裁剪好的环或折线。
// 这里不能再调 PolylineSegments：视界上的点深度恰为 0，会被判成不可见而把环拆开，
// 每个碎片再用 Z 直线闭合，就会在圆盘上留下横穿的弦。
func (frame Frame) writeOrthographicRing(builder *strings.Builder, points []geodata.GeoPoint, close bool) {
	written := 0
	for _, item := range points {
		x, y, visible := frame.Project(item.Longitude, item.Latitude)
		if !visible {
			continue
		}
		command := "L"
		if written == 0 {
			command = "M"
		}
		// 球面图上 0.1 像素已远小于海岸线本身的量化误差。
		fmt.Fprintf(builder, "%s%.1f %.1f", command, x, y)
		written++
	}
	if written >= 2 && close {
		builder.WriteString("Z")
	}
}

// writeOrthographicLand 在正射球面图上运行时投影并裁剪陆地。
func (frame Frame) writeOrthographicLand(builder *strings.Builder, clipID string) {
	view := frame.Clip()
	var path strings.Builder
	for _, ring := range landRings() {
		for _, fragment := range geodata.PolygonFragments(ring, view) {
			frame.writeOrthographicRing(&path, fragment, true)
		}
	}
	fmt.Fprintf(builder, `<g class="land-layer" clip-path="url(#%s)" fill="#d8d9d2" stroke="#a6aaa4" stroke-width="0.75" stroke-linejoin="round"><path class="land" d="%s" fill-rule="evenodd" vector-effect="non-scaling-stroke"/></g>`,
		clipID, path.String())
}

// writeOrthographicGraticule 绘制球面经纬网；经线是完整大圆，纬线是等纬圈，都按视界裁剪。
func (frame Frame) writeOrthographicGraticule(builder *strings.Builder, clipID string) {
	view := frame.Clip()
	fmt.Fprintf(builder, `<g class="graticule" clip-path="url(#%s)" fill="none" stroke="#b8c4c3" stroke-width="0.65">`, clipID)
	for longitude := -180.0; longitude < 180; longitude += 30 {
		points := make([]geodata.GeoPoint, 0, 91)
		for latitude := -90.0; latitude <= 90; latitude += 2 {
			points = append(points, geodata.GeoPoint{Longitude: longitude, Latitude: latitude})
		}
		builder.WriteString(`<path d="`)
		frame.writeOrthographicPath(builder, view, points, false)
		builder.WriteString(`"/>`)
	}
	for latitude := -60.0; latitude <= 60; latitude += 30 {
		points := make([]geodata.GeoPoint, 0, 181)
		for longitude := -180.0; longitude <= 180; longitude += 2 {
			points = append(points, geodata.GeoPoint{Longitude: longitude, Latitude: latitude})
		}
		builder.WriteString(`<path d="`)
		frame.writeOrthographicPath(builder, view, points, false)
		builder.WriteString(`"/>`)
	}
	builder.WriteString(`</g>`)
}
