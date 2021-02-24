package service

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"sync"

	"github.com/google/uuid"
)

// ImageStore ...
type ImageStore interface {
	Save(medicineID string, imageType string, imageData bytes.Buffer) (string, error)
}

// DiskImageStore stores image on disk, and its info on memory
type DiskImageStore struct {
	mutex       sync.RWMutex
	imageFolder string
	images      map[string]*ImageInfo
}

// ImageInfo contains information of the medicine image
type ImageInfo struct {
	MedicineID string
	Type       string
	Path       string
}

// NewDiskImageStore ...
func NewDiskImageStore(imageFolder string) *DiskImageStore {
	return &DiskImageStore{
		imageFolder: imageFolder,
		images:      make(map[string]*ImageInfo),
	}
}

// Save adds a new image to a medicine
func (store *DiskImageStore) Save(medicineID string, imageType string, imageData bytes.Buffer) (string, error) {
	imageID, err := uuid.NewRandom()
	if err != nil {
		return "", errors.New("cannot generate image id: " + err.Error())
	}

	imagePath := fmt.Sprintf("%s/%s%s", store.imageFolder, imageID, imageType)

	file, err := os.Create(imagePath)
	if err != nil {
		return "", errors.New("cannot create image file: " + err.Error())
	}

	_, err = imageData.WriteTo(file)
	if err != nil {
		return "", errors.New("cannot write image to file: " + err.Error())
	}

	store.mutex.Lock()
	store.images[imageID.String()] = &ImageInfo{
		MedicineID: medicineID,
		Type:       imageType,
		Path:       imagePath,
	}
	store.mutex.Unlock()

	return imageID.String(), nil
}
