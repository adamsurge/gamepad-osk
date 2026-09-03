package main

import (
	"math"
	"testing"
)

func TestKeyboardGeometryKeyCenters(t *testing.T) {
	geometry := NewKeyboardGeometry(LayoutQWERTY, 73, 5, 29)

	for rowIdx, row := range LayoutQWERTY {
		for colIdx := range row {
			want := KeyPosition{Row: rowIdx, Col: colIdx}
			rect, ok := geometry.KeyRect(want)
			if !ok {
				t.Fatalf("KeyRect(%v) returned false", want)
			}
			got, ok := geometry.HitTest(rect.X+rect.W/2, rect.Y+rect.H/2)
			if !ok || got != want {
				t.Errorf("center of key %v resolved to %v, %v", want, got, ok)
			}
		}
	}
}

func TestKeyboardGeometryNonKeyAreas(t *testing.T) {
	layout := [][]KeyDef{
		{{Width: 1}, {Width: 1}},
		{{Width: 1}},
	}
	geometry := NewKeyboardGeometry(layout, 10, 2, 5)

	tests := []struct {
		name string
		x    float32
		y    float32
	}{
		{name: "left padding", x: 1, y: 10},
		{name: "right padding", x: 25, y: 10},
		{name: "top padding", x: 3, y: 1},
		{name: "status bar", x: 3, y: 4},
		{name: "column gap", x: 12, y: 10},
		{name: "row gap", x: 3, y: 18},
		{name: "trailing row space", x: 15, y: 24},
		{name: "bottom padding", x: 3, y: 30},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if pos, ok := geometry.HitTest(tt.x, tt.y); ok {
				t.Errorf("HitTest(%v, %v) = %v, true; want no key", tt.x, tt.y, pos)
			}
		})
	}
}

func TestKeyboardGeometryHalfOpenBoundaries(t *testing.T) {
	geometry := NewKeyboardGeometry([][]KeyDef{{{Width: 1}, {Width: 1}}}, 10, 0, 0)

	tests := []struct {
		name string
		x    float32
		y    float32
		want KeyPosition
		ok   bool
	}{
		{name: "left inclusive", x: 0, y: 0, want: KeyPosition{Row: 0, Col: 0}, ok: true},
		{name: "top inclusive", x: 5, y: 0, want: KeyPosition{Row: 0, Col: 0}, ok: true},
		{name: "shared edge belongs to right key", x: 10, y: 5, want: KeyPosition{Row: 0, Col: 1}, ok: true},
		{name: "window right exclusive", x: 20, y: 5, ok: false},
		{name: "window bottom exclusive", x: 5, y: 10, ok: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := geometry.HitTest(tt.x, tt.y)
			if ok != tt.ok || got != tt.want {
				t.Errorf("HitTest(%v, %v) = %v, %v; want %v, %v", tt.x, tt.y, got, ok, tt.want, tt.ok)
			}
		})
	}
}

func TestKeyboardGeometryInvalidCoordinates(t *testing.T) {
	geometry := NewKeyboardGeometry([][]KeyDef{{{Width: 1}}}, 10, 2, 5)
	width, height := geometry.WindowSize()
	tests := []struct {
		name string
		x    float32
		y    float32
	}{
		{name: "negative x", x: -1, y: 8},
		{name: "negative y", x: 3, y: -1},
		{name: "nan x", x: float32(math.NaN()), y: 8},
		{name: "nan y", x: 3, y: float32(math.NaN())},
		{name: "positive infinity", x: float32(math.Inf(1)), y: 8},
		{name: "negative infinity", x: 3, y: float32(math.Inf(-1))},
		{name: "past width", x: float32(width), y: 8},
		{name: "past height", x: 3, y: float32(height)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if pos, ok := geometry.HitTest(tt.x, tt.y); ok {
				t.Errorf("HitTest(%v, %v) = %v, true; want no key", tt.x, tt.y, pos)
			}
		})
	}
}

func TestKeyboardGeometryFractionalWidthTruncation(t *testing.T) {
	layout := [][]KeyDef{{{Width: 1.25}, {Width: 0.75}}}
	geometry := NewKeyboardGeometry(layout, 7, 3, 4)

	first, ok := geometry.KeyRect(KeyPosition{Row: 0, Col: 0})
	if !ok {
		t.Fatal("first KeyRect returned false")
	}
	second, ok := geometry.KeyRect(KeyPosition{Row: 0, Col: 1})
	if !ok {
		t.Fatal("second KeyRect returned false")
	}
	if first.W != 8 || second.X != 14 || second.W != 5 {
		t.Errorf("fractional rects = first %#v, second %#v; want widths 8,5 and second x 14", first, second)
	}
}

func TestKeyboardGeometryWindowSizeMatchesLegacyCalculation(t *testing.T) {
	tests := []struct {
		name    string
		layout  [][]KeyDef
		unit    int32
		padding int32
		status  int32
	}{
		{name: "qwerty", layout: LayoutQWERTY, unit: 73, padding: 5, status: 29},
		{name: "fractional", layout: [][]KeyDef{{{Width: 1.25}, {Width: 0.75}}, {{Width: 3.5}}}, unit: 7, padding: 3, status: 4},
		{name: "empty layout", layout: nil, unit: 10, padding: 2, status: 5},
		{name: "empty row", layout: [][]KeyDef{{}}, unit: 10, padding: 2, status: 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wantWidth, wantHeight := legacyWindowSize(tt.layout, tt.unit, tt.padding, tt.status)
			geometry := NewKeyboardGeometry(tt.layout, tt.unit, tt.padding, tt.status)
			gotWidth, gotHeight := geometry.WindowSize()
			if gotWidth != wantWidth || gotHeight != wantHeight {
				t.Errorf("WindowSize() = (%d, %d), want (%d, %d)", gotWidth, gotHeight, wantWidth, wantHeight)
			}
			calcWidth, calcHeight := CalcWindowSize(tt.layout, tt.unit, tt.padding, tt.status)
			if calcWidth != gotWidth || calcHeight != gotHeight {
				t.Errorf("CalcWindowSize() = (%d, %d), geometry = (%d, %d)", calcWidth, calcHeight, gotWidth, gotHeight)
			}
		})
	}
}

func TestKeyboardGeometryInvalidPositionsAndEmptyRows(t *testing.T) {
	geometry := NewKeyboardGeometry([][]KeyDef{{}, {{Width: 1}}}, 10, 2, 5)
	for _, pos := range []KeyPosition{
		{Row: -1, Col: 0},
		{Row: 0, Col: 0},
		{Row: 1, Col: -1},
		{Row: 1, Col: 1},
		{Row: 2, Col: 0},
	} {
		if rect, ok := geometry.KeyRect(pos); ok {
			t.Errorf("KeyRect(%v) = %#v, true; want invalid", pos, rect)
		}
	}
}

func legacyWindowSize(layout [][]KeyDef, unit, padding, statusHeight int32) (int32, int32) {
	var maxWidth int32
	for _, row := range layout {
		rowWidth := padding
		for _, key := range row {
			rowWidth += int32(key.Width*float64(unit)) + padding
		}
		if rowWidth > maxWidth {
			maxWidth = rowWidth
		}
	}
	height := padding + statusHeight + int32(len(layout))*(unit+padding) //nolint:gosec // Test layouts fit in int32.
	return maxWidth, height
}
