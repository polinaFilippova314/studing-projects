package customerrors

import (
	"log"
	"net/http"
	"errors"
)

type HttpError struct {
	Err     error
	Message string
	Code    int
}

func (e *HttpError) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return e.Message
}

type HttpErrorHandler func(http.ResponseWriter, *http.Request) *HttpError

func (fn HttpErrorHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if e := fn(w, r); e != nil {
		if e.Err != nil {
			log.Println(e.Err)
		}
		http.Error(w, e.Message, e.Code)
	}
}

//Repository Errors
var (
	ErrDB = errors.New("database error")
	ErrNotFound = errors.New("resource not found")
	ErrConflict = errors.New("conflict")
	ErrInvalidReference = errors.New("invalid reference")
	ErrConcurrentModification = errors.New("movie was modified by another request")
)