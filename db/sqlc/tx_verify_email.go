package db

import (
	"context"
	"github.com/jackc/pgx/v5/pgtype"
)

type VerifyEmailTXParams struct {
	EmailID int64
	Secret  string
}
type VerifyEmailTXResult struct {
	User        User
	VerifyEmail VerifyEmail
}

func (store *SQLStore) VerifyEmailTX(ctx context.Context, arg VerifyEmailTXParams) (VerifyEmailTXResult, error) {
	var result VerifyEmailTXResult
	err := store.execTx(ctx, func(q *Queries) error {
		//update verify email
		var err error
		result.VerifyEmail, err = q.UpdateVerifyEmail(ctx, UpdateVerifyEmailParams{
			ID:     arg.EmailID,
			Secret: arg.Secret,
		})
		if err != nil {
			return err
		}
		result.User, err = q.UpdateUser(ctx, UpdateUserParams{
			Username: result.VerifyEmail.Username,
			IsEmailVerified: pgtype.Bool{
				Bool:  true,
				Valid: true,
			},
		})
		if err != nil {
			return err
		}
		return nil
	})
	return result, err
}
