package gapi

import (
	"context"
	"fmt"
	"github.com/GGjahoon/MySimpleBank/token"
	"google.golang.org/grpc/metadata"
	"strings"
)

const (
	authorizationHeader     = "authorization"
	authorizationTypeBearer = "bearer"
)

func (server *Server) authorizeUser(ctx context.Context, accessRoles []string) (*token.Payload, error) {
	//get the metadata in ctx
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, fmt.Errorf("missing metadata")
	}
	//get the value of key:authorizationHeader(authorization)
	values := md.Get(authorizationHeader)
	if len(values) == 0 {
		return nil, fmt.Errorf("missing authorization header")
	}
	//the authHeader should be the first parameter in values
	//the authHeader is : <authorizationType>space<Token>
	authHeader := values[0]
	//convert the authHeader(string format) to the slice
	fields := strings.Fields(authHeader)
	//check the format of authHeader
	if len(fields) < 2 {
		return nil, fmt.Errorf("incorrect authorization format")
	}
	if authorizationTypeBearer != strings.ToLower(fields[0]) {
		return nil, fmt.Errorf("cannot support the authorize type %s", fields[0])
	}
	accessToken := fields[1]
	payload, err := server.tokenMaker.VerifyToken(accessToken)
	if err != nil {
		return nil, fmt.Errorf("token is invalid")
	}
	if !hasPermission(payload.Role, accessRoles) {
		return nil, fmt.Errorf("permission denied")
	}
	return payload, nil
}
func hasPermission(userRole string, accessRoles []string) bool {
	for _, role := range accessRoles {
		if userRole == role {
			return true
		}
	}
	return false
}
