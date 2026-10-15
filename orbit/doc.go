// Package orbit propagates lightweight Sun-centered two-body conic orbits.
//
// Supported input styles:
//   - classical elliptic elements: A/E/I/Omega/W/M0 at EpochJD
//   - perihelion form: Q/E/I/Omega/W/TpJD for high-eccentricity, parabolic,
//     or hyperbolic comet-like trajectories
//
// All input angles are in degrees. EpochJD and TpJD are TT/TDB Julian days.
// Returned geometric positions do not include perturbations beyond the supplied
// elements. Astrometric results include down-leg light-time correction.
// Optional planetary perturbations are available through PropagationOptions and
// StateWithOptions; a zero-value options value preserves the two-body path.
// Apparent results add annual aberration, nutation, and of-date coordinate
// conversion to the astrometric position. The package also provides observer-facing
// altitude/azimuth/hour-angle and rise/transit/set helpers built on top of the
// apparent topocentric coordinates. Rise/set helpers return package-local
// sentinel errors ERR_ORBIT_NEVER_RISE / ERR_ORBIT_NEVER_SET, matching the
// convention used by other public observation packages in this repository.
//
// In addition, the package includes a small classical visual-binary helper
// based on the standard P/T/e/a/i/Omega/omega element set, returning apparent
// position angle and separation on the sky.
package orbit
