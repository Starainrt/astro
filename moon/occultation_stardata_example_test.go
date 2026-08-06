package moon_test

import (
	"fmt"

	"github.com/starainrt/astro/moon"
	"github.com/starainrt/astro/star"
)

func ExampleStarCoordinateFromStarData() {
	data, _ := star.StarDataByName("进贤增九")
	coordinate, _ := moon.StarCoordinateFromStarData(data)

	fmt.Printf("%s %.4f %.4f %s\n", coordinate.ID, coordinate.RA, coordinate.Dec, coordinate.Frame)
	// Output:
	// 进贤增九 189.1975 -5.8319 j2000
}
