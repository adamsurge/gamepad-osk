package main

import "testing"

func TestKeySelectionStateKeepsCursorAndTouchIndependent(t *testing.T) {
	kb := NewKeyboardState(LayoutQWERTY)
	kb.CursorRow = 2
	kb.CursorCol = 1
	touchPos := KeyPosition{Row: 3, Col: 2}

	isCursor, isTouch := keySelectionState(KeyPosition{Row: 2, Col: 1}, kb, touchPos, true)
	if !isCursor || isTouch {
		t.Errorf("cursor key state = %v, %v; want true, false", isCursor, isTouch)
	}
	isCursor, isTouch = keySelectionState(touchPos, kb, touchPos, true)
	if isCursor || !isTouch {
		t.Errorf("touch key state = %v, %v; want false, true", isCursor, isTouch)
	}
	if kb.CursorRow != 2 || kb.CursorCol != 1 {
		t.Errorf("touch selection moved cursor to (%d, %d)", kb.CursorRow, kb.CursorCol)
	}
}
