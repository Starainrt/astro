package svg

import "github.com/starainrt/astro/internal/svgmap"

// EclipseMapProjection 控制全球日月食地图投影；零值根据事件几何选择投影。
// EclipseMapProjection controls a global eclipse map projection. The zero value selects the projection from the event geometry.
type EclipseMapProjection string

const (
	// EclipseMapProjectionAuto 根据事件几何自动选择投影。
	// EclipseMapProjectionAuto selects a projection from event geometry.
	EclipseMapProjectionAuto EclipseMapProjection = ""
	// EclipseMapProjectionEquirectangular 使用等经纬投影。
	// EclipseMapProjectionEquirectangular uses the equirectangular projection.
	EclipseMapProjectionEquirectangular EclipseMapProjection = "equirectangular"
	// EclipseMapProjectionNorthPolar 使用北极方位等距投影。
	// EclipseMapProjectionNorthPolar uses the north-polar azimuthal equidistant projection.
	EclipseMapProjectionNorthPolar EclipseMapProjection = "north-polar"
	// EclipseMapProjectionSouthPolar 使用南极方位等距投影。
	// EclipseMapProjectionSouthPolar uses the south-polar azimuthal equidistant projection.
	EclipseMapProjectionSouthPolar EclipseMapProjection = "south-polar"
)

func internalEclipseMapProjection(value EclipseMapProjection) svgmap.Projection {
	return svgmap.Projection(value)
}

func validEclipseMapProjection(value EclipseMapProjection) bool {
	switch value {
	case EclipseMapProjectionAuto, EclipseMapProjectionEquirectangular,
		EclipseMapProjectionNorthPolar, EclipseMapProjectionSouthPolar:
		return true
	default:
		return false
	}
}

func eclipseMapProjectionLabel(projection svgmap.Projection, language string) string {
	if language == "en" {
		switch projection {
		case svgmap.ProjectionNorthPolar:
			return "North-polar azimuthal equidistant projection"
		case svgmap.ProjectionSouthPolar:
			return "South-polar azimuthal equidistant projection"
		default:
			return "Equirectangular projection"
		}
	}
	switch projection {
	case svgmap.ProjectionNorthPolar:
		return "北极方位等距投影"
	case svgmap.ProjectionSouthPolar:
		return "南极方位等距投影"
	default:
		return "等经纬投影"
	}
}
