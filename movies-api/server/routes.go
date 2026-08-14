package server

import (
	"database/sql"
	"net/http"

	"movies-api/customerrors"
	"movies-api/handler"
	"movies-api/repository"
	"movies-api/service"
)

func RegisterRoutes(mux *http.ServeMux, database *sql.DB) {
	movieRepo := repository.NewSQLiteMovieRepository(database)
	movieService := service.NewMovieService(movieRepo)
	movieHandler := handler.NewMovieHandler(movieService)

	//actors
	actorRepo := repository.NewSQLiteActorRepository(database)
	actorService := service.NewActorService(actorRepo)
	actorHandler := handler.NewActorHandler(actorService)
	//genre
	genreRepo := repository.NewSQLiteGenreRepository(database)
	genreService := service.NewGenreService(genreRepo)
	genreHandler := handler.NewGenreHandler(genreService)

	mux.Handle("GET /api/movies", customerrors.HttpErrorHandler(movieHandler.Get))
	mux.Handle("GET /api/movies/search", customerrors.HttpErrorHandler(movieHandler.Search))
	mux.Handle("GET /api/movies/{id}", customerrors.HttpErrorHandler(movieHandler.GetById))
	mux.Handle("GET /api/movies/{id}/actors", customerrors.HttpErrorHandler(movieHandler.GetActorsById))
	mux.Handle("POST /api/movies", customerrors.HttpErrorHandler(movieHandler.Create))
	mux.Handle("PATCH /api/movies/{id}", customerrors.HttpErrorHandler(movieHandler.Update))
	mux.Handle("DELETE /api/movies/{id}", customerrors.HttpErrorHandler(movieHandler.Delete))

	mux.Handle("GET /api/actors", customerrors.HttpErrorHandler(actorHandler.GetAll))
	mux.Handle("POST /api/actors", customerrors.HttpErrorHandler(actorHandler.Create))
	mux.Handle("GET /api/actors/{id}", customerrors.HttpErrorHandler(actorHandler.GetByID))
	mux.Handle("PATCH /api/actors/{id}", customerrors.HttpErrorHandler(actorHandler.Update))
	mux.Handle("DELETE /api/actors/{id}", customerrors.HttpErrorHandler(actorHandler.Delete))
	mux.Handle("DELETE /api/actors/deleteconnection/{id}", customerrors.HttpErrorHandler(actorHandler.DeleteConnection))
	mux.Handle("DELETE /api/actors/checkduplicate", customerrors.HttpErrorHandler(actorHandler.CheckDuplicates))

	mux.Handle("GET /api/genres", customerrors.HttpErrorHandler(genreHandler.GetAll))
	mux.Handle("POST /api/genres", customerrors.HttpErrorHandler(genreHandler.Create))
	mux.Handle("GET /api/genres/{id}", customerrors.HttpErrorHandler(genreHandler.GetByID))
	mux.Handle("PATCH /api/genres/{id}", customerrors.HttpErrorHandler(genreHandler.Update))
	mux.Handle("DELETE /api/genres/{id}", customerrors.HttpErrorHandler(genreHandler.Delete))
	mux.Handle("DELETE /api/genres/deleteconnection/{id}", customerrors.HttpErrorHandler(genreHandler.DeleteConnection))
}
