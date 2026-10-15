package basic

// 月缘等效半径必须与全部位置角的暴力判据一致，且站心全食与中心带多边形对同一地点结论一致。

import (
	"math"
	"testing"
	"time"

	"github.com/starainrt/astro/internal/geodata"
	"github.com/starainrt/astro/moon/limb"
)

func limbBruteMargins(contact lunarLimbContact) (outer, cover, inside float64) {
	outer, cover, inside = math.Inf(-1), math.Inf(1), math.Inf(1)
	step := contact.frame.stepDeg / 4
	start := math.Ceil(contact.view.angle/step)*step - contact.view.angle
	for offset := start - 180; offset < 180; offset += step {
		radius, ok := contact.radius(offset, true)
		if !ok {
			continue
		}
		// 远交点为负的射线背向日面，不是真实交点；月心在日面外时远交点为正即近交点为正。
		near, far, hit := contact.rayChord(offset)
		if hit && far > 0 {
			outer = math.Max(outer, radius-near)
			cover = math.Min(cover, radius-far)
		}
		if _, far, hit := contact.rayChord(offset + 180); hit && far > 0 {
			if radius, ok := contact.radius(offset+180, true); ok {
				inside = math.Min(inside, far-radius)
			}
		}
	}
	return
}

func TestLunarLimbContactMatchesBruteForce(t *testing.T) {
	seed := Date2JD(time.Date(2035, 9, 2, 0, 0, 0, 0, time.UTC))
	params := solarEclipseModelParams(SolarEclipseModelNASABulletinSplitK, SolarEclipseSunRadiusStandard)
	params.lunarLimb = limb.Default()
	greatest := localSolarEclipseGreatest(solarEclipseAnchorNewMoon(seed), 116.2310*rad, 39.7215*rad, 0, params)
	const tolerance = 1e-4 / 3600 * rad
	for _, dtSec := range []float64{-60, -20, -5, 0, 5, 20, 60} {
		context := newLocalSolarEclipseStateContext(greatest+dtSec/86400, params)
		for _, latDeg := range []float64{39.70, 39.7215, 39.73, 39.74} {
			observer := localSolarEclipseObserverXYZ(context.gst, 116.2310*rad, latDeg*rad, 0)
			moon := [3]float64{context.moonXYZ[0] - observer[0], context.moonXYZ[1] - observer[1], context.moonXYZ[2] - observer[2]}
			sun := [3]float64{context.sunXYZ[0] - observer[0], context.sunXYZ[1] - observer[1], context.sunXYZ[2] - observer[2]}
			view, ok := context.limbFrame.viewToward(moon, sun)
			if !ok {
				t.Fatal("view failed")
			}
			state := context.stateAt(116.2310*rad, latDeg*rad, 0)
			contact := lunarLimbContact{frame: context.limbFrame, view: view, distanceKM: math.Sqrt(lunarLimbDot(moon, moon)),
				separation: state.separationRad, targetRadius: state.sunRadiusRad}
			outer, cover, _ := limbBruteMargins(contact)
			low, high, _ := contact.bounds()
			relief := high - low
			d, sunRadius := state.separationRad, state.sunRadiusRad
			// 门控区只保证间隙符号，门控外须与暴力判据逐值一致。
			gotOuter, _ := contact.outerRadius()
			wantOuterGap := -outer
			if gotGap := d - gotOuter - sunRadius; d > sunRadius && math.Abs(wantOuterGap) < relief {
				if math.Abs(gotGap-wantOuterGap) > tolerance {
					t.Errorf("dt %.0f lat %.4f: outer gap %.4f\" want %.4f\"", dtSec, latDeg, gotGap/rad*3600, wantOuterGap/rad*3600)
				}
			} else if (gotGap > 0) != (wantOuterGap > 0) {
				t.Errorf("dt %.0f lat %.4f: outer gap sign %.4f\" want %.4f\"", dtSec, latDeg, gotGap/rad*3600, wantOuterGap/rad*3600)
			}
			gotInner, _ := contact.innerRadius(true)
			gotGap := d - math.Abs(gotInner-sunRadius)
			wantGap := math.Min(-cover, d)
			if math.Abs(cover) < relief {
				if math.Abs(gotGap-wantGap) > tolerance {
					t.Errorf("dt %.0f lat %.4f: inner gap %.4f\" want %.4f\"", dtSec, latDeg, gotGap/rad*3600, wantGap/rad*3600)
				}
			} else if (gotGap > 0) != (wantGap > 0) {
				t.Errorf("dt %.0f lat %.4f: inner gap sign %.4f\" want %.4f\"", dtSec, latDeg, gotGap/rad*3600, wantGap/rad*3600)
			}
		}
	}
}

func limbPointInBand(segments [][]SolarEclipsePathPoint, lon, lat float64) bool {
	inside := false
	for _, ring := range segments {
		for i, j := 0, len(ring)-1; i < len(ring); j, i = i, i+1 {
			a, b := ring[i], ring[j]
			if (a.Latitude > lat) != (b.Latitude > lat) &&
				lon < (b.Longitude-a.Longitude)*(lat-a.Latitude)/(b.Latitude-a.Latitude)+a.Longitude {
				inside = !inside
			}
		}
	}
	return inside
}

func TestLimbCentralBandMatchesLocalTotality(t *testing.T) {
	seed := Date2JD(time.Date(2035, 9, 2, 0, 0, 0, 0, time.UTC))
	profile := limb.Default()
	path := SolarEclipseCentralPathNASABulletinSplitKWithInstance(seed, SolarEclipsePathOptions{LunarLimb: profile}, nil)
	options := SolarEclipseOptions{RadiusModel: SolarEclipseModelNASABulletinSplitK, LunarLimb: profile}
	const lon, marginDeg = 116.2310, 0.5 / 111.2
	for lat := 39.60; lat <= 39.85; lat += 0.005 {
		inBand := limbPointInBand(path.CentralBandSegments, lon, lat)
		if inBand != limbPointInBand(path.CentralBandSegments, lon, lat-marginDeg) ||
			inBand != limbPointInBand(path.CentralBandSegments, lon, lat+marginDeg) {
			continue
		}
		local := LocalSolarEclipseWithOptionsWithInstance(seed, lon, lat, 0, options, nil)
		if local.HasCentral != inBand {
			t.Errorf("lat %.4f: local central %v, band %v", lat, local.HasCentral, inBand)
		}
	}
	local := LocalSolarEclipseWithOptionsWithInstance(seed, lon, 39.7215, 0, options, nil)
	if local.HasCentral || limbPointInBand(path.CentralBandSegments, lon, 39.7215) {
		t.Errorf("39.7215: local central %v must agree with band (outside)", local.HasCentral)
	}
}

// limbTestContact 2035-09-02 北京附近某一时刻的月缘接触上下文，日面半径与角距可替换以构造环食。
func limbTestContact(t *testing.T) lunarLimbContact {
	t.Helper()
	seed := Date2JD(time.Date(2035, 9, 2, 0, 0, 0, 0, time.UTC))
	params := solarEclipseModelParams(SolarEclipseModelNASABulletinSplitK, SolarEclipseSunRadiusStandard)
	params.lunarLimb = limb.Default()
	greatest := localSolarEclipseGreatest(solarEclipseAnchorNewMoon(seed), 116.2310*rad, 39.7215*rad, 0, params)
	context := newLocalSolarEclipseStateContext(greatest, params)
	observer := localSolarEclipseObserverXYZ(context.gst, 116.2310*rad, 39.7215*rad, 0)
	moon := [3]float64{context.moonXYZ[0] - observer[0], context.moonXYZ[1] - observer[1], context.moonXYZ[2] - observer[2]}
	sun := [3]float64{context.sunXYZ[0] - observer[0], context.sunXYZ[1] - observer[1], context.sunXYZ[2] - observer[2]}
	view, ok := context.limbFrame.viewToward(moon, sun)
	if !ok {
		t.Fatal("view failed")
	}
	state := context.stateAt(116.2310*rad, 39.7215*rad, 0)
	return lunarLimbContact{frame: context.limbFrame, view: view, distanceKM: math.Sqrt(lunarLimbDot(moon, moon)),
		separation: state.separationRad, targetRadius: state.sunRadiusRad}
}

func TestLunarLimbAnnularAndUnrangedMatchBruteForce(t *testing.T) {
	base := limbTestContact(t)
	const tolerance = 1e-4 / 3600 * rad
	meanRadius := math.Asin(1737.4 / base.distanceKM)
	unranged := *base.frame
	unranged.ranged = false
	for _, ranged := range []bool{true, false} {
		for _, separationArcsec := range []float64{0, 1, 5, 12, 20} {
			contact := base
			if !ranged {
				contact.frame = &unranged
			}
			contact.targetRadius = meanRadius + 6/3600.0*rad
			contact.separation = separationArcsec / 3600 * rad
			_, _, inside := limbBruteMargins(contact)
			got, ok := contact.innerRadius(false)
			if !ok {
				t.Fatalf("ranged %v d %.0f\": annular radius failed", ranged, separationArcsec)
			}
			// 环食等效内半径必须小于日面半径，否则食分会大于 1；只比间隙看不出 R 关于 s 的镜像错误。
			if got > contact.targetRadius {
				t.Errorf("ranged %v d %.0f\": annular radius %.4f\" exceeds the solar radius %.4f\"", ranged, separationArcsec, got/rad*3600, contact.targetRadius/rad*3600)
			}
			gotGap := contact.separation - math.Abs(got-contact.targetRadius)
			wantGap := math.Min(-inside, contact.separation)
			if math.Abs(gotGap-wantGap) > tolerance && !(gotGap > 0 && wantGap > 0) {
				t.Errorf("ranged %v d %.0f\": annular gap %.4f\" want %.4f\"", ranged, separationArcsec, gotGap/rad*3600, wantGap/rad*3600)
			}
			if !ranged {
				contact.targetRadius = meanRadius - 6/3600.0*rad
				contact.separation = (separationArcsec + 1900) / 3600 * rad
				outer, _, _ := limbBruteMargins(contact)
				got, ok := contact.outerRadius()
				if !ok {
					t.Fatal("unranged outer failed")
				}
				if gap := contact.separation - got - contact.targetRadius; math.Abs(gap+outer) > tolerance {
					t.Errorf("unranged d %.0f\": outer gap %.4f\" want %.4f\"", separationArcsec, gap/rad*3600, -outer/rad*3600)
				}
			}
		}
	}
}

func TestLimbCentralLimitMatchesLocalLimit(t *testing.T) {
	seed := Date2JD(time.Date(2035, 9, 2, 0, 0, 0, 0, time.UTC))
	profile := limb.Default()
	path := SolarEclipseCentralPathNASABulletinSplitKWithInstance(seed, SolarEclipsePathOptions{LunarLimb: profile}, nil)
	options := SolarEclipseOptions{RadiusModel: SolarEclipseModelNASABulletinSplitK, LunarLimb: profile}
	const lon = 116.2310
	bisect := func(inside func(float64) bool) float64 {
		low, high := 39.5, 39.9
		for i := 0; i < 40; i++ {
			mid := (low + high) / 2
			if inside(mid) {
				high = mid
			} else {
				low = mid
			}
		}
		return (low + high) / 2
	}
	local := bisect(func(lat float64) bool {
		return LocalSolarEclipseWithOptionsWithInstance(seed, lon, lat, 0, options, nil).HasCentral
	})
	band := bisect(func(lat float64) bool { return limbPointInBand(path.CentralBandSegments, lon, lat) })
	if d := math.Abs(local-band) * 111.2; d > 0.5 {
		t.Errorf("southern limit: local %.5f, band %.5f, %.2f km apart", local, band, d)
	}
}

func TestLocalSolarEclipseLimbCanFlipTotalToAnnular(t *testing.T) {
	seed := Date2JD(time.Date(2023, 4, 20, 0, 0, 0, 0, time.UTC))
	const lon, lat = 170.1481, 4.6940
	circular := LocalSolarEclipseWithOptionsWithInstance(seed, lon, lat, 0, SolarEclipseOptions{RadiusModel: SolarEclipseModelNASABulletinSplitK}, nil)
	if circular.Type != SolarEclipseTotal {
		t.Fatalf("circular type %v, want total near the hybrid transition", circular.Type)
	}
	// 常数月缘比圆月面内半径小 2 km：圆月面判全食，月缘下日面在各方向都露出，应为环食。
	innerKM := geodata.EarthEquatorialRadiusKM * solarEclipseUmbralK * localSolarMoonRadiusScale
	smaller := &recordingLunarLimb{radiusKM: innerKM - 2}
	limbed := LocalSolarEclipseWithOptionsWithInstance(seed, lon, lat, 0,
		SolarEclipseOptions{RadiusModel: SolarEclipseModelNASABulletinSplitK, LunarLimb: smaller}, nil)
	if limbed.Type != SolarEclipseAnnular || !limbed.HasCentral || !limbed.HasAnnular || limbed.HasTotal {
		t.Fatalf("smaller constant limb: type %v central %v annular %v total %v, want annular", limbed.Type, limbed.HasCentral, limbed.HasAnnular, limbed.HasTotal)
	}
	if limbed.CentralEnd <= limbed.CentralStart {
		t.Errorf("annular phase has no duration: %.8f–%.8f", limbed.CentralStart, limbed.CentralEnd)
	}
}

// sectorLunarLimb 只在以 center 为中心、半宽 halfWidth 的位置角扇区内给出剖面值，并报告整张表的半径范围。
type sectorLunarLimb struct {
	center, halfWidth float64
	profile           *limb.Profile
}

func (sector sectorLunarLimb) LimbRadiusKM(lon, lat, angle float64) (float64, bool) {
	if math.Abs(math.Mod(angle-sector.center+540, 360)-180) > sector.halfWidth {
		return 0, false
	}
	return sector.profile.LimbRadiusKM(lon, lat, angle)
}

func (sector sectorLunarLimb) RadiusRangeKM() (float64, float64) {
	return sector.profile.RadiusRangeKM()
}

func (sector sectorLunarLimb) AngleStepDeg() float64 { return sector.profile.AngleStepDeg() }

func TestLocalSolarEclipseStateSeparatesOuterAndInnerLimb(t *testing.T) {
	params := solarEclipseModelParams(SolarEclipseModelNASABulletinSplitK, SolarEclipseSunRadiusStandard)
	seed := Date2JD(time.Date(2035, 9, 2, 0, 0, 0, 0, time.UTC))
	const lon, lat = 116.2310 * rad, 39.70 * rad
	// 食甚附近月盘深入日面：外切间隙远离零点被门控，只查日心方向；内切接近零点必须全方向扫描。
	tt := localSolarEclipseGreatest(solarEclipseAnchorNewMoon(seed), lon, lat, 0, params)
	circular := newLocalSolarEclipseStateContext(tt, params).stateAt(lon, lat, 0)
	params.lunarLimb = limb.Default()
	full := newLocalSolarEclipseStateContext(tt, params)
	observer := localSolarEclipseObserverXYZ(full.gst, lon, lat, 0)
	view, ok := full.limbFrame.viewToward(
		[3]float64{full.moonXYZ[0] - observer[0], full.moonXYZ[1] - observer[1], full.moonXYZ[2] - observer[2]},
		[3]float64{full.sunXYZ[0] - observer[0], full.sunXYZ[1] - observer[1], full.sunXYZ[2] - observer[2]})
	if !ok {
		t.Fatal("view failed")
	}
	params.lunarLimb = sectorLunarLimb{center: view.angle, halfWidth: 2, profile: limb.Default()}
	state := newLocalSolarEclipseStateContext(tt, params).stateAt(lon, lat, 0)
	if !state.outerLimbApplied || state.limbApplied {
		t.Fatalf("outerLimbApplied %v limbApplied %v, want outer only", state.outerLimbApplied, state.limbApplied)
	}
	if state.moonInnerRadiusRad != circular.moonInnerRadiusRad || state.moonOuterRadiusRad == circular.moonOuterRadiusRad {
		t.Errorf("inner radius must stay circular and outer must move: inner %.9g vs %.9g, outer %.9g vs %.9g",
			state.moonInnerRadiusRad, circular.moonInnerRadiusRad, state.moonOuterRadiusRad, circular.moonOuterRadiusRad)
	}
}
