package db

import (
	"database/sql"

	_ "modernc.org/sqlite"
)

var DB *sql.DB

func InitDB() {
	var err error
	DB, err = sql.Open("sqlite", "api.sql")

	if err != nil {
		panic("Could not connect to the database!")
	}

	DB.SetMaxOpenConns(10)
	DB.SetMaxIdleConns(5)

	createTables()
}

func createTables() {

	createUsersTable := ` 
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		email TEXT NOT NULL UNIQUE,
		password TEXT NOT NULL,
		creation_date DATETIME NOT NULL,
		update_date DATETIME NOT NULL,
		last_login_date DATETIME
	)
	`
	_, err := DB.Exec(createUsersTable)

	if err != nil {
		panic("Could not create users table!")
	}

	createItinerariesTable := `
	CREATE TABLE IF NOT EXISTS itineraries (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		description TEXT NOT NULL,
		owner_id INTEGER NOT NULL,
		creation_date DATETIME NOT NULL,
		update_date DATETIME NOT NULL,
		FOREIGN KEY (owner_id) REFERENCES users(id)
	)
	`
	_, err = DB.Exec(createItinerariesTable)
	if err != nil {
		panic("Could not create itineraries table!")
	}

	createItinerariesTravelDestinationsTable := `
	CREATE TABLE IF NOT EXISTS itinerary_travel_destinations (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		country TEXT NOT NULL,
		city TEXT NOT NULL,
		itinerary_id INTEGER NOT NULL,
		arrival_date DATETIME NOT NULL,
		departure_date DATETIME NOT NULL,
		creation_date DATETIME NOT NULL,
		update_date DATETIME NOT NULL,
		FOREIGN KEY (itinerary_id) REFERENCES itineraries(id),
		UNIQUE (country, city, arrival_date, departure_date, itinerary_id)
	)
	`
	_, err = DB.Exec(createItinerariesTravelDestinationsTable)
	if err != nil {
		panic("Could not create itineraries travel destinations table!")
	}

	createItinerariesFileJobsTable := `
	CREATE TABLE IF NOT EXISTS itinerary_file_jobs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		itinerary_id INTEGER NOT NULL,
		status TEXT NOT NULL,
		status_description TEXT,
		creation_date DATETIME NOT NULL,
		start_date DATETIME,
		end_date DATETIME,
		file_path TEXT,
		file_manager VARCHAR(64) NOT NULL,
		async_task_id VARCHAR(64),
		FOREIGN KEY (itinerary_id) REFERENCES itineraries(id)
	)
	`
	_, err = DB.Exec(createItinerariesFileJobsTable)
	if err != nil {
		panic("Could not create itineraries file jobs table!")
	}

}

func HandleTransaction(tx *sql.Tx, err *error) {
	if p := recover(); p != nil {
		tx.Rollback()
		panic(p)
	} else if *err != nil {
		tx.Rollback()
	} else {
		*err = tx.Commit()
	}
}
