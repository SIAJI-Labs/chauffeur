package installers

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestImagickLibraryDirs(t *testing.T) {
	got := imagickLibraryDirs("-L/usr/lib -lMagickWand-7.Q16HDRI -L/opt/imagemagick/lib -L/usr/lib")
	want := []string{"/usr/lib", "/opt/imagemagick/lib"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("imagickLibraryDirs() = %#v, want %#v", got, want)
	}
}

func TestValidateImagick(t *testing.T) {
	bin := filepath.Join(t.TempDir(), "php")
	if err := os.WriteFile(bin, []byte("#!/bin/sh\nexit 0\n"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := validateImagick(bin, filepath.Join(t.TempDir(), "conf.d", "imagick.ini")); err != nil {
		t.Fatalf("validateImagick() error = %v", err)
	}
}

func TestValidateImagickReportsLoadFailure(t *testing.T) {
	bin := filepath.Join(t.TempDir(), "php")
	if err := os.WriteFile(bin, []byte("#!/bin/sh\necho missing MagickWand >&2\nexit 1\n"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := validateImagick(bin, "imagick.ini"); err == nil {
		t.Fatal("validateImagick() error = nil, want load failure")
	}
}
