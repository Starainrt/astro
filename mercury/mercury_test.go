package mercury

import (
	"fmt"
	"testing"
	"time"
)

func TestMercury(t *testing.T) {
	tz := time.FixedZone("CST", 8*3600)
	date := time.Date(2022, 01, 20, 00, 00, 00, 00, tz)
	if NextConjunction(date).Unix() != 1642933683 {
		t.Fatal(NextConjunction(date).Unix())
	}
	if CulminationTime(date, 115).Unix() != 1642654651 {
		t.Fatal(CulminationTime(date, 115).Unix())
	}
	date, err := (RiseTime(date, 115, 40, 0, false))
	if err != nil {
		t.Fatal(err)
	}
	if date.Unix() != 1642636481 {
		t.Fatal(date.Unix())
	}
	fmt.Println(DownTime(date, 115, 40, 0, false))
}
