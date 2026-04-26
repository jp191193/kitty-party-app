package contributiontracking

import (
	"context"
	"time"
)

// Service encapsulates the business rules for contribution tracking.
type Service interface {
	Create(ctx context.Context, req CreateTrackingRequest) (*ContributionTracking, error)
	Get(ctx context.Context, id string) (*ContributionTracking, error)
	GetByDueID(ctx context.Context, dueID string) (*ContributionTracking, error)
	ListByGroup(ctx context.Context, groupID string, cycleMonth, cycleYear *int) ([]*ContributionTracking, error)
	ListByMember(ctx context.Context, memberID string) ([]*ContributionTracking, error)
	ApplyPayment(ctx context.Context, id string, req UpdateTrackingRequest) (*ContributionTracking, error)
}

type service struct {
	repo Repository
}

// NewService returns a new tracking Service.
func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) Create(ctx context.Context, req CreateTrackingRequest) (*ContributionTracking, error) {
	t := &ContributionTracking{
		GroupID:     req.GroupID,
		MemberID:    req.MemberID,
		DueID:       req.DueID,
		CycleMonth:  req.CycleMonth,
		CycleYear:   req.CycleYear,
		CycleNumber: req.CycleNumber,
		TotalDue:    req.TotalDue,
		TotalPaid:   0,
		Balance:     req.TotalDue,
		Status:      StatusPending,
	}
	if err := s.repo.Create(ctx, t); err != nil {
		return nil, err
	}
	return t, nil
}

func (s *service) Get(ctx context.Context, id string) (*ContributionTracking, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *service) GetByDueID(ctx context.Context, dueID string) (*ContributionTracking, error) {
	return s.repo.GetByDueID(ctx, dueID)
}

func (s *service) ListByGroup(ctx context.Context, groupID string, cycleMonth, cycleYear *int) ([]*ContributionTracking, error) {
	return s.repo.ListByGroup(ctx, groupID, cycleMonth, cycleYear)
}

func (s *service) ListByMember(ctx context.Context, memberID string) ([]*ContributionTracking, error) {
	return s.repo.ListByMember(ctx, memberID)
}

func (s *service) ApplyPayment(ctx context.Context, id string, req UpdateTrackingRequest) (*ContributionTracking, error) {
	t, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	t.TotalPaid += req.AmountPaid
	t.Balance = t.TotalDue - t.TotalPaid
	if t.Balance < 0 {
		t.Balance = 0
	}

	paidAt := time.Now().UTC()
	if req.LastPaymentDate != nil {
		paidAt = *req.LastPaymentDate
	}
	t.LastPaymentDate = &paidAt

	t.Status = deriveStatus(t.TotalPaid, t.TotalDue)

	if err := s.repo.Update(ctx, t); err != nil {
		return nil, err
	}
	return t, nil
}

func deriveStatus(paid, due float64) string {
	switch {
	case paid <= 0:
		return StatusPending
	case paid >= due:
		return StatusPaid
	default:
		return StatusPartial
	}
}
