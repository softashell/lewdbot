package settings

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"log"
)

func createGroupEntry(db *sql.DB, id uint64) {
	db.Exec(`INSERT INTO Groups (id) VALUES (?)`, id)
	// YOLO
}

func IsGroupBlacklisted(db *sql.DB, id uint64) bool {
	stmt := `SELECT blacklisted FROM Groups WHERE id=?`
	var fakebool int
	err := db.QueryRow(stmt, id).Scan(&fakebool)
	switch {
	case err == sql.ErrNoRows:
		return false
	case err != nil:
		log.Fatal(err)
	}
	return fakebool == 1
}

func SetGroupBlacklisted(db *sql.DB, id uint64, value bool) {
	createGroupEntry(db, id)
	stmt := `UPDATE Groups SET blacklisted=? WHERE id=?`
	if _, err := db.Exec(stmt, value, id); err != nil {
		log.Fatal(err)
	}
}

func IsGroupQuiet(db *sql.DB, id uint64) bool {
	stmt := `SELECT quiet FROM Groups WHERE id=?`
	var fakebool int
	err := db.QueryRow(stmt, id).Scan(&fakebool)
	switch {
	case err == sql.ErrNoRows:
		return false
	case err != nil:
		log.Fatal(err)
	}
	return fakebool == 1
}

func SetGroupQuiet(db *sql.DB, id uint64, value bool) {
	createGroupEntry(db, id)
	stmt := `UPDATE Groups SET quiet=? WHERE id=?`
	if _, err := db.Exec(stmt, value, id); err != nil {
		log.Fatal(err)
	}
}

func Load() *sql.DB {
	db, err := sql.Open("sqlite3", "data/lewdbot.db")
	if err != nil {
		log.Fatalf("Opening settings: %s", err)
	}
	migrate(db)
	return db
}
