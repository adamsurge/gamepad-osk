package main

import "testing"

func TestFindFontconfigFontRequiresExactFamily(t *testing.T) {
	if got := findFontconfigFont("gamepad-osk-definitely-missing-font"); got != "" {
		t.Errorf("substituted family path = %q, want empty path", got)
	}
}

func TestFindFontconfigFontFindsExactInstalledFamily(t *testing.T) {
	path := findFontconfigFont("DejaVu Sans")
	if path == "" {
		t.Skip("DejaVu Sans is not installed")
	}
	if !isReadableFile(path) {
		t.Errorf("Fontconfig returned unreadable path %q", path)
	}
}
