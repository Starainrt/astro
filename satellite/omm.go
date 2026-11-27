package satellite

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
)

// ParseOMM 自动识别并解析轨道平均根数 / detect and parse orbit mean elements.
func ParseOMM(data []byte) ([]Elements, error) {
	data = bytes.TrimPrefix(data, []byte("\xef\xbb\xbf"))
	data = bytes.TrimSpace(data)
	if len(data) == 0 {
		return nil, fmt.Errorf("%w: empty message", ErrInvalidElements)
	}
	switch data[0] {
	case '[', '{':
		return ParseOMMJSON(data)
	case '<':
		return ParseOMMXML(data)
	default:
		return ParseOMMKVN(data)
	}
}

// ParseOMMJSON 解析JSON轨道平均根数 / parse JSON orbit mean elements.
func ParseOMMJSON(data []byte) ([]Elements, error) {
	data = bytes.TrimPrefix(data, []byte("\xef\xbb\xbf"))
	data = bytes.TrimSpace(data)
	decoder := json.NewDecoder(bytes.NewReader(data))
	array := len(data) > 0 && data[0] == '['
	if array {
		if _, err := decoder.Token(); err != nil {
			return nil, fmt.Errorf("%w: JSON message", ErrInvalidElements)
		}
	}
	var out []Elements
	for !array || decoder.More() {
		fields, err := ommJSONFields(decoder)
		if err != nil {
			return nil, err
		}
		e, err := ommElements(fields, false)
		if err != nil {
			return nil, err
		}
		out = append(out, e)
		if !array {
			break
		}
	}
	if array {
		if _, err := decoder.Token(); err != nil {
			return nil, fmt.Errorf("%w: JSON array", ErrInvalidElements)
		}
	}
	if _, err := decoder.Token(); err != io.EOF {
		return nil, fmt.Errorf("%w: trailing JSON data", ErrInvalidElements)
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("%w: no element records", ErrInvalidElements)
	}
	return out, nil
}

func ommJSONFields(decoder *json.Decoder) (map[string]string, error) {
	if token, err := decoder.Token(); err != nil || token != json.Delim('{') {
		return nil, fmt.Errorf("%w: JSON record", ErrInvalidElements)
	}
	fields := make(map[string]string)
	for decoder.More() {
		token, err := decoder.Token()
		key, isKey := token.(string)
		if err != nil || !isKey {
			return nil, fmt.Errorf("%w: JSON field", ErrInvalidElements)
		}
		if _, exists := fields[key]; exists {
			return nil, fmt.Errorf("%w: duplicate JSON field %s", ErrInvalidElements, key)
		}
		var value json.RawMessage
		if err := decoder.Decode(&value); err != nil {
			return nil, fmt.Errorf("%w: %s", ErrInvalidElements, key)
		}
		if !ommFieldKnown(key) {
			continue
		}
		var text string
		if err := json.Unmarshal(value, &text); err != nil {
			var number json.Number
			if err := json.Unmarshal(value, &number); err != nil {
				return nil, fmt.Errorf("%w: %s", ErrInvalidElements, key)
			}
			text = number.String()
		}
		fields[key] = strings.TrimSpace(text)
	}
	if _, err := decoder.Token(); err != nil {
		return nil, fmt.Errorf("%w: JSON record", ErrInvalidElements)
	}
	return fields, nil
}

func ommFieldKnown(key string) bool {
	return ommFieldSection(key) != ""
}

func ommFieldSection(key string) string {
	switch key {
	case "OBJECT_NAME", "OBJECT_ID", "CENTER_NAME", "REF_FRAME", "TIME_SYSTEM", "MEAN_ELEMENT_THEORY":
		return "metadata"
	case "EPOCH", "MEAN_MOTION", "ECCENTRICITY", "INCLINATION", "RA_OF_ASC_NODE", "ARG_OF_PERICENTER", "MEAN_ANOMALY":
		return "meanElements"
	case "CLASSIFICATION_TYPE", "NORAD_CAT_ID", "MEAN_MOTION_DOT", "MEAN_MOTION_DDOT", "BSTAR", "ELEMENT_SET_NO", "REV_AT_EPOCH", "EPHEMERIS_TYPE":
		return "tleParameters"
	}
	return ""
}

func validOMMUnit(key, unit string) bool {
	switch key {
	case "INCLINATION", "RA_OF_ASC_NODE", "ARG_OF_PERICENTER", "MEAN_ANOMALY":
		return unit == "deg"
	case "MEAN_MOTION":
		return unit == "rev/day"
	case "MEAN_MOTION_DOT":
		return unit == "rev/day**2"
	case "MEAN_MOTION_DDOT":
		return unit == "rev/day**3"
	case "BSTAR":
		return unit == "1/ER"
	}
	return false
}

func ommElements(f map[string]string, standardFormat bool) (Elements, error) {
	for _, m := range []struct{ key, want string }{{"CENTER_NAME", "EARTH"}, {"REF_FRAME", "TEME"}, {"TIME_SYSTEM", "UTC"}} {
		v, present := f[m.key]
		if (standardFormat || present) && v != m.want {
			return Elements{}, fmt.Errorf("%w: %s", ErrInvalidElements, m.key)
		}
	}
	if theory, present := f["MEAN_ELEMENT_THEORY"]; (standardFormat || present) && theory != "SGP4" && theory != "SGP/SGP4" {
		return Elements{}, fmt.Errorf("%w: mean element theory", ErrInvalidElements)
	}
	epoch, err := time.Parse(time.RFC3339Nano, f["EPOCH"])
	if err != nil {
		epoch, err = time.Parse("2006-01-02T15:04:05", f["EPOCH"])
	}
	if err != nil {
		epoch, err = time.Parse("2006-002T15:04:05", strings.TrimSuffix(f["EPOCH"], "Z"))
	}
	if err != nil {
		return Elements{}, fmt.Errorf("%w: epoch", ErrInvalidElements)
	}
	e := Elements{Name: f["OBJECT_NAME"], InternationalDesignator: f["OBJECT_ID"], Classification: f["CLASSIFICATION_TYPE"], Epoch: epoch.UTC()}
	if _, present := f["CLASSIFICATION_TYPE"]; standardFormat && !present {
		e.Classification = "U"
	}
	for _, v := range []struct {
		key string
		dst *int
	}{{"NORAD_CAT_ID", &e.CatalogNumber}, {"ELEMENT_SET_NO", &e.ElementSetNumber}, {"REV_AT_EPOCH", &e.RevolutionNumber}, {"EPHEMERIS_TYPE", &e.EphemerisType}} {
		value, present := f[v.key]
		if standardFormat && !present && v.key != "NORAD_CAT_ID" {
			continue
		}
		n, err := strconv.Atoi(value)
		if err != nil {
			return Elements{}, fmt.Errorf("%w: %s", ErrInvalidElements, v.key)
		}
		*v.dst = n
	}
	for _, v := range []struct {
		key string
		dst *float64
	}{{"MEAN_MOTION", &e.MeanMotion}, {"ECCENTRICITY", &e.Eccentricity}, {"INCLINATION", &e.Inclination}, {"RA_OF_ASC_NODE", &e.RightAscension}, {"ARG_OF_PERICENTER", &e.ArgumentOfPerigee}, {"MEAN_ANOMALY", &e.MeanAnomaly}, {"MEAN_MOTION_DOT", &e.MeanMotionDot}, {"MEAN_MOTION_DDOT", &e.MeanMotionDDot}, {"BSTAR", &e.BStar}} {
		value, present := f[v.key]
		if standardFormat && !present && (v.key == "MEAN_MOTION_DOT" || v.key == "MEAN_MOTION_DDOT") {
			continue
		}
		n, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return Elements{}, fmt.Errorf("%w: %s", ErrInvalidElements, v.key)
		}
		*v.dst = n
	}
	if !validElements(e) {
		return Elements{}, fmt.Errorf("%w: field range", ErrInvalidElements)
	}
	return e, nil
}

// ParseOMMXML 解析XML轨道平均根数 / parse XML orbit mean elements.
func ParseOMMXML(data []byte) ([]Elements, error) {
	data = bytes.TrimPrefix(data, []byte("\xef\xbb\xbf"))
	d := xml.NewDecoder(bytes.NewReader(data))
	var out []Elements
	var root string
	closed := false
	for {
		token, err := d.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("%w: XML message", ErrInvalidElements)
		}
		switch v := token.(type) {
		case xml.StartElement:
			if closed {
				return nil, fmt.Errorf("%w: multiple XML roots", ErrInvalidElements)
			}
			if root == "" {
				root = v.Name.Local
				if root != "omm" && root != "ndm" {
					return nil, fmt.Errorf("%w: XML root", ErrInvalidElements)
				}
				if root == "ndm" {
					continue
				}
			} else if v.Name.Local == "COMMENT" {
				if err := d.Skip(); err != nil {
					return nil, fmt.Errorf("%w: XML comment", ErrInvalidElements)
				}
				continue
			} else if v.Name.Local != "omm" {
				return nil, fmt.Errorf("%w: XML record", ErrInvalidElements)
			}
			f, err := ommXMLFields(d, v)
			if err != nil {
				return nil, err
			}
			e, err := ommElements(f, true)
			if err != nil {
				return nil, err
			}
			out = append(out, e)
			if root == "omm" {
				closed = true
			}
		case xml.EndElement:
			if root != "ndm" || v.Name.Local != "ndm" {
				return nil, fmt.Errorf("%w: XML closing tag", ErrInvalidElements)
			}
			closed = true
		case xml.CharData:
			if len(bytes.TrimSpace(v)) != 0 {
				return nil, fmt.Errorf("%w: XML text outside record", ErrInvalidElements)
			}
		}
	}
	if !closed || len(out) == 0 {
		return nil, fmt.Errorf("%w: no complete XML records", ErrInvalidElements)
	}
	return out, nil
}

func ommXMLFields(d *xml.Decoder, start xml.StartElement) (map[string]string, error) {
	fields := make(map[string]string)
	path := []string{start.Name.Local}
	segmentSeen := false
	for len(path) > 0 {
		token, err := d.Token()
		if err != nil {
			return nil, fmt.Errorf("%w: XML record", ErrInvalidElements)
		}
		switch v := token.(type) {
		case xml.StartElement:
			if len(path) == 2 && path[1] == "body" && v.Name.Local == "segment" {
				if segmentSeen {
					return nil, fmt.Errorf("%w: multiple XML segments", ErrInvalidElements)
				}
				segmentSeen = true
			}
			if section := ommFieldSection(v.Name.Local); section != "" {
				if !ommXMLFieldInSection(path, section) {
					return nil, fmt.Errorf("%w: %s outside XML %s", ErrInvalidElements, v.Name.Local, section)
				}
				if _, ok := fields[v.Name.Local]; ok {
					return nil, fmt.Errorf("%w: duplicate XML field", ErrInvalidElements)
				}
				for _, attr := range v.Attr {
					if attr.Name.Local == "units" && !validOMMUnit(v.Name.Local, attr.Value) {
						return nil, fmt.Errorf("%w: %s units", ErrInvalidElements, v.Name.Local)
					}
				}
				var value string
				if err = d.DecodeElement(&value, &v); err != nil {
					return nil, fmt.Errorf("%w: XML field", ErrInvalidElements)
				}
				fields[v.Name.Local] = strings.TrimSpace(value)
			} else {
				path = append(path, v.Name.Local)
			}
		case xml.EndElement:
			path = path[:len(path)-1]
		}
	}
	return fields, nil
}

func ommXMLFieldInSection(path []string, section string) bool {
	if len(path) < 4 || path[1] != "body" || path[2] != "segment" {
		return false
	}
	if section == "metadata" {
		return len(path) == 4 && path[3] == section
	}
	return len(path) == 5 && path[3] == "data" && path[4] == section
}

// ParseOMMKVN 解析KVN轨道平均根数 / parse KVN orbit mean elements.
func ParseOMMKVN(data []byte) ([]Elements, error) {
	data = bytes.TrimPrefix(data, []byte("\xef\xbb\xbf"))
	var out []Elements
	var fields map[string]string
	for _, raw := range strings.Split(string(data), "\n") {
		line := strings.TrimSpace(raw)
		if line == "" || line == "COMMENT" || strings.HasPrefix(line, "COMMENT ") || strings.HasPrefix(line, "COMMENT\t") {
			continue
		}
		if fields != nil && (line == "META_START" || line == "META_STOP") {
			continue
		}
		i := strings.IndexByte(line, '=')
		if i < 1 {
			return nil, fmt.Errorf("%w: KVN field", ErrInvalidElements)
		}
		key, value := strings.TrimSpace(line[:i]), strings.TrimSpace(line[i+1:])
		if key == "CCSDS_OMM_VERS" {
			if value == "" {
				return nil, fmt.Errorf("%w: KVN version", ErrInvalidElements)
			}
			if fields != nil {
				e, err := ommElements(fields, true)
				if err != nil {
					return nil, err
				}
				out = append(out, e)
			}
			fields = make(map[string]string)
			continue
		}
		if fields == nil {
			return nil, fmt.Errorf("%w: missing KVN header", ErrInvalidElements)
		}
		if !ommFieldKnown(key) {
			continue
		}
		if _, ok := fields[key]; ok {
			return nil, fmt.Errorf("%w: duplicate KVN field", ErrInvalidElements)
		}
		if i := strings.LastIndexByte(value, '['); i >= 0 && key != "OBJECT_NAME" && key != "OBJECT_ID" {
			number := strings.TrimRight(value[:i], " \t")
			if len(number) == i || !strings.HasSuffix(value, "]") || !validOMMUnit(key, value[i+1:len(value)-1]) {
				return nil, fmt.Errorf("%w: %s units", ErrInvalidElements, key)
			}
			value = number
		}
		fields[key] = value
	}
	if fields == nil {
		return nil, fmt.Errorf("%w: no KVN records", ErrInvalidElements)
	}
	e, err := ommElements(fields, true)
	if err != nil {
		return nil, err
	}
	return append(out, e), nil
}
