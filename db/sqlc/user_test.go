package db

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"schedule.sqlc.dev/app/db/random"
)

func createRandomUser(t *testing.T) User {
	arg := CreateUserParams{
		TelegramUserID: random.RandInt(),
		FullName:       random.RandString(),
	}

	user, err := testQueries.CreateUser(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, user)
	require.Equal(t, arg.TelegramUserID, user.TelegramUserID)
	require.Equal(t, arg.FullName, user.FullName)
	require.NotZero(t, user.CreatedAt)
	return user
}

func TestCreateUser(t *testing.T) {
	createRandomUser(t)
}

func TestGetUser(t *testing.T) {
	user1 := createRandomUser(t)

	user2, err := testQueries.GetUser(context.Background(), user1.TelegramUserID)
	require.NoError(t, err)
	require.NotEmpty(t, user2)
	require.Equal(t, user1.TelegramUserID, user2.TelegramUserID)
	require.Equal(t, user1.FullName, user2.FullName)
	require.WithinDuration(t, user1.CreatedAt, user2.CreatedAt, time.Second)
}

// func TestUpdateUser(t *testing.T) {
// 	user := createRandomUser(t)

// 	arg := UpdateUserParams{
// 		UserID:   user.UserID,
// 		FullName: random.RandString(),
// 	}

// 	updatedUser, err := testQueries.UpdateUser(context.Background(), arg)
// 	require.NoError(t, err)
// 	require.NotEmpty(t, updatedUser)
// 	require.Equal(t, user.UserID, updatedUser.UserID)
// 	require.WithinDuration(t, user.CreatedAt, updatedUser.CreatedAt, time.Second)
// 	require.NotEqual(t, user.FullName, updatedUser.FullName)
// 	require.Equal(t, updatedUser.FullName, arg.FullName)

// }
