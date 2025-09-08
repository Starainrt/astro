package venus

import (
	"fmt"
	"testing"
	"time"
)

func TestVenus(t *testing.T) {
	date := time.Now().Add(time.Hour * -24)
	fmt.Println(CulminationTime(date, 115))
	fmt.Println(RiseTime(date, 115, 23, 0, false))
	fmt.Println(SetTime(date, 115, 23, 0, false))
}
