package tools

import (
	"fmt"
	"math"
)

func Format(val float64, typed uint8) string {
	belowZero := false
	if val < 0 {
		belowZero = true
		val = -val
	}
	degree := math.Floor(val)
	min := math.Floor((val - degree) * 60)
	sec := (val - degree - min/60) * 3600
	if belowZero {
		degree = -degree
	}
	if typed == 0 {
		return fmt.Sprintf("%.0f°%.0f′%.2f″", degree, min, sec)
	}
	return fmt.Sprintf("%.0fh%.0fm%.2fs", degree, min, sec)
}
