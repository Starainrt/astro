package basic

import (
	"math"
	"testing"
	"time"

	. "github.com/starainrt/astro/tools"
)

// 本文件是「行星相位事件（留/合/冲/方照）」的顺序不变量回归。
//
// 背景：类型化「留」的结果必须满足
//   Next*(q) >= q - tol、Last*(q) <= q + tol、返回的是真值事件、且不得跳过更近的同类型事件。
// 2026-09 的 review 发现水星类型化「留」会跳过一个下合，导致 Last* 返回未来
// （例如查询 TT 2008-01-22 时 LastMercuryProgradeToRetrograde 返回 2008-01-28）。
// 真值来自对导出星历的独立二分扫描，不使用被测的 Next*/Last*，因此能独立复现该类缺陷。

const phaseInvariantToleranceDay = 0.6 / 86400.0 // 站事件同刻容差 0.5 s + 余量

type phaseTruthStation struct {
	jd  float64
	p2r bool
}

func phaseRate(ra func(float64) float64, jd float64) float64 {
	sub := ra(jd+0.01) - ra(jd-0.01)
	if sub > 180 {
		sub -= 360
	}
	if sub < -180 {
		sub += 360
	}
	return sub / 0.02
}

func phaseTruthStations(ra func(float64) float64, jd0, jd1, step float64) []phaseTruthStation {
	rate := func(jd float64) float64 { return phaseRate(ra, jd) }
	var out []phaseTruthStation
	prevJD, prev := jd0, rate(jd0)
	for jd := jd0 + step; jd <= jd1; jd += step {
		cur := rate(jd)
		if prev*cur < 0 {
			left, right := prevJD, jd
			for i := 0; i < 100; i++ {
				middle := (left + right) / 2
				if middle == left || middle == right {
					break
				}
				if (prev < 0) == (rate(middle) < 0) {
					left = middle
					continue
				}
				right = middle
			}
			out = append(out, phaseTruthStation{jd: (left + right) / 2, p2r: prev > 0})
		}
		prevJD, prev = jd, cur
	}
	return out
}

type phaseStationCase struct {
	name    string
	ra      func(float64) float64
	nextP2R func(float64) float64
	lastP2R func(float64) float64
	nextR2P func(float64) float64
	lastR2P func(float64) float64
	step    float64
	// radius 真值扫描半径（天）：必须覆盖「查询在站后几秒 ⇒ 答案是下一个同名站」的距离。
	radius float64
	// queryRadius 参与构造查询的站距中心的最大距离（天）。
	queryRadius float64
}

func phaseStationCases() []phaseStationCase {
	return []phaseStationCase{
		{"Mercury", MercuryApparentRa, NextMercuryProgradeToRetrograde, LastMercuryProgradeToRetrograde,
			NextMercuryRetrogradeToPrograde, LastMercuryRetrogradeToPrograde, 0.25, 700, 450},
		{"Mars", MarsApparentRa, NextMarsProgradeToRetrograde, LastMarsProgradeToRetrograde,
			NextMarsRetrogradeToPrograde, LastMarsRetrogradeToPrograde, 0.5, 1100, 250},
		{"Venus", VenusApparentRa, NextVenusProgradeToRetrograde, LastVenusProgradeToRetrograde,
			NextVenusRetrogradeToPrograde, LastVenusRetrogradeToPrograde, 0.5, 900, 250},
		{"Jupiter", JupiterApparentRa, NextJupiterProgradeToRetrograde, LastJupiterProgradeToRetrograde,
			NextJupiterRetrogradeToPrograde, LastJupiterRetrogradeToPrograde, 1, 700, 250},
	}
}

func phaseEpochTT(year int, month time.Month, day int) float64 {
	return TD2UT(Date2JDE(time.Date(year, month, day, 0, 0, 0, 0, time.UTC)), true)
}

// TestPlanetStationOrderInvariant 在固定的历史失败时点附近逐点检查顺序不变量。
func TestPlanetStationOrderInvariant(t *testing.T) {
	epochs := map[string][]time.Time{
		"Mercury": {
			time.Date(2008, 1, 22, 0, 0, 0, 0, time.UTC),
			time.Date(2006, 2, 23, 0, 0, 0, 0, time.UTC),
			time.Date(2007, 2, 6, 0, 0, 0, 0, time.UTC),
			time.Date(2007, 11, 8, 0, 0, 0, 0, time.UTC),
			time.Date(2003, 9, 27, 0, 0, 0, 0, time.UTC),
			time.Date(-209, 1, 14, 0, 0, 0, 0, time.UTC),
			time.Date(-208, 12, 11, 0, 0, 0, 0, time.UTC),
		},
		"Mars":    {time.Date(2003, 7, 30, 0, 0, 0, 0, time.UTC)},
		"Venus":   {time.Date(2025, 3, 1, 0, 0, 0, 0, time.UTC)},
		"Jupiter": {time.Date(2003, 7, 30, 0, 0, 0, 0, time.UTC)},
	}
	offsets := []float64{0, 1, -1, 30, -30, 300, -300, 3600, -3600, 86400, -86400, 10 * 86400, -10 * 86400, 30 * 86400, -30 * 86400}

	for _, tc := range phaseStationCases() {
		for _, epoch := range epochs[tc.name] {
			center := phaseEpochTT(epoch.Year(), epoch.Month(), epoch.Day())
			stations := phaseTruthStations(tc.ra, center-tc.radius, center+tc.radius, tc.step)
			if len(stations) == 0 {
				t.Fatalf("%s: no truth stations near %s", tc.name, epoch.Format("2006-01-02"))
			}
			funcs := []struct {
				label string
				fn    func(float64) float64
				next  bool
				p2r   bool
			}{
				{"NextP2R", tc.nextP2R, true, true},
				{"LastP2R", tc.lastP2R, false, true},
				{"NextR2P", tc.nextR2P, true, false},
				{"LastR2P", tc.lastR2P, false, false},
			}
			for _, f := range funcs {
				var queries []float64
				for _, st := range stations {
					if st.p2r != f.p2r || math.Abs(st.jd-center) > tc.queryRadius {
						continue
					}
					for _, off := range offsets {
						queries = append(queries, st.jd+off/86400.0)
					}
				}
				for _, q := range queries {
					gotUT := f.fn(q)
					if math.IsNaN(gotUT) {
						t.Fatalf("%s %s at %s returned NaN", tc.name, f.label,
							JDE2DateByZone(TD2UT(q, false), time.UTC, false).Format("2006-01-02 15:04:05"))
					}
					got := TD2UT(gotUT, true)
					// 1) 顺序不变量
					if f.next && got < q-phaseInvariantToleranceDay {
						t.Fatalf("%s %s at %s returned past event %s", tc.name, f.label,
							JDE2DateByZone(TD2UT(q, false), time.UTC, false).Format("2006-01-02 15:04:05"),
							JDE2DateByZone(gotUT, time.UTC, false).Format("2006-01-02 15:04:05"))
					}
					if !f.next && got > q+phaseInvariantToleranceDay {
						t.Fatalf("%s %s at %s returned future event %s", tc.name, f.label,
							JDE2DateByZone(TD2UT(q, false), time.UTC, false).Format("2006-01-02 15:04:05"),
							JDE2DateByZone(gotUT, time.UTC, false).Format("2006-01-02 15:04:05"))
					}
					// 2) 必须是真值事件
					nearest, nearestDev := math.NaN(), math.Inf(1)
					for _, st := range stations {
						if st.p2r != f.p2r {
							continue
						}
						if dev := math.Abs(st.jd - got); dev < nearestDev {
							nearest, nearestDev = st.jd, dev
						}
					}
					if nearestDev > 60.0/86400.0 {
						t.Fatalf("%s %s at %s returned non-event %.6f (nearest truth %.3f d away)", tc.name, f.label,
							JDE2DateByZone(TD2UT(q, false), time.UTC, false).Format("2006-01-02 15:04:05"), got, nearestDev)
					}
					_ = nearest
					// 3) 不得跳过更近的同类型事件
					for _, st := range stations {
						if st.p2r != f.p2r {
							continue
						}
						if f.next && st.jd > q+phaseInvariantToleranceDay && st.jd < got-60.0/86400.0 {
							t.Fatalf("%s %s at %s skipped %s", tc.name, f.label,
								JDE2DateByZone(TD2UT(q, false), time.UTC, false).Format("2006-01-02 15:04:05"),
								JDE2DateByZone(TD2UT(st.jd, false), time.UTC, false).Format("2006-01-02 15:04:05"))
						}
						if !f.next && st.jd < q-phaseInvariantToleranceDay && st.jd > got+60.0/86400.0 {
							t.Fatalf("%s %s at %s skipped %s", tc.name, f.label,
								JDE2DateByZone(TD2UT(q, false), time.UTC, false).Format("2006-01-02 15:04:05"),
								JDE2DateByZone(TD2UT(st.jd, false), time.UTC, false).Format("2006-01-02 15:04:05"))
						}
					}
				}
			}
		}
	}
}

// TestMercuryConjunctionNeverSkips 检查水星「合」搜索不会跨过更近的合
// （历史缺陷：启发式跳 + 2 天走法会走满一个会合周期，跳过一次下合）。
func TestMercuryConjunctionNeverSkips(t *testing.T) {
	delta := func(jd float64) float64 {
		sub := Limit360(MercuryApparentLo(jd) - HSunApparentLo(jd))
		if sub > 180 {
			sub -= 360
		}
		if sub < -180 {
			sub += 360
		}
		return sub
	}
	// 真值：局部细扫
	truth := func(jd0, jd1, step float64) []float64 {
		var out []float64
		prevJD, prev := jd0, delta(jd0)
		for jd := jd0 + step; jd <= jd1; jd += step {
			cur := delta(jd)
			if prev*cur < 0 {
				left, right := prevJD, jd
				for i := 0; i < 100; i++ {
					middle := (left + right) / 2
					if middle == left || middle == right {
						break
					}
					if (prev < 0) == (delta(middle) < 0) {
						left = middle
						continue
					}
					right = middle
				}
				out = append(out, (left+right)/2)
			}
			prevJD, prev = jd, cur
		}
		return out
	}
	for _, epoch := range []time.Time{
		time.Date(2008, 1, 22, 0, 0, 0, 0, time.UTC),
		time.Date(2006, 2, 23, 0, 0, 0, 0, time.UTC),
		time.Date(2007, 2, 6, 0, 0, 0, 0, time.UTC),
		time.Date(-209, 1, 14, 0, 0, 0, 0, time.UTC),
	} {
		center := phaseEpochTT(epoch.Year(), epoch.Month(), epoch.Day())
		truthEvents := truth(center-400, center+400, 0.25)
		if len(truthEvents) < 5 {
			t.Fatalf("truth scan too sparse near %s: %d", epoch.Format("2006-01-02"), len(truthEvents))
		}
		for _, q := range []float64{center, center + 0.5, center + 12, center - 12, center + 60, center - 60} {
			for _, next := range []uint8{0, 1} {
				got := TD2UT(mercuryConjunction(q, next), true)
				if math.IsNaN(got) {
					t.Fatalf("mercuryConjunction(%.6f, %d) = NaN", q, next)
				}
				want := math.NaN()
				if next == 1 {
					for _, e := range truthEvents {
						if e >= q-0.1/86400.0 {
							want = e
							break
						}
					}
				} else {
					for i := len(truthEvents) - 1; i >= 0; i-- {
						if truthEvents[i] <= q+0.1/86400.0 {
							want = truthEvents[i]
							break
						}
					}
				}
				if math.IsNaN(want) {
					continue
				}
				if math.Abs(got-want) > 60.0/86400.0 {
					t.Fatalf("mercuryConjunction(%.6f, %d) = %.6f want %.6f (%.3f d off) near %s",
						q, next, got, want, got-want, epoch.Format("2006-01-02"))
				}
			}
		}
	}
}

// TestGreatestElongationNoSkip 检查大距事件不会跳过相邻的极大。
//
// 口径说明：库内 Next*GreatestElongation 用的是「真距角」（mercuryTrueElongationN，
// 不含光行差/视位置），公开的 MercurySunElongation 是「视距角」，两者极大时刻相差约 10 分钟。
// 因此这里用视距角的极值作为参照，只做「不跳事件 / 顺序 / 是极值附近」的判定，
// 匹配容差取 30 分钟，足以覆盖口径差又远小于任何真实跳事件（数十天）。
func TestGreatestElongationNoSkip(t *testing.T) {
	type elongCase struct {
		name     string
		elongate func(float64) float64
		next     func(float64) float64
		last     func(float64) float64
	}
	cases := []elongCase{
		{"Mercury", MercurySunElongation, NextMercuryGreatestElongation, LastMercuryGreatestElongation},
		{"Venus", VenusSunElongation, NextVenusGreatestElongation, LastVenusGreatestElongation},
	}
	jd0 := phaseEpochTT(2024, 1, 1)
	jd1 := phaseEpochTT(2029, 1, 1)
	for _, tc := range cases {
		// 真值：视距角的局部极大
		var maxima []float64
		prev, cur := tc.elongate(jd0), tc.elongate(jd0+0.5)
		for jd := jd0 + 1.0; jd <= jd1; jd += 0.5 {
			next := tc.elongate(jd)
			if cur >= prev && cur >= next && (cur > prev || cur > next) {
				left, right := jd-1.0, jd
				for i := 0; i < 60; i++ {
					third := (right - left) / 3
					if tc.elongate(left+third) <= tc.elongate(right-third) {
						left += third
						continue
					}
					right -= third
				}
				maxima = append(maxima, (left+right)/2)
			}
			prev, cur = cur, next
		}
		if len(maxima) < 5 {
			t.Fatalf("%s: too few elongation maxima (%d)", tc.name, len(maxima))
		}
		const matchTolerance = 30.0 / 1440.0
		for i, maximum := range maxima {
			// 查询落在极大前一天：Next 必须命中该极大（不得跳过）
			next := TD2UT(tc.next(maximum-1), true)
			if math.IsNaN(next) {
				t.Fatalf("%s: Next at %s returned NaN", tc.name, tmpPhaseDate(maximum-1))
			}
			if dev := math.Abs(next - maximum); dev > matchTolerance {
				t.Fatalf("%s: Next at %s = %s, expected the maximum %s (%.2f min off)",
					tc.name, tmpPhaseDate(maximum-1), tmpPhaseDate(next), tmpPhaseDate(maximum), dev*1440)
			}
			if next < maximum-1-phaseInvariantToleranceDay {
				t.Fatalf("%s: Next returned an event before the query", tc.name)
			}
			// 查询落在极大后一天：Last 必须命中该极大，Next 必须命中下一个极大
			last := TD2UT(tc.last(maximum+1), true)
			if dev := math.Abs(last - maximum); dev > matchTolerance {
				t.Fatalf("%s: Last at %s = %s, expected the previous maximum %s (%.2f min off)",
					tc.name, tmpPhaseDate(maximum+1), tmpPhaseDate(last), tmpPhaseDate(maximum), dev*1440)
			}
			if last > maximum+1+phaseInvariantToleranceDay {
				t.Fatalf("%s: Last returned an event after the query", tc.name)
			}
			if i+1 < len(maxima) {
				following := TD2UT(tc.next(maximum+1), true)
				if dev := math.Abs(following - maxima[i+1]); dev > matchTolerance {
					t.Fatalf("%s: Next at %s = %s, expected the following maximum %s (%.2f min off)",
						tc.name, tmpPhaseDate(maximum+1), tmpPhaseDate(following), tmpPhaseDate(maxima[i+1]), dev*1440)
				}
			}
		}
	}
}

func tmpPhaseDate(jd float64) string {
	return JDE2DateByZone(TD2UT(jd, false), time.UTC, false).Format("2006-01-02 15:04")
}
