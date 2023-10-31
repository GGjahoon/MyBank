package db

import (
	"context"
	"database/sql"
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

func TestUpdateUserOnlyFullName(t *testing.T) {
	oldUser := CreateRandomUser(t)
	newFullName := util.RandomOwner()
	newUser, err := testQueries.UpdateUser(context.Background(), UpdateUserParams{
		FullName: sql.NullString{
			String: newFullName,
			Valid:  true,
		},
		Username: oldUser.Username,
	})
	require.NoError(t, err)
	require.NotEqual(t, newUser.FullName, oldUser.FullName)
	require.Equal(t, newUser.FullName, newFullName)
	require.Equal(t, newUser.Email, oldUser.Email)
	require.Equal(t, newUser.Username, oldUser.Username)
	require.Equal(t, newUser.HashedPassword, oldUser.HashedPassword)
}
func TestUpdateUserOnlyEmail(t *testing.T) {
	oldUser := CreateRandomUser(t)
	newEmail := util.RandomEmail()
	newUser, err := testQueries.UpdateUser(context.Background(), UpdateUserParams{
		Email: sql.NullString{
			String: newEmail,
			Valid:  true,
		},
		Username: oldUser.Username,
	})
	require.NoError(t, err)
	require.NotEqual(t, newUser.Email, oldUser.Email)
	require.Equal(t, newUser.Email, newEmail)
	require.Equal(t, newUser.FullName, oldUser.FullName)
	require.Equal(t, newUser.Username, oldUser.Username)
	require.Equal(t, newUser.HashedPassword, oldUser.HashedPassword)
}

func TestUpdateUserOnlyPassword(t *testing.T) {
	oldUser := CreateRandomUser(t)
	newPassword := util.RandomString(6)
	newHashedPassword, err := util.HashPassword(newPassword)
	require.NoError(t, err)

	newUser, err := testQueries.UpdateUser(context.Background(), UpdateUserParams{
		HashedPassword: sql.NullString{
			String: newHashedPassword,
			Valid:  true,
		},
		Username: oldUser.Username,
	})
	require.NoError(t, err)
	require.Equal(t, newUser.HashedPassword, newHashedPassword)
	require.NotEqual(t, newUser.HashedPassword, oldUser.HashedPassword)
	//the new password should not be match to the old hashed password
	err = util.CheckPassword(newPassword, oldUser.HashedPassword)
	require.Error(t, err)
	err = util.CheckPassword(newPassword, newUser.HashedPassword)
	require.NoError(t, err)

	require.Equal(t, newUser.FullName, oldUser.FullName)
	require.Equal(t, newUser.Username, oldUser.Username)
	require.Equal(t, newUser.Email, oldUser.Email)

}
func TestUpdateUserAllFields(t *testing.T) {
	oldUser := CreateRandomUser(t)

	newFullName := util.RandomOwner()
	newEmail := util.RandomEmail()

	newPassword := util.RandomString(6)
	newHashedPassword, err := util.HashPassword(newPassword)
	require.NoError(t, err)

	newUser, err := testQueries.UpdateUser(context.Background(), UpdateUserParams{
		FullName: sql.NullString{
			String: newFullName,
			Valid:  true,
		},
		Email: sql.NullString{
			String: newEmail,
			Valid:  true,
		},
		HashedPassword: sql.NullString{
			String: newHashedPassword,
			Valid:  true,
		},
		Username: oldUser.Username,
	})
	require.NoError(t, err)

	require.NotEqual(t, oldUser.FullName, newUser.FullName)
	require.Equal(t, newUser.FullName, newFullName)

	require.NotEqual(t, oldUser.Email, newUser.Email)
	require.Equal(t, newUser.Email, newEmail)

	require.NotEqual(t, oldUser.HashedPassword, newUser.HashedPassword)
	require.Equal(t, newUser.HashedPassword, newHashedPassword)
}
