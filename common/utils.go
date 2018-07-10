package common

import (
	"github.com/oklog/ulid"
	"math/rand"
	"time"
)

func NewULID() ulid.ULID {
	t := time.Now()
	entropy := rand.New(rand.NewSource(t.UnixNano()))

	return ulid.MustNew(ulid.Timestamp(t), entropy)
}
