-- name: CreateAppointment :one
INSERT INTO appointments (
training_id, internal_user_id, additional_child_number
) VALUES (
$1, $2, $3
) ON CONFLICT (training_id, internal_user_id)
DO UPDATE SET additional_child_number = EXCLUDED.additional_child_number
RETURNING *;

-- name: GetAppointment :one
SELECT * FROM appointments
WHERE appointment_id = $1 LIMIT 1;

-- name: ListUserTrainings :many
SELECT appointment_id, appointments.training_id, additional_child_number, users.telegram_user_id, type, date_and_time, price, trainer  FROM appointments
JOIN trainings ON appointments.training_id=trainings.training_id
JOIN users ON appointments.internal_user_id=users.internal_user_id
WHERE telegram_user_id = $1 AND date_and_time > now()
ORDER BY date_and_time;

-- name: ListTrainingUsers :many
SELECT appointment_id, training_id, appointments.internal_user_id, telegram_user_id, full_name, additional_child_number, appointments.created_at FROM appointments
JOIN users ON appointments.internal_user_id=users.internal_user_id
WHERE training_id = $1
ORDER BY appointments.created_at;

-- name: DeleteAppointment :exec
DELETE FROM appointments WHERE appointment_id = $1;

-- name: ListUsersForAlert :many
SELECT appointments.internal_user_id, telegram_user_id, date_and_time FROM appointments
JOIN trainings ON appointments.training_id=trainings.training_id
JOIN users ON appointments.internal_user_id=users.internal_user_id
WHERE users.telegram_user_id IS NOT NULL AND date_part('day', date_and_time)  = date_part('day', now() + INTERVAL '1' DAY) AND date_and_time > now();

-- name: GetAppointmentCount :one
SELECT COUNT(*) FROM appointments
WHERE training_id = $1;

