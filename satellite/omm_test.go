package satellite

import (
	"encoding/json"
	"errors"
	"math"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"
)

func ommSyntheticJSON(t *testing.T) []byte {
	t.Helper()
	data, err := json.Marshal(map[string]interface{}{
		"OBJECT_NAME": "  Example  ", "OBJECT_ID": "1998-067A", "CLASSIFICATION_TYPE": "U", "NORAD_CAT_ID": 25544,
		"EPOCH": "2026-11-13T21:59:13.395264", "MEAN_MOTION": 15.49359724, "ECCENTRICITY": 0.00073816, "INCLINATION": 51.6324,
		"RA_OF_ASC_NODE": 279.1461, "ARG_OF_PERICENTER": 22.6178, "MEAN_ANOMALY": 337.5134,
		"MEAN_MOTION_DOT": 0.00008668, "MEAN_MOTION_DDOT": 0, "BSTAR": 0.00016319629,
		"ELEMENT_SET_NO": 999, "REV_AT_EPOCH": 59496, "EPHEMERIS_TYPE": 0,
	})
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func TestParseOMMJSONSynthetic(t *testing.T) {
	raw := ommSyntheticJSON(t)
	got, err := ParseOMMJSON(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Name != "Example" || got[0].CatalogNumber != 25544 {
		t.Fatalf("unexpected %+v", got)
	}
	for _, input := range [][]byte{raw, append(append([]byte(" [ "), raw...), ']')} {
		auto, err := ParseOMM(input)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(got, auto) {
			t.Fatalf("auto=%+v want %+v", auto, got)
		}
	}
}

func TestParseOMMXMLSynthetic(t *testing.T) {
	var fixtures []elementFixture
	for _, f := range readElementFixtures(t) {
		if f.Format == "xml" {
			fixtures = append(fixtures, f)
		}
	}
	single := fixtures[0].Raw
	double := "<?xml version=\"1.0\"?><ndm>" + single + fixtures[1].Raw + "</ndm>"
	commented := strings.Replace(double, "<ndm>", "<ndm><COMMENT>Two OMM records</COMMENT><COMMENT>&lt;omm&gt; messages</COMMENT>", 1)
	for _, tc := range []struct {
		raw   string
		count int
	}{{single, 1}, {double, 2}, {commented, 2}} {
		got, err := ParseOMMXML([]byte(tc.raw))
		if err != nil {
			t.Fatal(err)
		}
		if len(got) != tc.count {
			t.Fatalf("count=%d want %d", len(got), tc.count)
		}
		for index, record := range got {
			if record != fixtures[index].Expected[0] {
				t.Fatalf("record %d: got %+v want %+v", index, record, fixtures[index].Expected[0])
			}
		}
		auto, err := ParseOMM([]byte(" \r\n" + tc.raw))
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(got, auto) {
			t.Fatal("auto XML differs")
		}
	}
}

func TestParseOMMXMLStructure(t *testing.T) {
	var raw string
	for _, fixture := range readElementFixtures(t) {
		if fixture.Label == "iss-xml" {
			raw = fixture.Raw
			break
		}
	}
	const theory = "<MEAN_ELEMENT_THEORY>SGP4</MEAN_ELEMENT_THEORY>"
	const inclination = "<INCLINATION>51.6324</INCLINATION>"
	const ephemerisType = "<EPHEMERIS_TYPE>0</EPHEMERIS_TYPE>"
	for _, tc := range []struct {
		name, remove, before, after string
	}{
		{"theory-in-mean-elements", theory, "<meanElements>", "<meanElements>" + theory},
		{"inclination-in-tle-parameters", inclination, "<tleParameters>", "<tleParameters>" + inclination},
		{"type-in-header", ephemerisType, "<header>", "<header>" + ephemerisType},
		{"split-segments", "", "</metadata><data>", "</metadata></segment><segment><data>"},
		{"extra-segment", "", "</body>", "<segment/></body>"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if !strings.Contains(raw, tc.remove) || !strings.Contains(raw, tc.before) {
				t.Fatal("fixture is missing XML field or section")
			}
			invalid := strings.Replace(strings.Replace(raw, tc.remove, "", 1), tc.before, tc.after, 1)
			for _, input := range []string{invalid, "<ndm>" + raw + invalid + "</ndm>"} {
				got, err := ParseOMMXML([]byte(input))
				if got != nil || !errors.Is(err, ErrInvalidElements) {
					t.Errorf("invalid XML accepted: %+v, %v", got, err)
				}
			}
		})
	}
	got, err := ParseOMMXML([]byte("<ndm><COMMENT>OMM records</COMMENT>" + raw + "<oem/></ndm>"))
	if got != nil || !errors.Is(err, ErrInvalidElements) {
		t.Errorf("unsupported NDM record accepted: %+v, %v", got, err)
	}
}

func TestParseOMMKVNSynthetic(t *testing.T) {
	var fixtures []elementFixture
	for _, f := range readElementFixtures(t) {
		if f.Format == "kvn" {
			fixtures = append(fixtures, f)
		}
	}
	single := fixtures[0].Raw
	multi := "COMMENT text without equals\r\n" + single + "\nCOMMENT between records\n" + fixtures[1].Raw
	units := strings.Replace(single, "15.49359724", "15.49359724 [rev/day]", 1)
	for _, tc := range []struct {
		raw   string
		count int
	}{{single, 1}, {multi, 2}, {units, 1}} {
		got, err := ParseOMMKVN([]byte(tc.raw))
		if err != nil {
			t.Fatal(err)
		}
		if len(got) != tc.count {
			t.Fatalf("count=%d want %d", len(got), tc.count)
		}
		auto, err := ParseOMM([]byte(" \r\n" + tc.raw))
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(got, auto) {
			t.Fatal("auto KVN differs")
		}
	}
}

func TestParseOMMOrdinalEpoch(t *testing.T) {
	for _, fixture := range readElementFixtures(t) {
		if fixture.Label != "iss-json" && fixture.Label != "iss-xml" && fixture.Label != "iss-kvn" {
			continue
		}
		for _, suffix := range []string{"", "Z"} {
			t.Run(fixture.Label+suffix, func(t *testing.T) {
				raw := strings.Replace(fixture.Raw, "2026-11-13T21:59:13.395264", "2026-317T21:59:13.395264"+suffix, 1)
				if raw == fixture.Raw {
					t.Fatal("missing fixture epoch")
				}
				got, err := ParseOMM([]byte(raw))
				if err != nil || !reflect.DeepEqual(got, fixture.Expected) {
					t.Fatalf("ordinal and calendar epochs differ: %+v, %v", got, err)
				}
			})
		}
	}
	got, err := ParseOMMJSON(ommJSONChange(t, "EPOCH", "2024-366T00:00:00Z", false))
	if err != nil || len(got) != 1 || got[0].Epoch != time.Date(2024, time.December, 31, 0, 0, 0, 0, time.UTC) {
		t.Fatalf("leap-year final day: %+v, %v", got, err)
	}
}

func TestParseOMMStandardFieldsSynthetic(t *testing.T) {
	for _, fixture := range readElementFixtures(t) {
		if fixture.Label != "iss-xml" && fixture.Label != "iss-kvn" {
			continue
		}
		t.Run(fixture.Format, func(t *testing.T) {
			raw := fixture.Raw
			fields := []string{
				"<EPHEMERIS_TYPE>0</EPHEMERIS_TYPE>", "<CLASSIFICATION_TYPE>U</CLASSIFICATION_TYPE>",
				"<ELEMENT_SET_NO>999</ELEMENT_SET_NO>", "<REV_AT_EPOCH>59496</REV_AT_EPOCH>",
				"<MEAN_MOTION_DOT>.8668E-4</MEAN_MOTION_DOT>", "<MEAN_MOTION_DDOT>0</MEAN_MOTION_DDOT>",
			}
			if fixture.Format == "kvn" {
				raw = strings.Replace(raw, "SGP/SGP4", "SGP4", 1)
				fields = []string{
					"EPHEMERIS_TYPE = 0\r\n", "CLASSIFICATION_TYPE = U\r\n", "ELEMENT_SET_NO = 999\r\n",
					"REV_AT_EPOCH   = 59496\r\n", "MEAN_MOTION_DOT = .8668E-4\r\n", "MEAN_MOTION_DDOT = 0\r\n",
				}
			}
			for _, field := range fields {
				if !strings.Contains(raw, field) {
					t.Fatalf("fixture is missing %q", field)
				}
				raw = strings.Replace(raw, field, "", 1)
			}
			want := fixture.Expected[0]
			want.ElementSetNumber, want.RevolutionNumber, want.MeanMotionDot, want.MeanMotionDDot = 0, 0, 0, 0
			got, err := ParseOMM([]byte(raw))
			if err != nil || len(got) != 1 || got[0] != want {
				t.Fatalf("optional fields: got %+v, %v; want %+v", got, err, want)
			}
		})
	}
}

func TestParseOMMDeclaredUnitsSynthetic(t *testing.T) {
	for _, fixture := range readElementFixtures(t) {
		var field, withUnits string
		var invalid []string
		switch fixture.Label {
		case "iss-xml":
			field = `<INCLINATION>51.6324</INCLINATION>`
			withUnits = `<INCLINATION units="deg">51.6324</INCLINATION>`
			invalid = []string{`<INCLINATION units="rad">0.901155</INCLINATION>`, `<INCLINATION units="">51.6324</INCLINATION>`}
		case "iss-kvn":
			field = "INCLINATION    = 51.6324"
			withUnits = field + " [deg]"
			invalid = []string{"INCLINATION    = 0.901155 [rad]", field + " []", field + "[deg]"}
		default:
			continue
		}
		t.Run(fixture.Format, func(t *testing.T) {
			if !strings.Contains(fixture.Raw, field) {
				t.Fatalf("fixture is missing %q", field)
			}
			got, err := ParseOMM([]byte(strings.Replace(fixture.Raw, field, withUnits, 1)))
			if err != nil || !reflect.DeepEqual(got, fixture.Expected) {
				t.Fatalf("correct units: %+v, %v", got, err)
			}
			for _, bad := range invalid {
				got, err := ParseOMM([]byte(strings.Replace(fixture.Raw, field, bad, 1)))
				if got != nil || !errors.Is(err, ErrInvalidElements) {
					t.Errorf("%q: got %+v, %v", bad, got, err)
				}
			}
		})
	}
}

func TestParseOMMJSONDuplicateFieldsSynthetic(t *testing.T) {
	raw := string(ommSyntheticJSON(t))
	for _, body := range []string{
		strings.Replace(raw, `"EPHEMERIS_TYPE":0`, `"EPHEMERIS_TYPE":4,"EPHEMERIS_TYPE":0`, 1),
		`{"MEAN_ELEMENT_THEORY":"SGP4-XP","MEAN_ELEMENT_THEORY":"SGP4",` + raw[1:],
		strings.Replace(raw, `"BSTAR":`, `"BST\u0041R":0,"BSTAR":`, 1),
	} {
		if body == raw {
			t.Fatal("fixture did not change")
		}
		for _, input := range []string{body, "[" + raw + "," + body + "]"} {
			got, err := ParseOMMJSON([]byte(input))
			if got != nil || !errors.Is(err, ErrInvalidElements) {
				t.Errorf("duplicate field accepted: %+v, %v", got, err)
			}
		}
	}
}

func TestParseOMMExternal(t *testing.T) {
	const exactTolerance = 0.0
	var maxDiff float64
	count, supgp := 0, 0
	for _, f := range readElementFixtures(t) {
		var parse func([]byte) ([]Elements, error)
		switch f.Format {
		case "json":
			parse = ParseOMMJSON
		case "xml":
			parse = ParseOMMXML
		case "kvn":
			parse = ParseOMMKVN
		default:
			continue
		}
		t.Run(f.Label, func(t *testing.T) {
			got, err := parse([]byte(f.Raw))
			if err != nil {
				t.Fatal(err)
			}
			if len(got) != len(f.Expected) {
				t.Fatalf("count=%d want %d", len(got), len(f.Expected))
			}
			count += len(got)
			if f.Label == "iss-supgp" {
				supgp = len(got)
			}
			for i, e := range got {
				gv, wv := reflect.ValueOf(e), reflect.ValueOf(f.Expected[i])
				for j := 0; j < gv.NumField(); j++ {
					a, b := gv.Field(j), wv.Field(j)
					if a.Kind() == reflect.Float64 {
						diff := math.Abs(a.Float() - b.Float())
						if diff > maxDiff {
							maxDiff = diff
						}
						if math.IsNaN(a.Float()) || math.IsInf(a.Float(), 0) || diff > exactTolerance {
							t.Errorf("record %d %s=%v want %v", i, gv.Type().Field(j).Name, a, b)
						}
					} else if !reflect.DeepEqual(a.Interface(), b.Interface()) {
						t.Errorf("record %d %s=%v want %v", i, gv.Type().Field(j).Name, a, b)
					}
				}
			}
		})
	}
	if count != 81 || supgp != 60 {
		t.Fatalf("count=%d supgp=%d", count, supgp)
	}
	t.Logf("81 element records including 60 supplemental segments; max numeric difference %.12g", maxDiff)
}

func TestElementsFormatsExternal(t *testing.T) {
	const angleResolution = 1e-4
	const eccentricityResolution = 1e-7
	const motionResolution = 1e-8
	const epochResolution = time.Duration(864000)
	canonical := make(map[int]Elements)
	for _, f := range readElementFixtures(t) {
		if f.Format == "json" && len(f.Expected) == 1 {
			canonical[f.Expected[0].CatalogNumber] = f.Expected[0]
		}
	}
	maxDiff := make(map[string]float64)
	var maxEpoch time.Duration
	count := 0
	for _, f := range readElementFixtures(t) {
		if f.Format != "tle" {
			continue
		}
		count++
		got, err := ParseTLEText([]byte(f.Raw))
		if err != nil {
			t.Fatal(err)
		}
		e := got[0]
		want, ok := canonical[e.CatalogNumber]
		if !ok {
			t.Fatalf("missing catalog %d", e.CatalogNumber)
		}
		lines := strings.Split(f.Raw, "\n")
		a := lines[1]
		ddExp, err := strconv.Atoi(a[50:52])
		if err != nil {
			t.Fatal(err)
		}
		bsExp, err := strconv.Atoi(a[59:61])
		if err != nil {
			t.Fatal(err)
		}
		gv, wv := reflect.ValueOf(e), reflect.ValueOf(want)
		for i := 0; i < gv.NumField(); i++ {
			name := gv.Type().Field(i).Name
			if gv.Field(i).Kind() == reflect.Float64 {
				var limit float64
				switch name {
				case "Inclination", "RightAscension", "ArgumentOfPerigee", "MeanAnomaly":
					limit = angleResolution
				case "Eccentricity":
					limit = eccentricityResolution
				case "MeanMotion", "MeanMotionDot":
					limit = motionResolution
				case "MeanMotionDDot":
					limit = math.Pow10(ddExp - 5)
				case "BStar":
					limit = math.Pow10(bsExp - 5)
				}
				diff := math.Abs(gv.Field(i).Float() - wv.Field(i).Float())
				if diff > maxDiff[name] {
					maxDiff[name] = diff
				}
				if diff > limit {
					t.Errorf("%s %s difference %.12g exceeds encoded resolution %.12g", f.Label, name, diff, limit)
				}
			} else if name == "Epoch" {
				delta := e.Epoch.Sub(want.Epoch)
				if delta < 0 {
					delta = -delta
				}
				if delta > maxEpoch {
					maxEpoch = delta
				}
				if delta > epochResolution {
					t.Errorf("%s epoch difference %v", f.Label, delta)
				}
			} else if !reflect.DeepEqual(gv.Field(i).Interface(), wv.Field(i).Interface()) {
				t.Errorf("%s %s differs", f.Label, name)
			}
		}
	}
	if count != 6 {
		t.Fatalf("snapshot groups=%d", count)
	}
	t.Logf("six four-format groups; max epoch difference %v; numeric differences %v", maxEpoch, maxDiff)
}

func ommJSONChange(t *testing.T, key string, value interface{}, remove bool) []byte {
	t.Helper()
	var obj map[string]interface{}
	if err := json.Unmarshal(ommSyntheticJSON(t), &obj); err != nil {
		t.Fatal(err)
	}
	if remove {
		delete(obj, key)
	} else {
		obj[key] = value
	}
	raw, err := json.Marshal(obj)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func TestParseOMMInvalidSynthetic(t *testing.T) {
	cases := []struct {
		label string
		data  []byte
	}{
		{"empty", nil}, {"null", []byte("null")}, {"empty-array", []byte("[]")}, {"null-item", []byte("[null]")},
		{"number", []byte("12")}, {"missing", []byte("{}")}, {"broken", []byte("{\"EPOCH\":")},
		{"trailing-json", append(ommSyntheticJSON(t), []byte(" {}")...)},
		{"json-string", []byte("\"text\"")}, {"xml-empty", []byte("<ndm/>")}, {"xml-wrong-root", []byte("<oem/>")},
		{"xml-unclosed", []byte("<ndm><omm>")}, {"xml-bodyless", []byte("<omm/>")},
		{"kvn-missing-header", []byte("EPOCH = 2026-01-01T00:00:00")}, {"kvn-empty", []byte("CCSDS_OMM_VERS = 2.0")},
	}
	for _, tc := range []struct {
		key   string
		value interface{}
	}{
		{"CENTER_NAME", "MARS"}, {"REF_FRAME", "GCRF"}, {"TIME_SYSTEM", "TT"}, {"MEAN_ELEMENT_THEORY", "OTHER"},
		{"CENTER_NAME", ""}, {"REF_FRAME", ""}, {"TIME_SYSTEM", ""}, {"MEAN_ELEMENT_THEORY", ""},
		{"EPOCH", "2026-02-30T00:00:00"}, {"EPOCH", "2026-01-01"}, {"EPOCH", true},
		{"EPOCH", "2026-366T00:00:00"}, {"EPOCH", "2024-000T00:00:00Z"}, {"EPOCH", "2024-01T00:00:00Z"},
		{"NORAD_CAT_ID", 0}, {"NORAD_CAT_ID", -1}, {"NORAD_CAT_ID", 1.5}, {"NORAD_CAT_ID", "9999999999999999999999999999"},
		{"CLASSIFICATION_TYPE", "X"}, {"INCLINATION", 180.1}, {"INCLINATION", -0.1},
		{"RA_OF_ASC_NODE", 360}, {"ARG_OF_PERICENTER", -1}, {"MEAN_ANOMALY", 360},
		{"ECCENTRICITY", 1}, {"ECCENTRICITY", -0.001}, {"MEAN_MOTION", 0},
		{"BSTAR", "NaN"}, {"MEAN_MOTION_DOT", "+Inf"}, {"MEAN_MOTION_DDOT", "-Inf"}, {"INCLINATION", nil},
		{"ELEMENT_SET_NO", -1}, {"REV_AT_EPOCH", -1}, {"EPHEMERIS_TYPE", -1}, {"EPHEMERIS_TYPE", 4}, {"EPHEMERIS_TYPE", 6},
	} {
		cases = append(cases, struct {
			label string
			data  []byte
		}{tc.key, ommJSONChange(t, tc.key, tc.value, false)})
	}
	for _, key := range []string{"EPOCH", "NORAD_CAT_ID", "CLASSIFICATION_TYPE", "INCLINATION", "RA_OF_ASC_NODE", "ARG_OF_PERICENTER", "MEAN_ANOMALY", "ECCENTRICITY", "MEAN_MOTION", "MEAN_MOTION_DOT", "MEAN_MOTION_DDOT", "BSTAR", "ELEMENT_SET_NO", "REV_AT_EPOCH", "EPHEMERIS_TYPE"} {
		cases = append(cases, struct {
			label string
			data  []byte
		}{"missing-" + key, ommJSONChange(t, key, nil, true)})
	}
	for _, f := range readElementFixtures(t) {
		if f.Label == "iss-xml" {
			for _, tc := range []struct{ before, after string }{{">EARTH<", ">MARS<"}, {">TEME<", ">GCRF<"}, {">UTC<", ">TT<"}, {">SGP4<", ">OTHER<"}} {
				cases = append(cases, struct {
					label string
					data  []byte
				}{"xml-metadata", []byte(strings.Replace(f.Raw, tc.before, tc.after, 1))})
			}
			cases = append(cases, struct {
				label string
				data  []byte
			}{"xml-multiple-roots", []byte(f.Raw + f.Raw)})
			cases = append(cases, struct {
				label string
				data  []byte
			}{"xml-missing-theory", []byte(strings.Replace(f.Raw, "<MEAN_ELEMENT_THEORY>SGP4</MEAN_ELEMENT_THEORY>", "", 1))})
		}
		if f.Label == "iss-kvn" {
			cases = append(cases, struct {
				label string
				data  []byte
			}{"kvn-missing-theory", []byte(strings.Replace(f.Raw, "MEAN_ELEMENT_THEORY = SGP/SGP4\r\n", "", 1))})
			for _, tc := range []struct{ before, after string }{{"EARTH", "MARS"}, {"TEME", "GCRF"}, {"UTC", "TT"}, {"SGP/SGP4", "OTHER"}} {
				cases = append(cases, struct {
					label string
					data  []byte
				}{"kvn-metadata", []byte(strings.Replace(f.Raw, tc.before, tc.after, 1))})
			}
			for _, tail := range []string{"\ninvalid line", "\nMEAN_MOTION = 1", "\nBSTAR = 1 [km"} {
				cases = append(cases, struct {
					label string
					data  []byte
				}{"kvn-malformed", []byte(f.Raw + tail)})
			}
		}
	}
	for _, tc := range cases {
		t.Run(tc.label, func(t *testing.T) {
			got, err := ParseOMM(tc.data)
			if !errors.Is(err, ErrInvalidElements) || got != nil {
				t.Fatalf("got %+v, %v", got, err)
			}
		})
	}
}

func TestParseOMMBracketedNameSynthetic(t *testing.T) {
	for _, f := range readElementFixtures(t) {
		if f.Label != "iss-kvn" {
			continue
		}
		raw := strings.Replace(f.Raw, "ISS (ZARYA)", "ISS [Segment 01]", 1)
		got, err := ParseOMMKVN([]byte(raw))
		if err != nil {
			t.Fatal(err)
		}
		if len(got) != 1 || got[0].Name != "ISS [Segment 01]" {
			t.Fatalf("name lost: %+v", got)
		}
	}
}

func TestParseOMMBOMSynthetic(t *testing.T) {
	for _, f := range readElementFixtures(t) {
		var parse func([]byte) ([]Elements, error)
		switch f.Label {
		case "iss-json":
			parse = ParseOMMJSON
		case "iss-xml":
			parse = ParseOMMXML
		case "iss-kvn":
			parse = ParseOMMKVN
		default:
			continue
		}
		t.Run(f.Format, func(t *testing.T) {
			want, err := parse([]byte(f.Raw))
			if err != nil {
				t.Fatal(err)
			}
			for _, fn := range []func([]byte) ([]Elements, error){parse, ParseOMM} {
				got, err := fn(append([]byte("\ufeff"), []byte(f.Raw)...))
				if err != nil {
					t.Fatal(err)
				}
				if !reflect.DeepEqual(got, want) {
					t.Fatalf("BOM changed elements: %+v", got)
				}
			}
		})
	}
}
