package star

import (
	"errors"
	"testing"
)

func TestStarDataByNameAmbiguity(t *testing.T) {
	if err := InitStarDatabase(); err != nil {
		t.Fatal(err)
	}
	if _, err := StarDataByName("杵增一"); !errors.Is(err, ErrStarNameAmbiguous) {
		t.Fatalf("StarDataByName(杵增一) error = %v, want ErrStarNameAmbiguous", err)
	}
	if _, err := StarDataByName("不存在的星名"); !errors.Is(err, ErrStarNotFound) {
		t.Fatalf("StarDataByName(不存在的星名) error = %v, want ErrStarNotFound", err)
	}
	if _, err := StarDataByNameIn("杵增一", ""); !errors.Is(err, ErrStarNameAmbiguous) {
		t.Fatalf("StarDataByNameIn(杵增一, \"\") error = %v, want ErrStarNameAmbiguous", err)
	}
	matches, err := StarDataByNameAll("杵增一")
	if err != nil {
		t.Fatal(err)
	}
	if len(matches) != 2 || matches[0].HR != 6743 || matches[1].HR != 8449 {
		t.Fatalf("StarDataByNameAll(杵增一) = %v, want HR 6743 and 8449", matches)
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
	}
	for _, tc := range tests {
		got, err := StarDataByNameIn(tc.name, tc.region)
		if err != nil {
			t.Fatalf("StarDataByNameIn(%q, %q): %v", tc.name, tc.region, err)
		}
		if got.HR != tc.hr || got.ChineseRegion != tc.region {
			t.Fatalf("StarDataByNameIn(%q, %q) = HR %d region %q, want HR %d region %q",
				tc.name, tc.region, got.HR, got.ChineseRegion, tc.hr, tc.region)
		}
	}
	english, err := StarDataByName("sirius")
	if err != nil {
		t.Fatal(err)
	}
	if english.HR != 2491 {
		t.Fatalf("StarDataByName(sirius) = HR %d, want 2491", english.HR)
	}
	if _, err := StarDataByName("Gienah"); !errors.Is(err, ErrStarNameAmbiguous) {
		t.Fatalf("StarDataByName(Gienah) error = %v, want ErrStarNameAmbiguous", err)
	}

	got, err := StarDataByName("杵增一[危宿]")
	if err != nil {
		t.Fatal(err)
	}
	if got.HR != 8449 {
		t.Fatalf("StarDataByName(杵增一[危宿]) = HR %d, want 8449", got.HR)
	}
}
