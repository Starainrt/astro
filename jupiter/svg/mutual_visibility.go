package svg

import (
	"math"
	"time"

	"github.com/starainrt/astro/basic"
	"github.com/starainrt/astro/jupiter"
)

type contactVisibility struct {
	hiddenBy int
	fraction float64
	jupiter  bool
	shadow   bool
}

// hidden 被遮星在地球视向上被前方卫星挡住过半、被木星掩蔽或处于木影中。
func (visibility contactVisibility) hidden() bool {
	return visibility.jupiter || visibility.shadow || visibility.fraction >= 0.5
}

// contactVisibilityAt 某时刻被遮星从地球看是否可见；遮挡比例按两圆交面积对被遮盘求。
func contactVisibilityAt(date time.Time, passive int) contactVisibility {
	jde := basic.UTC2TT(basic.Date2JD(date.UTC()))
	snapshot, ok := basic.JupiterGalileanDiagramWithInstance(jde, nil)
	if !ok {
		return contactVisibility{}
	}
	target := snapshot.Observations[passive-1]
	radius := snapshot.AngularRadiiJupiterR[passive-1]
	result := contactVisibility{
		jupiter: snapshot.Phenomena[passive-1].Occultation,
		shadow:  snapshot.Phenomena[passive-1].Eclipse,
	}
	for index, observation := range snapshot.Observations {
		if index == passive-1 || observation.OffsetZJupiterRadii >= target.OffsetZJupiterRadii {
			continue
		}
		distance := math.Hypot(observation.OffsetXJupiterRadii-target.OffsetXJupiterRadii, observation.OffsetYJupiterRadii-target.OffsetYJupiterRadii)
		fraction := diskOverlap(distance, snapshot.AngularRadiiJupiterR[index], radius) / (math.Pi * radius * radius)
		if fraction > result.fraction {
			result.hiddenBy, result.fraction = index+1, math.Min(1, fraction)
		}
	}
	return result
}

func diskOverlap(distance, first, second float64) float64 {
	if distance >= first+second {
		return 0
	}
	if distance <= math.Abs(first-second) {
		smaller := math.Min(first, second)
		return math.Pi * smaller * smaller
	}
	cosAlpha := math.Max(-1, math.Min(1, (distance*distance+first*first-second*second)/(2*distance*first)))
	cosBeta := math.Max(-1, math.Min(1, (distance*distance+second*second-first*first)/(2*distance*second)))
	alpha := math.Acos(cosAlpha)
	beta := math.Acos(cosBeta)
	kite := 0.5 * math.Sqrt(math.Max(0, (-distance+first+second)*(distance+first-second)*(distance-first+second)*(distance+first+second)))
	return first*first*alpha + second*second*beta - kite
}

func (drawing *chart) visibilityNote(visibility contactVisibility) string {
	switch {
	case visibility.jupiter:
		return drawing.words("木星掩蔽", "behind Jupiter")
	case visibility.shadow:
		return drawing.words("在木影中", "in Jupiter's shadow")
	case visibility.fraction >= 0.5:
		return drawing.words("被"+satelliteName(visibility.hiddenBy-1, false)+"掩", "occulted by "+satelliteName(visibility.hiddenBy-1, true))
	}
	return ""
}

// mutualVisibility 每个接触点的可见性；互掩时被遮星本来就被遮挡星挡住，只看木星与木影。
func mutualVisibility(event jupiter.GalileanMutualEvent, contacts []mutualContact) []contactVisibility {
	result := make([]contactVisibility, len(contacts))
	for index, contact := range contacts {
		visibility := contactVisibilityAt(contact.date, event.Passive)
		if event.Type == jupiter.GalileanMutualOccultation && visibility.hiddenBy == event.Active {
			visibility.hiddenBy, visibility.fraction = 0, 0
		}
		result[index] = visibility
	}
	return result
}
