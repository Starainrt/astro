package phenomena_test

import (
	"fmt"
	"time"

	"github.com/starainrt/astro"
	"github.com/starainrt/astro/phenomena"
)

func ExampleSurfacePositionAt() {
	observer := phenomena.BodyObserver{Body: phenomena.Mars, EastLongitude: 137.4, Latitude: -4.6}
	date := time.Date(2026, 11, 14, 0, 0, 0, 0, time.UTC)
	position, err := phenomena.SurfacePositionAt(phenomena.Earth, observer, date)
	fmt.Println(err == nil, position.Altitude >= -90 && position.Altitude <= 90)
	// Output: true true
}

func ExampleNextSurfaceRise() {
	context, err := astro.NewContext(astro.ModelSemianalytic)
	if err != nil {
		panic(err)
	}
	observer := phenomena.BodyObserver{Body: phenomena.Mars, EastLongitude: 137.4, Latitude: -4.6}
	after := time.Date(2026, 11, 14, 0, 0, 0, 0, time.UTC)
	event, err := context.Phenomena().NextSurfaceRise(phenomena.Earth, observer, after, phenomena.SurfaceOptions{})
	if err != nil {
		panic(err)
	}
	local, err := context.Phenomena().SurfaceLocalTimeAt(observer, event.Time)
	fmt.Println(event.Kind == phenomena.SurfaceRise, !event.Time.Before(after), err == nil && event.Local == local)
	// Output: true true true
}

func ExampleNextStarSurfaceCulmination() {
	star, err := phenomena.StarCoordinateByName("Sirius")
	if err != nil {
		panic(err)
	}
	observer := phenomena.BodyObserver{Body: phenomena.Mars, EastLongitude: 137.4, Latitude: -4.6}
	after := time.Date(2026, 11, 14, 0, 0, 0, 0, time.UTC)
	event, err := phenomena.NextStarSurfaceCulmination(star, observer, after)
	fmt.Println(err == nil, event.Kind == phenomena.SurfaceUpperCulmination, !event.Time.Before(after))
	// Output: true true true
}
