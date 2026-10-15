package calendar

import (
	"testing"
	"time"
)

func addBase(t *testing.T, y, m, d int) Time {
	t.Helper()
	res, err := LunarToSolarByYMD(y, 1, 1, false) // 只是拿一个合法 Time，日期另行构造
	_ = res
	if err != nil {
		t.Fatal(err)
	}
	fwd, err := SolarToLunarByYMD(y, m, d)
	if err != nil {
		t.Fatalf("SolarToLunarByYMD(%d,%d,%d) error: %v", y, m, d, err)
	}
	return fwd
}

// Add 必须与 time.Time 一样精确，不得经儒略日往返引入微秒漂移。
func TestTimeAddIsExact(t *testing.T) {
	base := addBase(t, 2024, 3, 1)
	cases := []struct {
		d    time.Duration
		want string
	}{
		{time.Millisecond, "2024-03-01 00:00:00.001"},
		{time.Second, "2024-03-01 00:00:01.000"},
		{time.Minute, "2024-03-01 00:01:00.000"},
		{time.Hour, "2024-03-01 01:00:00.000"},
		{24 * time.Hour, "2024-03-02 00:00:00.000"},
		{30 * 24 * time.Hour, "2024-03-31 00:00:00.000"},
		{-time.Hour, "2024-02-29 23:00:00.000"},
		{-time.Millisecond, "2024-02-29 23:59:59.999"},
	}
	for _, c := range cases {
		if got := base.Add(c.d).Solar().Format("2006-01-02 15:04:05.000"); got != c.want {
			t.Errorf("Add(%v) = %s, want %s", c.d, got, c.want)
		}
	}
}

// 1582 改历跳过的 10 天不能出现在结果里：跨过该窗口的偏移要落在 10 天之后。
// The ten days skipped by the 1582 reform must never appear: an offset crossing that window lands
// after them.
func TestTimeAddSkipsGregorianReformGap(t *testing.T) {
	cases := []struct {
		y, m, d int
		delta   time.Duration
		want    string
	}{
		{1582, 10, 4, 24 * time.Hour, "1582-10-15"},
		{1582, 10, 15, -24 * time.Hour, "1582-10-04"},
		{1582, 9, 20, 30 * 24 * time.Hour, "1582-10-30"},
		{1582, 10, 20, -30 * 24 * time.Hour, "1582-09-10"},
		{1582, 9, 20, 24 * time.Hour, "1582-09-21"},
		{1582, 10, 20, 24 * time.Hour, "1582-10-21"},
	}
	for _, c := range cases {
		got := addBase(t, c.y, c.m, c.d).Add(c.delta).Solar().Format("2006-01-02")
		if got != c.want {
			t.Errorf("%d-%02d-%02d Add(%v) = %s, want %s", c.y, c.m, c.d, c.delta, got, c.want)
		}
	}
}

func TestAddOutsideLunarRangeKeepsSolarTime(t *testing.T) {
	base, err := SolarToLunarByYMD(-721, 1, 15)
	if err != nil {
		t.Fatal(err)
	}
	got := base.Add(-30 * 24 * time.Hour)
	want := base.Solar().Add(-30 * 24 * time.Hour)
	if !got.Solar().Equal(want) {
		t.Fatalf("Solar() = %v, want %v", got.Solar(), want)
	}
	if len(got.Lunars()) != 0 {
		t.Fatalf("Lunars() = %v, want empty", got.Lunars())
	}
}
