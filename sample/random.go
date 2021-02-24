package sample

import (
	"math/rand"
	"time"

	"github.com/google/uuid"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func randomStringFromSet(a ...string) string {
	n := len(a)
	if n == 0 {
		return ""
	}
	return a[rand.Intn(n)]
}

func randomInt(min, max int) int {
	return min + rand.Int()%(max-min+1)
}

func randomID() string {
	return uuid.New().String()
}

func randomMedicineBrand() string {
	return randomStringFromSet("Cipla", "Morpheus", "AYQ")
}
func randomQuantity() uint32 {
	return uint32(randomInt(0, 100))
}
func randomMedicineName(brand string) string {
	switch brand {
	case "Cipla":
		return randomStringFromSet("Crocin", "Combiflame")
	case "AYQ":
		return randomStringFromSet("Hearteo", "Delzium", "VSAR")
	default:
		return randomStringFromSet("Handplast", "Tape", "Nobel")
	}
}
