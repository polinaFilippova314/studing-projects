package repository

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
	"unicode"

	"movies-api/entity"
)

type SQLiteActorRepository struct {
	db *sql.DB
}

func NewSQLiteActorRepository(db *sql.DB) *SQLiteActorRepository {
	return &SQLiteActorRepository{db: db}
}

type ActorRepository interface {
	Create(actor *entity.Actor) (int64, error)
	GetAll(moviesFlag bool, page int, size int, pagination bool) (entity.PaginatedActorResponse, error)
	GetByID(id int) (entity.Actor, error)
	GetByName(name string) ([]entity.Actor, error)
	Update(id int, actor entity.ActorPatchRequest) (entity.Actor, error)
	Delete(id int, force bool) (int64, error)
	DeleteConnection(id int, movies []int) (int64, error)
	CheckDuplicates() ([]entity.Actor, error)
}

func (a *SQLiteActorRepository) Create(actor *entity.Actor) (int64, error) {
	tx, err := a.db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	query := `INSERT INTO actors (name, birthdate, version) VALUES (?, ?, ?);`
	actor.Name = nameStyle(actor.Name)
	result, err := tx.Exec(query, actor.Name, actor.BirthDate.Format("2006-01-02"), 1)
	if err != nil {
		return 0, fmt.Errorf("%w: actor name or actor birthday already exists", entity.ErrAlreadyExists)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	if actor.MovieIds != nil {
		if err := CreateActorConnection(tx, id, actor.MovieIds); err != nil {
			return 0, err
		}
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return id, nil
}
func (a *SQLiteActorRepository) GetAll(moviesFlag bool, page int, size int, pagination bool) (entity.PaginatedActorResponse, error) {
	offset := page * size
	query := `SELECT id, name, birthdate, version FROM actors`
	queryPagination := `SELECT id, name, birthdate, version FROM actors ORDER BY id LIMIT ? OFFSET ?`
	queryCount := `SELECT COUNT(*) FROM actors`
	var err error
	var rows *sql.Rows
	if pagination {
		rows, err = a.db.Query(queryPagination, size, offset)
	} else {
		rows, err = a.db.Query(query)
	}
	if err != nil {
		return entity.PaginatedActorResponse{}, err
	}
	defer rows.Close()
	actors := []entity.Actor{}
	for rows.Next() {
		var id uint
		var name, birthdate string
		var version int
		if err := rows.Scan(&id, &name, &birthdate, &version); err != nil {
			return entity.PaginatedActorResponse{}, err
		}
		birthTime, err := time.Parse("2006-01-02", birthdate)
		if err != nil {
			return entity.PaginatedActorResponse{}, err
		}
		if moviesFlag {
			movies, err := a.GetMovies(int(id))
			if err != nil {
				return entity.PaginatedActorResponse{}, err
			}
			actors = append(actors, entity.Actor{Id: id, Name: name, BirthDate: birthTime, Version: version, Movies: movies})
		} else {
			actors = append(actors, entity.Actor{Id: id, Name: name, BirthDate: birthTime, Version: version})
		}
	}
	countActors := 0
	if err = a.db.QueryRow(queryCount).Scan(&countActors); err != nil {
		return entity.PaginatedActorResponse{}, err
	}
	var result entity.PaginatedActorResponse
	result.Actors = actors
	result.Page = uint(page)
	result.Size = uint(size)
	if !pagination {
		result.Size = uint(len(actors))
	}
	result.Total = uint(countActors)
	return result, nil
}
func (a *SQLiteActorRepository) GetByID(id int) (entity.Actor, error) {
	query := `SELECT name, birthdate, version FROM actors WHERE id = ?`
	row := a.db.QueryRow(query, id)
	var name, birthdate string
	var version int
	err := row.Scan(&name, &birthdate, &version)
	if err == sql.ErrNoRows {
		return entity.Actor{}, fmt.Errorf("%w actor id: %d", entity.ErrNotFound, id)
	} else if err != nil {
		return entity.Actor{}, err
	}
	birthTime, err := time.Parse("2006-01-02", birthdate)
	if err != nil {
		return entity.Actor{}, err
	}
	movies, err := a.GetMovies(id)
	if err != nil {
		return entity.Actor{}, err
	}
	return entity.Actor{Id: uint(id), Name: name, BirthDate: birthTime, Version: version, Movies: movies}, nil
}
func (a *SQLiteActorRepository) GetByName(name string) ([]entity.Actor, error) {
	query := `SELECT id, name, birthdate, version FROM actors WHERE name LIKE ?`
	searchPattern := "%" + name + "%"
	rows, err := a.db.Query(query, searchPattern)
	if err != nil {
		return []entity.Actor{}, err
	}
	defer rows.Close()
	actors := []entity.Actor{}
	for rows.Next() {
		var id uint
		var version int
		var nameActual, birthdate string
		if err := rows.Scan(&id, &nameActual, &birthdate, &version); err != nil {
			return []entity.Actor{}, err
		}
		birthTime, err := time.Parse("2006-01-02", birthdate)
		if err != nil {
			return []entity.Actor{}, err
		}
		movies, err := a.GetMovies(int(id))
		if err != nil {
			return []entity.Actor{}, err
		}
		actors = append(actors, entity.Actor{Id: uint(id), Name: nameActual, BirthDate: birthTime, Version: version, Movies: movies})
	}
	return actors, nil
}
func (a *SQLiteActorRepository) Update(id int, actor entity.ActorPatchRequest) (entity.Actor, error) {
	tx, err := a.db.Begin()
	if err != nil {
		return entity.Actor{}, err
	}
	defer tx.Rollback()
	query := `SELECT name, birthdate, version FROM actors WHERE id = ?`
	queryConnection := `SELECT movie_id FROM movie_actors WHERE actor_id = ?`
	row := tx.QueryRow(query, id)
	var name, birthdate string
	var version int
	err = row.Scan(&name, &birthdate, &version)
	if err == sql.ErrNoRows {
		return entity.Actor{}, fmt.Errorf("%w actor id: %d", entity.ErrNotFound, id)
	} else if err != nil {
		return entity.Actor{}, err
	}

	if actor.BirthDate != nil {
		birthdate = *actor.BirthDate
	}
	if actor.Name != nil {
		name = *actor.Name
	}
	birthTime, err := time.Parse("2006-01-02", birthdate)
	if err != nil {
		return entity.Actor{}, err
	}
	newQuery := `UPDATE actors SET name = ?, birthdate = ?, version = version + 1 WHERE id = ? AND version = ?`
	result, err := tx.Exec(newQuery, name, birthTime.Format("2006-01-02"), id, actor.Version)
	if err != nil {
		return entity.Actor{}, err
	}
	//if someone delete this entity before this func updates everything
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return entity.Actor{}, fmt.Errorf("%w: actor was updated by someone else, refetch and try again", entity.ErrVersionConflict)
	}
	if actor.MovieIds != nil {
		rowConnection, err := tx.Query(queryConnection, id)
		if err != nil {
			return entity.Actor{}, err
		}
		defer rowConnection.Close()
		var currentMoviesId []int
		for rowConnection.Next() {
			var idCurrent int
			err = rowConnection.Scan(&idCurrent)
			if err != nil {
				return entity.Actor{}, err
			}
			currentMoviesId = append(currentMoviesId, idCurrent)
		}
		toAdd, toDelete := computeMovieDiff(currentMoviesId, actor.MovieIds)
		err = CreateActorConnection(tx, int64(id), toAdd)
		if err != nil {
			return entity.Actor{}, err
		}
		err = DeleteActorConnection(tx, int64(id), toDelete)
		if err != nil {
			return entity.Actor{}, err
		}
	}
	if err := tx.Commit(); err != nil {
		return entity.Actor{}, err
	}
	return entity.Actor{Id: uint(id), Name: name, BirthDate: birthTime, Version: version + 1}, nil
}
func (a *SQLiteActorRepository) Delete(id int, force bool) (int64, error) {
	query := `SELECT name, birthdate, version FROM actors WHERE id = ?`
	row := a.db.QueryRow(query, id)
	var name, birthdate string
	var version int
	err := row.Scan(&name, &birthdate, &version)
	if err == sql.ErrNoRows {
		return 0, fmt.Errorf("%w actor id: %d", entity.ErrNotFound, id)
	} else if err != nil {
		return 0, err
	}
	countFilms := 0
	queryCount := `SELECT COUNT(*) FROM movie_actors WHERE actor_id = ? `
	if err = a.db.QueryRow(queryCount, id).Scan(&countFilms); err != nil {
		return 0, err
	}
	if countFilms > 0 && !force {
		return 0, fmt.Errorf("%w: %s plays in %d films", entity.ErrHasRelations, name, countFilms)
	}
	if force {
		queryDeleteConnection := `DELETE FROM movie_actors WHERE actor_id=?`
		_, err = a.db.Exec(queryDeleteConnection, id)
		if err != nil {
			return 0, err
		}
	}
	queryDelete := `DELETE FROM actors WHERE id = ?`
	result, err := a.db.Exec(queryDelete, id)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}
func (a *SQLiteActorRepository) DeleteConnection(id int, movies []int) (int64, error) {
	str := make([]string, len(movies))
	args := []any{id}
	for i, movieID := range movies {
		str[i] = "?"
		args = append(args, movieID)
	}
	placeholder := strings.Join(str, ",")
	queryDeleteConnection := fmt.Sprintf(`DELETE FROM movie_actors WHERE actor_id=? AND movie_id IN (%s)`, placeholder)
	result, err := a.db.Exec(queryDeleteConnection, args...)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}
func (a *SQLiteActorRepository) CheckDuplicates() ([]entity.Actor, error) {
	query := `SELECT id, name, version FROM actors`
	rows, err := a.db.Query(query)
	if err != nil {
		return []entity.Actor{}, err
	}
	defer rows.Close()
	actors := []entity.Actor{}
	for rows.Next() {
		var id uint
		var name, birthdate string
		var version int
		if err := rows.Scan(&id, &name, &birthdate, &version); err != nil {
			return []entity.Actor{}, err
		}
		birthTime, err := time.Parse("2006-01-02", birthdate)
		if err != nil {
			return []entity.Actor{}, err
		}
		actors = append(actors, entity.Actor{Id: id, Name: name, BirthDate: birthTime, Version: version})
	}
	actorsDuplicated, err := checkDuplicates(actors)
	if err != nil {
		return []entity.Actor{}, err
	}
	return actorsDuplicated, nil
}

// helpers
func computeMovieDiff(current []int, desired []int) ([]int, []int) {
	currentSet := make(map[int]bool)
	for _, v := range current {
		currentSet[v] = true
	}
	newSet := make(map[int]bool)
	for _, v := range desired {
		newSet[v] = true
	}
	toAdd := []int{}
	toDelete := []int{}
	for _, v := range desired {
		if !currentSet[v] {
			toAdd = append(toAdd, v)
		}
	}
	for _, v := range current {
		if !newSet[v] {
			toDelete = append(toDelete, v)
		}
	}
	return toAdd, toDelete
}
func nameStyle(name string) string {
	nameSlice := strings.Split(name, " ")
	for i := range nameSlice {
		nameSlice[i] = strings.ToLower(nameSlice[i])
		nameSlice[i] = firstUpperCase(nameSlice[i])
	}
	return strings.Join(nameSlice, " ")
}
func firstUpperCase(name string) string {
	s := []rune(name)
	s[0] = unicode.ToUpper(s[0])
	return string(s)
}
func CreateActorConnection(tx *sql.Tx, idActor int64, idMovies []int) error {
	for _, id := range idMovies {
		_, err := tx.Exec(`INSERT OR IGNORE INTO movie_actors(movie_id, actor_id) VALUES (?,?)`,
			id, idActor)
		if err != nil {
			return err
		}
	}
	return nil
}
func DeleteActorConnection(tx *sql.Tx, idActor int64, idMovies []int) error {
	if len(idMovies) == 0 {
		return nil
	}
	str := make([]string, len(idMovies))
	args := []any{idActor}
	for i, movieID := range idMovies {
		str[i] = "?"
		args = append(args, movieID)
	}
	placeholder := strings.Join(str, ",")
	query := fmt.Sprintf(`DELETE FROM movie_actors WHERE actor_id=? AND movie_id IN (%s)`, placeholder)
	_, err := tx.Exec(query, args...)
	return err
}
func (a *SQLiteActorRepository) GetMovies(id int) ([]entity.Movie, error) {
	query := `SELECT movies.id, movies.title, movies.release_year, movies.duration
	FROM movies
	JOIN movie_actors ON movies.id=movie_actors.movie_id
	WHERE movie_actors.actor_id = ?`
	rows, err := a.db.Query(query, id)
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
func checkDuplicates(actors []entity.Actor) ([]entity.Actor, error) {
	result := []entity.Actor{}
	for i := range actors {
		actors[i].Name = strings.ToLower(actors[i].Name)
	}
	for i := 0; i < len(actors)-1; i++ {
		for j := i + 1; j < len(actors); j++ {

		}
	}
	return result, nil
}
