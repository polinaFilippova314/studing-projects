package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"movies-api/customerrors"
	"movies-api/entity"
	"movies-api/service"
)

type GenreHandler struct {
	service *service.GenreService
}

func NewGenreHandler(service *service.GenreService) *GenreHandler {
	return &GenreHandler{service: service}
}

func (h *GenreHandler) Create(w http.ResponseWriter, r *http.Request) *customerrors.HttpError {
	var genre entity.Genre
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&genre)
	if err != nil {
		return &customerrors.HttpError{Err: err, Message: "invalid json", Code: http.StatusBadRequest}
	}
	genre.Version = 1
	id, err := h.service.CreateGenre(&genre)
	genre.Id = uint(id)
	if err != nil {
		if errors.Is(err, entity.ErrInvalidContent) {
			return &customerrors.HttpError{Err: err, Message: err.Error(), Code: http.StatusBadRequest}
		}
		if errors.Is(err, entity.ErrAlreadyExists) {
			return &customerrors.HttpError{Err: err, Message: err.Error(), Code: http.StatusConflict}
		}
		return &customerrors.HttpError{Err: err, Message: err.Error(), Code: http.StatusInternalServerError}
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(genre)
	return nil
}
func (h *GenreHandler) GetAll(w http.ResponseWriter, r *http.Request) *customerrors.HttpError {
	gotMovies := r.URL.Query().Get("movies")
	movies := gotMovies == "true"
	genres, err := h.service.GetAll(movies)
	if err != nil {
		if errors.Is(err, entity.ErrNotFound) {
			return &customerrors.HttpError{Err: err, Message: err.Error(), Code: http.StatusNotFound}
		}
		return &customerrors.HttpError{Err: err, Message: err.Error(), Code: http.StatusInternalServerError}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(genres)
	return nil

}
func (h *GenreHandler) GetByID(w http.ResponseWriter, r *http.Request) *customerrors.HttpError {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		return &customerrors.HttpError{Err: err, Message: "invalid id", Code: http.StatusBadRequest}
	}
	genre, err := h.service.GetByID(id)
	if err != nil {
		if errors.Is(err, entity.ErrNotFound) {
			return &customerrors.HttpError{Err: err, Message: err.Error(), Code: http.StatusNotFound}
		}
		return &customerrors.HttpError{Err: err, Message: err.Error(), Code: http.StatusInternalServerError}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(genre)
	return nil
}
func (h *GenreHandler) Update(w http.ResponseWriter, r *http.Request) *customerrors.HttpError {
	idGenre := r.PathValue("id")
	id, err := strconv.Atoi(idGenre)
	if err != nil || id <= 0 {
		return &customerrors.HttpError{Err: err, Message: "invalid id", Code: http.StatusBadRequest}
	}
	var genreUpdate entity.GenrePatchRequest
	err1 := json.NewDecoder(r.Body).Decode(&genreUpdate)
	if err1 != nil {
		return &customerrors.HttpError{Err: err, Message: "invalid json", Code: http.StatusBadRequest}
	}
	actor, err := h.service.Update(id, genreUpdate)
	if err != nil {
		if errors.Is(err, entity.ErrNotFound) {
			return &customerrors.HttpError{Err: err, Message: err.Error(), Code: http.StatusNotFound}
		}
		if errors.Is(err, entity.ErrVersionConflict) {
			return &customerrors.HttpError{Err: err, Message: err.Error(), Code: http.StatusConflict}
		}
		return &customerrors.HttpError{Err: err, Message: err.Error(), Code: http.StatusInternalServerError}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(actor)
	return nil
}
func (h *GenreHandler) Delete(w http.ResponseWriter, r *http.Request) *customerrors.HttpError {
	idGenre := r.PathValue("id")
	id, err := strconv.Atoi(idGenre)
	if err != nil || id <= 0 {
		return &customerrors.HttpError{Err: err, Message: "invalid id", Code: http.StatusBadRequest}
	}
	gotForce := r.URL.Query().Get("force")
	force := gotForce == "true"
	err = h.service.Delete(id, force)
	if err != nil {
		if errors.Is(err, entity.ErrNotFound) {
			return &customerrors.HttpError{Err: err, Message: err.Error(), Code: http.StatusNotFound}
		}
		if errors.Is(err, entity.ErrHasRelations) {
			return &customerrors.HttpError{Err: err, Message: err.Error(), Code: http.StatusConflict}
		}
		return &customerrors.HttpError{Err: err, Message: err.Error(), Code: http.StatusInternalServerError}
	}
	w.WriteHeader(http.StatusNoContent)
	return nil
}
func (h *GenreHandler) DeleteConnection(w http.ResponseWriter, r *http.Request) *customerrors.HttpError {
	idGenre := r.PathValue("id")
	id, err := strconv.Atoi(idGenre)
	if err != nil || id <= 0 {
		return &customerrors.HttpError{Err: err, Message: "invalid id", Code: http.StatusBadRequest}
	}
	var moviesId entity.DeleteMoviesConnectionRequest
	err = json.NewDecoder(r.Body).Decode(&moviesId)
	if err != nil {
		return &customerrors.HttpError{Err: err, Message: "invalid json", Code: http.StatusBadRequest}
	}
	err = h.service.DeleteConnection(id, moviesId.MovieIds)
	if err != nil {
		if errors.Is(err, entity.ErrNotFound) {
			return &customerrors.HttpError{Err: err, Message: err.Error(), Code: http.StatusNotFound}
		}
		return &customerrors.HttpError{Err: err, Message: err.Error(), Code: http.StatusInternalServerError}
	}
	w.WriteHeader(http.StatusNoContent)
	return nil
}
