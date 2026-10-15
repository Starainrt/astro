package astro_test

import (
	"fmt"
	"math"
	"time"

	"github.com/starainrt/astro"
)

func ExampleContext() {
	context, err := astro.NewContext(astro.ModelSemianalytic)
	if err != nil {
		panic(err)
	}
	date := time.Date(2024, 10, 6, 12, 0, 0, 0, time.UTC)
	longitude := context.Moon().ApparentLo(date)
	fmt.Println(context.Model() == astro.ModelSemianalytic)
	fmt.Println(!math.IsNaN(longitude) && longitude >= 0 && longitude < 360)
	// Output:
	// true
	// true
}
