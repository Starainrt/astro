package astro

import (
	"testing"

	"github.com/starainrt/astro/internal/ephem"
)

func TestContextDefaultRefreshAndExistingModelSetter(t *testing.T) {
	previousModel := GetModel()
	previousInstance := ephem.Global()
	t.Cleanup(func() {
		SetModel(previousModel)
		ephem.SetGlobal(previousInstance)
	})
	SetModel(ModelLegacy)
	first, err := NewContext(ModelLegacy)
	if err != nil {
		t.Fatal(err)
	}
	ephem.SetGlobal(ephem.NewInstance(nil))
	SetModel(ModelLegacy)
	if ephem.Global().Model() != ephem.Legacy || GetModel() != ModelLegacy {
		t.Fatal("repeated SetModel retained an incompatible global instance")
	}
	ephem.SetActive(ephem.Semianalytic)
	if GetModel() != ModelSemianalytic || ephem.Global().Model() != ephem.Semianalytic || first.Model() != ModelLegacy {
		t.Fatal("existing model setter broke global refresh or explicit isolation")
	}
}
