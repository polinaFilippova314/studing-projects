package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"movies-api/customerrors"
	"movies-api/entity"
	"strings"
	"time"

	sqlite3 "github.com/mattn/go-sqlite3"
)

type SQLiteMovieRepository struct {
	db *sql.DB
}

func NewSQLiteMovieRepository(db *sql.DB) *SQLiteMovieRepository {
	return &SQLiteMovieRepository{db: db}
}

type MovieRepository interface {
	FindAll() ([]*entity.Movie, error)
	FindWithPagination(page, size int) ([]*entity.Movie, error)
	FindById(id int) (*entity.Movie, error)
	FindByGenre(genreId int) ([]*entity.Movie, error)
	FindByYear(year int) ([]*entity.Movie, error)
	FindByActor(actorId int) ([]*entity.Movie, error)
	FindActors(id int) ([]entity.Actor, error)

	Create(movie *entity.Movie) (int64, error)
	Update(id int, patch *entity.MoviePatch) (int64, error)
	Delete(id int, force bool) (int64, error)
	//extra
	FindByExactTitle(title string) (*entity.Movie, error)
	FindByTitleContains(title string) ([]*entity.Movie, error)
}

func (r *SQLiteMovieRepository) FindAll() ([]*entity.Movie, error) {
	queryMoviesTable := `SELECT * FROM movies ORDER BY Id`

	rows, err := r.db.Query(queryMoviesTable)
	if err != nil {
		return nil, fmt.Errorf("%w: select movies: %w", customerrors.ErrDB, err)
	}
	defer rows.Close()

	movies := []*entity.Movie{}

	for rows.Next() {
		m := &entity.Movie{}
		err := rows.Scan(&m.Id, &m.Title, &m.ReleaseYear, &m.Duration, &m.Version)
		if err != nil {
			return nil, fmt.Errorf("%w: scan movie row: %w", customerrors.ErrDB, err)
		}
		movies = append(movies, m)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%w: iterate movie rows: %w", customerrors.ErrDB, err)
	}

	if err := r.populateRelations(movies); err != nil {
		return nil, err
	}

	return movies, nil

}

func (r *SQLiteMovieRepository) FindWithPagination(page, size int) ([]*entity.Movie, error) {
	offset := page * size

	queryMoviesTable := `SELECT * FROM movies ORDER BY Id LIMIT ? OFFSET ?`

	rows, err := r.db.Query(queryMoviesTable, size, offset)
	if err != nil {
		return nil, fmt.Errorf("%w: select movies: %w", customerrors.ErrDB, err)
	}
	defer rows.Close()

	movies := []*entity.Movie{}

	for rows.Next() {
		m := &entity.Movie{}
		err := rows.Scan(&m.Id, &m.Title, &m.ReleaseYear, &m.Duration, &m.Version)
		if err != nil {
			return nil, fmt.Errorf("%w: scan movie row: %w", customerrors.ErrDB, err)
		}
		movies = append(movies, m)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%w: iterate movie rows: %w", customerrors.ErrDB, err)
	}

	if err := r.populateRelations(movies); err != nil {
		return nil, err
	}

	return movies, nil

}

func (r *SQLiteMovieRepository) FindById(id int) (*entity.Movie, error) {
	queryMoviesTable := `SELECT * FROM movies WHERE id = ?`
	row := r.db.QueryRow(queryMoviesTable, id)
	movie := &entity.Movie{}

	err := row.Scan(&movie.Id, &movie.Title, &movie.ReleaseYear, &movie.Duration, &movie.Version)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w: not found (movieId=%d): %w", customerrors.ErrNotFound, id, err)
		}
		return nil, fmt.Errorf("%w: select movie by id: %w", customerrors.ErrDB, err)
	}

	if err := r.populateRelations([]*entity.Movie{movie}); err != nil {
		return nil, err
	}

	return movie, nil
}

func (r *SQLiteMovieRepository) FindByGenre(genreId int) ([]*entity.Movie, error) {

	query := `
        SELECT 
            m.id, 
            m.title, 
            m.release_year,
            m.duration,
            m.version
        FROM movies m
        JOIN movie_genres mg
        ON mg.movie_id = m.id
        WHERE mg.genre_id = ?
        ORDER BY m.id;
    `

	rows, err := r.db.Query(query, genreId)
	if err != nil {
		return nil, fmt.Errorf("%w: select movies by genre: %w", customerrors.ErrDB, err)
	}
	defer rows.Close()

	foundMovies := []*entity.Movie{}

	for rows.Next() {
		m := &entity.Movie{}
		err := rows.Scan(&m.Id, &m.Title, &m.ReleaseYear, &m.Duration, &m.Version)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, fmt.Errorf("%w: not found (genreId=%d): %w", customerrors.ErrNotFound, genreId, err)
			}
			return nil, fmt.Errorf("%w: scan movie row: %w", customerrors.ErrDB, err)
		}
		foundMovies = append(foundMovies, m)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("%w: iterate movie rows: %w", customerrors.ErrDB, err)
	}

	if err := r.populateRelations(foundMovies); err != nil {
		return nil, err
	}

	return foundMovies, nil
}

func (r *SQLiteMovieRepository) FindByYear(year int) ([]*entity.Movie, error) {

	query := `SELECT * FROM movies WHERE release_year = ?`

	rows, err := r.db.Query(query, year)
	if err != nil {
		return nil, fmt.Errorf("%w: select movies by year: %w", customerrors.ErrDB, err)
	}
	defer rows.Close()

	movies := []*entity.Movie{}

	for rows.Next() {
		m := &entity.Movie{}

		if err := rows.Scan(&m.Id, &m.Title, &m.ReleaseYear, &m.Duration, &m.Version); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, fmt.Errorf("%w: not found (year=%d): %w", customerrors.ErrNotFound, year, err)
			}
			return nil, fmt.Errorf("%w: scan movie row: %w", customerrors.ErrDB, err)
		}
		movies = append(movies, m)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("%w: iterate movie rows: %w", customerrors.ErrDB, err)
	}

	if err := r.populateRelations(movies); err != nil {
		return nil, err
	}

	return movies, nil
}

func (r *SQLiteMovieRepository) FindByActor(actorId int) ([]*entity.Movie, error) {

	query := `
        SELECT 
            m.id, 
            m.title, 
            m.release_year,
            m.duration,
            m.version
        FROM movies m
        JOIN movie_actors ma
        ON ma.movie_id = m.id
        WHERE ma.actor_id = ?
        ORDER BY m.id;
    `

	rows, err := r.db.Query(query, actorId)
	if err != nil {
		return nil, fmt.Errorf("%w: select movies by actor: %w", customerrors.ErrDB, err)
	}
	defer rows.Close()

	foundMovies := []*entity.Movie{}

	for rows.Next() {
		m := &entity.Movie{}
		err := rows.Scan(&m.Id, &m.Title, &m.ReleaseYear, &m.Duration, &m.Version)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, fmt.Errorf("%w: not found (actorId=%d): %w", customerrors.ErrNotFound, actorId, err)
			}
			return nil, fmt.Errorf("%w: scan movie row: %w", customerrors.ErrDB, err)
		}
		foundMovies = append(foundMovies, m)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("%w: iterate movie rows: %w", customerrors.ErrDB, err)
	}

	if err := r.populateRelations(foundMovies); err != nil {
		return nil, err
	}

	return foundMovies, nil
}

func (r *SQLiteMovieRepository) FindActors(id int) ([]entity.Actor, error) {
	movie, err := r.FindById(id)

	if err != nil {
		return nil, err
	}

	actors := []entity.Actor{}

	for _, actor := range movie.Actors {
		actors = append(actors, actor)
	}

	return actors, nil
}

func (r *SQLiteMovieRepository) Create(movie *entity.Movie) (int64, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return 0, fmt.Errorf("%w: transaction movie create start: %w", customerrors.ErrDB, err)
	}

	defer tx.Rollback()

	queryForMovies := `INSERT INTO movies (title, release_year, duration) VALUES (?, ?, ?);`
	result, err := tx.Exec(queryForMovies, movie.Title, movie.ReleaseYear, movie.Duration)

	if err != nil {
		if isUniqueConstraint(err) {
			return 0, fmt.Errorf("%w: movie '%s' released in %d already exists", customerrors.ErrConflict,
				movie.Title,
				movie.ReleaseYear,
			)
		}
		return 0, fmt.Errorf("%w: insert movie: %w", customerrors.ErrDB, err)
	}

	movieId, err := result.LastInsertId()

	if err != nil {
		return 0, fmt.Errorf("%w: get inserted movie id: %w", customerrors.ErrDB, err)
	}

	queryForMovieActors := `INSERT INTO movie_actors (movie_id, actor_id) VALUES (?, ?);`

	for _, actor := range movie.Actors {
		_, err := tx.Exec(queryForMovieActors, movieId, actor.Id)
		if err != nil {
			var sqliteErr sqlite3.Error
			if errors.As(err, &sqliteErr) && sqliteErr.ExtendedCode == sqlite3.ErrConstraintForeignKey {
				return 0, fmt.Errorf("%w: non existing (actorId=%d): %w", customerrors.ErrInvalidReference, actor.Id, err)
			}
			return 0, fmt.Errorf("%w: insert movie_actors (movie=%d, actor=%d): %w", customerrors.ErrDB, movieId, actor.Id, err)
		}
	}

	queryForMovieGenres := `INSERT INTO movie_genres (movie_id, genre_id) VALUES (?, ?);`

	for _, genre := range movie.Genres {
		_, err := tx.Exec(queryForMovieGenres, movieId, genre.Id)
		if err != nil {
			var sqliteErr sqlite3.Error
			if errors.As(err, &sqliteErr) && sqliteErr.ExtendedCode == sqlite3.ErrConstraintForeignKey {
				return 0, fmt.Errorf("%w: non existing (genreId=%d): %w", customerrors.ErrInvalidReference, genre.Id, err)
			}
			return 0, fmt.Errorf("%w: insert movie_genres (movie=%d, genre=%d): %w", customerrors.ErrDB, movieId, genre.Id, err)
		}
	}

	err = tx.Commit()
	if err != nil {
		return 0, fmt.Errorf("%w: transaction movie create commit: %w", customerrors.ErrDB, err)
	}

	movie.Id = uint(movieId)

	return movieId, nil
}

func (r *SQLiteMovieRepository) Update(id int, patch *entity.MoviePatch) (int64, error) {

	tx, err := r.db.Begin()
	if err != nil {
		return 0, fmt.Errorf("%w: transaction movie update start: %w", customerrors.ErrDB, err)
	}

	defer tx.Rollback()

	// Check that the movie exists and get its current version.
	var currentVersion int

	err = tx.QueryRow(`SELECT version FROM movies WHERE id = ?`, id).Scan(&currentVersion)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, fmt.Errorf("%w: movie does not exist (movie=%d)", customerrors.ErrNotFound, id)
		}

		return 0, fmt.Errorf("%w: check movie version (movie=%d): %w", customerrors.ErrDB, id, err)
	}

	// Optimistic locking check.
	if currentVersion != *patch.Version {
		return 0, fmt.Errorf("%w: movie was modified by another user (movie=%d, expectedVersion=%d, currentVersion=%d)",
			customerrors.ErrConcurrentModification,
			id,
			*patch.Version,
			currentVersion,
		)
	}

	// Build scalar updates.
	updates := []string{}
	args := []any{}

	if patch.Title != nil {
		updates = append(updates, "title = ?")
		args = append(args, *patch.Title)
	}

	if patch.ReleaseYear != nil {
		updates = append(updates, "release_year = ?")
		args = append(args, *patch.ReleaseYear)
	}

	if patch.Duration != nil {
		updates = append(updates, "duration = ?")
		args = append(args, *patch.Duration)
	}

	// If anything is being updated, increment version.
	if len(updates) > 0 {
		updates = append(updates, "version = version + 1")

		args = append(args, id, *patch.Version)

		query := fmt.Sprintf(`UPDATE movies SET %s WHERE id = ? AND version = ?`, strings.Join(updates, ", "))

		result, err := tx.Exec(query, args...)
		if err != nil {
			return 0, fmt.Errorf("%w: update movie (movie=%d): %w", customerrors.ErrDB, id, err)
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return 0, fmt.Errorf("%w: get affected rows (movie=%d): %w", customerrors.ErrDB, id, err)
		}

		if rowsAffected == 0 {
			return 0, fmt.Errorf("%w: movie was modified by another user (movie=%d)", customerrors.ErrConflict, id)
		}
	}

	// Replace actors if supplied.
	if patch.Actors != nil {

		if _, err := tx.Exec(`DELETE FROM movie_actors WHERE movie_id = ?`, id); err != nil {
			return 0, fmt.Errorf("%w: delete movie actors (movie=%d): %w", customerrors.ErrDB, id, err)
		}

		query := `INSERT INTO movie_actors (movie_id, actor_id) VALUES (?, ?)`

		for _, actor := range *patch.Actors {
			if _, err := tx.Exec(query, id, actor.Id); err != nil {

				var sqliteErr sqlite3.Error

				if errors.As(err, &sqliteErr) &&
					sqliteErr.ExtendedCode == sqlite3.ErrConstraintForeignKey {

					return 0, fmt.Errorf("%w: non existing actor (actorId=%d)", customerrors.ErrInvalidReference, actor.Id)
				}

				return 0, fmt.Errorf("%w: insert movie actor (movie=%d, actor=%d): %w", customerrors.ErrDB, id, actor.Id, err)
			}
		}
	}

	// Replace genres if supplied.
	if patch.Genres != nil {

		if _, err := tx.Exec(`DELETE FROM movie_genres WHERE movie_id = ?`, id); err != nil {
			return 0, fmt.Errorf("%w: delete movie genres (movie=%d): %w", customerrors.ErrDB, id, err)
		}

		query := `INSERT INTO movie_genres (movie_id, genre_id) VALUES (?, ?)`

		for _, genre := range *patch.Genres {
			if _, err := tx.Exec(query, id, genre.Id); err != nil {

				var sqliteErr sqlite3.Error

				if errors.As(err, &sqliteErr) &&
					sqliteErr.ExtendedCode == sqlite3.ErrConstraintForeignKey {

					return 0, fmt.Errorf("%w: non existing genre (genreId=%d)", customerrors.ErrInvalidReference, genre.Id)
				}

				return 0, fmt.Errorf("%w: insert movie genre (movie=%d, genre=%d): %w", customerrors.ErrDB, id, genre.Id, err)
			}
		}
	}

	// If only actors or genres were updated, we still need
	// to increment the movie version.
	if len(updates) == 0 && (patch.Actors != nil || patch.Genres != nil) {

		result, err := tx.Exec(`UPDATE movies SET version = version + 1 WHERE id = ? AND version = ?`, id, *patch.Version)

		if err != nil {
			return 0, fmt.Errorf("%w: update movie version (movie=%d): %w", customerrors.ErrDB, id, err)
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return 0, fmt.Errorf("%w: get affected rows (movie=%d): %w", customerrors.ErrDB, id, err)
		}

		if rowsAffected == 0 {
			return 0, fmt.Errorf("%w: movie was modified by another user (movie=%d)", customerrors.ErrConflict, id)
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("%w: transaction movie update commit: %w", customerrors.ErrDB, err)
	}

	return 1, nil
}

func (r *SQLiteMovieRepository) Delete(id int, force bool) (int64, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return 0, fmt.Errorf("%w: transaction movie delete start: %w", customerrors.ErrDB, err)
	}

	defer tx.Rollback()

	///exists

	var movieTitle string

	err = tx.QueryRow(`SELECT title FROM movies WHERE id = ?`, id).Scan(&movieTitle)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, fmt.Errorf("%w: movie does not exist (movie=%d)", customerrors.ErrNotFound, id)
		}

		return 0, fmt.Errorf("%w: check movie before delete (movie=%d): %w", customerrors.ErrDB, id, err)
	}

	//check conflict

	var actorCount int
	var genreCount int

	err = tx.QueryRow(`SELECT COUNT(*) FROM movie_actors WHERE movie_id = ?`, id).Scan(&actorCount)

	if err != nil {
		return 0, fmt.Errorf("%w: count movie actors (movie=%d): %w", customerrors.ErrDB, id, err)
	}

	err = tx.QueryRow(`SELECT COUNT(*) FROM movie_genres WHERE movie_id = ?`, id).Scan(&genreCount)

	if err != nil {
		return 0, fmt.Errorf("%w: count movie genres (movie=%d): %w", customerrors.ErrDB, id, err)
	}

	//delete

	if !force && (actorCount > 0 || genreCount > 0) {
		return 0, fmt.Errorf("%w: cannot delete movie '%s' because it has %d actors and %d genres",
			customerrors.ErrConflict,
			movieTitle,
			actorCount,
			genreCount,
		)
	}

	//delete force

	if force {
		if _, err := tx.Exec(`DELETE FROM movie_actors WHERE movie_id = ?`, id); err != nil {
			return 0, fmt.Errorf("%w: delete movie actors (movie=%d): %w", customerrors.ErrDB, id, err)
		}

		if _, err := tx.Exec(`DELETE FROM movie_genres WHERE movie_id = ?`, id); err != nil {
			return 0, fmt.Errorf("%w: delete movie genres (movie=%d): %w", customerrors.ErrDB, id, err)
		}
	}

	//delete movie

	if _, err := tx.Exec(`DELETE FROM movies WHERE id = ?`, id); err != nil {
		return 0, fmt.Errorf("%w: delete movie (movie=%d): %w", customerrors.ErrDB, id, err)
	}

	//commit

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("%w: transaction movie delete commit: %w", customerrors.ErrDB, err)
	}

	return 1, nil
}

// extra
func (r *SQLiteMovieRepository) FindByExactTitle(title string) (*entity.Movie, error) {
	queryMoviesTable := `SELECT * FROM movies WHERE Title = ?`

	row := r.db.QueryRow(queryMoviesTable, title)
	movie := &entity.Movie{}

	err := row.Scan(&movie.Id, &movie.Title, &movie.ReleaseYear, &movie.Duration, &movie.Version)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("%w: select movie by id: %w", customerrors.ErrDB, err)
	}

	if err := r.populateRelations([]*entity.Movie{movie}); err != nil {
		return nil, err
	}

	return movie, nil

}

func (r *SQLiteMovieRepository) FindByTitleContains(title string) ([]*entity.Movie, error) {
	queryMoviesTable := `SELECT * FROM movies WHERE LOWER(title) LIKE ? ORDER BY Id`

	rows, err := r.db.Query(queryMoviesTable, "%"+title+"%")
	if err != nil {
		return nil, fmt.Errorf("%w: select movies: %w", customerrors.ErrDB, err)
	}
	defer rows.Close()

	movies := []*entity.Movie{}

	for rows.Next() {
		m := &entity.Movie{}
		err := rows.Scan(&m.Id, &m.Title, &m.ReleaseYear, &m.Duration, &m.Version)
		if err != nil {
			return nil, fmt.Errorf("%w: scan movie row: %w", customerrors.ErrDB, err)
		}
		movies = append(movies, m)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%w: iterate movie rows: %w", customerrors.ErrDB, err)
	}

	if err := r.populateRelations(movies); err != nil {
		return nil, err
	}

	return movies, nil

}

// Helpers
func (r *SQLiteMovieRepository) GetActorsByMovieIds(movieIds []uint) (map[uint][]entity.Actor, error) {

	actorsByMovie := make(map[uint][]entity.Actor)

	if len(movieIds) == 0 {
		return actorsByMovie, nil
	}

	placeholders := make([]string, len(movieIds))
	args := make([]any, len(movieIds))

	for i, id := range movieIds {
		placeholders[i] = "?"
		args[i] = id
	}

	queryMoviesAndActors := fmt.Sprintf(`
        SELECT 
            ma.movie_id, 
            a.id, 
            a.name, 
            a.birthdate,
            a.version
        FROM movie_actors ma
        JOIN actors a ON a.id = ma.actor_id
        WHERE ma.movie_id IN (%s)
        ORDER BY ma.movie_id;
    `, strings.Join(placeholders, ","))

	rows, err := r.db.Query(queryMoviesAndActors, args...)
	if err != nil {
		return nil, fmt.Errorf("%w: select actors by movie ids=%v: %w", customerrors.ErrDB, movieIds, err)
	}
	defer rows.Close()

	for rows.Next() {
		var (
			movieId   uint
			actor     entity.Actor
			birthdate string
		)

		if err := rows.Scan(&movieId, &actor.Id, &actor.Name, &birthdate, &actor.Version); err != nil {
			return nil, fmt.Errorf("%w: scan actor for movie=%d: %w", customerrors.ErrDB, movieId, err)
		}

		actor.BirthDate, err = time.Parse("2006-01-02", birthdate)

		if err != nil {
			return nil, fmt.Errorf("%w: invalid actor birthdate (actor=%d, value=%s): %w",
				customerrors.ErrInvalidReference,
				actor.Id,
				birthdate,
				err,
			)
		}

		actorsByMovie[movieId] = append(actorsByMovie[movieId], actor)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(
			"%w: iterate actors by movie ids=%v: %w",
			customerrors.ErrDB,
			movieIds,
			err,
		)
	}

	return actorsByMovie, nil

}

func (r *SQLiteMovieRepository) GetGenresByMovieIds(movieIds []uint) (map[uint][]entity.Genre, error) {

	genresByMovie := make(map[uint][]entity.Genre)

	if len(movieIds) == 0 {
		return genresByMovie, nil
	}

	placeholders := make([]string, len(movieIds))
	args := make([]any, len(movieIds))

	for i, id := range movieIds {
		placeholders[i] = "?"
		args[i] = id
	}

	queryMovieGenres := fmt.Sprintf(`
        SELECT 
            mg.movie_id,
            g.id,
            g.name,
            g.version
        FROM movie_genres mg
        JOIN genres g ON g.id = mg.genre_id
        WHERE mg.movie_id IN (%s)
        ORDER BY mg.movie_id
    `, strings.Join(placeholders, ","))

	rows, err := r.db.Query(queryMovieGenres, args...)
	if err != nil {
		return nil, fmt.Errorf("%w: select genres by movie ids=%v: %w", customerrors.ErrDB, movieIds, err)
	}
	defer rows.Close()

	for rows.Next() {
		var (
			movieId uint
			genre   entity.Genre
		)

		if err := rows.Scan(&movieId, &genre.Id, &genre.Name, &genre.Version); err != nil {
			return nil, fmt.Errorf("%w: scan genre for movie=%d: %w", customerrors.ErrDB, movieId, err)
		}

		genresByMovie[movieId] = append(genresByMovie[movieId], genre)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%w: iterate genres by movie ids=%v: %w", customerrors.ErrDB, movieIds, err)
	}

	return genresByMovie, nil
}

func (r *SQLiteMovieRepository) populateRelations(movies []*entity.Movie) error {
	if len(movies) == 0 {
		return nil
	}

	movieIds := []uint{}
	for _, movie := range movies {
		movieIds = append(movieIds, movie.Id)
	}

	actorsByMovieIds, err := r.GetActorsByMovieIds(movieIds)
	if err != nil {
		return err
	}

	genresByMovieIds, err := r.GetGenresByMovieIds(movieIds)
	if err != nil {
		return err
	}

	for _, movie := range movies {
		movie.Actors = actorsByMovieIds[movie.Id]
		movie.Genres = genresByMovieIds[movie.Id]
	}

	return nil
}

func isUniqueConstraint(err error) bool {
	var sqliteErr sqlite3.Error

	if errors.As(err, &sqliteErr) {
		return sqliteErr.Code == sqlite3.ErrConstraint &&
			sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique
	}

	return false
}

func containsActor(actors []entity.Actor, actorId int) bool {
	for _, actor := range actors {
		if actor.Id == uint(actorId) {
			return true
		}
	}

	return false
}

func containsGenre(genres []entity.Genre, genreId int) bool {
	for _, genre := range genres {
		if genre.Id == uint(genreId) {
			return true
		}
	}

	return false
}
