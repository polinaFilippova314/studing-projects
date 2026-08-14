package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"movies-api/customerrors"
	"movies-api/entity"
	"movies-api/service"
)

type ActorHandler struct {
	service *service.ActorService
}

func NewActorHandler(service *service.ActorService) *ActorHandler {
	return &ActorHandler{service: service}
}
func (h *ActorHandler) Create(w http.ResponseWriter, r *http.Request) *customerrors.HttpError {
	var actorRaw entity.ActorCreateRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&actorRaw)
	if err != nil {
		return &customerrors.HttpError{Err: err, Message: "invalid json", Code: http.StatusBadRequest}
	}
	var actor entity.Actor
	actor.Name = actorRaw.Name
	actor.MovieIds = actorRaw.MovieIds
	date, err := time.Parse("2006-01-02", actorRaw.BirthDate)
	if err != nil {
		return &customerrors.HttpError{Err: err, Message: "invalid json: invalid data format", Code: http.StatusBadRequest}
	}
	actor.BirthDate = date
	actor.Version = 1
	id, err := h.service.CreateActor(&actor)
	actor.Id = uint(id)
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
	json.NewEncoder(w).Encode(actor)
	return nil
}
func (h *ActorHandler) GetAll(w http.ResponseWriter, r *http.Request) *customerrors.HttpError {
	name := r.URL.Query().Get("name")
	gotMovies := r.URL.Query().Get("movies")
	movies := gotMovies == "true"
	if name == "" {
		page := r.URL.Query().Get("page")
		size := r.URL.Query().Get("size")
		pagination := true
		pageInt, sizeInt := 0, 0
		var err error
		if page == "" && size == "" {
			pagination = false
		}
		if pagination {
			pageInt, err = strconv.Atoi(page)
			if err != nil || pageInt < 0 {
				return &customerrors.HttpError{Err: err, Message: "invalid page number", Code: http.StatusBadRequest}
			}
			sizeInt, err = strconv.Atoi(size)
			if err != nil || sizeInt <= 0 {
				return &customerrors.HttpError{Err: err, Message: "invalid size number", Code: http.StatusBadRequest}
			}
		}
		actors, err := h.service.GetAll(movies, pageInt, sizeInt, pagination)
		if err != nil {
			if errors.Is(err, entity.ErrNotFound) {
				return &customerrors.HttpError{Err: err, Message: err.Error(), Code: http.StatusNotFound}
			}
			return &customerrors.HttpError{Err: err, Message: err.Error(), Code: http.StatusInternalServerError}
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(actors)
		return nil
	}
	actors, err := h.service.GetByName(name)
	if err != nil {
		return &customerrors.HttpError{Err: err, Message: err.Error(), Code: http.StatusInternalServerError}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(actors)
	return nil
}
func (h *ActorHandler) GetByID(w http.ResponseWriter, r *http.Request) *customerrors.HttpError {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		return &customerrors.HttpError{Err: err, Message: "invalid id", Code: http.StatusBadRequest}
	}
	actor, err := h.service.GetByID(id)
	if err != nil {
		if errors.Is(err, entity.ErrNotFound) {
			return &customerrors.HttpError{Err: err, Message: err.Error(), Code: http.StatusNotFound}
		}
		return &customerrors.HttpError{Err: err, Message: err.Error(), Code: http.StatusInternalServerError}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(actor)
	return nil
}
func (h *ActorHandler) Update(w http.ResponseWriter, r *http.Request) *customerrors.HttpError {
	idActor := r.PathValue("id")
	id, err := strconv.Atoi(idActor)
	if err != nil || id <= 0 {
		return &customerrors.HttpError{Err: err, Message: "invalid id", Code: http.StatusBadRequest}
	}
	var actorUpdate entity.ActorPatchRequest
	err1 := json.NewDecoder(r.Body).Decode(&actorUpdate)
	if err1 != nil {
		return &customerrors.HttpError{Err: err, Message: "invalid json", Code: http.StatusBadRequest}
	}
	actor, err := h.service.Update(id, actorUpdate)
	if err != nil {
		if errors.Is(err, entity.ErrNotFound) {
			return &customerrors.HttpError{Err: err, Message: err.Error(), Code: http.StatusNotFound}
		}
		if errors.Is(err, entity.ErrInvalidContent) {
			return &customerrors.HttpError{Err: err, Message: err.Error(), Code: http.StatusBadRequest}
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
func (h *ActorHandler) Delete(w http.ResponseWriter, r *http.Request) *customerrors.HttpError {
	idActor := r.PathValue("id")
	id, err := strconv.Atoi(idActor)
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
func (h *ActorHandler) DeleteConnection(w http.ResponseWriter, r *http.Request) *customerrors.HttpError {
	idActor := r.PathValue("id")
	id, err := strconv.Atoi(idActor)
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
func (h *ActorHandler) CheckDuplicates(w http.ResponseWriter, r *http.Request) *customerrors.HttpError {
	actorsDuplicate, err := h.service.CheckDuplicates()
	if err != nil {
		return &customerrors.HttpError{Err: err, Message: err.Error(), Code: http.StatusInternalServerError}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(actorsDuplicate)
	return nil
}
