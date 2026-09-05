package main

import (
	"math"
	"testing"
	"time"
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

func TestPointerRepeatContactsAreIndependent(t *testing.T) {
	geometry, width, height := testTouchGeometry()
	pos := KeyPosition{Row: 0, Col: 0}
	now := time.Unix(100, 0)
	firstEvent := touchEventAt(t, TouchDown, 10, 20, 30, pos, geometry, width, height)
	secondEvent := touchEventAt(t, TouchDown, 10, 20, 31, pos, geometry, width, height)
	var touch2 TouchInput
	touch2.Handle(firstEvent, 10, width, height, geometry)
	touch2.Handle(secondEvent, 10, width, height, geometry)
	if !touch2.StartRepeat(firstEvent, pos, now) {
		t.Fatal("StartRepeat returned false")
	}
	if touch2.IsRepeating(secondEvent) {
		t.Error("starting one contact armed another contact")
	}
	if !touch2.StartRepeat(secondEvent, pos, now) {
		t.Fatal("second StartRepeat returned false")
	}
	if got := touch2.RepeatableContacts(now.Add(399*time.Millisecond), 400*time.Millisecond, 80*time.Millisecond); len(got) != 0 {
		t.Fatalf("early repeats = %v; want none", got)
	}
	if got := touch2.RepeatableContacts(now.Add(400*time.Millisecond), 400*time.Millisecond, 80*time.Millisecond); len(got) != 2 {
		t.Fatalf("initial repeats = %v; want two", got)
	}
	up := touchEventAt(t, TouchUp, 10, 20, 30, pos, geometry, width, height)
	repeating := touch2.IsRepeating(up)
	activate, changed := touch2.Handle(up, 10, width, height, geometry)
	if activate == nil || !changed || !repeating {
		t.Fatal("first contact release was not suppressible")
	}
	if next, active := touch2.NextRepeat(now.Add(400*time.Millisecond), 400*time.Millisecond, 80*time.Millisecond); !active || next <= 0 {
		t.Fatalf("remaining repeat = %v, %v; want active future deadline", next, active)
	}
}

func TestPointerRepeatStopsOnMotionAndSuppressesRelease(t *testing.T) {
	geometry, width, height := testTouchGeometry()
	var touch TouchInput
	pos := KeyPosition{Row: 0, Col: 0}
	other := KeyPosition{Row: 0, Col: 1}
	now := time.Unix(100, 0)
	touch.Handle(touchEventAt(t, TouchDown, 10, 20, 30, pos, geometry, width, height), 10, width, height, geometry)
	down := touchEventAt(t, TouchDown, 10, 20, 30, pos, geometry, width, height)
	touch.StartRepeat(down, pos, now)
	if got := touch.RepeatableContacts(now.Add(400*time.Millisecond), 400*time.Millisecond, 80*time.Millisecond); len(got) != 1 {
		t.Fatalf("initial repeats = %v; want one", got)
	}
	touch.Handle(touchEventAt(t, TouchMotion, 10, 20, 30, other, geometry, width, height), 10, width, height, geometry)
	if _, active := touch.NextRepeat(now.Add(400*time.Millisecond), 400*time.Millisecond, 80*time.Millisecond); active {
		t.Error("motion onto another key kept repeat scheduled")
	}
	touch.Handle(touchEventAt(t, TouchMotion, 10, 20, 30, pos, geometry, width, height), 10, width, height, geometry)
	if _, active := touch.NextRepeat(now.Add(480*time.Millisecond), 400*time.Millisecond, 80*time.Millisecond); active {
		t.Error("motion back to backspace restarted repeat schedule")
	}
	up := touchEventAt(t, TouchUp, 10, 20, 30, pos, geometry, width, height)
	repeating := touch.IsRepeating(up)
	activate, changed := touch.Handle(up, 10, width, height, geometry)
	if activate == nil || !changed || !repeating {
		t.Fatal("release after motion was not suppressible")
	}
	if touch.Cancel() {
		t.Error("release did not clear contact")
	}
	next, active := touch.NextRepeat(now, time.Second, time.Second)
	if next != 0 || active {
		t.Error("release did not clear repeat state")
	}
}

func TestPointerRepeatStartsFromSelectedKey(t *testing.T) {
	geometry, width, height := testTouchGeometry()
	var touch TouchInput
	backspace := KeyPosition{Row: 0, Col: 1}
	event := touchEventAt(t, TouchDown, 10, 20, 30, backspace, geometry, width, height)
	touch.Handle(event, 10, width, height, geometry)

	kb := &KeyboardState{Layout: [][]KeyDef{{{Code: KEY_A}, {Code: KEY_BACKSPACE}}}}
	kb.Navigate(0, 1)
	if !touch.StartRepeat(event, backspace, time.Unix(100, 0)) {
		t.Fatal("StartRepeat returned false")
	}
	if !pressPointerKey(kb, backspace, nil) {
		t.Fatal("pointer repeat start did not press selected key")
	}
	if kb.CursorRow != backspace.Row || kb.CursorCol != backspace.Col {
		t.Errorf("pointer repeat cursor = (%d, %d); want (%d, %d)", kb.CursorRow, kb.CursorCol, backspace.Row, backspace.Col)
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

func TestMouseAndTouchContactsRemainIndependent(t *testing.T) {
	geometry, width, height := testTouchGeometry()
	var touch TouchInput
	touchPos := KeyPosition{Row: 0, Col: 0}
	mousePos := KeyPosition{Row: 0, Col: 1}

	touch.Handle(touchEventAt(t, TouchDown, 10, 20, 30, touchPos, geometry, width, height), 10, width, height, geometry)
	mouseDown, ok := mouseEventFromSDLEvent(mouseSDLEventAt(t, SDL_EVENT_MOUSE_BUTTON_DOWN, 10, mousePos, geometry), width, height)
	if !ok {
		t.Fatal("mouse button down was not converted")
	}
	if activate, changed := touch.Handle(mouseDown, 10, width, height, geometry); activate != nil || !changed {
		t.Fatalf("mouse down = %v, %v; want nil, true", activate, changed)
	}
	if !touch.IsSelected(touchPos) || !touch.IsSelected(mousePos) || len(touch.contacts) != 2 {
		t.Fatalf("contacts = %#v; want independent touch and mouse selections", touch.contacts)
	}
	now := time.Unix(100, 0)
	touchDown := touchEventAt(t, TouchDown, 10, 20, 30, touchPos, geometry, width, height)
	if !touch.StartRepeat(touchDown, touchPos, now) || !touch.StartRepeat(mouseDown, mousePos, now) {
		t.Fatal("mouse and touch repeats did not start independently")
	}
	if got := touch.RepeatableContacts(now.Add(400*time.Millisecond), 400*time.Millisecond, 80*time.Millisecond); len(got) != 2 {
		t.Fatalf("mouse and touch repeats = %v; want two", got)
	}

	mouseUp, ok := mouseEventFromSDLEvent(mouseSDLEventAt(t, SDL_EVENT_MOUSE_BUTTON_UP, 10, mousePos, geometry), width, height)
	if !ok {
		t.Fatal("mouse button up was not converted")
	}
	repeating := touch.IsRepeating(mouseUp)
	activate, changed := touch.Handle(mouseUp, 10, width, height, geometry)
	if activate == nil || *activate != mousePos || !changed || !repeating {
		t.Fatalf("mouse up = %v, %v, repeating %v; want %v, true, true", activate, changed, repeating, mousePos)
	}
	if !touch.IsSelected(touchPos) || len(touch.contacts) != 1 {
		t.Errorf("mouse release changed touch contact: %#v", touch.contacts)
	}
	if _, active := touch.NextRepeat(now.Add(400*time.Millisecond), 400*time.Millisecond, 80*time.Millisecond); !active {
		t.Error("mouse release stopped touch repeat")
	}
}

func TestMouseDragAndInvalidRelease(t *testing.T) {
	geometry, width, height := testTouchGeometry()
	first := KeyPosition{Row: 0, Col: 0}
	final := KeyPosition{Row: 0, Col: 1}
	var touch TouchInput

	for _, event := range []SDLEvent{
		mouseSDLEventAt(t, SDL_EVENT_MOUSE_BUTTON_DOWN, 10, first, geometry),
		mouseSDLEventAt(t, SDL_EVENT_MOUSE_MOTION, 10, final, geometry),
	} {
		pointer, ok := mouseEventFromSDLEvent(event, width, height)
		if !ok {
			t.Fatalf("mouseEventFromSDLEvent(%#v) returned false", event)
		}
		if activate, changed := touch.Handle(pointer, 10, width, height, geometry); activate != nil || !changed {
			t.Fatalf("Handle(%#v) = %v, %v; want nil, true", pointer, activate, changed)
		}
	}
	if touch.IsSelected(first) || !touch.IsSelected(final) {
		t.Fatalf("mouse drag contacts = %#v; want %v selected", touch.contacts, final)
	}

	release, ok := mouseEventFromSDLEvent(mouseSDLEventAt(t, SDL_EVENT_MOUSE_BUTTON_UP, 10, final, geometry), width, height)
	if !ok {
		t.Fatal("mouse button up was not converted")
	}
	activate, changed := touch.Handle(release, 10, width, height, geometry)
	if activate == nil || *activate != final || !changed {
		t.Fatalf("mouse release = %v, %v; want %v, true", activate, changed, final)
	}
	kb := NewKeyboardState(LayoutQWERTY)
	if !pressPointerKey(kb, *activate, nil) {
		t.Fatal("mouse activation did not press selected key")
	}
	if kb.CursorRow != final.Row || kb.CursorCol != final.Col {
		t.Errorf("mouse activation cursor = (%d, %d); want (%d, %d)", kb.CursorRow, kb.CursorCol, final.Row, final.Col)
	}
	if activate, changed := touch.Handle(release, 10, width, height, geometry); activate != nil || changed {
		t.Errorf("duplicate mouse release = %v, %v; want ignored", activate, changed)
	}

	down, _ := mouseEventFromSDLEvent(mouseSDLEventAt(t, SDL_EVENT_MOUSE_BUTTON_DOWN, 10, first, geometry), width, height)
	touch.Handle(down, 10, width, height, geometry)
	outside, ok := mouseEventFromSDLEvent(SDLEvent{Type: SDL_EVENT_MOUSE_BUTTON_UP, WindowID: 10, Button: SDL_BUTTON_LEFT, X: -1, Y: 0}, width, height)
	if !ok {
		t.Fatal("outside mouse release was not converted")
	}
	if activate, changed := touch.Handle(outside, 10, width, height, geometry); activate != nil || !changed || len(touch.contacts) != 0 {
		t.Errorf("outside mouse release = %v, %v, %#v; want nil, true, empty", activate, changed, touch.contacts)
	}
}

func TestMouseEventConversionRejectsInvalidSequences(t *testing.T) {
	geometry, width, height := testTouchGeometry()
	valid := mouseSDLEventAt(t, SDL_EVENT_MOUSE_BUTTON_DOWN, 10, KeyPosition{Row: 0, Col: 0}, geometry)

	tests := []struct {
		name         string
		event        SDLEvent
		windowWidth  int32
		windowHeight int32
	}{
		{name: "right button", event: SDLEvent{Type: SDL_EVENT_MOUSE_BUTTON_DOWN, WindowID: 10, Button: 3, X: valid.X, Y: valid.Y}, windowWidth: width, windowHeight: height},
		{name: "unknown event", event: SDLEvent{Type: SDL_EVENT_QUIT, WindowID: 10, X: valid.X, Y: valid.Y}, windowWidth: width, windowHeight: height},
		{name: "zero width", event: valid, windowHeight: height},
		{name: "zero height", event: valid, windowWidth: width},
		{name: "non-finite coordinate", event: SDLEvent{Type: SDL_EVENT_MOUSE_MOTION, WindowID: 10, X: float32(math.NaN()), Y: valid.Y}, windowWidth: width, windowHeight: height},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if event, ok := mouseEventFromSDLEvent(tt.event, tt.windowWidth, tt.windowHeight); ok {
				t.Errorf("mouseEventFromSDLEvent() = %#v, true; want false", event)
			}
		})
	}

	var touch TouchInput
	motion, ok := mouseEventFromSDLEvent(mouseSDLEventAt(t, SDL_EVENT_MOUSE_MOTION, 10, KeyPosition{Row: 0, Col: 0}, geometry), width, height)
	if !ok {
		t.Fatal("mouse motion was not converted")
	}
	if activate, changed := touch.Handle(motion, 10, width, height, geometry); activate != nil || changed {
		t.Errorf("orphan mouse motion = %v, %v; want ignored", activate, changed)
	}

	wrongWindow, _ := mouseEventFromSDLEvent(valid, width, height)
	if activate, changed := touch.Handle(wrongWindow, 11, width, height, geometry); activate != nil || changed {
		t.Errorf("wrong-window mouse down = %v, %v; want ignored", activate, changed)
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
			var touch TouchInput
			contacts := map[KeyPosition]int64{shift: 30, q: 31}
			for _, pos := range []KeyPosition{shift, q} {
				touch.Handle(touchEventAt(t, TouchDown, 10, 20, contacts[pos], pos, geometry, width, height), 10, width, height, geometry)
			}
			for _, pos := range tt.releaseOrder {
				activate, _ := touch.Handle(touchEventAt(t, TouchUp, 10, 20, contacts[pos], pos, geometry, width, height), 10, width, height, geometry)
				if activate == nil || !pressPointerKey(kb, *activate, nil) {
					t.Fatalf("TouchUp did not activate %v", pos)
				}
			}
			if kb.ShiftActive != tt.wantShiftActive {
				t.Errorf("ShiftActive = %v; want %v", kb.ShiftActive, tt.wantShiftActive)
			}
			last := tt.releaseOrder[len(tt.releaseOrder)-1]
			if kb.CursorRow != last.Row || kb.CursorCol != last.Col {
				t.Errorf("touch activation cursor = (%d, %d); want (%d, %d)", kb.CursorRow, kb.CursorCol, last.Row, last.Col)
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
		Source:   PointerTouch,
		Phase:    phase,
		WindowID: windowID,
		TouchID:  touchID,
		FingerID: fingerID,
		X:        (rect.X + rect.W/2) / float32(width),
		Y:        (rect.Y + rect.H/2) / float32(height),
	}
}

func mouseSDLEventAt(t *testing.T, eventType uint32, windowID uint32, pos KeyPosition, geometry KeyboardGeometry) SDLEvent {
	t.Helper()
	rect, ok := geometry.KeyRect(pos)
	if !ok {
		t.Fatalf("KeyRect(%v) returned false", pos)
	}
	return SDLEvent{
		Type:     eventType,
		WindowID: windowID,
		Button:   SDL_BUTTON_LEFT,
		X:        rect.X + rect.W/2,
		Y:        rect.Y + rect.H/2,
	}
}
