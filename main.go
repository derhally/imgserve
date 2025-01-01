package main

import (
	"flag"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"log"
	"log/slog"
	"math/rand"
	"os"
	"path/filepath"
	"picserve/internal"
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/nao1215/imaging"
)

var settings internal.ServiceSettings
var photos []string

func init() {
	flag.StringVar(&settings.PhotoDir, "photoDir", "./photos", "Directory where photos are stored")
	flag.StringVar(&settings.CertFile, "certFile", "", "Path to the certificate file")
	flag.StringVar(&settings.CertKeyFile, "certKeyFile", "", "Path to the certificate key file")
	flag.StringVar(&settings.LogLevel, "logLevel", "INFO", "Path to the certificate key file")
	flag.StringVar(&settings.Port, "port", "8080", "The port to listen on")
}

func getPhotos(dir string) ([]string, error) {
	var photos []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			photos = append(photos, path)
		}
		return nil
	})
	return photos, err
}

func randomPhoto() string {
	return photos[rand.Intn(len(photos))]
}

func thumbnail(r io.Reader, w io.Writer, mimetype string, settings internal.ImageSettings) error {
	var src image.Image
	var err error
	var dst image.Image

	switch mimetype {
	case "image/jpeg":
		src, err = jpeg.Decode(r)
	case "image/png":
		src, err = png.Decode(r)
	}

	if err != nil {
		return err
	}

	slog.Debug("settings", "width", settings.Width, "height", settings.Height, "grayscale", settings.Grayscale, "blur", settings.Blur, "resizeMode", settings.ResizeMode)

	if settings.ResizeMode == "fill" {
		dst = imaging.Fill(src, settings.Width, settings.Height, imaging.Center, imaging.CatmullRom)
	} else if settings.ResizeMode == "fit" {
		dst = imaging.Fit(src, settings.Width, settings.Height, imaging.CatmullRom)
	} else {
		dst = imaging.Resize(src, settings.Width, settings.Height, imaging.CatmullRom)
	}

	if settings.Grayscale {
		dst = imaging.Grayscale(dst)
	}

	if settings.Blur > 0 {
		dst = imaging.Blur(dst, settings.Blur)
	}

	err = jpeg.Encode(w, dst, nil)
	if err != nil {
		return err
	}

	return nil
}

func resizePhoto(photoPath string, w io.Writer, imageSettings internal.ImageSettings) error {
	file, err := os.Open(photoPath)
	if err != nil {
		return err
	}
	defer file.Close()
	return thumbnail(file, w, "image/jpeg", imageSettings)
}

func photoHandler(c fiber.Ctx) error {

	// pick a random photo
	photo := randomPhoto()
	slog.Debug("Serving photo", "image", photo)

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

	resizeMode := c.Query("resizemode", "fit")
	if resizeMode != "none" && resizeMode != "fill" && resizeMode != "fit" {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid fit parameter. Must be none, fill, or fit")
	}

	imageSettings := internal.ImageSettings{
		Width:      width,
		Height:     height,
		Blur:       blur,
		Grayscale:  grayscale,
		ResizeMode: resizeMode,
	}

	c.Set(fiber.HeaderContentType, "image/jpeg")
	err = resizePhoto(photo, c, imageSettings)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to resize image")
	}

	return c.SendStatus(fiber.StatusOK)
}

func loadPhotos() error {
	var err error
	photos, err = getPhotos(settings.PhotoDir)
	if err != nil {
		return fmt.Errorf("failed to load photos: %v", err)
	}
	if len(photos) == 0 {
		return fmt.Errorf("no photos found in %s", settings.PhotoDir)
	}
	slog.Info("Loaded photos", "directory", settings.PhotoDir, "count", len(photos))
	return nil
}

func setLogLevel() error {
	var level slog.Level
	var err = level.UnmarshalText([]byte(settings.LogLevel))
	if err != nil {
		slog.Error("Failed to set log level", "error", err)
		return err
	}
	slog.SetLogLoggerLevel(level)
	return nil
}

func main() {
	flag.Parse()

	var err error
	var photoDir string
	setLogLevel()

	photoDir, err = filepath.Abs(settings.PhotoDir)
	if err != nil {
		log.Fatalf("Failed to resolve photo directory path: %v", err)
	}

	if _, err := os.Stat(photoDir); os.IsNotExist(err) {
		log.Fatalf("Photo directory does not exist: %s", photoDir)
	}

	err = loadPhotos()
	if err != nil {
		log.Fatalf("Failed to load photos: %v", err)
	}

	app := fiber.New()
	app.Use(recover.New())

	app.Get("/:width<int>", photoHandler)
	app.Get("/:width<int>/:height<int>", photoHandler)

	listeningPort := fmt.Sprintf(":%s", settings.Port)
	slog.Info("Starting server", "port", listeningPort)
	log.Fatal(app.Listen(listeningPort, fiber.ListenConfig{CertFile: settings.CertFile, CertKeyFile: settings.CertKeyFile}))
}
