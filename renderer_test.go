package main

import "testing"

func TestKeySelectionStateKeepsCursorAndTouchIndependent(t *testing.T) {
	kb := NewKeyboardState(LayoutQWERTY)
	kb.CursorRow = 2
	kb.CursorCol = 1
	touchPositions := []KeyPosition{{Row: 3, Col: 2}, {Row: 4, Col: 3}}
	touch := TouchInput{contacts: map[TouchContact]TouchSequence{
		{TouchID: 1, FingerID: 1}: {selected: touchPositions[0], hasSelected: true},
		{TouchID: 1, FingerID: 2}: {selected: touchPositions[1], hasSelected: true},
	}}

	isCursor, isTouch := keySelectionState(KeyPosition{Row: 2, Col: 1}, kb, &touch)
	if !isCursor || isTouch {
		t.Errorf("cursor key state = %v, %v; want true, false", isCursor, isTouch)
	}
	for _, touchPos := range touchPositions {
		isCursor, isTouch = keySelectionState(touchPos, kb, &touch)
		if isCursor || !isTouch {
			t.Errorf("touch key %v state = %v, %v; want false, true", touchPos, isCursor, isTouch)
		}
	}
	if kb.CursorRow != 2 || kb.CursorCol != 1 {
		t.Errorf("touch selection moved cursor to (%d, %d)", kb.CursorRow, kb.CursorCol)
	}
}
