package tools

import (
	"math"
)

func Sin(x float64) float64 {
	return (math.Sin(x * math.Pi / 180.00000))
}

func Cos(x float64) float64 {
	return (math.Cos(x * math.Pi / 180.00000))
}

func Tan(x float64) float64 {
	return (math.Tan(x * math.Pi / 180.00000))
}

func ArcSin(x float64) float64 {
	return (math.Asin(x) / math.Pi * 180.00000)
}

func ArcCos(x float64) float64 {
	return (math.Acos(x) / math.Pi * 180.00000)
}

func ArcTan(x float64) float64 {
	return (math.Atan(x) / math.Pi * 180.00000)
}

// ArcTan2 计算两个变量的反正切并转换为角度，处理所有象限
func ArcTan2(y, x float64) float64 {
	angle := math.Atan2(y, x) * 180.0 / math.Pi
	if angle < 0 {
		angle += 360.0
	}
	return angle
}

func FloatRound(f float64, n int) float64 {
	p := math.Pow10(n)
	return math.Floor(f*p+0.5) / p
}

func Limit360(x float64) float64 {
	x = math.Mod(x, 360)
	if x < 0 {
		x += 360
	}
	return x
}

func FR(f float64) float64 {
	return FloatRound(f, 14)
}

func Abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
