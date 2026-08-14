package repository

import (
	"database/sql"
	"fmt"
	"strings"

	"movies-api/entity"
)

type SQLiteGenreRepository struct {
	db *sql.DB
}

func NewSQLiteGenreRepository(db *sql.DB) *SQLiteGenreRepository {
	return &SQLiteGenreRepository{db: db}
}

type GenreRepository interface {
	Create(genre *entity.Genre) (int64, error)
	GetAll(moviesFlag bool) ([]entity.Genre, error)
	GetByID(id int) (entity.Genre, error)
	Update(id int, genre entity.GenrePatchRequest) (entity.Genre, error)
	Delete(id int, force bool) (int64, error)
	DeleteConnection(id int, movies []int) (int64, error)
}

func (g *SQLiteGenreRepository) Create(genre *entity.Genre) (int64, error) {
	tx, err := g.db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	query := `INSERT INTO genres (name, version) VALUES (?, ?);`
	genre.Name = nameStyle(genre.Name)
	result, err := tx.Exec(query, genre.Name, 1)
	if err != nil {
		return 0, fmt.Errorf("%w: genre already exists", entity.ErrAlreadyExists)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	if genre.MovieIds != nil {
		if err := CreateGenreConnection(tx, id, genre.MovieIds); err != nil {
			return 0, err
		}
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return id, nil
}
func (g *SQLiteGenreRepository) GetAll(moviesFlag bool) ([]entity.Genre, error) {
	query := `SELECT id, name, version FROM genres`
	rows, err := g.db.Query(query)
	if err != nil {
		return []entity.Genre{}, err
	}
	defer rows.Close()
	genres := []entity.Genre{}
	for rows.Next() {
		var id uint
		var name string
		var version int
		if err := rows.Scan(&id, &name, &version); err != nil {
			return []entity.Genre{}, err
		}

		if moviesFlag {
			movies, err := g.GetMovies(int(id))
			if err != nil {
				return []entity.Genre{}, err
			}
			genres = append(genres, entity.Genre{Id: id, Name: name, Version: version, Movies: movies})
		} else {
			genres = append(genres, entity.Genre{Id: id, Name: name, Version: version})
		}
	}
	return genres, nil
}
func (g *SQLiteGenreRepository) GetByID(id int) (entity.Genre, error) {
	query := `SELECT name, version FROM genres WHERE id = ?`
	row := g.db.QueryRow(query, id)
	var name string
	var version int
	err := row.Scan(&name, &version)
	if err == sql.ErrNoRows {
		return entity.Genre{}, fmt.Errorf("%w genre id: %d", entity.ErrNotFound, id)
	} else if err != nil {
		return entity.Genre{}, err
	}
	movies, err := g.GetMovies(int(id))
	if err != nil {
		return entity.Genre{}, err
	}
	return entity.Genre{Id: uint(id), Name: name, Version: version, Movies: movies}, nil
}
func (g *SQLiteGenreRepository) Update(id int, genre entity.GenrePatchRequest) (entity.Genre, error) {
	tx, err := g.db.Begin()
	if err != nil {
		return entity.Genre{}, err
	}
	defer tx.Rollback()
	query := `SELECT name, version FROM genres WHERE id =?`
	queryConnection := `SELECT movie_id FROM movie_genres WHERE genre_id = ?`
	row := tx.QueryRow(query, id)
	var name string
	var version int
	err = row.Scan(&name, &version)
	if err == sql.ErrNoRows {
		return entity.Genre{}, fmt.Errorf("%w genre id: %d", entity.ErrNotFound, id)
	} else if err != nil {
		return entity.Genre{}, err
	}
	if genre.Name != nil {
		name = *genre.Name
	}
	newQuery := `UPDATE genres SET name = ?, version= version+1 WHERE id = ? AND version = ?`
	result, err := tx.Exec(newQuery, name, id, genre.Version)
	if err != nil {
		return entity.Genre{}, err
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return entity.Genre{}, fmt.Errorf("%w: genre was updated by someone else, refetch and try again", entity.ErrVersionConflict)
	}
	if genre.MovieIds != nil {
		rowConnection, err := tx.Query(queryConnection, id)
		if err != nil {
			return entity.Genre{}, err
		}
		defer rowConnection.Close()
		var currentMoviesId []int
		for rowConnection.Next() {
			var idCurrent int
			err = rowConnection.Scan(&idCurrent)
			if err != nil {
				return entity.Genre{}, err
			}
			currentMoviesId = append(currentMoviesId, idCurrent)
		}
		toAdd, toDelete := computeMovieDiff(currentMoviesId, genre.MovieIds)
		err = CreateGenreConnection(tx, int64(id), toAdd)
		if err != nil {
			return entity.Genre{}, err
		}
		err = DeleteGenreConnection(tx, int64(id), toDelete)
		if err != nil {
			return entity.Genre{}, err
		}
	}
	if err := tx.Commit(); err != nil {
		return entity.Genre{}, err
	}
	return entity.Genre{Id: uint(id), Name: name, Version: version + 1}, nil
}
func (g *SQLiteGenreRepository) Delete(id int, force bool) (int64, error) {
	query := `SELECT name, version FROM genres WHERE id = ?`
	row := g.db.QueryRow(query, id)
	var name string
	var version int
	err := row.Scan(&name, &version)
	if err == sql.ErrNoRows {
		return 0, fmt.Errorf("%w genre id: %d", entity.ErrNotFound, id)
	} else if err != nil {
		return 0, err
	}
	countFilms := 0
	queryCount := `SELECT COUNT(*) FROM movie_genres WHERE genre_id = ?`
	if err := g.db.QueryRow(queryCount, id).Scan(&countFilms); err != nil {
		return 0, err
	}
	if countFilms > 0 && !force {
		return 0, fmt.Errorf("%w: %s is in %d films", entity.ErrHasRelations, name, id)
	}
	if force {
		queryDeleteConnection := `DELETE FROM movie_genres WHERE genre_id=?`
		_, err = g.db.Exec(queryDeleteConnection, id)
		if err != nil {
			return 0, err
		}
	}
	queryDelete := `DELETE FROM genres WHERE id = ?`
	result, err := g.db.Exec(queryDelete, id)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}
func (g *SQLiteGenreRepository) DeleteConnection(id int, movies []int) (int64, error) {
	str := make([]string, len(movies))
	args := []any{id}
	for i, movieID := range movies {
		str[i] = "?"
		args = append(args, movieID)
	}
	placeholder := strings.Join(str, ",")
	queryDeleteConnection := fmt.Sprintf(`DELETE FROM movie_genres WHERE genre_id=? AND movie_id IN (%s)`, placeholder)
	result, err := g.db.Exec(queryDeleteConnection, args...)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// helpers

func CreateGenreConnection(tx *sql.Tx, idGenre int64, idMovies []int) error {
	for _, id := range idMovies {
		_, err := tx.Exec(`INSERT OR IGNORE INTO movie_genres(movie_id, genre_id) VALUES (?,?)`, id, idGenre)
		if err != nil {
			return err
		}
	}
	return nil
}
func DeleteGenreConnection(tx *sql.Tx, idGenre int64, idMovies []int) error {
	if len(idMovies) == 0 {
		return nil
	}
	str := make([]string, len(idMovies))
	args := []any{idGenre}
	for i, movieID := range idMovies {
		str[i] = "?"
		args = append(args, movieID)
	}
	placeholder := strings.Join(str, ",")
	query := fmt.Sprintf(`DELETE FROM movie_genres WHERE genre_id=? AND movie_id IN (%s)`, placeholder)
	_, err := tx.Exec(query, args...)
	return err
}
func (g *SQLiteGenreRepository) GetMovies(id int) ([]entity.Movie, error) {
	query := `SELECT movies.id, movies.title, movies.release_year, movies.duration
	FROM movies
	JOIN movie_genres ON movies.id=movie_genres.movie_id
	WHERE movie_genres.genre_id = ?`
	rows, err := g.db.Query(query, id)
	if err != nil {
		return []entity.Movie{}, err
	}
	defer rows.Close()
	movies := []entity.Movie{}
	for rows.Next() {
		var id uint
		var title string
		var year int
		var duration int
		err := rows.Scan(&id, &title, &year, &duration)
		if err != nil {
			return []entity.Movie{}, err
		}
		movies = append(movies, entity.Movie{Id: id, Title: title, ReleaseYear: year, Duration: duration})
	}
	return movies, nil
}
