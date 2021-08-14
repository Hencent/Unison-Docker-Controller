package uuid

import (
	"github.com/gofrs/uuid"
)

func GenerateRandomUUID() string {
	id, _ := uuid.NewV4()
	return id.String()
}
