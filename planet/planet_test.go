package planet

import (
	"fmt"
	"testing"
)

func Test_WherePlanet(t *testing.T) {
	fmt.Println(WherePlanet(0, 0, 2452545.56))
	fmt.Println(WherePlanet(0, 1, 2452545.56))
	fmt.Println(WherePlanet(0, 2, 2452545.56))
	fmt.Println(WherePlanet(1, 0, 2452545.56))
	fmt.Println(WherePlanet(1, 1, 2452545.56))
	fmt.Println(WherePlanet(1, 2, 2452545.56))
	/*
	184.76617673228
1.623745085328E-6
1.0021086365503
5.917793283752
-4.7404269321589
0.34835392797302
*/
}
