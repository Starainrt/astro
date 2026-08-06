package svgmap

import "github.com/starainrt/astro/internal/geodata"

// SphericalCircle 返回球面小圆上的等间隔采样点 / SphericalCircle returns evenly spaced points on a small circle on the
// 球面小圆；方位角从地理北方顺时针采样 / sphere. Bearings are sampled clockwise from geographic north.
func SphericalCircle(center GeoPoint, radiusDegrees float64, points int) []GeoPoint {
	return geodata.SphericalCircle(center, radiusDegrees, points)
}

// JoinPolylineSegments 按最近端点连接无序边界线段 / JoinPolylineSegments joins unordered boundary segments by their nearest
// 端点连接；输入线段不会被修改 / endpoints. The input segments are not modified.
func JoinPolylineSegments(segments [][]GeoPoint) []GeoPoint {
	return geodata.JoinPolylineSegments(segments)
}

// ShortestCircleArc 返回两点之间较短的采样圆弧 / ShortestCircleArc returns the shorter sampled arc from one point to another.
func ShortestCircleArc(circle []GeoPoint, from, to GeoPoint) []GeoPoint {
	return geodata.ShortestCircleArc(circle, from, to)
}

// SameGeoPoint 判断两个经纬度点是否在拓扑所需精度内相等 / SameGeoPoint reports whether two longitude/latitude points are equal within
// 地图拓扑辅助函数所需的精度内相等 / the precision needed by the map topology helpers.
func SameGeoPoint(a, b GeoPoint) bool {
	return geodata.SameGeoPoint(a, b)
}

// VisibleHemispherePolygons 返回以指定中心为中心的半球多边形 / VisibleHemispherePolygons returns polygons for the hemisphere centered on
// 中心的半球多边形，并裁剪到请求的地图投影 / center, clipped to the requested map projection.
func VisibleHemispherePolygons(center GeoPoint, projection Projection, samples int) [][]GeoPoint {
	return geodata.VisibleHemispherePolygons(center, projection, samples)
}
