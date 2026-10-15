package basic

import (
	"math"
	"testing"
	"time"
)

func solarEclipseIsochroneTestSeed(t *testing.T, year int, month time.Month, day int) float64 {
	t.Helper()
	return UTC2TT(Date2JD(time.Date(year, month, day, 0, 0, 0, 0, time.UTC)))
}

func solarEclipseIsochroneTestSeedB(b *testing.B, year int, month time.Month, day int) float64 {
	b.Helper()
	return UTC2TT(Date2JD(time.Date(year, month, day, 0, 0, 0, 0, time.UTC)))
}

// 等时线必须与站心食甚定义自洽：支路上任意点由站心算法独立求出的食甚时刻等于该支路电平。
func TestSolarEclipseGreatestTimeContoursMatchLocalCircumstances(t *testing.T) {
	for _, tc := range []struct {
		name  string
		seed  float64
		first time.Time
		count int
	}{
		{name: "2009Total", seed: solarEclipseIsochroneTestSeed(t, 2009, time.July, 22),
			first: time.Date(2009, time.July, 22, 1, 30, 0, 0, time.UTC), count: 6},
		{name: "2010Annular", seed: solarEclipseIsochroneTestSeed(t, 2010, time.January, 15),
			first: time.Date(2010, time.January, 15, 5, 0, 0, 0, time.UTC), count: 6},
	} {
		t.Run(tc.name, func(t *testing.T) {
			levels := make([]float64, 0, tc.count)
			for index := 0; index < tc.count; index++ {
				levels = append(levels, UTC2TT(Date2JD(tc.first.Add(time.Duration(index)*30*time.Minute))))
			}
			result := SolarEclipsePartialFootprintsNASABulletinSplitK(tc.seed, SolarEclipsePartialFootprintOptions{
				StepDays:             10.0 / 1440.0,
				BoundaryPoints:       24,
				DisableRiseSetCurves: true,
				GreatestTimeValues:   levels,
			})
			if len(result.GreatestTimeContours) == 0 {
				t.Fatal("no greatest-time contours")
			}
			verified := 0
			worst := 0.0
			for _, contour := range result.GreatestTimeContours {
				branchPoints := 0
				for _, segment := range contour.Segments {
					if len(segment) < 2 {
						t.Fatalf("contour %.6f has a degenerate branch", contour.JDE)
					}
					branchPoints += len(segment)
				}
				// 弧长步长上限决定了单条等时线的点数上界；点数爆表意味着延拓没有真正前进。
				if branchPoints > 4000 {
					t.Fatalf("contour %.6f has %d points, continuation is not advancing", contour.JDE, branchPoints)
				}
				for _, segment := range contour.Segments {
					// 独立回验很贵（每次是完整站心求解），按点数预算抽样而不是全量遍历；
					// 每个时刻取值都要抽到，不能只看第一条等时线。
					for index := 0; index < len(segment) && verified < 240; index += 1 + len(segment)/12 {
						point := segment[index]
						local := LocalSolarEclipseNASABulletinSplitK(
							contour.JDE, point.Longitude, point.Latitude, 0,
						)
						if local.GreatestEclipse == 0 {
							continue
						}
						delta := math.Abs(local.GreatestEclipse-contour.JDE) * 86400
						if delta > worst {
							worst = delta
						}
						verified++
					}
				}
			}
			if verified < 10 {
				t.Fatalf("verified only %d points", verified)
			}
			t.Logf("%s: %d contours, %d points verified, worst %.2f s",
				tc.name, len(result.GreatestTimeContours), verified, worst)
			if worst > 3 {
				t.Fatalf("isochrone disagrees with local greatest eclipse by %.2f s", worst)
			}
		})
	}
}

func TestSolarEclipseGreatestTimeContoursDisabledByDefault(t *testing.T) {
	seed := solarEclipseIsochroneTestSeed(t, 2009, time.July, 22)
	result := SolarEclipsePartialFootprintsNASABulletinSplitK(seed, SolarEclipsePartialFootprintOptions{
		StepDays: 10.0 / 1440.0, DisableRiseSetCurves: true,
	})
	if len(result.GreatestTimeContours) != 0 {
		t.Fatalf("contours computed without a request: %d", len(result.GreatestTimeContours))
	}
}

func TestSolarEclipseGreatestTimeContoursRejectOutOfWindowLevels(t *testing.T) {
	seed := solarEclipseIsochroneTestSeed(t, 2009, time.July, 22)
	result := SolarEclipsePartialFootprintsNASABulletinSplitK(seed, SolarEclipsePartialFootprintOptions{
		StepDays:             10.0 / 1440.0,
		DisableRiseSetCurves: true,
		GreatestTimeValues:   []float64{seed - 5, seed + 5, math.NaN()},
	})
	if len(result.GreatestTimeContours) != 0 {
		t.Fatalf("accepted out-of-window levels: %d", len(result.GreatestTimeContours))
	}
}

// 同一时刻取值的等时线必须是一条连通曲线，不得由多个种子重复描出支路。
func TestSolarEclipseGreatestTimeContoursHaveSingleBranchPerLevel(t *testing.T) {
	for _, tc := range []struct {
		name string
		seed float64
		hour int
	}{
		{"2009Total", solarEclipseIsochroneTestSeed(t, 2009, time.July, 22), 2},
		{"2024Total", solarEclipseIsochroneTestSeed(t, 2024, time.April, 8), 18},
		{"2014NonCentral", solarEclipseIsochroneTestSeed(t, 2014, time.April, 29), 6},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// 用事件当天同一个世界时小时作为时刻取值。
			seedTime := JD2DateByZone(TT2UTC(tc.seed), time.UTC, false)
			eventLevel := UTC2TT(Date2JD(time.Date(seedTime.Year(), seedTime.Month(), seedTime.Day(), tc.hour, 0, 0, 0, time.UTC)))
			result := SolarEclipsePartialFootprintsNASABulletinSplitK(tc.seed, SolarEclipsePartialFootprintOptions{
				StepDays: 10.0 / 1440.0, BoundaryPoints: 24, DisableRiseSetCurves: true,
				GreatestTimeValues: []float64{eventLevel},
			})
			if len(result.GreatestTimeContours) != 1 {
				t.Fatalf("expected one contour, got %d", len(result.GreatestTimeContours))
			}
			contour := result.GreatestTimeContours[0]
			if len(contour.Segments) != 1 {
				t.Fatalf("contour %.6f has %d branches, want a single connected curve", contour.JDE, len(contour.Segments))
			}
			if points := len(contour.Segments[0]); points < 20 {
				t.Fatalf("contour has only %d points", points)
			}
		})
	}
}

// 按步长请求多个时刻取值时，每条等时线同样只能有一条支路。
func TestSolarEclipseGreatestTimeStepHasSingleBranchPerLevel(t *testing.T) {
	for _, tc := range []struct {
		name string
		seed float64
	}{
		{"2009Total", solarEclipseIsochroneTestSeed(t, 2009, time.July, 22)},
		{"2024Total", solarEclipseIsochroneTestSeed(t, 2024, time.April, 8)},
		{"2014NonCentral", solarEclipseIsochroneTestSeed(t, 2014, time.April, 29)},
		{"2010Annular", solarEclipseIsochroneTestSeed(t, 2010, time.January, 15)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			result := SolarEclipsePartialFootprintsNASABulletinSplitK(tc.seed, SolarEclipsePartialFootprintOptions{
				StepDays: 10.0 / 1440.0, BoundaryPoints: 24, DisableRiseSetCurves: true,
				GreatestTimeStep: 30 * time.Minute,
			})
			if len(result.GreatestTimeContours) < 2 {
				t.Fatalf("step request produced %d contours", len(result.GreatestTimeContours))
			}
			branches := 0
			for _, contour := range result.GreatestTimeContours {
				branches += len(contour.Segments)
				if len(contour.Segments) != 1 {
					t.Fatalf("contour %.6f has %d branches, want a single connected curve", contour.JDE, len(contour.Segments))
				}
			}
			if branches != len(result.GreatestTimeContours) {
				t.Fatalf("%d branches for %d contours", branches, len(result.GreatestTimeContours))
			}
		})
	}
}

// 延拓必须覆盖整个零集：0.5°×2° 细扫找到的每个根都要落在已绘折线一个覆盖容差内。
func TestSolarEclipseGreatestTimeContoursCoverZeroSet(t *testing.T) {
	for _, tc := range []struct {
		name string
		seed float64
		hour int
	}{
		{"2009Total", solarEclipseIsochroneTestSeed(t, 2009, time.July, 22), 2},
		{"2014NonCentral", solarEclipseIsochroneTestSeed(t, 2014, time.April, 29), 6},
	} {
		t.Run(tc.name, func(t *testing.T) {
			seedTime := JD2DateByZone(TT2UTC(tc.seed), time.UTC, false)
			level := UTC2TT(Date2JD(time.Date(seedTime.Year(), seedTime.Month(), seedTime.Day(), tc.hour, 0, 0, 0, time.UTC)))
			solver := newSolarEclipseSolver(CalcMoonSHByJDE(tc.seed, 0), SolarEclipseModelNASABulletinSplitK)
			segments := solver.greatestTimeContourSegments(level)
			if len(segments) == 0 {
				t.Fatal("no isochrone for the requested level")
			}
			arc := solarEclipseGreatestTimeArc{evaluation: solver.magnitudeEvaluationAt(level)}
			roots, missed := 0, 0
			for latitude := -86.0; latitude <= 86.0; latitude += 2.0 {
				previous := arc.residual(-180, latitude)
				for longitude := -179.5; longitude <= 180; longitude += 0.5 {
					current := arc.residual(longitude, latitude)
					if finite(previous) && finite(current) && previous*current <= 0 {
						left, right, leftValue := longitude-0.5, longitude, previous
						for iteration := 0; iteration < 60; iteration++ {
							middle := (left + right) / 2
							value := arc.residual(middle, latitude)
							if !finite(value) {
								break
							}
							if math.Abs(value) <= greatestTimeContourResidualTolerance || right-left <= 1e-9 {
								left, right = middle, middle
								break
							}
							if leftValue*value <= 0 {
								right = middle
							} else {
								left, leftValue = middle, value
							}
						}
						root := (left + right) / 2
						roots++
						best := math.Inf(1)
						for _, segment := range segments {
							for index := 1; index < len(segment); index++ {
								distance := greatestTimeContourPointSegmentKM(
									root, latitude,
									segment[index-1].Longitude, segment[index-1].Latitude,
									segment[index].Longitude, segment[index].Latitude,
								)
								if distance < best {
									best = distance
								}
							}
						}
						if best > greatestTimeContourCoverToleranceKM {
							missed++
						}
					}
					previous = current
				}
			}
			if roots < 10 {
				t.Fatalf("fine scan found only %d roots", roots)
			}
			if missed > 0 {
				t.Fatalf("%d of %d zero-set roots are farther than %.0f km from the traced isochrone", missed, roots, greatestTimeContourCoverToleranceKM)
			}
		})
	}
}

// 按步长请求时：取值对齐到 UTC 整刻度、落在偏食窗口内、数量受上限保护。
func TestSolarEclipseGreatestTimeStepAlignsAndCapsLevels(t *testing.T) {
	seed := solarEclipseIsochroneTestSeed(t, 2009, time.July, 22)
	result := SolarEclipsePartialFootprintsNASABulletinSplitK(seed, SolarEclipsePartialFootprintOptions{
		StepDays: 10.0 / 1440.0, BoundaryPoints: 24, DisableRiseSetCurves: true,
		GreatestTimeStep: 30 * time.Minute,
	})
	// 对齐网格的取值是固定契约，写成字面量以免与被测的取值生成函数互相自证。
	want := []float64{
		2455034.542432685, 2455034.563266018, 2455034.584099352, 2455034.604932685,
		2455034.625766018, 2455034.646599352, 2455034.667432685,
	}
	if len(result.GreatestTimeContours) != len(want) {
		t.Fatalf("step request produced %d contours, want %d", len(result.GreatestTimeContours), len(want))
	}
	// 独立重算网格：只用导出的 TT/UTC 换算与整刻度截断，不经过被测的取值生成函数。
	begin := JD2DateByZone(TT2UTC(result.Eclipse.PartialBeginOnEarth), time.UTC, false)
	first := begin.Truncate(30 * time.Minute)
	if first.Before(begin) {
		first = first.Add(30 * time.Minute)
	}
	end := JD2DateByZone(TT2UTC(result.Eclipse.PartialEndOnEarth), time.UTC, false)
	ticked := 0
	for current := first; !current.After(end); current = current.Add(30 * time.Minute) {
		ticked++
	}
	// 没有支路的取值会被丢掉，所以返回的是独立重算网格的子集，但必须是子集。
	if ticked < len(want) {
		t.Fatalf("独立重算的对齐网格只有 %d 个刻度，容不下 %d 条等时线", ticked, len(want))
	}
	for index, contour := range result.GreatestTimeContours {
		if math.Abs(contour.JDE-want[index]) > 1e-9 {
			t.Fatalf("contour %d JDE = %.9f, want %.9f", index, contour.JDE, want[index])
		}
		if len(contour.Segments) == 0 {
			t.Fatalf("contour %d has no branches", index)
		}
	}
	// 上限：给一个极细的步长，返回的条数不得超过共享上限。
	tiny := SolarEclipsePartialFootprintsNASABulletinSplitK(seed, SolarEclipsePartialFootprintOptions{
		StepDays: 10.0 / 1440.0, BoundaryPoints: 24, DisableRiseSetCurves: true,
		GreatestTimeStep: time.Minute,
	})
	if len(tiny.GreatestTimeContours) > greatestTimeContourMaxLevels {
		t.Fatalf("level cap not applied: %d contours", len(tiny.GreatestTimeContours))
	}
}

func BenchmarkSolarEclipseGreatestTimeContours(b *testing.B) {
	seed := solarEclipseIsochroneTestSeedB(b, 2009, time.July, 22)
	levels := make([]float64, 0, 6)
	for index := 0; index < 6; index++ {
		levels = append(levels, UTC2TT(Date2JD(time.Date(2009, time.July, 22, 1, 30, 0, 0, time.UTC).Add(time.Duration(index)*30*time.Minute))))
	}
	options := SolarEclipsePartialFootprintOptions{
		StepDays: 10.0 / 1440.0, BoundaryPoints: 24, DisableRiseSetCurves: true,
		GreatestTimeValues: levels,
	}
	for index := 0; index < b.N; index++ {
		SolarEclipsePartialFootprintsNASABulletinSplitK(seed, options)
	}
}
