// Package group – service (business-logic) layer.
//
// All business rules live here. The service knows nothing about HTTP or storage
// details — it depends only on the Repository interface.
package group

// Service defines the use-case contract for the Group domain.
type Service interface {
	GetAll() ([]*Group, error)
	GetByID(id string) (*Group, error)
	GetByCreator(createdBy string) ([]*Group, error)
	Create(req CreateGroupRequest) (*Group, error)
	Update(id string, req UpdateGroupRequest) (*Group, error)
	Delete(id string) error
}

type service struct {
	repo Repository
}

// NewService constructs a Service backed by the provided Repository.
func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) GetAll() ([]*Group, error) {
	return s.repo.FindAll()
}

func (s *service) GetByID(id string) (*Group, error) {
	return s.repo.FindByID(id)
}

func (s *service) GetByCreator(createdBy string) ([]*Group, error) {
	return s.repo.FindByCreator(createdBy)
}

func (s *service) Create(req CreateGroupRequest) (*Group, error) {
	return s.repo.Create(req)
}

func (s *service) Update(id string, req UpdateGroupRequest) (*Group, error) {
	return s.repo.Update(id, req)
}

func (s *service) Delete(id string) error {
	return s.repo.Delete(id)
}
