package settings

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"log"
)

type Settings struct {
	db *sql.DB
}

func (this Settings) createGroupEntry(id uint64) {
	this.db.Exec(`INSERT INTO Groups (id) VALUES (?)`, id)
	// YOLO
}

func (this Settings) IsGroupBlacklisted(id uint64) bool {
	stmt := `SELECT blacklisted FROM Groups WHERE id=?`
	var fakebool int
	err := this.db.QueryRow(stmt, id).Scan(&fakebool)
	switch {
	case err == sql.ErrNoRows:
		return false
	case err != nil:
		log.Fatal(err)
	}
	return fakebool == 1
}

func (this Settings) SetGroupBlacklisted(id uint64, value bool) {
	this.createGroupEntry(id)
	stmt := `UPDATE Groups SET blacklisted=? WHERE id=?`
	if _, err := this.db.Exec(stmt, value, id); err != nil {
		log.Fatal(err)
	}
}

func (this Settings) IsGroupQuiet(id uint64) bool {
	stmt := `SELECT quiet FROM Groups WHERE id=?`
	var fakebool int
	err := this.db.QueryRow(stmt, id).Scan(&fakebool)
	switch {
	case err == sql.ErrNoRows:
		return false
	case err != nil:
		log.Fatal(err)
	}
	return fakebool == 1
}

func (this Settings) SetGroupQuiet(id uint64, value bool) {
	this.createGroupEntry(id)
	stmt := `UPDATE Groups SET quiet=? WHERE id=?`
	if _, err := this.db.Exec(stmt, value, id); err != nil {
		log.Fatal(err)
	}
}

func LoadSettings() Settings {
	db, err := sql.Open("sqlite3", "data/lewdbot.db")
	if err != nil {
		log.Fatalf("Opening settings: %s", err)
	}
	migrate(db)
	return Settings{db}
}

func (this Settings) Close() {
	this.db.Close()
}
