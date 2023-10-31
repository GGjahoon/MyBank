package gapi

import (
	"context"
	"google.golang.org/grpc"
)

func gRPCLogger(ctx context.Context,
	req any,
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (resp any, err error) {

	return nil, nil
}
