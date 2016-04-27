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

func (settings *Settings) createGroupEntry(id steamid.SteamId) {
	settings.db.Exec(`INSERT INTO Groups (id) VALUES (?)`, id)
	// YOLO
}

func (settings *Settings) createUserEntry(id steamid.SteamId) {
	settings.db.Exec(`INSERT INTO Users (id) VALUES (?)`, id)
}

// IsGroupBlacklisted looks up whether the group has been remembered as
// blacklisted.
func (settings *Settings) IsGroupBlacklisted(id steamid.SteamId) bool {
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
func (settings *Settings) SetGroupBlacklisted(id steamid.SteamId, value bool) {
	settings.createGroupEntry(id)
	stmt := `UPDATE Groups SET blacklisted=? WHERE id=?`
	if _, err := settings.db.Exec(stmt, value, id); err != nil {
		log.Fatal(err)
	}
}

// ListGroupBlacklisted returns all blacklisted groups ids
func (settings *Settings) ListGroupBlacklisted() []steamid.SteamId {
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

// IsGroupAutojoin looks up whether the group has been remembered as should be
// autojoined.
func (settings *Settings) IsGroupAutojoin(id steamid.SteamId) bool {
	stmt := `SELECT autojoin FROM Groups WHERE id=?`
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

// SetGroupAutojoin sets whether a group should be autojoined.
func (settings *Settings) SetGroupAutojoin(id steamid.SteamId, value bool) {
	settings.createGroupEntry(id)
	stmt := `UPDATE Groups SET autojoin=? WHERE id=?`
	if _, err := settings.db.Exec(stmt, value, id); err != nil {
		log.Fatal(err)
	}
}

// ListGroupAutojoin looks up all groups that are remembered as should be
// autojoined
func (settings *Settings) ListGroupAutojoin() []steamid.SteamId {
	stmt := `SELECT id FROM Groups WHERE autojoin=1`
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
func (settings *Settings) IsGroupQuiet(id steamid.SteamId) bool {
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
func (settings *Settings) SetGroupQuiet(id steamid.SteamId, value bool) {
	settings.createGroupEntry(id)
	stmt := `UPDATE Groups SET quiet=? WHERE id=?`
	if _, err := settings.db.Exec(stmt, value, id); err != nil {
		log.Fatal(err)
	}
}

// IsUserMaster looks up whether a user has been remembered as an admin.
func (settings *Settings) IsUserMaster(id steamid.SteamId) bool {
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

// SetUserMaster sets whether a user should be considered an admin.
func (settings *Settings) SetUserMaster(id steamid.SteamId, value bool) {
	settings.createUserEntry(id)
	stmt := `UPDATE Users SET admin=? WHERE id=?`
	if _, err := settings.db.Exec(stmt, value, id); err != nil {
		log.Fatal(err)
	}
}

// ListUserMaster lists all users that are considered admins.
func (settings *Settings) ListUserMaster() []steamid.SteamId {
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
		users = append(users, sid)
	}
	if err = rows.Err(); err != nil {
		log.Fatal(err)
	}
	return users
}

// SetUserBanned sets whether a user should be ignored
func (settings *Settings) SetUserBanned(id steamid.SteamId, value bool) {
	settings.createUserEntry(id)
	stmt := `UPDATE Users SET banned=? WHERE id=?`
	if _, err := settings.db.Exec(stmt, value, id); err != nil {
		log.Fatal(err)
	}
}

// IsUserBanned looks up whether a user has been set to be ignored
func (settings *Settings) IsUserBanned(id steamid.SteamId) bool {
	stmt := `SELECT banned FROM Users WHERE id=?`
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

// ListUserBanned lists all users that are considered subhumans
func (settings *Settings) ListUserBanned() []steamid.SteamId {
	stmt := `SELECT id FROM Users WHERE banned=1`
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
		users = append(users, sid)
	}
	if err = rows.Err(); err != nil {
		log.Fatal(err)
	}
	return users
}

// SetUserLastUse sets unix timestamp of last interaction with bot
func (settings *Settings) SetUserLastUse(id steamid.SteamId, value int64) {
	settings.createUserEntry(id)
	stmt := `UPDATE Users SET lastuse=? WHERE id=?`
	if _, err := settings.db.Exec(stmt, value, id); err != nil {
		log.Fatal(err)
	}
}

// GetUserLastUse gets unix timestamp of last interaction with bot
func (settings *Settings) GetUserLastUse(id steamid.SteamId) int64 {
	stmt := `SELECT lastuse FROM Users WHERE id=?`
	var lastuse int64
	err := settings.db.QueryRow(stmt, id).Scan(&lastuse)
	switch {
	case err == sql.ErrNoRows:
		return 0
	case err != nil:
		log.Fatal(err)
	}
	return lastuse
}

// LoadSettings should be called before anything else, and will give you the
// object you look up all your settings from.
func LoadSettings(databasename string) *Settings {
	db, err := sql.Open("sqlite3", databasename)
	if err != nil {
		log.Fatalf("Opening settings: %s", err)
	}
	migrate(db)
	return &Settings{db}
}

// Close tears down the database. Don't forget this!
func (settings *Settings) Close() {
	settings.db.Close()
}
