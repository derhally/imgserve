package main

import (
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"picserve/internal"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/recover"
)

var settings internal.ServiceSettings

func init() {
	flag.StringVar(&settings.ImageDir, "imageDir", "./images", "Directory where images are stored")
	flag.StringVar(&settings.CertFile, "certFile", "", "Path to the certificate file")
	flag.StringVar(&settings.CertKeyFile, "certKeyFile", "", "Path to the certificate key file")
	flag.StringVar(&settings.LogLevel, "logLevel", "INFO", "Path to the certificate key file")
	flag.StringVar(&settings.Port, "port", "8080", "The port to listen on")
	flag.StringVar(&settings.CacheDir, "cacheDir", "", "The directory to store temporary files")
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

func resolvePath(path string) (string, error) {
	imageDir, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}

	return imageDir, nil
}

func main() {
	flag.Parse()

	var err error
	setLogLevel()

	imageDir, err := resolvePath(settings.ImageDir)
	if err != nil {
		slog.Error("Failed to resolve image directory", "error", err)
		os.Exit(1)
	}

	cacheDir, err := resolvePath(settings.CacheDir)
	if err != nil {
		slog.Error("Failed to resolve cache directory", "error", err)
		os.Exit(1)
	}

	if imageDir == cacheDir {
		slog.Error("Image directory and cache directory cannot be the same")
		os.Exit(1)
	}

	imageTransformer := internal.NewImageTransfomer()
	imageStorage, err := internal.NewImageStorage(internal.ImageStoreTypeLocal, imageTransformer, imageDir)
	if err != nil {
		log.Fatalf("Error creating image storage: %v", err)
	}

	imageCacheStorage, err := internal.NewImageStorage(internal.ImageStoreTypeLocal, imageTransformer, cacheDir)
	imageCache, err := internal.NewImageCacheLocal(imageStorage, imageCacheStorage)
	if err != nil {
		log.Fatalf("Error creating image storage: %v", err)
	}

	imagePicker := internal.NewRandomImagePicker(imageStorage)
	imageHandler := internal.NewImageHandler(settings, imageCache, imagePicker)

	app := fiber.New()
	app.Use(recover.New())

	app.Get("/:width<int>", imageHandler.HandleRequest)
	app.Get("/:width<int>/:height<int>", imageHandler.HandleRequest)

	listeningPort := fmt.Sprintf(":%s", settings.Port)
	slog.Info("Starting server", "port", listeningPort)
	log.Fatal(app.Listen(listeningPort, fiber.ListenConfig{CertFile: settings.CertFile, CertKeyFile: settings.CertKeyFile}))
}
