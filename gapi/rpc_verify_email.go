package gapi

import (
	"context"
	db "github.com/GGjahoon/MySimpleBank/db/sqlc"
	"github.com/GGjahoon/MySimpleBank/pb"
	"github.com/GGjahoon/MySimpleBank/val"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (server *Server) VerifyEmail(ctx context.Context, req *pb.VerifyEmailRequest) (*pb.VerifyEmailResponse, error) {
	violations := ValidateVerifyEmailRequest(req)
	if violations != nil {
		return nil, invalidArgumentError(violations)
	}
	txResult, err := server.store.VerifyEmailTX(ctx, db.VerifyEmailTXParams{
		EmailID: req.GetEmailId(),
		Secret:  req.GetSecret(),
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to verify email")
	}
	rsp := &pb.VerifyEmailResponse{IsVerified: txResult.User.IsEmailVerified}
	return rsp, nil
}
func ValidateVerifyEmailRequest(req *pb.VerifyEmailRequest) (violations []*errdetails.BadRequest_FieldViolation) {
	var err error
	if err = val.ValidateEmailID(req.GetEmailId()); err != nil {
		violations = append(violations, fieldViolations("email_id", err))
	}
	if err = val.ValidateSecret(req.GetSecret()); err != nil {
		violations = append(violations, fieldViolations("secret", err))
	}
	return violations

}
