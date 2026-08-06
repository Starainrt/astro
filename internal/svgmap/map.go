// Package svgmap 提供日食和月掩 SVG 共用的无国界 Natural Earth 地图 / Package svgmap provides the shared, border-free Natural Earth map used by
// 日食、月食和月掩 SVG 渲染器共用 / eclipse and lunar-occultation SVG renderers.
package svgmap

import (
	_ "embed"
	"fmt"
	"math"
	"strings"

	"github.com/starainrt/astro/internal/geodata"
)

const (
	worldLandWidth  = 5760
	worldLandHeight = 2880
	polarLandSize   = 2880
)

// Projection 标识受支持的地图投影 / Projection identifies one of the supported map projections.
type Projection = geodata.Projection

const (
	ProjectionEquirectangular = geodata.ProjectionEquirectangular
	ProjectionNorthPolar      = geodata.ProjectionNorthPolar
	ProjectionSouthPolar      = geodata.ProjectionSouthPolar
)

// GeoPoint 是以度表示的地理点，东经为正 / GeoPoint is a geographic point in degrees, with east longitude positive.
type GeoPoint = geodata.GeoPoint

// Frame 描述一个投影地图在 SVG 画布中的位置 / Frame places one projected map in an SVG canvas.
type Frame struct {
	X          float64
	Y          float64
	Width      float64
	Height     float64
	Projection Projection
}

//go:generate go run ./internal/mapgen -output-dir .

//go:embed land_equirectangular.path
var equirectangularLandPath string

//go:embed land_north_polar.path
var northPolarLandPath string

//go:embed land_south_polar.path
var southPolarLandPath string

// ResolveProjection 校验显式投影，或为高纬事件选择极区视图 / ResolveProjection validates an explicit projection or selects a polar view
// 为始终位于一个半球内的高纬事件选择极区视图 / for a high-latitude event that stays in one hemisphere.
func ResolveProjection(requested Projection, focusLatitude, minimumLatitude, maximumLatitude float64) Projection {
	switch requested {
	case ProjectionEquirectangular, ProjectionNorthPolar, ProjectionSouthPolar:
		return requested
	}
	if focusLatitude >= 60 && minimumLatitude >= -2 {
		return ProjectionNorthPolar
	}
	if focusLatitude <= -60 && maximumLatitude <= 2 {
		return ProjectionSouthPolar
	}
	return ProjectionEquirectangular
}

// IsPolar 判断当前画布是否使用半球方位投影 / IsPolar reports whether the frame uses a hemispheric azimuthal projection.
func (frame Frame) IsPolar() bool {
	return frame.Projection == ProjectionNorthPolar || frame.Projection == ProjectionSouthPolar
}

// Project 将经纬度映射为 SVG 坐标；布尔值为 false 表示点在极区投影半球外 / Project maps longitude and latitude to SVG coordinates. The boolean is false
// 点位于极区投影的可见半球之外时返回 false / when a point lies outside a polar projection's visible hemisphere.
func (frame Frame) Project(longitude, latitude float64) (float64, float64, bool) {
	if !frame.IsPolar() {
		x := frame.X + (longitude+180)/360*frame.Width
		y := frame.Y + (90-latitude)/180*frame.Height
		return x, y, true
	}
	hemisphere := frame.hemisphere()
	if latitude*hemisphere < -1e-9 {
		return 0, 0, false
	}
	radius := (90 - hemisphere*latitude) / 90 * math.Min(frame.Width, frame.Height) / 2
	angle := longitude * math.Pi / 180
	return frame.X + frame.Width/2 + radius*math.Sin(angle),
		frame.Y + frame.Height/2 - radius*math.Cos(angle), true
}

// ClipDefinition 写入地图的矩形或圆形裁剪路径 / ClipDefinition writes the rectangular or circular clipping path for a map.
func (frame Frame) ClipDefinition(id string) string {
	if frame.IsPolar() {
		return fmt.Sprintf(`<clipPath id="%s"><circle cx="%.3f" cy="%.3f" r="%.3f"/></clipPath>`,
			id, frame.X+frame.Width/2, frame.Y+frame.Height/2, math.Min(frame.Width, frame.Height)/2)
	}
	return fmt.Sprintf(`<clipPath id="%s"><rect x="%.3f" y="%.3f" width="%.3f" height="%.3f"/></clipPath>`,
		id, frame.X, frame.Y, frame.Width, frame.Height)
}

// WriteOcean 绘制地图的物理范围 / WriteOcean renders the map's physical extent.
func (frame Frame) WriteOcean(builder *strings.Builder) {
	if frame.IsPolar() {
		fmt.Fprintf(builder, `<circle class="map-ocean" cx="%.3f" cy="%.3f" r="%.3f" fill="#edf3f2"/>`,
			frame.X+frame.Width/2, frame.Y+frame.Height/2, math.Min(frame.Width, frame.Height)/2)
		return
	}
	fmt.Fprintf(builder, `<rect class="map-ocean" x="%.3f" y="%.3f" width="%.3f" height="%.3f" fill="#edf3f2"/>`,
		frame.X, frame.Y, frame.Width, frame.Height)
}

// WriteGraticule 绘制符合投影的经线和纬线 / WriteGraticule renders projection-correct meridians and parallels.
func (frame Frame) WriteGraticule(builder *strings.Builder, clipID string) {
	fmt.Fprintf(builder, `<g class="graticule" clip-path="url(#%s)" fill="none" stroke="#b8c4c3" stroke-width="0.65">`, clipID)
	if !frame.IsPolar() {
		for longitude := -150.0; longitude <= 150; longitude += 30 {
			x, _, _ := frame.Project(longitude, 0)
			fmt.Fprintf(builder, `<line x1="%.3f" y1="%.3f" x2="%.3f" y2="%.3f"/>`, x, frame.Y, x, frame.Y+frame.Height)
		}
		for latitude := -60.0; latitude <= 60; latitude += 30 {
			_, y, _ := frame.Project(0, latitude)
			fmt.Fprintf(builder, `<line x1="%.3f" y1="%.3f" x2="%.3f" y2="%.3f"/>`, frame.X, y, frame.X+frame.Width, y)
		}
	} else {
		centerX := frame.X + frame.Width/2
		centerY := frame.Y + frame.Height/2
		radius := math.Min(frame.Width, frame.Height) / 2
		for _, fraction := range []float64{1.0 / 3, 2.0 / 3, 1} {
			fmt.Fprintf(builder, `<circle cx="%.3f" cy="%.3f" r="%.3f"/>`, centerX, centerY, radius*fraction)
		}
		for longitude := -150.0; longitude <= 180; longitude += 30 {
			x, y, _ := frame.Project(longitude, 0)
			fmt.Fprintf(builder, `<line x1="%.3f" y1="%.3f" x2="%.3f" y2="%.3f"/>`, centerX, centerY, x, y)
		}
	}
	builder.WriteString(`</g>`)
}

// WriteLand 绘制无国界的 Natural Earth 1:50m 陆地 / WriteLand renders Natural Earth 1:50m physical land without borders.
func (frame Frame) WriteLand(builder *strings.Builder, clipID string) {
	path := equirectangularLandPath
	viewWidth := float64(worldLandWidth)
	viewHeight := float64(worldLandHeight)
	if frame.Projection == ProjectionNorthPolar {
		path = northPolarLandPath
		viewWidth, viewHeight = polarLandSize, polarLandSize
	} else if frame.Projection == ProjectionSouthPolar {
		path = southPolarLandPath
		viewWidth, viewHeight = polarLandSize, polarLandSize
	}
	fmt.Fprintf(builder, `<g class="land-layer" clip-path="url(#%s)" fill="#d8d9d2" stroke="#a6aaa4" stroke-width="0.75" stroke-linejoin="round"><path class="land" d="`, clipID)
	builder.WriteString(path)
	fmt.Fprintf(builder, `" transform="matrix(%.9f 0 0 %.9f %.3f %.3f)" fill-rule="evenodd" vector-effect="non-scaling-stroke"/></g>`,
		frame.Width/viewWidth, frame.Height/viewHeight, frame.X, frame.Y)
}

// WriteFrame 绘制地图轮廓 / WriteFrame renders the map outline.
func (frame Frame) WriteFrame(builder *strings.Builder) {
	if frame.IsPolar() {
		fmt.Fprintf(builder, `<circle class="map-frame" cx="%.3f" cy="%.3f" r="%.3f" fill="none" stroke="#707879" stroke-width="1.1"/>`,
			frame.X+frame.Width/2, frame.Y+frame.Height/2, math.Min(frame.Width, frame.Height)/2)
		return
	}
	fmt.Fprintf(builder, `<rect class="map-frame" x="%.3f" y="%.3f" width="%.3f" height="%.3f" fill="none" stroke="#707879" stroke-width="1.1"/>`,
		frame.X, frame.Y, frame.Width, frame.Height)
}

// PolylineSegments 将地理折线裁剪到选定投影并 / PolylineSegments clips a geographic polyline to the selected projection and
// 在等经纬投影中按日界线拆分路径 / splits equirectangular paths at the antimeridian.
func PolylineSegments(points []GeoPoint, projection Projection) [][]GeoPoint {
	return geodata.PolylineSegments(points, projection)
}

// PolygonFragments 将地理多边形裁剪到选定地图范围 / PolygonFragments clips a geographic polygon to the selected map extent.
func PolygonFragments(points []GeoPoint, projection Projection) [][]GeoPoint {
	return geodata.PolygonFragments(points, projection)
}

func (frame Frame) hemisphere() float64 {
	if frame.Projection == ProjectionSouthPolar {
		return -1
	}
	return 1
}
