package gapi

import (
	"context"
	db "github.com/GGjahoon/MySimpleBank/db/sqlc"
	"github.com/GGjahoon/MySimpleBank/pb"
	"github.com/GGjahoon/MySimpleBank/util"
	"github.com/GGjahoon/MySimpleBank/val"
	"github.com/GGjahoon/MySimpleBank/worker"
	"github.com/hibiken/asynq"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"time"
)

func (server *Server) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (rsp *pb.CreateUserResponse, err error) {
	// call Valid first
	violations := validateCreateUserRequest(req)
	//if violations is not nil,means there are at least one invalid parameter,must return badRequest to client
	if violations != nil {
		return nil, invalidArgumentError(violations)
	}
	hashedPassword, err := util.HashPassword(req.GetPassword())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "hashed password failed: %s", err)
	}

	arg := db.CreateUserTxParams{
		CreateUserParams: db.CreateUserParams{
			Username:       req.GetUsername(),
			HashedPassword: hashedPassword,
			FullName:       req.GetFullName(),
			Email:          req.GetEmail(),
		},
		AfterCreate: func(user db.User) error {
			payload := &worker.PayloadSendVerifyEmail{Username: user.Username}
			opts := []asynq.Option{
				//最大重试次数：10
				asynq.MaxRetry(10),
				//放入critical queue中
				asynq.Queue(worker.QueueCritical),
				//延迟worker执行时间为10s后
				asynq.ProcessIn(time.Second * 1),
			}
			return server.taskDistributor.DistributeTaskSendVerifyEmail(ctx, payload, opts...)
		},
	}
	txResult, err := server.store.CreateUserTX(ctx, arg)
	if err != nil {
		if db.ErrorCode(err) == db.UniqueViolation {
			return nil, status.Errorf(codes.AlreadyExists, "%s", err)
		}
		return nil, status.Errorf(codes.Internal, "failed to create a new user: %s", err)
	}
	rsp = &pb.CreateUserResponse{
		User: convertUser(txResult.User),
	}
	return rsp, nil
}
func validateCreateUserRequest(req *pb.CreateUserRequest) (violations []*errdetails.BadRequest_FieldViolation) {
	if err := val.ValidateUsername(req.GetUsername()); err != nil {
		violations = append(violations, fieldViolations("username", err))
	}
	if err := val.ValidatePassword(req.GetPassword()); err != nil {
		violations = append(violations, fieldViolations("password", err))
	}
	if err := val.ValidateFullName(req.GetFullName()); err != nil {
		violations = append(violations, fieldViolations("full_name", err))
	}
	if err := val.ValidateEmail(req.GetEmail()); err != nil {
		violations = append(violations, fieldViolations("email", err))
	}
	return violations
}
