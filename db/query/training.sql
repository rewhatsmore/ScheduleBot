-- name: CreateTraining :one
INSERT INTO trainings (
date_and_time, group_type, column_number
) VALUES (
$1, $2, $3
)
RETURNING *;

-- name: GetTraining :one
SELECT * FROM trainings
WHERE training_id = $1 LIMIT 1;

-- name: ListTrainings :many
SELECT * FROM trainings
WHERE date_and_time > now()
ORDER BY date_and_time;

-- name: ListChildrenTrainings :many
SELECT * FROM trainings
WHERE date_and_time > now() AND group_type = 'child'
ORDER BY date_and_time;

-- name: ListAdultTrainings :many
SELECT * FROM trainings
WHERE date_and_time > now() AND group_type = 'adult'
ORDER BY date_and_time;

-- name: ListLastWeekTrainings :many
SELECT * FROM trainings
WHERE date_and_time > now() - INTERVAL '7' DAY
ORDER BY date_and_time;

-- -- name: UpdateTraining :one
-- UPDATE trainings SET trainer = $2
-- WHERE training_id = $1
-- RETURNING *;

-- name: DeleteTraining :exec
DELETE FROM trainings WHERE training_id = $1;

-- name: ListTrainingsForSend :many
SELECT trainings.training_id, date_and_time, column_number, COALESCE (U.appointment_id, 0) AS appointment_id, COALESCE (additional_child_number, -1) AS additional_child_number
FROM trainings
LEFT JOIN (SELECT * FROM appointments WHERE user_id=$1) AS U
ON trainings.training_id = U.training_id
WHERE date_and_time > now() AND group_type = $2
ORDER BY date_and_time;

