package db

import (
	"database/sql"
	"os"
	"strconv"

	_ "modernc.org/sqlite"
)

var DB *sql.DB

func InitDB() {
	var err error

	// Load environment variables
	dbDriver := os.Getenv("DB_DRIVER")
	if dbDriver == "" {
		panic("DB_DRIVER environment variable is not set!")
	}

	dbDataSource := os.Getenv("DB_DATASOURCE")
	if dbDataSource == "" {
		panic("DB_DATASOURCE environment variable is not set!")
	}

	DB, err = sql.Open(dbDriver, dbDataSource)

	if err != nil {
		panic("Could not connect to the database!")
	}
	maxOpenConnections := os.Getenv("DB_MAX_OPEN_CONNECTIONS")
	if maxOpenConnections == "" {
		panic("DB_MAX_OPEN_CONNECTIONS environment variable is not set!")
	}
	maxIdleConnections := os.Getenv("DB_MAX_IDLE_CONNECTIONS")
	if maxIdleConnections == "" {
		panic("DB_MAX_IDLE_CONNECTIONS environment variable is not set!")
	}

	maxOpenConnsInt, err := strconv.Atoi(maxOpenConnections)
	if err != nil {
		panic("DB_MAX_OPEN_CONNECTIONS must be an integer!")
	}
	maxIdleConnsInt, err := strconv.Atoi(maxIdleConnections)
	if err != nil {
		panic("DB_MAX_IDLE_CONNECTIONS must be an integer!")
	}

	DB.SetMaxOpenConns(maxOpenConnsInt)
	DB.SetMaxIdleConns(maxIdleConnsInt)

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
		notes TEXT,
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
		UNIQUE (itinerary_id, arrival_date, departure_date, city, country)
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
