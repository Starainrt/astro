package basic

import "math"

// MoonHorizon 返回 UT 儒略日下海平面几何月球中心地平圈（月球恰好在地平线上的观测者轨迹，
// 即月下点周围的地平圈）的 [经度, 纬度] 顶点，单位为度，不重复首点。视差与椭球口径同
// HMoonHeight，不含折射；与 HMoonHeight(经, 纬, ..., 0) 配合时该圈上的点高度角为 0。
// samples<=0 取 360，其余夹到 [12, 1440]。
// MoonHorizon returns sea-level geometric Moon-centre horizon vertices in degrees for a UT Julian
// day: the locus of observers that see the Moon exactly on the horizon. Parallax and the observer
// ellipsoid match HMoonHeight; refraction is excluded.
func MoonHorizon(jdUT float64, samples int) [][2]float64 {
	if !finite(jdUT) {
		return nil
	}
	if samples <= 0 {
		samples = 360
	}
	if samples < 12 {
		samples = 12
	} else if samples > 1440 {
		samples = 1440
	}
	tt := TD2UT(jdUT, true)
	ra, dec := HMoonTrueRaDec(tt)
	distanceAU := HMoonAway(tt) / angularDiameterAstronomicalUnitKM
	parallax := math.Sin(0.0024427777777*rad) / distanceAU
	longitude := (ra - ApparentSiderealTime(jdUT)*15) * rad
	latitude := dec * rad
	if !finite(parallax) || parallax <= 0 || parallax >= 1 || !finite(longitude) || !finite(latitude) {
		return nil
	}
	center := [3]float64{math.Cos(latitude) * math.Cos(longitude), math.Cos(latitude) * math.Sin(longitude), math.Sin(latitude)}
	north := [3]float64{-math.Sin(latitude) * math.Cos(longitude), -math.Sin(latitude) * math.Sin(longitude), math.Cos(latitude)}
	east := [3]float64{-math.Sin(longitude), math.Cos(longitude), 0}
	points := make([][2]float64, samples)
	for index := range points {
		bearing := 2 * math.Pi * float64(index) / float64(samples)
		radius := math.Acos(parallax)
		var point [3]float64
		for iteration := 0; iteration < 8; iteration++ {
			for axis := range point {
				point[axis] = center[axis]*math.Cos(radius) +
					(north[axis]*math.Cos(bearing)+east[axis]*math.Sin(bearing))*math.Sin(radius)
			}
			lat := math.Asin(math.Max(-1, math.Min(1, point[2]))) / rad
			// The topocentric direction is horizontal when its dot product
			// with the geodetic zenith vanishes: cos(radius)=observer/range.
			next := math.Acos(parallax * (pcosi(lat, 0)*math.Cos(lat*rad) + psini(lat, 0)*math.Sin(lat*rad)))
			if math.Abs(next-radius) < 1e-14 {
				break
			}
			radius = next
		}
		points[index] = [2]float64{math.Atan2(point[1], point[0]) / rad, math.Asin(math.Max(-1, math.Min(1, point[2]))) / rad}
	}
	return points
}
