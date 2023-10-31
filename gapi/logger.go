package gapi

import (
	"context"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"time"
)

func GrpcLogger(ctx context.Context,
	req any,
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (resp any, err error) {
	//get the time started
	startTime := time.Now()
	//get the result and err from handler
	result, err := handler(ctx, req)
	//compute the duration
	duration := time.Since(startTime)
	//initialize the statusCode is unknown
	statusCode := codes.Unknown
	//get the status from err
	if st, ok := status.FromError(err); ok {
		statusCode = st.Code()
	}
	logger := log.Info()
	//if err is not nil,convert the level of log to error
	if err != nil {
		logger = log.Error().Err(err)
	}
	logger.Str("protocol", "gRPC").
		Str("method", info.FullMethod).
		Int("status_code", int(statusCode)).
		Str("status_text", statusCode.String()).
		Dur("duration", duration).
		Msg("receive a gRPC request")
	return result, err
}
