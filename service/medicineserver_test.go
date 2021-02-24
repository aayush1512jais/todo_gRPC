package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.com/aayushjaiswal/meds/pb"
	"gitlab.com/aayushjaiswal/meds/sample"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestCreateMedicine(t *testing.T) {
	medicineNoID := sample.NewMedicine()
	medicineNoID.Id = ""

	medicineInvalidID := sample.NewMedicine()
	medicineInvalidID.Id = "invalid-uuid"

	medicineDuplicateID := sample.NewMedicine()
	storeDuplicateID := NewInMemoryMedicineStore()
	err := storeDuplicateID.Save(medicineDuplicateID)
	assert.Nil(t, err)

	cases := map[string]struct {
		medicine *pb.Medicine
		store    MedicineStore
		code     codes.Code
	}{
		"when medicine is created successfully with id": {
			medicine: sample.NewMedicine(),
			store:    NewInMemoryMedicineStore(),
			code:     codes.OK,
		},
		"when medicine is created successfully without id": {
			medicine: medicineNoID,
			store:    NewInMemoryMedicineStore(),
			code:     codes.OK,
		},
		"when medicine id is invalid": {
			medicine: medicineInvalidID,
			store:    NewInMemoryMedicineStore(),
			code:     codes.InvalidArgument,
		},
		"when duplicate id is being used": {
			medicine: medicineDuplicateID,
			store:    storeDuplicateID,
			code:     codes.AlreadyExists,
		},
	}
	for k, v := range cases {
		t.Run(k, func(t *testing.T) {
			req := &pb.CreateMedicineRequest{
				Med: v.medicine,
			}
			server := NewMedicineServer(v.store, nil)
			res, err := server.CreateMedicine(context.Background(), req)
			if v.code == codes.OK {
				assert.NoError(t, err)
				assert.NotNil(t, res)
				assert.NotEmpty(t, res.Id)
				if len(v.medicine.Id) > 0 {
					assert.Equal(t, v.medicine.Id, res.Id)
				}
			} else {
				assert.Error(t, err)
				assert.Nil(t, res)
				st, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, v.code, st.Code())
			}
		})
	}
}
