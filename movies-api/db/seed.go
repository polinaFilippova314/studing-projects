package db

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

type ActorSeed struct {
	Name      string
	BirthDate string
}

type GenreSeed struct {
	Name    string
}

type MovieSeed struct {
	Title       string
	ReleaseYear int
	Duration    float64

	Actors []string
	Genres []string
}

func SeedTables(db *sql.DB) error {

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	//genres

	genres := []GenreSeed{
		{"Action"},
		{"Adventure"},
		{"Comedy"},
		{"Drama"},
		{"Science Fiction"},
	}

	for _, genre := range genres {
		_, err := tx.Exec(`
			INSERT OR IGNORE INTO genres (name)
			VALUES (?)
		`, genre.Name)

		if err != nil {
			return err
		}
	}

	//actors

	actors := []ActorSeed{
		{"Leonardo DiCaprio", "1974-11-11"},
		{"Tom Hanks", "1956-07-09"},
		{"Scarlett Johansson", "1984-11-22"},
		{"Morgan Freeman", "1937-06-01"},
		{"Natalie Portman", "1981-06-09"},
		{"Christian Bale", "1974-01-30"},
		{"Emma Stone", "1988-11-06"},
		{"Robert Downey Jr.", "1965-04-04"},
		{"Chris Evans", "1981-06-13"},
		{"Jennifer Lawrence", "1990-08-15"},
		{"Brad Pitt", "1963-12-18"},
		{"Angelina Jolie", "1975-06-04"},
		{"Denzel Washington", "1954-12-28"},
		{"Anne Hathaway", "1982-11-12"},
		{"Ryan Gosling", "1980-11-12"},
		{"Matthew McConaughey", "1969-11-04"},
		{"Russell Crowe", "1964-04-07"},
		{"Matt Damon", "1970-10-08"},
		{"Keanu Reeves", "1964-09-02"},
		{"Robert Pattinson", "1986-05-13"},
	}

	for _, actor := range actors {
		_, err := tx.Exec(`
			INSERT OR IGNORE INTO actors (name, birthdate)
			VALUES (?, ?)
		`, actor.Name, actor.BirthDate)

		if err != nil {
			return err
		}
	}

	//movies

	movies := []MovieSeed{
		{
			"Inception",
			2010,
			148,
			[]string{"Leonardo DiCaprio"},
			[]string{"Science Fiction", "Action"},
		},
		{
			"The Dark Knight",
			2008,
			152,
			[]string{"Christian Bale", "Morgan Freeman"},
			[]string{"Action", "Drama"},
		},
		{
			"Interstellar",
			2014,
			169,
			[]string{"Matthew McConaughey"},
			[]string{"Science Fiction", "Adventure"},
		},
		{
			"Forrest Gump",
			1994,
			142,
			[]string{"Tom Hanks"},
			[]string{"Drama", "Comedy"},
		},
		{
			"The Avengers",
			2012,
			143,
			[]string{"Robert Downey Jr.", "Chris Evans", "Scarlett Johansson"},
			[]string{"Action", "Adventure"},
		},
		{
			"Iron Man",
			2008,
			126,
			[]string{"Robert Downey Jr."},
			[]string{"Action", "Science Fiction"},
		},
		{
			"Black Widow",
			2021,
			134,
			[]string{"Scarlett Johansson"},
			[]string{"Action", "Adventure"},
		},
		{
			"Gladiator",
			2000,
			155,
			[]string{"Russell Crowe"},
			[]string{"Action", "Drama"},
		},
		{
			"Fight Club",
			1999,
			139,
			[]string{"Brad Pitt"},
			[]string{"Drama"},
		},
		{
			"Mr. & Mrs. Smith",
			2005,
			120,
			[]string{"Brad Pitt", "Angelina Jolie"},
			[]string{"Action", "Comedy"},
		},
		{
			"La La Land",
			2016,
			128,
			[]string{"Ryan Gosling", "Emma Stone"},
			[]string{"Drama", "Comedy"},
		},
		{
			"Passengers",
			2016,
			116,
			[]string{"Jennifer Lawrence"},
			[]string{"Science Fiction", "Drama"},
		},
		{
			"The Martian",
			2015,
			144,
			[]string{"Matt Damon"},
			[]string{"Science Fiction", "Adventure"},
		},
		{
			"Black Swan",
			2010,
			108,
			[]string{"Natalie Portman"},
			[]string{"Drama"},
		},
		{
			"Lucy",
			2014,
			90,
			[]string{"Scarlett Johansson"},
			[]string{"Action", "Science Fiction"},
		},
		{
			"John Wick",
			2014,
			101,
			[]string{"Keanu Reeves"},
			[]string{"Action"},
		},
		{
			"The Prestige",
			2006,
			130,
			[]string{"Christian Bale"},
			[]string{"Drama", "Science Fiction"},
		},
		{
			"Ocean's Eleven",
			2001,
			116,
			[]string{"Brad Pitt"},
			[]string{"Comedy", "Action"},
		},
		{
			"The Hunger Games",
			2012,
			142,
			[]string{"Jennifer Lawrence"},
			[]string{"Adventure", "Science Fiction"},
		},
		{
			"Tenet",
			2020,
			150,
			[]string{"Robert Pattinson"},
			[]string{"Science Fiction", "Action"},
		},
	}

	for _, movie := range movies {

		_, err := tx.Exec(`
			INSERT OR IGNORE INTO movies(title, release_year, duration)
			VALUES (?, ?, ?)
		`,
			movie.Title,
			movie.ReleaseYear,
			movie.Duration,
		)

		if err != nil {
			return err
		}

		var movieID int

		err = tx.QueryRow(`
			SELECT id
			FROM movies
			WHERE title = ? AND release_year = ?
		`, movie.Title, movie.ReleaseYear).Scan(&movieID)

		if err != nil {
			return err
		}

		for _, actorName := range movie.Actors {
			var actorID int

			err := tx.QueryRow(`
				SELECT id 
				FROM actors 
				WHERE name = ?
			`, actorName).Scan(&actorID)

			if err != nil {
				return err
			}

			_, err = tx.Exec(`
				INSERT OR IGNORE INTO movie_actors(movie_id, actor_id)
				VALUES (?, ?)
			`, movieID, actorID)

			if err != nil {
				return err
			}
		}

		for _, genreName := range movie.Genres {
			var genreID int

			err := tx.QueryRow(`
				SELECT id FROM genres WHERE name = ?
			`, genreName).Scan(&genreID)

			if err != nil {
				return err
			}

			_, err = tx.Exec(`
				INSERT OR IGNORE INTO movie_genres(movie_id, genre_id)
				VALUES (?, ?)
			`, movieID, genreID)

			if err != nil {
				return err
			}
		}
	}

	return tx.Commit()

}
