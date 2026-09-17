package eclipse

import (
	"testing"
	"time"
)

// 请求的电平必须原样返回：JDE 往返只有微秒级误差，但整分电平会落到 59.999 秒。
func TestSolarEclipsePartialFootprintsReturnsRequestedGreatestTimes(t *testing.T) {
	date := time.Date(2009, time.July, 22, 0, 0, 0, 0, time.UTC)
	levels := []time.Time{
		time.Date(2009, time.July, 22, 1, 0, 0, 0, time.UTC),
		time.Date(2009, time.July, 22, 2, 0, 0, 0, time.UTC),
		time.Date(2009, time.July, 22, 3, 0, 0, 0, time.UTC),
	}
	info, ok := SolarEclipsePartialFootprintsNASABulletinSplitK(date, SolarEclipsePartialFootprintOptions{
		Step: 10 * time.Minute, BoundaryPoints: 24, DisableRiseSet: true,
		GreatestTimeValues: levels,
	})
	if !ok {
		t.Fatal("missing footprints")
	}
	if len(info.GreatestTimeContours) == 0 {
		t.Fatal("no greatest-time contours")
	}
	for _, contour := range info.GreatestTimeContours {
		matched := false
		for _, level := range levels {
			if contour.Time.Equal(level) {
				matched = true
				break
			}
		}
		if !matched {
			t.Fatalf("contour time %s is not one of the requested levels",
				contour.Time.Format("15:04:05.000"))
		}
		for _, segment := range contour.Segments {
			if len(segment) < 2 {
				t.Fatalf("contour %s has a degenerate branch", contour.Time.Format("15:04"))
			}
		}
	}
}

// 按步长请求等时线：时刻按输入时区对齐到整刻度、落在偏食窗口内，并回填 JDE。
func TestSolarEclipsePartialFootprintsGreatestTimeStep(t *testing.T) {
	location := time.FixedZone("UTC+8", 8*3600)
	date := time.Date(2009, time.July, 22, 0, 0, 0, 0, location)
	info, ok := SolarEclipsePartialFootprintsNASABulletinSplitK(date, SolarEclipsePartialFootprintOptions{
		Step: 10 * time.Minute, BoundaryPoints: 24, DisableRiseSet: true,
		GreatestTimeStep: 30 * time.Minute,
	})
	if !ok {
		t.Fatal("missing footprints")
	}
	if len(info.GreatestTimeContours) < 2 {
		t.Fatalf("step request produced %d contours", len(info.GreatestTimeContours))
	}
	for _, contour := range info.GreatestTimeContours {
		if contour.JDE == 0 {
			t.Fatal("contour JDE is not populated")
		}
		if contour.Time.Nanosecond() != 0 || contour.Time.Second() != 0 || contour.Time.Minute()%30 != 0 {
			t.Fatalf("contour time %s is not aligned to the requested step",
				contour.Time.Format("15:04:05.000000000"))
		}
		if contour.Time.Before(info.Eclipse.PartialBeginOnEarth) || contour.Time.After(info.Eclipse.PartialEndOnEarth) {
			t.Fatalf("contour %s is outside the partial-eclipse window", contour.Time.Format("15:04"))
		}
		for _, segment := range contour.Segments {
			if len(segment) < 2 {
				t.Fatalf("contour %s has a degenerate branch", contour.Time.Format("15:04"))
			}
		}
	}
}
