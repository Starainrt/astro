// 被测契约：持有 nil 指针的 Engine 与裸 nil 等价，构造器必须拦下而不是留到调用时 panic。
package ephem

import "testing"

type nilCheckEngine struct{ offset float64 }

func (engine *nilCheckEngine) State(target, center Body, jdTDB float64) (pos, vel [3]float64, err error) {
	pos[0] = engine.offset
	return pos, vel, nil
}

func (engine *nilCheckEngine) Coverage(target Body) (start, end float64, ok bool) {
	return 0, engine.offset, true
}

func TestConstructorsRejectTypedNilEngine(t *testing.T) {
	var typedNil *nilCheckEngine
	cases := map[string]func() error{
		"NewInstance": func() error {
			_, _, err := NewInstance(typedNil).State(Mars, Sun, 2451545)
			return err
		},
		"NewStack": func() error {
			_, _, err := NewInstance(NewStack(typedNil)).State(Mars, Sun, 2451545)
			return err
		},
	}
	for name, call := range cases {
		if err := call(); err == nil {
			t.Errorf("%s: err = nil, want an error", name)
		}
	}
}

func TestInstanceKeepsUsableEngine(t *testing.T) {
	pos, _, err := NewInstance(&nilCheckEngine{offset: 7}).State(Mars, Sun, 2451545)
	if err != nil {
		t.Fatalf("State err = %v", err)
	}
	if pos[0] != 7 {
		t.Errorf("pos[0] = %v, want 7", pos[0])
	}
}
