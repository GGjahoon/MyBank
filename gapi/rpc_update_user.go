package gapi

import (
	"context"
	"database/sql"
	db "github.com/GGjahoon/MySimpleBank/db/sqlc"
	"github.com/GGjahoon/MySimpleBank/pb"
	"github.com/GGjahoon/MySimpleBank/util"
	"github.com/GGjahoon/MySimpleBank/val"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"time"
)

func (server *Server) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.UpdateUserResponse, error) {
	payload, err := server.authorizeUser(ctx)
	if err != nil {
		return nil, unAuthenticatedError(err)
	}
	violations := validateUpdateRequest(req)
	if violations != nil {
		return nil, invalidArgumentError(violations)
	}
	//check the username who send request is same as username of payload or not
	//avoid a user update other user's info
	if req.Username != payload.Username {
		return nil, status.Errorf(codes.PermissionDenied, "cannot update other user's info")
	}

	// create the arg to update user
	arg := db.UpdateUserParams{
		FullName: sql.NullString{
			String: req.GetFullName(),
			Valid:  req.FullName != nil,
		},
		Email: sql.NullString{
			String: req.GetEmail(),
			Valid:  req.Email != nil,
		},
		Username: req.GetUsername(),
	}
	if req.Password != nil {
		hashedPassword, err := util.HashPassword(req.GetPassword())
		if err != nil {
			return nil, status.Errorf(codes.Internal, "hashed password failed: %s", err)
		}
		arg.HashedPassword = sql.NullString{
			String: hashedPassword,
			Valid:  true,
		}
		arg.PasswordChangedAt = sql.NullTime{
			Time:  time.Now(),
			Valid: true,
		}
	}
	user, err := server.store.UpdateUser(ctx, arg)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Errorf(codes.NotFound, "user not found in db")
		}
		return nil, status.Errorf(codes.Internal, "failed to update user")
	}
	rsp := &pb.UpdateUserResponse{
		User: convertUser(user),
	}

	return rsp, nil
}
func validateUpdateRequest(req *pb.UpdateUserRequest) (violations []*errdetails.BadRequest_FieldViolation) {
	if err := val.ValidateUsername(req.GetUsername()); err != nil {
		violations = append(violations, fieldViolations("username", err))
	}
	if req.Password != nil {
		if err := val.ValidatePassword(req.GetPassword()); err != nil {
			violations = append(violations, fieldViolations("password", err))
		}
	}
	if req.FullName != nil {
		if err := val.ValidateFullName(req.GetFullName()); err != nil {
			violations = append(violations, fieldViolations("full_name", err))
		}
	}
	if req.Email != nil {
		if err := val.ValidateEmail(req.GetEmail()); err != nil {
			violations = append(violations, fieldViolations("email", err))
		}
	}
	return violations
}
