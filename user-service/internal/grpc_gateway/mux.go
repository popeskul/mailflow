package grpc_gateway

import "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"

func NewGatewayMux() *runtime.ServeMux {
	return runtime.NewServeMux(
		runtime.WithIncomingHeaderMatcher(headerMatcher),
		runtime.WithErrorHandler(ErrorHandler),
	)
}
