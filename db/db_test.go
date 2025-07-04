package db

import (
	"database/sql"
	"os"
	"testing"
)

func setTestEnv() {
	// Use a named in-memory database with shared cache so all connections share the same DB
	os.Setenv("DB_DRIVER", "sqlite")
	os.Setenv("DB_DATASOURCE", "file::memory:?cache=shared")
	os.Setenv("DB_MAX_OPEN_CONNECTIONS", "1")
	os.Setenv("DB_MAX_IDLE_CONNECTIONS", "1")
}

func unsetTestEnv() {
	os.Unsetenv("DB_DRIVER")
	os.Unsetenv("DB_DATASOURCE")
	os.Unsetenv("DB_MAX_OPEN_CONNECTIONS")
	os.Unsetenv("DB_MAX_IDLE_CONNECTIONS")
}

func TestInitDB_Success(t *testing.T) {
	setTestEnv()
	defer unsetTestEnv()

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("InitDB panicked: %v", r)
		}
		if DB != nil {
			DB.Close()
		}
	}()

	InitDB()

	if DB == nil {
		t.Fatal("DB should not be nil after InitDB")
	}

	// Check if tables exist
	tables := []string{"users", "itineraries", "itinerary_travel_destinations", "itinerary_file_jobs", "audit_events"}
	for _, table := range tables {
		query := "SELECT name FROM sqlite_master WHERE type='table' AND name=?"
		row := DB.QueryRow(query, table)
		var name string
		err := row.Scan(&name)
		if err != nil {
			t.Errorf("Table %s does not exist: %v", table, err)
		} else if name != table {
			t.Errorf("Expected table name %s, got %s", table, name)
		}
	}
}

func TestInitDB_MissingEnvVars(t *testing.T) {
	unsetTestEnv()
	defer unsetTestEnv()

	tests := []struct {
		name string
		set  func()
		want string
	}{
		{
			name: "missing DB_DRIVER",
			set:  func() {},
			want: "DB_DRIVER environment variable is not set!",
		},
		{
			name: "missing DB_DATASOURCE",
			set: func() {
				os.Setenv("DB_DRIVER", "sqlite")
			},
			want: "DB_DATASOURCE environment variable is not set!",
		},
		{
			name: "missing DB_MAX_OPEN_CONNECTIONS",
			set: func() {
				os.Setenv("DB_DRIVER", "sqlite")
				os.Setenv("DB_DATASOURCE", ":memory:")
			},
			want: "DB_MAX_OPEN_CONNECTIONS environment variable is not set!",
		},
		{
			name: "missing DB_MAX_IDLE_CONNECTIONS",
			set: func() {
				os.Setenv("DB_DRIVER", "sqlite")
				os.Setenv("DB_DATASOURCE", ":memory:")
				os.Setenv("DB_MAX_OPEN_CONNECTIONS", "1")
			},
			want: "DB_MAX_IDLE_CONNECTIONS environment variable is not set!",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			unsetTestEnv()
			tt.set()
			defer func() {
				if r := recover(); r == nil {
					t.Errorf("Expected panic for %s, but did not panic", tt.name)
				} else if r != tt.want {
					t.Errorf("Expected panic message %q, got %q", tt.want, r)
				}
			}()
			InitDB()
		})
	}
}

func TestInitDB_InvalidConnectionNumbers(t *testing.T) {
	setTestEnv()
	defer unsetTestEnv()

	os.Setenv("DB_MAX_OPEN_CONNECTIONS", "notanint")
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic for invalid DB_MAX_OPEN_CONNECTIONS, but did not panic")
		}
	}()
	InitDB()
}

func TestHandleTransaction_Commit(t *testing.T) {
	setTestEnv()
	defer unsetTestEnv()
	InitDB()
	defer DB.Close()

	tx, err := DB.Begin()
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}
	var txErr error
	HandleTransaction(tx, &txErr)
	if txErr != nil {
		t.Errorf("Expected nil error after commit, got %v", txErr)
	}
}

func TestHandleTransaction_RollbackOnError(t *testing.T) {
	setTestEnv()
	defer unsetTestEnv()
	InitDB()
	defer DB.Close()

	tx, err := DB.Begin()
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}
	someErr := sql.ErrConnDone
	HandleTransaction(tx, &someErr)
	// No panic expected, just rollback
}

func TestHandleTransaction_Panic(t *testing.T) {
	setTestEnv()
	defer unsetTestEnv()
	InitDB()
	defer DB.Close()

	tx, err := DB.Begin()
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic to propagate from HandleTransaction, but did not panic")
		}
	}()
	func() {
		defer HandleTransaction(tx, &err)
		panic("test panic")
	}()
}
