package main

import (
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"picserve/internal"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/recover"
)

const defaultImageDir = "./images"

var settings internal.ServiceSettings

// stringList is a flag.Value that collects every use of a repeatable flag.
type stringList []string

func (s *stringList) String() string { return strings.Join(*s, ", ") }

func (s *stringList) Set(value string) error {
	*s = append(*s, value)
	return nil
}

var imageDirs stringList
var configFile string
var defaultDirName string

func init() {
	flag.StringVar(&configFile, "config", "", "Path to a YAML config file. Flags given on the command line override its values, "+
		"and any -imageDir flags replace its directories")
	flag.Var(&imageDirs, "imageDir", "Directory to serve images from, as `dir` or name=dir. "+
		"Repeat to serve several directories; each is served at /{name}/{width}[/{height}], "+
		"where name defaults to the directory's base name. Defaults to "+defaultImageDir)
	flag.StringVar(&defaultDirName, "default", "", "Path name of the directory that also answers /{width}[/{height}]. "+
		"Overrides default: in the config file. When nothing sets it, the first directory is the default")
	flag.StringVar(&settings.CertFile, "certFile", "", "Path to the certificate file")
	flag.StringVar(&settings.CertKeyFile, "certKeyFile", "", "Path to the certificate key file")
	flag.StringVar(&settings.LogLevel, "logLevel", "INFO", "Log level: DEBUG, INFO, WARN or ERROR")
	flag.StringVar(&settings.Port, "port", "8080", "The port to listen on")
	flag.StringVar(&settings.CacheDir, "cacheDir", "", "The directory to store transformed images in. "+
		"Each image directory gets its own subdirectory named after its path name")
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

// applyConfig fills in settings from the config file for every flag that
// wasn't given on the command line, and returns the config's directories.
func applyConfig(cfg internal.Config) ([]internal.ImageSource, error) {
	setFlags := map[string]bool{}
	flag.Visit(func(f *flag.Flag) { setFlags[f.Name] = true })

	apply := func(flagName string, target *string, value string) {
		if !setFlags[flagName] && value != "" {
			*target = value
		}
	}
	apply("port", &settings.Port, cfg.Port)
	apply("logLevel", &settings.LogLevel, cfg.LogLevel)
	apply("cacheDir", &settings.CacheDir, cfg.CacheDir)
	apply("certFile", &settings.CertFile, cfg.CertFile)
	apply("certKeyFile", &settings.CertKeyFile, cfg.CertKeyFile)

	return cfg.ImageSources()
}

func checkFlags(configSources []internal.ImageSource) (sources []internal.ImageSource, cacheDir string, err error) {
	switch {
	case len(settings.ImageDirs) > 0:
		for _, value := range settings.ImageDirs {
			source, err := internal.ParseImageSource(value)
			if err != nil {
				slog.Error("Invalid image directory", "value", value, "error", err)
				return nil, "", err
			}
			sources = append(sources, source)
		}
	case len(configSources) > 0:
		sources = configSources
	default:
		source, err := internal.ParseImageSource(defaultImageDir)
		if err != nil {
			return nil, "", err
		}
		sources = []internal.ImageSource{source}
	}

	if defaultDirName != "" {
		found := false
		for i := range sources {
			sources[i].Default = strings.EqualFold(sources[i].Name, defaultDirName)
			found = found || sources[i].Default
		}
		if !found {
			slog.Error("The -default directory is not being served", "default", defaultDirName)
			return nil, "", fmt.Errorf("no directory has the path name %q", defaultDirName)
		}
	}

	if err := internal.ValidateImageSources(sources); err != nil {
		slog.Error("Invalid image directories", "error", err)
		return nil, "", err
	}

	if settings.CacheDir != "" {
		cacheDir, err = filepath.Abs(settings.CacheDir)
		if err != nil {
			slog.Error("Failed to resolve cache directory", "error", err)
			return nil, "", err
		}
	} else {
		slog.Info("Cache directory not set. Caching is DISABLED.")
	}

	for _, source := range sources {
		if cacheDir != "" && source.Dir == cacheDir {
			slog.Error("Image directory and cache directory cannot be the same", "directory", source.Dir)
			return nil, "", fmt.Errorf("image directory and cache directory cannot be the same")
		}
	}

	return sources, cacheDir, nil
}

// newSourceHandler builds the storage, optional cache, picker and handler for one image directory.
func newSourceHandler(source internal.ImageSource, cacheDir string, transformer internal.ImageTransformerInterface) (*internal.ImageHandler, int, error) {
	imageStorage, err := internal.NewImageStorage(internal.ImageStoreTypeLocal, transformer, source.Dir)
	if err != nil {
		return nil, 0, fmt.Errorf("error creating image storage for %s: %w", source.Dir, err)
	}

	var store internal.ImageStorageInterface = imageStorage
	if cacheDir != "" {
		// Separate cache per source so images with the same file name in two
		// directories never share cache entries.
		sourceCacheDir := filepath.Join(cacheDir, source.Name)
		if err := os.MkdirAll(sourceCacheDir, 0o755); err != nil {
			return nil, 0, fmt.Errorf("error creating cache directory %s: %w", sourceCacheDir, err)
		}
		cacheStorage, err := internal.NewImageStorage(internal.ImageStoreTypeLocal, transformer, sourceCacheDir)
		if err != nil {
			return nil, 0, fmt.Errorf("error creating cache storage for %s: %w", sourceCacheDir, err)
		}
		store, err = internal.NewImageCacheLocal(imageStorage, cacheStorage)
		if err != nil {
			return nil, 0, fmt.Errorf("error creating image cache for %s: %w", source.Dir, err)
		}
	}

	picker := internal.NewRandomImagePicker(imageStorage)
	return internal.NewImageHandler(settings, store, picker), imageStorage.ImageCount(), nil
}

func main() {
	flag.Parse()
	settings.ImageDirs = imageDirs

	var configSources []internal.ImageSource
	if configFile != "" {
		cfg, err := internal.LoadConfig(configFile)
		if err != nil {
			slog.Error("Invalid config file", "error", err)
			os.Exit(1)
		}
		configSources, err = applyConfig(cfg)
		if err != nil {
			slog.Error("Invalid directory in config file", "file", configFile, "error", err)
			os.Exit(1)
		}
	}

	setLogLevel()
	if configFile != "" {
		slog.Info("Loaded config file", "file", configFile)
	}

	sources, cacheDir, err := checkFlags(configSources)
	if err != nil {
		os.Exit(1)
	}

	imageTransformer := internal.NewImageTransfomer()

	app := fiber.New()
	app.Use(recover.New())

	defaultIndex := internal.DefaultSourceIndex(sources)
	for i, source := range sources {
		handler, count, err := newSourceHandler(source, cacheDir, imageTransformer)
		if err != nil {
			log.Fatal(err)
		}

		prefix := "/" + source.Name
		app.Get(prefix+"/:width<int>", handler.HandleRequest)
		app.Get(prefix+"/:width<int>/:height<int>", handler.HandleRequest)
		slog.Info("Serving image directory", "path", prefix, "directory", source.Dir, "images", count)

		// Keep the original /{width} and /{width}/{height} URLs working by
		// pointing them at the default directory.
		if i == defaultIndex {
			app.Get("/:width<int>", handler.HandleRequest)
			app.Get("/:width<int>/:height<int>", handler.HandleRequest)
			slog.Info("Serving default image directory", "path", "/", "directory", source.Dir)
		}
	}

	listeningPort := fmt.Sprintf(":%s", settings.Port)
	slog.Info("Starting server", "port", listeningPort)
	log.Fatal(app.Listen(listeningPort, fiber.ListenConfig{CertFile: settings.CertFile, CertKeyFile: settings.CertKeyFile}))
}
