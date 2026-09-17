package calendar

import (
	"testing"
)

// 年号表的每条记录都必须带朝代，否则 LunarInfo.Dynasty 会随该条记录变空。
// Every era-table record must carry a dynasty, otherwise LunarInfo.Dynasty goes empty for that record.
func TestEraTablesCarryDynasty(t *testing.T) {
	tables := map[string][]Era{
		"hanEras":               hanEras(),
		"weiJinNanBeiChaoEras":  weiJinNanBeiChaoEras(),
		"tangEras":              tangEras(),
		"wudaiSongYuanEras":     wudaiSongYuanEras(),
		"mingQingEras":          mingQingEras(),
		"liaoJinYuanEras":       liaoJinYuanEras(),
		"houQinEras":            houQinEras(),
		"weiZhouSuiEras":        weiZhouSuiEras(),
		"beiLiangEras":          beiLiangEras(),
		"dongWeiBeiQiEras":      dongWeiBeiQiEras(),
		"shuEras":               shuEras(),
		"wuEras":                wuEras(),
		"nanMingEras01":         nanMingEras01(),
		"nanMingEras02":         nanMingEras02(),
		"zhouAncientEras":       zhouAncientEras,
		"luAncientEras":         luAncientEras,
		"qinWarringAncientEras": qinWarringAncientEras,
		"qinHanAncientEras":     qinHanAncientEras,
	}
	for name, eras := range tables {
		for i, e := range eras {
			if e.Dynasty == "" {
				t.Errorf("%s[%d]: Year=%d Emperor=%q Nianhao=%q OtherNianHaoStart=%q has empty Dynasty", name, i, e.Year, e.Emperor, e.Nianhao, e.OtherNianHaoStart)
			}
		}
	}
}

func TestXijinYongxingDynasty(t *testing.T) {
	got, err := SolarToLunarByYMD(304, 6, 1)
	if err != nil {
		t.Fatalf("SolarToLunarByYMD(304,6,1) error: %v", err)
	}
	info := got.LunarInfo()
	if len(info) == 0 {
		t.Fatal("no lunar info")
	}
	for _, li := range info {
		if li.Nianhao == "永兴" && li.Dynasty != "西晋" {
			t.Errorf("永兴 record dynasty = %q, want 西晋", li.Dynasty)
		}
	}
	if got.Lunar().Eras()[0].Dynasty != "西晋" {
		t.Errorf("primary dynasty = %q, want 西晋", got.Lunar().Eras()[0].Dynasty)
	}
}

// 补充年号分支在农历年早于该表首年时，曾追加一条与主流日期完全相同却没有朝代的候选。
func TestSupplementLunarsCarryDynasty(t *testing.T) {
	for _, d := range [][3]int{
		{221, 1, 5}, {222, 1, 5}, {384, 1, 5}, {397, 1, 5}, {534, 1, 5},
		{1647, 1, 1}, {1648, 1, 1},
	} {
		got, err := SolarToLunarByYMD(d[0], d[1], d[2])
		if err != nil {
			t.Fatalf("%v error: %v", d, err)
		}
		for i, l := range got.Lunars() {
			if len(l.Eras()) == 0 {
				t.Errorf("%v record %d: no era", d, i)
			}
			for _, info := range l.LunarInfo() {
				if info.Dynasty == "" {
					t.Errorf("%v record %d: empty dynasty (%q)", d, i, info.EraDesc)
				}
			}
		}
	}
}

// 永历年号表只有一条记录，曾因 innerEras 把“年==表首年”的唯一命中放进 secEra 而整条丢失。
func TestYongliFirstYear(t *testing.T) {
	for _, d := range [][3]int{{1647, 6, 1}, {1648, 1, 1}} {
		got, err := SolarToLunarByYMD(d[0], d[1], d[2])
		if err != nil {
			t.Fatalf("%v error: %v", d, err)
		}
		found := false
		for _, info := range got.LunarInfo() {
			if info.Nianhao != "永历" {
				continue
			}
			found = true
			if info.Dynasty != "南明" || info.Emperor != "南明/明郑" || info.YearOfNianhao != 1 {
				t.Errorf("%v: 永历 record = %+v, want 南明/明郑 永历元年", d, info)
			}
		}
		if !found {
			t.Errorf("%v: 永历 record missing", d)
		}
	}
}
