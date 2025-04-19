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

}
