package internal

import (
	"fmt"
	"image"
	"iter"
	"os"
)

type ImageStorageInterface interface {
	Images() iter.Seq[string]

	Image(key string) (image.Image, error)

	ImageWithTransform(key string, settings ImageSettings) (image.Image, error)

	MimeType(key string) string

	Contains(key string) bool

	Add(key string, mimeType string, img image.Image) error

	ImageCount() int

	Keys() []string

	Empty() bool

	Clear() error
}

func NewImageStorage(storageType string, imageTransformer ImageTransformerInterface, path string) (ImageStorageInterface, error) {

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("directory does not exist: %s", path)
	}

	if storageType == ImageStoreTypeLocal {
		diskStore := &ImageStorageDisk{location: path}
		err := diskStore.LoadImages()
		return diskStore, err
	}

	return nil, fmt.Errorf("unsupported storage type: %s", storageType)
}
