package gapi

import (
	"context"
	"database/sql"
	"fmt"
	mockdb "github.com/GGjahoon/MySimpleBank/db/mock"
	db "github.com/GGjahoon/MySimpleBank/db/sqlc"
	"github.com/GGjahoon/MySimpleBank/pb"
	"github.com/GGjahoon/MySimpleBank/util"
	"github.com/GGjahoon/MySimpleBank/worker"
	mockwk "github.com/GGjahoon/MySimpleBank/worker/mock"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"reflect"
	"testing"
)

// Create a new Matcher because hashed password is not same in twice generate
type eqCreateUserParamsMatcher struct {
	arg      db.CreateUserTxParams
	password string
	user     db.User
}

func (expected eqCreateUserParamsMatcher) Matches(x interface{}) bool {
	//将x断言为db.CreateUserTxParams
	actualArg, ok := x.(db.CreateUserTxParams)
	if !ok {
		return false
	}
	err := util.CheckPassword(expected.password, actualArg.CreateUserParams.HashedPassword)
	if err != nil {
		return false
	}
	expected.arg.HashedPassword = actualArg.CreateUserParams.HashedPassword

	if !reflect.DeepEqual(expected.arg.CreateUserParams, actualArg.CreateUserParams) {
		return false
	}
	// call the AfterCreate func here
	err = actualArg.AfterCreate(expected.user)
	return err == nil
}

func (expected eqCreateUserParamsMatcher) String() string {
	return fmt.Sprintf("matches arg %v and password %v", expected.arg, expected.password)
}

func EqCreateUserTXParams(arg db.CreateUserTxParams, password string, user db.User) gomock.Matcher {
	return eqCreateUserParamsMatcher{arg, password, user}
}

func TestCreateUserAPI(t *testing.T) {
	user, password := randomUser(t)
	//表驱动测试
	testCases := []struct {
		name          string
		req           *pb.CreateUserRequest
		buildStubs    func(store *mockdb.MockStore, taskDistributor *mockwk.MockTaskDistributor)
		checkResponse func(t *testing.T, rsp *pb.CreateUserResponse, err error)
	}{
		{
			name: "ok",
			req: &pb.CreateUserRequest{
				Username: user.Username,
				FullName: user.FullName,
				Email:    user.Email,
				Password: password,
			},
			buildStubs: func(store *mockdb.MockStore, taskDistributor *mockwk.MockTaskDistributor) {
				arg := db.CreateUserTxParams{
					CreateUserParams: db.CreateUserParams{
						Username: user.Username,
						FullName: user.FullName,
						Email:    user.Email,
					},
				}
				taskPayload := &worker.PayloadSendVerifyEmail{Username: user.Username}
				store.EXPECT().
					CreateUserTX(gomock.Any(), EqCreateUserTXParams(arg, password, user)).
					Times(1).Return(db.CreateUserTxResult{User: user}, nil)
				taskDistributor.EXPECT().
					DistributeTaskSendVerifyEmail(gomock.Any(), taskPayload, gomock.Any()).Times(1).Return(nil)
			},
			checkResponse: func(t *testing.T, rsp *pb.CreateUserResponse, err error) {
				require.NoError(t, err)
				require.NotNil(t, rsp)
				createdUser := rsp.GetUser()
				require.Equal(t, createdUser.Username, user.Username)
				require.Equal(t, createdUser.FullName, user.FullName)
				require.Equal(t, createdUser.Email, user.Email)
			},
		},
		{
			name: "InternalError",
			req: &pb.CreateUserRequest{
				Username: user.Username,
				FullName: user.FullName,
				Email:    user.Email,
				Password: password,
			},
			buildStubs: func(store *mockdb.MockStore, taskDistributor *mockwk.MockTaskDistributor) {
				//taskPayload := &worker.PayloadSendVerifyEmail{Username: user.Username}
				store.EXPECT().
					CreateUserTX(gomock.Any(), gomock.Any()).
					Times(1).Return(db.CreateUserTxResult{}, sql.ErrConnDone)
				taskDistributor.EXPECT().
					DistributeTaskSendVerifyEmail(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(t *testing.T, rsp *pb.CreateUserResponse, err error) {
				require.Error(t, err)
				s, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.Internal, s.Code())
			},
		},
		{
			name: "DuplicateUsername",
			req: &pb.CreateUserRequest{
				Username: user.Username,
				FullName: user.FullName,
				Email:    user.Email,
				Password: password,
			},
			buildStubs: func(store *mockdb.MockStore, taskDistributor *mockwk.MockTaskDistributor) {
				//taskPayload := &worker.PayloadSendVerifyEmail{Username: user.Username}
				store.EXPECT().
					CreateUserTX(gomock.Any(), gomock.Any()).
					Times(1).Return(db.CreateUserTxResult{}, db.ErrUniqueViolation)
				taskDistributor.EXPECT().
					DistributeTaskSendVerifyEmail(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(t *testing.T, rsp *pb.CreateUserResponse, err error) {
				require.Error(t, err)
				s, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.AlreadyExists, s.Code())
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dbCtrl := gomock.NewController(t)
			defer dbCtrl.Finish()
			wkCtrl := gomock.NewController(t)
			defer wkCtrl.Finish()
			store := mockdb.NewMockStore(dbCtrl)
			taskDistributor := mockwk.NewMockTaskDistributor(wkCtrl)
			tc.buildStubs(store, taskDistributor)

			server := NewTestServer(t, store, taskDistributor)
			rsp, err := server.CreateUser(context.Background(), tc.req)
			//此处的t与全局t(即t.Run的t)不同，其已被func的input t覆盖，所以其是子测试的t对象，由Run()创建
			//因此，未来添加更多测试案例时，每个案例的checkResponse调用将是独立的，不会互相干扰
			tc.checkResponse(t, rsp, err)
		})
	}
}
func randomUser(t *testing.T) (user db.User, password string) {
	password = util.RandomString(6)
	hashedPassword, err := util.HashPassword(password)
	require.NoError(t, err)

	user = db.User{
		Username:       util.RandomOwner(),
		HashedPassword: hashedPassword,
		FullName:       util.RandomOwner(),
		Email:          util.RandomEmail(),
		Role:           util.DepositorRole,
	}
	return
}
