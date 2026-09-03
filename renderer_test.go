package main

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestFontSearchDirs(t *testing.T) {
	first := t.TempDir()
	second := t.TempDir()
	t.Setenv(fontDirsEnv, first+string(os.PathListSeparator)+second)

	want := append([]string{first, second}, fallbackFontDirs...)
	if got := fontSearchDirs(); !slices.Equal(got, want) {
		t.Fatalf("fontSearchDirs() = %q, want %q", got, want)
	}
}

func TestFontSearchDirsFallsBackWithoutEnvironment(t *testing.T) {
	t.Setenv(fontDirsEnv, "")

	if got := fontSearchDirs(); !slices.Equal(got, fallbackFontDirs) {
		t.Fatalf("fontSearchDirs() = %q, want %q", got, fallbackFontDirs)
	}
}

func TestFontSearchDirsIgnoresEmptyEntries(t *testing.T) {
	dir := t.TempDir()
	separator := string(os.PathListSeparator)
	t.Setenv(fontDirsEnv, separator+dir+separator)

	want := append([]string{dir}, fallbackFontDirs...)
	if got := fontSearchDirs(); !slices.Equal(got, want) {
		t.Fatalf("fontSearchDirs() = %q, want %q", got, want)
	}
}

func TestFindFontInDirs(t *testing.T) {
	first := t.TempDir()
	second := t.TempDir()
	fontPath := filepath.Join(second, "LiberationSans-Regular.ttf")
	if err := os.WriteFile(fontPath, nil, 0600); err != nil {
		t.Fatalf("write test font: %v", err)
	}

	if got := findFontInDirs([]string{"DejaVu Sans", "Liberation Sans"}, []string{first, second}); got != fontPath {
		t.Fatalf("findFontInDirs() = %q, want %q", got, fontPath)
	}
}

func TestFindFontInDirsPrefersEarlierDirectory(t *testing.T) {
	first := t.TempDir()
	second := t.TempDir()
	firstPath := filepath.Join(first, "FreeSans.ttf")
	secondPath := filepath.Join(second, "DejaVuSans.ttf")
	for _, path := range []string{firstPath, secondPath} {
		if err := os.WriteFile(path, nil, 0600); err != nil {
			t.Fatalf("write test font: %v", err)
		}
	}

	if got := findFontInDirs([]string{"DejaVu Sans", "FreeSans"}, []string{first, second}); got != firstPath {
		t.Fatalf("findFontInDirs() = %q, want earlier-directory font %q", got, firstPath)
	}
}

func TestFindFontInDirsMissing(t *testing.T) {
	if got := findFontInDirs([]string{"DejaVu Sans", "unknown"}, []string{t.TempDir()}); got != "" {
		t.Fatalf("findFontInDirs() = %q, want empty string", got)
	}
}
