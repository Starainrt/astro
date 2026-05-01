package star

import (
	"testing"
	"time"
)

func TestConstellationVariants(t *testing.T) {
	date := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	ra := 88.792939
	dec := 7.407064
	if code := ConstellationCode(ra, dec, date); code != "ORI" {
		t.Fatalf("ConstellationCode() = %q, want %q", code, "ORI")
	}
	if name := ConstellationEN(ra, dec, date); name != "Orion" {
		t.Fatalf("ConstellationEN() = %q, want %q", name, "Orion")
	}
	if name := Constellation(ra, dec, date); name != "猎户座" {
		t.Fatalf("Constellation() = %q, want %q", name, "猎户座")
	}
}
