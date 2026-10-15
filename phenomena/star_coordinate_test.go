package phenomena

import (
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
	if _, err := StarCoordinateByName("不存在的恒星"); err == nil {
		t.Fatal("StarCoordinateByName() expected an error")
	}
	if _, err := StarCoordinateFromStarData(StarData{}); err == nil {
		t.Fatal("StarCoordinateFromStarData() expected an error")
	}
}
