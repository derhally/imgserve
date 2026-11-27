package internal

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

// ImageSource is a directory of images served under its own URL path.
type ImageSource struct {
	Name string // URL path segment, e.g. "screensaver" for /screensaver/{width}
	Dir  string // Absolute path to the directory holding the images
}

var validSourceName = regexp.MustCompile(`^[A-Za-z0-9._-]+$`)
var allDigits = regexp.MustCompile(`^[0-9]+$`)

// ParseImageSource parses an -imageDir value. The value is either a plain
// directory ("/mnt/media/photos", served as /photos) or "name=directory"
// ("screensaver=/mnt/images/ss", served as /screensaver).
func ParseImageSource(value string) (ImageSource, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return ImageSource{}, fmt.Errorf("image directory must not be empty")
	}

	name, dir, hasName := strings.Cut(value, "=")
	if !hasName {
		dir = value
		name = ""
	}

	if hasName && strings.TrimSpace(name) == "" {
		return ImageSource{}, fmt.Errorf("path name must not be empty in %q", value)
	}

	return NewImageSource(name, dir)
}

// NewImageSource builds an ImageSource from a directory and an optional name.
// An empty name defaults to the directory's base name.
func NewImageSource(name, dir string) (ImageSource, error) {
	dir = strings.TrimSpace(dir)
	if dir == "" {
		return ImageSource{}, fmt.Errorf("image directory must not be empty")
	}

	absDir, err := filepath.Abs(dir)
	if err != nil {
		return ImageSource{}, fmt.Errorf("failed to resolve image directory %q: %w", dir, err)
	}

	name = strings.TrimSpace(name)
	if name == "" {
		name = filepath.Base(absDir)
	}

	if err := validateSourceName(name); err != nil {
		return ImageSource{}, fmt.Errorf("invalid path name for %q: %w", dir, err)
	}

	return ImageSource{Name: name, Dir: absDir}, nil
}

func validateSourceName(name string) error {
	if name == "" {
		return fmt.Errorf("name must not be empty")
	}
	if name == "." || name == ".." {
		return fmt.Errorf("name %q is not allowed", name)
	}
	if !validSourceName.MatchString(name) {
		return fmt.Errorf("name %q may only contain letters, digits, '.', '_' and '-'", name)
	}
	// An all-digit name would clash with the /{width} routes.
	if allDigits.MatchString(name) {
		return fmt.Errorf("name %q must not be only digits", name)
	}
	return nil
}

// ValidateImageSources checks that every source has a unique name.
// Names are compared case-insensitively so /Photos and /photos can't both exist.
func ValidateImageSources(sources []ImageSource) error {
	if len(sources) == 0 {
		return fmt.Errorf("at least one image directory is required")
	}
	seen := make(map[string]string, len(sources))
	for _, s := range sources {
		key := strings.ToLower(s.Name)
		if other, ok := seen[key]; ok {
			return fmt.Errorf("path name %q is used by both %q and %q; give one a name with name=directory", s.Name, other, s.Dir)
		}
		seen[key] = s.Dir
	}
	return nil
}
