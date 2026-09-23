# Accuracy and performance

[中文](../accuracy.md) | [README](../../../README.en.md)

These tables describe the built-in models and previously measured comparisons. A sampled maximum is not an error bound for every date or site.

External comparisons require matching the time scale, coordinate frame, observer height and refraction model.

See [Time scales](timescale.md) for UTC, UT1 and TT.

## Contents

- [Sun and planets](#sun-and-planets)
- [Moon](#moon)
- [Lite lightweight chains](#lite-lightweight-chains)
- [Accuracy references](#accuracy-references)

## Sun and planets

The Sun and planets use built-in VSOP87 analytical terms. The current table entries cover roughly 4000 years around J2000. The table below lists truncation errors relative to the complete VSOP87 tables:

| Target | Longitude / latitude | Distance |
| --- | --- | --- |
| Sun / Earth | about `0.1"` | about `0.1 x 10^-6 AU` |
| Mercury, Venus | about `0.2"` | about `0.2 x 10^-6 AU` |
| Mars | about `0.5"` | about `1 x 10^-6 AU` |
| Jupiter | about `0.5"` | about `3 x 10^-6 AU` |
| Saturn | about `0.5"` | about `5 x 10^-6 AU` |
| Uranus | about `1"` | about `20 x 10^-6 AU` |
| Neptune | about `1"` | about `40 x 10^-6 AU` |

This is suitable for ordinary calendrical work, observing support, outreach, and personal research; spacecraft navigation, precise occultation prediction, and strict dynamical integration fall outside that range and usually need a professional ephemeris such as JPL DE.

## Moon

The Moon uses a built-in truncated ELP2000/82-style analytical series. The package stays lightweight and does not require external ephemeris files.

It is suitable for Chinese-calendar new moons, lunar phases, rise/set, lunar eclipses, amateur occultation prediction, and ordinary positional work; extremely high-precision lunar laser ranging, long-term physical libration, and professional occultation work fall outside that range and are best served by JPL or a dedicated lunar ephemeris.

## Lite lightweight chains

`lite/sun` and `lite/moon` are independent approximation chains. They do not depend on the VSOP87 or ELP2000/82 series used by `sun` / `moon`, and are intended for CPU- or memory-constrained environments.

- `lite/sun`: simplified true/apparent solar longitude formulas plus lightweight equatorial conversion
- `lite/moon`: Schlyter-style lunar approximation with about 15 perturbation terms plus lightweight topocentric correction
- rise/set search: fixed-step scanning plus bisection, without the high-precision nutation iteration used by the main chain
- zero heap allocation in the computation path (0 allocs/op); against the main chain, pure evaluation entry points such as position and phase run about `7.5-27.1x` faster, and rise/set entry points about `0.9-3.5x`

Capability boundaries:

| Package | Position model | Rise/set search | Main use |
| --- | --- | --- | --- |
| `lite/sun` | simplified true/apparent solar longitude plus lightweight equatorial conversion | `30 min` scan plus bisection | sunrise/sunset, solar altitude, watch faces, frontend refresh loops |
| `lite/moon` | Schlyter / vFPS lunar approximation plus lightweight topocentric correction | `15 min` scan plus bisection | moonrise/moonset, lunar phase, lunar age, lightweight lunar observing helpers |

Error against the `sun` / `moon` packages (year 2026, 8 observing sites; rise/set sampled every 7 or 15 days, phase/age every 6 hours):

| Capability | Mean absolute error | P95 | Max absolute error | Notes |
| --- | --- | --- | --- | --- |
| `lite/sun` sunrise | `0.02 min` | `0.04 min` | `0.31 min` | no event-existence mismatch in the sample set |
| `lite/sun` sunset | `0.02 min` | `0.06 min` | `0.35 min` | `2` high-latitude samples differ only in day-attribution semantics across midnight |
| `lite/moon` moonrise | `0.28 min` | `0.57 min` | `1.44 min` | no event-existence mismatch in the sample set |
| `lite/moon` moonset | `0.36 min` | `0.86 min` | `1.24 min` | `1` high-latitude sample differs on whether the moonset belongs to the same civil day |
| `lite/moon` `Phase()` | `0.00089` | `0.00185` | `0.00243` | compared with `moon.Phase` |
| `lite/moon` `PhaseAge()` | `0.003 d` | `0.010 d` | `0.014 d` | about 4.3 min mean, 14.4 min P95, 20.2 min max |
| `lite/moon` geocentric longitude | `2.41'` | `6.82'` | `9.91'` | relative to the main lunar chain |
| `lite/moon` geocentric latitude | `0.87'` | `1.83'` | `2.92'` | relative to the main lunar chain |

`Go testing.Benchmark` reference values (single-machine measurements for comparison; absolute values vary with hardware):

| Entry point | Main chain | `lite` | Speedup | Main-chain allocation | `lite` allocation |
| --- | --- | --- | --- | --- | --- |
| `Sun ApparentRaDec` | `6.031 µs/op` | `222.4 ns/op` | `27.1x` | `0 B/op, 0 allocs/op` | `0 B/op, 0 allocs/op` |
| `Sun Altitude` | `6.127 µs/op` | `672.3 ns/op` | `9.1x` | `0 B/op, 0 allocs/op` | `0 B/op, 0 allocs/op` |
| `Sun RiseTime` | `101.874 µs/op` | `29.168 µs/op` | `3.5x` | `0 B/op, 0 allocs/op` | `0 B/op, 0 allocs/op` |
| `Moon ApparentRaDec` | `16.897 µs/op` | `1.070 µs/op` | `15.8x` | `0 B/op, 0 allocs/op` | `0 B/op, 0 allocs/op` |
| `Moon Phase` | `15.441 µs/op` | `935.6 ns/op` | `16.5x` | `0 B/op, 0 allocs/op` | `0 B/op, 0 allocs/op` |
| `Moon Altitude` | `9.714 µs/op` | `1.294 µs/op` | `7.5x` | `0 B/op, 0 allocs/op` | `0 B/op, 0 allocs/op` |
| `Moon RiseTime` | `121.772 µs/op` | `132.312 µs/op` | `0.9x` | `0 B/op, 0 allocs/op` | `0 B/op, 0 allocs/op` |

The main-chain/`lite` gap depends on the scenario: pure evaluation entry points (position, phase) run about `7.5-27.1x` faster in `lite`, while rise/set entry points narrow to `0.9-3.5x` because both sides perform a time search; `Moon RiseTime` is close to parity.

Use the main `sun` / `moon` chains for eclipses, physical libration, or high-latitude edge cases.

## Accuracy references

The following entry points have been checked against JPL Horizons, NASA GSFC, and other public references; use them to judge the order of magnitude to expect:

- apparent diameters of the Sun, planets, and Moon: maximum differences from the external baseline range from `0.000002"` to `0.194598"` depending on the body; the Moon is the most sensitive because of parallax and distance changes
- solar physical ephemerides `P/B0/L0`: maximum differences are about `0.003349° / 0.003986° / 0.047394°`
- planetary rise, transit, and set: checked against JPL Horizons rise/transit/set events; that baseline is generated at a 1-minute step, and current results align with the Horizons event times at the minute level
- Moon rise/set: `aero=true` uses dynamic standard refraction and the instantaneous lunar semidiameter for an upper-limb crossing.

  Across 14 sea-level events at 7 sites, the current mean/maximum differences against JPL Horizons DE441 are about `0.30s / 0.75s`.
- Moon rise/set with other conventions: mean/maximum differences are about `38.77s / 76.22s` against MET Norway's fixed `-0.8333°` convention (Skyfield 1.53 + DE440s). Against IMCCE Miriade, whose horizon convention is not exposed, the mean is about `2m13.46s`; the low-elevation `61°N` sample reaches about `6m41.82s`.
- Earth perihelion and aphelion: maximum time difference about `1m28.84s`, maximum distance difference about `0.000000039837 AU`
- main-chain lunar position: the current algorithm is a truncated ELP2000/82-style analytical series; across four JPL/Horizons `JDTT` samples in year `-2000`, the maximum difference from JPL/Horizons is about `219.6"` in longitude, `25.8"` in latitude, and `34.3 km` in distance
- Moon perigee and apogee: maximum time difference about `15m53.45s`, maximum distance difference about `39.758 km`
- maximum lunar declination: maximum time difference about `2.43s`, maximum declination difference about `0.00006431°`
