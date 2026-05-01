package basic

import (
	"math"
	"testing"
)

func TestStarDataCacheMatchesRawRows(t *testing.T) {
	if err := LoadStarData(); err != nil {
		t.Fatal(err)
	}
	if len(parsedStarData) != len(stardat) {
		t.Fatalf("parsed cache length mismatch: got %d want %d", len(parsedStarData), len(stardat))
	}
	if len(cachedStarData) != len(stardat) {
		t.Fatalf("full cache length mismatch: got %d want %d", len(cachedStarData), len(stardat))
	}
	for i, row := range stardat {
		parsed, err := parseStarData(row)
		if err != nil {
			t.Fatalf("parse row %d failed: %v", i+1, err)
		}
		if parsed != parsedStarData[i] {
			t.Fatalf("parsed cache mismatch at HR %d", i+1)
		}
		full := fullStarData(parsed)
		if full != cachedStarData[i] {
			t.Fatalf("full cache mismatch at HR %d", i+1)
		}
	}
}

func TestStarDataRegressionSamples(t *testing.T) {
	if err := LoadStarData(); err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		hr               int
		hip              uint32
		name             string
		chineseName      string
		chineseAlias     string
		chineseBayerName string
		commonName       string
		cst              string
		cstChinese       string
		ra               float64
		dec              float64
		mag              float64
		pmRA             float64
		pmDec            float64
	}{
		{15, 677, "21Alp And", "壁宿二", "", "仙女座α", "Alpheratz", "Andromeda", "仙女座", 2.097083333333, 29.090555555556, 2.06, 0.136, -0.163},
		{424, 11767, "1Alp UMi", "勾陈一", "北极星", "小熊座α", "Polaris", "UrsaMinor", "小熊座", 37.952916666667, 89.264166666667, 2.02, 0.038, -0.015},
		{2491, 32349, "9Alp CMa", "天狼", "", "大犬座α", "Sirius", "CanisMajor", "大犬座", 101.287083333333, -16.716111111111, -1.46, -0.553, -1.205},
		{7001, 91262, "3Alp Lyr", "织女一", "织女", "天琴座α", "Vega", "Lyra", "天琴座", 279.234583333333, 38.783611111111, 0.03, 0.202, 0.286},
		{9100, 330, "9    Cas", "", "", "", "", "", "", 1.056666666667, 62.287777777778, 5.88, -0.004, 0.006},
	}
	for _, tc := range tests {
		got, err := StarDataByHR(tc.hr)
		if err != nil {
			t.Fatalf("StarDataByHR(%d) failed: %v", tc.hr, err)
		}
		if got.HR != uint16(tc.hr) || got.HIP != tc.hip || got.Name != tc.name ||
			got.ChineseName != tc.chineseName || got.ChineseAlias != tc.chineseAlias ||
			got.ChineseBayerName != tc.chineseBayerName || got.CommonName != tc.commonName ||
			got.Cst != tc.cst || got.CstChinese != tc.cstChinese {
			t.Fatalf("unexpected metadata for HR %d: %+v", tc.hr, got)
		}
		assertNear(t, "Ra", tc.hr, got.Ra, tc.ra, 1e-12)
		assertNear(t, "Dec", tc.hr, got.Dec, tc.dec, 1e-12)
		assertNear(t, "Mag", tc.hr, got.Mag, tc.mag, 1e-12)
		assertNear(t, "PmRA", tc.hr, got.PmRA, tc.pmRA, 1e-12)
		assertNear(t, "PmDec", tc.hr, got.PmDec, tc.pmDec, 1e-12)
	}
}

func TestStarDataByChineseAlias(t *testing.T) {
	if err := LoadStarData(); err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name string
		hr   uint16
		hip  uint32
	}{
		{"天狼", 2491, 32349},
		{"天狼星", 2491, 32349},
		{"勾陈一", 424, 11767},
		{"北极", 424, 11767},
		{"北极星", 424, 11767},
		{"织女", 7001, 91262},
	}
	for _, tc := range tests {
		got, err := StarDataByChinese(tc.name)
		if err != nil {
			t.Fatalf("StarDataByChinese(%q) failed: %v", tc.name, err)
		}
		if got.HR != tc.hr || got.HIP != tc.hip {
			t.Fatalf("StarDataByChinese(%q) got HR=%d HIP=%d", tc.name, got.HR, got.HIP)
		}
	}
}

func TestStarDataByHRInvalid(t *testing.T) {
	if err := LoadStarData(); err != nil {
		t.Fatal(err)
	}
	for _, hr := range []int{0, -1, len(stardat) + 1} {
		if _, err := StarDataByHR(hr); err == nil {
			t.Fatalf("StarDataByHR(%d) expected error", hr)
		}
	}
}

func BenchmarkInitStarData(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		stardat = nil
		hr2detail = nil
		hr2hip = nil
		chnidx = nil
		parsedStarData = nil
		cachedStarData = nil
		if err := initStarData(); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkStarDataByHR(b *testing.B) {
	if err := LoadStarData(); err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := StarDataByHR(2491); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkStarDataByChinese(b *testing.B) {
	if err := LoadStarData(); err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := StarDataByChinese("天狼星"); err != nil {
			b.Fatal(err)
		}
	}
}

func assertNear(t *testing.T, field string, hr int, got, want, tolerance float64) {
	t.Helper()
	if math.Abs(got-want) > tolerance {
		t.Fatalf("HR %d %s mismatch: got %.15f want %.15f", hr, field, got, want)
	}
}
