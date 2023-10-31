package gapi

import (
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func fieldViolations(field string, err error) *errdetails.BadRequest_FieldViolation {
	return &errdetails.BadRequest_FieldViolation{
		Field:       field,
		Description: err.Error(),
	}
}
func invalidArgumentError(violations []*errdetails.BadRequest_FieldViolation) error {
	//create a bad request object
	badRequest := &errdetails.BadRequest{FieldViolations: violations}
	//create a new statusInvalid object
	statusInvalid := status.New(codes.InvalidArgument, "invalid parameters")

	//add more details of those invalid parameters into status object
	statusDetails, err := statusInvalid.WithDetails(badRequest)
	if err != nil {
		return statusInvalid.Err()
	}
	return statusDetails.Err()
}
func unAuthenticatedError(err error) error {
	return status.Errorf(codes.Unauthenticated, "unauthorized : %s", err)
}
