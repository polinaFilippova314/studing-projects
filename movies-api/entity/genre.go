package entity

import "fmt"

type Genre struct {
	Id      uint
	Name    string
	Version int

	Movies   []Movie `json:"movies,omitempty"`
	MovieIds []int   `json:"movieIds,omitempty"`
}
type GenrePatchRequest struct {
	Name     *string
	Version  int
	MovieIds []int
}

func (g *Genre) Validate() error {
	if g.Name == "" {
		return fmt.Errorf("%w: there is no name for genre", ErrInvalidContent)
	}
	return nil
}
func (g *GenrePatchRequest) Validate() error {
	if g.Version < 1 {
		return fmt.Errorf("%w: version number is required", ErrInvalidContent)
	}
	return nil
}
