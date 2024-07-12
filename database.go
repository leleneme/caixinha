package main

import (
	"database/sql"
	"log"
)

var sqlCreate = `
create table if not exists file_record (
  id blob(16) primary key,
  name text not null,
  size integer not null,
  content_type text not null,
  upload_timestamp integer not null,
  seconds_to_live integer not null
)
`

func initDatabase(name string) *sql.DB {
	db, err := sql.Open("sqlite3", name)
	if err != nil {
		log.Fatalf("Failed to open SQLite database: %s\n", err)
	}

	_, err = db.Exec(sqlCreate)
	if err != nil {
		db.Close()
		log.Fatalf("Failed to initialize DB tables: %s\n", err)
	}

	return db
}
