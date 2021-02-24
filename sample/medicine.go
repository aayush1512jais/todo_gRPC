package sample

import "gitlab.com/aayushjaiswal/meds/pb"

// NewMedicine ...
func NewMedicine() *pb.Medicine {
	brand := randomMedicineBrand()
	name := randomMedicineName(brand)
	medicine := &pb.Medicine{
		Id:       randomID(),
		Name:     name,
		Brand:    brand,
		Quantity: randomQuantity(),
	}
	return medicine
}
