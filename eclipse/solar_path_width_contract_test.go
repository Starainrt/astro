package eclipse

import (
	"math"
	"testing"
	"time"
)

// basic 与 eclipse 两层必须给出同一个带宽口径标志：单侧极限事件带宽无定义且值恒为 0。

func TestSolarEclipsePathCarriesPathWidthDefinition(t *testing.T) {
	testCases := []struct {
		name    string
		date    time.Time
		defined bool
		widthKM float64
	}{
		{name: "-1404-01-07 single-sided limit", date: time.Date(-1404, 1, 7, 0, 0, 0, 0, time.UTC)},
		{name: "2003-05-31 single-sided limit with limit lines", date: time.Date(2003, 5, 31, 0, 0, 0, 0, time.UTC)},
		{name: "2024-04-08 total", date: time.Date(2024, 4, 8, 0, 0, 0, 0, time.UTC), defined: true, widthKM: 198.6161},
		{name: "2010-01-15 annular", date: time.Date(2010, 1, 15, 0, 0, 0, 0, time.UTC), defined: true, widthKM: 335.0298},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			info := ClosestSolarEclipse(tc.date)
			if info.PathWidthDefined != tc.defined {
				t.Fatalf("info.PathWidthDefined=%v want %v", info.PathWidthDefined, tc.defined)
			}
			if !tc.defined && info.PathWidthKM != 0 {
				t.Fatalf("undefined info width %.4f", info.PathWidthKM)
			}
			path, ok := SolarEclipseCentralPath(tc.date, SolarEclipsePathOptions{})
			if !ok {
				t.Fatalf("no central path")
			}
			if path.PathWidthDefined != tc.defined || path.Eclipse.PathWidthDefined != tc.defined {
				t.Fatalf("defined=%v eclipse=%v want %v", path.PathWidthDefined, path.Eclipse.PathWidthDefined, tc.defined)
			}
			if !tc.defined {
				if path.Eclipse.PathWidthKM != 0 || path.Greatest.WidthKM != 0 {
					t.Fatalf("undefined width leaks: eclipse=%.4f greatest=%.4f",
						path.Eclipse.PathWidthKM, path.Greatest.WidthKM)
				}
				return
			}
			if math.Abs(path.Eclipse.PathWidthKM-tc.widthKM) > 0.01 {
				t.Fatalf("path width %.4f want %.4f", path.Eclipse.PathWidthKM, tc.widthKM)
			}
			if path.Eclipse.PathWidthKM > 0 && (len(path.NorthernLimit) == 0 || len(path.SouthernLimit) == 0) {
				t.Fatalf("width %.4f km with limits %d/%d", path.Eclipse.PathWidthKM,
					len(path.NorthernLimit), len(path.SouthernLimit))
			}
		})
	}
}
