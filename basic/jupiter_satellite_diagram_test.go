package basic

import (
	"math"
	"reflect"
	"testing"

	"github.com/starainrt/astro/internal/ephem"
)

func TestJupiterGalileanDiagramConsistencyProduction(t *testing.T) {
	for _, jde := range []float64{2461466.5, 2461555.5, 2461586.5} {
		got, ok := JupiterGalileanDiagramWithInstance(jde, nil)
		if !ok {
			t.Fatal("missing diagram")
		}
		context := newJupiterGalileanObservationContext(jde)
		for index := range got.Observations {
			if !reflect.DeepEqual(got.Observations[index], context.observationForSatellite(index)) || got.RadiiJupiterR[index] != jupiterGalileanSatelliteRadiusJupiterRadii(index+1) {
				t.Fatalf("satellite %d snapshot differs", index+1)
			}
			wantAngular := math.Atan2(got.RadiiJupiterR[index]*jupiterGalileanEquatorialRadiusKM/astronomicalUnitKM, got.Observations[index].Distance) / math.Atan2(jupiterGalileanEquatorialRadiusKM/astronomicalUnitKM, context.jupiterDistance)
			if got.AngularRadiiJupiterR[index] != wantAngular {
				t.Fatal("angular radius ratio differs")
			}
		}
		if got.MinorRadiusRatio != context.earthMinorRadius {
			t.Fatal("projected limb differs")
		}
	}
}

func TestJupiterGalileanMutualDiagramConsistencyProduction(t *testing.T) {
	for _, kind := range []JupiterGalileanMutualPhenomenonType{JupiterGalileanMutualOccultation, JupiterGalileanMutualEclipse} {
		event := NextJupiterGalileanMutualEventWithInstance(2461466.5, 1, 3, kind, nil)
		if !event.Valid {
			t.Fatal("missing fixture")
		}
		for _, jde := range []float64{event.Start, event.Greatest, event.End} {
			got, ok := JupiterGalileanMutualDiagramWithInstance(jde, 1, 3, kind, nil)
			reversed, reverseOK := JupiterGalileanMutualDiagramWithInstance(jde, 3, 1, kind, nil)
			if !ok || !reverseOK || !reflect.DeepEqual(got, reversed) || got.Active != event.Active || got.Passive != event.Passive {
				t.Fatalf("pair/role mismatch: %+v", got)
			}
			if kind == JupiterGalileanMutualEclipse {
				geometry, _ := jupiterGalileanMutualEclipseGeometryAt(got.GeometryJDE, 1, 3, nil)
				if got.UmbraRadiusJupiterR != geometry.UmbraKM/jupiterGalileanEquatorialRadiusKM || got.PenumbraRadiusJupiterR != geometry.PenumbraKM/jupiterGalileanEquatorialRadiusKM || got.Coverage != geometry.Coverage {
					t.Fatal("cone/coverage differs")
				}
				if math.Abs(math.Hypot(got.XJupiterR, got.YJupiterR)-geometry.SeparationKM/jupiterGalileanEquatorialRadiusKM) > 1e-12 || math.Abs(jupiterGalileanEclipseReception(got.GeometryJDE, got.Passive, nil)-jde)*86400 > 0.001 {
					t.Fatal("projection or reception time differs")
				}
			} else {
				geometry, _ := jupiterGalileanMutualGeometryAt(jde, 1, 3)
				if got.XJupiterR != geometry.PassiveX-geometry.ActiveX || got.YJupiterR != geometry.PassiveY-geometry.ActiveY || got.Coverage != geometry.Coverage {
					t.Fatal("occultation projection differs")
				}
			}
		}
	}
}

func TestJupiterGalileanDiagramSyntheticInvalid(t *testing.T) {
	if got, ok := JupiterGalileanDiagramWithInstance(math.NaN(), nil); ok || !reflect.DeepEqual(got, JupiterGalileanDiagram{}) {
		t.Fatal("invalid system is not zero")
	}
	for _, pair := range [][2]int{{0, 1}, {1, 1}, {1, 5}} {
		if got, ok := JupiterGalileanMutualDiagramWithInstance(2461466.5, pair[0], pair[1], JupiterGalileanMutualEclipse, nil); ok || !reflect.DeepEqual(got, JupiterGalileanMutualDiagram{}) {
			t.Fatal("invalid pair is not zero")
		}
	}
}

func TestJupiterGalileanDiagramConsistencyInstances(t *testing.T) {
	for _, model := range []ephem.Model{ephem.Legacy, ephem.Semianalytic} {
		instance := ephem.NewModelInstance(model)
		jde := 2461466.5
		got, ok := JupiterGalileanDiagramWithInstance(jde, instance)
		if !ok || !reflect.DeepEqual(got.Observations, JupiterGalileanSatelliteObservationsWithInstance(jde, instance)) {
			t.Fatal("explicit snapshot differs")
		}
		for _, kind := range []JupiterGalileanMutualPhenomenonType{JupiterGalileanMutualOccultation, JupiterGalileanMutualEclipse} {
			got, ok := JupiterGalileanMutualDiagramWithInstance(jde, 1, 3, kind, instance)
			if !ok || !isFinite(got.MinimumSeparationArcsec) {
				t.Fatal("explicit mutual geometry invalid")
			}
		}
	}
	defaultView, _ := JupiterGalileanDiagramWithInstance(2461466.5, nil)
	explicit, _ := JupiterGalileanDiagramWithInstance(2461466.5, ephem.Global())
	if !reflect.DeepEqual(defaultView, explicit) {
		t.Fatal("nil instance differs from process instance")
	}
}
