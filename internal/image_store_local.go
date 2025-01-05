package internal

import (
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"iter"
	"log/slog"
	"mime"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/go-set/v3"
)

type ImageStorageDisk struct {
	location         string
	imageTransformer ImageTransformerInterface
	images           set.Set[string]
}

func (p *ImageStorageDisk) ImageCount() int {
	return p.images.Size()
}

func (p *ImageStorageDisk) Empty() bool {
	return p.images.Empty()
}

func (p *ImageStorageDisk) Clear() error {

	files, err := os.ReadDir(p.location)
	if err != nil {
		slog.Error("Failed to read image directory", "err", err)
		return err
	}
	for _, file := range files {
		path := filepath.Join(p.location, file.Name())
		if err := os.Remove(path); err != nil {
			slog.Error("Failed to delete file", "path", path, "err", err)
			return err
		}
	}
	p.images = *set.New[string](0)
	return nil
}

func (p *ImageStorageDisk) MimeType(key string) string {
	ext := filepath.Ext(key)
	return mime.TypeByExtension(ext)
}

func (p *ImageStorageDisk) Keys() []string {
	return p.images.Slice()
}

func (p *ImageStorageDisk) Images() iter.Seq[string] {
	return p.images.Items()
}

func (p *ImageStorageDisk) Add(key string, mimeType string, img image.Image) error {

	path := filepath.Join(p.location, key)

	file, err := os.Create(path)
	if err != nil {
		return err
	}

	defer file.Close()

	switch mimeType {
	case MIMEImageJpeg:
		err = jpeg.Encode(file, img, nil)
	case MIMEImagePng:
		err = png.Encode(file, img)
	default:
		err = fmt.Errorf("unsupported image format: %s", mimeType)
	}

	if err != nil {
		return err
	}

	p.images.Insert(key)

	return nil
}

func (p *ImageStorageDisk) Contains(key string) bool {
	return p.images.Contains(key)
}

func (p *ImageStorageDisk) Image(key string) (image.Image, error) {
	path := filepath.Join(p.location, key)

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var img image.Image
	mimeType := p.MimeType(path)

	switch mimeType {
	case MIMEImageJpeg:
		img, err = jpeg.Decode(file)
	case MIMEImagePng:
		img, err = png.Decode(file)
	case "":
		err = fmt.Errorf("Unknown image type")
	}

	if err != nil {
		return nil, err
	}

	return img, nil
}

func (p *ImageStorageDisk) ImageWithTransform(key string, settings ImageSettings) (image.Image, error) {

	img, err := p.Image(key)
	if err != nil {
		return nil, err
	}

	bounds := img.Bounds()
	if bounds.Dx() == settings.Width && bounds.Dy() == settings.Height && !settings.Grayscale && settings.Blur == 0 {
		return img, nil
	}

	return p.imageTransformer.Transform(img, settings)
}

func (p *ImageStorageDisk) LoadImages() error {

	if _, err := os.Stat(p.location); os.IsNotExist(err) {
		slog.Error("Image directory does not exist", "directory", p.location)
		return err
	}

	err := filepath.Walk(p.location, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		_, filename := filepath.Split(path)
		if !info.IsDir() {

			fileExt := filepath.Ext(filename)
			if supportedImageFile(fileExt) {
				p.images.Insert(filename)
			}
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to load images: %v", err)
	}

	slog.Info("Loaded images", "directory", p.location, "count", p.images.Size())

	return err
}

func supportedImageFile(fileExt string) bool {
	fileExt = strings.ToLower(fileExt)
	switch fileExt {
	case ".jpg", ".jpeg", ".png":
		return true
	default:
		return false
	}
}
