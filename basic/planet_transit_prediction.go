package basic

import (
	"math"

	"github.com/starainrt/astro/internal/ephem"
)

func planetTransitPredictedSeason(seedTT float64, cfg planetTransitConfig, inst *Instance) bool {
	centuries := (seedTT - 2451545) / 36525
	if inst != nil && inst.Model() == ephem.Custom || centuries < -40 || centuries > 20 {
		return true
	}
	node := 76.679920 + centuries*(0.9011206+centuries*(0.00040618-0.000000093*centuries))
	if cfg.planetIndex == 1 {
		node = 48.330893 + centuries*(1.1861883+centuries*(0.00017542+0.000000215*centuries))
	}
	solar := 280.46646 + centuries*(36000.76983+0.0003032*centuries)
	phase := math.Abs(math.Remainder(solar-node, 180))
	return phase <= 24
}

func planetTransitCoarseConjunction(seedTT float64, cfg planetTransitConfig, inst *Instance) float64 {
	current := seedTT
	for iteration := 0; iteration < 5; iteration++ {
		value := planetTransitLongitudeDeltaN(current, cfg, planetTransitCoarseN, inst)
		slope := (planetTransitLongitudeDeltaN(current+0.02, cfg, planetTransitCoarseN, inst) -
			planetTransitLongitudeDeltaN(current-0.02, cfg, planetTransitCoarseN, inst)) / 0.04
		if !isFiniteFloat(slope) || slope == 0 {
			return math.NaN()
		}
		step := value / slope
		current -= step
		if math.Abs(step) < 0.0001 {
			break
		}
	}
	return current
}

// PlanetTransitSeries 凌日周期族，Number 为零表示超出编号适用范围 / transit family; a zero Number means unassigned.
type PlanetTransitSeries struct {
	Number      int
	PeriodYears int
}

func planetTransitSeriesAt(greatestTT float64, cfg planetTransitConfig) PlanetTransitSeries {
	period, synodicCycles := 243, 152
	firstYear, lastYear := -2000, 4000
	anchors := [][4]int{{-1892, 5, 21, 1}, {-1998, 11, 18, 2}, {-1884, 5, 19, 3}, {-1763, 11, 20, 4}, {554, 5, 22, 5}, {1631, 12, 7, 6}}
	if cfg.planetIndex == 1 {
		period, synodicCycles = 46, 145
		firstYear, lastYear = 1601, 2300
		anchors = [][4]int{{1644, 11, 9, 1}, {1631, 11, 7, 2}, {1628, 5, 5, 3}, {1618, 11, 4, 4}, {1615, 5, 3, 5}, {1605, 11, 1, 6}, {1740, 5, 2, 7}, {1776, 11, 2, 8}, {1957, 5, 6, 9}, {1993, 11, 6, 10}, {2174, 5, 8, 11}, {2210, 11, 9, 12}}
	}
	result := PlanetTransitSeries{PeriodYears: period}
	if greatestTT < JDCalc(firstYear, 1, 1) || greatestTT >= JDCalc(lastYear+1, 1, 1) {
		return result
	}
	cycle := int(math.Round((greatestTT - cfg.anchorInferiorTT) / cfg.synodicPeriodDays))
	for _, anchor := range anchors {
		anchorCycle := int(math.Round((JDCalc(anchor[0], anchor[1], float64(anchor[2])) - cfg.anchorInferiorTT) / cfg.synodicPeriodDays))
		if (cycle-anchorCycle)%synodicCycles == 0 {
			result.Number = anchor[3]
			return result
		}
	}
	return result
}
