package gapi

import (
	"context"
	mockdb "github.com/GGjahoon/MySimpleBank/db/mock"
	db "github.com/GGjahoon/MySimpleBank/db/sqlc"
	"github.com/GGjahoon/MySimpleBank/pb"
	"github.com/GGjahoon/MySimpleBank/token"
	"github.com/GGjahoon/MySimpleBank/util"
	"github.com/golang/mock/gomock"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"testing"
	"time"
)

func TestUpdateUserAPI(t *testing.T) {
	user, _ := randomUser(t)
	newName := util.RandomOwner()
	newEmail := util.RandomEmail()
	invalidEmail := "invalid-email"
	errUser := "errUser"
	//表驱动测试
	testCases := []struct {
		name          string
		req           *pb.UpdateUserRequest
		buildStubs    func(store *mockdb.MockStore)
		buildContext  func(t *testing.T, tokenMaker token.Maker) context.Context
		checkResponse func(t *testing.T, rsp *pb.UpdateUserResponse, err error)
	}{
		{
			name: "ok",
			req: &pb.UpdateUserRequest{
				Username: user.Username,
				FullName: &newName,
				Email:    &newEmail,
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.UpdateUserParams{
					Username: user.Username,
					Email: pgtype.Text{
						String: newEmail,
						Valid:  true,
					},
					FullName: pgtype.Text{
						String: newName,
						Valid:  true,
					},
				}
				store.EXPECT().
					UpdateUser(gomock.Any(), gomock.Eq(arg)).
					Times(1).Return(db.User{
					Username:          user.Username,
					HashedPassword:    user.HashedPassword,
					FullName:          newName,
					Email:             newEmail,
					PasswordChangedAt: user.PasswordChangedAt,
					CreatedAt:         user.CreatedAt,
					IsEmailVerified:   user.IsEmailVerified,
				}, nil)
			},
			buildContext: func(t *testing.T, tokenMaker token.Maker) context.Context {
				return NewContextWithToken(t, tokenMaker, user.Username, user.Role, time.Minute)
			},
			checkResponse: func(t *testing.T, rsp *pb.UpdateUserResponse, err error) {
				require.NoError(t, err)
				require.NotNil(t, rsp)
				updateUser := rsp.GetUser()
				require.Equal(t, updateUser.Username, user.Username)
				require.Equal(t, updateUser.FullName, newName)
				require.Equal(t, updateUser.Email, newEmail)
			},
		},
		{
			name: "expiredToken",
			req: &pb.UpdateUserRequest{
				Username: user.Username,
				FullName: &newName,
				Email:    &newEmail,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					UpdateUser(gomock.Any(), gomock.Any()).
					Times(0)
			},
			buildContext: func(t *testing.T, tokenMaker token.Maker) context.Context {
				return NewContextWithToken(t, tokenMaker, user.Username, user.Role, -time.Minute)
			},
			checkResponse: func(t *testing.T, rsp *pb.UpdateUserResponse, err error) {
				require.Error(t, err)
				require.Nil(t, rsp)
				s, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, s.Code(), codes.Unauthenticated)
			},
		},
		{
			name: "unauthorized",
			req: &pb.UpdateUserRequest{
				Username: user.Username,
				FullName: &newName,
				Email:    &newEmail,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					UpdateUser(gomock.Any(), gomock.Any()).
					Times(0)
			},
			buildContext: func(t *testing.T, tokenMaker token.Maker) context.Context {
				return context.Background()
			},
			checkResponse: func(t *testing.T, rsp *pb.UpdateUserResponse, err error) {
				require.Error(t, err)
				require.Nil(t, rsp)
				s, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, s.Code(), codes.Unauthenticated)
			},
		},
		{
			name: "UserNotFound",
			req: &pb.UpdateUserRequest{
				Username: user.Username,
				FullName: &newName,
				Email:    &newEmail,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					UpdateUser(gomock.Any(), gomock.Any()).
					Times(1).Return(db.User{}, db.ErrRecordNotFound)
			},
			buildContext: func(t *testing.T, tokenMaker token.Maker) context.Context {
				return NewContextWithToken(t, tokenMaker, user.Username, user.Role, time.Minute)
			},
			checkResponse: func(t *testing.T, rsp *pb.UpdateUserResponse, err error) {
				require.Error(t, err)
				require.Nil(t, rsp)
				s, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, s.Code(), codes.NotFound)
			},
		},
		{
			name: "InvalidEmail",
			req: &pb.UpdateUserRequest{
				Username: user.Username,
				FullName: &newName,
				Email:    &invalidEmail,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					UpdateUser(gomock.Any(), gomock.Any()).
					Times(0)
			},
			buildContext: func(t *testing.T, tokenMaker token.Maker) context.Context {
				return NewContextWithToken(t, tokenMaker, user.Username, user.Role, time.Minute)
			},
			checkResponse: func(t *testing.T, rsp *pb.UpdateUserResponse, err error) {
				require.Error(t, err)
				require.Nil(t, rsp)
				s, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, s.Code(), codes.InvalidArgument)
			},
		},
		{
			name: "ErrUser",
			req: &pb.UpdateUserRequest{
				Username: user.Username,
				FullName: &newName,
				Email:    &newEmail,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					UpdateUser(gomock.Any(), gomock.Any()).
					Times(0)
			},
			buildContext: func(t *testing.T, tokenMaker token.Maker) context.Context {
				return NewContextWithToken(t, tokenMaker, errUser, user.Role, time.Minute)
			},
			checkResponse: func(t *testing.T, rsp *pb.UpdateUserResponse, err error) {
				require.Error(t, err)
				require.Nil(t, rsp)
				s, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, s.Code(), codes.PermissionDenied)
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dbCtrl := gomock.NewController(t)
			defer dbCtrl.Finish()

			store := mockdb.NewMockStore(dbCtrl)

			tc.buildStubs(store)

			server := NewTestServer(t, store, nil)
			ctx := tc.buildContext(t, server.tokenMaker)
			rsp, err := server.UpdateUser(ctx, tc.req)
			//此处的t与全局t(即t.Run的t)不同，其已被func的input t覆盖，所以其是子测试的t对象，由Run()创建
			//因此，未来添加更多测试案例时，每个案例的checkResponse调用将是独立的，不会互相干扰
			tc.checkResponse(t, rsp, err)
		})
	}
}
