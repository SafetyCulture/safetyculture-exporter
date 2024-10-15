package util

import (
	"github.com/gofrs/uuid"
	"strings"
)

func ConvertS12ToUUID(s string) uuid.UUID {
	idx := strings.LastIndex(s, "_")
	if idx == -1 {
		return uuid.Nil
	}
	return uuid.FromStringOrNil(s[idx+1:])
}
