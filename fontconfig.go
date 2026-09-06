package main

/*
#cgo pkg-config: fontconfig
#include <fontconfig/fontconfig.h>
#include <stdlib.h>
#include <string.h>

static char* find_exact_fontconfig_font(const char *family) {
	FcPattern *pattern;
	FcPattern *match;
	FcResult result;
	FcChar8 *matchedFamily;
	FcChar8 *file;
	char *path = NULL;
	int index;
	int exact = 0;

	if (!FcInit()) return NULL;
	pattern = FcPatternCreate();
	if (!pattern) return NULL;
	if (!FcPatternAddString(pattern, FC_FAMILY, (const FcChar8 *)family)) goto done;
	FcConfigSubstitute(NULL, pattern, FcMatchPattern);
	FcDefaultSubstitute(pattern);
	match = FcFontMatch(NULL, pattern, &result);
	if (!match) goto done;
	for (index = 0; FcPatternGetString(match, FC_FAMILY, index, &matchedFamily) == FcResultMatch; index++) {
		if (strcmp((const char *)matchedFamily, family) == 0) {
			exact = 1;
			break;
		}
	}
	if (!exact) goto destroy_match;
	if (FcPatternGetString(match, FC_FILE, 0, &file) != FcResultMatch) goto destroy_match;
	path = malloc(strlen((const char *)file) + 1);
	if (path) strcpy(path, (const char *)file);

destroy_match:
	FcPatternDestroy(match);
done:
	FcPatternDestroy(pattern);
	return path;
}
*/
import "C"

import (
	"os"
	"path/filepath"
	"unsafe"
)

var executablePath = os.Executable

// executablePrefix returns the installation prefix for an executable in <prefix>/bin.
func executablePrefix() string {
	executable, err := executablePath()
	if err != nil || !filepath.IsAbs(executable) {
		return ""
	}

	binDir := filepath.Dir(filepath.Clean(executable))
	if filepath.Base(binDir) != "bin" {
		return ""
	}
	return filepath.Dir(binDir)
}

// packageDataPath returns a path below the installed package data directory.
func packageDataPath(elements ...string) string {
	prefix := executablePrefix()
	if prefix == "" {
		return ""
	}
	return filepath.Join(append([]string{prefix, "share", "gamepad-osk"}, elements...)...)
}

// findFontconfigFont returns a readable font file for an exact family match.
func findFontconfigFont(family string) string {
	if family == "" {
		return ""
	}

	requestedFamily := C.CString(family)
	defer C.free(unsafe.Pointer(requestedFamily))
	path := C.find_exact_fontconfig_font(requestedFamily)
	if path == nil {
		return ""
	}
	defer C.free(unsafe.Pointer(path))

	fontPath := C.GoString(path)
	if !isReadableFile(fontPath) {
		return ""
	}
	return fontPath
}

func isReadableFile(path string) bool {
	file, err := os.Open(path) //nolint:gosec // G304: font paths come from Fontconfig or fixed fallback paths
	if err != nil {
		return false
	}
	defer func() { _ = file.Close() }()
	info, err := file.Stat()
	return err == nil && info.Mode().IsRegular()
}
