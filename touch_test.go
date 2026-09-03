package main

import (
	"math"
	"testing"
)

func TestTouchDownSelectsWithoutActivation(t *testing.T) {
	geometry, width, height := testTouchGeometry()
	var touch TouchInput
	want := KeyPosition{Row: 0, Col: 0}

	activate, changed := touch.Handle(touchEventAt(t, TouchDown, 10, 20, 30, want, geometry, width, height), 10, width, height, geometry)
	if activate != nil || !changed {
		t.Fatalf("TouchDown = %v, %v; want nil, true", activate, changed)
	}
	if !touch.IsSelected(want) || len(touch.contacts) != 1 {
		t.Errorf("TouchDown contacts = %#v; want %v selected", touch.contacts, want)
	}
}

func TestTouchContactsActivateIndependentlyInReleaseOrder(t *testing.T) {
	geometry, width, height := testTouchGeometry()
	var touch TouchInput
	first := KeyPosition{Row: 0, Col: 0}
	second := KeyPosition{Row: 0, Col: 1}
	touch.Handle(touchEventAt(t, TouchDown, 10, 20, 30, first, geometry, width, height), 10, width, height, geometry)
	touch.Handle(touchEventAt(t, TouchDown, 10, 20, 31, second, geometry, width, height), 10, width, height, geometry)

	var activated []KeyPosition
	for _, event := range []TouchEvent{
		touchEventAt(t, TouchUp, 10, 20, 31, second, geometry, width, height),
		touchEventAt(t, TouchUp, 10, 20, 30, first, geometry, width, height),
	} {
		activate, changed := touch.Handle(event, 10, width, height, geometry)
		if activate == nil || !changed {
			t.Fatalf("TouchUp = %v, %v; want activation", activate, changed)
		}
		activated = append(activated, *activate)
	}
	if len(activated) != 2 || activated[0] != second || activated[1] != first {
		t.Errorf("activation order = %v; want [%v %v]", activated, second, first)
	}
	if len(touch.contacts) != 0 {
		t.Errorf("TouchUp retained contacts: %#v", touch.contacts)
	}
}

func TestTouchSlideAndReleaseActivatesFinalKey(t *testing.T) {
	geometry, width, height := testTouchGeometry()
	var touch TouchInput
	first := KeyPosition{Row: 0, Col: 0}
	final := KeyPosition{Row: 0, Col: 1}
	touch.Handle(touchEventAt(t, TouchDown, 10, 20, 30, first, geometry, width, height), 10, width, height, geometry)

	activate, changed := touch.Handle(touchEventAt(t, TouchMotion, 10, 20, 30, final, geometry, width, height), 10, width, height, geometry)
	if activate != nil || !changed || touch.IsSelected(first) || !touch.IsSelected(final) {
		t.Fatalf("TouchMotion = %v, %v, contacts %#v; want final key selected", activate, changed, touch.contacts)
	}
	activate, changed = touch.Handle(touchEventAt(t, TouchUp, 10, 20, 30, final, geometry, width, height), 10, width, height, geometry)
	if activate == nil || *activate != final || !changed {
		t.Fatalf("TouchUp = %v, %v; want %v, true", activate, changed, final)
	}
	if activate, changed = touch.Handle(touchEventAt(t, TouchUp, 10, 20, 30, final, geometry, width, height), 10, width, height, geometry); activate != nil || changed {
		t.Errorf("duplicate TouchUp = %v, %v; want ignored", activate, changed)
	}
}

func TestTouchMotionAndCancellationAffectOnlyMatchingContact(t *testing.T) {
	geometry, width, height := testTouchGeometry()
	var touch TouchInput
	first := KeyPosition{Row: 0, Col: 0}
	second := KeyPosition{Row: 0, Col: 1}
	touch.Handle(touchEventAt(t, TouchDown, 10, 20, 30, first, geometry, width, height), 10, width, height, geometry)
	touch.Handle(touchEventAt(t, TouchDown, 10, 20, 31, second, geometry, width, height), 10, width, height, geometry)

	outside := TouchEvent{Phase: TouchMotion, WindowID: 10, TouchID: 20, FingerID: 30, X: -0.1, Y: 0.5}
	if activate, changed := touch.Handle(outside, 10, width, height, geometry); activate != nil || !changed {
		t.Fatalf("outside motion = %v, %v; want nil, true", activate, changed)
	}
	if touch.IsSelected(first) || !touch.IsSelected(second) || len(touch.contacts) != 2 {
		t.Errorf("outside motion changed wrong contact: %#v", touch.contacts)
	}

	if activate, changed := touch.Handle(touchEventAt(t, TouchMotion, 10, 20, 30, first, geometry, width, height), 10, width, height, geometry); activate != nil || !changed {
		t.Fatalf("recovery motion = %v, %v; want nil, true", activate, changed)
	}
	if !touch.IsSelected(first) || !touch.IsSelected(second) {
		t.Errorf("recovery did not restore both selections: %#v", touch.contacts)
	}

	if activate, changed := touch.Handle(TouchEvent{Phase: TouchCanceled, WindowID: 10, TouchID: 20, FingerID: 30}, 10, width, height, geometry); activate != nil || !changed {
		t.Fatalf("TouchCanceled = %v, %v; want nil, true", activate, changed)
	}
	if touch.IsSelected(first) || !touch.IsSelected(second) || len(touch.contacts) != 1 {
		t.Errorf("TouchCanceled changed wrong contact: %#v", touch.contacts)
	}
}

func TestTouchSharedKeyRemainsSelectedUntilLastContactLeaves(t *testing.T) {
	geometry, width, height := testTouchGeometry()
	var touch TouchInput
	shared := KeyPosition{Row: 0, Col: 0}
	for _, fingerID := range []int64{30, 31} {
		touch.Handle(touchEventAt(t, TouchDown, 10, 20, fingerID, shared, geometry, width, height), 10, width, height, geometry)
	}

	for index, fingerID := range []int64{30, 31} {
		activate, changed := touch.Handle(touchEventAt(t, TouchUp, 10, 20, fingerID, shared, geometry, width, height), 10, width, height, geometry)
		if activate == nil || *activate != shared || !changed {
			t.Fatalf("contact %d TouchUp = %v, %v; want %v, true", index, activate, changed, shared)
		}
		wantSelected := index == 0
		if touch.IsSelected(shared) != wantSelected {
			t.Errorf("selected after contact %d release = %v; want %v", index, touch.IsSelected(shared), wantSelected)
		}
	}
}

func TestTouchCompositeIdentityDistinguishesDevices(t *testing.T) {
	geometry, width, height := testTouchGeometry()
	var touch TouchInput
	first := KeyPosition{Row: 0, Col: 0}
	second := KeyPosition{Row: 0, Col: 1}
	touch.Handle(touchEventAt(t, TouchDown, 10, 20, 30, first, geometry, width, height), 10, width, height, geometry)
	touch.Handle(touchEventAt(t, TouchDown, 10, 21, 30, second, geometry, width, height), 10, width, height, geometry)

	activate, changed := touch.Handle(touchEventAt(t, TouchUp, 10, 20, 30, first, geometry, width, height), 10, width, height, geometry)
	if activate == nil || *activate != first || !changed {
		t.Fatalf("first device TouchUp = %v, %v; want %v, true", activate, changed, first)
	}
	if !touch.IsSelected(second) || len(touch.contacts) != 1 {
		t.Errorf("first device release removed second device: %#v", touch.contacts)
	}
}

func TestTouchDuplicateDownDoesNotReplaceSequence(t *testing.T) {
	geometry, width, height := testTouchGeometry()
	var touch TouchInput
	first := KeyPosition{Row: 0, Col: 0}
	second := KeyPosition{Row: 0, Col: 1}
	touch.Handle(touchEventAt(t, TouchDown, 10, 20, 30, first, geometry, width, height), 10, width, height, geometry)

	activate, changed := touch.Handle(touchEventAt(t, TouchDown, 10, 20, 30, second, geometry, width, height), 10, width, height, geometry)
	if activate != nil || changed {
		t.Fatalf("duplicate TouchDown = %v, %v; want ignored", activate, changed)
	}
	if !touch.IsSelected(first) || touch.IsSelected(second) || len(touch.contacts) != 1 {
		t.Errorf("duplicate TouchDown replaced sequence: %#v", touch.contacts)
	}
}

func TestTouchUnknownAndWrongWindowEventsAreIgnored(t *testing.T) {
	geometry, width, height := testTouchGeometry()
	var touch TouchInput
	selected := KeyPosition{Row: 0, Col: 0}
	other := KeyPosition{Row: 0, Col: 1}
	touch.Handle(touchEventAt(t, TouchDown, 10, 20, 30, selected, geometry, width, height), 10, width, height, geometry)

	tests := []TouchEvent{
		touchEventAt(t, TouchMotion, 11, 20, 30, other, geometry, width, height),
		touchEventAt(t, TouchMotion, 10, 20, 31, other, geometry, width, height),
		touchEventAt(t, TouchUp, 10, 21, 30, selected, geometry, width, height),
		{Phase: TouchCanceled, WindowID: 10, TouchID: 20, FingerID: 31},
	}
	for _, event := range tests {
		if activate, changed := touch.Handle(event, 10, width, height, geometry); activate != nil || changed {
			t.Errorf("Handle(%#v) = %v, %v; want ignored", event, activate, changed)
		}
		if !touch.IsSelected(selected) || len(touch.contacts) != 1 {
			t.Fatalf("ignored event changed contact: %#v", touch.contacts)
		}
	}
}

func TestTouchDownOutsideDoesNotCreateRecoverableContact(t *testing.T) {
	geometry, width, height := testTouchGeometry()
	var touch TouchInput
	outside := TouchEvent{Phase: TouchDown, WindowID: 10, TouchID: 20, FingerID: 30, X: 0, Y: 0}
	if activate, changed := touch.Handle(outside, 10, width, height, geometry); activate != nil || changed {
		t.Fatalf("outside TouchDown = %v, %v; want ignored", activate, changed)
	}
	valid := KeyPosition{Row: 0, Col: 0}
	if activate, changed := touch.Handle(touchEventAt(t, TouchMotion, 10, 20, 30, valid, geometry, width, height), 10, width, height, geometry); activate != nil || changed {
		t.Errorf("motion after outside down = %v, %v; want ignored", activate, changed)
	}
	if len(touch.contacts) != 0 || touch.IsSelected(valid) {
		t.Errorf("outside sequence acquired contact: %#v", touch.contacts)
	}
}

func TestTouchReleaseRequiresFinalStoredSelectionMatch(t *testing.T) {
	geometry, width, height := testTouchGeometry()
	first := KeyPosition{Row: 0, Col: 0}
	second := KeyPosition{Row: 0, Col: 1}
	tests := []struct {
		name           string
		clearSelection bool
		up             TouchEvent
	}{
		{name: "different key without motion", up: touchEventAt(t, TouchUp, 10, 20, 30, second, geometry, width, height)},
		{name: "outside", up: TouchEvent{Phase: TouchUp, WindowID: 10, TouchID: 20, FingerID: 30, X: 0.5, Y: 1}},
		{name: "malformed", up: TouchEvent{Phase: TouchUp, WindowID: 10, TouchID: 20, FingerID: 30, X: float32(math.NaN()), Y: 0.5}},
		{name: "cleared selection", clearSelection: true, up: touchEventAt(t, TouchUp, 10, 20, 30, first, geometry, width, height)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var touch TouchInput
			touch.Handle(touchEventAt(t, TouchDown, 10, 20, 30, first, geometry, width, height), 10, width, height, geometry)
			if tt.clearSelection {
				touch.Handle(TouchEvent{Phase: TouchMotion, WindowID: 10, TouchID: 20, FingerID: 30, X: 0, Y: 0}, 10, width, height, geometry)
			}
			activate, changed := touch.Handle(tt.up, 10, width, height, geometry)
			if activate != nil || !changed || len(touch.contacts) != 0 {
				t.Errorf("TouchUp = %v, %v, contacts %#v; want nil, true, empty", activate, changed, touch.contacts)
			}
		})
	}
}

func TestTouchCancelClearsAllContacts(t *testing.T) {
	geometry, width, height := testTouchGeometry()
	var touch TouchInput
	for index, pos := range []KeyPosition{{Row: 0, Col: 0}, {Row: 0, Col: 1}} {
		touch.Handle(touchEventAt(t, TouchDown, 10, 20, int64(30+index), pos, geometry, width, height), 10, width, height, geometry)
	}
	if !touch.Cancel() || len(touch.contacts) != 0 {
		t.Errorf("Cancel contacts = %#v; want empty", touch.contacts)
	}
	if touch.Cancel() {
		t.Error("Cancel reported change without contacts")
	}
}

func TestActionCloseCancelsAllTouchContacts(t *testing.T) {
	geometry, width, height := testTouchGeometry()
	for _, daemon := range []bool{false, true} {
		t.Run(map[bool]string{false: "exit", true: "daemon hide"}[daemon], func(t *testing.T) {
			app := NewApp(Config{})
			app.daemon = daemon
			app.running = true
			for index, pos := range []KeyPosition{{Row: 0, Col: 0}, {Row: 0, Col: 1}} {
				app.touch.Handle(touchEventAt(t, TouchDown, 10, 20, int64(30+index), pos, geometry, width, height), 10, width, height, geometry)
			}

			app.handleAction(Action{Type: ActionClose}, NewKeyboardState(LayoutQWERTY), nil, &Renderer{})
			if len(app.touch.contacts) != 0 {
				t.Errorf("ActionClose retained contacts: %#v", app.touch.contacts)
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

func TestTouchModifierOverlapUsesReleaseOrder(t *testing.T) {
	geometry := NewKeyboardGeometry(LayoutQWERTY, 40, 2, 20)
	width, height := geometry.WindowSize()
	shift := findKeyPosition(t, func(key KeyDef) bool { return key.ModifierType == "shift" })
	q := findKeyPosition(t, func(key KeyDef) bool { return key.Code == KEY_Q })

	tests := []struct {
		name            string
		releaseOrder    []KeyPosition
		wantShiftActive bool
	}{
		{name: "modifier released first", releaseOrder: []KeyPosition{shift, q}, wantShiftActive: false},
		{name: "letter released first", releaseOrder: []KeyPosition{q, shift}, wantShiftActive: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			kb := NewKeyboardState(LayoutQWERTY)
			startRow, startCol := kb.CursorRow, kb.CursorCol
			var touch TouchInput
			contacts := map[KeyPosition]int64{shift: 30, q: 31}
			for _, pos := range []KeyPosition{shift, q} {
				touch.Handle(touchEventAt(t, TouchDown, 10, 20, contacts[pos], pos, geometry, width, height), 10, width, height, geometry)
			}
			for _, pos := range tt.releaseOrder {
				activate, _ := touch.Handle(touchEventAt(t, TouchUp, 10, 20, contacts[pos], pos, geometry, width, height), 10, width, height, geometry)
				if activate == nil || !kb.PressAt(*activate, nil) {
					t.Fatalf("TouchUp did not activate %v", pos)
				}
			}
			if kb.ShiftActive != tt.wantShiftActive {
				t.Errorf("ShiftActive = %v; want %v", kb.ShiftActive, tt.wantShiftActive)
			}
			if kb.CursorRow != startRow || kb.CursorCol != startCol {
				t.Errorf("touch activation moved cursor to (%d, %d)", kb.CursorRow, kb.CursorCol)
			}
		})
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
	fingerID int64,
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
		FingerID: fingerID,
		X:        (rect.X + rect.W/2) / float32(width),
		Y:        (rect.Y + rect.H/2) / float32(height),
	}
}
