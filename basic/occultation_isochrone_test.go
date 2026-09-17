package basic

import (
	"math"
	"testing"
	"time"
)

func occultationIsochroneTestStar() StarCoordinate {
	return StarCoordinate{
		ID:    "Regulus",
		RA:    152.09292,
		Dec:   11.96719,
		Epoch: time.Date(2000, time.January, 1, 12, 0, 0, 0, time.UTC),
		Frame: CoordinateFrameJ2000,
	}
}

// 等时线必须与站心掩甚定义自洽：支路上任意点由站心事件搜索独立求出的掩甚时刻等于该支路电平。
func TestOccultationGreatestTimeContoursMatchLocalGreatest(t *testing.T) {
	star := occultationIsochroneTestStar()
	start := time.Date(2025, time.August, 22, 0, 0, 0, 0, time.UTC)
	end := time.Date(2025, time.August, 24, 0, 0, 0, 0, time.UTC)
	options := OccultationPathOptions{
		Step: 2 * time.Minute, DisableFootprints: true, GreatestTimeStep: 15 * time.Minute,
	}
	paths, err := FindStarOccultationPaths(start, end, star, options)
	if err != nil || len(paths) == 0 {
		t.Fatalf("no path with isochrones: %v", err)
	}
	contours := paths[0].GreatestTimeContours
	if len(contours) < 4 {
		t.Fatalf("expected a full isochrone fan, got %d contours", len(contours))
	}
	for _, contour := range contours {
		if contour.Time.Second() != 0 || contour.Time.Nanosecond() != 0 || contour.Time.Minute()%15 != 0 {
			t.Fatalf("contour time %s is not aligned to the requested step",
				contour.Time.Format("15:04:05.000"))
		}
	}

	verified := 0
	worst := 0.0
	for _, contour := range contours {
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
		level := occultationTTToLocation(contour.JDE, time.UTC)
		for _, segment := range contour.Segments {
			for index := 0; index < len(segment) && verified < 80; index += 1 + len(segment)/8 {
				point := segment[index]
				infos, searchErr := FindStarOccultations(
					level.Add(-2*time.Hour), level.Add(2*time.Hour), star,
					point.Longitude, point.Latitude, 0, OccultationSearchOptions{},
				)
				if searchErr != nil || len(infos) == 0 {
					continue
				}
				delta := math.Abs(infos[0].Greatest.Sub(point.Time).Seconds())
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
	t.Logf("%d contours, %d points verified, worst %.2f s", len(contours), verified, worst)
	if worst > 1 {
		t.Fatalf("isochrone disagrees with local greatest occultation by %.2f s", worst)
	}
}

func TestOccultationGreatestTimeContoursDisabledByDefault(t *testing.T) {
	star := occultationIsochroneTestStar()
	start := time.Date(2025, time.August, 22, 0, 0, 0, 0, time.UTC)
	end := time.Date(2025, time.August, 24, 0, 0, 0, 0, time.UTC)
	paths, err := FindStarOccultationPaths(start, end, star, OccultationPathOptions{
		Step: 2 * time.Minute, DisableFootprints: true,
	})
	if err != nil || len(paths) == 0 {
		t.Fatalf("no occultation path: %v", err)
	}
	if len(paths[0].GreatestTimeContours) != 0 {
		t.Fatalf("contours computed without a request: %d", len(paths[0].GreatestTimeContours))
	}
}

// 行星是有限盘面，等时线按外接触锥门控，必须与站心行星掩甚一致。
func TestPlanetOccultationGreatestTimeContoursMatchLocalGreatest(t *testing.T) {
	start := time.Date(2025, time.February, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2025, time.February, 2, 0, 0, 0, 0, time.UTC)
	options := OccultationPathOptions{
		Step: 2 * time.Minute, DisableFootprints: true, GreatestTimeStep: 30 * time.Minute,
	}
	paths, err := FindPlanetOccultationPaths(start, end, OccultationSaturn, options)
	if err != nil || len(paths) == 0 {
		t.Fatalf("no planetary path with isochrones: %v", err)
	}
	contours := paths[0].GreatestTimeContours
	if len(contours) == 0 {
		t.Fatal("no greatest-time contours")
	}
	verified := 0
	worst := 0.0
	for _, contour := range contours {
		level := occultationTTToLocation(contour.JDE, time.UTC)
		for _, segment := range contour.Segments {
			if len(segment) < 2 {
				t.Fatalf("contour %.6f has a degenerate branch", contour.JDE)
			}
			for index := 0; index < len(segment) && verified < 60; index += 1 + len(segment)/8 {
				point := segment[index]
				infos, searchErr := FindPlanetOccultations(
					level.Add(-3*time.Hour), level.Add(3*time.Hour), OccultationSaturn,
					point.Longitude, point.Latitude, 0, OccultationSearchOptions{},
				)
				if searchErr != nil || len(infos) == 0 {
					continue
				}
				delta := math.Abs(infos[0].Greatest.Sub(point.Time).Seconds())
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
	t.Logf("%d contours, %d points verified, worst %.2f s", len(contours), verified, worst)
	if worst > 1 {
		t.Fatalf("isochrone disagrees with local greatest occultation by %.2f s", worst)
	}
}

// 同一时刻取值的等时线必须是单条连通曲线：曾因"覆盖判据用点到顶点距离"而被重复延拓
// （行星掩星每个取值 3 条重叠支路）。
func TestOccultationGreatestTimeContoursHaveSingleBranchPerLevel(t *testing.T) {
	star := occultationIsochroneTestStar()
	starPaths, err := FindStarOccultationPaths(
		time.Date(2025, time.August, 22, 0, 0, 0, 0, time.UTC),
		time.Date(2025, time.August, 24, 0, 0, 0, 0, time.UTC), star,
		OccultationPathOptions{Step: 2 * time.Minute, DisableFootprints: true, GreatestTimeStep: 15 * time.Minute},
	)
	if err != nil || len(starPaths) == 0 {
		t.Fatalf("star: %v", err)
	}
	planetPaths, err := FindPlanetOccultationPaths(
		time.Date(2025, time.February, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2025, time.February, 2, 0, 0, 0, 0, time.UTC), OccultationSaturn,
		OccultationPathOptions{Step: 2 * time.Minute, DisableFootprints: true, GreatestTimeStep: 30 * time.Minute},
	)
	if err != nil || len(planetPaths) == 0 {
		t.Fatalf("planet: %v", err)
	}
	for _, tc := range []struct {
		name     string
		contours []OccultationGreatestTimeContour
	}{
		{"star", starPaths[0].GreatestTimeContours},
		{"planet", planetPaths[0].GreatestTimeContours},
	} {
		if len(tc.contours) == 0 {
			t.Fatalf("%s: no contours", tc.name)
		}
		for _, contour := range tc.contours {
			if len(contour.Segments) != 1 {
				t.Fatalf("%s contour %s has %d branches, want a single connected curve",
					tc.name, contour.Time.Format("15:04"), len(contour.Segments))
			}
		}
	}
}

// 显式时刻取值：按 TT 儒略日给出，回显时刻取整到毫秒，窗口外取值被丢弃，且同样受上限保护。
func TestOccultationGreatestTimeContoursAcceptExplicitValues(t *testing.T) {
	star := occultationIsochroneTestStar()
	start := time.Date(2025, time.August, 22, 0, 0, 0, 0, time.UTC)
	end := time.Date(2025, time.August, 24, 0, 0, 0, 0, time.UTC)
	base := OccultationPathOptions{Step: 2 * time.Minute, DisableFootprints: true, GreatestTimeStep: 15 * time.Minute}
	stepPaths, err := FindStarOccultationPaths(start, end, star, base)
	if err != nil || len(stepPaths) == 0 || len(stepPaths[0].GreatestTimeContours) == 0 {
		t.Fatalf("step request failed: %v", err)
	}
	stepContours := stepPaths[0].GreatestTimeContours
	// 用步长路径得到的取值当显式取值，外加 8 个窗口外取值，检验回显与过滤。
	values := make([]float64, 0, len(stepContours)+8)
	for _, contour := range stepContours {
		values = append(values, contour.JDE)
	}
	for index := 0; index < 8; index++ {
		values = append(values, occultationTimeToTT(start.Add(-time.Duration(index+1)*time.Hour)))
	}
	explicitPaths, err := FindStarOccultationPaths(start, end, star, OccultationPathOptions{
		Step: 2 * time.Minute, DisableFootprints: true, GreatestTimeValues: values,
	})
	if err != nil || len(explicitPaths) == 0 {
		t.Fatalf("explicit request failed: %v", err)
	}
	contours := explicitPaths[0].GreatestTimeContours
	if len(contours) != len(stepContours) {
		t.Fatalf("explicit values produced %d contours, step request produced %d", len(contours), len(stepContours))
	}
	for index, contour := range contours {
		if contour.JDE != stepContours[index].JDE {
			t.Fatalf("contour %d JDE %.9f does not match the requested value %.9f", index, contour.JDE, stepContours[index].JDE)
		}
		if contour.Time.Nanosecond() != 0 {
			t.Fatalf("contour time %s is not rounded to the millisecond", contour.Time.Format("15:04:05.000000000"))
		}
	}

	// 上限：72 个互不相同、都落在窗口内的取值，返回条数不得超过共享上限。
	dense := make([]float64, 0, greatestTimeContourMaxLevels+8)
	for index := 0; index < greatestTimeContourMaxLevels+8; index++ {
		dense = append(dense, stepContours[0].JDE+float64(index)*1e-6)
	}
	cappedPaths, err := FindStarOccultationPaths(start, end, star, OccultationPathOptions{
		Step: 2 * time.Minute, DisableFootprints: true, GreatestTimeValues: dense,
	})
	if err != nil || len(cappedPaths) == 0 {
		t.Fatalf("capped request failed: %v", err)
	}
	if count := len(cappedPaths[0].GreatestTimeContours); count > greatestTimeContourMaxLevels {
		t.Fatalf("level cap not applied: %d contours", count)
	}
}

func BenchmarkOccultationGreatestTimeContours(b *testing.B) {
	star := occultationIsochroneTestStar()
	start := time.Date(2025, time.August, 22, 0, 0, 0, 0, time.UTC)
	end := time.Date(2025, time.August, 24, 0, 0, 0, 0, time.UTC)
	options := OccultationPathOptions{
		Step: 2 * time.Minute, DisableFootprints: true, GreatestTimeStep: 15 * time.Minute,
	}
	for index := 0; index < b.N; index++ {
		FindStarOccultationPaths(start, end, star, options)
	}
}
