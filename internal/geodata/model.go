// Package geodata 提供与投影无关的地理拓扑辅助函数，用于 / Package geodata provides projection-neutral geographic topology helpers for
// 地图渲染器和传输编码器使用 / map renderers and transport encoders.
package geodata

// Projection 标识受支持的地图投影 / Projection identifies one of the supported map projections.
type Projection string

const (
	ProjectionEquirectangular Projection = "equirectangular"
	ProjectionNorthPolar      Projection = "north-polar"
	ProjectionSouthPolar      Projection = "south-polar"
	// ProjectionOrthographic 是正射（球面）投影，只画朝向视点的半个地球。
	// ProjectionOrthographic is the orthographic (globe) projection, drawing only the hemisphere facing the view point.
	ProjectionOrthographic Projection = "orthographic"
)

// ClipView 是裁剪所依赖的投影及其参数 / ClipView carries the projection and the parameters clipping depends on.
type ClipView struct {
	Projection Projection
	// Center 是正射投影的视点，也就是可见半球的中心；其他投影忽略它。
	// Center is the orthographic view point, the middle of the visible hemisphere; other projections ignore it.
	Center GeoPoint
}

// Orthographic 判断视图是否为正射投影 / Orthographic reports whether the view uses the orthographic projection.
func (view ClipView) Orthographic() bool {
	return view.Projection == ProjectionOrthographic
}

// GeoPoint 是以度表示的地理点，东经为正 / GeoPoint is a geographic point in degrees, with east longitude positive.
type GeoPoint struct {
	Longitude float64
	Latitude  float64
}
