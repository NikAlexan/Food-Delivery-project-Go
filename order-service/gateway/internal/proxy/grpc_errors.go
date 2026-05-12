package proxy

import (
	"net/http"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var genericMessages = map[codes.Code]string{
	codes.NotFound:           "resource not found",
	codes.AlreadyExists:      "resource already exists",
	codes.Unauthenticated:    "unauthenticated",
	codes.PermissionDenied:   "permission denied",
	codes.InvalidArgument:    "invalid request",
	codes.FailedPrecondition: "operation not allowed",
}

func grpcError(err error) (int, string) {
	s, ok := status.FromError(err)
	if !ok {
		return http.StatusInternalServerError, "internal server error"
	}
	if msg, found := genericMessages[s.Code()]; found {
		return grpcToHTTP(s.Code()), msg
	}
	return http.StatusInternalServerError, "internal server error"
}

func grpcToHTTP(code codes.Code) int {
	switch code {
	case codes.NotFound:
		return http.StatusNotFound
	case codes.AlreadyExists:
		return http.StatusConflict
	case codes.Unauthenticated:
		return http.StatusUnauthorized
	case codes.PermissionDenied:
		return http.StatusForbidden
	case codes.InvalidArgument:
		return http.StatusBadRequest
	case codes.FailedPrecondition:
		return http.StatusUnprocessableEntity
	default:
		return http.StatusInternalServerError
	}
}
