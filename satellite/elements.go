package satellite

import (
	"errors"
	"math"
	"time"
)

// ErrInvalidElements 轨道根数无效 / invalid orbital elements.
var ErrInvalidElements = errors.New("invalid orbital elements")

// Elements 卫星平均轨道根数 / satellite mean orbital elements.
type Elements struct {
	Name                    string
	CatalogNumber           int
	InternationalDesignator string
	Classification          string
	Epoch                   time.Time // Epoch 根数历元，UTC / elements epoch in UTC.
	MeanMotionDot           float64   // MeanMotionDot 一阶字段，rev/day² / first derivative field in rev/day squared.
	MeanMotionDDot          float64   // MeanMotionDDot 二阶字段，rev/day³ / second derivative field in rev/day cubed.
	BStar                   float64   // BStar 阻力项，地球半径⁻¹ / drag term in inverse Earth radii.
	Inclination             float64   // Inclination 轨道倾角，度 / orbital inclination in degrees.
	RightAscension          float64   // RightAscension 升交点赤经，度 / ascending node right ascension in degrees.
	Eccentricity            float64
	ArgumentOfPerigee       float64 // ArgumentOfPerigee 近地点幅角，度 / argument of perigee in degrees.
	MeanAnomaly             float64 // MeanAnomaly 平近点角，度 / mean anomaly in degrees.
	MeanMotion              float64 // MeanMotion 平均运动，rev/day / mean motion in revolutions per day.
	ElementSetNumber        int
	RevolutionNumber        int
	EphemerisType           int
}

// Age 根数历元到指定时刻的时长 / elapsed duration from the elements epoch.
func (e Elements) Age(t time.Time) time.Duration { return t.Sub(e.Epoch) }

func validElements(e Elements) bool {
	if e.CatalogNumber <= 0 || e.Epoch.IsZero() || (e.Classification != "U" && e.Classification != "C" && e.Classification != "S") || e.ElementSetNumber < 0 || e.RevolutionNumber < 0 || (e.EphemerisType != 0 && e.EphemerisType != 2) {
		return false
	}
	for _, v := range []float64{e.MeanMotionDot, e.MeanMotionDDot, e.BStar, e.Inclination, e.RightAscension, e.Eccentricity, e.ArgumentOfPerigee, e.MeanAnomaly, e.MeanMotion} {
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return false
		}
	}
	return e.Inclination >= 0 && e.Inclination <= 180 && e.RightAscension >= 0 && e.RightAscension < 360 && e.Eccentricity >= 0 && e.Eccentricity < 1 && e.ArgumentOfPerigee >= 0 && e.ArgumentOfPerigee < 360 && e.MeanAnomaly >= 0 && e.MeanAnomaly < 360 && e.MeanMotion > 0
}
