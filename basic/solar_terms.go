package basic

import "math"

func GetMoonLoops(year float64, loop int) []float64 {
	var start float64
	var newMoon, lastNewMoon float64
	moonLoops := make([]float64, loop)
	if year < 6000 {
		start = year + 11.00/12.00 + 5.00/30.00/12.00
	} else {
		start = year + 9.00/12.00 + 5.00/30.00/12.00
	}
	i := 1
	for j := 0; j < loop; j++ {
		if year > 3000 {
			newMoon = TD2UT(CalcMoonSH(start+float64(i-1)/12.5, 0)+8.0/24.0, false)
		} else {
			newMoon = TD2UT(CalcMoonS(start+float64(i-1)/12.5, 0)+8.0/24.0, false)
		}
		if i != 1 {
			if newMoon == lastNewMoon {
				j--
				i++
				continue
			}
		}
		moonLoops[j] = newMoon
		lastNewMoon = moonLoops[j]
		i++
	}
	return moonLoops
}

func GetJieqiLoops(year, loop int) []float64 {
	start := 270
	jq := make([]float64, loop)
	for i := 1; i <= loop; i++ {
		angle := start + 15*(i-1)
		if angle > 360 {
			angle -= 360
		}
		jq[i-1] = GetJQTime(year+int(math.Ceil(float64(i-1)/24.000)), angle) + 8.0/24.0
	}
	return jq
}

func GetJQTime(year, angle int) float64 {
	// Calculate initial day based on angle parity
	var initialDay float64
	if angle%2 == 0 {
		initialDay = 18
	} else {
		initialDay = 3
	}

	// Calculate temporary factor for month offset
	var tempFactor float64
	if angle%10 != 0 {
		tempFactor = float64(angle+15) / 30.0
	} else {
		tempFactor = float64(angle) / 30.0
	}

	// Calculate initial month, adjusting if超过 12
	initialMonth := 3.0 + tempFactor
	if initialMonth > 12.0 {
		initialMonth -= 12.0
	}

	// Calculate initial Julian date
	initialJD := JDECalc(year, int(initialMonth), initialDay)

	// Set target angle for iteration; if angle is 0, use 360
	targetAngle := float64(angle)
	if angle == 0 {
		targetAngle = 360.0
	}

	// Newton-Raphson iteration to find precise Julian date
	currentJD := initialJD
	for {
		previousJD := currentJD
		errorValue := JQLospec(previousJD, targetAngle) - targetAngle
		derivative := (JQLospec(previousJD+0.000005, targetAngle) - JQLospec(previousJD-0.000005, targetAngle)) / 0.00001
		currentJD = previousJD - errorValue/derivative

		// Check for convergence
		if math.Abs(currentJD-previousJD) <= 0.00001 {
			break
		}
	}

	// Convert to UT and return
	return TD2UT(currentJD, false)
}

func JQLospec(jd float64, target float64) float64 {
	sunLo := HSunApparentLo(jd)
	if target >= 345 {
		if sunLo <= 12 {
			sunLo += 360
		}
	} else if target <= 15 {
		if sunLo >= 350 {
			sunLo -= 360
		}
	}
	return sunLo
}
