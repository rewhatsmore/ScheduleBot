# ScheduleTelegramBot

The bot adds information about the new user to the database,
and gives him access to an interactive list of workouts,
where you can sign up for training or cancel the registration.
The bot reminds you of the upcoming workout and notifies you of the cancellation.
The bot also automatically generates a weekly schedule by copying
it from last week.

## Install (run commands in console)

* `git clone`
* `cp .env.example app.env`
* `make postgres` need Docker
* `make createdb`
* `make migrateup` need https://github.com/golang-migrate/migrate
* `go run .`

