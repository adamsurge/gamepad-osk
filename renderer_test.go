package main

import (
	"path/filepath"
	"testing"
)

func TestFindFontPrefersFontconfig(t *testing.T) {
	previousLookup, previousExists := fontconfigLookup, fontFileExists
	fontconfigLookup = func(family string) string {
		if family == "DejaVu Sans" {
			return "/home/user/.local/share/fonts/DejaVuSans.ttf"
		}
		return ""
	}
	fontFileExists = func(path string) bool { return path == "/home/user/.local/share/fonts/DejaVuSans.ttf" }
	t.Cleanup(func() {
		fontconfigLookup = previousLookup
		fontFileExists = previousExists
	})

	if got, want := findFont("DejaVu Sans"), "/home/user/.local/share/fonts/DejaVuSans.ttf"; got != want {
		t.Errorf("findFont() = %q, want %q", got, want)
	}
}

func TestFindFontFallsBackToPackageAndFHSPaths(t *testing.T) {
	tmp := t.TempDir()
	useExecutablePath(t, filepath.Join(tmp, "prefix", "bin", "gamepad-osk"))
	previousLookup, previousExists, previousDirs := fontconfigLookup, fontFileExists, fhsFontDirs
	fontconfigLookup = func(string) string { return "" }
	packageFont := filepath.Join(tmp, "prefix", "share", "gamepad-osk", "fonts", "DejaVuSans.ttf")
	fhsFont := filepath.Join(tmp, "legacy", "DejaVuSans.ttf")
	fontFileExists = func(path string) bool { return path == packageFont || path == fhsFont }
	fhsFontDirs = []string{filepath.Join(tmp, "legacy")}
	t.Cleanup(func() {
		fontconfigLookup = previousLookup
		fontFileExists = previousExists
		fhsFontDirs = previousDirs
	})

	if got := findFont("DejaVu Sans"); got != packageFont {
		t.Errorf("package fallback = %q, want %q", got, packageFont)
	}
	fontFileExists = func(path string) bool { return path == fhsFont }
	if got := findFont("DejaVu Sans"); got != fhsFont {
		t.Errorf("FHS fallback = %q, want %q", got, fhsFont)
	}
}
