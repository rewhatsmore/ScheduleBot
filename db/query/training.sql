-- name: CreateTraining :one
INSERT INTO trainings (
date_and_time, place
) VALUES (
$1, $2
)
RETURNING *;

-- name: GetTraining :one
SELECT * FROM trainings
WHERE training_id = $1 LIMIT 1;

-- name: ListTrainings :many
SELECT * FROM trainings
WHERE date_and_time > now()
ORDER BY date_and_time;

-- name: ListLastWeekTrainings :many
SELECT * FROM trainings
WHERE date_and_time > now() - INTERVAL '7' DAY
ORDER BY date_and_time;

-- name: UpdateTraining :one
UPDATE trainings SET trainer = $2
WHERE training_id = $1
RETURNING *;

-- name: DeleteTraining :exec
DELETE FROM trainings WHERE training_id = $1;

-- name: ListTrainingsForSend :many
SELECT trainings.training_id, place, date_and_time, COALESCE (U.appointment_id, 0) AS appointment_id 
FROM trainings
LEFT JOIN (SELECT * FROM appointments WHERE user_id=$1) AS U
ON trainings.training_id = U.training_id
WHERE date_and_time > now()
ORDER BY date_and_time;

