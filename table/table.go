package table

import (
	"context"
	"log"
	"time"

	"github.com/robfig/cron/v3"
	db "schedule.sqlc.dev/app/db/sqlc"
)

func StartcCreateTable(queries *db.Queries) {
	c := cron.New()
	c.AddFunc("0 14 * * 0", func() {
		trainings, err := queries.ListLastWeekTrainings(context.Background())
		if err != nil {
			log.Println(err)
		}
		for _, training := range trainings {
			arg := db.CreateTrainingParams{
				Place:       training.Place,
				DateAndTime: training.DateAndTime.Add(time.Hour * 24 * 7),
			}
			trainingNew, err := queries.CreateTraining(context.Background(), arg)
			log.Println("inserted:", trainingNew.TrainingID, trainingNew.DateAndTime)
			if err != nil {
				log.Println(err)
			}
		}
	})
	c.Start()
}
