// Package kittycycle – service (business-logic) layer.
package kittycycle

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"kitty-party-app/internal/apperrors"
)

// DuesGenerator generates contribution dues for a cycle's first active month.
// Implementations wrap the contribution domain's service, avoiding a direct import.
type DuesGenerator interface {
	GenerateDues(ctx context.Context, groupID, hostMemberID string,
		cycleNumber, cycleMonth, cycleYear int,
		kittyDate, dueDate time.Time, amount float64) (int, error)
}

// PayoutScheduler schedules a payout for the host of a kitty cycle month.
// Implementations wrap the payout domain's service, avoiding a direct import.
type PayoutScheduler interface {
	SchedulePayout(callerID, groupID, recipientID string,
		cycleNumber, cycleMonth, cycleYear int,
		scheduledDate time.Time, poolAmount float64) (string, error)
}

// Service defines the use-case contract for the KittyCycle domain.
type Service interface {
	// ScheduleCycle creates a kitty_cycle and all its kitty_schedule rows.
	// Only the group's creator or an ADMIN member may do this.
	ScheduleCycle(callerID string, req ScheduleCycleRequest) (*KittyCycleWithSchedule, error)

	// GetCycle returns a cycle and its full schedule.
	GetCycle(id string) (*KittyCycleWithSchedule, error)

	// GetCyclesByGroup returns every cycle ever created for a group.
	GetCyclesByGroup(groupID string) ([]*KittyCycle, error)

	// CancelCycle cancels a cycle that has not yet completed.
	CancelCycle(callerID string, id string, req CancelCycleRequest) error

	// StartCycle transitions a PENDING cycle to ACTIVE, then automatically
	// generates contribution dues and schedules a payout for the first month.
	// Only the group's creator or an ADMIN member may do this.
	StartCycle(callerID string, id string) (*StartCycleResult, error)

	// RollHost picks a random eligible group member and assigns them as
	// the host of the given schedule entry. Eligible members are those who
	// have not already been picked as host in any other entry of the same
	// cycle. Only the group's creator or an ADMIN member may do this.
	RollHost(callerID string, scheduleID string) (*RollHostResult, error)
}

type service struct {
	repo            Repository
	groupProvider   GroupInfoProvider
	duesGenerator   DuesGenerator   // nil if activation extras not configured
	payoutScheduler PayoutScheduler // nil if activation extras not configured
}

// NewService constructs a KittyCycle Service without activation side-effects.
func NewService(repo Repository, groupProvider GroupInfoProvider) Service {
	return NewServiceWithActivation(repo, groupProvider, nil, nil)
}

// NewServiceWithActivation constructs a KittyCycle Service that also generates
// contribution dues and schedules a payout when a cycle is started.
func NewServiceWithActivation(repo Repository, groupProvider GroupInfoProvider, duesGen DuesGenerator, payoutSch PayoutScheduler) Service {
	return &service{
		repo:            repo,
		groupProvider:   groupProvider,
		duesGenerator:   duesGen,
		payoutScheduler: payoutSch,
	}
}

func (s *service) ScheduleCycle(callerID string, req ScheduleCycleRequest) (*KittyCycleWithSchedule, error) {
	g, err := s.groupProvider.GetGroup(req.GroupID)
	if err != nil {
		return nil, err
	}

	isAdmin, err := s.groupProvider.IsAdminOrCreator(req.GroupID, callerID)
	if err != nil {
		return nil, err
	}
	if !isAdmin {
		return nil, apperrors.New(http.StatusForbidden, "only group admin or creator can schedule a kitty cycle")
	}

	if !req.EndDate.After(req.StartDate) {
		return nil, apperrors.New(http.StatusBadRequest, "end_date must be after start_date")
	}

	totalCycles := len(req.Schedule)
	if totalCycles > g.Duration {
		return nil, apperrors.New(http.StatusBadRequest,
			fmt.Sprintf("schedule entries (%d) exceed group duration of %d", totalCycles, g.Duration))
	}

	poolAmount := req.MonthlyAmount * float64(totalCycles)

	seenNumbers := make(map[int]bool, totalCycles)
	seenHosts := make(map[string]bool, totalCycles)

	entries := make([]*KittySchedule, 0, totalCycles)
	for _, e := range req.Schedule {
		if seenNumbers[e.CycleNumber] {
			return nil, apperrors.New(http.StatusBadRequest,
				fmt.Sprintf("duplicate cycle_number %d in schedule", e.CycleNumber))
		}
		if e.CycleNumber > totalCycles {
			return nil, apperrors.New(http.StatusBadRequest,
				fmt.Sprintf("cycle_number %d exceeds schedule length %d", e.CycleNumber, totalCycles))
		}
		if !e.DueDate.Before(e.ScheduledDate) && !e.DueDate.Equal(e.ScheduledDate) {
			return nil, apperrors.New(http.StatusBadRequest,
				fmt.Sprintf("cycle %d: due_date must be on or before scheduled_date", e.CycleNumber))
		}

		// Hosts are now optional at scheduling time — they're picked later
		// via the dice roll. Validate only when a host was supplied up front.
		if e.HostMemberID != nil && *e.HostMemberID != "" {
			hostID := *e.HostMemberID
			if seenHosts[hostID] {
				return nil, apperrors.New(http.StatusBadRequest,
					fmt.Sprintf("host %s appears more than once in schedule", hostID))
			}
			isMember, err := s.groupProvider.IsMember(req.GroupID, hostID)
			if err != nil {
				return nil, err
			}
			if !isMember {
				return nil, apperrors.New(http.StatusBadRequest,
					fmt.Sprintf("host %s is not a member of this group", hostID))
			}
			seenHosts[hostID] = true
		}

		seenNumbers[e.CycleNumber] = true

		entries = append(entries, &KittySchedule{
			CycleNumber:   e.CycleNumber,
			CycleMonth:    e.CycleMonth,
			CycleYear:     e.CycleYear,
			HostMemberID:  e.HostMemberID,
			ScheduledDate: e.ScheduledDate,
			DueDate:       e.DueDate,
			PoolAmount:    poolAmount,
			Venue:         e.Venue,
			Notes:         e.Notes,
		})
	}

	cycle := &KittyCycle{
		GroupID:       req.GroupID,
		Name:          req.Name,
		TotalCycles:   totalCycles,
		MonthlyAmount: req.MonthlyAmount,
		PoolAmount:    poolAmount,
		StartDate:     req.StartDate,
		EndDate:       req.EndDate,
		Notes:         req.Notes,
		CreatedBy:     callerID,
	}

	return s.repo.CreateCycleWithSchedule(cycle, entries)
}

func (s *service) GetCycle(id string) (*KittyCycleWithSchedule, error) {
	return s.repo.GetCycle(id)
}

func (s *service) GetCyclesByGroup(groupID string) ([]*KittyCycle, error) {
	return s.repo.ListByGroup(groupID)
}

func (s *service) StartCycle(callerID string, id string) (*StartCycleResult, error) {
	cycle, err := s.repo.GetCycle(id)
	if err != nil {
		return nil, err
	}

	isAdmin, err := s.groupProvider.IsAdminOrCreator(cycle.GroupID, callerID)
	if err != nil {
		return nil, err
	}
	if !isAdmin {
		return nil, apperrors.New(http.StatusForbidden, "only group admin or creator can start a kitty cycle")
	}

	started, err := s.repo.StartCycle(id)
	if err != nil {
		return nil, err
	}

	result := &StartCycleResult{KittyCycleWithSchedule: *started}

	// Find the first (now IN_PROGRESS) schedule entry.
	var firstEntry *KittySchedule
	for _, e := range started.Schedule {
		if firstEntry == nil || e.CycleNumber < firstEntry.CycleNumber {
			firstEntry = e
		}
	}

	if firstEntry != nil && firstEntry.HostMemberID != nil && *firstEntry.HostMemberID != "" {
		ctx := context.Background()
		hostID := *firstEntry.HostMemberID

		if s.duesGenerator != nil {
			n, genErr := s.duesGenerator.GenerateDues(ctx,
				started.GroupID, hostID,
				firstEntry.CycleNumber, firstEntry.CycleMonth, firstEntry.CycleYear,
				firstEntry.ScheduledDate, firstEntry.DueDate, started.MonthlyAmount)
			if genErr == nil {
				result.DuesGenerated = n
			}
		}

		if s.payoutScheduler != nil {
			payoutID, payErr := s.payoutScheduler.SchedulePayout(callerID,
				started.GroupID, hostID,
				firstEntry.CycleNumber, firstEntry.CycleMonth, firstEntry.CycleYear,
				firstEntry.ScheduledDate, started.PoolAmount)
			if payErr == nil {
				result.PayoutID = payoutID
			}
		}
	}

	return result, nil
}

func (s *service) RollHost(callerID string, scheduleID string) (*RollHostResult, error) {
	entry, err := s.repo.GetSchedule(scheduleID)
	if err != nil {
		return nil, err
	}

	isAdmin, err := s.groupProvider.IsAdminOrCreator(entry.GroupID, callerID)
	if err != nil {
		return nil, err
	}
	if !isAdmin {
		return nil, apperrors.New(http.StatusForbidden, "only group admin or creator can roll the host dice")
	}

	if entry.HostMemberID != nil && *entry.HostMemberID != "" {
		return nil, apperrors.New(http.StatusConflict, "this cycle already has a host assigned")
	}
	if entry.Status == "CANCELLED" || entry.Status == "COMPLETED" {
		return nil, apperrors.New(http.StatusConflict, "cannot roll a host for a "+entry.Status+" cycle")
	}

	members, err := s.groupProvider.ListMembers(entry.GroupID)
	if err != nil {
		return nil, err
	}
	if len(members) == 0 {
		return nil, apperrors.New(http.StatusBadRequest, "no members in this group to roll a host from")
	}

	taken, err := s.repo.ListAssignedHosts(entry.CycleID)
	if err != nil {
		return nil, err
	}

	eligible := make([]MemberInfo, 0, len(members))
	for _, m := range members {
		if !taken[m.MemberID] {
			eligible = append(eligible, m)
		}
	}
	if len(eligible) == 0 {
		return nil, apperrors.New(http.StatusConflict, "every member has already hosted in this cycle")
	}

	picked := eligible[rand.Intn(len(eligible))]

	updated, err := s.repo.AssignHost(scheduleID, picked.MemberID)
	if err != nil {
		return nil, err
	}

	return &RollHostResult{
		ScheduleID:     updated.ID,
		HostMemberID:   picked.MemberID,
		HostMemberName: picked.MemberName,
		Schedule:       updated,
	}, nil
}

func (s *service) CancelCycle(callerID string, id string, req CancelCycleRequest) error {
	cycle, err := s.repo.GetCycle(id)
	if err != nil {
		return err
	}

	isAdmin, err := s.groupProvider.IsAdminOrCreator(cycle.GroupID, callerID)
	if err != nil {
		return err
	}
	if !isAdmin {
		return apperrors.New(http.StatusForbidden, "only group admin or creator can cancel a kitty cycle")
	}

	return s.repo.CancelCycle(id, req.Notes)
}
