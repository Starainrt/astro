package ephem

import (
	"errors"
	"sort"
)

// BodyProvider 可选地报告引擎能够触及的天体编号 / optionally reports bodies reachable by an engine.
type BodyProvider interface {
	Bodies() []Body
}

// Stack 按顺序叠加多个只读星历引擎 / ordered stack of read-only ephemeris engines.
type Stack struct {
	engines []Engine
	// extra 为各源声明、且不在 chainBodies 中的天体，创建时算好。
	extra []Body
}

// stackInlineBodies 中心链搜索在栈上预留的候选数，超出时 append 退回堆分配。
const stackInlineBodies = 24

type stackNode struct {
	body     Body
	pos, vel [3]float64
}

// chainBodies 中心链搜索的固定候选，顺序决定同等路径的取舍。
var chainBodies = [...]Body{SSB, Sun, Mercury, Venus, EMB, Earth, Moon, Mars, Jupiter, Saturn, Uranus, Neptune}

// NewStack 创建多源星历；先传入的引擎优先 / creates a multi-source ephemeris; earlier engines have priority.
func NewStack(engines ...Engine) *Stack {
	usable := make([]Engine, 0, len(engines))
	for _, engine := range engines {
		if !IsNilEngine(engine) {
			usable = append(usable, engine)
		}
	}
	stack := &Stack{engines: usable}
	for _, body := range stack.Bodies() {
		if !containsBody(chainBodies[:], body) {
			stack.extra = append(stack.extra, body)
		}
	}
	return stack
}

// Engines 返回底层引擎快照 / returns a snapshot of the underlying engines.
func (stack *Stack) Engines() []Engine {
	if stack == nil {
		return nil
	}
	return append([]Engine(nil), stack.engines...)
}

// Bodies 返回各源声明的天体编号 / returns body identifiers declared by the sources.
func (stack *Stack) Bodies() []Body {
	if stack == nil {
		return nil
	}
	seen := make(map[Body]bool)
	result := make([]Body, 0)
	for _, engine := range stack.engines {
		provider, ok := engine.(BodyProvider)
		if !ok {
			continue
		}
		for _, body := range provider.Bodies() {
			if !seen[body] {
				seen[body] = true
				result = append(result, body)
			}
		}
	}
	sort.Slice(result, func(left, right int) bool { return result[left] < result[right] })
	return result
}

// Coverage 返回所有源覆盖区间的包络 / returns the envelope of all source coverage intervals.
func (stack *Stack) Coverage(target Body) (start, end float64, ok bool) {
	if stack == nil {
		return 0, 0, false
	}
	for _, engine := range stack.engines {
		engineStart, engineEnd, engineOK := engine.Coverage(target)
		if !engineOK {
			continue
		}
		if !ok || engineStart < start {
			start = engineStart
		}
		if !ok || engineEnd > end {
			end = engineEnd
		}
		ok = true
	}
	return start, end, ok
}

// State 返回目标相对中心的状态，并在源之间拼接中心链 / returns target-relative state, composing center chains across sources.
func (stack *Stack) State(target, center Body, jdTDB float64) (pos, vel [3]float64, err error) {
	if target == center {
		return pos, vel, nil
	}
	if stack == nil || len(stack.engines) == 0 {
		return pos, vel, ErrNoEngine
	}
	if !finiteEpoch(jdTDB) {
		return pos, vel, ErrOutOfRange
	}
	failures := &stackFailures{}
	if pos, vel, ok := stack.direct(target, center, jdTDB, failures); ok {
		return pos, vel, nil
	}

	var candidateBuf [stackInlineBodies]Body
	candidates := stack.candidates(candidateBuf[:0], target, center)
	var queueBuf [stackInlineBodies + 1]stackNode
	queue := append(queueBuf[:0], stackNode{body: target})
	var visitedBuf [stackInlineBodies]bool
	visited := visitedBuf[:0]
	for _, candidate := range candidates {
		visited = append(visited, candidate == target)
	}
	for head := 0; head < len(queue); head++ {
		current := queue[head]
		if current.body == center {
			return current.pos, current.vel, nil
		}
		for index, candidate := range candidates {
			if visited[index] || candidate == current.body {
				continue
			}
			edgePos, edgeVel, ok := stack.direct(current.body, candidate, jdTDB, failures)
			if !ok {
				continue
			}
			addVector(&edgePos, current.pos)
			addVector(&edgeVel, current.vel)
			visited[index] = true
			queue = append(queue, stackNode{body: candidate, pos: edgePos, vel: edgeVel})
		}
	}
	return pos, vel, failures.err()
}

func (stack *Stack) direct(target, center Body, jdTDB float64, failures *stackFailures) (pos, vel [3]float64, ok bool) {
	for _, engine := range stack.engines {
		start, end, covered := engine.Coverage(target)
		if !covered || jdTDB < start || jdTDB > end {
			if covered && (jdTDB < start || jdTDB > end) {
				failures.add(ErrOutOfRange)
			}
			start, end, covered = engine.Coverage(center)
			if !covered || jdTDB < start || jdTDB > end {
				if covered && (jdTDB < start || jdTDB > end) {
					failures.add(ErrOutOfRange)
				}
				continue
			}
		}
		pos, vel, err := engine.State(target, center, jdTDB)
		if err == nil {
			return pos, vel, true
		}
		failures.add(err)
	}
	return pos, vel, false
}

// candidates 按 center、固定候选、target、各源额外天体的顺序去重。
func (stack *Stack) candidates(result []Body, target, center Body) []Body {
	result = appendUniqueBody(result, center)
	for _, body := range chainBodies {
		result = appendUniqueBody(result, body)
	}
	result = appendUniqueBody(result, target)
	for _, body := range stack.extra {
		result = appendUniqueBody(result, body)
	}
	return result
}

func appendUniqueBody(bodies []Body, body Body) []Body {
	if containsBody(bodies, body) {
		return bodies
	}
	return append(bodies, body)
}

func containsBody(bodies []Body, body Body) bool {
	for _, existing := range bodies {
		if existing == body {
			return true
		}
	}
	return false
}

type stackFailures struct {
	last        error
	outOfRange  bool
	unsupported bool
}

func (failures *stackFailures) add(err error) {
	if err == nil {
		return
	}
	failures.last = err
	if errors.Is(err, ErrOutOfRange) {
		failures.outOfRange = true
	}
	if errors.Is(err, ErrUnsupportedBody) {
		failures.unsupported = true
	}
}

func (failures *stackFailures) err() error {
	if failures.outOfRange {
		return ErrOutOfRange
	}
	if failures.unsupported {
		return ErrUnsupportedBody
	}
	if failures.last != nil {
		return failures.last
	}
	return ErrUnsupportedBody
}

func finiteEpoch(value float64) bool {
	return value == value && value < 1.7976931348623157e308 && value > -1.7976931348623157e308
}

func addVector(dst *[3]float64, src [3]float64) {
	for index := range dst {
		dst[index] += src[index]
	}
}

var _ Engine = (*Stack)(nil)
var _ BodyProvider = (*Stack)(nil)
