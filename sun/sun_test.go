package sun

import (
	"fmt"
	"testing"
	"time"
)

func TestSun(t *testing.T) {
	ja, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		t.Fatal(err)
	}
	now, err := time.ParseInLocation("2006-01-02 15:04:05", "2020-01-01 00:00:00", ja)
	if err != nil {
		t.Fatal(err)
	}
	d, err := RiseTime(now, 115, 40, 0, true)
	if err != nil {
		t.Fatal(err)
	}
	if d.Format("2006-01-02 15:04:05") != "2020-01-01 08:41:45" {
		t.Fatal(d.Format("2006-01-02 15:04:05"))
	}
	fmt.Println(CulminationTime(now, 115))
	fmt.Println(DownTime(now, 115, 40, 0, true))
	fmt.Println(MorningTwilight(now, 115, 40, -6))
	fmt.Println(EveningTwilight(now, 115, 40, -6))
}
