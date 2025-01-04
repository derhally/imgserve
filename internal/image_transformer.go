package internal

import (
	"image"
	"log/slog"

	"github.com/nao1215/imaging"
)

type ImageTransformerInterface interface {
	Transform(img image.Image, imageSettings ImageSettings) (image.Image, error)
}

type ImageTransformer struct {
}

func NewImageTransfomer() *ImageTransformer {
	return &ImageTransformer{}
}

func (t *ImageTransformer) Transform(img image.Image, imageSettings ImageSettings) (image.Image, error) {
	slog.Debug("settings", "width", imageSettings.Width, "height", imageSettings.Height, "grayscale", imageSettings.Grayscale, "blur", imageSettings.Blur, "resizeMode", imageSettings.ResizeMode)

	if imageSettings.ResizeMode == "fill" {
		img = imaging.Fill(img, imageSettings.Width, imageSettings.Height, imaging.Center, imaging.CatmullRom)
	} else if imageSettings.ResizeMode == "fit" {
		img = imaging.Fit(img, imageSettings.Width, imageSettings.Height, imaging.CatmullRom)
	} else {
		img = imaging.Resize(img, imageSettings.Width, imageSettings.Height, imaging.CatmullRom)
	}

	if imageSettings.Grayscale {
		img = imaging.Grayscale(img)
	}

	if imageSettings.Blur > 0 {
		img = imaging.Blur(img, imageSettings.Blur)
	}

	return img, nil
}
