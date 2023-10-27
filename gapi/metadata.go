package gapi

import (
	"context"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
)

const (
	grpcGatewayUserAgentHeader = "grpcgateway-user-agent"
	xForwardedForHeader        = "x-forwarded-for"
	grpcUserAgentHeader        = "user-agent"
)

type Metadata struct {
	UserAgent string
	ClientIP  string
}

// extractMetadata to get the userAgent and client ip information in ctx
func (server *Server) extractMetadata(ctx context.Context) *Metadata {
	MtDt := &Metadata{}
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if userAgents := md.Get(grpcGatewayUserAgentHeader); len(userAgents) > 0 {
			MtDt.UserAgent = userAgents[0]
		}
		if clientIPs := md.Get(xForwardedForHeader); len(clientIPs) > 0 {
			MtDt.ClientIP = clientIPs[0]
		}
		if userAgents := md.Get(grpcUserAgentHeader); len(userAgents) > 0 {
			MtDt.UserAgent = userAgents[0]
		}
	}
	if p, ok := peer.FromContext(ctx); ok {
		MtDt.ClientIP = p.Addr.String()
	}
	return MtDt
}
