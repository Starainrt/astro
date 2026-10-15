package calendar

import (
	"testing"
	"time"
)

func TestGanZhiOfDayUsesCSTCivilDay(t *testing.T) {
	instant := time.Date(2020, 1, 24, 17, 0, 0, 0, time.UTC)
	want := GanZhiOfDay(instant.In(getCst()))
	if got := GanZhiOfDay(instant); got != want {
		t.Fatalf("GanZhiOfDay(UTC) = %s, want %s", got, want)
	}
	if want != "丁卯" {
		t.Fatalf("GanZhiOfDay(2020-01-25 CST) = %s, want 丁卯", want)
	}
}

func TestParseChineseDateAcceptsExplicitGongYuan(t *testing.T) {
	plain, err := parseChineseDate("1988年1月5日")
	if err != nil {
		t.Fatal(err)
	}
	prefixed, err := parseChineseDate("公元1988年1月5日")
	if err != nil {
		t.Fatal(err)
	}
	if prefixed.comment != plain.comment || prefixed.year != plain.year ||
		prefixed.month != plain.month || prefixed.day != plain.day {
		t.Fatalf("公元 prefix parsed as %+v, want %+v", prefixed, plain)
	}
}
