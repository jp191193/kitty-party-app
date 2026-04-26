package router

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"kitty-party-app/internal/auth"
	"kitty-party-app/internal/contribution"
	"kitty-party-app/internal/contributiontracking"
	"kitty-party-app/internal/dateproposal"
	"kitty-party-app/internal/group"
	"kitty-party-app/internal/kittycycle"
	"kitty-party-app/internal/member"
	"kitty-party-app/internal/middleware"
	"kitty-party-app/internal/payout"
	"kitty-party-app/internal/profile"
	"kitty-party-app/internal/scheduleinvitation"
)

// Options groups all dependencies needed to build the router.
type Options struct {
	Logger                    *zap.Logger
	AuthHandler               *auth.Handler
	MemberHandler             *member.Handler
	GroupHandler              *group.Handler
	MembershipHandler         *group.MembershipHandler
	ContributionHandler       *contribution.Handler
	TrackingHandler           *contributiontracking.Handler
	ProfileHandler            *profile.Handler
	PayoutHandler             *payout.Handler
	KittyCycleHandler         *kittycycle.Handler
	ScheduleInvitationHandler *scheduleinvitation.Handler
	DateProposalHandler       *dateproposal.Handler
}

// New constructs and returns the root Gin engine with all routes registered.
func New(opts Options) *gin.Engine {
	engine := gin.New()

	// Global middleware
	engine.Use(
		middleware.Recovery(opts.Logger),
		middleware.Logger(opts.Logger),
		middleware.CORS(),
	)

	// Health-check – no auth required
	engine.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Versioned API group
	v1 := engine.Group("/api/v1")
	{
		if opts.AuthHandler != nil {
			opts.AuthHandler.RegisterRoutes(v1)
		}
		opts.MemberHandler.RegisterRoutes(v1)
		opts.GroupHandler.RegisterRoutes(v1)
		opts.MembershipHandler.RegisterRoutes(v1)
		opts.ContributionHandler.RegisterRoutes(v1)
		if opts.TrackingHandler != nil {
			opts.TrackingHandler.RegisterRoutes(v1)
		}
		opts.ProfileHandler.RegisterRoutes(v1)
		opts.PayoutHandler.RegisterRoutes(v1)
		opts.KittyCycleHandler.RegisterRoutes(v1)
		if opts.ScheduleInvitationHandler != nil {
			opts.ScheduleInvitationHandler.RegisterRoutes(v1)
		}
		if opts.DateProposalHandler != nil {
			opts.DateProposalHandler.RegisterRoutes(v1)
		}
	}

	return engine
}
