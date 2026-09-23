package svg

import (
	"testing"
	"time"

	"github.com/starainrt/astro/basic"
	eclipsecore "github.com/starainrt/astro/eclipse"
	"github.com/starainrt/astro/internal/svgmap"
)

// TestLunarEclipseMapSVGPolarWitnessStaysOutsideEquirectangularVisibility 固定全球月食地图在
// 极区的可见性契约。等距圆柱投影在高纬把经度拉伸：1° 的球面地平弧在极点附近可以横跨近
// 180° 经度，只连端点就会把地平线以外画进可见区。1904-09-24 的 P1 在南极点附近有一个
// 月心高度约 −0.05° 的见证点，曾经被画进 visible-at-p1 多边形。
// TestLunarEclipseMapSVGPolarWitnessStaysOutsideEquirectangularVisibility pins the polar
// visibility contract of the global lunar-eclipse map. Equirectangular projection stretches
// longitude near a pole: a 1-degree spherical horizon arc can span almost 180 degrees there, so
// joining endpoints alone paints sky below the horizon as visible. At P1 of 1904-09-24 a witness
// near the south pole sits about 0.05 degrees below the Moon's horizon and used to be painted
// inside the visible-at-p1 polygon.
func TestLunarEclipseMapSVGPolarWitnessStaysOutsideEquirectangularVisibility(t *testing.T) {
	date := time.Date(1904, 9, 24, 0, 0, 0, 0, time.UTC)
	info, ok := eclipsecore.LunarEclipseOnDate(date)
	if !ok {
		t.Fatal("missing lunar eclipse")
	}
	frame := eclipseMapFrame(lunarEclipseMapDefaultWidth, lunarEclipseMapDefaultHeight,
		svgmap.ProjectionEquirectangular, 142, 92)
	path, boundary := lunarEclipseVisibilityPath(info.PenumbralStart, frame)
	if len(boundary) == 0 {
		t.Fatal("visibility path has no boundary")
	}
	rings := svgClosedPathRings(t, `<path d="`+path+`"/>`)
	if len(rings) == 0 {
		t.Fatal("visibility path has no rings")
	}
	jd := basic.Date2JD(info.PenumbralStart)
	// 只取足够深入地平线以下的见证点：−89.9° 处月心高度约 −0.054°，而更靠近极点的
	// −89.99° 只有约 −0.004°，落在可见性判定本身的数值容差里，不能作为回归依据。
	// Only a witness clearly below the horizon is used: at -89.9 degrees the Moon centre sits
	// about 0.054 degrees down, while -89.99 degrees is only about 0.004 degrees down and lies
	// inside the numerical tolerance of the visibility test itself.
	witnesses := []struct {
		longitude float64
		latitude  float64
	}{
		{-115, -89.9},
	}
	for _, witness := range witnesses {
		altitude := basic.HMoonHeight(jd, witness.longitude, witness.latitude, 0)
		if altitude >= -0.01 {
			// 见证点必须先确实是地平线以下，否则这个回归就失去意义。
			t.Fatalf("witness (%g,%g) altitude=%g, want below the horizon",
				witness.longitude, witness.latitude, altitude)
		}
		x, y, projectable := frame.Project(witness.longitude, witness.latitude)
		if !projectable {
			continue
		}
		inside := false
		for _, ring := range rings {
			inside = inside || pointInLunarVisibilityPolygon(x, y, ring)
		}
		if inside {
			t.Fatalf("below-horizon witness (%g,%g) altitude=%g is painted inside the visible region",
				witness.longitude, witness.latitude, altitude)
		}
	}
}
