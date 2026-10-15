package catalog

import (
	"math"
	"sort"
	"strings"
)

// Near 返回星表历元下的锥形邻域，按角距由近到远 / cone search at catalog epochs, nearest first.
func (c *Catalog) Near(ra, dec, radius float64) []Entry {
	if c == nil || !finite(ra) || !finite(dec) || !finite(radius) || radius < 0 || dec < -90 || dec > 90 {
		return nil
	}
	ra = normalizeRA(ra)
	// 锥面不含天极时，小圆的 RA 半宽恰为 asin(sin r / cos δ)，且 |δ|+r < 90 保证比值 < 1；含天极时整带扫描。
	halfWidth := 180.0
	if math.Abs(dec)+radius < 90 {
		halfWidth = math.Asin(math.Sin(radius*math.Pi/180)/math.Cos(dec*math.Pi/180)) * 180 / math.Pi
	}
	type hit struct {
		index    int
		distance float64
	}
	hits := make([]hit, 0)
	collect := func(indexes []int) {
		for _, index := range indexes {
			coordinate := c.entries[index].Coordinate
			if distance := angularDistanceDeg(ra, dec, coordinate.RA, coordinate.Dec); distance <= radius+1e-12 {
				hits = append(hits, hit{index: index, distance: distance})
			}
		}
	}
	for band := decBand(dec - radius); band <= decBand(dec+radius); band++ {
		indexes := c.bands[band]
		if halfWidth >= 180 {
			collect(indexes)
			continue
		}
		low, high := ra-halfWidth-1e-9, ra+halfWidth+1e-9
		collect(c.raRange(indexes, math.Max(low, 0), math.Min(high, 360)))
		if low < 0 {
			collect(c.raRange(indexes, low+360, 360))
		}
		if high > 360 {
			collect(c.raRange(indexes, 0, high-360))
		}
	}
	sort.SliceStable(hits, func(i, j int) bool { return hits[i].distance < hits[j].distance })
	result := make([]Entry, len(hits))
	for i, item := range hits {
		result[i] = cloneEntry(c.entries[item.index])
	}
	return result
}

// raRange 返回带内 RA ∈ [low, high] 的下标切片；带内已按 RA 升序。
func (c *Catalog) raRange(indexes []int, low, high float64) []int {
	start := sort.Search(len(indexes), func(i int) bool { return c.entries[indexes[i]].Coordinate.RA >= low })
	end := sort.Search(len(indexes), func(i int) bool { return c.entries[indexes[i]].Coordinate.RA > high })
	if start >= end {
		return nil
	}
	return indexes[start:end]
}

// Band 返回赤纬带内的记录 / records in a declination band.
func (c *Catalog) Band(decMin, decMax float64) []Entry {
	if c == nil || !finite(decMin) || !finite(decMax) || decMin > decMax || decMax < -90 || decMin > 90 {
		return nil
	}
	if decMin < -90 {
		decMin = -90
	}
	if decMax > 90 {
		decMax = 90
	}
	result := make([]Entry, 0)
	for band := decBand(decMin); band <= decBand(decMax); band++ {
		for _, index := range c.bands[band] {
			entry := c.entries[index]
			if entry.Coordinate.Dec >= decMin && entry.Coordinate.Dec <= decMax {
				result = append(result, cloneEntry(entry))
			}
		}
	}
	return result
}

// EclipticBandFilter 创建按 J2000 平黄道纬度过滤的函数 / build a mean-ecliptic latitude filter.
func EclipticBandFilter(halfWidthDeg float64) func(Entry) bool {
	if !finite(halfWidthDeg) || halfWidthDeg < 0 {
		return func(Entry) bool { return false }
	}
	if halfWidthDeg >= 90 {
		return func(Entry) bool { return true }
	}
	const obliquity = 23.439291111
	return func(entry Entry) bool {
		ra := entry.Coordinate.RA * math.Pi / 180
		dec := entry.Coordinate.Dec * math.Pi / 180
		eps := obliquity * math.Pi / 180
		beta := math.Asin(math.Sin(dec)*math.Cos(eps) - math.Cos(dec)*math.Sin(eps)*math.Sin(ra))
		return math.Abs(beta*180/math.Pi) <= halfWidthDeg
	}
}

// MagnitudeFilter 创建按 V 星等过滤的函数 / build a V-magnitude filter.
func MagnitudeFilter(maxV float64) func(Entry) bool {
	if !finite(maxV) {
		return func(Entry) bool { return false }
	}
	return func(entry Entry) bool {
		value, ok := VMagnitude(entry)
		return ok && value <= maxV
	}
}

// VMagnitude 将支持的波段换算为 V / convert supported bands to V magnitude.
func VMagnitude(entry Entry) (float64, bool) {
	if !finite(entry.Magnitude) {
		return 0, false
	}
	switch strings.ToUpper(string(entry.Band)) {
	case "V":
		return entry.Magnitude, true
	case "G":
		if !finite(entry.BPRP) || entry.BPRP <= -0.5 || entry.BPRP >= 5.0 {
			return 0, false
		}
		x := entry.BPRP
		gMinusV := -0.02704 + 0.01424*x - 0.2156*x*x + 0.01426*x*x*x
		return entry.Magnitude - gMinusV, true
	default:
		return 0, false
	}
}

func angularDistanceDeg(ra1, dec1, ra2, dec2 float64) float64 {
	const rad = math.Pi / 180
	lat1, lat2 := dec1*rad, dec2*rad
	dra := math.Abs((ra1 - ra2) * rad)
	if dra > math.Pi {
		dra = 2*math.Pi - dra
	}
	value := math.Sin(lat1)*math.Sin(lat2) + math.Cos(lat1)*math.Cos(lat2)*math.Cos(dra)
	if value > 1 {
		value = 1
	}
	if value < -1 {
		value = -1
	}
	return math.Acos(value) / rad
}

func normalizeRA(value float64) float64 {
	value = math.Mod(value, 360)
	if value < 0 {
		value += 360
	}
	return value
}

func finite(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}
