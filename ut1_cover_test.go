package astro_test

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/starainrt/astro"
	"github.com/starainrt/astro/basic"
	"github.com/starainrt/astro/eclipse"
	"github.com/starainrt/astro/moon"
)

// UT1 转换器的字段级契约（反射穷举，而不是逐字段手写）：
//  1. 结果里每一个非零 time.Time 都必须等于该时刻的 UT1 读数，零值原样保留；
//  2. 含时刻的切片必须整体复制，不能与输入共享底层数组（改输出会改到调用方原数据）。
//
// Field-level contract of the UT1 converters, enumerated by reflection rather than by hand:
// every non-zero time must move to its UT1 reading, zero instants stay zero, and every
// time-bearing slice must be deep-copied.
func TestUT1ConvertersCoverEveryTimeField(t *testing.T) {
	cases := []struct {
		name    string
		input   interface{}
		convert func(interface{}) interface{}
	}{
		{"SolarEclipseInfo", synthetic(eclipse.SolarEclipseInfo{}), func(v interface{}) interface{} {
			return eclipse.SolarEclipseInfoInUT1(v.(eclipse.SolarEclipseInfo))
		}},
		{"SolarEclipsePartialFootprintsInfo", synthetic(eclipse.SolarEclipsePartialFootprintsInfo{}), func(v interface{}) interface{} {
			return eclipse.SolarEclipsePartialFootprintsInUT1(v.(eclipse.SolarEclipsePartialFootprintsInfo))
		}},
		{"SolarEclipsePath", synthetic(eclipse.SolarEclipsePath{}), func(v interface{}) interface{} {
			return eclipse.SolarEclipsePathInUT1(v.(eclipse.SolarEclipsePath))
		}},
		{"LunarEclipseInfo", synthetic(eclipse.LunarEclipseInfo{}), func(v interface{}) interface{} {
			return eclipse.LunarEclipseInfoInUT1(v.(eclipse.LunarEclipseInfo))
		}},
		{"LocalSolarEclipseInfo", synthetic(eclipse.LocalSolarEclipseInfo{}), func(v interface{}) interface{} {
			return eclipse.LocalSolarEclipseInfoInUT1(v.(eclipse.LocalSolarEclipseInfo))
		}},
		{"SolarEclipseGeocentricPanel", synthetic(eclipse.SolarEclipseGeocentricPanel{}), func(v interface{}) interface{} {
			return eclipse.SolarEclipseGeocentricPanelInUT1(v.(eclipse.SolarEclipseGeocentricPanel))
		}},
		{"StarOccultationInfo", synthetic(basic.StarOccultationInfo{}), func(v interface{}) interface{} {
			return moon.StarOccultationInfoInUT1(v.(basic.StarOccultationInfo))
		}},
		{"StarOccultationPath", synthetic(basic.StarOccultationPath{}), func(v interface{}) interface{} {
			return moon.StarOccultationPathInUT1(v.(basic.StarOccultationPath))
		}},
		{"PlanetOccultationInfo", synthetic(basic.PlanetOccultationInfo{}), func(v interface{}) interface{} {
			return moon.PlanetOccultationInfoInUT1(v.(basic.PlanetOccultationInfo))
		}},
		{"PlanetOccultationPath", synthetic(basic.PlanetOccultationPath{}), func(v interface{}) interface{} {
			return moon.PlanetOccultationPathInUT1(v.(basic.PlanetOccultationPath))
		}},
	}
	for _, testCase := range cases {
		output := testCase.convert(testCase.input)
		problems := checkUT1Coverage(reflect.ValueOf(testCase.input), reflect.ValueOf(output), "")
		timeFields := countTimeFields(reflect.TypeOf(testCase.input), 0)
		if timeFields == 0 {
			t.Errorf("%s: synthetic value carries no time field, test would be vacuous", testCase.name)
		}
		for _, problem := range problems {
			t.Errorf("%s: %s", testCase.name, problem)
		}
	}
}

var timeType = reflect.TypeOf(time.Time{})

// synthetic 递归填充每个字段，time.Time 取固定非零时刻，切片给两个元素。
func synthetic(prototype interface{}) interface{} {
	value := reflect.New(reflect.TypeOf(prototype))
	fillSynthetic(value.Elem(), 0)
	return value.Elem().Interface()
}

func fillSynthetic(value reflect.Value, depth int) {
	if depth > 8 || !value.CanSet() {
		return
	}
	if value.Type() == timeType {
		value.Set(reflect.ValueOf(time.Date(2026, 3, 3, 11, 0, 0, 0, time.UTC)))
		return
	}
	switch value.Kind() {
	case reflect.Struct:
		for index := 0; index < value.NumField(); index++ {
			fillSynthetic(value.Field(index), depth+1)
		}
	case reflect.Slice:
		value.Set(reflect.MakeSlice(value.Type(), 2, 2))
		for index := 0; index < value.Len(); index++ {
			fillSynthetic(value.Index(index), depth+1)
		}
	case reflect.Ptr:
		value.Set(reflect.New(value.Type().Elem()))
		fillSynthetic(value.Elem(), depth+1)
	case reflect.Float32, reflect.Float64:
		value.SetFloat(1)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		value.SetInt(1)
	case reflect.String:
		value.SetString("synthetic")
	case reflect.Bool:
		value.SetBool(true)
	}
}

// checkUT1Coverage 并行遍历输入与输出，检查时刻换算与切片别名。
func checkUT1Coverage(input, output reflect.Value, path string) []string {
	if input.Type() != output.Type() {
		return []string{path + ": type mismatch"}
	}
	if input.Type() == timeType {
		at := input.Interface().(time.Time)
		got := output.Interface().(time.Time)
		if at.IsZero() {
			if !got.IsZero() {
				return []string{path + ": zero instant became " + got.String()}
			}
			return nil
		}
		want := astro.LabelIn(astro.TimeScaleUT1, at)
		if !got.Equal(want) {
			return []string{fmt.Sprintf("%s: got %v, want UT1 %v", path, got, want)}
		}
		return nil
	}
	switch input.Kind() {
	case reflect.Struct:
		var problems []string
		for index := 0; index < input.NumField(); index++ {
			problems = append(problems, checkUT1Coverage(
				input.Field(index), output.Field(index),
				path+"."+input.Type().Field(index).Name,
			)...)
		}
		return problems
	case reflect.Slice:
		var problems []string
		if containsTime(input.Type().Elem(), 0) && input.Len() > 0 && output.Len() > 0 &&
			input.Pointer() == output.Pointer() {
			problems = append(problems, path+": output shares the input slice (aliasing)")
		}
		for index := 0; index < input.Len() && index < output.Len(); index++ {
			problems = append(problems, checkUT1Coverage(
				input.Index(index), output.Index(index), fmt.Sprintf("%s[%d]", path, index))...)
		}
		return problems
	case reflect.Ptr:
		if input.IsNil() || output.IsNil() {
			return nil
		}
		return checkUT1Coverage(input.Elem(), output.Elem(), path)
	}
	return nil
}

func containsTime(value reflect.Type, depth int) bool {
	if value == timeType {
		return true
	}
	if depth > 8 {
		return false
	}
	switch value.Kind() {
	case reflect.Slice, reflect.Array, reflect.Ptr:
		return containsTime(value.Elem(), depth+1)
	case reflect.Struct:
		for index := 0; index < value.NumField(); index++ {
			if containsTime(value.Field(index).Type, depth+1) {
				return true
			}
		}
	}
	return false
}

func countTimeFields(value reflect.Type, depth int) int {
	if value == timeType {
		return 1
	}
	if depth > 8 {
		return 0
	}
	count := 0
	switch value.Kind() {
	case reflect.Slice, reflect.Array, reflect.Ptr:
		count += countTimeFields(value.Elem(), depth+1)
	case reflect.Struct:
		for index := 0; index < value.NumField(); index++ {
			count += countTimeFields(value.Field(index).Type, depth+1)
		}
	}
	return count
}

var _ = strings.TrimSpace
