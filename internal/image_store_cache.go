package internal

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"image"
	"iter"
	"log/slog"
)

type ImageStoreCacheLocal struct {
	imageStore ImageStorageInterface
	cacheStore ImageStorageInterface
}

func NewImageCacheLocal(imageStore ImageStorageInterface, cacheStore ImageStorageInterface) (*ImageStoreCacheLocal, error) {
	return &ImageStoreCacheLocal{
		imageStore: imageStore,
		cacheStore: cacheStore,
	}, nil
}

func (c *ImageStoreCacheLocal) Keys() []string {
	return c.imageStore.Keys()
}

func (c *ImageStoreCacheLocal) Images() iter.Seq[string] {
	return c.imageStore.Images()
}

func (c *ImageStoreCacheLocal) Image(key string) (image.Image, error) {
	return c.imageStore.Image(key)
}

func (c *ImageStoreCacheLocal) Empty() bool {
	c.cacheStore.Empty()
	return c.imageStore.Empty()
}

func (c *ImageStoreCacheLocal) ClearCache() error {
	return c.cacheStore.Clear()
}

func (c *ImageStoreCacheLocal) Clear() error {
	err := c.ClearCache()
	if err != nil {
		return err
	}
	return c.imageStore.Clear()
}

func (c *ImageStoreCacheLocal) ImageCount() int {
	return c.imageStore.ImageCount()
}

func (c *ImageStoreCacheLocal) MimeType(key string) string {
	return c.imageStore.MimeType(key)
}

func (c *ImageStoreCacheLocal) Contains(key string) bool {
	return c.imageStore.Contains(key)
}

func (c *ImageStoreCacheLocal) Add(key string, mimeType string, img image.Image) error {
	return c.imageStore.Add(key, mimeType, img)
}

func (c *ImageStoreCacheLocal) ImageWithTransform(key string, imageSettings ImageSettings) (image.Image, error) {

	var targetMimeType = MIMEImageJpeg
	fileEx := fileExtFromMimeType(targetMimeType)
	cacheKey := fmt.Sprintf("n:%s_w:%d_h:%d_b:%f_g:%t_m:%s%s", key,
		imageSettings.Width,
		imageSettings.Height,
		imageSettings.Blur,
		imageSettings.Grayscale,
		imageSettings.ResizeMode,
		fileEx,
	)

	hashedName := fmt.Sprintf("%s%s", hashString(cacheKey), fileEx)
	if c.cacheStore.Contains(hashedName) {
		slog.Debug("Cache hit", "key", key, "cacheKey", cacheKey)
		return c.cacheStore.Image(hashedName)
	}

	img, err := c.imageStore.ImageWithTransform(key, imageSettings)
	if err != nil {
		return nil, err
	}

	if err := c.cacheStore.Add(hashedName, targetMimeType, img); err != nil {
		// A failed cache write shouldn't fail the request.
		slog.Warn("Failed to cache image", "key", key, "err", err)
	}
	return img, nil
}

func hashString(str string) string {
	hasher := sha256.New()
	hasher.Write([]byte(str))
	hashBytes := hasher.Sum(nil)
	return hex.EncodeToString(hashBytes)
}
