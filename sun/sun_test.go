package sun

import (
	"fmt"
	"testing"
	"time"
)

func TestSun(t *testing.T) {
	now := time.Now()
	fmt.Println(RiseTime(now, 115, 40, 0, true))
	fmt.Println(CulminationTime(now, 115))
	fmt.Println(DownTime(now, 115, 40, 0, true))
}
