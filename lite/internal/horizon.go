package internal

import "math"

const (
	earthEquatorialRadius = 6378137.0
	earthPolarRadius      = 6356752.3
)

// HeightDegreeByLat 高度对应地平线下角 / horizon dip at latitude.
func HeightDegreeByLat(height, lat float64) float64 {
	radius := geocentricRadius(lat)
	return math.Acos(radius/(radius+height)) * 180 / math.Pi
}

func geocentricRadius(lat float64) float64 {
	cosLat := math.Cos(lat * math.Pi / 180)
	sinLat := math.Sin(lat * math.Pi / 180)
	a := earthEquatorialRadius * earthEquatorialRadius * cosLat
	a *= a
	b := earthPolarRadius * earthPolarRadius * sinLat
	b *= b
	c := earthEquatorialRadius * cosLat
	c *= c
	d := earthPolarRadius * sinLat
	d *= d
	return math.Sqrt((a + b) / (c + d))
}
