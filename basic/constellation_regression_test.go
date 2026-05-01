package basic

import (
	_ "embed"
	"encoding/json"
	"testing"
	"time"
)

type constellationBaselineSample struct {
	RA   float64 `json:"ra"`
	Dec  float64 `json:"dec"`
	JDE  float64 `json:"jde"`
	Code string  `json:"code"`
	ZH   string  `json:"zh"`
}

//go:embed testdata/cst_baseline.json
var constellationBaselineJSON []byte

func loadCstBaseline(t *testing.T) []constellationBaselineSample {
	t.Helper()
	var samples []constellationBaselineSample
	if err := json.Unmarshal(constellationBaselineJSON, &samples); err != nil {
		t.Fatalf("unmarshal constellation baseline: %v", err)
	}
	return samples
}

func TestConstellationBaseline(t *testing.T) {
	samples := loadCstBaseline(t)
	for index, sample := range samples {
		if code := resolveConstellationCode(sample.RA, sample.Dec, sample.JDE); code != sample.Code {
			t.Fatalf("sample %d code mismatch: ra=%.6f dec=%.6f jde=%.6f got=%q want=%q", index, sample.RA, sample.Dec, sample.JDE, code, sample.Code)
		}
		if code := ConstellationCode(sample.RA, sample.Dec, sample.JDE); code != sample.Code {
			t.Fatalf("sample %d wrapper mismatch: ra=%.6f dec=%.6f jde=%.6f got=%q want=%q", index, sample.RA, sample.Dec, sample.JDE, code, sample.Code)
		}
		if zh := ConstellationNameZH(sample.RA, sample.Dec, sample.JDE); zh != sample.ZH {
			t.Fatalf("sample %d zh mismatch: ra=%.6f dec=%.6f jde=%.6f got=%q want=%q", index, sample.RA, sample.Dec, sample.JDE, zh, sample.ZH)
		}
	}
}

func TestConstellationNameLookups(t *testing.T) {
	tests := []struct {
		code string
		zh   string
		en   string
	}{
		{code: "ORI", zh: "猎户座", en: "Orion"},
		{code: "PSC", zh: "双鱼座", en: "Pisces"},
		{code: "SER1", zh: "巨蛇座", en: "Serpens Caput"},
		{code: "SER2", zh: "巨蛇座", en: "Serpens Cauda"},
		{code: "UMI", zh: "小熊座", en: "Ursa Minor"},
	}
	for _, testCase := range tests {
		if zh := ConstellationNameByCodeZH(testCase.code); zh != testCase.zh {
			t.Fatalf("ConstellationNameByCodeZH(%q) = %q, want %q", testCase.code, zh, testCase.zh)
		}
		if en := ConstellationNameByCodeEN(testCase.code); en != testCase.en {
			t.Fatalf("ConstellationNameByCodeEN(%q) = %q, want %q", testCase.code, en, testCase.en)
		}
	}
}

func TestConstellationNameEN(t *testing.T) {
	jde := Date2JDE(time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC))
	if en := ConstellationNameEN(88.792939, 7.407064, jde); en != "Orion" {
		t.Fatalf("ConstellationNameEN() = %q, want %q", en, "Orion")
	}
}
