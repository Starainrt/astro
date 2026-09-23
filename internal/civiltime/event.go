package civiltime

import "time"

// Event 合并变长民用日的固定偏移候选，handled=false 表示普通日期 / merges fixed-offset candidates on a civil day with an offset change.
func Event(date time.Time, missing error, calculate func(time.Time) (time.Time, error)) (result time.Time, err error, handled bool) {
	year, month, day := date.Date()
	start := time.Date(year, month, day, 0, 0, 0, 0, date.Location())
	end := time.Date(year, month, day+1, 0, 0, 0, 0, date.Location())
	_, firstOffset := start.Zone()
	_, lastOffset := end.Add(-time.Nanosecond).Zone()
	if firstOffset == lastOffset {
		return time.Time{}, nil, false
	}
	var firstErr error
	for _, offset := range []int{firstOffset, lastOffset} {
		fixed := time.Date(year, month, day, 0, 0, 0, 0, time.FixedZone("", offset))
		candidate, candidateErr := calculate(fixed)
		if candidateErr != nil {
			if firstErr == nil {
				firstErr = candidateErr
			}
			continue
		}
		if !candidate.Before(start) && candidate.Before(end) && (result.IsZero() || candidate.Before(result)) {
			result = candidate.In(date.Location())
		}
	}
	if !result.IsZero() {
		return result, nil, true
	}
	if firstErr != nil {
		return time.Time{}, firstErr, true
	}
	return time.Time{}, missing, true
}
