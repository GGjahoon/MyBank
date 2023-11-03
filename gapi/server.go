package gapi

import (
	"fmt"
	db "github.com/GGjahoon/MySimpleBank/db/sqlc"
	"github.com/GGjahoon/MySimpleBank/pb"
	"github.com/GGjahoon/MySimpleBank/token"
	"github.com/GGjahoon/MySimpleBank/util"
	"github.com/GGjahoon/MySimpleBank/worker"
)

// Server serves gRPC request for bank service
type Server struct {
	pb.UnimplementedSimpleBankServer
	config          util.Config
	store           db.Store
	tokenMaker      token.Maker
	taskDistributor worker.TaskDistributor
}

// NewServer create a new gRPC server
func NewServer(config util.Config, store db.Store, taskDistributor worker.TaskDistributor) (server *Server, err error) {
	tokenMaker, err := token.NewPasetoMaker(config.TokenSymmetricKey)
	if err != nil {
		return nil, fmt.Errorf("cannot create token maker : %w", err)
	}
	server = &Server{
		config:          config,
		store:           store,
		tokenMaker:      tokenMaker,
		taskDistributor: taskDistributor,
	}
	return server, nil
}
