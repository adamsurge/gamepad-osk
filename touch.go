package main

import "time"

type TouchPhase uint8

const (
	TouchDown TouchPhase = iota
	TouchMotion
	TouchUp
	TouchCanceled
)

type PointerSource uint8

const (
	PointerTouch PointerSource = iota
	PointerMouse
)

type TouchEvent struct {
	Source   PointerSource
	Phase    TouchPhase
	WindowID uint32
	TouchID  uint64
	FingerID int64
	X        float32
	Y        float32
}

type TouchContact struct {
	Source   PointerSource
	TouchID  uint64
	FingerID int64
}

type TouchSequence struct {
	selected    KeyPosition
	hasSelected bool
	repeat      pointerRepeat
}

type pointerRepeat struct {
	active  bool
	started bool
	last    time.Time
	initial bool
}

type TouchInput struct {
	contacts map[TouchContact]TouchSequence
}

func (t *TouchInput) Handle(
	event TouchEvent,
	expectedWindowID uint32,
	windowWidth, windowHeight int32,
	geometry KeyboardGeometry,
) (activate *KeyPosition, changed bool) {
	if event.WindowID != expectedWindowID {
		return nil, false
	}

	switch event.Phase {
	case TouchDown:
		contact := contactFromEvent(event)
		if _, exists := t.contacts[contact]; exists {
			return nil, false
		}
		pos, ok := touchHitTest(event.X, event.Y, windowWidth, windowHeight, geometry)
		if !ok {
			return nil, false
		}
		if t.contacts == nil {
			t.contacts = make(map[TouchContact]TouchSequence)
		}
		t.contacts[contact] = TouchSequence{selected: pos, hasSelected: true}
		return nil, true

	case TouchMotion:
		contact := contactFromEvent(event)
		sequence, exists := t.contacts[contact]
		if !exists {
			return nil, false
		}
		pos, ok := touchHitTest(event.X, event.Y, windowWidth, windowHeight, geometry)
		if !setTouchSelection(&sequence, pos, ok) {
			return nil, false
		}
		sequence.repeat.active = false
		t.contacts[contact] = sequence
		return nil, true

	case TouchUp:
		contact := contactFromEvent(event)
		sequence, exists := t.contacts[contact]
		if !exists {
			return nil, false
		}
		pos, ok := touchHitTest(event.X, event.Y, windowWidth, windowHeight, geometry)
		activateSelected := ok && sequence.hasSelected && pos == sequence.selected
		selected := sequence.selected
		delete(t.contacts, contact)
		if activateSelected {
			return &selected, true
		}
		return nil, true

	case TouchCanceled:
		contact := contactFromEvent(event)
		if _, exists := t.contacts[contact]; !exists {
			return nil, false
		}
		delete(t.contacts, contact)
		return nil, true
	}

	return nil, false
}

// StartRepeat arms repeat for event's contact when it selects pos.
func (t *TouchInput) StartRepeat(event TouchEvent, pos KeyPosition, now time.Time) bool {
	contact := contactFromEvent(event)
	sequence, ok := t.contacts[contact]
	if !ok || !sequence.hasSelected || sequence.selected != pos {
		return false
	}
	sequence.repeat = pointerRepeat{active: true, started: true, last: now, initial: true}
	t.contacts[contact] = sequence
	return true
}

func (t *TouchInput) IsRepeating(event TouchEvent) bool {
	sequence, ok := t.contacts[contactFromEvent(event)]
	return ok && sequence.repeat.started
}

// RepeatableContacts returns contacts whose next repeat is due and advances them.
func (t *TouchInput) RepeatableContacts(now time.Time, delay, rate time.Duration) []KeyPosition {
	var due []KeyPosition
	for contact, sequence := range t.contacts {
		if !sequence.repeat.active || !sequence.hasSelected {
			continue
		}
		wait := rate
		if sequence.repeat.initial {
			wait = delay
		}
		if now.Sub(sequence.repeat.last) >= wait {
			due = append(due, sequence.selected)
			sequence.repeat.last = now
			sequence.repeat.initial = false
			t.contacts[contact] = sequence
		}
	}
	return due
}

func (t *TouchInput) NextRepeat(now time.Time, delay, rate time.Duration) (time.Duration, bool) {
	var next time.Duration
	active := false
	for _, sequence := range t.contacts {
		if !sequence.repeat.active {
			continue
		}
		active = true
		wait := rate
		if sequence.repeat.initial {
			wait = delay
		}
		remaining := wait - now.Sub(sequence.repeat.last)
		if next == 0 || remaining < next {
			next = remaining
		}
	}
	return next, active
}

func (t *TouchInput) IsSelected(pos KeyPosition) bool {
	for _, sequence := range t.contacts {
		if sequence.hasSelected && sequence.selected == pos {
			return true
		}
	}
	return false
}

func (t *TouchInput) Cancel() bool {
	if len(t.contacts) == 0 {
		return false
	}
	t.contacts = nil
	return true
}

func contactFromEvent(event TouchEvent) TouchContact {
	return TouchContact{Source: event.Source, TouchID: event.TouchID, FingerID: event.FingerID}
}

func setTouchSelection(sequence *TouchSequence, pos KeyPosition, ok bool) bool {
	if sequence.hasSelected == ok && (!ok || sequence.selected == pos) {
		return false
	}
	sequence.selected = pos
	sequence.hasSelected = ok
	return true
}

func touchHitTest(
	x, y float32,
	windowWidth, windowHeight int32,
	geometry KeyboardGeometry,
) (KeyPosition, bool) {
	if !finiteCoordinate(x) || !finiteCoordinate(y) || x < 0 || y < 0 || x >= 1 || y >= 1 ||
		windowWidth <= 0 || windowHeight <= 0 {
		return KeyPosition{}, false
	}
	return geometry.HitTest(x*float32(windowWidth), y*float32(windowHeight))
}
