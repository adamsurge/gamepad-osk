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

func TestKeySelectionColorsDistinguishTouchFromGamepadHighlight(t *testing.T) {
	theme := Theme{
		KeyBgPressed:    c(10, 20, 30),
		HighlightBg:     c(40, 50, 60),
		HighlightBorder: c(70, 80, 90),
	}

	tests := []struct {
		name                       string
		isCursor, isTouch, flashed bool
		wantBg, wantBorder         Color
		wantSelected               bool
	}{
		{name: "gamepad cursor", isCursor: true, wantBg: theme.HighlightBg, wantBorder: theme.HighlightBorder, wantSelected: true},
		{name: "shortcut flash", flashed: true, wantBg: theme.HighlightBg, wantBorder: theme.HighlightBorder, wantSelected: true},
		{name: "touch", isTouch: true, wantBg: theme.KeyBgPressed, wantBorder: theme.HighlightBorder, wantSelected: true},
		{name: "touch over gamepad cursor", isCursor: true, isTouch: true, wantBg: theme.KeyBgPressed, wantBorder: theme.HighlightBorder, wantSelected: true},
		{name: "unselected"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bg, border, selected := keySelectionColors(theme, tt.isCursor, tt.isTouch, tt.flashed)
			if bg != tt.wantBg || border != tt.wantBorder || selected != tt.wantSelected {
				t.Errorf("keySelectionColors() = %v, %v, %v; want %v, %v, %v", bg, border, selected, tt.wantBg, tt.wantBorder, tt.wantSelected)
			}
		})
	}
}

func TestKeySelectionTextColorPreservesTouchContrast(t *testing.T) {
	theme := Theme{KeyText: c(10, 20, 30)}
	white := c(255, 255, 255)

	tests := []struct {
		name              string
		isTouch, selected bool
		want              Color
	}{
		{name: "gamepad highlight", selected: true, want: white},
		{name: "touch", isTouch: true, selected: true, want: theme.KeyText},
		{name: "touch over gamepad cursor", isTouch: true, selected: true, want: theme.KeyText},
		{name: "unselected", want: theme.KeyText},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := keySelectionTextColor(theme, tt.isTouch, tt.selected); got != tt.want {
				t.Errorf("keySelectionTextColor() = %v; want %v", got, tt.want)
			}
		})
	}
}
