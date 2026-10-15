package ephem

import (
	"errors"
	"testing"
)

type stackTestEngine struct {
	bodies []Body
	pairs  map[[2]Body][3]float64
	start  float64
	end    float64
}

func (engine stackTestEngine) State(target, center Body, jdTDB float64) (pos, vel [3]float64, err error) {
	if value, ok := engine.pairs[[2]Body{target, center}]; ok && jdTDB >= engine.start && jdTDB <= engine.end {
		pos = value
		return pos, vel, nil
	}
	for _, body := range engine.bodies {
		if body == target || body == center {
			return pos, vel, ErrOutOfRange
		}
	}
	return pos, vel, ErrUnsupportedBody
}

func (engine stackTestEngine) Coverage(target Body) (start, end float64, ok bool) {
	for _, body := range engine.bodies {
		if body == target {
			return engine.start, engine.end, true
		}
	}
	return 0, 0, false
}

func (engine stackTestEngine) Bodies() []Body { return append([]Body(nil), engine.bodies...) }

func TestStackComposesAcrossSources(t *testing.T) {
	first := stackTestEngine{
		bodies: []Body{Sun, SSB},
		pairs: map[[2]Body][3]float64{
			{Sun, SSB}: {10, 20, 30},
		},
		start: 0,
		end:   3000000,
	}
	second := stackTestEngine{
		bodies: []Body{Body(2000001), Sun},
		pairs: map[[2]Body][3]float64{
			{Body(2000001), Sun}: {1, 2, 3},
		},
		start: 0,
		end:   3000000,
	}
	stack := NewStack(first, second)
	pos, _, err := stack.State(Body(2000001), SSB, 2451545)
	if err != nil {
		t.Fatal(err)
	}
	if pos != [3]float64{11, 22, 33} {
		t.Fatalf("position = %v", pos)
	}
}

func TestStackFirstSourceWinsDirectState(t *testing.T) {
	first := stackTestEngine{bodies: []Body{Mars, Earth}, pairs: map[[2]Body][3]float64{{Mars, Earth}: {1, 2, 3}}, start: 0, end: 10}
	second := stackTestEngine{bodies: []Body{Mars, Earth}, pairs: map[[2]Body][3]float64{{Mars, Earth}: {4, 5, 6}}, start: 0, end: 10}
	pos, _, err := NewStack(first, second).State(Mars, Earth, 5)
	if err != nil || pos != [3]float64{1, 2, 3} {
		t.Fatalf("state = %v, %v", pos, err)
	}
}

func TestStackErrorPriority(t *testing.T) {
	engine := stackTestEngine{bodies: []Body{Mars}, start: 0, end: 1}
	_, _, err := NewStack(engine).State(Mars, Earth, 2)
	if !errors.Is(err, ErrOutOfRange) {
		t.Fatalf("error = %v", err)
	}
}

func TestStackGenericApparentAcceptsArbitraryBody(t *testing.T) {
	engine := stackTestEngine{
		bodies: []Body{Earth, Sun, Body(2000001), SSB},
		pairs: map[[2]Body][3]float64{
			{Earth, SSB}:         {0, 0, 0},
			{Sun, SSB}:           {1000000, 0, 0},
			{Body(2000001), SSB}: {1001000, 0, 0},
		},
		start: 0,
		end:   3000000,
	}
	instance := NewInstance(NewStack(engine))
	if _, distance, _, err := instance.Apparent(Body(2000001), 2451545); err != nil || distance != 1001000 {
		t.Fatalf("apparent distance = %.3f, err = %v", distance, err)
	}
}

func BenchmarkStackStateChain(b *testing.B) {
	first := stackTestEngine{bodies: []Body{Sun, SSB}, pairs: map[[2]Body][3]float64{{Sun, SSB}: {10, 20, 30}}, start: 0, end: 3000000}
	second := stackTestEngine{bodies: []Body{Body(2000001), Sun}, pairs: map[[2]Body][3]float64{{Body(2000001), Sun}: {1, 2, 3}}, start: 0, end: 3000000}
	stack := NewStack(first, second)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, _, err := stack.State(Body(2000001), SSB, 2451545); err != nil {
			b.Fatal(err)
		}
	}
}

func TestStackChainBeyondInlineCapacity(t *testing.T) {
	bodies := []Body{SSB}
	pairs := map[[2]Body][3]float64{}
	previous := SSB
	for index := 0; index < 2*stackInlineBodies; index++ {
		body := Body(3000000 + index)
		bodies = append(bodies, body)
		pairs[[2]Body{body, previous}] = [3]float64{1, 0, 0}
		previous = body
	}
	engine := stackTestEngine{bodies: bodies, pairs: pairs, start: 0, end: 10}
	pos, _, err := NewStack(engine).State(previous, SSB, 5)
	if err != nil || pos != [3]float64{2 * stackInlineBodies, 0, 0} {
		t.Fatalf("state = %v, %v", pos, err)
	}
}
