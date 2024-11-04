package db

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func createRandomAppointment(t *testing.T, user User, training Training) Appointment {
	arg := CreateAppointmentParams{
		TrainingID:     training.TrainingID,
		InternalUserID: int64(user.InternalUserID),
	}

	appointment, err := testQueries.CreateAppointment(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, appointment)
	require.NotZero(t, appointment.AppointmentID)
	require.NotZero(t, appointment.CreatedAt)
	require.Equal(t, arg.TrainingID, appointment.TrainingID)
	require.Equal(t, appointment.InternalUserID, arg.InternalUserID)
	return appointment
}

func TestCreateAppointment(t *testing.T) {
	user := createRandomUser(t)
	training := createRandomTraining(t)
	createRandomAppointment(t, user, training)
}

func TestGetAppointment(t *testing.T) {
	user := createRandomUser(t)
	training := createRandomTraining(t)
	appointment1 := createRandomAppointment(t, user, training)

	appointment2, err := testQueries.GetAppointment(context.Background(), appointment1.AppointmentID)
	require.NoError(t, err)
	require.NotEmpty(t, appointment2)
	require.Equal(t, appointment1.AppointmentID, appointment2.AppointmentID)
	require.Equal(t, appointment1.TrainingID, appointment2.TrainingID)
	require.Equal(t, appointment1.InternalUserID, appointment2.InternalUserID)
	require.WithinDuration(t, appointment1.CreatedAt, appointment2.CreatedAt, time.Second)
}

func TestDeleteAppointment(t *testing.T) {
	user := createRandomUser(t)
	training := createRandomTraining(t)
	appointment1 := createRandomAppointment(t, user, training)

	err := testQueries.DeleteAppointment(context.Background(), appointment1.AppointmentID)
	require.NoError(t, err)

	appointment2, err := testQueries.GetAppointment(context.Background(), appointment1.AppointmentID)
	require.Error(t, err)
	require.EqualError(t, err, sql.ErrNoRows.Error())
	require.Empty(t, appointment2)
}

func TestListTrainingUsers(t *testing.T) {
	n := 5
	users := make([]User, n)
	appointments := make([]Appointment, n)
	training := createRandomTraining(t)

	for i := 0; i < n; i++ {
		user := createRandomUser(t)
		users[i] = user
		appointment := createRandomAppointment(t, user, training)
		appointments[i] = appointment
	}

	trainingUsers, err := testQueries.ListTrainingUsers(context.Background(), training.TrainingID)
	require.NoError(t, err)
	require.NotEmpty(t, trainingUsers)
	require.Equal(t, len(trainingUsers), n)

	for i, trainingUser := range trainingUsers {
		require.NotEmpty(t, trainingUser)
		require.Equal(t, trainingUser.TrainingID, training.TrainingID)
		require.Equal(t, trainingUser.InternalUserID, appointments[i].InternalUserID)
		require.Equal(t, trainingUser.AppointmentID, appointments[i].AppointmentID)
		require.Equal(t, trainingUser.FullName, users[i].FullName)
		require.WithinDuration(t, trainingUser.CreatedAt, appointments[i].CreatedAt, time.Second)
	}
}

func TestListUserTrainings(t *testing.T) {
	n := 5
	trainings := make([]Training, n)
	appointments := make([]Appointment, n)
	user := createRandomUser(t)

	for i := 0; i < n; i++ {
		training := createRandomTraining(t)
		trainings[i] = training
		appointment := createRandomAppointment(t, user, training)
		appointments[i] = appointment
	}

	userTrainings, err := testQueries.ListUserTrainings(context.Background(), user.TelegramUserID)
	require.NoError(t, err)
	require.NotEmpty(t, userTrainings)
	require.Equal(t, len(userTrainings), n)

	for _, userTraining := range userTrainings {
		require.NotEmpty(t, userTraining)
		require.Equal(t, userTraining.TelegramUserID, user.TelegramUserID)
	}
}
