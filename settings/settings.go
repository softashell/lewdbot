package settings

import (
  "database/sql"
  _ "github.com/mattn/go-sqlite3"
  "log"
)

var totalMigrations int = 1

func Migrate(db *sql.DB) {
  rows, err := db.Query("SELECT * FROM Migrations")
  if err != nil {
    MigrateAll(db)
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
  if totalMigrations - latestMigration == 0 {
    return
  }
  log.Printf("%d migrations behind...", totalMigrations - latestMigration)
  // wish I could dynamically call this or something
  if latestMigration == 0 {
    Migrate1(db)
    latestMigration = 1
  }
}

func MigrateAll(db *sql.DB) {
  stmt := `CREATE TABLE Migrations (
             id INT PRIMARY KEY NOT NULL
           );`
  log.Print(stmt)
  if _, err := db.Exec(stmt); err != nil {
    log.Fatalf("Failed to create Migrations table! %s", err)
  }
  Migrate1(db)
}

func Migrate1(db *sql.DB) {
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
  Migrate(db)
  return db
}
