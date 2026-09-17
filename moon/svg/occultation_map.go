package svg

import "github.com/starainrt/astro/internal/svgmap"

// MapProjection 控制全球月掩地图使用的投影。
// 零值根据事件几何自动选择投影。
// 与 eclipse/svg.EclipseMapProjection 是同一概念的两套名字，取值字符串逐对相同。
// MapProjection controls the projection used by a global occultation map.
// The zero value selects a projection from the event geometry.
type MapProjection string

const (
	// MapProjectionAuto 根据事件几何自动选择投影。
	// MapProjectionAuto selects a projection from event geometry.
	MapProjectionAuto MapProjection = ""
	// MapProjectionEquirectangular 使用等经纬投影。
	// MapProjectionEquirectangular uses the equirectangular projection.
	MapProjectionEquirectangular MapProjection = "equirectangular"
	// MapProjectionNorthPolar 使用北极方位等距投影。
	// MapProjectionNorthPolar uses the north-polar azimuthal equidistant projection.
	MapProjectionNorthPolar MapProjection = "north-polar"
	// MapProjectionSouthPolar 使用南极方位等距投影。
	// MapProjectionSouthPolar uses the south-polar azimuthal equidistant projection.
	MapProjectionSouthPolar MapProjection = "south-polar"
	// MapProjectionOrthographic 使用正射球面投影，视点取事件中心，只画朝向观察者的半球。
	// MapProjectionOrthographic uses the orthographic globe, viewed from the event centre; only the facing hemisphere is drawn.
	MapProjectionOrthographic MapProjection = "orthographic"
)

func internalMapProjection(value MapProjection) svgmap.Projection {
	return svgmap.Projection(value)
}
