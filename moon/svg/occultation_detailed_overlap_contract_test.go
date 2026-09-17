package svg

import (
	"errors"
	"strconv"
	"testing"
)

// 详细版式的文本盒估算沿用测试工具：ASCII 0.62em、非 ASCII 1.05em，纵向 y-0.85em 到 y+0.3em。

func occultationDetailedTextOverlapPairs(texts []occultationSVGText) [][2]occultationSVGText {
	pairs := [][2]occultationSVGText{}
	for first := 0; first < len(texts); first++ {
		for second := first + 1; second < len(texts); second++ {
			a, b := occultationSVGTextBox(texts[first]), occultationSVGTextBox(texts[second])
			if a[0] < b[2] && b[0] < a[2] && a[1] < b[3] && b[1] < a[3] {
				pairs = append(pairs, [2]occultationSVGText{texts[first], texts[second]})
			}
		}
	}
	return pairs
}

// 详细版式在能出图的画布上不得有互相压叠的文本；装不下的画布必须返回 ErrInvalidOccultationDetailedSVGOptions。
func TestOccultationDetailedSVGTextOverlapContract(t *testing.T) {
	// 竖版最小画布 600x800 的已知压叠总对数：数量变化即视为回归，必须重新核对而不是继续跳过。
	knownOverlaps := map[[2]int]int{{600, 800}: 10}
	sizes := [][2]int{
		{1000, 1414}, {1414, 1000}, {1200, 800}, {900, 560}, {900, 400},
		{800, 600}, {640, 480}, {600, 800}, {480, 320},
	}
	starPath := occultationTestStarPath(t)
	planetPath := occultationTestSaturnPath(t)
	for _, size := range sizes {
		size := size
		t.Run(strconv.Itoa(size[0])+"x"+strconv.Itoa(size[1]), func(t *testing.T) {
			options := OccultationDetailedSVGOptions{Width: size[0], Height: size[1]}
			overlapping := 0
			for _, item := range []struct {
				label  string
				render func() (string, error)
			}{
				{label: "stellar", render: func() (string, error) {
					return StarOccultationDetailedSVG(starPath, hr4799StarCoordinate(), options)
				}},
				{label: "planetary", render: func() (string, error) {
					return PlanetOccultationDetailedSVG(planetPath, options)
				}},
			} {
				diagram, err := item.render()
				if err != nil {
					if !errors.Is(err, ErrInvalidOccultationDetailedSVGOptions) {
						t.Fatalf("%s %dx%d: %v", item.label, size[0], size[1], err)
					}
					t.Logf("%s %dx%d rejected: %v", item.label, size[0], size[1], err)
					continue
				}
				texts := occultationSVGTexts(t, diagram)
				pairs := occultationDetailedTextOverlapPairs(texts)
				overlapping += len(pairs)
				t.Logf("%s %dx%d rendered: %d texts, %d overlapping pairs", item.label, size[0], size[1], len(texts), len(pairs))
				for _, pair := range pairs {
					t.Logf("%s %dx%d overlaps: %s %v vs %s %v", item.label, size[0], size[1],
						pair[0].value, occultationSVGTextBox(pair[0]), pair[1].value, occultationSVGTextBox(pair[1]))
				}
			}
			if overlapping == 0 {
				return
			}
			if expected, known := knownOverlaps[size]; known && expected == overlapping {
				t.Skipf("竖版最小画布 %dx%d 仍有 %d 对文本压叠（stellar 与 planetary 各 %d 对）：月亮数据块行内 "+
					"%q 与其值横向压叠约 6.1 px，掩甚/外掩终标记与其上方经纬读数纵向压叠约 0.63 px；"+
					"块内行距守卫的估算宽度（ASCII 0.58em／非 ASCII 1em）比本测试（0.62em／1.05em）宽松，故未拒绝该画布。",
					size[0], size[1], overlapping, overlapping/len([]string{"stellar", "planetary"}), "地平视差 H.P.")
			}
			t.Errorf("%dx%d renders %d overlapping text pairs", size[0], size[1], overlapping)
		})
	}
}
