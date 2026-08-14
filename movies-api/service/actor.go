package service

import (
	"time"

	"movies-api/entity"
	"movies-api/repository"
)

type ActorService struct {
	repo repository.ActorRepository
}

func NewActorService(repo repository.ActorRepository) *ActorService {
	return &ActorService{
		repo: repo,
	}
}
func (s *ActorService) CreateActor(actor *entity.Actor) (int64, error) {
	if err := actor.Validate(); err != nil {
		return 0, err
	}
	id, err := s.repo.Create(actor)
	if err != nil {
		return 0, err
	}
	return id, nil
}
func (s *ActorService) GetAll(movies bool, page int, size int, pagination bool) (entity.PaginatedActorResponse, error) {
	actors, err := s.repo.GetAll(movies, page, size, pagination)
	if err != nil {
		return entity.PaginatedActorResponse{}, err
	}
	return actors, nil
}
func (s *ActorService) GetByID(id int) (entity.Actor, error) {
	actor, err := s.repo.GetByID(id)
	if err != nil {
		return entity.Actor{}, err
	}
	return actor, nil
}
func (s *ActorService) GetByName(name string) ([]entity.Actor, error) {
	actors, err := s.repo.GetByName(name)
	if err != nil {
		return []entity.Actor{}, err
	}
	return actors, nil
}
func (s *ActorService) Update(id int, actor entity.ActorPatchRequest) (entity.Actor, error) {
	if err := actor.Validate(); err != nil {
		return entity.Actor{}, err
	}
	if actor.BirthDate != nil {
		date, err := time.Parse("2006-01-02", *actor.BirthDate)
		if err != nil {
			return entity.Actor{}, err
		}
		name := ""
		if actor.Name != nil {
			name = *actor.Name
		}
		if err = entity.ValidateDate(date, name); err != nil {
			return entity.Actor{}, err
		}
	}
	actorUpdated, err := s.repo.Update(id, actor)
	if err != nil {
		return entity.Actor{}, err
	}
	return actorUpdated, nil
}
func (s *ActorService) Delete(id int, force bool) error {
	_, err := s.repo.Delete(id, force)
	if err != nil {
		return err
	}
	return nil
}
func (s *ActorService) DeleteConnection(id int, moviesId []int) error {
	_, err := s.repo.DeleteConnection(id, moviesId)
	if err != nil {
		return err
	}
	return nil
}
func (s *ActorService) CheckDuplicates() ([]entity.Actor, error) {
	actors, err := s.repo.CheckDuplicates()
	if err != nil {
		return []entity.Actor{}, err
	}
	return actors, nil
}
