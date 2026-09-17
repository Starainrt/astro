package tools

import (
	"math"
)

// Sin 正弦，角度参数单位为度 / sine; the angle is in degrees.
func Sin(x float64) float64 {
	return (math.Sin(x * math.Pi / 180.00000))
}

// Cos 余弦，角度参数单位为度 / cosine; the angle is in degrees.
func Cos(x float64) float64 {
	return (math.Cos(x * math.Pi / 180.00000))
}

// Tan 正切，角度参数单位为度 / tangent; the angle is in degrees.
func Tan(x float64) float64 {
	return (math.Tan(x * math.Pi / 180.00000))
}

// ArcSin 反正弦，结果单位为度 / arc sine; the result is in degrees.
func ArcSin(x float64) float64 {
	return (math.Asin(x) / math.Pi * 180.00000)
}

// ArcCos 反余弦，结果单位为度 / arc cosine; the result is in degrees.
func ArcCos(x float64) float64 {
	return (math.Acos(x) / math.Pi * 180.00000)
}

// ArcTan 反正切，结果单位为度 / arc tangent; the result is in degrees.
func ArcTan(x float64) float64 {
	return (math.Atan(x) / math.Pi * 180.00000)
}

// ArcTan2 两变量反正切，结果单位为度并归一到 [0,360) / two-argument arc tangent; the result is in degrees and normalized to [0,360).
func ArcTan2(y, x float64) float64 {
	angle := math.Atan2(y, x) * 180.0 / math.Pi
	if angle < 0 {
		angle += 360.0
	}
	return angle
}

// FloatRound 四舍五入到 n 位小数，实现为 floor(f*10^n+0.5) / round to n decimal places via floor(f*10^n+0.5).
func FloatRound(f float64, n int) float64 {
	p := math.Pow10(n)
	return math.Floor(f*p+0.5) / p
}

// Limit360 把角度归一到 [0,360) / normalize an angle into [0,360).
func Limit360(x float64) float64 {
	x = math.Mod(x, 360)
	if x < 0 {
		x += 360
	}
	return x
}

// FR 四舍五入到 14 位小数 / round to 14 decimal places.
func FR(f float64) float64 {
	return FloatRound(f, 14)
}

// Abs 取 int 的绝对值，只接受 int，与 math.Abs 同名且不支持浮点 / absolute value of an int; same name as math.Abs but accepts only int.
func Abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
