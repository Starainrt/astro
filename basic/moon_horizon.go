package basic

// MoonHorizon 返回 UTC 儒略日下海平面几何月球中心地平圈（月球恰好在地平线上的观测者轨迹，
// 即月下点周围的地平圈）的 [经度, 纬度] 顶点，单位为度，不重复首点。视差与椭球口径同
// HMoonHeight，不含折射；与 HMoonHeight(经, 纬, ..., 0) 配合时该圈上的点高度角为 0。
// samples<=0 取 360，其余夹到 [12, 1440]。
// MoonHorizon returns sea-level geometric Moon-centre horizon vertices in degrees for a UTC Julian
// day: the locus of observers that see the Moon exactly on the horizon. Parallax and the observer
// ellipsoid match HMoonHeight; refraction is excluded.
func MoonHorizon(jdUTC float64, samples int) [][2]float64 {
	if !finite(jdUTC) {
		return nil
	}
	return MoonStateAt(jdUTC).MoonHorizon(samples)
}
