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
	// ProjectionOrthographic 是正射（球面）投影 / ProjectionOrthographic is the orthographic (globe) projection.
	ProjectionOrthographic = geodata.ProjectionOrthographic
)

// ClipView 是裁剪所依赖的投影与视点 / ClipView carries the projection and its view point.
type ClipView = geodata.ClipView

// GeoPoint 是以度表示的地理点，东经为正 / GeoPoint is a geographic point in degrees, with east longitude positive.
type GeoPoint = geodata.GeoPoint

// Frame 描述一个投影地图在 SVG 画布中的位置 / Frame places one projected map in an SVG canvas.
type Frame struct {
	X          float64
	Y          float64
	Width      float64
	Height     float64
	Projection Projection
	// CenterLongitude 与 CenterLatitude 是正射投影的视点；其他投影忽略。
	// CenterLongitude and CenterLatitude are the orthographic view point; other projections ignore them.
	CenterLongitude float64
	CenterLatitude  float64
}

// Clip 返回该画布对应的裁剪视图 / Clip returns the clip view of this frame.
func (frame Frame) Clip() ClipView {
	return ClipView{
		Projection: frame.Projection,
		Center:     GeoPoint{Longitude: frame.CenterLongitude, Latitude: frame.CenterLatitude},
	}
}

// 底图资产已随仓库提交；重生成工具不随仓库发布。

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
	case ProjectionEquirectangular, ProjectionNorthPolar, ProjectionSouthPolar, ProjectionOrthographic:
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

// IsDisk 判断画布是否为圆盘版式：极区方位投影与正射球面图都只画一个圆。
// IsDisk reports whether the frame is disk shaped: both hemispheric azimuthal and orthographic views draw one circle.
func (frame Frame) IsDisk() bool {
	return frame.IsPolar() || frame.Projection == ProjectionOrthographic
}

// Project 将经纬度映射为 SVG 坐标；布尔值为 false 表示点在极区投影半球外 / Project maps longitude and latitude to SVG coordinates. The boolean is false
// 点位于极区投影的可见半球之外时返回 false / when a point lies outside a polar projection's visible hemisphere.
func (frame Frame) Project(longitude, latitude float64) (float64, float64, bool) {
	if frame.Projection == ProjectionOrthographic {
		x, y, visible := geodata.OrthographicDiskPoint(
			GeoPoint{Longitude: longitude, Latitude: latitude},
			GeoPoint{Longitude: frame.CenterLongitude, Latitude: frame.CenterLatitude},
		)
		if !visible {
			return 0, 0, false
		}
		radius := math.Min(frame.Width, frame.Height) / 2
		return frame.X + frame.Width/2 + radius*x, frame.Y + frame.Height/2 - radius*y, true
	}
	if !frame.IsPolar() {
		offset := longitude + 180
		if frame.CenterLongitude != 0 {
			offset = equirectangularLongitudeOffset(longitude, frame.CenterLongitude)
		}
		x := frame.X + offset/360*frame.Width
		y := frame.Y + (90-latitude)/180*frame.Height
		return x, y, true
	}
	hemisphere := frame.hemisphere()
	if latitude*hemisphere < -1e-9 {
		return 0, 0, false
	}
	radius := (90 - hemisphere*latitude) / 90 * math.Min(frame.Width, frame.Height) / 2
	// 极点俯视下横坐标是 sin(θ)，南北两极的视点方向相反，θ 的符号也相反：
	// 北极 θ = C−λ+180、南极 θ = λ−C，各自把事件经线放到背向极点屏幕方向的一侧，
	// 使事件附近都是北在上、东在右。符号写错就会得到镜像图（东京跑到北京西边）。
	angle := (frame.CenterLongitude - longitude) * math.Pi / 180
	if hemisphere > 0 {
		angle += math.Pi
	} else {
		angle = -angle
	}
	return frame.X + frame.Width/2 + radius*math.Sin(angle),
		frame.Y + frame.Height/2 - radius*math.Cos(angle), true
}

// ClipDefinition 写入地图的矩形或圆形裁剪路径 / ClipDefinition writes the rectangular or circular clipping path for a map.
func (frame Frame) ClipDefinition(id string) string {
	if frame.IsDisk() {
		return fmt.Sprintf(`<clipPath id="%s"><circle cx="%.3f" cy="%.3f" r="%.3f"/></clipPath>`,
			id, frame.X+frame.Width/2, frame.Y+frame.Height/2, math.Min(frame.Width, frame.Height)/2)
	}
	return fmt.Sprintf(`<clipPath id="%s"><rect x="%.3f" y="%.3f" width="%.3f" height="%.3f"/></clipPath>`,
		id, frame.X, frame.Y, frame.Width, frame.Height)
}

// WriteOcean 绘制地图的物理范围 / WriteOcean renders the map's physical extent.
func (frame Frame) WriteOcean(builder *strings.Builder) {
	if frame.IsDisk() {
		fmt.Fprintf(builder, `<circle class="map-ocean" cx="%.3f" cy="%.3f" r="%.3f" fill="#edf3f2"/>`,
			frame.X+frame.Width/2, frame.Y+frame.Height/2, math.Min(frame.Width, frame.Height)/2)
		return
	}
	fmt.Fprintf(builder, `<rect class="map-ocean" x="%.3f" y="%.3f" width="%.3f" height="%.3f" fill="#edf3f2"/>`,
		frame.X, frame.Y, frame.Width, frame.Height)
}

// WriteGraticule 绘制符合投影的经线和纬线 / WriteGraticule renders projection-correct meridians and parallels.
func (frame Frame) WriteGraticule(builder *strings.Builder, clipID string) {
	if frame.Projection == ProjectionOrthographic {
		frame.writeOrthographicGraticule(builder, clipID)
		return
	}
	fmt.Fprintf(builder, `<g class="graticule" clip-path="url(#%s)" fill="none" stroke="#b8c4c3" stroke-width="0.65">`, clipID)
	if !frame.IsPolar() {
		first, last := -150.0, 150.0
		if frame.CenterLongitude != 0 {
			first, last = frame.CenterLongitude-180, frame.CenterLongitude+180
		}
		for longitude := first; longitude <= last; longitude += 30 {
			x, _, _ := frame.Project(longitude, 0)
			if frame.CenterLongitude != 0 && (x < frame.X-0.5 || x > frame.X+frame.Width+0.5) {
				continue
			}
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
		first, last := -150.0, 180.0
		if frame.CenterLongitude != 0 {
			first, last = frame.CenterLongitude-180, frame.CenterLongitude+180
		}
		for longitude := first; longitude <= last; longitude += 30 {
			x, y, _ := frame.Project(longitude, 0)
			fmt.Fprintf(builder, `<line x1="%.3f" y1="%.3f" x2="%.3f" y2="%.3f"/>`, centerX, centerY, x, y)
		}
	}
	builder.WriteString(`</g>`)
}

// WriteLand 绘制无国界的 Natural Earth 1:50m 陆地 / WriteLand renders Natural Earth 1:50m physical land without borders.
func (frame Frame) WriteLand(builder *strings.Builder, clipID string) {
	if frame.Projection == ProjectionOrthographic {
		frame.writeOrthographicLand(builder, clipID)
		return
	}
	if frame.Projection == ProjectionEquirectangular && frame.CenterLongitude != 0 {
		frame.writeCenteredEquirectangularLand(builder, clipID)
		return
	}
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
	fmt.Fprintf(builder, `<g class="land-layer" clip-path="url(#%s)" fill="#d8d9d2" stroke="#a6aaa4" stroke-width="0.75" stroke-linejoin="round">`, clipID)
	landX, landScaleX := frame.X, frame.Width/viewWidth
	if frame.IsPolar() {
		// 极地底图都按 θ=λ 烘焙。北极的投影与之手性相反，需先水平镜像变成 −λ 再转 C+180；
		// 南极手性相同，直接转 −C 即可。两侧目标角度都是 θ = ±(C−λ)+偏移。
		landRotation := -frame.CenterLongitude
		if frame.hemisphere() > 0 {
			landX, landScaleX = frame.X+frame.Width, -frame.Width/viewWidth
			landRotation = frame.CenterLongitude + 180
		}
		fmt.Fprintf(builder, `<g transform="rotate(%.4f %.3f %.3f)">`,
			landRotation, frame.X+frame.Width/2, frame.Y+frame.Height/2)
	}
	fmt.Fprintf(builder, `<path class="land" d="`)
	builder.WriteString(path)
	fmt.Fprintf(builder, `" transform="matrix(%.9f 0 0 %.9f %.3f %.3f)" fill-rule="evenodd" vector-effect="non-scaling-stroke"/>`,
		landScaleX, frame.Height/viewHeight, landX, frame.Y)
	if frame.IsPolar() {
		builder.WriteString(`</g>`)
	}
	builder.WriteString(`</g>`)
}

// WriteFrame 绘制地图轮廓 / WriteFrame renders the map outline.
func (frame Frame) WriteFrame(builder *strings.Builder) {
	if frame.IsDisk() {
		fmt.Fprintf(builder, `<circle class="map-frame" cx="%.3f" cy="%.3f" r="%.3f" fill="none" stroke="#707879" stroke-width="1.1"/>`,
			frame.X+frame.Width/2, frame.Y+frame.Height/2, math.Min(frame.Width, frame.Height)/2)
		return
	}
	fmt.Fprintf(builder, `<rect class="map-frame" x="%.3f" y="%.3f" width="%.3f" height="%.3f" fill="none" stroke="#707879" stroke-width="1.1"/>`,
		frame.X, frame.Y, frame.Width, frame.Height)
}

// PolylineSegments 将地理折线裁剪到选定投影并 / PolylineSegments clips a geographic polyline to the selected projection and
// 在等经纬投影中按日界线拆分路径 / splits equirectangular paths at the antimeridian.
func PolylineSegments(points []GeoPoint, view ClipView) [][]GeoPoint {
	return geodata.PolylineSegments(points, view)
}

// PolygonFragments 将地理多边形裁剪到选定地图范围 / PolygonFragments clips a geographic polygon to the selected map extent.
func PolygonFragments(points []GeoPoint, view ClipView) [][]GeoPoint {
	return geodata.PolygonFragments(points, view)
}

func (frame Frame) hemisphere() float64 {
	if frame.Projection == ProjectionSouthPolar {
		return -1
	}
	return 1
}

// equirectangularLongitudeOffset 返回经度相对居中经线的偏移，换算成 0…360 的剂量。
// 居中经线落在画面正中，其对面的经线落在左右任一边界上。
func equirectangularLongitudeOffset(longitude, center float64) float64 {
	offset := math.Mod(longitude-center, 360)
	if offset < 0 {
		offset += 360
	}
	offset += 180
	if offset >= 360 {
		offset -= 360
	}
	return offset
}

// writeCenteredEquirectangularLand 画按经度居中后的陆地：底图路径只出现一次，
// 再向图框缺的那一侧补一份平移副本覆盖环绕部分，因此输出大小与不居中时同量级。
func (frame Frame) writeCenteredEquirectangularLand(builder *strings.Builder, clipID string) {
	shift := frame.CenterLongitude / 360 * frame.Width
	x := frame.X - shift
	fmt.Fprintf(builder, `<g class="land-layer" clip-path="url(#%s)" fill="#d8d9d2" stroke="#a6aaa4" stroke-width="0.75" stroke-linejoin="round">`, clipID)
	builder.WriteString(`<defs><path class="land" id="land-equirectangular-tile" d="`)
	builder.WriteString(equirectangularLandPath)
	fmt.Fprintf(builder, `" transform="matrix(%.9f 0 0 %.9f 0 0)" fill-rule="evenodd"/></defs>`,
		frame.Width/float64(worldLandWidth), frame.Height/float64(worldLandHeight))
	// 居中经线为负时窗口相对图块整体右移，只有向左补一份才能盖住图框左侧的 |C|° 空带。
	offsets := []float64{0, frame.Width}
	if shift < 0 {
		offsets = []float64{-frame.Width, 0}
	}
	for _, dx := range offsets {
		fmt.Fprintf(builder, `<use href="#land-equirectangular-tile" transform="translate(%.3f %.3f)" vector-effect="non-scaling-stroke"/>`,
			x+dx, frame.Y)
	}
	builder.WriteString(`</g>`)
}
