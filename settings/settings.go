// Package settings handles setting up, tearing down, and migrating SQL stuff
// and in the end gives you a bunch of easy functions to use for looking up and
// setting settings.
package settings

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3" // sql driver
	"log"
)

// Settings is the holder struct for the backend database handler and is also
// the object you call all your functions from.
type Settings struct {
	db *sql.DB
}

func (settings Settings) createGroupEntry(id uint64) {
	settings.db.Exec(`INSERT INTO Groups (id) VALUES (?)`, id)
	// YOLO
}

// IsGroupBlacklisted looks up whether the group has been remembered as
// blacklisted.
func (settings Settings) IsGroupBlacklisted(id uint64) bool {
	stmt := `SELECT blacklisted FROM Groups WHERE id=?`
	var fakebool int
	err := settings.db.QueryRow(stmt, id).Scan(&fakebool)
	switch {
	case err == sql.ErrNoRows:
		return false
	case err != nil:
		log.Fatal(err)
	}
	return fakebool == 1
}

// SetGroupBlacklisted sets whether a group should be considered blacklisted.
func (settings Settings) SetGroupBlacklisted(id uint64, value bool) {
	settings.createGroupEntry(id)
	stmt := `UPDATE Groups SET blacklisted=? WHERE id=?`
	if _, err := settings.db.Exec(stmt, value, id); err != nil {
		log.Fatal(err)
	}
}

// IsGroupQuiet looks up whether the group has been remembered as should be
// treated quietly.
func (settings Settings) IsGroupQuiet(id uint64) bool {
	stmt := `SELECT quiet FROM Groups WHERE id=?`
	var fakebool int
	err := settings.db.QueryRow(stmt, id).Scan(&fakebool)
	switch {
	case err == sql.ErrNoRows:
		return false
	case err != nil:
		log.Fatal(err)
	}
	return fakebool == 1
}

// SetGroupQuiet sets whether a group should be treated quietly.
func (settings Settings) SetGroupQuiet(id uint64, value bool) {
	settings.createGroupEntry(id)
	stmt := `UPDATE Groups SET quiet=? WHERE id=?`
	if _, err := settings.db.Exec(stmt, value, id); err != nil {
		log.Fatal(err)
	}
}

// LoadSettings should be called before anything else, and will give you the
// object you look up all your settings from.
func LoadSettings(databasename string) Settings {
	db, err := sql.Open("sqlite3", databasename)
	if err != nil {
		log.Fatalf("Opening settings: %s", err)
	}
	migrate(db)
	return Settings{db}
}

// Close tears down the database. Don't forget this!
func (settings Settings) Close() {
	settings.db.Close()
}
