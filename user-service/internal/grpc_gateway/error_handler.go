package grpc_gateway

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ErrorHandler handles errors from gRPC services
func ErrorHandler(
	ctx context.Context,
	_ *runtime.ServeMux,
	_ runtime.Marshaler,
	w http.ResponseWriter,
	_ *http.Request,
	err error,
) {
	s, ok := status.FromError(err)
	if !ok {
		s = status.New(codes.Unknown, err.Error())
	}

	w.Header().Set("Content-Type", "application/json")
	statusCode := grpcToHTTPStatus(s.Code())
	w.WriteHeader(statusCode)

	errorResponse := map[string]interface{}{
		"error": map[string]interface{}{
			"code":    int(s.Code()),
			"message": s.Message(),
			"details": s.Details(),
		},
	}

	if jsonErr := json.NewEncoder(w).Encode(errorResponse); jsonErr != nil {
		// If we can't encode the error response, write a simple error
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// grpcToHTTPStatus converts gRPC status codes to HTTP status codes
func grpcToHTTPStatus(code codes.Code) int {
	switch code {
	case codes.InvalidArgument:
		return http.StatusBadRequest
	case codes.NotFound:
		return http.StatusNotFound
	case codes.AlreadyExists:
		return http.StatusConflict
	case codes.PermissionDenied:
		return http.StatusForbidden
	case codes.Unauthenticated:
		return http.StatusUnauthorized
	case codes.ResourceExhausted:
		return http.StatusTooManyRequests
	case codes.FailedPrecondition:
		return http.StatusBadRequest
	case codes.Aborted:
		return http.StatusConflict
	case codes.OutOfRange:
		return http.StatusBadRequest
	case codes.Unimplemented:
		return http.StatusNotImplemented
	case codes.Internal:
		return http.StatusInternalServerError
	case codes.Unavailable:
		return http.StatusServiceUnavailable
	case codes.DataLoss:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}
