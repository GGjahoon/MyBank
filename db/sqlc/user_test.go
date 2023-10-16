package db

import (
	"context"
	"github.com/GGjahoon/MySimpleBank/util"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

// CreateRandomUser create a random user
func CreateRandomUser(t *testing.T) User {
	hashedPassword, err := util.HashPassword(util.RandomString(6))
	require.NoError(t, err)
	require.NotEmpty(t, hashedPassword)
	arg := CreateUserParams{
		Username:       util.RandomOwner(),
		HashedPassword: hashedPassword,
		FullName:       util.RandomOwner(),
		Email:          util.RandomEmail(),
	}
	user, err := testQueries.CreateUser(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, user)

	require.Equal(t, arg.Username, user.Username)
	require.Equal(t, arg.HashedPassword, user.HashedPassword)
	require.Equal(t, arg.FullName, user.FullName)
	require.Equal(t, arg.Email, user.Email)

	require.NotZero(t, user.CreatedAt)
	require.True(t, user.PasswordChangedAt.IsZero())
	return user
}

func TestCreateUser(t *testing.T) {
	CreateRandomUser(t)
}

func TestGetUser(t *testing.T) {
	user1 := CreateRandomUser(t)
	user2, err := testQueries.GetUser(context.Background(), user1.Username)
	require.NoError(t, err)
	require.NotEmpty(t, user2)

	require.Equal(t, user2.Username, user1.Username)
	require.Equal(t, user2.FullName, user1.FullName)
	require.Equal(t, user2.HashedPassword, user1.HashedPassword)
	require.Equal(t, user2.Email, user1.Email)

	require.WithinDuration(t, user2.CreatedAt, user1.CreatedAt, time.Second)
	require.WithinDuration(t, user2.PasswordChangedAt, user1.PasswordChangedAt, time.Second)
}
