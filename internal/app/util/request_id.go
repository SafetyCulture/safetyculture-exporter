// Copyright (c) 2018 SafetyCulture Pty Ltd. All Rights Reserved.

package util

import (
	"context"
	"encoding/hex"

	"github.com/gofrs/uuid"
)

// RequestIDFromContext returns the Request ID stored in context or a new RequestID.
func RequestIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return newRequestID()
	}
	if ctxValue := ctx.Value(ContextKeyRequestID); ctxValue != nil {
		if ctxRequestID, ok := ctxValue.(string); ok && ctxRequestID != "" {
			return ctxRequestID
		}
	}

	return newRequestID()
}

func newRequestID() string {
	id := uuid.Must(uuid.NewV4())
	buf := make([]byte, 32)
	hex.Encode(buf, id.Bytes())

	return string(buf)
}
