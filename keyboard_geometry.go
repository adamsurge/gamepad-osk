package main

import "math"

type KeyPosition struct {
	Row int
	Col int
}

type KeyboardGeometry struct {
	Layout       [][]KeyDef
	Unit         int32
	Padding      int32
	StatusHeight int32
}

func NewKeyboardGeometry(layout [][]KeyDef, unit, padding, statusHeight int32) KeyboardGeometry {
	return KeyboardGeometry{
		Layout:       layout,
		Unit:         unit,
		Padding:      padding,
		StatusHeight: statusHeight,
	}
}

func (g KeyboardGeometry) KeyRect(pos KeyPosition) (FRect, bool) {
	if pos.Row < 0 || pos.Row >= len(g.Layout) {
		return FRect{}, false
	}
	row := g.Layout[pos.Row]
	if pos.Col < 0 || pos.Col >= len(row) {
		return FRect{}, false
	}

	x := g.Padding
	for col := range pos.Col {
		x += g.keyWidth(row[col]) + g.Padding
	}
	y := g.Padding + g.StatusHeight + int32(pos.Row)*(g.Unit+g.Padding) //nolint:gosec // G115: layout row index fits in int32

	return FRect{
		X: float32(x),
		Y: float32(y),
		W: float32(g.keyWidth(row[pos.Col])),
		H: float32(g.Unit),
	}, true
}

func (g KeyboardGeometry) HitTest(x, y float32) (KeyPosition, bool) {
	if !finiteCoordinate(x) || !finiteCoordinate(y) {
		return KeyPosition{}, false
	}
	width, height := g.WindowSize()
	if x < 0 || y < 0 || x >= float32(width) || y >= float32(height) {
		return KeyPosition{}, false
	}

	for rowIdx, row := range g.Layout {
		for colIdx := range row {
			pos := KeyPosition{Row: rowIdx, Col: colIdx}
			rect, _ := g.KeyRect(pos)
			if x >= rect.X && x < rect.X+rect.W && y >= rect.Y && y < rect.Y+rect.H {
				return pos, true
			}
		}
	}
	return KeyPosition{}, false
}

func (g KeyboardGeometry) WindowSize() (int32, int32) {
	var maxWidth int32
	for _, row := range g.Layout {
		rowWidth := g.Padding
		for _, key := range row {
			rowWidth += g.keyWidth(key) + g.Padding
		}
		if rowWidth > maxWidth {
			maxWidth = rowWidth
		}
	}
	height := g.Padding + g.StatusHeight + int32(len(g.Layout))*(g.Unit+g.Padding) //nolint:gosec // G115: layout row count fits in int32
	return maxWidth, height
}

func (g KeyboardGeometry) keyWidth(key KeyDef) int32 {
	return int32(key.Width * float64(g.Unit))
}

func finiteCoordinate(value float32) bool {
	return !math.IsNaN(float64(value)) && !math.IsInf(float64(value), 0)
}

func CalcWindowSize(layout [][]KeyDef, unit, pad, statusH int32) (int32, int32) {
	return NewKeyboardGeometry(layout, unit, pad, statusH).WindowSize()
}
