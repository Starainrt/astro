package satellite

import (
	"errors"
	"math"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestParseTLEExternal(t *testing.T) {
	const decodingTolerance = 1e-12
	var maxDiff float64
	count := 0
	for _, f := range readElementFixtures(t) {
		if f.Format != "tle" && f.Format != "3le" {
			continue
		}
		count++
		t.Run(f.Label, func(t *testing.T) {
			lines := strings.Split(strings.ReplaceAll(f.Raw, "\r\n", "\n"), "\n")
			i := 0
			for !strings.HasPrefix(lines[i], "1 ") {
				i++
			}
			got, err := ParseTLE(lines[i], lines[i+1])
			if err != nil {
				t.Fatal(err)
			}
			want := f.Expected[0]
			want.Name = ""
			gv, wv := reflect.ValueOf(got), reflect.ValueOf(want)
			for j := 0; j < gv.NumField(); j++ {
				a, b := gv.Field(j), wv.Field(j)
				if a.Kind() == reflect.Float64 {
					diff := math.Abs(a.Float() - b.Float())
					if diff > maxDiff {
						maxDiff = diff
					}
					if math.IsNaN(a.Float()) || math.IsInf(a.Float(), 0) || diff > decodingTolerance {
						t.Errorf("%s=%v want %v", gv.Type().Field(j).Name, a, b)
					}
				} else if !reflect.DeepEqual(a.Interface(), b.Interface()) {
					t.Errorf("%s=%v want %v", gv.Type().Field(j).Name, a, b)
				}
			}
		})
	}
	if count != 12 {
		t.Fatalf("snapshot count=%d want 12", count)
	}
	t.Logf("%d snapshots; max numeric decoding difference %.12g", count, maxDiff)
}

func tleSyntheticLines(t *testing.T) (string, string) {
	t.Helper()
	f := readElementFixtures(t)[0]
	lines := strings.Split(strings.ReplaceAll(f.Raw, "\r\n", "\n"), "\n")
	for i, line := range lines {
		if strings.HasPrefix(line, "1 ") {
			return line, lines[i+1]
		}
	}
	t.Fatal("missing two-line record")
	return "", ""
}

func tleSyntheticReplace(line string, start, end int, value string) string {
	line = line[:start] + value + line[end:68]
	sum := 0
	for i := range line {
		if line[i] >= '0' && line[i] <= '9' {
			sum += int(line[i] - '0')
		} else if line[i] == '-' {
			sum++
		}
	}
	return line + string(byte('0'+sum%10))
}

func TestUnsupportedEphemerisType(t *testing.T) {
	line1, line2 := tleSyntheticLines(t)
	accepted, err := ParseTLE(tleSyntheticReplace(line1, 62, 63, "2"), line2)
	if err != nil || accepted.EphemerisType != 2 {
		t.Fatalf("explicit SGP4 type: %+v, %v", accepted, err)
	}
	for typ := 0; typ <= 9; typ++ {
		if typ == 0 || typ == 2 {
			continue
		}
		if _, err := ParseTLE(tleSyntheticReplace(line1, 62, 63, string(rune('0'+typ))), line2); !errors.Is(err, ErrInvalidElements) {
			t.Errorf("TLE type %d: %v", typ, err)
		}
		elements := accepted
		elements.EphemerisType = typ
		if _, err := Propagate(elements, elements.Epoch); !errors.Is(err, ErrInvalidElements) {
			t.Errorf("Propagate type %d: %v", typ, err)
		}
		if _, err := LookAt(elements, Observer{}, elements.Epoch, ObserveOptions{}); !errors.Is(err, ErrInvalidElements) {
			t.Errorf("LookAt type %d: %v", typ, err)
		}
	}
}

func TestParseTLEFieldsSynthetic(t *testing.T) {
	a, b := tleSyntheticLines(t)
	for _, tc := range []struct {
		label         string
		first, second string
		check         func(Elements) bool
	}{
		{"alpha-first", tleSyntheticReplace(a, 2, 7, "A0000"), tleSyntheticReplace(b, 2, 7, "A0000"), func(e Elements) bool { return e.CatalogNumber == 100000 }},
		{"alpha-last", tleSyntheticReplace(a, 2, 7, "Z9999"), tleSyntheticReplace(b, 2, 7, "Z9999"), func(e Elements) bool { return e.CatalogNumber == 339999 }},
		{"negative-exponents", tleSyntheticReplace(tleSyntheticReplace(a, 44, 52, "-12345-6"), 53, 61, "+54321+2"), b, func(e Elements) bool {
			return math.Abs(e.MeanMotionDDot+0.00000012345) < 1e-20 && math.Abs(e.BStar-54.321) < 1e-12
		}},
		{"blank-designator", tleSyntheticReplace(a, 9, 17, "        "), b, func(e Elements) bool { return e.InternationalDesignator == "" }},
		{"pivot-1957", tleSyntheticReplace(tleSyntheticReplace(a, 18, 32, "57001.00000000"), 9, 17, "57001A  "), b, func(e Elements) bool {
			return e.Epoch.Equal(time.Date(1957, 1, 1, 0, 0, 0, 0, time.UTC)) && e.InternationalDesignator == "1957-001A"
		}},
		{"pivot-2056", tleSyntheticReplace(a, 18, 32, "56366.99999999"), b, func(e Elements) bool { return e.Epoch.Equal(time.Date(2056, 12, 31, 23, 59, 59, 999136000, time.UTC)) }},
		{"leap-2000", tleSyntheticReplace(a, 18, 32, "00366.00000001"), b, func(e Elements) bool { return e.Epoch.Equal(time.Date(2000, 12, 31, 0, 0, 0, 864000, time.UTC)) }},
	} {
		t.Run(tc.label, func(t *testing.T) {
			e, err := ParseTLE("  "+tc.first+"\r\n", tc.second+"  ")
			if err != nil {
				t.Fatal(err)
			}
			if !tc.check(e) {
				t.Fatalf("unexpected fields %+v", e)
			}
		})
	}
}

func TestParseTLETextSynthetic(t *testing.T) {
	a, b := tleSyntheticLines(t)
	cases := []struct {
		label, text string
		names       []string
	}{
		{"two-line", a + "\n" + b, []string{""}},
		{"mixed-crlf", " \r\n 0   First  \r\n" + a + "\r\n" + b + "\r\n\r\n" + a + "\r\n" + b + "\r\n Second \r\n" + a + "\r\n" + b + "\r\n  ", []string{"First", "", "Second"}},
		{"numeric-prefix-names", "1 KUNS-PF\n" + a + "\n" + b + "\n2 TEST-SAT\n" + a + "\n" + b, []string{"1 KUNS-PF", "2 TEST-SAT"}},
	}
	for _, tc := range cases {
		t.Run(tc.label, func(t *testing.T) {
			got, err := ParseTLEText([]byte(tc.text))
			if err != nil {
				t.Fatal(err)
			}
			if len(got) != len(tc.names) {
				t.Fatalf("count=%d", len(got))
			}
			for i, name := range tc.names {
				if got[i].Name != name {
					t.Errorf("name=%q want %q", got[i].Name, name)
				}
			}
		})
	}
}

func TestParseTLEInvalidSynthetic(t *testing.T) {
	a, b := tleSyntheticLines(t)
	cases := []struct{ label, first, second string }{
		{"short", a[:68], b}, {"long", a + "0", b}, {"checksum", a[:68] + string('0'+(a[68]-'0'+1)%10), b},
		{"first-number", tleSyntheticReplace(a, 0, 1, "2"), b}, {"second-number", a, tleSyntheticReplace(b, 0, 1, "1")},
		{"catalog-mismatch", a, tleSyntheticReplace(b, 2, 7, "12345")},
		{"alpha-I", tleSyntheticReplace(a, 2, 7, "I0000"), tleSyntheticReplace(b, 2, 7, "I0000")},
		{"alpha-O", tleSyntheticReplace(a, 2, 7, "O0000"), tleSyntheticReplace(b, 2, 7, "O0000")},
		{"alpha-lower", tleSyntheticReplace(a, 2, 7, "a0000"), tleSyntheticReplace(b, 2, 7, "a0000")},
		{"catalog-zero", tleSyntheticReplace(a, 2, 7, "00000"), tleSyntheticReplace(b, 2, 7, "00000")},
		{"catalog-tail", tleSyntheticReplace(a, 2, 7, "A00A0"), b},
		{"classification", tleSyntheticReplace(a, 7, 8, "X"), b},
		{"designator", tleSyntheticReplace(a, 9, 17, "98A67A  "), b},
		{"epoch-zero", tleSyntheticReplace(a, 18, 32, "26000.00000000"), b},
		{"epoch-nonleap", tleSyntheticReplace(a, 18, 32, "26366.00000000"), b},
		{"epoch-decimal", tleSyntheticReplace(a, 18, 32, "26317,91612726"), b},
		{"dot", tleSyntheticReplace(a, 33, 43, " 0.0000866"), b},
		{"ddot", tleSyntheticReplace(a, 44, 52, " 12x45-3"), b},
		{"bstar", tleSyntheticReplace(a, 53, 61, " 12345 3"), b},
		{"inclination", a, tleSyntheticReplace(b, 8, 16, "181.0000")},
		{"node", a, tleSyntheticReplace(b, 17, 25, "360.0000")},
		{"eccentricity", a, tleSyntheticReplace(b, 26, 33, "123x567")},
		{"perigee", a, tleSyntheticReplace(b, 34, 42, " -1.0000")},
		{"anomaly", a, tleSyntheticReplace(b, 43, 51, "360.0000")},
		{"motion-zero", a, tleSyntheticReplace(b, 52, 63, " 0.00000000")},
		{"motion-exponent", a, tleSyntheticReplace(b, 52, 63, " 1.000000e1")},
		{"element-set", tleSyntheticReplace(a, 64, 68, "9x99"), b},
		{"ephemeris", tleSyntheticReplace(a, 62, 63, "x"), b},
		{"revolution", a, tleSyntheticReplace(b, 63, 68, "123x5")},
	}
	for _, pos := range []int{1, 8, 17, 32, 43, 52, 61, 63} {
		cases = append(cases, struct{ label, first, second string }{"first-separator", tleSyntheticReplace(a, pos, pos+1, "0"), b})
	}
	for _, pos := range []int{1, 7, 16, 25, 33, 42, 51} {
		cases = append(cases, struct{ label, first, second string }{"second-separator", a, tleSyntheticReplace(b, pos, pos+1, "0")})
	}
	for _, tc := range cases {
		t.Run(tc.label, func(t *testing.T) {
			got, err := ParseTLE(tc.first, tc.second)
			if !errors.Is(err, ErrInvalidElements) || !reflect.DeepEqual(got, Elements{}) {
				t.Fatalf("got %+v, %v", got, err)
			}
		})
	}
	for _, text := range []string{"", "Name", a, b, "Name\nOther\n" + a + "\n" + b, a + "\nName", "0 \n" + a + "\n" + b, a + "\n" + b + "\n" + a} {
		got, err := ParseTLEText([]byte(text))
		if !errors.Is(err, ErrInvalidElements) || got != nil {
			t.Errorf("text=%q got %+v, %v", text, got, err)
		}
	}
}

func TestParseTLETextBOMSynthetic(t *testing.T) {
	a, b := tleSyntheticLines(t)
	raw := []byte("\ufeffISS\r\n" + a + "\r\n" + b)
	got, err := ParseTLEText(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Name != "ISS" {
		t.Fatalf("BOM leaked into name: %+v", got)
	}
}
