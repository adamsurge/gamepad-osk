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

type TouchInput struct {
	owned       bool
	touchID     uint64
	fingerID    int64
	selected    KeyPosition
	hasSelected bool
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
		if t.owned {
			return nil, false
		}
		pos, ok := touchHitTest(event.X, event.Y, windowWidth, windowHeight, geometry)
		if !ok {
			return nil, false
		}
		t.owned = true
		t.touchID = event.TouchID
		t.fingerID = event.FingerID
		t.selected = pos
		t.hasSelected = true
		return nil, true

	case TouchMotion:
		if !t.owns(event) {
			return nil, false
		}
		pos, ok := touchHitTest(event.X, event.Y, windowWidth, windowHeight, geometry)
		return nil, t.setSelection(pos, ok)

	case TouchUp:
		if !t.owns(event) {
			return nil, false
		}
		pos, ok := touchHitTest(event.X, event.Y, windowWidth, windowHeight, geometry)
		activateSelected := ok && t.hasSelected && pos == t.selected
		selected := t.selected
		t.reset()
		if activateSelected {
			return &selected, true
		}
		return nil, true

	case TouchCanceled:
		if !t.owns(event) {
			return nil, false
		}
		t.reset()
		return nil, true
	}

	return nil, false
}

func (t *TouchInput) Selected() (KeyPosition, bool) {
	return t.selected, t.hasSelected
}

func (t *TouchInput) Cancel() bool {
	if !t.owned {
		return false
	}
	t.reset()
	return true
}

func (t *TouchInput) owns(event TouchEvent) bool {
	return t.owned && event.TouchID == t.touchID && event.FingerID == t.fingerID
}

func (t *TouchInput) setSelection(pos KeyPosition, ok bool) bool {
	if t.hasSelected == ok && (!ok || t.selected == pos) {
		return false
	}
	t.selected = pos
	t.hasSelected = ok
	return true
}

func (t *TouchInput) reset() {
	t.owned = false
	t.touchID = 0
	t.fingerID = 0
	t.selected = KeyPosition{}
	t.hasSelected = false
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
