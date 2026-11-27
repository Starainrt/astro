package phenomena

import (
	"time"

	"github.com/starainrt/astro/basic"
)

const (
	// SurfaceRise 参考点升过目标高度 / reference point rising through the target altitude.
	SurfaceRise = basic.SurfaceRise
	// SurfaceSet 参考点落过目标高度 / reference point setting through the target altitude.
	SurfaceSet = basic.SurfaceSet
	// SurfaceUpperCulmination 上中天 / upper culmination.
	SurfaceUpperCulmination = basic.SurfaceUpperCulmination
	// SurfaceLowerCulmination 下中天 / lower culmination.
	SurfaceLowerCulmination = basic.SurfaceLowerCulmination
)

// ErrSurfaceEventNotFound 搜索上限内未找到指定事件 / requested event was not found within the search horizon.
var ErrSurfaceEventNotFound = basic.ErrSurfaceEventNotFound

// SurfaceEvent 行星表面的无折射事件，Time 为地球绝对时刻并保留输入时区 / airless surface event at an Earth instant in the input time zone.
type SurfaceEvent struct {
	Kind     SurfaceEventKind
	Time     time.Time
	Position SurfacePosition
	// Local 按 Time 时刻求得的本地时 / local times evaluated at Time.
	Local SurfaceLocalTime
}

func (module *Phenomena) surfaceEvent(event basic.SurfaceEvent, observer BodyObserver, loc *time.Location) (SurfaceEvent, error) {
	date := basic.JD2DateByZone(basic.TT2UTC(event.JDE), loc, false)
	local, err := module.SurfaceLocalTimeAt(observer, date)
	if err != nil {
		return SurfaceEvent{}, err
	}
	return SurfaceEvent{Kind: event.Kind, Time: date, Position: event.Position, Local: local}, nil
}

func (module *Phenomena) surfaceEventList(list []basic.SurfaceEvent, observer BodyObserver, loc *time.Location) ([]SurfaceEvent, error) {
	result := make([]SurfaceEvent, 0, len(list))
	for _, event := range list {
		converted, err := module.surfaceEvent(event, observer, loc)
		if err != nil {
			return nil, err
		}
		result = append(result, converted)
	}
	return result, nil
}

func (module *Phenomena) nextSurface(target Body, observer BodyObserver, kind SurfaceEventKind, after time.Time, options SurfaceOptions) (SurfaceEvent, error) {
	event, err := basic.NextSurfaceEventWithInstance(target, observer, kind, toTT(after), options, module.inst)
	if err != nil {
		return SurfaceEvent{}, err
	}
	return module.surfaceEvent(event, observer, after.Location())
}

func (module *Phenomena) nextStarSurface(star StarCoordinate, observer BodyObserver, kind SurfaceEventKind, after time.Time) (SurfaceEvent, error) {
	event, err := basic.NextStarSurfaceEventWithInstance(star, observer, kind, toTT(after), SurfaceOptions{}, module.inst)
	if err != nil {
		return SurfaceEvent{}, err
	}
	return module.surfaceEvent(event, observer, after.Location())
}

// SurfacePositionAt 行星表面的无折射站心视位置，不支持月球 / airless apparent position on a planetary surface, excluding the Moon.
func (module *Phenomena) SurfacePositionAt(target Body, observer BodyObserver, date time.Time) (SurfacePosition, error) {
	return basic.SurfacePositionAtWithInstance(target, observer, toTT(date), module.inst)
}

// StarSurfacePositionAt 行星表面的恒星无折射站心视位置 / airless apparent stellar position on a planetary surface.
func (module *Phenomena) StarSurfacePositionAt(star StarCoordinate, observer BodyObserver, date time.Time) (SurfacePosition, error) {
	return basic.StarSurfacePositionAtWithInstance(star, observer, toTT(date), module.inst)
}

// SurfaceLocalTimeAt 行星表面的恒星时与太阳时 / sidereal and solar times on a planetary surface.
func (module *Phenomena) SurfaceLocalTimeAt(observer BodyObserver, date time.Time) (SurfaceLocalTime, error) {
	return basic.SurfaceLocalTimeAtWithInstance(observer, toTT(date), module.inst)
}

// SurfaceEvents 最多十年区间内全部无折射表面事件与穿越状态 / all airless surface events and crossing state in an interval of at most ten years.
func (module *Phenomena) SurfaceEvents(target Body, observer BodyObserver, start, end time.Time, options SurfaceOptions) ([]SurfaceEvent, AltitudeState, error) {
	list, state, err := basic.SurfaceEventsWithInstance(target, observer, toTT(start), toTT(end), options, module.inst)
	if err != nil {
		return nil, state, err
	}
	result, err := module.surfaceEventList(list, observer, start.Location())
	return result, state, err
}

// StarSurfaceEvents 最多十年区间内全部恒星表面事件与穿越状态 / all stellar surface events and crossing state in an interval of at most ten years.
func (module *Phenomena) StarSurfaceEvents(star StarCoordinate, observer BodyObserver, start, end time.Time) ([]SurfaceEvent, AltitudeState, error) {
	list, state, err := basic.StarSurfaceEventsWithInstance(star, observer, toTT(start), toTT(end), SurfaceOptions{}, module.inst)
	if err != nil {
		return nil, state, err
	}
	result, err := module.surfaceEventList(list, observer, start.Location())
	return result, state, err
}

// NextSurfaceRise 下一次无折射目标升起，包含起点 / next airless target rise including the starting instant.
func (module *Phenomena) NextSurfaceRise(target Body, observer BodyObserver, after time.Time, options SurfaceOptions) (SurfaceEvent, error) {
	return module.nextSurface(target, observer, SurfaceRise, after, options)
}

// NextSurfaceSet 下一次无折射目标落下，包含起点 / next airless target set including the starting instant.
func (module *Phenomena) NextSurfaceSet(target Body, observer BodyObserver, after time.Time, options SurfaceOptions) (SurfaceEvent, error) {
	return module.nextSurface(target, observer, SurfaceSet, after, options)
}

// NextSurfaceCulmination 下一次无折射目标上中天，包含起点 / next airless target upper culmination including the starting instant.
func (module *Phenomena) NextSurfaceCulmination(target Body, observer BodyObserver, after time.Time, options SurfaceOptions) (SurfaceEvent, error) {
	return module.nextSurface(target, observer, SurfaceUpperCulmination, after, options)
}

// NextStarSurfaceRise 下一次无折射恒星升起，包含起点 / next airless stellar rise including the starting instant.
func (module *Phenomena) NextStarSurfaceRise(star StarCoordinate, observer BodyObserver, after time.Time) (SurfaceEvent, error) {
	return module.nextStarSurface(star, observer, SurfaceRise, after)
}

// NextStarSurfaceSet 下一次无折射恒星落下，包含起点 / next airless stellar set including the starting instant.
func (module *Phenomena) NextStarSurfaceSet(star StarCoordinate, observer BodyObserver, after time.Time) (SurfaceEvent, error) {
	return module.nextStarSurface(star, observer, SurfaceSet, after)
}

// NextStarSurfaceCulmination 下一次无折射恒星上中天，包含起点 / next airless stellar upper culmination including the starting instant.
func (module *Phenomena) NextStarSurfaceCulmination(star StarCoordinate, observer BodyObserver, after time.Time) (SurfaceEvent, error) {
	return module.nextStarSurface(star, observer, SurfaceUpperCulmination, after)
}
