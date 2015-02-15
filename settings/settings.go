// Package settings handles setting up, tearing down, and migrating SQL stuff
// and in the end gives you a bunch of easy functions to use for looking up and
// setting settings.
package settings

import (
	"database/sql"
	"github.com/Philipp15b/go-steam/steamid"
	_ "github.com/mattn/go-sqlite3" // sql driver
	"log"
)

// Settings is the holder struct for the backend database handler and is also
// the object you call all your functions from.
type Settings struct {
	db *sql.DB
}

func (settings Settings) createGroupEntry(id steamid.SteamId) {
	settings.db.Exec(`INSERT INTO Groups (id) VALUES (?)`, id)
	// YOLO
}

func (settings Settings) createUserEntry(id steamid.SteamId) {
	settings.db.Exec(`INSERT INTO Users (id) VALUES (?)`, id)
}

// IsGroupBlacklisted looks up whether the group has been remembered as
// blacklisted.
func (settings Settings) IsGroupBlacklisted(id steamid.SteamId) bool {
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
func (settings Settings) SetGroupBlacklisted(id steamid.SteamId, value bool) {
	settings.createGroupEntry(id)
	stmt := `UPDATE Groups SET blacklisted=? WHERE id=?`
	if _, err := settings.db.Exec(stmt, value, id); err != nil {
		log.Fatal(err)
	}
}

func (settings Settings) ListGroupBlacklisted() []steamid.SteamId {
	stmt := `SELECT id FROM Groups WHERE blacklisted=1`
	rows, err := settings.db.Query(stmt)
	if err != nil {
		log.Fatal(err)
	}
	var groups []steamid.SteamId
	for rows.Next() {
		var id string
		err = rows.Scan(&id)
		if err != nil {
			log.Fatal(err)
		}
		sid, err := steamid.NewId(id)
		if err != nil {
			log.Fatal(err)
		}
		groups = append(groups, sid)
	}
	if err = rows.Err(); err != nil {
		log.Fatal(err)
	}
	return groups
}

// IsGroupQuiet looks up whether the group has been remembered as should be
// treated quietly.
func (settings Settings) IsGroupQuiet(id steamid.SteamId) bool {
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
func (settings Settings) SetGroupQuiet(id steamid.SteamId, value bool) {
	settings.createGroupEntry(id)
	stmt := `UPDATE Groups SET quiet=? WHERE id=?`
	if _, err := settings.db.Exec(stmt, value, id); err != nil {
		log.Fatal(err)
	}
}

// IsUserAdmin looks up whether a user has been remembered as an admin.
func (settings Settings) IsUserAdmin(id steamid.SteamId) bool {
	stmt := `SELECT admin FROM Users WHERE id=?`
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

// SetUserAdmin sets whether a user should be considered an admin.
func (settings Settings) SetUserAdmin(id steamid.SteamId, value bool) {
	settings.createUserEntry(id)
	stmt := `UPDATE Users SET admin=? WHERE id=?`
	if _, err := settings.db.Exec(stmt, value, id); err != nil {
		log.Fatal(err)
	}
}

// ListUserAdmin lists all users that are considered admins.
func (settings Settings) ListUserAdmin() []steamid.SteamId {
	stmt := `SELECT id FROM Users WHERE admin=1`
	rows, err := settings.db.Query(stmt)
	if err != nil {
		log.Fatal(err)
	}
	var users []steamid.SteamId
	for rows.Next() {
		var id string
		err = rows.Scan(&id)
		if err != nil {
			log.Fatal(err)
		}
		sid, err := steamid.NewId(id)
		if err != nil {
			log.Fatal(err)
		}
		groups = append(groups, sid)
	}
	if err = rows.Err(); err != nil {
		log.Fatal(err)
	}
	return users
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
