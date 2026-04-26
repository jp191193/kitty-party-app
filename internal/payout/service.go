// Package payout – service (business-logic) layer.
package payout

import (
	"fmt"
	"net/http"

	"kitty-party-app/internal/apperrors"
)

// Service defines the use-case contract for the Payout domain.
type Service interface {
	// SchedulePayout creates a payout schedule entry for one cycle of a group.
	SchedulePayout(callerID string, req CreateScheduleRequest) (*PayoutSchedule, error)

	// RecordDisbursement records the actual money transfer that settles a scheduled payout.
	RecordDisbursement(callerID string, req RecordDisbursementRequest) (*Disbursement, error)

	// GetGroupPayouts returns all payout schedules for a group, each enriched with
	// its disbursement record if one exists.
	GetGroupPayouts(groupID string) ([]*PayoutSummary, error)

	// GetRecipientPayouts returns all payouts a specific member is set to receive.
	GetRecipientPayouts(recipientID string) ([]*PayoutSummary, error)

	// GetPayout returns a single payout schedule enriched with its disbursement.
	GetPayout(id string) (*PayoutSummary, error)

	// CancelPayout cancels a SCHEDULED payout. Fails if already DISBURSED.
	CancelPayout(callerID string, id string, req CancelScheduleRequest) error
}

type service struct {
	repo          Repository
	groupProvider GroupInfoProvider
}

// NewService constructs a payout Service.
func NewService(repo Repository, groupProvider GroupInfoProvider) Service {
	return &service{repo: repo, groupProvider: groupProvider}
}

func (s *service) SchedulePayout(callerID string, req CreateScheduleRequest) (*PayoutSchedule, error) {
	// Validate group exists and fetch its metadata.
	g, err := s.groupProvider.GetGroup(req.GroupID)
	if err != nil {
		return nil, err
	}

	// Only ADMIN or creator may schedule payouts.
	isAdmin, err := s.groupProvider.IsAdminOrCreator(req.GroupID, callerID)
	if err != nil {
		return nil, err
	}
	if !isAdmin {
		return nil, apperrors.New(http.StatusForbidden, "only group admin or creator can schedule payouts")
	}

	// cycle_number must be within the group's duration.
	if req.CycleNumber > g.Duration {
		return nil, apperrors.New(http.StatusBadRequest,
			fmt.Sprintf("cycle_number %d exceeds group duration of %d months", req.CycleNumber, g.Duration))
	}

	// Recipient must be an active member of the group.
	isMember, err := s.groupProvider.IsMember(req.GroupID, req.RecipientID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, apperrors.New(http.StatusBadRequest, "recipient is not a member of this group")
	}

	p := &PayoutSchedule{
		GroupID:       req.GroupID,
		RecipientID:   req.RecipientID,
		CycleNumber:   req.CycleNumber,
		CycleMonth:    req.CycleMonth,
		CycleYear:     req.CycleYear,
		ScheduledDate: req.ScheduledDate,
		PoolAmount:    req.PoolAmount,
		Notes:         req.Notes,
		CreatedBy:     callerID,
	}
	return s.repo.CreateSchedule(p)
}

func (s *service) RecordDisbursement(callerID string, req RecordDisbursementRequest) (*Disbursement, error) {
	// Load the payout to perform the admin check.
	schedule, err := s.repo.GetSchedule(req.PayoutID)
	if err != nil {
		return nil, err
	}

	isAdmin, err := s.groupProvider.IsAdminOrCreator(schedule.GroupID, callerID)
	if err != nil {
		return nil, err
	}
	if !isAdmin {
		return nil, apperrors.New(http.StatusForbidden, "only group admin or creator can record disbursements")
	}

	d := &Disbursement{
		PayoutID:       req.PayoutID,
		AmountPaid:     req.AmountPaid,
		PaymentMode:    req.PaymentMode,
		TransactionRef: req.TransactionRef,
		DisbursedBy:    callerID,
		Notes:          req.Notes,
	}
	return s.repo.RecordDisbursement(d)
}

func (s *service) GetGroupPayouts(groupID string) ([]*PayoutSummary, error) {
	schedules, err := s.repo.ListByGroup(groupID)
	if err != nil {
		return nil, err
	}
	return s.enrichWithDisbursements(schedules)
}

func (s *service) GetRecipientPayouts(recipientID string) ([]*PayoutSummary, error) {
	schedules, err := s.repo.ListByRecipient(recipientID)
	if err != nil {
		return nil, err
	}
	return s.enrichWithDisbursements(schedules)
}

func (s *service) GetPayout(id string) (*PayoutSummary, error) {
	schedule, err := s.repo.GetSchedule(id)
	if err != nil {
		return nil, err
	}
	disbursement, err := s.repo.GetDisbursementByPayout(id)
	if err != nil {
		return nil, err
	}
	return &PayoutSummary{PayoutSchedule: *schedule, Disbursement: disbursement}, nil
}

func (s *service) CancelPayout(callerID string, id string, req CancelScheduleRequest) error {
	schedule, err := s.repo.GetSchedule(id)
	if err != nil {
		return err
	}

	isAdmin, err := s.groupProvider.IsAdminOrCreator(schedule.GroupID, callerID)
	if err != nil {
		return err
	}
	if !isAdmin {
		return apperrors.New(http.StatusForbidden, "only group admin or creator can cancel payouts")
	}

	return s.repo.CancelSchedule(id, req.Notes)
}

// enrichWithDisbursements fetches the disbursement for each schedule and bundles them.
func (s *service) enrichWithDisbursements(schedules []*PayoutSchedule) ([]*PayoutSummary, error) {
	summaries := make([]*PayoutSummary, 0, len(schedules))
	for _, ps := range schedules {
		disbursement, err := s.repo.GetDisbursementByPayout(ps.ID)
		if err != nil {
			return nil, err
		}
		summaries = append(summaries, &PayoutSummary{
			PayoutSchedule: *ps,
			Disbursement:   disbursement,
		})
	}
	return summaries, nil
}
