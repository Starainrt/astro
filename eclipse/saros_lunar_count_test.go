package eclipse

import (
	"testing"
	"time"
)

// 把某个时间窗内落入指定沙罗系列的月食枚举出来。
func lunarSarosMembersInWindow(t *testing.T, series int, start, end time.Time) []time.Time {
	t.Helper()
	seen := map[int64]bool{}
	members := make([]time.Time, 0, 4)
	for cursor := start; cursor.Before(end); cursor = cursor.AddDate(0, 1, 0) {
		info := ClosestLunarEclipse(cursor)
		if info.Maximum.IsZero() || info.Saros.Series != series {
			continue
		}
		key := info.Maximum.UTC().Unix()
		if seen[key] {
			continue
		}
		seen[key] = true
		members = append(members, info.Maximum)
	}
	return members
}

// 月食沙罗系列的总成员数与成员序号必须与 NASA 目录一致。
// 123 号 72 次（NASA LEsaros123：24N 6P 25T 8P 9N = 72），157 号 73 次；
// 2025-03-14 是 123 号的第 53 个成员。
func TestLunarSarosSeriesMemberCounts(t *testing.T) {
	for _, test := range []struct {
		name   string
		probe  time.Time
		series int
		member int
		count  int
	}{
		{name: "123-2025", probe: time.Date(2025, time.March, 14, 0, 0, 0, 0, time.UTC), series: 123, member: 53, count: 72},
		{name: "157-2360", probe: time.Date(2360, time.April, 2, 0, 0, 0, 0, time.UTC), series: 157, member: 4, count: 73},
	} {
		info := ClosestLunarEclipse(test.probe)
		if info.Saros.Series != test.series || info.Saros.Member != test.member || info.Saros.Count != test.count {
			t.Fatalf("%s: series %d member %d/%d, want %d member %d/%d",
				test.name, info.Saros.Series, info.Saros.Member, info.Saros.Count,
				test.series, test.member, test.count)
		}
	}
}

// 123 号系列的边界必须与 NASA 目录一致：首 1087-08-16、末 2367-10-08，两端之外再无成员。
func TestLunarSarosSeries123Boundaries(t *testing.T) {
	// 系列开始前的三年窗口里不应有任何 123 号成员。
	if members := lunarSarosMembersInWindow(t, 123,
		time.Date(1084, time.January, 1, 0, 0, 0, 0, time.UTC),
		time.Date(1087, time.January, 1, 0, 0, 0, 0, time.UTC)); len(members) != 0 {
		t.Fatalf("series 123 starts before 1087: %v", members)
	}
	if members := lunarSarosMembersInWindow(t, 123,
		time.Date(1087, time.January, 1, 0, 0, 0, 0, time.UTC),
		time.Date(1090, time.January, 1, 0, 0, 0, 0, time.UTC)); len(members) != 1 ||
		members[0].UTC().Format("2006-01-02") != "1087-08-16" {
		t.Fatalf("series 123 first members = %v, want one on 1087-08-16", members)
	}
	if members := lunarSarosMembersInWindow(t, 123,
		time.Date(2367, time.January, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2371, time.January, 1, 0, 0, 0, 0, time.UTC)); len(members) != 1 ||
		members[0].UTC().Format("2006-01-02") != "2367-10-08" {
		t.Fatalf("series 123 last members = %v, want one on 2367-10-08", members)
	}
}
