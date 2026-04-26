// Package member – service (business-logic) layer.
package member

import "golang.org/x/crypto/bcrypt"

// Service defines the use-case contract for the Member domain.
type Service interface {
	GetAll() ([]*Member, error)
	GetByID(id string) (*Member, error)
	Create(req CreateMemberRequest) (*Member, error)
	Update(id string, req UpdateMemberRequest) (*Member, error)
	Delete(id string) error
}

type service struct {
	repo Repository
}

// NewService constructs a Service backed by the provided Repository.
func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) GetAll() ([]*Member, error) {
	return s.repo.FindAll()
}

func (s *service) GetByID(id string) (*Member, error) {
	return s.repo.FindByID(id)
}

func (s *service) Create(req CreateMemberRequest) (*Member, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	req.PasswordHash = string(hash)
	return s.repo.Create(req)
}

func (s *service) Update(id string, req UpdateMemberRequest) (*Member, error) {
	return s.repo.Update(id, req)
}

func (s *service) Delete(id string) error {
	return s.repo.Delete(id)
}
