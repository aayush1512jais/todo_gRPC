package main

import (
	"flag"
	"log"

	"gitlab.com/aayushjaiswal/meds/appclient"
	"gitlab.com/aayushjaiswal/meds/pb"
	"gitlab.com/aayushjaiswal/meds/sample"
	"google.golang.org/grpc"
)

func testCreateMedicine(medicineClient *appclient.MedicineClient) {
	medicineClient.CreateMedicine(sample.NewMedicine())
}

func testSearchMedicine(medicineClient *appclient.MedicineClient) {
	for i := 0; i < 10; i++ {
		medicineClient.CreateMedicine(sample.NewMedicine())
	}
	filter := &pb.Filter{
		Brand: "Cipla",
	}
	medicineClient.SearchMedicine(filter)
}

func testUploadImage(medicineClient *appclient.MedicineClient) {
	medicine := sample.NewMedicine()
	medicineClient.CreateMedicine(medicine)
	medicineClient.UploadImage(medicine.GetId(), "tmp/medicine.jpg")
}

func main() {
	serverAddress := flag.String("address", "0.0.0.0:8000", "the server address")
	flag.Parse()
	log.Printf("dial server %s", *serverAddress)

	cc, err := grpc.Dial(*serverAddress, grpc.WithInsecure())
	if err != nil {
		log.Fatal("cannot dial server: ", err)
	}

	medicineClient := appclient.NewMedicineClient(cc)
	//testCreateMedicine(medicineClient)
	//testSearchMedicine(medicineClient)
	testUploadImage(medicineClient)
}
