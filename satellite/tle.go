package satellite

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// ParseTLE 解析两行轨道根数并验证列位与校验和 / parse two element lines and validate their columns and checksums.
func ParseTLE(line1, line2 string) (Elements, error) {
	a, b := strings.TrimSpace(line1), strings.TrimSpace(line2)
	if !tleLineValid(a, '1', []int{1, 8, 17, 32, 43, 52, 61, 63}) || !tleLineValid(b, '2', []int{1, 7, 16, 25, 33, 42, 51}) {
		return Elements{}, fmt.Errorf("%w: line layout or checksum", ErrInvalidElements)
	}
	id, ok := tleCatalog(a[2:7])
	other, otherOK := tleCatalog(b[2:7])
	if !ok || !otherOK || id != other {
		return Elements{}, fmt.Errorf("%w: catalog number", ErrInvalidElements)
	}
	epoch, err := tleEpoch(a[18:32])
	if err != nil {
		return Elements{}, err
	}
	design, ok := tleDesignator(a[9:17])
	if !ok {
		return Elements{}, fmt.Errorf("%w: international designator", ErrInvalidElements)
	}
	e := Elements{CatalogNumber: id, Classification: a[7:8], InternationalDesignator: design, Epoch: epoch}
	fields := []struct {
		raw   string
		point int
		dst   *float64
	}{
		{a[33:43], 1, &e.MeanMotionDot}, {b[8:16], 3, &e.Inclination}, {b[17:25], 3, &e.RightAscension},
		{b[34:42], 3, &e.ArgumentOfPerigee}, {b[43:51], 3, &e.MeanAnomaly}, {b[52:63], 2, &e.MeanMotion},
	}
	for _, f := range fields {
		if !tleDecimal(f.raw, f.point) {
			return Elements{}, fmt.Errorf("%w: decimal field", ErrInvalidElements)
		}
		*f.dst, err = strconv.ParseFloat(strings.TrimSpace(f.raw), 64)
		if err != nil {
			return Elements{}, fmt.Errorf("%w: decimal field", ErrInvalidElements)
		}
	}
	if !tleDigits(b[26:33]) {
		return Elements{}, fmt.Errorf("%w: eccentricity", ErrInvalidElements)
	}
	e.Eccentricity, err = strconv.ParseFloat("0."+b[26:33], 64)
	if err != nil {
		return Elements{}, fmt.Errorf("%w: eccentricity", ErrInvalidElements)
	}
	e.MeanMotionDDot, ok = tleExponent(a[44:52])
	if !ok {
		return Elements{}, fmt.Errorf("%w: second derivative", ErrInvalidElements)
	}
	e.BStar, ok = tleExponent(a[53:61])
	if !ok {
		return Elements{}, fmt.Errorf("%w: drag term", ErrInvalidElements)
	}
	for _, f := range []struct {
		raw string
		dst *int
	}{{a[62:63], &e.EphemerisType}, {a[64:68], &e.ElementSetNumber}, {b[63:68], &e.RevolutionNumber}} {
		value := strings.TrimLeft(f.raw, " ")
		if !tleDigits(value) {
			return Elements{}, fmt.Errorf("%w: integer field", ErrInvalidElements)
		}
		*f.dst, err = strconv.Atoi(value)
		if err != nil {
			return Elements{}, fmt.Errorf("%w: integer field", ErrInvalidElements)
		}
	}
	if !validElements(e) {
		return Elements{}, fmt.Errorf("%w: field range", ErrInvalidElements)
	}
	return e, nil
}

// ParseTLEText 解析混排的两行与三行根数 / parse mixed two-line and named three-line records.
func ParseTLEText(data []byte) ([]Elements, error) {
	data = bytes.TrimPrefix(data, []byte("\xef\xbb\xbf"))
	var out []Elements
	var name, first string
	for _, raw := range strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n") {
		line := strings.TrimSpace(raw)
		if line == "" {
			continue
		}
		if first != "" {
			e, err := ParseTLE(first, line)
			if err != nil {
				return nil, err
			}
			e.Name = name
			out = append(out, e)
			name, first = "", ""
			continue
		}
		if len(line) == 69 && strings.HasPrefix(line, "1 ") {
			first = line
			continue
		}
		if (len(line) == 69 && strings.HasPrefix(line, "2 ")) || name != "" {
			return nil, fmt.Errorf("%w: orphan line", ErrInvalidElements)
		}
		name = strings.TrimSpace(strings.TrimPrefix(line, "0 "))
		if name == "" || line == "0" {
			return nil, fmt.Errorf("%w: empty name line", ErrInvalidElements)
		}
	}
	if first != "" || name != "" || len(out) == 0 {
		return nil, fmt.Errorf("%w: incomplete element text", ErrInvalidElements)
	}
	return out, nil
}

func tleLineValid(line string, number byte, spaces []int) bool {
	if len(line) != 69 || line[0] != number || line[68] < '0' || line[68] > '9' {
		return false
	}
	for _, i := range spaces {
		if line[i] != ' ' {
			return false
		}
	}
	sum := 0
	for i := 0; i < 68; i++ {
		c := line[i]
		if c < 32 || c > 126 {
			return false
		}
		if c >= '0' && c <= '9' {
			sum += int(c - '0')
		} else if c == '-' {
			sum++
		}
	}
	return sum%10 == int(line[68]-'0')
}

func tleDigits(s string) bool {
	if len(s) == 0 {
		return false
	}
	for i := range s {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}

func tleCatalog(s string) (int, bool) {
	if len(s) != 5 {
		return 0, false
	}
	if s[0] >= 'A' && s[0] <= 'Z' {
		const alphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ"
		n := strings.IndexByte(alphabet, s[0])
		if n < 0 || !tleDigits(s[1:]) {
			return 0, false
		}
		tail, _ := strconv.Atoi(s[1:])
		return (n+10)*10000 + tail, true
	}
	s = strings.TrimLeft(s, " ")
	if !tleDigits(s) {
		return 0, false
	}
	n, err := strconv.Atoi(s)
	return n, err == nil && n > 0
}

func tleDecimal(s string, point int) bool {
	if point >= len(s) || s[point] != '.' || !tleDigits(s[point+1:]) {
		return false
	}
	before := strings.TrimLeft(s[:point], " ")
	if point == 1 {
		return s[0] == ' ' || s[0] == '+' || s[0] == '-'
	}
	return tleDigits(before)
}

func tleExponent(s string) (float64, bool) {
	if len(s) != 8 || (s[0] != ' ' && s[0] != '+' && s[0] != '-') || !tleDigits(s[1:6]) || (s[6] != '+' && s[6] != '-') || !tleDigits(s[7:]) {
		return 0, false
	}
	sign := ""
	if s[0] == '-' {
		sign = "-"
	}
	v, err := strconv.ParseFloat(sign+"0."+s[1:6]+"e"+s[6:], 64)
	return v, err == nil
}

func tleYear(year int) int {
	if year >= 57 {
		return 1900 + year
	}
	return 2000 + year
}

func tleEpoch(s string) (time.Time, error) {
	if len(s) != 14 || s[5] != '.' || !tleDigits(s[:5]) || !tleDigits(s[6:]) {
		return time.Time{}, fmt.Errorf("%w: epoch layout", ErrInvalidElements)
	}
	year, _ := strconv.Atoi(s[:2])
	day, _ := strconv.Atoi(s[2:5])
	fraction, _ := strconv.ParseInt(s[6:], 10, 64)
	year = tleYear(year)
	start := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
	days := time.Date(year+1, 1, 1, 0, 0, 0, 0, time.UTC).Sub(start) / (24 * time.Hour)
	if day < 1 || day > int(days) {
		return time.Time{}, fmt.Errorf("%w: epoch day", ErrInvalidElements)
	}
	return start.AddDate(0, 0, day-1).Add(time.Duration(fraction) * 864000), nil
}

func tleDesignator(s string) (string, bool) {
	s = strings.TrimRight(s, " ")
	if s == "" {
		return "", true
	}
	if len(s) < 5 || !tleDigits(s[:5]) {
		return "", false
	}
	for _, c := range s[5:] {
		if c < 'A' || c > 'Z' {
			return "", false
		}
	}
	year, _ := strconv.Atoi(s[:2])
	return fmt.Sprintf("%04d-%s", tleYear(year), s[2:]), true
}
