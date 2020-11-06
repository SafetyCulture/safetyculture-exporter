// Copyright (c) 2018 SafetyCulture Pty Ltd. All Rights Reserved.

package util

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Defines Errors.
var (
	ErrUnauthenticated  = status.Errorf(codes.Unauthenticated, "Unauthenticated")
	ErrPermissionDenied = status.Errorf(codes.PermissionDenied, "Permission Denied")
	ErrInternal         = status.Errorf(codes.Internal, "Internal Server Error")
	ErrInvalidArgument  = status.Errorf(codes.InvalidArgument, "Invalid Arguments")
	ErrNotFound         = status.Errorf(codes.NotFound, "Item not found")
	ErrAlreadyExists    = status.Errorf(codes.AlreadyExists, "AlreadyExists")
)
