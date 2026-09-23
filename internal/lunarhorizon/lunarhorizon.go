// Package lunarhorizon 在等距圆柱地图坐标里细化月食可见区使用的站心地平线。
//
// 极点附近一小段球面地平弧可以横跨近 180° 经度，把它的端点直接连成多边形会在极区
// 切出地平线以外的假可见帽。这里按地图坐标误差自适应细分，并把新增顶点用同一个站心
// 地平方程校正回零高度。
//
// Package lunarhorizon refines the topocentric horizon used by lunar-eclipse visibility
// regions in equirectangular map coordinates.
//
// Near a pole a short spherical horizon arc can span almost 180 degrees of longitude, so
// joining its endpoints directly cuts a false visible cap outside the horizon. This package
// subdivides by error measured in map coordinates and corrects every inserted vertex onto the
// same topocentric zero-altitude curve.
package lunarhorizon

import (
	"math"
	"time"

	"github.com/starainrt/astro/basic"
	"github.com/starainrt/astro/internal/geodata"
)

// normalizeLongitude 把经度归一化到 [-180, 180)；internal/geodata 与 geojson 各有一份等价的未导出实现，无法跨包复用。
func normalizeLongitude(value float64) float64 {
	value = math.Mod(value+180, 360)
	if value < 0 {
		value += 360
	}
	return value - 180
}

// defaultToleranceDegrees 是 GeoJSON 导出使用的误差门限。GeoJSON 客户端可以无限放大，
// 因此这里取远小于任何地图像素的角度。
// defaultToleranceDegrees is the error tolerance used by the GeoJSON export. GeoJSON clients
// can zoom without limit, so it stays far below any map pixel.
const defaultToleranceDegrees = 0.002

// DefaultToleranceDegrees 暴露默认门限，供固定分辨率的调用方设置下限。
// DefaultToleranceDegrees exposes the default tolerance so fixed-resolution callers can floor it.
const DefaultToleranceDegrees = defaultToleranceDegrees

// Refine 用 GeoJSON 导出的默认门限细分一条站心地平线。
// Refine subdivides one topocentric horizon with the GeoJSON export's default tolerance.
//
// Near a pole, a short spherical horizon arc can span almost 180 degrees of
// longitude. Refine in map coordinates before clipping; extra vertices are
// corrected to the same topocentric zero-altitude curve as the source ring.
func Refine(points []geodata.GeoPoint, at time.Time) []geodata.GeoPoint {
	return RefineWithin(points, at, defaultToleranceDegrees)
}

// RefineWithin 用给定误差门限（单位：度）细分一条站心地平线。插入的顶点被校正回零高度，
// 因此细分只增加描述精度，不改变曲线本身；固定分辨率的目标（例如 SVG 地图）可以用
// 与像素尺度相称的门限，避免为了显示不出来的精度生成成千上万个顶点。
// RefineWithin subdivides one topocentric horizon with the given error tolerance in degrees.
// Inserted vertices are corrected back onto zero altitude, so refinement adds description
// accuracy without moving the curve. A fixed-resolution target such as an SVG map can pass a
// tolerance matched to its pixel scale instead of emitting vertices no display can resolve.
func RefineWithin(points []geodata.GeoPoint, at time.Time, toleranceDegrees float64) []geodata.GeoPoint {
	if !(toleranceDegrees > 0) {
		toleranceDegrees = defaultToleranceDegrees
	}
	if len(points) < 3 {
		return points
	}
	jd := basic.Date2JD(at.UTC())
	tt := basic.UTC2TT(jd)
	ra, dec := basic.HMoonTrueRaDec(tt)
	distanceAU := basic.HMoonAway(tt) / 149597870.7
	sidereal := basic.ApparentSiderealTime(basic.UTC2UT1(jd)) * 15
	center := geodata.GeoPoint{Longitude: normalizeLongitude(ra - sidereal), Latitude: dec}
	// Every correction is at the same instant. Reuse its full ephemeris while
	// retaining the ellipsoid and topocentric transform used by HMoonHeight.
	altitudeAt := func(point geodata.GeoPoint) float64 {
		ra, dec := basic.TopocentricRaDec(ra, dec, point.Latitude, point.Longitude, jd, distanceAU, 0)
		hourAngle := (sidereal + point.Longitude - ra) * math.Pi / 180
		latitude := point.Latitude * math.Pi / 180
		declination := dec * math.Pi / 180
		return math.Asin(math.Sin(latitude)*math.Sin(declination)+
			math.Cos(declination)*math.Cos(latitude)*math.Cos(hourAngle)) * 180 / math.Pi
	}
	result := make([]geodata.GeoPoint, 0, len(points))
	var refine func(geodata.GeoPoint, geodata.GeoPoint, int)
	refine = func(first, second geodata.GeoPoint, depth int) {
		middle := geodata.InterpolateGreatCircle(first, second, 0.5)
		linearLongitude := first.Longitude + math.Remainder(second.Longitude-first.Longitude, 360)/2
		errorDeg := math.Hypot(math.Remainder(middle.Longitude-linearLongitude, 360), middle.Latitude-(first.Latitude+second.Latitude)/2)
		if errorDeg <= toleranceDegrees || depth >= 20 {
			result = append(result, first)
			return
		}
		for iteration := 0; iteration < 5; iteration++ {
			altitude := altitudeAt(middle)
			if math.Abs(altitude) < 1e-10 {
				break
			}
			angleDeg := basic.StarAngularSeparation(middle.Longitude, middle.Latitude, center.Longitude, center.Latitude)
			if !(angleDeg > 0) {
				// 只有传入的环并非地平线、中点与该瞬时月下点重合时才会走到这里；此时无法
				// 沿"朝向月下点"的方向修正，保留未修正的中点比产生 NaN 顶点安全。
				// Reached only when the input ring is not a horizon ring and the midpoint
				// coincides with that instant's sub-lunar point. There is no direction toward
				// the sub-lunar point to correct along, so keep the uncorrected midpoint rather
				// than emit a NaN vertex.
				break
			}
			middle = geodata.InterpolateGreatCircle(middle, center, -altitude/angleDeg)
		}
		refine(first, middle, depth+1)
		refine(middle, second, depth+1)
	}
	for index, point := range points {
		refine(point, points[(index+1)%len(points)], 0)
	}
	return result
}
