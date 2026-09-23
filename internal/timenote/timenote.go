// Package timenote 生成出图时写入图内或图注的时标声明。
package timenote

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/starainrt/astro"
)

const english = "en"

// Scale 生成图内时标声明：UTC 口径写显示时区与 UTC 偏移，UT1 口径补 DUT1 差值 / builds the in-figure time-scale declaration.
func Scale(scale astro.TimeScale, instant time.Time, location *time.Location, language string) string {
	if scale == astro.TimeScaleUT1 {
		dut1 := astro.DUT1(instant)
		if language == english {
			return fmt.Sprintf("Times are UT1 (Universal Time 1); DUT1 = UT1−UTC = %+.2f s.", dut1)
		}
		return fmt.Sprintf("图中时刻为 UT1（世界时），DUT1 = UT1−UTC = %+.2f s。", dut1)
	}
	zone, offset := displayZone(instant, location)
	if offset == 0 {
		if language == english {
			return "All times are UTC"
		}
		return "图中时刻为 UTC"
	}
	if named, ok := namedZoneOffset(zone); ok && named == offset {
		// 时区名本身就是偏移标签（FixedZone("UTC+08:00")、"Etc/GMT+8" 之类）时再写一遍偏移就是重复文案。
		if language == english {
			return fmt.Sprintf("All times are UTC (shown in %s)", zone)
		}
		return fmt.Sprintf("图中时刻为 UTC（显示时区 %s）", zone)
	}
	sign := "+"
	if offset < 0 {
		sign, offset = "-", -offset
	}
	hours, minutes := offset/3600, offset%3600/60
	if language == english {
		return fmt.Sprintf("All times are UTC (shown in %s, UTC%s%02d:%02d)", zone, sign, hours, minutes)
	}
	return fmt.Sprintf("图中时刻为 UTC（显示时区 %s，UTC%s%02d:%02d）", zone, sign, hours, minutes)
}

// Merge 把时标声明并进图注，已含声明时不重复 / appends the declaration to a footer note once.
func Merge(text, note string) string {
	if note == "" {
		return text
	}
	if text == "" {
		return note
	}
	if strings.Contains(text, note) {
		return text
	}
	return text + " " + note
}

// namedZoneOffset 解析时区名里写死的 UTC 偏移，秒为单位；名字不含偏移时返回 false。
func namedZoneOffset(name string) (int, bool) {
	rest := strings.ToUpper(name)
	for _, prefix := range []string{"UTC", "GMT", "UT"} {
		if strings.HasPrefix(rest, prefix) {
			rest = rest[len(prefix):]
			break
		}
	}
	if rest == "" {
		return 0, false
	}
	sign := 1
	switch rest[0] {
	case '+':
		rest = rest[1:]
	case '-':
		sign, rest = -1, rest[1:]
	default:
		return 0, false
	}
	digits := strings.ReplaceAll(rest, ":", "")
	if digits == "" || len(digits) > 6 {
		return 0, false
	}
	for _, symbol := range digits {
		if symbol < '0' || symbol > '9' {
			return 0, false
		}
	}
	hours, minutes, seconds := 0, 0, 0
	switch len(digits) {
	case 1, 2:
		hours, _ = strconv.Atoi(digits)
	case 3, 4:
		hours, _ = strconv.Atoi(digits[:len(digits)-2])
		minutes, _ = strconv.Atoi(digits[len(digits)-2:])
	case 5, 6:
		hours, _ = strconv.Atoi(digits[:len(digits)-4])
		minutes, _ = strconv.Atoi(digits[len(digits)-4 : len(digits)-2])
		seconds, _ = strconv.Atoi(digits[len(digits)-2:])
	}
	if minutes > 59 || seconds > 59 {
		return 0, false
	}
	return sign * (hours*3600 + minutes*60 + seconds), true
}

func displayZone(instant time.Time, location *time.Location) (string, int) {
	if location != nil {
		instant = instant.In(location)
	}
	zone, offset := instant.Zone()
	if zone == "" {
		zone = "UTC"
	}
	return zone, offset
}
