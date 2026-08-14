package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"movies-api/customerrors"
	"movies-api/entity"
	"movies-api/service"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type MovieHandler struct {
	service *service.MovieService
}

func NewMovieHandler(s *service.MovieService) *MovieHandler {
	return &MovieHandler{
		service: s,
	}
}

func (h *MovieHandler) Get(w http.ResponseWriter, r *http.Request) *customerrors.HttpError {
	if r.Method != http.MethodGet {
		return &customerrors.HttpError{Message: "method not allowed", Code: http.StatusMethodNotAllowed}
	}

	movies, page, size, err := h.getMovies(r)
	if err != nil {
		return errorResponse(err)
	}

	var response []byte

	responseWith := struct {
		Page   int
		Size   int
		Movies []*entity.Movie
	}{
		Page:   page,
		Size:   size,
		Movies: movies,
	}

	if page == 0 && size == 0 {
		response, err = json.Marshal(movies)
	} else {
		response, err = json.Marshal(responseWith)
	}

	if err != nil {
		return &customerrors.HttpError{Message: "failed to encode JSON", Code: http.StatusInternalServerError}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(response)

	return nil

}

func (h *MovieHandler) getMovies(r *http.Request) ([]*entity.Movie, int, int, error) {
	params := r.URL.Query()

	switch {
	case params.Get("actor") != "":
		id, err := parseIdFromParam(params.Get("actor"), "actor id")
		if err != nil {
			return nil, 0, 0, err
		}

		movies, err := h.service.FindMoviesByActor(id)
		return movies, 0, 0, err

	case params.Get("genre") != "":
		id, err := parseIdFromParam(params.Get("genre"), "genre id")
		if err != nil {
			return nil, 0, 0, err
		}

		movies, err := h.service.FindMoviesByGenre(id)
		return movies, 0, 0, err

	case params.Get("year") != "":
		id, err := parseIdFromParam(params.Get("year"), "release year")
		if err != nil {
			return nil, 0, 0, err
		}

		movies, err := h.service.FindMoviesByYear(id)
		return movies, 0, 0, err

	default:
		pageStr := strings.TrimSpace(params.Get("page"))
		sizeStr := strings.TrimSpace(params.Get("size"))

		if pageStr == "" && sizeStr == "" {
			movies, err := h.service.GetAllMovies()
			return movies, 0, 0, err
		}

		page, err := parsePageAndSize(pageStr, "page")
		if err != nil {
			return nil, 0, 0, err
		}

		size, err := parsePageAndSize(sizeStr, "size")
		if err != nil {
			return nil, 0, 0, err
		}

		movies, err := h.service.GetMoviesWithPagination(page, size)
		return movies, page, size, err
	}
}

func (h *MovieHandler) GetById(w http.ResponseWriter, r *http.Request) *customerrors.HttpError {
	if r.Method != http.MethodGet {
		return &customerrors.HttpError{Message: "method not allowed", Code: http.StatusMethodNotAllowed}
	}

	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)

	if err != nil || id <= 0 {
		return &customerrors.HttpError{Message: "invalid movie id", Code: http.StatusBadRequest}
	}

	movie, err := h.service.GetMovieById(id)

	if err != nil {
		return errorResponse(err)
	}

	response, err := json.Marshal(movie)

	if err != nil {
		return &customerrors.HttpError{Message: "failed to encode JSON", Code: http.StatusInternalServerError}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(response)

	return nil

}

func (h *MovieHandler) GetActorsById(w http.ResponseWriter, r *http.Request) *customerrors.HttpError {
	if r.Method != http.MethodGet {
		return &customerrors.HttpError{Message: "method not allowed", Code: http.StatusMethodNotAllowed}
	}

	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)

	if err != nil || id <= 0 {
		return &customerrors.HttpError{Message: "invalid movie id", Code: http.StatusBadRequest}
	}

	actors, err := h.service.FindMovieActors(id)

	if err != nil {
		return errorResponse(err)
	}

	response, err := json.Marshal(actors)

	if err != nil {
		return &customerrors.HttpError{Message: "failed to encode JSON", Code: http.StatusInternalServerError}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(response)

	return nil

}

func (h *MovieHandler) Create(w http.ResponseWriter, r *http.Request) *customerrors.HttpError {
	if r.Method != http.MethodPost {
		return &customerrors.HttpError{Message: "method not allowed", Code: http.StatusMethodNotAllowed}
	}

	var movie entity.Movie

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	err := decoder.Decode(&movie)
	if err != nil {
		return &customerrors.HttpError{Message: "invalid json: " + err.Error(), Code: http.StatusBadRequest}
	}

	if hasDuplicateActorIds(movie.Actors) {
		return &customerrors.HttpError{Message: "duplicate actor id provided", Code: http.StatusBadRequest}
	}

	if hasDuplicateGenreIds(movie.Genres) {
		return &customerrors.HttpError{Message: "duplicate genre id provided", Code: http.StatusBadRequest}
	}

	if err := movie.ValidateMovie(); err != nil {
		return &customerrors.HttpError{Message: err.Error(), Code: http.StatusBadRequest}
	}

	createdId, err := h.service.CreateMovie(&movie)

	if err != nil {
		return errorResponse(err)
	}

	movie.Id = uint(createdId)
	movie.Version = 1

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(movie)

	return nil
}

func (h *MovieHandler) Update(w http.ResponseWriter, r *http.Request) *customerrors.HttpError {
	if r.Method != http.MethodPatch {
		return &customerrors.HttpError{Message: "method not allowed", Code: http.StatusMethodNotAllowed}
	}

	movieId, err := parseIdFromParam(r.PathValue("id"), "movie id")
	if err != nil {
		return errorResponse(err)
	}

	var patch entity.MoviePatch

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	err = decoder.Decode(&patch)

	if err != nil {
		return &customerrors.HttpError{Message: "invalid request body", Code: http.StatusBadRequest}
	}

	// Version is required for optimistic locking.
	if patch.Version == nil {
		return &customerrors.HttpError{Message: "movie version is required", Code: http.StatusBadRequest}
	}

	if *patch.Version <= 0 {
		return &customerrors.HttpError{Message: "movie version must be greater than zero", Code: http.StatusBadRequest}
	}

	// Validate only supplied scalar fields.
	if patch.Title != nil {
		if strings.TrimSpace(*patch.Title) == "" {
			return &customerrors.HttpError{Message: "movie title cannot be empty", Code: http.StatusBadRequest}
		}
	}

	currentYear := time.Now().Year()

	if patch.ReleaseYear != nil {
		if *patch.ReleaseYear < 1888 {
			return &customerrors.HttpError{Message: "movie release year cannot be before 1888", Code: http.StatusBadRequest}
		}

		if *patch.ReleaseYear > currentYear {
			return &customerrors.HttpError{
				Message: fmt.Sprintf("movie has not been released yet: %d", *patch.ReleaseYear),
				Code:    http.StatusBadRequest,
			}
		}
	}

	if patch.Duration != nil && *patch.Duration <= 0 {
		return &customerrors.HttpError{Message: "movie duration must be greater than zero", Code: http.StatusBadRequest}
	}

	_, err = h.service.UpdateMovie(movieId, &patch)

	if err != nil {
		return errorResponse(err)
	}

	movie, err := h.service.GetMovieById(movieId)

	if err != nil {
		return errorResponse(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(movie); err != nil {
		return &customerrors.HttpError{Message: "failed to encode JSON", Code: http.StatusInternalServerError}
	}

	return nil
}

func (h *MovieHandler) Delete(w http.ResponseWriter, r *http.Request) *customerrors.HttpError {
	if r.Method != http.MethodDelete {
		return &customerrors.HttpError{Message: "method not allowed", Code: http.StatusMethodNotAllowed}
	}

	movieId, err := parseIdFromParam(
		r.PathValue("id"),
		"movie id",
	)

	if err != nil {
		return errorResponse(err)
	}

	force := false

	forceParam := r.URL.Query().Get("force")

	if forceParam != "" {
		parsedForce, err := strconv.ParseBool(forceParam)

		if err != nil {
			return &customerrors.HttpError{
				Message: "force must be true or false",
				Code:    http.StatusBadRequest,
			}
		}

		force = parsedForce
	}

	_, err = h.service.DeleteMovie(movieId, force)

	if err != nil {
		return errorResponse(err)
	}

	w.WriteHeader(http.StatusNoContent)

	return nil
}

// extra
func (h *MovieHandler) Search(w http.ResponseWriter, r *http.Request) *customerrors.HttpError {
	if r.Method != http.MethodGet {
		return &customerrors.HttpError{Message: "method not allowed", Code: http.StatusMethodNotAllowed}
	}

	title := r.URL.Query().Get("title")

	movies, err := h.service.SearchMovies(title)

	if err != nil {
		return errorResponse(err)
	}

	response, err := json.Marshal(movies)

	if err != nil {
		return &customerrors.HttpError{Message: "failed to encode JSON", Code: http.StatusInternalServerError}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(response)

	return nil
}

// helper
func parsePageAndSize(param, paramName string) (int, error) {
	param = strings.TrimSpace(param)
	value, err := strconv.Atoi(param)

	if paramName == "page" {
		if err != nil || value < 0 {
			return 0, &customerrors.HttpError{
				Message: "page should be a non-negative number",
				Code:    http.StatusBadRequest,
			}
		}
	} else if paramName == "size" {
		if err != nil || value <= 0 || value > 100 {
			return 0, &customerrors.HttpError{
				Message: "size should be between 1 and 100",
				Code:    http.StatusBadRequest,
			}
		}
	}

	return value, nil
}

func parseIdFromParam(param, paramName string) (int, error) {
	param = strings.TrimSpace(param)
	id, err := strconv.Atoi(param)

	message := fmt.Sprintf("%s should be positive number", paramName)
	code := http.StatusBadRequest

	if err != nil || id <= 0 {
		return 0, &customerrors.HttpError{Message: message, Code: code}
	}

	return id, nil
}

func parseIdsFromParam(param string) ([]int, error) {
	if len(param) == 0 {
		return []int{}, nil
	}

	param = strings.TrimSpace(param)
	paramArr := strings.Split(param, ",")
	ids := []int{}

	for _, val := range paramArr {
		id, err := strconv.Atoi(val)
		if err != nil || id <= 0 {
			return nil, fmt.Errorf("invalid id %q", val)
		}
		ids = append(ids, id)
	}

	return ids, nil
}

func errorResponse(err error) *customerrors.HttpError {
	if err == nil {
		return nil
	}

	var httpErr *customerrors.HttpError
	if errors.As(err, &httpErr) {
		return httpErr
	}

	switch {
	case errors.Is(err, customerrors.ErrDB):
		return &customerrors.HttpError{
			Message: "internal server error",
			Code:    http.StatusInternalServerError,
		}
	case errors.Is(err, customerrors.ErrNotFound):
		return &customerrors.HttpError{
			Message: "not found",
			Code:    http.StatusNotFound,
		}
	case errors.Is(err, customerrors.ErrConcurrentModification):
		return &customerrors.HttpError{
			Message: "movie was modified by another user",
			Code:    http.StatusConflict,
		}
	case errors.Is(err, customerrors.ErrConflict):
		return &customerrors.HttpError{
			Message: err.Error(),
			Code:    http.StatusConflict,
		}
	case errors.Is(err, customerrors.ErrInvalidReference):
		return &customerrors.HttpError{
			Message: "non existing id passed",
			Code:    http.StatusBadRequest,
		}
	default:
		log.Println("Unknown error", err)
		return &customerrors.HttpError{
			Message: "internal server error",
			Code:    http.StatusInternalServerError,
		}
	}
}

func hasDuplicateActorIds(actors []entity.Actor) bool {
	seen := make(map[uint]bool)

	for _, actor := range actors {
		if seen[actor.Id] {
			return true
		}

		seen[actor.Id] = true
	}

	return false
}

func hasDuplicateGenreIds(genres []entity.Genre) bool {
	seen := make(map[uint]bool)

	for _, genre := range genres {
		if seen[genre.Id] {
			return true
		}

		seen[genre.Id] = true
	}

	return false
}
