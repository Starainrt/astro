package svg

import (
	"strings"
	"testing"
	"time"
)

// EventsTitle 覆盖“全球阶段”数据块的标题；空值使用该块的本地化默认标题。
func TestSolarEclipseEventsTitleLocalizesAndOverrides(t *testing.T) {
	for _, test := range []struct {
		name     string
		language string
		title    string
		want     string
	}{
		{name: "zh default", want: "食甚点的地方情况"},
		{name: "zh override", title: "全球阶段（自定义）", want: "全球阶段（自定义）"},
		{name: "en default", language: "en", want: "Local circumstances at greatest"},
		{name: "en override", language: "en", title: "Global phases (custom)", want: "Global phases (custom)"},
	} {
		t.Run(test.name, func(t *testing.T) {
			_, circumstances, _ := solarEclipseContactTitles(SolarEclipseMapSVGOptions{
				Language: test.language, EventsTitle: test.title,
			})
			if circumstances != test.want {
				t.Fatalf("global-phase block title = %q, want %q", circumstances, test.want)
			}
		})
	}

	// 两种版式都要把标题写进产物，横版与竖版各验证一次。
	for _, layout := range []struct {
		name          string
		width, height int
		language      string
	}{
		{name: "landscape", width: 960, height: 640},
		{name: "portrait", width: 800, height: 820, language: "en"},
	} {
		t.Run("rendered/"+layout.name, func(t *testing.T) {
			title := "全球阶段（自定义）"
			unwanted := "食甚点的地方情况"
			if layout.language == "en" {
				title = "Global phases (custom)"
				unwanted = "Local circumstances at greatest"
			}
			rendered, ok := SolarEclipseMapSVG(
				time.Date(2024, time.April, 8, 0, 0, 0, 0, time.UTC),
				SolarEclipseMapSVGOptions{
					Width: layout.width, Height: layout.height, Location: time.UTC,
					PartialStep: 10 * time.Minute, Language: layout.language, EventsTitle: title,
				},
			)
			if !ok {
				t.Fatal("missing 2024-04-08 map")
			}
			if !strings.Contains(rendered, title) {
				t.Fatalf("map is missing the events title %q", title)
			}
			if strings.Contains(rendered, unwanted) {
				t.Fatalf("map still carries the replaced events title %q", unwanted)
			}
			if err := validateEclipseMapXML(rendered); err != nil {
				t.Fatalf("map is not valid XML: %v", err)
			}
		})
	}
}
