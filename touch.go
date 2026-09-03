package main

type TouchPhase uint8

const (
	TouchDown TouchPhase = iota
	TouchMotion
	TouchUp
	TouchCanceled
)

type TouchEvent struct {
	Phase    TouchPhase
	WindowID uint32
	TouchID  uint64
	FingerID int64
	X        float32
	Y        float32
}

type TouchContact struct {
	TouchID  uint64
	FingerID int64
}

type TouchSequence struct {
	selected    KeyPosition
	hasSelected bool
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
	return TouchContact{TouchID: event.TouchID, FingerID: event.FingerID}
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
