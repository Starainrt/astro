package basic

import (
	"fmt"
	"math"
	"testing"
)

func Test_EarthFn(t *testing.T) {
	fmt.Println(HeightDistance(10000))
	//近似算法，差距在接受范围内？
	fmt.Println(math.Sqrt(((EARTH_AVERAGE_RADIUS)*2 + 10000) * 10000))
}
