package basic

import (
	"math"

	. "github.com/starainrt/astro/tools"
)

//地球常数

const (
	EARTH_EQUATORIAL_RADIUS float64 = 6378137.0
	EARTH_POLAR_RADIUS      float64 = 6356752.3
	EARTH_AVERAGE_RADIUS    float64 = 6371008.8
)

// HeightDistance 高度与地平线距离的关系（单位：米）
func HeightDistance(height float64) float64 {
	return math.Acos((EARTH_AVERAGE_RADIUS)/(EARTH_AVERAGE_RADIUS+height)) * EARTH_AVERAGE_RADIUS
}

// HeightDistance 高度（单位：米）与地平线下角度的关系（单位：度）
func HeightDegree(height float64) float64 {
	return math.Acos((EARTH_AVERAGE_RADIUS)/(EARTH_AVERAGE_RADIUS+height)) * 180 / math.Pi
}

// HeightDistanceByLat 不同纬度下高度与地平线距离的关系（单位：米）
func HeightDistanceByLat(height, lat float64) float64 {
	radius := GeocentricRadius(lat)
	return math.Acos((radius)/(radius+height)) * radius
}

// HeightDegreeByLat 不同纬度下高度（单位：米）与地平线下角度的关系（单位：度）
func HeightDegreeByLat(height, lat float64) float64 {
	radius := GeocentricRadius(lat)
	return math.Acos((radius)/(radius+height)) * 180 / math.Pi
}

// GeocentricRadius 给定纬度处的地心半径，单位米 / geocentric radius at the given latitude, in metres.
func GeocentricRadius(lat float64) float64 {
	a := (EARTH_EQUATORIAL_RADIUS * EARTH_EQUATORIAL_RADIUS * Cos(lat))
	a *= a
	b := (EARTH_POLAR_RADIUS * EARTH_POLAR_RADIUS * Sin(lat))
	b *= b
	c := (EARTH_EQUATORIAL_RADIUS * Cos(lat))
	c *= c
	d := (EARTH_POLAR_RADIUS * Sin(lat))
	d *= d
	return math.Sqrt((a + b) / (c + d))
}
