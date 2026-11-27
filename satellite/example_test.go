package satellite_test

import (
	"fmt"
	"time"

	"github.com/starainrt/astro"
	"github.com/starainrt/astro/satellite"
)

const issExampleTLE = `ISS (ZARYA)
1 25544U 98067A   26317.91612726  .00008668  00000+0  16320-3 0  9995
2 25544  51.6324 279.1461 0007381  22.6178 337.5134 15.49359724594966`

func ExamplePropagate() {
	elements, err := satellite.ParseTLEText([]byte(issExampleTLE))
	if err != nil {
		panic(err)
	}
	iss := elements[0]
	at := iss.Epoch.Add(10 * time.Minute)
	state, err := satellite.Propagate(iss, at)
	if err != nil {
		panic(err)
	}
	point, err := satellite.GroundPointAt(iss, at)
	if err != nil {
		panic(err)
	}
	fmt.Println(iss.CatalogNumber, state.Time.Equal(at), iss.Age(at))
	fmt.Println(point.HeightKM > 300 && point.HeightKM < 500)
	// Output:
	// 25544 true 10m0s
	// true
}

func ExampleSatellite_NextPass() {
	elements, err := satellite.ParseTLEText([]byte(issExampleTLE))
	if err != nil {
		panic(err)
	}
	ctx, err := astro.NewContext(astro.ModelSemianalytic)
	if err != nil {
		panic(err)
	}
	station := satellite.Observer{Longitude: -105, Latitude: 40, Height: 1600}
	pass, err := ctx.Satellite().NextPass(elements[0], station, elements[0].Epoch, satellite.PassOptions{
		ObserveOptions: satellite.ObserveOptions{NoRefraction: true},
		MinElevation:   10,
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(pass.Rise.Time.Before(pass.Set.Time), pass.Culmination.Elevation >= 10)
	// Output:
	// true true
}
