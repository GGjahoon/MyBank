package main

import (
	"context"
	"database/sql"
	"github.com/GGjahoon/MySimpleBank/api"
	db "github.com/GGjahoon/MySimpleBank/db/sqlc"
	"github.com/GGjahoon/MySimpleBank/gapi"
	"github.com/GGjahoon/MySimpleBank/pb"
	"github.com/GGjahoon/MySimpleBank/util"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/encoding/protojson"
	"log"
	"net"
	"net/http"
)

func main() {
	config, err := util.LoadConfig(".")
	if err != nil {
		log.Fatal("cannot load config :", err)
	}
	conn, err := sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}
	store := db.NewStore(conn)

	runGinServer(config, store)
	//go runGatewayServer(config, store)
	//runGrpcServer(config, store)
}
func runGrpcServer(config util.Config, store db.Store) {
	server, err := gapi.NewServer(config, store)
	if err != nil {
		log.Fatal("cannot create a new server", err)
	}
	grpcServer := grpc.NewServer()
	pb.RegisterSimpleBankServer(grpcServer, server)
	reflection.Register(grpcServer)
	listener, err := net.Listen("tcp", config.GRPCServerAddress)
	if err != nil {
		log.Fatal("cannot create listener", err)
	}
	log.Printf("start gRPC server at %s", listener.Addr().String())
	err = grpcServer.Serve(listener)
	if err != nil {
		log.Fatal("cannot create grpc Server", err)
	}
}
func runGatewayServer(config util.Config, store db.Store) {
	server, err := gapi.NewServer(config, store)
	if err != nil {
		log.Fatal("cannot create a new server:", err)
	}
	jsonOption := runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
		MarshalOptions: protojson.MarshalOptions{
			UseProtoNames: true,
		},
		UnmarshalOptions: protojson.UnmarshalOptions{
			DiscardUnknown: true,
		},
	})
	grpcMux := runtime.NewServeMux(jsonOption)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	err = pb.RegisterSimpleBankHandlerServer(ctx, grpcMux, server)
	if err != nil {
		log.Fatal("cannot register handler server  ")
	}
	//create a new http serve mux , it receive http request from client
	mux := http.NewServeMux()
	//in order to convert the request from client to grpc format,rerouter them to grpcMux
	//"/"作为第一个参数以覆盖所有路由
	mux.Handle("/", grpcMux)
	listener, err := net.Listen("tcp", config.HTTPServerAddress)
	if err != nil {
		log.Fatal("cannot create listener", err)
	}
	log.Printf("start REST server at %s", config.HTTPServerAddress)
	err = http.Serve(listener, mux)
	if err != nil {
		log.Fatal("cannot start server", err)
	}
}

func runGinServer(config util.Config, store db.Store) {
	server, err := api.NewServer(config, store)
	if err != nil {
		log.Fatal("cannot create server", err)
	}
	err = server.Start(config.HTTPServerAddress)
	if err != nil {
		log.Fatal("server cannot start", err)
	}
}
