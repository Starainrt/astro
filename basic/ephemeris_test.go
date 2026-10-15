// 被测契约：默认构建以旧级数开始；切换模型使缓存世代递增，重复设置同一模型不递增，未知模型被拒绝。
package basic

import (
	"testing"

	"github.com/starainrt/astro/internal/ephem"
)

func TestEphemerisDefaultIsBuildDefault(t *testing.T) {
	if ephem.BuildDefault != ephem.Legacy && ephem.BuildDefault != ephem.Semianalytic {
		t.Fatalf("build default %d", ephem.BuildDefault)
	}
	if GetEphemerisModel() != EphemerisModel(ephem.BuildDefault) {
		t.Fatalf("default model %d, build default %d", GetEphemerisModel(), ephem.BuildDefault)
	}
}

func TestSetEphemerisModelBumpsGeneration(t *testing.T) {
	if SetEphemerisModel(EphemerisModel(99)) {
		t.Fatal("unknown model accepted")
	}
	prev := GetEphemerisModel()
	t.Cleanup(func() { SetEphemerisModel(prev) })
	SetEphemerisModel(EphemerisLegacy)
	g := deltaTGenerationValue()
	SetEphemerisModel(EphemerisSemianalytic)
	if deltaTGenerationValue() == g {
		t.Fatal("generation did not change")
	}
	g = deltaTGenerationValue()
	SetEphemerisModel(EphemerisSemianalytic)
	if deltaTGenerationValue() != g {
		t.Fatal("setting the same model changed the generation")
	}
	before := HMoonTrueLo(2451545.0)
	SetEphemerisModel(EphemerisLegacy)
	if HMoonTrueLo(2451545.0) == before {
		t.Fatal("moon longitude did not follow the model switch")
	}
}
