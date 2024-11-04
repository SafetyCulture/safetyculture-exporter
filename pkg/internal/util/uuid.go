package util

import (
	"github.com/gofrs/uuid"
	"strings"
)

// ConvertS12ToUUID - attempt to convert a s12id to UUID
func ConvertS12ToUUID(s string) uuid.UUID {
	maybeUUID, err := uuid.FromString(s)
	if err == nil {
		return maybeUUID
	}

	idx := strings.LastIndex(s, "_")
	if idx == -1 {
		return uuid.Nil
	}
	return uuid.FromStringOrNil(s[idx+1:])
}
