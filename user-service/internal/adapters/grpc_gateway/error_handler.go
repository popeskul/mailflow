package grpc_gateway

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.uber.org/zap"
	"google.golang.org/grpc/status"
)

func errorHandler(ctx context.Context, _ *runtime.ServeMux, _ runtime.Marshaler, w http.ResponseWriter, _ *http.Request, err error) {
	logger := ctx.Value("logger").(*zap.Logger)

	code := runtime.HTTPStatusFromCode(status.Code(err))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	if encErr := json.NewEncoder(w).Encode(map[string]string{
		"error": err.Error(),
	}); encErr != nil {
		logger.Error("failed to encode error response",
			zap.Error(encErr),
			zap.Error(err),
		)
	}
}
