package subscription

import (
	"errors"
	"net/http"

	"fitslot/internal/auth"
	"fitslot/internal/wallet"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"fitslot/internal/metrics"
	"fitslot/internal/logger"
)

type Handler struct {
	repo       *Repository
	walletRepo *wallet.Repository
}

func NewHandler(db *sqlx.DB) *Handler {
	return &Handler{
		repo:       NewRepository(db),
		walletRepo: wallet.NewRepository(db),
	}
}

type Plan struct {
	Type        string `json:"type"`
	Name        string `json:"name"`
	Description string `json:"description"`
	PriceCents  int64  `json:"price_cents"`
	VisitsLimit *int   `json:"visits_limit,omitempty"`
	GymRequired bool   `json:"gym_required"`
}

func getPlans() []Plan {
	singleLimit := 8
	multiLimit := 20

	return []Plan{
		{
			Type:        "single_gym_lite",
			Name:        "Single Gym Lite",
			Description: "1 выбранный зал, 8 посещений в месяц",
			PriceCents:  10000,
			VisitsLimit: &singleLimit,
			GymRequired: true,
		},
		{
			Type:        "multi_gym_flex",
			Name:        "Multi Gym Flex",
			Description: "Любые залы, 20 посещений в месяц",
			PriceCents:  18000,
			VisitsLimit: &multiLimit,
			GymRequired: false,
		},
		{
			Type:        "unlimited_pro",
			Name:        "Unlimited Pro",
			Description: "Любые залы, безлимит на месяц",
			PriceCents:  25000,
			VisitsLimit: nil,
			GymRequired: false,
		},
	}
}

func findPlan(planType string) (Plan, error) {
	for _, p := range getPlans() {
		if p.Type == planType {
			return p, nil
		}
	}
	return Plan{}, errors.New("unknown plan type")
}

type CreateSubscriptionRequest struct {
	Type  string `json:"type" binding:"required"`
	GymID *int   `json:"gym_id,omitempty"`
}

type CreateSubscriptionResponse struct {
	Subscription *Subscription `json:"subscription"`
	PaidWith     string        `json:"paid_with"`
	AmountCents  int64         `json:"amount_cents"`
}

func (h *Handler) Create(c *gin.Context) {
	userID, ok := auth.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req CreateSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	plan, err := findPlan(req.Type)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unknown subscription type"})
		return
	}

	if plan.GymRequired && req.GymID == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "gym_id is required for single_gym_lite"})
		return
	}
	if !plan.GymRequired {
		req.GymID = nil
	}

	ctx := c.Request.Context()

	if err := h.walletRepo.AddTransaction(ctx, userID, -plan.PriceCents, "subscription_payment"); err != nil {
		if err.Error() == "insufficient balance" {
			c.JSON(http.StatusPaymentRequired, gin.H{"error": "insufficient wallet balance"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to charge wallet"})
		return
	}

	subType := SubscriptionType(plan.Type)
	sub, err := h.repo.CreateSubscription(ctx, userID, req.GymID, subType, plan.PriceCents, plan.VisitsLimit)
	if err != nil {
		logger.Errorf("Failed to create subscription for user %d: %v", userID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create subscription"})
		return
	}
	logger.Infof("Subscription created: Type=%s, User=%d", plan.Type, userID)
	metrics.RecordSubscription(plan.Type)

	c.JSON(http.StatusCreated, CreateSubscriptionResponse{
		Subscription: sub,
		PaidWith:     "wallet",
		AmountCents:  plan.PriceCents,
	})
}

func (h *Handler) ListMy(c *gin.Context) {
	userID, ok := auth.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	subs, err := h.repo.ListActiveByUser(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load subscriptions"})
		return
	}

	c.JSON(http.StatusOK, subs)
}

func (h *Handler) ListPlans(c *gin.Context) {
	c.JSON(http.StatusOK, getPlans())
}
