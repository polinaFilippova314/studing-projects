package entity

import (
	"errors"
	"fmt"
	"time"
)

type Actor struct {
	Id        uint
	Name      string
	BirthDate time.Time
	Version   int

	Movies   []Movie `json:"movies,omitempty"`
	MovieIds []int   `json:"movieIds,omitempty"`
}

func (a *Actor) Validate() error {
	var err1, err2 error
	if a.Name == "" {
		err1 = fmt.Errorf("%w: name is empty", ErrInvalidContent)
	}
	err2 = ValidateDate(a.BirthDate, a.Name)
	return errors.Join(err1, err2)
}
func ValidateDate(date time.Time, name string) error {
	if date.After(time.Now()) {
		return fmt.Errorf("%w: the actor %s isn't born yet", ErrInvalidContent, name)
	}
	return nil
}

type PaginatedActorResponse struct {
	Actors []Actor
	Page   uint
	Size   uint
	Total  uint
}
type ActorCreateRequest struct {
	Name      string
	BirthDate string
	MovieIds  []int
}
type ActorPatchRequest struct {
	Name      *string
	BirthDate *string
	Version   int
	MovieIds  []int
}

func (a *ActorPatchRequest) Validate() error {
	if a.Version < 1 {
		return fmt.Errorf("%w: version number is required", ErrInvalidContent)
	}
	return nil
}
