package grpc_gateway

import "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"

func headerMatcher(key string) (string, bool) {
	switch key {
	case "Accept", "Content-Type":
		return key, true
	default:
		return runtime.DefaultHeaderMatcher(key)
	}
}
