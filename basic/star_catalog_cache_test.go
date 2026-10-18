package basic

import (
	"errors"
	"math"
	"testing"
)

func TestStarDataCacheMatchesDecodedRows(t *testing.T) {
	if err := LoadStarData(); err != nil {
		t.Fatal(err)
	}
	if len(cachedStarData) != len(parsedStarData) {
		t.Fatalf("full cache length mismatch: got %d want %d", len(cachedStarData), len(parsedStarData))
	}
	for i, parsed := range parsedStarData {
		if full := fullStarData(parsed); full != cachedStarData[i] {
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
		{4799, 61558, "25    Vir", "进贤增九", "", "室女座25", "", "Virgo", "室女座", 189.1975, -5.831944444444, 5.87, -0.028, -0.018},
		{7001, 91262, "3Alp Lyr", "织女一", "织女", "天琴座α", "Vega", "Lyra", "天琴座", 279.234583333333, 38.783611111111, 0.03, 0.202, 0.286},
		{1140, 17489, "16    Tau", "昴宿增九", "", "金牛座16", "Celaeno", "Taurus", "金牛座", 56.200833333333, 24.289444444444, 5.46, 0.011, -0.046},
		{1144, 17527, "18    Tau", "昴宿增八", "", "金牛座18", "", "Taurus", "金牛座", 56.290416666667, 24.839166666667, 5.64, 0.02, -0.047},
		{1151, 17579, "21    Tau", "昴宿三", "", "金牛座21", "Asterope", "Taurus", "金牛座", 56.476666666667, 24.554722222222, 5.76, 0.011, -0.042},
		{1180, 17851, "28    Tau", "昴宿增十二", "", "金牛座28", "Pleione", "Taurus", "金牛座", 57.296666666667, 24.136666666667, 5.09, 0.013, -0.05},
		{8610, 111710, "63Kap Aqr", "虚梁三", "", "宝瓶座κ", "Situla", "Aquarius", "宝瓶座", 339.439166666667, -4.228055555556, 5.03, -0.067, -0.12},
		{390, 6411, "46Xi  And", "天大将军增一", "天大将军增二", "仙女座ξ", "Adhil", "Andromeda", "仙女座", 20.585, 45.528888888889, 4.88, 0.032, 0.009},
		{2585, 33485, "16    Lyn", "座旗二", "", "天猫座16", "", "Lynx", "天猫座", 104.404583333333, 45.094166666667, 4.90, -0.021, -0.004},
		{1412, 20894, "78The2Tau", "毕宿增十三", "", "金牛座θ2", "", "Taurus", "金牛座", 67.165416666667, 15.870833333333, 3.40, 0.103, -0.025},
		{5533, 72487, "38    Boo", "玄戈增二", "", "牧夫座38", "Merga", "Bootes", "牧夫座", 222.327916666667, 46.116111111111, 5.74, -0.002, -0.076},
		{6008, 79043, "7Kap Her", "天市右垣三", "", "武仙座κ", "Marsic", "Hercules", "武仙座", 242.01875, 17.046944444444, 5.00, -0.032, -0.007},
		{7735, 99675, "31    Cyg", "天津增三十八", "", "天鹅座ο1", "", "Cygnus", "天鹅座", 303.407916666667, 46.741388888889, 3.79, 0.004, 0.003},
		{8560, 111043, "Del2Gru", "鹤十一", "", "天鹤座δ2", "", "Grus", "天鹤座", 337.439583333333, -43.749444444444, 4.11, -0.007, 0},
		{9109, 399, "", "", "", "", "", "", "", 1.233333333333, 26.648888888889, 6.25, 0.11, -0.012},
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
		{"进贤增九", 4799, 61558},
		{"昴宿增九", 1140, 17489},
		{"昴宿增八", 1144, 17527},
		{"昴宿三", 1151, 17579},
		{"昴宿增十二", 1180, 17851},
		{"虚梁三", 8610, 111710},
		{"壁宿增十", 4, 476},
		{"刍藁增二", 681, 10826},
		{"云雨增七", 9004, 117245},
		{"王良增十", 9085, 124},
		{"天大将军增二", 390, 6411},
		{"座旗二", 2585, 33485},
		{"天街一", 1388, 20641},
		{"毕宿六", 1411, 20885},
		{"毕宿增十三", 1412, 20894},
		{"天市右垣三", 6008, 79043},
		{"玄戈增二", 5533, 72487},
		{"天津三", 7730, 99639},
		{"天津增三十八", 7735, 99675},
		{"鹤十一", 8560, 111043},
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

func TestStarDataByChineseAmbiguity(t *testing.T) {
	if err := LoadStarData(); err != nil {
		t.Fatal(err)
	}
	if _, err := StarDataByChinese("杵增一"); !errors.Is(err, ErrStarNameAmbiguous) {
		t.Fatalf("StarDataByChinese(杵增一) error = %v, want ErrStarNameAmbiguous", err)
	}
	if _, err := StarDataByChinese("不存在的星名"); !errors.Is(err, ErrStarNotFound) {
		t.Fatalf("StarDataByChinese(不存在的星名) error = %v, want ErrStarNotFound", err)
	}
	if got, err := StarDataByChinese(" 天狼星 "); err != nil || got.HR != 2491 {
		t.Fatalf("StarDataByChinese with whitespace = HR %d err %v, want HR 2491", got.HR, err)
	}
	if got, err := StarDataByChineseIn(" 杵增一 ", " 危宿 "); err != nil || got.HR != 8449 {
		t.Fatalf("StarDataByChineseIn with whitespace = HR %d err %v, want HR 8449", got.HR, err)
	}
	if _, err := StarDataByChineseIn("杵增一", ""); !errors.Is(err, ErrStarNameAmbiguous) {
		t.Fatalf("StarDataByChineseIn(杵增一, \"\") error = %v, want ErrStarNameAmbiguous", err)
	}
	if got, err := StarDataByChineseIn("天狼", ""); err != nil || got.HR != 2491 {
		t.Fatalf("StarDataByChineseIn(天狼, \"\") = HR %d err %v, want HR 2491", got.HR, err)
	}
	matches, err := StarDataByChineseAll("杵增一")
	if err != nil {
		t.Fatal(err)
	}
	if len(matches) != 2 || matches[0].HR != 6743 || matches[1].HR != 8449 {
		t.Fatalf("StarDataByChineseAll(杵增一) = %v, want HR 6743 and 8449", matches)
	}
	tests := []struct {
		name   string
		region string
		hr     uint16
	}{
		{"杵增一", "箕宿", 6743},
		{"杵增一", "危宿", 8449},
		{"杵一", "箕宿", 6537},
		{"杵一", "危宿", 8498},
		{"柱一", "毕宿", 1605},
		{"柱一", "角宿", 5260},
		{"太子", "太微垣", 4527},
		{"太子", "紫微垣", 5735},
	}
	for _, tc := range tests {
		got, err := StarDataByChineseIn(tc.name, tc.region)
		if err != nil {
			t.Fatalf("StarDataByChineseIn(%q, %q): %v", tc.name, tc.region, err)
		}
		if got.HR != tc.hr || got.ChineseRegion != tc.region {
			t.Fatalf("StarDataByChineseIn(%q, %q) = HR %d region %q, want HR %d region %q",
				tc.name, tc.region, got.HR, got.ChineseRegion, tc.hr, tc.region)
		}
	}
	for _, tc := range []struct {
		name string
		hr   uint16
	}{
		{"杵增一[危宿]", 8449},
		{"杵增一[箕宿]", 6743},
	} {
		got, err := StarDataByChinese(tc.name)
		if err != nil {
			t.Fatalf("StarDataByChinese(%q): %v", tc.name, err)
		}
		if got.HR != tc.hr {
			t.Fatalf("StarDataByChinese(%q) = HR %d, want %d", tc.name, got.HR, tc.hr)
		}
	}
}

func TestStarDataByNameCombinesChineseAndEnglish(t *testing.T) {
	if err := LoadStarData(); err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name string
		hr   uint16
	}{
		{"Sirius", 2491},
		{"sirius", 2491},
		{"  SIRIUS  ", 2491},
		{"Vega", 7001},
		{"Polaris", 424},
		{"Adhil", 390},
		{"天狼", 2491},
		{"天狼星", 2491},
	}
	for _, tc := range tests {
		got, err := StarDataByName(tc.name)
		if err != nil {
			t.Fatalf("StarDataByName(%q): %v", tc.name, err)
		}
		if got.HR != tc.hr {
			t.Fatalf("StarDataByName(%q) = HR %d, want %d", tc.name, got.HR, tc.hr)
		}
	}
	if _, err := StarDataByName("not-a-star"); !errors.Is(err, ErrStarNotFound) {
		t.Fatalf("StarDataByName(not-a-star) error = %v, want ErrStarNotFound", err)
	}
	if _, err := StarDataByName("Gienah"); !errors.Is(err, ErrStarNameAmbiguous) {
		t.Fatalf("StarDataByName(Gienah) error = %v, want ErrStarNameAmbiguous", err)
	}
	matches, err := StarDataByNameAll("gienah")
	if err != nil {
		t.Fatal(err)
	}
	if len(matches) != 2 || matches[0].HR != 4662 || matches[1].HR != 7949 {
		t.Fatalf("StarDataByNameAll(gienah) = %v, want HR 4662 and 7949", matches)
	}
	if got, err := StarDataByNameIn("sirius", ""); err != nil || got.HR != 2491 {
		t.Fatalf("StarDataByNameIn(sirius, \"\") = HR %d err %v, want HR 2491", got.HR, err)
	}
}

func TestStarDataByHRInvalid(t *testing.T) {
	if err := LoadStarData(); err != nil {
		t.Fatal(err)
	}
	for _, hr := range []int{0, -1, len(cachedStarData) + 1} {
		if _, err := StarDataByHR(hr); !errors.Is(err, ErrStarNotFound) {
			t.Fatalf("StarDataByHR(%d) error = %v, want ErrStarNotFound", hr, err)
		}
	}
}

func BenchmarkInitStarData(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		hr2name = nil
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
