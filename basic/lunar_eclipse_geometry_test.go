package basic

import (
	"math"
	"testing"
	"time"
)

// Axis 与 Gamma 是同一个量的两种刻度：Gamma × 月球处的地球视差 = Axis。
func TestLunarEclipseShadowGeometryAxisMatchesGamma(t *testing.T) {
	for _, date := range []time.Time{
		time.Date(2025, time.March, 14, 0, 0, 0, 0, time.UTC),
		time.Date(2026, time.March, 3, 0, 0, 0, 0, time.UTC),
		time.Date(2028, time.July, 6, 0, 0, 0, 0, time.UTC),
	} {
		result := LunarEclipse(UTC2TT(Date2JD(date)))
		geometry := LunarEclipseShadowGeometryAt(result.Maximum)
		// 影几何内部用的是小角近似的地球视差（R⊕/月距），这里必须用同一形式，否则二级差会露出来。
		earthParallaxDegrees := (1 / geometry.MoonDistanceEarthRadii) * 180 / math.Pi
		if delta := geometry.Gamma*earthParallaxDegrees - geometry.AxisDegrees; math.Abs(delta) > 1e-9 {
			t.Fatalf("%s: Gamma×parallax - Axis = %g", date.Format("2006-01-02"), delta)
		}
	}
}

// 食分公式反解出的月心到影轴距离必须等于 Axis：这条把影几何与食分求解器对起来。
// 食分 = (影半径 + 月视半径 − Axis) / (2 × 月视半径)。
func TestLunarEclipseMagnitudeInvertsToAxis(t *testing.T) {
	for _, date := range []time.Time{
		time.Date(2025, time.March, 14, 0, 0, 0, 0, time.UTC),
		time.Date(2026, time.March, 3, 0, 0, 0, 0, time.UTC),
		time.Date(2028, time.July, 6, 0, 0, 0, 0, time.UTC),
		time.Date(2020, time.November, 30, 0, 0, 0, 0, time.UTC),
	} {
		result := LunarEclipse(UTC2TT(Date2JD(date)))
		geometry := LunarEclipseShadowGeometryAt(result.Maximum)
		moonSemidiameter := MoonSemidiameter(result.Maximum) / 3600
		fromUmbral := geometry.UmbralRadiusDegrees + moonSemidiameter -
			2*moonSemidiameter*result.Magnitude
		fromPenumbral := geometry.PenumbralRadiusDegrees + moonSemidiameter -
			2*moonSemidiameter*result.PenumbralMagnitude
		t.Logf("%s %s: Axis=%.4f 本影反推=%.4f 半影反推=%.4f",
			date.Format("2006-01-02"), result.Type, geometry.AxisDegrees, fromUmbral, fromPenumbral)
		if math.Abs(fromUmbral-geometry.AxisDegrees) > 0.002 {
			t.Fatalf("%s: umbral magnitude implies axis %.4f, geometry gives %.4f",
				date.Format("2006-01-02"), fromUmbral, geometry.AxisDegrees)
		}
		if math.Abs(fromPenumbral-geometry.AxisDegrees) > 0.002 {
			t.Fatalf("%s: penumbral magnitude implies axis %.4f, geometry gives %.4f",
				date.Format("2006-01-02"), fromPenumbral, geometry.AxisDegrees)
		}
	}
}
