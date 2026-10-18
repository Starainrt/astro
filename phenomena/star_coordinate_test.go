package phenomena

import (
	"errors"
	"testing"

	"github.com/starainrt/astro/basic"
)

func TestStarCoordinateLookup(t *testing.T) {
	byHR, err := StarCoordinateByHR(2491)
	if err != nil {
		t.Fatalf("StarCoordinateByHR() error = %v", err)
	}
	byName, err := StarCoordinateByName("天狼")
	if err != nil {
		t.Fatalf("StarCoordinateByName() error = %v", err)
	}
	if byHR != byName {
		t.Fatalf("lookup mismatch: HR=%+v name=%+v", byHR, byName)
	}
	byEnglish, err := StarCoordinateByName("sirius")
	if err != nil {
		t.Fatalf("StarCoordinateByName(sirius) error = %v", err)
	}
	if byHR != byEnglish {
		t.Fatalf("English lookup mismatch: HR=%+v name=%+v", byHR, byEnglish)
	}

	data, err := basic.StarDataByHR(2491)
	if err != nil {
		t.Fatalf("basic star lookup error = %v", err)
	}
	converted, err := StarCoordinateFromStarData(data)
	if err != nil {
		t.Fatalf("StarCoordinateFromStarData() error = %v", err)
	}
	if byHR != converted {
		t.Fatalf("wrapper mismatch: lookup=%+v converted=%+v", byHR, converted)
	}
}

func TestStarCoordinateLookupInvalid(t *testing.T) {
	if _, err := StarCoordinateByHR(0); err == nil {
		t.Fatal("StarCoordinateByHR(0) expected an error")
	}
	if _, err := StarCoordinateByName("不存在的恒星"); !errors.Is(err, ErrStarNotFound) {
		t.Fatalf("StarCoordinateByName(不存在的恒星) error = %v, want ErrStarNotFound", err)
	}
	if _, err := StarCoordinateByName("杵增一"); !errors.Is(err, ErrStarNameAmbiguous) {
		t.Fatalf("StarCoordinateByName(杵增一) error = %v, want ErrStarNameAmbiguous", err)
	}
	if _, err := StarCoordinateFromStarData(StarData{}); err == nil {
		t.Fatal("StarCoordinateFromStarData() expected an error")
	}
}
