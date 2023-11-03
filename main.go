package main

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/GGjahoon/MySimpleBank/api"
	db "github.com/GGjahoon/MySimpleBank/db/sqlc"
	"github.com/GGjahoon/MySimpleBank/gapi"
	"github.com/GGjahoon/MySimpleBank/pb"
	"github.com/GGjahoon/MySimpleBank/util"
	"github.com/GGjahoon/MySimpleBank/worker"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/hibiken/asynq"
	_ "github.com/lib/pq"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/encoding/protojson"
	"net"
	"net/http"
	"os"
)

func main() {
	config, err := util.LoadConfig(".")
	if err != nil {
		log.Fatal().Err(err).Msg("cannot load config ")
	}
	if config.Environment == "development" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}
	conn, err := sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot connect to db")
	}
	fmt.Println(config.MigrationUrl)
	runDBMigration(config.MigrationUrl, config.DBSource)
	store := db.NewStore(conn)
	redisOptions := asynq.RedisClientOpt{
		Addr: config.RedisAddress,
	}
	//create a new taskDistributor
	taskDistributor := worker.NewRedisTaskDistributor(redisOptions)
	//run a new processor
	go runProcessor(redisOptions, store)
	//runGinServer(config, store)
	go runGatewayServer(config, store, taskDistributor)
	runGrpcServer(config, store, taskDistributor)
}
func runProcessor(redisOpt asynq.RedisClientOpt, store db.Store) {
	processor := worker.NewRedisTaskProcessor(redisOpt, store)
	log.Info().Msg("start task processor")
	err := processor.Start()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to start task processor")
	}
}

func runDBMigration(migrateURL string, DBSource string) {
	m, err := migrate.New(migrateURL, DBSource)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot create a new migrate instance")
	}
	if err = m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatal().Err(err).Msg("cannot migrate up")
	}
	log.Info().Msg("migrate successfully")
}

func runGrpcServer(config util.Config, store db.Store, taskDistributor worker.TaskDistributor) {
	server, err := gapi.NewServer(config, store, taskDistributor)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot create a new server")
	}
	grpcLogger := grpc.UnaryInterceptor(gapi.GrpcLogger)
	grpcServer := grpc.NewServer(grpcLogger)
	pb.RegisterSimpleBankServer(grpcServer, server)
	reflection.Register(grpcServer)
	listener, err := net.Listen("tcp", config.GRPCServerAddress)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot create listener")
	}
	log.Info().Msgf("start gRPC server at %s", listener.Addr().String())
	err = grpcServer.Serve(listener)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot create grpc Server")
	}
}
func runGatewayServer(config util.Config, store db.Store, taskDistributor worker.TaskDistributor) {
	server, err := gapi.NewServer(config, store, taskDistributor)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot create a new server")
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
	//函数 RegisterSimpleBankHandlerServer 在HTTP和gRPC之间执行进程内转换
	//即不通过gRPC interceptor 直接调用gRPC服务器的Handler
	err = pb.RegisterSimpleBankHandlerServer(ctx, grpcMux, server)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot register handler server  ")
	}
	//create a new http serve mux , it receive http request from client
	mux := http.NewServeMux()
	//in order to convert the request from client to grpc format,rerouter them to grpcMux
	//"/"作为第一个参数以覆盖所有路由
	mux.Handle("/", grpcMux)
	listener, err := net.Listen("tcp", config.HTTPServerAddress)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot create listener")
	}
	log.Info().Msgf("start REST server at %s", config.HTTPServerAddress)
	logger := gapi.HttpLogger(mux)
	err = http.Serve(listener, logger)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot start server")
	}
}

func runGinServer(config util.Config, store db.Store) {
	server, err := api.NewServer(config, store)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot create server")
	}
	err = server.Start(config.HTTPServerAddress)
	if err != nil {
		log.Fatal().Err(err).Msg("server cannot start")
	}
}
