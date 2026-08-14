package entity

import (
	"fmt"
	"errors"
	"time"
	"strings"
)

type Movie struct {
	Id uint
	Title string
	ReleaseYear int
	Duration int
	Version int

	Actors []Actor
	Genres []Genre
}

func(m Movie) ValidateMovie() error {
	if strings.TrimSpace(m.Title) == "" {
		return errors.New("movie title cannot be empty")
	}

	if m.ReleaseYear == 0 {
		return errors.New("movie release year cannot be empty")
	}

	currentYear := time.Now().Year()

	if m.ReleaseYear < 1888 {
		return errors.New("movie release year cannot be before 1888")
	}

	if m.ReleaseYear > currentYear {
		return fmt.Errorf("movie has not been released yet: %d", m.ReleaseYear)
	}

	if m.Duration <= 0 {
		return errors.New("movie duration must be greater than zero")
	}

	return nil
} 

type MoviePatch struct {
    Title *string 
    ReleaseYear *int    
    Duration *int   

    Actors *[]Actor 
    Genres *[]Genre
 	Version     *int `json:"version"`
}