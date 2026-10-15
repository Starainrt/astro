package basic

import (
	"math"
	"testing"

	"github.com/starainrt/astro/internal/ephem"
)

// 带宽契约：解析式 2r/|sin(太阳高度)| 只在南北两限都存在时有定义，单侧极限事件置 0 并标记未定义。

func TestSolarEclipsePathWidthRequiresBothLimits(t *testing.T) {
	testCases := []struct {
		name    string
		date    [3]int
		defined bool
		widthKM float64
		semiKM  float64
	}{
		{name: "-1404-01-07 single-sided limit without limits", date: [3]int{-1404, 1, 7}},
		{name: "2003-05-31 single-sided limit with limit lines", date: [3]int{2003, 5, 31}},
		{name: "1874-10-10 single-sided limit with limit lines", date: [3]int{1874, 10, 10}},
		{name: "2024-04-08 total", date: [3]int{2024, 4, 8}, defined: true, widthKM: 198.6161, semiKM: 198.6214},
		{name: "2010-01-15 annular", date: [3]int{2010, 1, 15}, defined: true, widthKM: 335.0298, semiKM: 335.0272},
		{name: "4862-09-28 annular", date: [3]int{4862, 9, 28}, defined: true, widthKM: 259.4316, semiKM: 193.8401},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			seed := JDCalc(tc.date[0], tc.date[1], float64(tc.date[2]))
			path := SolarEclipseCentralPath(seed, SolarEclipsePathOptions{})
			info := SolarEclipse(seed)
			if path.Eclipse.PathWidthDefined != tc.defined {
				t.Fatalf("path.Eclipse.PathWidthDefined=%v want %v", path.Eclipse.PathWidthDefined, tc.defined)
			}
			if path.PathWidthDefined != tc.defined {
				t.Fatalf("path.PathWidthDefined=%v want %v", path.PathWidthDefined, tc.defined)
			}
			if info.PathWidthDefined != tc.defined {
				t.Fatalf("info.PathWidthDefined=%v want %v", info.PathWidthDefined, tc.defined)
			}
			if info.PathWidthDefined != (info.Centrality == SolarEclipseCentralTwoLimits) {
				t.Fatalf("info flag %v disagrees with centrality %s", info.PathWidthDefined, info.Centrality)
			}
			if !tc.defined {
				if path.Eclipse.PathWidthKM != 0 || path.Greatest.WidthKM != 0 || info.PathWidthKM != 0 {
					t.Fatalf("undefined width leaks: eclipse=%.4f greatest=%.4f info=%.4f",
						path.Eclipse.PathWidthKM, path.Greatest.WidthKM, info.PathWidthKM)
				}
			} else {
				want := tc.widthKM
				if ephem.Active() == ephem.Semianalytic {
					want = tc.semiKM
				}
				if math.Abs(path.Eclipse.PathWidthKM-want) > 0.01 {
					t.Fatalf("path width %.4f want %.4f", path.Eclipse.PathWidthKM, want)
				}
			}
			if path.Eclipse.PathWidthKM > 0 && (len(path.NorthernLimit) == 0 || len(path.SouthernLimit) == 0) {
				t.Fatalf("width %.4f km with limits %d/%d", path.Eclipse.PathWidthKM,
					len(path.NorthernLimit), len(path.SouthernLimit))
			}
			if (len(path.NorthernLimit) == 0) != (len(path.SouthernLimit) == 0) {
				t.Fatalf("limits are not paired: %d/%d", len(path.NorthernLimit), len(path.SouthernLimit))
			}
		})
	}
}

func TestSolarEclipsePathWidthContractAcrossScan(t *testing.T) {
	start := JDCalc(1998, 1, 1)
	end := JDCalc(2018, 1, 1)
	seenOneLimit, seenTwoLimits := 0, 0
	lastGreatest := 0.0
	for jd := start; jd < end; jd += 15 {
		path := SolarEclipseCentralPath(jd, SolarEclipsePathOptions{SkipCentralBand: true})
		if path.Eclipse.Type == SolarEclipseNone {
			continue
		}
		if lastGreatest != 0 && path.Eclipse.GreatestEclipse-lastGreatest < 1 {
			continue
		}
		lastGreatest = path.Eclipse.GreatestEclipse
		switch path.Eclipse.Centrality {
		case SolarEclipseCentralOneLimit:
			seenOneLimit++
		case SolarEclipseCentralTwoLimits:
			seenTwoLimits++
		}
		bothLimits := len(path.NorthernLimit) > 0 && len(path.SouthernLimit) > 0
		if path.PathWidthDefined && (!bothLimits || path.Eclipse.Centrality != SolarEclipseCentralTwoLimits) {
			t.Fatalf("defined at %s with limits %d/%d", path.Eclipse.Centrality,
				len(path.NorthernLimit), len(path.SouthernLimit))
		}
		if !path.PathWidthDefined && path.Eclipse.PathWidthKM != 0 {
			t.Fatalf("undefined width %.4f km at %.6f", path.Eclipse.PathWidthKM, path.Eclipse.GreatestEclipse)
		}
		if path.Eclipse.PathWidthKM > 0 && (len(path.NorthernLimit) == 0 || len(path.SouthernLimit) == 0) {
			t.Fatalf("width %.4f km with limits %d/%d at %.6f", path.Eclipse.PathWidthKM,
				len(path.NorthernLimit), len(path.SouthernLimit), path.Eclipse.GreatestEclipse)
		}
	}
	if seenOneLimit == 0 || seenTwoLimits == 0 {
		t.Fatalf("scan covered %d one-limit and %d two-limit events, want both kinds", seenOneLimit, seenTwoLimits)
	}
}
