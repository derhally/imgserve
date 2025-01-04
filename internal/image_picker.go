package internal

import (
	"math/rand"
)

type ImagePickerInterface interface {
	Image() (imageKey string)
}

func NewRandomImagePicker(storage ImageStorageInterface) *RandomImagePicker {

	keys := storage.Keys()
	return &RandomImagePicker{
		imageStorage: storage,
		keys:         keys,
	}
}

type RandomImagePicker struct {
	imageStorage ImageStorageInterface
	keys         []string
}

func (r *RandomImagePicker) Image() string {
	if len(r.keys) == 0 {
		return ""
	}
	return r.keys[rand.Intn(len(r.keys))]
}
