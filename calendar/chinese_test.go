package calendar

import (
	"fmt"
	"testing"
)

func Test_Solar(t *testing.T) {
	fmt.Println(Solar(2021, 1, 1, false))
	fmt.Println(Solar(2020, 1, 1, false))
}
