package svg

import (
	"encoding/xml"
	"fmt"
	"html"
	"io"
	"math"
	"strings"
	"time"

	"github.com/starainrt/astro/internal/svgchart"
)

const (
	canvasWidth  = 900.0
	canvasHeight = 600.0
	serifFont    = `Georgia, 'Times New Roman', serif`
)

type chart struct {
	strings.Builder
	options   chartOptions
	valid     bool
	labels    svgchart.LabelTable
	titleText string
}

func newChart(options chartOptions, title string, defs string) *chart {
	drawing := &chart{options: options, valid: true, titleText: svgchart.EllipsizeText(title, 820, 26)}
	fmt.Fprintf(drawing, `<svg xmlns="http://www.w3.org/2000/svg" width="%.0f" height="%.0f" viewBox="0 0 900 600" role="img" aria-label="%s">`, options.width, options.height, html.EscapeString(title))
	drawing.WriteString(`<defs>` + defs + `</defs>`)
	drawing.WriteString(`<rect width="100%" height="100%" fill="#efefed"/>`)
	drawing.WriteString(`<rect x="22" y="18" width="856" height="564" fill="#ffffff" stroke="#c9c9c6" stroke-width="1.2"/>`)
	return drawing
}

func (drawing *chart) text(x, baseline, size float64, anchor, fill, weight, value string) {
	box := svgchart.LabelTextBox(x, baseline, size, value, anchor)
	if box.X < 0 || box.Y < 0 || box.X+box.Width > canvasWidth || box.Y+box.Height > canvasHeight {
		drawing.valid = false
	}
	weightAttribute := ""
	if weight != "" {
		weightAttribute = ` font-weight="` + weight + `"`
	}
	fmt.Fprintf(drawing, `<text x="%.3f" y="%.3f" fill="%s" font-family="%s" font-size="%.0f"%s text-anchor="%s">%s</text>`,
		x, baseline, fill, serifFont, size, weightAttribute, anchor, html.EscapeString(value))
}

func (drawing *chart) placeText(size float64, fill, value string, candidates []svgchart.LabelPlacement) {
	placement, ok := drawing.labels.Place(value, size, candidates)
	if !ok {
		placement = candidates[0]
	}
	drawing.text(placement.X, placement.Y, size, placement.Anchor, fill, "", value)
}

func (drawing *chart) header(title string, lines []string) float64 {
	drawing.text(450, 46, 26, "middle", "#111111", "700", svgchart.EllipsizeText(title, 820, 26))
	drawing.WriteString(`<line x1="372" y1="59" x2="528" y2="59" stroke="#555" stroke-width="1"/>`)
	baseline := 84.0
	for index, line := range lines {
		size, fill := 13.0, "#333"
		if index == 0 {
			size, fill = 14, "#222"
		}
		drawing.text(450, baseline, size, "middle", fill, "", svgchart.EllipsizeText(line, 820, size))
		drawing.labels.ReserveText(450, baseline, size, line, "middle")
		baseline += 21
	}
	return baseline - 21
}

func (drawing *chart) warning(baseline float64, value string) {
	drawing.text(450, baseline, 12, "middle", "#b21f16", "", svgchart.EllipsizeText(value, 820, 12))
}

func (drawing *chart) footer(notes []string) {
	drawing.footerWithin(notes, 820)
}

func (drawing *chart) footerWithin(notes []string, width float64) {
	var lines []string
	for _, note := range notes {
		lines = append(lines, svgchart.WrapText(note, width, 12)...)
	}
	lines = svgchart.TruncateTextLines(lines, width, 12, 3)
	first := 572 - 17*float64(len(lines)-1)
	for index, line := range lines {
		drawing.text(38, first+17*float64(index), 12, "start", "#444", "", line)
	}
}

func (drawing *chart) finish() (string, bool) {
	drawing.WriteString(`</svg>`)
	result := drawing.String()
	if !drawing.valid || !finiteChartXML(result, drawing.titleText) {
		return "", false
	}
	return result, true
}

func finiteChartXML(document, titleText string) bool {
	decoder := xml.NewDecoder(strings.NewReader(document))
	var text strings.Builder
	textDepth := 0
	titleAllowed := true
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			return textDepth == 0
		}
		if err != nil {
			return false
		}
		switch value := token.(type) {
		case xml.StartElement:
			if value.Name.Local == "text" {
				text.Reset()
				textDepth = 1
			}
			for _, attribute := range value.Attr {
				if attribute.Name.Local != "aria-label" && containsNonFiniteNumber(attribute.Value) {
					return false
				}
			}
		case xml.CharData:
			if textDepth > 0 {
				text.Write(value)
			}
		case xml.EndElement:
			if value.Name.Local == "text" && textDepth > 0 {
				content := text.String()
				if containsNonFiniteNumber(content) {
					if !titleAllowed || content != titleText {
						return false
					}
					titleAllowed = false
				}
				textDepth = 0
			}
		}
	}
}

func containsNonFiniteNumber(value string) bool {
	for _, token := range []string{"NaN", "Inf"} {
		for offset := 0; offset < len(value); {
			index := strings.Index(value[offset:], token)
			if index < 0 {
				break
			}
			index += offset
			end := index + len(token)
			if (index == 0 || !isASCIIWord(value[index-1])) && (end == len(value) || !isASCIIWord(value[end])) {
				return true
			}
			offset = end
		}
	}
	return false
}

func isASCIIWord(value byte) bool {
	return value >= 'a' && value <= 'z' || value >= 'A' && value <= 'Z' || value >= '0' && value <= '9' || value == '_'
}

func (drawing *chart) words(chinese, english string) string {
	if drawing.options.english {
		return english
	}
	return chinese
}

func (drawing *chart) clock(date time.Time) string {
	return date.In(drawing.options.location).Format("15:04:05")
}

func (drawing *chart) stamp(date time.Time) string {
	return date.In(drawing.options.location).Format("2006-01-02 15:04:05") + " " + zoneLabel(date.In(drawing.options.location))
}

// zoneLabel 按该时刻的实际偏移给出 UTC±h[:mm]，不依赖时区名，夏令时也跟随时刻。
func zoneLabel(date time.Time) string {
	_, offset := date.Zone()
	sign := "+"
	if offset < 0 {
		sign, offset = "-", -offset
	}
	if offset%3600 == 0 {
		return fmt.Sprintf("UTC%s%d", sign, offset/3600)
	}
	return fmt.Sprintf("UTC%s%d:%02d", sign, offset/3600, offset/60%60)
}

func (drawing *chart) timeNote(date time.Time) string {
	return fmt.Sprintf(drawing.words("地球接收时刻（%s）。", "Earth-received times, %s."), zoneLabel(date.In(drawing.options.location)))
}

func formatDuration(duration time.Duration) string {
	seconds := int(math.Round(duration.Seconds()))
	return fmt.Sprintf("%02d:%02d:%02d", seconds/3600, seconds/60%60, seconds%60)
}

func satelliteName(index int, english bool) string {
	if english {
		return [...]string{"Io", "Europa", "Ganymede", "Callisto"}[index]
	}
	return [...]string{"木卫一", "木卫二", "木卫三", "木卫四"}[index]
}

func satelliteColor(index int) string {
	return [...]string{"#e8c46a", "#d8ccb2", "#b9ad98", "#8f8172"}[index]
}

func (drawing *chart) compass(cx, cy, north, south, east, west float64) {
	drawing.text(cx, north, 14, "middle", "#111", "700", drawing.words("北", "N"))
	drawing.text(cx, south, 14, "middle", "#111", "700", drawing.words("南", "S"))
	drawing.text(east, cy+5, 14, "middle", "#111", "700", drawing.words("东", "E"))
	drawing.text(west, cy+5, 14, "middle", "#111", "700", drawing.words("西", "W"))
	for _, box := range [][4]float64{{cx, north}, {cx, south}, {east, cy + 5}, {west, cy + 5}} {
		drawing.labels.ReserveText(box[0], box[1], 14, "北", "middle")
	}
}
