package geojson

import (
	"math"
	"sort"

	"github.com/starainrt/astro/internal/geodata"
)

// GeoJSON joins vertices with straight longitude/latitude segments. Bound
// their deviation from the spherical edges before antimeridian clipping,
// especially for a short arc that passes close to either pole.
func sampleSphericalMapRing(points []geodata.GeoPoint) []geodata.GeoPoint {
	return sampleSphericalMapRingWithin(points, sphericalMapChordLimitKM, sphericalMapChordErrorDegrees)
}

// Exported map chords are bounded so that a straight lon/lat segment still
// follows the spherical edge. Filled footprint polygons tolerate a much coarser
// bound than the strokes that carry the path limits: they are drawn as
// translucent fills, and their rings dominate the payload (a solar eclipse
// exports a hundred penumbral footprints).
const (
	sphericalMapChordLimitKM       = 100.0
	sphericalMapChordErrorDegrees  = 0.002
	sphericalFillChordLimitKM      = 400.0
	sphericalFillChordErrorDegrees = 0.02
)

// sampleSphericalMapRingWithin is sampleSphericalMapRing with explicit chord
// limits, used for fill-only geometry.
func sampleSphericalMapRingWithin(
	points []geodata.GeoPoint,
	chordLimitKM, chordErrorDegrees float64,
) []geodata.GeoPoint {
	if len(points) < 3 {
		return points
	}
	if len(points) == 3 {
		return sampleSphericalMapTriangle(points)
	}
	result := make([]geodata.GeoPoint, 0, len(points))
	for index, point := range points {
		result = appendSphericalMapArcWithin(
			result, point, points[(index+1)%len(points)], 0, chordLimitKM, chordErrorDegrees,
		)
	}
	return result
}

func sphericalMapChordError(first, middle, last geodata.GeoPoint) float64 {
	longitude := first.Longitude + math.Remainder(last.Longitude-first.Longitude, 360)/2
	return math.Hypot(math.Remainder(middle.Longitude-longitude, 360), middle.Latitude-(first.Latitude+last.Latitude)/2)
}

func appendSphericalMapArc(points []geodata.GeoPoint, first, last geodata.GeoPoint, depth int) []geodata.GeoPoint {
	return appendSphericalMapArcWithin(
		points, first, last, depth, sphericalMapChordLimitKM, sphericalMapChordErrorDegrees,
	)
}

func appendSphericalMapArcWithin(
	points []geodata.GeoPoint,
	first, last geodata.GeoPoint,
	depth int,
	chordLimitKM, chordErrorDegrees float64,
) []geodata.GeoPoint {
	middle := geodata.InterpolateGreatCircle(first, last, 0.5)
	// Keep exported map chords bounded even when a great-circle arc is nearly
	// linear in lon/lat (notably the long horizon closure edges of shallow
	// polar eclipses). The adaptive angular-error test alone cannot see that
	// case and leaves visually abrupt 250+ km segments.
	arcDistanceKM := solarCentralBandGeoPointDistanceKM(first, last)
	if (sphericalMapChordError(first, middle, last) <= chordErrorDegrees && arcDistanceKM <= chordLimitKM) || depth >= 20 {
		return append(points, first)
	}
	points = appendSphericalMapArcWithin(points, first, middle, depth+1, chordLimitKM, chordErrorDegrees)
	return appendSphericalMapArcWithin(points, middle, last, depth+1, chordLimitKM, chordErrorDegrees)
}

func sampleSphericalMapPath(source []pathSample) []pathSample {
	if len(source) < 2 {
		return source
	}
	result := make([]pathSample, 0, len(source))
	var refine func(pathSample, pathSample, int)
	refine = func(first, last pathSample, depth int) {
		a := geodata.GeoPoint{Longitude: first.Longitude, Latitude: first.Latitude}
		b := geodata.GeoPoint{Longitude: last.Longitude, Latitude: last.Latitude}
		middle := geodata.InterpolateGreatCircle(a, b, 0.5)
		at := first.Time.Add(last.Time.Sub(first.Time) / 2)
		if sphericalMapChordError(a, middle, b) <= 0.002 || depth >= 20 || at.Equal(first.Time) || at.Equal(last.Time) {
			result = append(result, first)
			return
		}
		point := pathSample{Time: at, Longitude: middle.Longitude, Latitude: middle.Latitude}
		refine(first, point, depth+1)
		refine(point, last, depth+1)
	}
	for i := 1; i < len(source); i++ {
		refine(source[i-1], source[i], 0)
	}
	return append(result, source[len(source)-1])
}

// Match the subdivisions on both long sides of a thin spherical triangle.
// Independent chord approximations can cross even with a small absolute error.
func sampleSphericalMapTriangle(points []geodata.GeoPoint) []geodata.GeoPoint {
	apex, shortest := 0, math.Inf(1)
	for i := range points {
		if length := solarCentralBandGeoPointDistanceKM(points[(i+1)%3], points[(i+2)%3]); length < shortest {
			apex, shortest = i, length
		}
	}
	a, b, c := points[apex], points[(apex+1)%3], points[(apex+2)%3]
	var left, right []geodata.GeoPoint
	var refine func(geodata.GeoPoint, geodata.GeoPoint, geodata.GeoPoint, geodata.GeoPoint, int)
	refine = func(a, b, c, d geodata.GeoPoint, depth int) {
		m, n := geodata.InterpolateGreatCircle(a, b, 0.5), geodata.InterpolateGreatCircle(c, d, 0.5)
		if depth >= 20 || math.Max(sphericalMapChordError(a, m, b), sphericalMapChordError(c, n, d)) <= 0.002 {
			left, right = append(left, a), append(right, c)
			return
		}
		refine(a, m, c, n, depth+1)
		refine(m, b, n, d, depth+1)
	}
	refine(a, b, a, c, 0)
	left, right = append(left, b), append(right, c)
	left, right = alignSphericalMapTriangleSides(left, right)
	left = left[:len(left)-1]
	left = appendSphericalMapArc(left, b, c, 0)
	left = append(left, c)
	for i := len(right) - 2; i > 0; i-- {
		left = append(left, right[i])
	}
	return left
}

func alignSphericalMapTriangleSides(left, right []geodata.GeoPoint) ([]geodata.GeoPoint, []geodata.GeoPoint) {
	var longitudes []float64
	for _, side := range [][]geodata.GeoPoint{left, right} {
		for i := range side {
			if i > 0 {
				side[i].Longitude = side[i-1].Longitude + math.Remainder(side[i].Longitude-side[i-1].Longitude, 360)
			}
			longitudes = append(longitudes, side[i].Longitude)
		}
	}
	sort.Float64s(longitudes)
	align := func(side []geodata.GeoPoint) []geodata.GeoPoint {
		result := []geodata.GeoPoint{side[0]}
		for i := 1; i < len(side); i++ {
			a, b := side[i-1], side[i]
			lo, hi := math.Min(a.Longitude, b.Longitude), math.Max(a.Longitude, b.Longitude)
			first, last := sort.SearchFloat64s(longitudes, lo), sort.SearchFloat64s(longitudes, hi)
			for j := first; j < last; j++ {
				index := j
				if a.Longitude > b.Longitude {
					index = first + last - 1 - j
				}
				lon := longitudes[index]
				if lon <= lo+1e-12 || lon >= hi-1e-12 || index > 0 && lon == longitudes[index-1] {
					continue
				}
				// The great-circle plane intersects each intermediate meridian once.
				lonA, latA, lonB, latB := a.Longitude*math.Pi/180, a.Latitude*math.Pi/180, b.Longitude*math.Pi/180, b.Latitude*math.Pi/180
				x, y, z := math.Cos(latA)*math.Cos(lonA), math.Cos(latA)*math.Sin(lonA), math.Sin(latA)
				u, v, w := math.Cos(latB)*math.Cos(lonB), math.Cos(latB)*math.Sin(lonB), math.Sin(latB)
				nx, ny, nz := y*w-z*v, z*u-x*w, x*v-y*u
				lat := math.Atan(-(nx*math.Cos(lon*math.Pi/180)+ny*math.Sin(lon*math.Pi/180))/nz) * 180 / math.Pi
				result = append(result, geodata.GeoPoint{Longitude: lon, Latitude: lat})
			}
			result = append(result, b)
		}
		return result
	}
	return align(left), align(right)
}
