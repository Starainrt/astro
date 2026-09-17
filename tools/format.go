package tools

import (
	"fmt"
	"math"
)

// Format 把角度格式化为度分秒；typed 为 0 时输出度分秒，非 0 时输出时分秒 / format an angle as degrees-minutes-seconds, or hours-minutes-seconds when typed is non-zero.
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
