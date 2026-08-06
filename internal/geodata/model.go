// Package geodata 提供与投影无关的地理拓扑辅助函数，用于 / Package geodata provides projection-neutral geographic topology helpers for
// 地图渲染器和传输编码器使用 / map renderers and transport encoders.
package geodata

// Projection 标识受支持的地图投影 / Projection identifies one of the supported map projections.
type Projection string

const (
	ProjectionEquirectangular Projection = "equirectangular"
	ProjectionNorthPolar      Projection = "north-polar"
	ProjectionSouthPolar      Projection = "south-polar"
)

// GeoPoint 是以度表示的地理点，东经为正 / GeoPoint is a geographic point in degrees, with east longitude positive.
type GeoPoint struct {
	Longitude float64
	Latitude  float64
}
