package settings

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"log"
)

var totalMigrations = 1

func migrate(db *sql.DB) {
	rows, err := db.Query("SELECT * FROM Migrations")
	if err != nil {
		migrateAll(db)
		return
	}
	defer rows.Close()
	latestMigration := 0
	for rows.Next() {
		if err := rows.Scan(&latestMigration); err != nil {
			log.Fatal(err)
		}
		log.Printf("Migration #%d already in place", latestMigration)
	}
	if totalMigrations-latestMigration == 0 {
		return
	}
	log.Printf("%d migrations behind...", totalMigrations-latestMigration)
	// wish I could dynamically call this or something
	if latestMigration == 0 {
		migrate1(db)
		latestMigration = 1
	}
}

func migrateAll(db *sql.DB) {
	stmt := `CREATE TABLE Migrations (
             id INT PRIMARY KEY NOT NULL
           );`
	log.Print(stmt)
	if _, err := db.Exec(stmt); err != nil {
		log.Fatalf("Failed to create Migrations table! %s", err)
	}
	migrate1(db)
}

func migrate1(db *sql.DB) {
	log.Print("Migration #1")
	stmt := `CREATE TABLE Groups (
             id VARCHAR(17) PRIMARY KEY NOT NULL,
             blacklisted INT DEFAULT 0,
             quiet INT DEFAULT 0
           );`
	log.Print(stmt)
	if _, err := db.Exec(stmt); err != nil {
		log.Fatalf("Failed! %s", err)
	}
	db.Exec(`INSERT INTO migrations VALUES (1)`)
}
