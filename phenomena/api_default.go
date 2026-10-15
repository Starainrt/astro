package phenomena

import "time"

// LongitudeCrossings 视黄经经过 longitude 的全部时刻 / times when the apparent longitude equals longitude.
func LongitudeCrossings(body Body, longitude float64, start, end time.Time) []Event {
	return defaultModule.LongitudeCrossings(body, longitude, start, end)
}

// Stations 视黄经留 / longitude stations.
func Stations(body Body, start, end time.Time) []Event {
	return defaultModule.Stations(body, start, end)
}

// AspectCrossings a 与 b 视黄经差（a − b）等于 angle 的时刻 / times when lon(a) − lon(b) equals angle.
func AspectCrossings(a, b Body, angle float64, start, end time.Time) []Event {
	return defaultModule.AspectCrossings(a, b, angle, start, end)
}

// ClosestApproaches 两天体最近角距 / closest approaches of two bodies.
func ClosestApproaches(a, b Body, maxSeparation float64, start, end time.Time) []Event {
	return defaultModule.ClosestApproaches(a, b, maxSeparation, start, end)
}

// StarApproaches 天体与恒星最近角距 / closest approaches of a body to a star.
func StarApproaches(body Body, star StarCoordinate, maxSeparation float64, start, end time.Time) []Event {
	return defaultModule.StarApproaches(body, star, maxSeparation, start, end)
}

// PlanetHeliacalEvents 行星偕日事件 / planet heliacal events.
func PlanetHeliacalEvents(body Body, start, end time.Time, lon, lat float64, options HeliacalOptions) []HeliacalEvent {
	return defaultModule.PlanetHeliacalEvents(body, start, end, lon, lat, options)
}

// StarHeliacalEvents 恒星偕日事件 / star heliacal events.
func StarHeliacalEvents(star StarCoordinate, magnitude float64, start, end time.Time, lon, lat float64, options HeliacalOptions) []HeliacalEvent {
	return defaultModule.StarHeliacalEvents(star, magnitude, start, end, lon, lat, options)
}

// Apsides 行星或地球的近日点与远日点 / perihelia and aphelia.
func Apsides(body Body, start, end time.Time) []OrbitalEvent {
	return defaultModule.Apsides(body, start, end)
}

// Nodes 行星的升交点与降交点 / node passages.
func Nodes(body Body, start, end time.Time) []OrbitalEvent {
	return defaultModule.Nodes(body, start, end)
}

// AltitudeCrossing 日、月或行星穿过任意视高度的时刻 / altitude crossings with the default ephemeris.
func AltitudeCrossing(body Body, date time.Time, lon, lat, height, altitude float64, options AltitudeOptions) (AltitudeCrossingResult, error) {
	return defaultModule.AltitudeCrossing(body, date, lon, lat, height, altitude, options)
}
