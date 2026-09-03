package main

import (
	"math"
	"testing"
)

func TestTouchDownSelectsWithoutActivation(t *testing.T) {
	geometry, width, height := testTouchGeometry()
	var touch TouchInput
	want := KeyPosition{Row: 0, Col: 0}

	activate, changed := touch.Handle(touchEventAt(t, TouchDown, 10, 20, want, geometry, width, height), 10, width, height, geometry)
	if activate != nil {
		t.Fatalf("TouchDown activated %v", *activate)
	}
	if !changed {
		t.Error("TouchDown did not report selection change")
	}
	if got, ok := touch.Selected(); !ok || got != want {
		t.Errorf("Selected() = %v, %v; want %v, true", got, ok, want)
	}
}

func TestTouchSlideAndReleaseActivatesFinalKeyOnce(t *testing.T) {
	geometry, width, height := testTouchGeometry()
	var touch TouchInput
	first := KeyPosition{Row: 0, Col: 0}
	final := KeyPosition{Row: 0, Col: 1}
	touch.Handle(touchEventAt(t, TouchDown, 10, 20, first, geometry, width, height), 10, width, height, geometry)

	activate, changed := touch.Handle(touchEventAt(t, TouchMotion, 10, 20, final, geometry, width, height), 10, width, height, geometry)
	if activate != nil || !changed {
		t.Fatalf("TouchMotion = %v, %v; want nil, true", activate, changed)
	}
	if got, ok := touch.Selected(); !ok || got != final {
		t.Errorf("selection after slide = %v, %v; want %v, true", got, ok, final)
	}

	activate, changed = touch.Handle(touchEventAt(t, TouchUp, 10, 20, final, geometry, width, height), 10, width, height, geometry)
	if activate == nil || *activate != final || !changed {
		t.Fatalf("TouchUp = %v, %v; want %v, true", activate, changed, final)
	}
	if _, ok := touch.Selected(); ok {
		t.Error("TouchUp retained selection")
	}
	activate, changed = touch.Handle(touchEventAt(t, TouchUp, 10, 20, final, geometry, width, height), 10, width, height, geometry)
	if activate != nil || changed {
		t.Errorf("second TouchUp = %v, %v; want nil, false", activate, changed)
	}
}

func TestTouchSlideOutsideClearsSelectionButRetainsOwnership(t *testing.T) {
	geometry, width, height := testTouchGeometry()
	var touch TouchInput
	first := KeyPosition{Row: 0, Col: 0}
	second := KeyPosition{Row: 0, Col: 1}
	touch.Handle(touchEventAt(t, TouchDown, 10, 20, first, geometry, width, height), 10, width, height, geometry)

	outside := TouchEvent{Phase: TouchMotion, WindowID: 10, TouchID: 20, FingerID: 30, X: -0.1, Y: 0.5}
	activate, changed := touch.Handle(outside, 10, width, height, geometry)
	if activate != nil || !changed {
		t.Fatalf("outside motion = %v, %v; want nil, true", activate, changed)
	}
	if _, ok := touch.Selected(); ok {
		t.Error("outside motion retained selection")
	}

	secondary := touchEventAt(t, TouchDown, 10, 21, second, geometry, width, height)
	secondary.FingerID = 31
	if activate, changed = touch.Handle(secondary, 10, width, height, geometry); activate != nil || changed {
		t.Errorf("secondary down = %v, %v; want ignored", activate, changed)
	}
	if activate, changed = touch.Handle(touchEventAt(t, TouchMotion, 10, 20, second, geometry, width, height), 10, width, height, geometry); activate != nil || !changed {
		t.Fatalf("owner slide back = %v, %v; want nil, true", activate, changed)
	}
	if got, ok := touch.Selected(); !ok || got != second {
		t.Errorf("selection after slide back = %v, %v; want %v, true", got, ok, second)
	}
}

func TestTouchReleaseOutsideNeverActivates(t *testing.T) {
	geometry, width, height := testTouchGeometry()
	var touch TouchInput
	first := KeyPosition{Row: 0, Col: 0}
	touch.Handle(touchEventAt(t, TouchDown, 10, 20, first, geometry, width, height), 10, width, height, geometry)

	activate, changed := touch.Handle(TouchEvent{
		Phase: TouchUp, WindowID: 10, TouchID: 20, FingerID: 30, X: 0.5, Y: 1,
	}, 10, width, height, geometry)
	if activate != nil || !changed {
		t.Errorf("outside TouchUp = %v, %v; want nil, true", activate, changed)
	}
	if touch.owned {
		t.Error("outside TouchUp retained ownership")
	}
}

func TestTouchReleaseDoesNotActivateUnselectedKey(t *testing.T) {
	geometry, width, height := testTouchGeometry()
	first := KeyPosition{Row: 0, Col: 0}
	second := KeyPosition{Row: 0, Col: 1}

	tests := []struct {
		name       string
		clearFirst bool
	}{
		{name: "release moved without motion"},
		{name: "release after selection cleared", clearFirst: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var touch TouchInput
			touch.Handle(touchEventAt(t, TouchDown, 10, 20, first, geometry, width, height), 10, width, height, geometry)
			if tt.clearFirst {
				touch.Handle(TouchEvent{
					Phase: TouchMotion, WindowID: 10, TouchID: 20, FingerID: 30, X: -0.1, Y: 0.5,
				}, 10, width, height, geometry)
			}

			activate, changed := touch.Handle(touchEventAt(t, TouchUp, 10, 20, second, geometry, width, height), 10, width, height, geometry)
			if activate != nil || !changed {
				t.Errorf("TouchUp on unselected key = %v, %v; want nil, true", activate, changed)
			}
		})
	}
}

func TestTouchInvalidCoordinatesClearOwnerWithoutActivationOnRelease(t *testing.T) {
	geometry, width, height := testTouchGeometry()
	first := KeyPosition{Row: 0, Col: 0}
	tests := []struct {
		name string
		x    float32
		y    float32
	}{
		{name: "negative", x: -0.1, y: 0.5},
		{name: "one", x: 1, y: 0.5},
		{name: "nan", x: float32(math.NaN()), y: 0.5},
		{name: "infinity", x: 0.5, y: float32(math.Inf(1))},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var touch TouchInput
			touch.Handle(touchEventAt(t, TouchDown, 10, 20, first, geometry, width, height), 10, width, height, geometry)
			activate, changed := touch.Handle(TouchEvent{
				Phase: TouchUp, WindowID: 10, TouchID: 20, FingerID: 30, X: tt.x, Y: tt.y,
			}, 10, width, height, geometry)
			if activate != nil || !changed || touch.owned {
				t.Errorf("invalid TouchUp = %v, %v, owned %v; want nil, true, false", activate, changed, touch.owned)
			}
		})
	}
}

func TestTouchIgnoresWrongWindowAndNonOwnerEvents(t *testing.T) {
	geometry, width, height := testTouchGeometry()
	var touch TouchInput
	first := KeyPosition{Row: 0, Col: 0}
	second := KeyPosition{Row: 0, Col: 1}
	touch.Handle(touchEventAt(t, TouchDown, 10, 20, first, geometry, width, height), 10, width, height, geometry)

	tests := []TouchEvent{
		touchEventAt(t, TouchMotion, 11, 20, second, geometry, width, height),
		touchEventAt(t, TouchMotion, 10, 21, second, geometry, width, height),
		touchEventAt(t, TouchUp, 10, 20, second, geometry, width, height),
		{Phase: TouchCanceled, WindowID: 10, TouchID: 20, FingerID: 31},
	}
	tests[2].FingerID = 31

	for _, event := range tests {
		if activate, changed := touch.Handle(event, 10, width, height, geometry); activate != nil || changed {
			t.Errorf("Handle(%#v) = %v, %v; want ignored", event, activate, changed)
		}
		if got, ok := touch.Selected(); !ok || got != first {
			t.Fatalf("non-owner event changed selection to %v, %v", got, ok)
		}
	}
}

func TestTouchDownOutsideDoesNotAcquireOwnership(t *testing.T) {
	geometry, width, height := testTouchGeometry()
	var touch TouchInput

	activate, changed := touch.Handle(TouchEvent{
		Phase: TouchDown, WindowID: 10, TouchID: 20, FingerID: 30, X: 0, Y: 0,
	}, 10, width, height, geometry)
	if activate != nil || changed || touch.owned {
		t.Fatalf("outside TouchDown = %v, %v, owned %v; want ignored", activate, changed, touch.owned)
	}

	valid := KeyPosition{Row: 0, Col: 0}
	event := touchEventAt(t, TouchDown, 10, 21, valid, geometry, width, height)
	if activate, changed = touch.Handle(event, 10, width, height, geometry); activate != nil || !changed || !touch.owned {
		t.Errorf("valid TouchDown after outside = %v, %v, owned %v", activate, changed, touch.owned)
	}
}

func TestTouchCancellationResetsOwnership(t *testing.T) {
	geometry, width, height := testTouchGeometry()
	first := KeyPosition{Row: 0, Col: 0}

	t.Run("event", func(t *testing.T) {
		var touch TouchInput
		touch.Handle(touchEventAt(t, TouchDown, 10, 20, first, geometry, width, height), 10, width, height, geometry)
		activate, changed := touch.Handle(TouchEvent{
			Phase: TouchCanceled, WindowID: 10, TouchID: 20, FingerID: 30,
		}, 10, width, height, geometry)
		if activate != nil || !changed || touch.owned {
			t.Errorf("TouchCanceled = %v, %v, owned %v", activate, changed, touch.owned)
		}
	})

	t.Run("lifecycle cancel", func(t *testing.T) {
		var touch TouchInput
		touch.Handle(touchEventAt(t, TouchDown, 10, 20, first, geometry, width, height), 10, width, height, geometry)
		if !touch.Cancel() || touch.owned {
			t.Error("Cancel did not reset active ownership")
		}
		if touch.Cancel() {
			t.Error("Cancel reported change without ownership")
		}
	})
}

func TestActionCloseCancelsTouchOwnership(t *testing.T) {
	geometry, width, height := testTouchGeometry()
	first := KeyPosition{Row: 0, Col: 0}

	for _, daemon := range []bool{false, true} {
		t.Run(map[bool]string{false: "exit", true: "daemon hide"}[daemon], func(t *testing.T) {
			app := NewApp(Config{})
			app.daemon = daemon
			app.running = true
			app.touch.Handle(touchEventAt(t, TouchDown, 10, 20, first, geometry, width, height), 10, width, height, geometry)

			app.handleAction(Action{Type: ActionClose}, NewKeyboardState(LayoutQWERTY), nil, &Renderer{})
			if app.touch.owned {
				t.Error("ActionClose retained touch ownership")
			}
			if daemon && !app.togglePending {
				t.Error("daemon ActionClose did not request hide")
			}
			if !daemon && app.running {
				t.Error("non-daemon ActionClose did not stop app")
			}
		})
	}
}

func TestTouchActivationReusesKeyboardModifierSemantics(t *testing.T) {
	geometry := NewKeyboardGeometry(LayoutQWERTY, 40, 2, 20)
	width, height := geometry.WindowSize()
	shift := findKeyPosition(t, func(key KeyDef) bool { return key.ModifierType == "shift" })
	q := findKeyPosition(t, func(key KeyDef) bool { return key.Code == KEY_Q })
	kb := NewKeyboardState(LayoutQWERTY)
	startRow, startCol := kb.CursorRow, kb.CursorCol
	var touch TouchInput

	for sequence, pos := range []KeyPosition{shift, q} {
		touchID := uint64(sequence + 1)
		touch.Handle(touchEventAt(t, TouchDown, 10, touchID, pos, geometry, width, height), 10, width, height, geometry)
		activate, _ := touch.Handle(touchEventAt(t, TouchUp, 10, touchID, pos, geometry, width, height), 10, width, height, geometry)
		if activate == nil || !kb.PressAt(*activate, nil) {
			t.Fatalf("touch sequence %d did not activate %v", sequence, pos)
		}
		if sequence == 0 && !kb.ShiftActive {
			t.Error("touching Shift did not activate one-shot shift")
		}
	}
	if kb.ShiftActive {
		t.Error("touching Q did not consume one-shot shift")
	}
	if kb.CursorRow != startRow || kb.CursorCol != startCol {
		t.Errorf("touch activation moved gamepad cursor to (%d, %d)", kb.CursorRow, kb.CursorCol)
	}
}

func TestTouchPhaseFromSDLEvent(t *testing.T) {
	tests := []struct {
		eventType uint32
		phase     TouchPhase
		ok        bool
	}{
		{eventType: SDL_EVENT_FINGER_DOWN, phase: TouchDown, ok: true},
		{eventType: SDL_EVENT_FINGER_MOTION, phase: TouchMotion, ok: true},
		{eventType: SDL_EVENT_FINGER_UP, phase: TouchUp, ok: true},
		{eventType: SDL_EVENT_FINGER_CANCELED, phase: TouchCanceled, ok: true},
		{eventType: SDL_EVENT_QUIT, ok: false},
	}
	for _, tt := range tests {
		phase, ok := touchPhaseFromSDLEvent(tt.eventType)
		if phase != tt.phase || ok != tt.ok {
			t.Errorf("touchPhaseFromSDLEvent(%#x) = %v, %v; want %v, %v", tt.eventType, phase, ok, tt.phase, tt.ok)
		}
	}
}

func testTouchGeometry() (KeyboardGeometry, int32, int32) {
	geometry := NewKeyboardGeometry([][]KeyDef{{{Width: 1}, {Width: 1}}}, 20, 2, 6)
	width, height := geometry.WindowSize()
	return geometry, width, height
}

func touchEventAt(
	t *testing.T,
	phase TouchPhase,
	windowID uint32,
	touchID uint64,
	pos KeyPosition,
	geometry KeyboardGeometry,
	width, height int32,
) TouchEvent {
	t.Helper()
	rect, ok := geometry.KeyRect(pos)
	if !ok {
		t.Fatalf("KeyRect(%v) returned false", pos)
	}
	return TouchEvent{
		Phase:    phase,
		WindowID: windowID,
		TouchID:  touchID,
		FingerID: 30,
		X:        (rect.X + rect.W/2) / float32(width),
		Y:        (rect.Y + rect.H/2) / float32(height),
	}
}
