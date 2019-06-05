package utility

import (
	"math/rand"
)

func GenerateID() uint64 {
	return rand.Uint64()
}
