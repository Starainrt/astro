package svg

import (
	"errors"
	"testing"
)

// 数据块行距不足时详细版必须拒绝画布，而不是画出压叠的面板文字。
func TestOccultationDetailedSVGRejectsCanvasesThatCrowdPanels(t *testing.T) {
	for _, testCase := range []struct {
		name          string
		width, height int
		rejected      bool
	}{
		{name: "landscape too short", width: 900, height: 400, rejected: true},
		{name: "landscape borderline", width: 900, height: 560, rejected: true},
		{name: "portrait too narrow", width: 600, height: 800, rejected: true},
		{name: "smallest landscape", width: 800, height: 600},
		{name: "default page", width: 1000, height: 1414},
		{name: "landscape page", width: 1414, height: 1000},
	} {
		for name, render := range map[string]func() (string, error){
			"star": func() (string, error) {
				return StarOccultationDetailedSVG(occultationTestStarPath(t), hr4799StarCoordinate(),
					OccultationDetailedSVGOptions{Width: testCase.width, Height: testCase.height})
			},
			"planet": func() (string, error) {
				return PlanetOccultationDetailedSVG(occultationTestSaturnPath(t),
					OccultationDetailedSVGOptions{Width: testCase.width, Height: testCase.height})
			},
		} {
			document, err := render()
			if testCase.rejected {
				if !errors.Is(err, ErrInvalidOccultationDetailedSVGOptions) {
					t.Fatalf("%s %s: err = %v, want ErrInvalidOccultationDetailedSVGOptions", name, testCase.name, err)
				}
				continue
			}
			if err != nil {
				t.Fatalf("%s %s: %v", name, testCase.name, err)
			}
			if len(document) == 0 {
				t.Fatalf("%s %s: empty document", name, testCase.name)
			}
		}
	}
}
