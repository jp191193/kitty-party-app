package contribution

import (
	"context"
	"time"

	"kitty-party-app/internal/apperrors"

	"github.com/google/uuid"
)

// GroupMemberProvider defines how we fetch members for a group.
type GroupMemberProvider interface {
	GetMemberIDs(ctx context.Context, groupID string) ([]string, error)
}

// Service contains the business logic for contributions.
type Service interface {
	GenerateMonthlyDues(ctx context.Context, req GenerateDuesRequest) ([]*ContributionDue, error)
	RecordPayment(ctx context.Context, req RecordPaymentRequest) (*ContributionPayment, error)
	GetGroupDues(ctx context.Context, groupID string, cycleMonth *int) ([]*ContributionDue, error)
}

type service struct {
	repo           Repository
	memberProvider GroupMemberProvider
}

// NewService returns a new Contribution Service.
func NewService(repo Repository, memberProvider GroupMemberProvider) Service {
	return &service{
		repo:           repo,
		memberProvider: memberProvider,
	}
}

func (s *service) GenerateMonthlyDues(ctx context.Context, req GenerateDuesRequest) ([]*ContributionDue, error) {
	memberIDs, err := s.memberProvider.GetMemberIDs(ctx, req.GroupID)
	if err != nil {
		return nil, err
	}
	
	if len(memberIDs) == 0 {
		return nil, apperrors.ErrBadRequest // Or another custom error
	}

	var generated []*ContributionDue
	for _, mID := range memberIDs {
		due := &ContributionDue{
			ID:             uuid.NewString(),
			GroupID:        req.GroupID,
			MemberID:       mID,
			HostMemberID:   req.HostMemberID,
			CycleMonth:     req.CycleMonth,
			CycleYear:      req.CycleYear,
			CycleNumber:    req.CycleNumber,
			KittyPartyDate: req.KittyPartyDate,
			DueDate:        req.DueDate,
			Amount:         req.Amount,
			Status:         "PENDING",
		}
		
		// Set to OVERDUE immediately if the due date is already in the past
		if time.Now().After(req.DueDate) {
			due.Status = "OVERDUE"
		}

		err := s.repo.GenerateDue(ctx, due)
		if err != nil {
			// If already exists, we might want to skip or return conflict.
			// Depending on business rules, we can return the error.
			return nil, err
		}
		generated = append(generated, due)
	}

	return generated, nil
}

func (s *service) RecordPayment(ctx context.Context, req RecordPaymentRequest) (*ContributionPayment, error) {
	// Let's verify the due exists
	_, err := s.repo.GetDue(ctx, req.DueID)
	if err != nil {
		return nil, err
	}

	payment := &ContributionPayment{
		ID:             uuid.NewString(),
		DueID:          req.DueID,
		AmountPaid:     req.AmountPaid,
		PaymentDate:    time.Now().UTC(),
		PaymentMode:    req.PaymentMode,
		TransactionRef: req.TransactionRef,
		Status:         req.Status,
	}

	err = s.repo.RecordPayment(ctx, payment)
	if err != nil {
		return nil, err
	}

	return payment, nil
}

func (s *service) GetGroupDues(ctx context.Context, groupID string, cycleMonth *int) ([]*ContributionDue, error) {
	return s.repo.ListGroupDues(ctx, groupID, cycleMonth)
}
