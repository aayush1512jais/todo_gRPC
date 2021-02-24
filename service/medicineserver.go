package service

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log"

	"github.com/google/uuid"
	"gitlab.com/aayushjaiswal/meds/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const maxImageSize = 1 << 20

// MedicineServer is the m that provides medicine services
type MedicineServer struct {
	medicineStore MedicineStore
	imageStore    ImageStore
}

// NewMedicineServer ....
func NewMedicineServer(medicineStore MedicineStore, imageStore ImageStore) *MedicineServer {
	return &MedicineServer{
		medicineStore: medicineStore,
		imageStore:    imageStore,
	}
}

// CreateMedicine is a unary RPC to create a new medicine
func (m *MedicineServer) CreateMedicine(ctx context.Context, req *pb.CreateMedicineRequest) (*pb.CreateMedicineResponse, error) {
	medicine := req.GetMed()
	log.Println("receive a create-medicine request with id: ", medicine.Id)

	if len(medicine.Id) > 0 {
		// check if it's a valid UUID
		_, err := uuid.Parse(medicine.Id)
		if err != nil {
			return nil, logError(status.Errorf(codes.InvalidArgument, "medicine ID is not a valid UUID: %v", err))
		}
	} else {
		id, err := uuid.NewRandom()
		if err != nil {
			return nil, logError(status.Errorf(codes.Internal, "cannot generate a new medicine ID: %v", err))
		}
		medicine.Id = id.String()
	}

	// some heavy processing
	// time.Sleep(6 * time.Second)

	if err := contextError(ctx); err != nil {
		return nil, err
	}

	// save the medicine to store
	err := m.medicineStore.Save(medicine)
	if err != nil {
		code := codes.Internal
		if errors.Is(err, ErrAlreadyExists) {
			code = codes.AlreadyExists
		}

		return nil, logError(status.Errorf(code, "cannot save medicine to the store: %v", err))
	}

	log.Println("saved medicine with id: ", medicine.Id)

	res := &pb.CreateMedicineResponse{
		Id: medicine.Id,
	}
	return res, nil
}

// SearchMedicine is a server-streaming RPC to search for medicines
func (m *MedicineServer) SearchMedicine(req *pb.SearchMedicineRequest,
	stream pb.MedicineService_SearchMedicineServer) error {

	filter := req.GetFilter()
	log.Println("receive a search-medicine request with filter: ", filter)

	err := m.medicineStore.Search(
		stream.Context(),
		filter,
		func(medicine *pb.Medicine) error {
			res := &pb.SearchMedicineResponse{Med: medicine}
			err := stream.Send(res)
			if err != nil {
				return err
			}

			log.Println("sent medicine with id: ", medicine.GetId())
			return nil
		},
	)

	if err != nil {
		return logError(status.Errorf(codes.Internal, "unexpected error: %v", err))
	}
	return nil
}

// UploadImage is a client-streaming RPC to upload a medicine image
func (m *MedicineServer) UploadImage(stream pb.MedicineService_UploadImageServer) error {
	req, err := stream.Recv()
	if err != nil {
		return logError(status.Errorf(codes.Unknown, "cannot receive image info"))
	}

	medicineID := req.GetInfo().GetMedicineID()
	imageType := req.GetInfo().GetImageType()
	log.Printf("receive an upload-image request for medicine %s with image type %s", medicineID, imageType)

	medicine, err := m.medicineStore.Find(medicineID)
	if err != nil {
		return logError(status.Errorf(codes.Internal, "cannot find medicine: %v", err))
	}
	if medicine == nil {
		return logError(status.Errorf(codes.InvalidArgument, "medicine id %s doesn't exist", medicineID))
	}

	imageData := bytes.Buffer{}
	imageSize := 0

	for {
		err := contextError(stream.Context())
		if err != nil {
			return err
		}

		log.Print("waiting to receive more data")

		req, err := stream.Recv()
		if err == io.EOF {
			log.Print("no more data")
			break
		}
		if err != nil {
			return logError(status.Errorf(codes.Unknown, "cannot receive chunk data: %v", err))
		}

		chunk := req.GetChunkData()
		size := len(chunk)

		log.Printf("received a chunk with size: %d", size)

		imageSize += size
		if imageSize > maxImageSize {
			return logError(status.Errorf(codes.InvalidArgument, "image is too large: %d > %d", imageSize, maxImageSize))
		}

		// write slowly
		// time.Sleep(time.Second)

		_, err = imageData.Write(chunk)
		if err != nil {
			return logError(status.Errorf(codes.Internal, "cannot write chunk data: %v", err))
		}
	}

	imageID, err := m.imageStore.Save(medicineID, imageType, imageData)
	if err != nil {
		return logError(status.Errorf(codes.Internal, "cannot save image to the store: %v", err))
	}

	res := &pb.UploadImageResponse{
		Id:   imageID,
		Size: uint32(imageSize),
	}

	err = stream.SendAndClose(res)
	if err != nil {
		return logError(status.Errorf(codes.Unknown, "cannot send response: %v", err))
	}

	log.Printf("saved image with id: %s, size: %d", imageID, imageSize)
	return nil
}

func contextError(ctx context.Context) error {
	switch ctx.Err() {
	case context.Canceled:
		return logError(status.Error(codes.Canceled, "request is canceled"))
	case context.DeadlineExceeded:
		return logError(status.Error(codes.DeadlineExceeded, "deadline is exceeded"))
	default:
		return nil
	}
}

func logError(err error) error {
	if err != nil {
		log.Print(err)
	}
	return err
}

func (m *MedicineServer) mustEmbedUnimplementedMedicineServiceServer() {}
