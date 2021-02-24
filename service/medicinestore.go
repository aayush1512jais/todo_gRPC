package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/jinzhu/copier"
	"gitlab.com/aayushjaiswal/meds/pb"
)

// ErrAlreadyExists is returned when a record with the same ID already exists in the store
var ErrAlreadyExists = errors.New("record already exists")

// MedicineStore is an interface to store medicine
type MedicineStore interface {
	// Save saves the medicine to the store
	Save(medicine *pb.Medicine) error
	// Find finds a medicine by ID
	Find(id string) (*pb.Medicine, error)
	// Search searches for medicines with filter, returns one by one via the found function
	Search(ctx context.Context, filter *pb.Filter, found func(medicine *pb.Medicine) error) error
}

// InMemoryMedicineStore stores medicine in memory
type InMemoryMedicineStore struct {
	mutex sync.RWMutex
	data  map[string]*pb.Medicine
}

// NewInMemoryMedicineStore returns a new InMemoryMedicineStore
func NewInMemoryMedicineStore() *InMemoryMedicineStore {
	return &InMemoryMedicineStore{
		data: make(map[string]*pb.Medicine),
	}
}

// Save saves the medicine to the store
func (store *InMemoryMedicineStore) Save(medicine *pb.Medicine) error {
	if store.data[medicine.Id] != nil {
		return ErrAlreadyExists
	}

	clonedMed, err := clone(medicine)
	if err != nil {
		return err
	}
	store.mutex.Lock()
	store.data[clonedMed.Id] = clonedMed
	store.mutex.Unlock()
	return nil
}

// Find finds a medicine by ID
func (store *InMemoryMedicineStore) Find(id string) (*pb.Medicine, error) {
	store.mutex.RLock()
	medicine := store.data[id]
	store.mutex.RUnlock()

	if medicine == nil {
		return nil, nil
	}
	return clone(medicine)
}

// Search searches for medicines with filter, returns one by one via the found function
func (store *InMemoryMedicineStore) Search(
	ctx context.Context,
	filter *pb.Filter,
	found func(medicine *pb.Medicine) error,
) error {
	store.mutex.RLock()
	for _, medicine := range store.data {
		if ctx.Err() == context.Canceled || ctx.Err() == context.DeadlineExceeded {
			log.Print("context is cancelled")
			return nil
		}

		// time.Sleep(time.Second)
		// log.Print("checking medicine id: ", medicine.GetId())

		if isQualified(filter, medicine) {
			clonedMed, err := clone(medicine)
			if err != nil {
				return err
			}
			err = found(clonedMed)
			if err != nil {
				return err
			}
		}
	}
	store.mutex.RUnlock()

	return nil
}

func isQualified(filter *pb.Filter, medicine *pb.Medicine) bool {
	if strings.Compare(filter.GetBrand(), medicine.GetBrand()) != 0 {
		return false
	}
	return true
}

func clone(medicine *pb.Medicine) (*pb.Medicine, error) {
	clonedMed := &pb.Medicine{}

	err := copier.Copy(clonedMed, medicine)
	if err != nil {
		return nil, fmt.Errorf("cannot copy medicine data: %w", err)
	}
	return clonedMed, nil
}
