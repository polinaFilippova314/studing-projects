package db

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

// create table relationship
func CreateTables(db *sql.DB) (sql.Result, error) {
	query := `
	CREATE TABLE IF NOT EXISTS movies (
		id INTEGER PRIMARY KEY,
		title TEXT NOT NULL,
		release_year INTEGER NOT NULL,
		duration INTEGER NOT NULL,
		version INTEGER NOT NULL DEFAULT 1,
		UNIQUE(title, release_year)
	);

	CREATE TABLE IF NOT EXISTS actors (
		id INTEGER PRIMARY KEY,
		name TEXT NOT NULL,
		birthdate TEXT NOT NULL,
		version INTEGER NOT NULL DEFAULT 1,
		UNIQUE(name, birthdate)
	);

	CREATE TABLE IF NOT EXISTS genres (
		id INTEGER PRIMARY KEY,
		name TEXT NOT NULL UNIQUE,
		version INTEGER NOT NULL DEFAULT 1
	);

	CREATE TABLE IF NOT EXISTS movie_actors (
		movie_id INTEGER NOT NULL,
		actor_id INTEGER NOT NULL,
		PRIMARY KEY(movie_id, actor_id),
		FOREIGN KEY(movie_id) REFERENCES movies(id),
		FOREIGN KEY(actor_id) REFERENCES actors(id)
	);

	CREATE TABLE IF NOT EXISTS movie_genres (
		movie_id INTEGER NOT NULL,
		genre_id INTEGER NOT NULL,
		PRIMARY KEY(movie_id, genre_id),
		FOREIGN KEY(movie_id) REFERENCES movies(id),
		FOREIGN KEY(genre_id) REFERENCES genres(id)
	);
	`

	return db.Exec(query)
}
