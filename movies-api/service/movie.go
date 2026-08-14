package service

import (
	"movies-api/entity"
	"movies-api/repository"
	"strings"
)

type MovieService struct {
	repo repository.MovieRepository
}

func NewMovieService(repo repository.MovieRepository) *MovieService {
	return &MovieService{
		repo: repo,
	}
}

func (s *MovieService) GetMovieById(id int) (*entity.Movie, error) {
	movie, err := s.repo.FindById(id)

	if err != nil {
		return nil, err
	}

	return movie, nil
}

func (s *MovieService) GetAllMovies() ([]*entity.Movie, error) {
	movies, err := s.repo.FindAll()

	if err != nil {
		return nil, err
	}

	return movies, nil
}

func (s *MovieService) GetMoviesWithPagination(page, size int) ([]*entity.Movie, error) {
	movies, err := s.repo.FindWithPagination(page, size)

	if err != nil {
		return nil, err
	}

	return movies, nil
}

func (s *MovieService) FindMoviesByActor(id int) ([]*entity.Movie, error) {
	movies, err := s.repo.FindByActor(id)

	if err != nil {
		return nil, err
	}

	return movies, nil
}

func (s *MovieService) FindMoviesByGenre(id int) ([]*entity.Movie, error) {
	movies, err := s.repo.FindByGenre(id)

	if err != nil {
		return nil, err
	}

	return movies, nil
}

func (s *MovieService) FindMoviesByYear(id int) ([]*entity.Movie, error) {
	movies, err := s.repo.FindByYear(id)

	if err != nil {
		return nil, err
	}

	return movies, nil
}

func (s *MovieService) FindMovieActors(id int) ([]entity.Actor, error) {
	actors, err := s.repo.FindActors(id)

	if err != nil {
		return nil, err
	}

	return actors, nil
}

func (s *MovieService) CreateMovie(movie *entity.Movie) (int64, error) {
	//make first letter upper and rest lower
	movie.Title = strings.Title(strings.ToLower(movie.Title))

	createdId, err := s.repo.Create(movie)

	if err != nil {
		return 0, err
	}

	return createdId, nil
}

func (s *MovieService) UpdateMovie(id int, patch *entity.MoviePatch) (int64, error) {
	return s.repo.Update(id, patch)
}

func (s *MovieService) DeleteMovie(id int, force bool) (int64, error) {
	return s.repo.Delete(id, force)
}

// extra
func (s *MovieService) SearchMovies(title string) ([]*entity.Movie, error) {
	title = strings.ToLower(strings.TrimSpace(title))
	exactMatch := strings.Title(title)

	movie, err := s.repo.FindByExactTitle(exactMatch)
	if err != nil {
		return nil, err
	}

	if movie != nil {
		return []*entity.Movie{movie}, nil
	}

	return s.repo.FindByTitleContains(title)
}
