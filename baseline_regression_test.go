// 根包回归：testdata 的位级基线与 n 截断契约（n<0 用全项，n>=0 保留约 n 个主项）。
package astro_test

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"testing"
	"time"

	"github.com/starainrt/astro/basic"
	"github.com/starainrt/astro/moon"
	"github.com/starainrt/astro/planet"
	"github.com/starainrt/astro/sun"

	"github.com/starainrt/astro/internal/ephemtest"
)

type baselinePlanetSnapshot struct {
	Name    string `json:"name"`
	XT      int    `json:"xt"`
	LonBits uint64 `json:"lon_bits"`
	LatBits uint64 `json:"lat_bits"`
	RadBits uint64 `json:"rad_bits"`
}

type baselineMoonSnapshot struct {
	LonBits uint64 `json:"lon_bits"`
	LatBits uint64 `json:"lat_bits"`
	DisBits uint64 `json:"dis_bits"`
}

type baselineSample struct {
	UTC     string                   `json:"utc"`
	TTJD    float64                  `json:"tt_jd"`
	Planets []baselinePlanetSnapshot `json:"planets"`
	Moon    baselineMoonSnapshot     `json:"moon"`
}

func loadBaselineSamples(t *testing.T) []baselineSample {
	t.Helper()

	data, err := os.ReadFile("testdata/planet_moon_baseline.json")
	if err != nil {
		t.Fatal(err)
	}

	var samples []baselineSample
	if err := json.Unmarshal(data, &samples); err != nil {
		t.Fatal(err)
	}
	if len(samples) == 0 {
		t.Fatal("empty baseline samples")
	}
	return samples
}

// truncationTermCounts 是截断检验抽取的 n 值。
var truncationTermCounts = []int{2, 4, 8, 16}

// angularDifference 两个角度量的最小夹角，单位度。
func angularDifference(a, b float64) float64 {
	d := math.Mod(a-b, 360)
	if d > 180 {
		d -= 360
	}
	if d < -180 {
		d += 360
	}
	return math.Abs(d)
}

// truncationCase 是一个截断探针：err 给出第 sample 个采样点在截断项数 n 下相对全项 n<0 的偏差。
type truncationCase struct {
	name string
	tol  []float64
	err  func(sample, n int) float64
}

// scalarTruncation 构造标量探针；jd 按采样索引取值，angular 为真时偏差取角度最小夹角。
func scalarTruncation(name string, jd func(index int) float64, f func(float64, int) float64, angular bool, tol []float64) truncationCase {
	return truncationCase{
		name: name,
		tol:  tol,
		err: func(sample, n int) float64 {
			got := f(jd(sample), n)
			full := f(jd(sample), -1)
			if angular {
				return angularDifference(got, full)
			}
			return math.Abs(got - full)
		},
	}
}

// pairTruncation 构造经纬成对返回的探针，取两个分量偏差的较大者。
func pairTruncation(name string, jd func(index int) float64, f func(float64, int) (float64, float64), tol []float64) truncationCase {
	return truncationCase{
		name: name,
		tol:  tol,
		err: func(sample, n int) float64 {
			gotA, gotB := f(jd(sample), n)
			fullA, fullB := f(jd(sample), -1)
			return math.Max(angularDifference(gotA, fullA), angularDifference(gotB, fullB))
		},
	}
}

// dateTruncation 构造以 time.Time 为入口的探针，偏差按角度最小夹角计。
func dateTruncation(name string, date func(index int) time.Time, f func(time.Time, int) float64, tol []float64) truncationCase {
	return truncationCase{
		name: name,
		tol:  tol,
		err: func(sample, n int) float64 {
			return angularDifference(f(date(sample), n), f(date(sample), -1))
		},
	}
}

// baselineDates 解析基线样本的 UTC 时刻。
func baselineDates(t *testing.T, samples []baselineSample) []time.Time {
	t.Helper()

	dates := make([]time.Time, len(samples))
	for i, sample := range samples {
		parsed, err := time.Parse(time.RFC3339Nano, sample.UTC)
		if err != nil {
			t.Fatal(err)
		}
		dates[i] = parsed
	}
	return dates
}

// assertTruncationConverges 逐采样点检验误差上限，并要求最坏误差随 n 不增、多数采样点严格下降。
// 单点截断误差不是嵌套部分和，允许同量级抖动，因此“不增”约束的是采样集上的误差包络。
func assertTruncationConverges(t *testing.T, cases []truncationCase, samples int) {
	t.Helper()

	for _, tc := range cases {
		envelope := make([]float64, len(truncationTermCounts))
		strict := 0
		for sample := 0; sample < samples; sample++ {
			first, last := 0.0, 0.0
			for i, n := range truncationTermCounts {
				e := tc.err(sample, n)
				if e > tc.tol[i] {
					t.Fatalf("%s: sample %d truncation error %.6g at n=%d exceeds %.6g", tc.name, sample, e, n, tc.tol[i])
				}
				if e > envelope[i] {
					envelope[i] = e
				}
				if i == 0 {
					first = e
				}
				last = e
			}
			if last < first {
				strict++
			}
		}
		for i := 1; i < len(envelope); i++ {
			if envelope[i] > envelope[i-1] {
				t.Fatalf("%s: worst truncation error grows from %.6g at n=%d to %.6g at n=%d",
					tc.name, envelope[i-1], truncationTermCounts[i-1], envelope[i], truncationTermCounts[i])
			}
		}
		if 2*strict < samples {
			t.Fatalf("%s: truncation error shrank with n at only %d of %d samples", tc.name, strict, samples)
		}
	}
}

func TestPlanetMoonBaselineRegression(t *testing.T) {
	ephemtest.PinLegacy(t)
	samples := loadBaselineSamples(t)
	for _, sample := range samples {
		for _, body := range sample.Planets {
			gotLon := planet.WherePlanet(body.XT, 0, sample.TTJD)
			if math.Float64bits(gotLon) != body.LonBits {
				t.Fatalf("%s lon regression at %s", body.Name, sample.UTC)
			}

			gotLat := planet.WherePlanet(body.XT, 1, sample.TTJD)
			if math.Float64bits(gotLat) != body.LatBits {
				t.Fatalf("%s lat regression at %s", body.Name, sample.UTC)
			}

			gotRad := planet.WherePlanet(body.XT, 2, sample.TTJD)
			if math.Float64bits(gotRad) != body.RadBits {
				t.Fatalf("%s rad regression at %s", body.Name, sample.UTC)
			}
		}

		if math.Float64bits(basic.HMoonTrueLo(sample.TTJD)) != sample.Moon.LonBits {
			t.Fatalf("moon lon regression at %s", sample.UTC)
		}
		if math.Float64bits(basic.HMoonTrueBo(sample.TTJD)) != sample.Moon.LatBits {
			t.Fatalf("moon lat regression at %s", sample.UTC)
		}
		if math.Float64bits(basic.HMoonAway(sample.TTJD)) != sample.Moon.DisBits {
			t.Fatalf("moon distance regression at %s", sample.UTC)
		}
	}
}

// planetTruncationName 与 planetTruncationTolerance 的下标同 WherePlanetN 的 xt。
var planetTruncationName = []string{"earth", "mercury", "venus", "mars", "jupiter", "saturn", "uranus", "neptune"}

// planetTruncationTolerance 是 zn=黄经/黄纬（度）、日心距（AU）两种量纲的上限，取实测最坏值的 5 倍以上。
var planetTruncationTolerance = [][]float64{
	{20, 2, 0.5, 0.1},
	{5, 0.5, 0.05, 0.01},
	{0.3, 0.1, 0.05, 0.02},
}

func planetMoonTruncationCases(samples []baselineSample) []truncationCase {
	jdOf := func(index int) float64 { return samples[index].TTJD }

	wherePlanet := func(xt, zn int) truncationCase {
		return scalarTruncation(
			fmt.Sprintf("planet.WherePlanetN(%s, zn=%d)", planetTruncationName[xt], zn),
			jdOf,
			func(jd float64, n int) float64 { return planet.WherePlanetN(xt, zn, jd, n) },
			zn != 2,
			planetTruncationTolerance[zn],
		)
	}

	cases := make([]truncationCase, 0, 27)
	for xt := 0; xt <= 7; xt++ {
		for zn := 0; zn <= 2; zn++ {
			cases = append(cases, wherePlanet(xt, zn))
		}
	}
	return append(cases,
		scalarTruncation("basic.HMoonTrueLoN", jdOf, basic.HMoonTrueLoN, true, []float64{5, 2.5, 1, 0.25}),
		scalarTruncation("basic.HMoonTrueBoN", jdOf, basic.HMoonTrueBoN, true, []float64{3, 0.75, 0.15, 0.06}),
		scalarTruncation("basic.HMoonAwayN", jdOf, basic.HMoonAwayN, false, []float64{40000, 7000, 2500, 400}),
	)
}

func TestPlanetMoonTruncationConvergesToFullSeries(t *testing.T) {
	ephemtest.PinLegacy(t)
	samples := loadBaselineSamples(t)
	assertTruncationConverges(t, planetMoonTruncationCases(samples), len(samples))
}

func publicSunMoonTruncationCases(t *testing.T, samples []baselineSample) []truncationCase {
	t.Helper()

	dates := baselineDates(t, samples)
	dateOf := func(index int) time.Time { return dates[index] }
	return []truncationCase{
		dateTruncation("sun.TrueLoN", dateOf, sun.TrueLoN, []float64{0.15, 0.05, 0.02, 0.006}),
		dateTruncation("sun.TrueBoN", dateOf, sun.TrueBoN, []float64{1e-3, 5e-4, 3e-4, 1.5e-4}),
		dateTruncation("moon.TrueLoN", dateOf, moon.TrueLoN, []float64{5, 2.5, 1, 0.25}),
		dateTruncation("moon.TrueBoN", dateOf, moon.TrueBoN, []float64{3, 0.75, 0.15, 0.06}),
	}
}

func TestPublicSunMoonTruncationConvergesToFullSeries(t *testing.T) {
	ephemtest.PinLegacy(t)
	samples := loadBaselineSamples(t)
	assertTruncationConverges(t, publicSunMoonTruncationCases(t, samples), len(samples))
}
