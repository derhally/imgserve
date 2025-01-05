package internal

import (
	"image/jpeg"
	"log/slog"
	"strconv"

	"github.com/gofiber/fiber/v3"
)

type ImageHandlerInterface interface {
	HandleRequest(c fiber.Ctx) error
}

type ImageHandler struct {
	settings     ServiceSettings
	imageStorage ImageStorageInterface
	imagePicker  ImagePickerInterface
}

func NewImageHandler(settings ServiceSettings, imageStorage ImageStorageInterface, imagePicker ImagePickerInterface) *ImageHandler {
	return &ImageHandler{
		settings:     settings,
		imageStorage: imageStorage,
		imagePicker:  imagePicker,
	}
}

func (p *ImageHandler) HandleRequest(c fiber.Ctx) error {

	// pick a random photo
	imageKey := p.imagePicker.Image()
	if imageKey == "" {
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to pick a photo")
	}

	slog.Debug("Serving image", "image", imageKey)

	width, err := strconv.Atoi(c.Params("width"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid width")
	}

	height, err := strconv.Atoi(c.Params("height", "0"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid height")
	}

	blur, err := strconv.ParseFloat(c.Query("blur", "0"), 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid blur")
	}

	grayscale, err := strconv.ParseBool(c.Query("grayscale", "false"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid grayscale")
	}

	if grayscale == false {
		grayscale, err = strconv.ParseBool(c.Query("greyscale", "false"))
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("Invalid greyscale")
		}
	}

	resizeMode := c.Query("resizemode", "fit")
	if resizeMode != "none" && resizeMode != "fill" && resizeMode != "fit" {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid resizemode parameter. Must be none, fill, or fit")
	}

	imageSettings := ImageSettings{
		Width:      width,
		Height:     height,
		Blur:       blur,
		Grayscale:  grayscale,
		ResizeMode: resizeMode,
	}

	slog.Debug("settings", "width", imageSettings.Width,
		"height", imageSettings.Height,
		"grayscale", imageSettings.Grayscale,
		"blur", imageSettings.Blur,
		"resizeMode", imageSettings.ResizeMode)

	img, err := p.imageStorage.ImageWithTransform(imageKey, imageSettings)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to resize image")
	}

	c.Set(fiber.HeaderContentType, MIMEImageJpeg)
	jpeg.Encode(c, img, nil)
	return c.SendStatus(fiber.StatusOK)
}
