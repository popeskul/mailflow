package grpc_gateway

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc/status"

	"github.com/popeskul/email-service-platform/logger"
	"github.com/popeskul/email-service-platform/user-service/internal/grpc"
)

func errorHandler(ctx context.Context, _ *runtime.ServeMux, _ runtime.Marshaler, w http.ResponseWriter, _ *http.Request, err error) {
	l := ctx.Value("logger").(grpc.Logger)

	code := runtime.HTTPStatusFromCode(status.Code(err))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	if encErr := json.NewEncoder(w).Encode(map[string]string{
		"error": err.Error(),
	}); encErr != nil {
		l.Error("failed to encode error response",
			logger.Field{Key: "encode_error", Value: encErr},
			logger.Field{Key: "original_error", Value: err},
		)
	}
}
