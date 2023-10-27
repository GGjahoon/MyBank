package gapi

import (
	"fmt"
	db "github.com/GGjahoon/MySimpleBank/db/sqlc"
	"github.com/GGjahoon/MySimpleBank/pb"
	"github.com/GGjahoon/MySimpleBank/token"
	"github.com/GGjahoon/MySimpleBank/util"
)

// Server serves gRPC request for bank service
type Server struct {
	pb.UnimplementedSimpleBankServer
	config     util.Config
	store      db.Store
	tokenMaker token.Maker
}

// NewServer create a new gRPC server
func NewServer(config util.Config, store db.Store) (server *Server, err error) {
	tokenMaker, err := token.NewPasetoMaker(config.TokenSymmetricKey)
	if err != nil {
		return nil, fmt.Errorf("cannot create token maker : %w", err)
	}
	server = &Server{
		config:     config,
		store:      store,
		tokenMaker: tokenMaker,
	}
	return server, nil
}
