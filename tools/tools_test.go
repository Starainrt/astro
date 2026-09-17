// tools 的对外契约：三角函数以度为角度单位，反三角以度为结果单位，取整为 floor 语义。
package tools_test

import (
	"math"
	"testing"

	"github.com/starainrt/astro/tools"
)

// 度数约定：与 math 的弧度结果必须显著不同，否则说明换算被删掉了。
func TestTrigTakesDegrees(t *testing.T) {
	if math.Abs(tools.Sin(1)-0.01745240643728351) > 1e-15 {
		t.Fatalf("Sin(1) 应为 sin 1° = 0.01745240643728351，得到 %v", tools.Sin(1))
	}
	if math.Abs(tools.Sin(1)-math.Sin(1)) < 0.8 {
		t.Fatal("Sin 的入参似乎被当成弧度")
	}
	if math.Abs(tools.ArcSin(0.5)-30) > 1e-12 {
		t.Fatalf("ArcSin(0.5) 应为 30 度，得到 %v", tools.ArcSin(0.5))
	}
}

func TestTrigExactAngles(t *testing.T) {
	cases := []struct {
		name string
		got  float64
		want float64
	}{
		{"Sin(0)", tools.Sin(0), 0},
		{"Sin(30)", tools.Sin(30), 0.5},
		{"Sin(90)", tools.Sin(90), 1},
		{"Sin(180)", tools.Sin(180), 0},
		{"Sin(270)", tools.Sin(270), -1},
		{"Cos(0)", tools.Cos(0), 1},
		{"Cos(60)", tools.Cos(60), 0.5},
		{"Cos(90)", tools.Cos(90), 0},
		{"Cos(180)", tools.Cos(180), -1},
		{"Tan(0)", tools.Tan(0), 0},
		{"Tan(45)", tools.Tan(45), 1},
		{"Tan(135)", tools.Tan(135), -1},
	}
	for _, tc := range cases {
		if math.Abs(tc.got-tc.want) > 1e-14 {
			t.Fatalf("%s = %v，期望 %v", tc.name, tc.got, tc.want)
		}
	}
	if math.Abs(tools.Tan(90)) < 1e15 || math.IsInf(tools.Tan(90), 0) {
		t.Fatalf("Tan(90) 应为极大有限值，得到 %v", tools.Tan(90))
	}
}

func TestInverseTrigReturnsDegrees(t *testing.T) {
	cases := []struct {
		name string
		got  float64
		want float64
	}{
		{"ArcSin(-1)", tools.ArcSin(-1), -90},
		{"ArcSin(0)", tools.ArcSin(0), 0},
		{"ArcSin(1)", tools.ArcSin(1), 90},
		{"ArcCos(-1)", tools.ArcCos(-1), 180},
		{"ArcCos(0)", tools.ArcCos(0), 90},
		{"ArcCos(0.5)", tools.ArcCos(0.5), 60},
		{"ArcCos(1)", tools.ArcCos(1), 0},
		{"ArcTan(-1)", tools.ArcTan(-1), -45},
		{"ArcTan(0)", tools.ArcTan(0), 0},
		{"ArcTan(1)", tools.ArcTan(1), 45},
	}
	for _, tc := range cases {
		if math.Abs(tc.got-tc.want) > 1e-12 {
			t.Fatalf("%s = %v，期望 %v", tc.name, tc.got, tc.want)
		}
	}
	for _, deg := range []float64{-89.5, -45, -0.25, 0, 30, 44.9, 89.5} {
		if got := tools.ArcSin(tools.Sin(deg)); math.Abs(got-deg) > 1e-11 {
			t.Fatalf("ArcSin(Sin(%v)) = %v", deg, got)
		}
	}
}

func TestArcTan2Quadrants(t *testing.T) {
	cases := []struct {
		y, x float64
		want float64
	}{
		{0, 1, 0},
		{1, 1, 45},
		{1, 0, 90},
		{1, -1, 135},
		{0, -1, 180},
		{-1, -1, 225},
		{-1, 0, 270},
		{-1, 1, 315},
		{-0.0, -1, 180},
	}
	for _, tc := range cases {
		if got := tools.ArcTan2(tc.y, tc.x); math.Abs(got-tc.want) > 1e-12 {
			t.Fatalf("ArcTan2(%v, %v) = %v，期望 %v", tc.y, tc.x, got, tc.want)
		}
	}
	for _, tc := range cases {
		got := tools.ArcTan2(tc.y, tc.x)
		if got < 0 || got >= 360 {
			t.Fatalf("ArcTan2(%v, %v) = %v 未落在 [0,360)", tc.y, tc.x, got)
		}
	}
}

func TestLimit360(t *testing.T) {
	cases := []struct {
		in   float64
		want float64
	}{
		{0, 0},
		{359.5, 359.5},
		{360, 0},
		{720, 0},
		{-1, 359},
		{-360, 0},
		{-720.5, 359.5},
		{1e9, 280},
		{-1e9, 80},
	}
	for _, tc := range cases {
		if got := tools.Limit360(tc.in); math.Abs(got-tc.want) > 1e-9 {
			t.Fatalf("Limit360(%v) = %v，期望 %v", tc.in, got, tc.want)
		}
	}
	if !math.IsNaN(tools.Limit360(math.NaN())) {
		t.Fatal("Limit360(NaN) 应为 NaN")
	}
}

func TestFloatRoundFloorSemantics(t *testing.T) {
	cases := []struct {
		f    float64
		n    int
		want float64
	}{
		{0.5, 0, 1},
		{1.5, 0, 2},
		{2.5, 0, 3},
		{-0.5, 0, 0},
		{-1.5, 0, -1},
		{-2.5, 0, -2},
		{1.25, 1, 1.3},
		{-1.25, 1, -1.2},
		{3.14159, 2, 3.14},
		{2.71828, 3, 2.718},
		{1, 5, 1},
	}
	for _, tc := range cases {
		if got := tools.FloatRound(tc.f, tc.n); got != tc.want {
			t.Fatalf("FloatRound(%v, %d) = %v，期望 %v", tc.f, tc.n, got, tc.want)
		}
	}
	if got := tools.FR(1.0 / 3.0); got != 0.33333333333333 {
		t.Fatalf("FR(1/3) = %v，期望 14 位", got)
	}
	if tools.FR(1.0/3.0) != tools.FloatRound(1.0/3.0, 14) {
		t.Fatal("FR 应为 FloatRound(f, 14)")
	}
}

func TestAbs(t *testing.T) {
	for _, tc := range []struct {
		in   int
		want int
	}{{0, 0}, {7, 7}, {-7, 7}, {-1, 1}} {
		if got := tools.Abs(tc.in); got != tc.want {
			t.Fatalf("Abs(%d) = %d，期望 %d", tc.in, got, tc.want)
		}
	}
}

func TestFormatDegreesAndHours(t *testing.T) {
	cases := []struct {
		val   float64
		typed uint8
		want  string
	}{
		{0, 0, "0\u00b00\u20320.00\u2033"},
		{12.3456, 0, "12\u00b020\u203244.16\u2033"},
		{-12.3456, 0, "-12\u00b020\u203244.16\u2033"},
		{359.999, 0, "359\u00b059\u203256.40\u2033"},
		{12.3456, 1, "12h20m44.16s"},
		{23.5, 1, "23h30m0.00s"},
		{0, 1, "0h0m0.00s"},
	}
	for _, tc := range cases {
		if got := tools.Format(tc.val, tc.typed); got != tc.want {
			t.Fatalf("Format(%v, %d) = %q，期望 %q", tc.val, tc.typed, got, tc.want)
		}
	}
}
