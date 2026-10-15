package ephem

import "testing"

func TestGlobalDefaultInstanceIsStable(t *testing.T) {
	model, generation := Active(), Generation()
	first := Global()
	if first == nil || first.Model() != model {
		t.Fatalf("default instance = %v, active model = %v", first, model)
	}
	if second := Global(); second != first {
		t.Error("lazy default must retain its instance identity")
	}
	if Active() != model || Generation() != generation {
		t.Error("reading the default changed the active model or generation")
	}
}

func TestGlobalExplicitInstanceIsStable(t *testing.T) {
	previous := Global()
	t.Cleanup(func() { SetGlobal(previous) })
	custom := NewInstance(testEngine{marker: 7})
	SetGlobal(custom)
	if Global() != custom || Global() != custom {
		t.Error("explicit default lost its instance identity")
	}
	SetGlobal(nil)
	if Global() != custom {
		t.Error("nil must not replace an explicit default")
	}
}
