package settings

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3" // sql driver
	"log"
)

var totalMigrations = 5

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
	if latestMigration == 1 {
		migrate2(db)
		latestMigration = 2
	}
	if latestMigration == 2 {
		migrate3(db)
		latestMigration = 3
	}
	if latestMigration == 3 {
		migrate4(db)
		latestMigration = 4
	}
	if latestMigration == 4 {
		migrate5(db)
		latestMigration = 5
	}
}

func execAndPrint(db *sql.DB, stmt string) {
	log.Print(stmt)
	if _, err := db.Exec(stmt); err != nil {
		log.Fatalf("Failed! %s", err)
	}
}

func migrateAll(db *sql.DB) {
	execAndPrint(db, `CREATE TABLE Migrations (
                      id INT PRIMARY KEY NOT NULL
                    );`)
	migrate1(db)
	migrate2(db)
	migrate3(db)
	migrate4(db)
	migrate5(db)
}

func migrate1(db *sql.DB) {
	log.Print("Migration #1")

	execAndPrint(db, `CREATE TABLE Groups (
                      id VARCHAR(17) PRIMARY KEY NOT NULL,
                      blacklisted INT DEFAULT 0,
                      quiet INT DEFAULT 0
                    );`)

	db.Exec(`INSERT INTO migrations VALUES (1)`)
}

func migrate2(db *sql.DB) {
	log.Print("Migration #2")

	db.Exec(`BEGIN TRANSACTION`)
	execAndPrint(db, `ALTER TABLE Groups RENAME TO Groups_tmp`)
	execAndPrint(db, `CREATE TABLE Groups (
                      id VARCHAR(19) PRIMARY KEY NOT NULL,
                      blacklisted INT DEFAULT 0,
                      quiet INT DEFAULT 0
                    );`)
	execAndPrint(db, `INSERT INTO Groups (id, blacklisted, quiet)
                    SELECT id, blacklisted, quiet
                    FROM Groups_tmp`)
	execAndPrint(db, `DROP TABLE Groups_tmp`)
	db.Exec(`COMMIT`)

	execAndPrint(db, `CREATE TABLE Users (
                      id VARCHAR(19),
                      admin INT DEFAULT 0
                    );`)

	db.Exec(`INSERT INTO migrations VALUES (2)`)
}

func migrate3(db *sql.DB) {
	log.Print("Migration #3")

	execAndPrint(db, `ALTER TABLE Groups ADD COLUMN autojoin INT DEFAULT 0`)

	db.Exec(`INSERT INTO migrations VALUES (3)`)
}

func migrate4(db *sql.DB) {
	log.Print("Migration #4")

	execAndPrint(db, `ALTER TABLE Users ADD COLUMN banned INT DEFAULT 0`)

	db.Exec(`INSERT INTO migrations VALUES (4)`)
}

func migrate5(db *sql.DB) {
	log.Print("Migration #5")

	execAndPrint(db, `ALTER TABLE Users ADD COLUMN lastuse INT DEFAULT 0`)

	db.Exec(`INSERT INTO migrations VALUES (5)`)
}
