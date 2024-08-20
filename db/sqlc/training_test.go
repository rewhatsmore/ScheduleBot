package db

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"schedule.sqlc.dev/app/db/random"
)

func createRandomTraining(t *testing.T) Training {
	arg := CreateTrainingParams{
		DateAndTime: random.RandTrainingTime(),
		GroupType:   GroupTypeEnumAdult,
	}

	training, err := testQueries.CreateTraining(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, training)
	require.WithinDuration(t, arg.DateAndTime, training.DateAndTime, time.Second)
	require.NotZero(t, training.Price)
	require.NotZero(t, training.Trainer)
	require.NotZero(t, training.TrainingID)
	require.NotZero(t, training.Type)
	return training
}

func TestCreateTraining(t *testing.T) {
	createRandomTraining(t)
}

func TestGetTraining(t *testing.T) {
	training1 := createRandomTraining(t)

	training2, err := testQueries.GetTraining(context.Background(), training1.TrainingID)
	require.NoError(t, err)
	require.NotEmpty(t, training2)
	require.Equal(t, training1.TrainingID, training2.TrainingID)
	require.Equal(t, training1.Place, training2.Place)
	require.Equal(t, training1.Price, training2.Price)
	require.Equal(t, training1.Type, training2.Type)
	require.Equal(t, training1.Trainer, training2.Trainer)
	require.WithinDuration(t, training1.DateAndTime, training2.DateAndTime, time.Second)
}

// func TestUpdateTraining(t *testing.T) {
// 	training1 := createRandomTraining(t)

// 	arg := UpdateTrainingParams{
// 		TrainingID: training1.TrainingID,
// 		Trainer:    "Саша Колесова",
// 	}

// 	training2, err := testQueries.UpdateTraining(context.Background(), arg)
// 	require.NoError(t, err)
// 	require.NotEmpty(t, training2)
// 	require.Equal(t, training1.TrainingID, training2.TrainingID)
// 	require.Equal(t, training1.Place, training2.Place)
// 	require.Equal(t, training1.Price, training2.Price)
// 	require.Equal(t, training1.Type, training2.Type)
// 	require.NotEqual(t, training1.Trainer, training2.Trainer)
// 	require.Equal(t, arg.Trainer, training2.Trainer)
// 	require.WithinDuration(t, training1.DateAndTime, training2.DateAndTime, time.Second)
// }

func TestDeleteTraining(t *testing.T) {
	training1 := createRandomTraining(t)
	ctx := context.Background()

	err := testQueries.DeleteTraining(ctx, training1.TrainingID)
	require.NoError(t, err)

	training2, err := testQueries.GetTraining(ctx, training1.TrainingID)
	require.Error(t, err)
	require.EqualError(t, err, sql.ErrNoRows.Error())
	require.Empty(t, training2)
}

func TestListTrainings(t *testing.T) {
	trainings, err := testQueries.ListTrainings(context.Background())
	require.NoError(t, err)
	require.NotEmpty(t, trainings)

	for _, training := range trainings {
		require.NotEmpty(t, training)
		require.GreaterOrEqual(t, training.DateAndTime, time.Now())
	}

}
