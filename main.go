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

func checkFlags() (imageDir string, cacheDir string, err error) {
	imageDir, err = resolvePath(settings.ImageDir)
	if err != nil {
		slog.Error("Failed to resolve image directory", "error", err)
		return "", "", err
	}

	if settings.CacheDir != "" {
		cacheDir, err = resolvePath(settings.CacheDir)
		if err != nil {
			slog.Error("Failed to resolve cache directory", "error", err)
			return "", "", err
		}
	} else {
		slog.Info("Cache directory not set. Caching is DISABLED.")
		cacheDir = ""
	}

	if imageDir == cacheDir {
		slog.Error("Image directory and cache directory cannot be the same")
		return "", "", fmt.Errorf("image directory and cache directory cannot be the same")
	}

	return imageDir, cacheDir, nil
}

func main() {
	flag.Parse()

	var err error
	var imageCache internal.ImageStorageInterface

	setLogLevel()

	imageDir, cacheDir, err := checkFlags()
	if err != nil {
		os.Exit(1)
	}

	imageTransformer := internal.NewImageTransfomer()
	imageStorage, err := internal.NewImageStorage(internal.ImageStoreTypeLocal, imageTransformer, imageDir)
	if err != nil {
		log.Fatalf("Error creating image storage: %v", err)
	}

	if cacheDir != "" {
		imageCacheStorage, err := internal.NewImageStorage(internal.ImageStoreTypeLocal, imageTransformer, cacheDir)
		imageCache, err = internal.NewImageCacheLocal(imageStorage, imageCacheStorage)
		if err != nil {
			log.Fatalf("Error creating image storage: %v", err)
		}
	} else {
		// Caching is disabled, so just use the image storage
		imageCache = imageStorage
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
