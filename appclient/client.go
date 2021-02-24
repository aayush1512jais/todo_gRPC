package appclient

import (
	"bufio"
	"context"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"gitlab.com/aayushjaiswal/meds/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// MedicineClient is a client to call medicine service RPCs
type MedicineClient struct {
	service pb.MedicineServiceClient
}

// NewMedicineClient returns a new medicine client
func NewMedicineClient(cc *grpc.ClientConn) *MedicineClient {
	return &MedicineClient{
		service: pb.NewMedicineServiceClient(cc),
	}
}

// CreateMedicine calls create medicine RPC
func (medicineClient *MedicineClient) CreateMedicine(medicine *pb.Medicine) {
	req := &pb.CreateMedicineRequest{
		Med: medicine,
	}

	// set timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := medicineClient.service.CreateMedicine(ctx, req)
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.AlreadyExists {
			log.Print("medicine already exists")
		} else {
			log.Fatal("cannot create medicine: ", err)
		}
		return
	}

	log.Printf("created medicine with id: %s", res.Id)
}

// SearchMedicine calls search medicine RPC
func (medicineClient *MedicineClient) SearchMedicine(filter *pb.Filter) {
	log.Print("search filter: ", filter)

	//ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	//	defer cancel()

	req := &pb.SearchMedicineRequest{Filter: filter}
	stream, err := medicineClient.service.SearchMedicine(context.Background(), req)
	if err != nil {
		log.Fatal("cannot search medicine: ", err)
	}

	for {
		res, err := stream.Recv()
		if err == io.EOF {
			return
		}
		if err != nil {
			log.Fatal("cannot receive response: ", err)
		}

		medicine := res.GetMed()
		log.Print("- found: ", medicine.GetId())
		log.Print("  + name: ", medicine.GetName())
		log.Print("  + brand: ", medicine.GetBrand())
		log.Print("  + Quantity: ", medicine.GetQuantity())
	}
}

// UploadImage calls upload image RPC
func (medicineClient *MedicineClient) UploadImage(medicineID string, imagePath string) {
	file, err := os.Open(imagePath)
	if err != nil {
		log.Fatal("cannot open image file: ", err)
	}
	defer file.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stream, err := medicineClient.service.UploadImage(ctx)
	if err != nil {
		log.Fatal("cannot upload image: ", err)
	}

	req := &pb.UploadImageRequest{
		Data: &pb.UploadImageRequest_Info{
			Info: &pb.ImageInfo{
				MedicineID: medicineID,
				ImageType:  filepath.Ext(imagePath),
			},
		},
	}

	err = stream.Send(req)
	if err != nil {
		log.Fatal("cannot send image info to server: ", err, stream.RecvMsg(nil))
	}

	reader := bufio.NewReader(file)
	buffer := make([]byte, 1024)

	for {
		n, err := reader.Read(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal("cannot read chunk to buffer: ", err)
		}

		req := &pb.UploadImageRequest{
			Data: &pb.UploadImageRequest_ChunkData{
				ChunkData: buffer[:n],
			},
		}

		err = stream.Send(req)
		if err != nil {
			log.Fatal("cannot send chunk to server: ", err, stream.RecvMsg(nil))
		}
	}

	res, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatal("cannot receive response: ", err)
	}

	log.Printf("image uploaded with id: %s, size: %d", res.GetId(), res.GetSize())
}
