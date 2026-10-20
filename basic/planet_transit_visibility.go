package basic

import "math"

// PlanetTransitVisibilityColumn 子午线上的可见纬度区间，端点单位为度 / visible latitude intervals on a meridian, in degrees.
type PlanetTransitVisibilityColumn struct {
	Longitude float64
	Intervals [][2]float64
}

// PlanetTransitVisibilityColumnsWithInstance 按 2° 经度、1° 纬度探测站心可见区并细化边界 / sample topocentric visibility at 2-degree longitude and 1-degree latitude spacing and refine boundaries.
func PlanetTransitVisibilityColumnsWithInstance(transit PlanetTransitResult, inst *Instance) []PlanetTransitVisibilityColumn {
	snapshot, ok := newPlanetTransitSnapshot(transit, inst)
	if !ok {
		return nil
	}
	snapshot.visibilityOnce.Do(func() {
		for column := 0; column <= 180; column++ {
			longitude := -180 + float64(column)*2
			visible := func(latitude float64) bool {
				local := snapshot.local(transit, Observer{Longitude: longitude, Latitude: latitude})
				return local.Valid && local.Visible
			}
			value := PlanetTransitVisibilityColumn{Longitude: longitude}
			previous, start := visible(-90), -90.0
			for latitude := -89.0; latitude <= 90; latitude++ {
				current := visible(latitude)
				if current != previous {
					left, right := latitude-1, latitude
					for right-left > 0.0001 {
						middle := (left + right) / 2
						if visible(middle) == previous {
							left = middle
						} else {
							right = middle
						}
					}
					boundary := (left + right) / 2
					if current {
						start = boundary
					} else {
						value.Intervals = append(value.Intervals, [2]float64{start, boundary})
					}
				}
				previous = current
			}
			if previous {
				value.Intervals = append(value.Intervals, [2]float64{start, 90})
			}
			snapshot.visibility = append(snapshot.visibility, value)
		}
	})
	result := make([]PlanetTransitVisibilityColumn, len(snapshot.visibility))
	for index, column := range snapshot.visibility {
		result[index] = PlanetTransitVisibilityColumn{Longitude: column.Longitude, Intervals: append([][2]float64(nil), column.Intervals...)}
	}
	return result
}

// PlanetTransitVisibilityPolygonsWithInstance 可见区的经度条带多边形 / longitude-strip polygons of the visible region.
func PlanetTransitVisibilityPolygonsWithInstance(transit PlanetTransitResult, inst *Instance) [][]PlanetTransitIsochronePoint {
	columns := PlanetTransitVisibilityColumnsWithInstance(transit, inst)
	var result [][]PlanetTransitIsochronePoint
	for index := 1; index < len(columns); index++ {
		left, right := columns[index-1], columns[index]
		for _, interval := range left.Intervals {
			for _, other := range right.Intervals {
				if math.Min(interval[1], other[1]) <= math.Max(interval[0], other[0]) {
					continue
				}
				result = append(result, []PlanetTransitIsochronePoint{
					{Longitude: left.Longitude, Latitude: interval[0]}, {Longitude: right.Longitude, Latitude: other[0]},
					{Longitude: right.Longitude, Latitude: other[1]}, {Longitude: left.Longitude, Latitude: interval[1]},
				})
			}
		}
	}
	return result
}
