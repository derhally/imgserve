package internal

import (
	"path/filepath"
	"testing"
)

func TestParseImageSource(t *testing.T) {
	tests := []struct {
		value    string
		wantName string
		wantDir  string
		wantErr  bool
	}{
		{value: "/mnt/media/photos", wantName: "photos", wantDir: "/mnt/media/photos"},
		{value: "/mnt/media/photos/", wantName: "photos", wantDir: "/mnt/media/photos"},
		{value: "screensaver=/mnt/backup/ss", wantName: "screensaver", wantDir: "/mnt/backup/ss"},
		{value: " family = /mnt/backup/fam ", wantName: "family", wantDir: "/mnt/backup/fam"},
		{value: "my-pics_2.0=/mnt/x", wantName: "my-pics_2.0", wantDir: "/mnt/x"},
		{value: "", wantErr: true},
		{value: "name=", wantErr: true},
		{value: "=/mnt/x", wantErr: true},
		{value: "1920=/mnt/x", wantErr: true},
		{value: "/mnt/2024", wantErr: true},
		{value: "bad name=/mnt/x", wantErr: true},
		{value: "a/b=/mnt/x", wantErr: true},
		{value: "..=/mnt/x", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			got, err := ParseImageSource(tt.value)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got %+v", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Name != tt.wantName {
				t.Errorf("name = %q, want %q", got.Name, tt.wantName)
			}
			if got.Dir != filepath.Clean(tt.wantDir) {
				t.Errorf("dir = %q, want %q", got.Dir, tt.wantDir)
			}
		})
	}
}

func TestParseImageSourceRelativeDir(t *testing.T) {
	got, err := ParseImageSource("./images")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !filepath.IsAbs(got.Dir) {
		t.Errorf("dir %q is not absolute", got.Dir)
	}
	if got.Name != "images" {
		t.Errorf("name = %q, want images", got.Name)
	}
}

func TestValidateImageSources(t *testing.T) {
	if err := ValidateImageSources(nil); err == nil {
		t.Error("expected error for no sources")
	}

	ok := []ImageSource{{Name: "photos", Dir: "/a"}, {Name: "screensaver", Dir: "/b"}}
	if err := ValidateImageSources(ok); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	dup := []ImageSource{{Name: "photos", Dir: "/a/photos"}, {Name: "Photos", Dir: "/b/Photos"}}
	if err := ValidateImageSources(dup); err == nil {
		t.Error("expected error for duplicate names")
	}

	twoDefaults := []ImageSource{{Name: "a", Dir: "/a", Default: true}, {Name: "b", Dir: "/b", Default: true}}
	if err := ValidateImageSources(twoDefaults); err == nil {
		t.Error("expected error for two defaults")
	}
}

func TestDefaultSourceIndex(t *testing.T) {
	none := []ImageSource{{Name: "a"}, {Name: "b"}}
	if got := DefaultSourceIndex(none); got != 0 {
		t.Errorf("no default marked: got %d, want 0", got)
	}

	second := []ImageSource{{Name: "a"}, {Name: "b", Default: true}, {Name: "c"}}
	if got := DefaultSourceIndex(second); got != 1 {
		t.Errorf("second marked: got %d, want 1", got)
	}
}
