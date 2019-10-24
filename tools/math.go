package tools

import (
	"fmt"
	"math"
	"strconv"
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

func FloatRound(f float64, n int) float64 {
	format := "%." + strconv.Itoa(n) + "f"
	res, _ := strconv.ParseFloat(fmt.Sprintf(format, f), 64)
	return res
}

func Limit360(x float64) float64 {
	for x > 360 {
		x -= 360
	}
	for x < 0 {
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
