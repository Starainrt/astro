package basic

import (
	. "github.com/starainrt/astro/tools"
	"math"
)

//地球常数

const (
	EARTH_EQUATORIAL_RADIUS float64 = 6378137.0
	EARTH_POLAR_RADIUS      float64 = 6356752.3
	EARTH_AVERAGE_RADIUS    float64 = 6371393.0
)

// HeightDistance 高度与地平线距离的关系（单位：米）
func HeightDistance(height float64) float64 {
	return math.Acos((EARTH_AVERAGE_RADIUS)/(EARTH_AVERAGE_RADIUS+height)) * EARTH_AVERAGE_RADIUS
}

// HeightDistance 高度（单位：米）与地平线下角度的关系（单位：度）
func HeightDegree(height float64) float64 {
	return math.Acos((EARTH_AVERAGE_RADIUS)/(EARTH_AVERAGE_RADIUS+height)) * 180 / math.Pi / 2
}

// HeightDistanceByLat 不同纬度下高度与地平线距离的关系（单位：米）
func HeightDistanceByLat(height, lat float64) float64 {
	raduis := GeocentricRadius(lat)
	return math.Acos((raduis)/(raduis+height)) * raduis
}

// HeightDegreeByLat 不同纬度下高度（单位：米）与地平线下角度的关系（单位：度）
func HeightDegreeByLat(height, lat float64) float64 {
	raduis := GeocentricRadius(lat)
	return math.Acos((raduis)/(raduis+height)) * 180 / math.Pi / 2
}

// GeocentricRadius 地心直径与纬度的关系
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
