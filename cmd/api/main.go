// main is the application entry point.
// It wires all dependencies together (poor-man's DI container) and starts
// the HTTP server.
package main

import (
	"context"
	"log"
	"time"

	"kitty-party-app/internal/auth"
	"kitty-party-app/internal/config"
	"kitty-party-app/internal/contribution"
	"kitty-party-app/internal/contributiontracking"
	"kitty-party-app/internal/database"
	"kitty-party-app/internal/group"
	"kitty-party-app/internal/kittycycle"
	"kitty-party-app/internal/logger"
	"kitty-party-app/internal/member"
	"kitty-party-app/internal/payout"
	"kitty-party-app/internal/profile"
	"kitty-party-app/internal/router"
	"kitty-party-app/internal/server"

	"go.uber.org/zap"
)

// contributionDuesGenAdapter adapts contribution.Service to kittycycle.DuesGenerator.
type contributionDuesGenAdapter struct {
	svc contribution.Service
}

func (a *contributionDuesGenAdapter) GenerateDues(ctx context.Context,
	groupID, hostMemberID string,
	cycleNumber, cycleMonth, cycleYear int,
	kittyDate, dueDate time.Time, amount float64) (int, error) {

	dues, err := a.svc.GenerateMonthlyDues(ctx, contribution.GenerateDuesRequest{
		GroupID:        groupID,
		HostMemberID:   hostMemberID,
		CycleNumber:    cycleNumber,
		CycleMonth:     cycleMonth,
		CycleYear:      cycleYear,
		KittyPartyDate: kittyDate,
		DueDate:        dueDate,
		Amount:         amount,
	})
	if err != nil {
		return 0, err
	}
	return len(dues), nil
}

// payoutSchedulerAdapter adapts payout.Service to kittycycle.PayoutScheduler.
type payoutSchedulerAdapter struct {
	svc payout.Service
}

func (a *payoutSchedulerAdapter) SchedulePayout(callerID, groupID, recipientID string,
	cycleNumber, cycleMonth, cycleYear int,
	scheduledDate time.Time, poolAmount float64) (string, error) {

	schedule, err := a.svc.SchedulePayout(callerID, payout.CreateScheduleRequest{
		GroupID:       groupID,
		RecipientID:   recipientID,
		CycleNumber:   cycleNumber,
		CycleMonth:    cycleMonth,
		CycleYear:     cycleYear,
		ScheduledDate: scheduledDate,
		PoolAmount:    poolAmount,
	})
	if err != nil {
		return "", err
	}
	return schedule.ID, nil
}

func main() {
	// ── 1. Load configuration ────────────────────────────────────────────────
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	// ── 2. Initialise logger ─────────────────────────────────────────────────
	l, err := logger.New(cfg.AppEnv)
	if err != nil {
		log.Fatalf("logger: %v", err)
	}
	defer func() { _ = l.Sync() }()

	l.Info("starting application",
		zap.String("name", cfg.AppName),
		zap.String("env", cfg.AppEnv),
		zap.String("port", cfg.Port),
	)

	// ── 3. Initialize Database ───────────────────────────────────────────────
	dbPool, err := database.NewPostgresPool(context.Background(), cfg.DatabaseURL)
	if err != nil {
		l.Fatal("failed to connect to database", zap.Error(err))
	}
	defer dbPool.Close()

	// ── 4. Wire up dependencies (Repository → Service → Handler) ─────────────

	// Member domain
	memberRepo := member.NewPostgresRepository(dbPool)
	memberSvc := member.NewService(memberRepo)
	memberHandler := member.NewHandler(memberSvc)

	// Group domain
	groupRepo := group.NewPostgresRepository(dbPool)
	groupSvc := group.NewService(groupRepo)
	groupHandler := group.NewHandler(groupSvc)

	// Group Membership domain
	// groupRepo is shared so the membership service can validate group existence.
	membershipRepo := group.NewPostgresMembershipRepository(dbPool)
	membershipSvc := group.NewMembershipService(groupRepo, membershipRepo)
	membershipHandler := group.NewMembershipHandler(membershipSvc)

	// Contribution domain
	contributionRepo := contribution.NewPostgresRepository(dbPool)
	memberProvider := contribution.NewGroupMemberProviderAdapter(membershipRepo)
	contributionSvc := contribution.NewService(contributionRepo, memberProvider)
	contributionHandler := contribution.NewHandler(contributionSvc)

	// Contribution Tracking domain (per-member, per-cycle rollup)
	trackingRepo := contributiontracking.NewPostgresRepository(dbPool)
	trackingSvc := contributiontracking.NewService(trackingRepo)
	trackingHandler := contributiontracking.NewHandler(trackingSvc)

	// Payout domain
	payoutRepo := payout.NewPostgresRepository(dbPool)
	groupInfoProvider := payout.NewGroupInfoProviderAdapter(groupRepo, membershipRepo)
	payoutSvc := payout.NewService(payoutRepo, groupInfoProvider)
	payoutHandler := payout.NewHandler(payoutSvc)

	// Kitty Cycle domain — wired with activation side-effects:
	// starting a cycle auto-generates dues and schedules a payout for month 1.
	kittyCycleRepo := kittycycle.NewPostgresRepository(dbPool)
	kittyCycleGroupProvider := kittycycle.NewGroupInfoProviderAdapter(groupRepo, membershipRepo)
	kittyCycleSvc := kittycycle.NewServiceWithActivation(
		kittyCycleRepo,
		kittyCycleGroupProvider,
		&contributionDuesGenAdapter{svc: contributionSvc},
		&payoutSchedulerAdapter{svc: payoutSvc},
	)
	kittyCycleHandler := kittycycle.NewHandler(kittyCycleSvc)

	// Member Profile domain
	profileRepo := profile.NewPostgresRepository(dbPool)
	memberChecker := profile.NewMemberChecker(memberRepo)
	profileSvc := profile.NewService(profileRepo, memberChecker)
	profileHandler := profile.NewHandler(profileSvc)

	// Auth domain
	authHandler := auth.NewHandler()

	// ── 5. Build router ──────────────────────────────────────────────────────
	r := router.New(router.Options{
		Logger:              l,
		AuthHandler:         authHandler,
		MemberHandler:       memberHandler,
		GroupHandler:        groupHandler,
		MembershipHandler:   membershipHandler,
		ContributionHandler: contributionHandler,
		TrackingHandler:     trackingHandler,
		ProfileHandler:      profileHandler,
		PayoutHandler:       payoutHandler,
		KittyCycleHandler:   kittyCycleHandler,
	})

	// ── 6. Start server ──────────────────────────────────────────────────────
	srv := server.New(r, cfg.Port, l)
	srv.Run()
}
