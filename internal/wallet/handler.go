package wallet

import (
	"net/http"
	"strconv"

	"fitslot/internal/api"
	"fitslot/internal/auth"
	"fitslot/internal/metrics"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	repo Repository
}

func NewHandler(repo Repository) *Handler {
	return &Handler{
		repo: repo,
	}
}

type TopUpRequest struct {
	AmountCents int64 `json:"amount_cents" binding:"required"`
}

// @Summary      Get wallet balance
// @Tags         wallet
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} wallet.Wallet
// @Failure      401 {object} api.ErrorResponse
// @Failure      500 {object} api.ErrorResponse
// @Router       /wallet [get]
func (h *Handler) GetBalance(c *gin.Context) {
	userID, ok := auth.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, api.ErrorResponse{Error: "user not authenticated"})
		return
	}

	w, err := h.repo.GetOrCreateWallet(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, api.ErrorResponse{Error: "failed to load wallet"})
		return
	}

	c.JSON(http.StatusOK, w)
}

// @Summary      Top up wallet
// @Tags         wallet
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body wallet.TopUpRequest true "Top up payload"
// @Success      200 {object} wallet.TopUpResponse
// @Failure      400 {object} api.ErrorResponse
// @Failure      401 {object} api.ErrorResponse
// @Failure      500 {object} api.ErrorResponse
// @Router       /wallet/topup [post]
func (h *Handler) TopUp(c *gin.Context) {
	userID, ok := auth.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, api.ErrorResponse{Error: "user not authenticated"})
		return
	}

	var req TopUpRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.AmountCents <= 0 {
		c.JSON(http.StatusBadRequest, api.ErrorResponse{Error: "amount_cents must be positive"})
		return
	}

	if err := h.repo.TopUp(c.Request.Context(), userID, req.AmountCents); err != nil {
		c.JSON(http.StatusInternalServerError, api.ErrorResponse{Error: "failed to top up wallet"})
		return
	}

	metrics.RecordWalletTopUp()
	w, err := h.repo.GetOrCreateWallet(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, api.ErrorResponse{Error: "failed to load wallet after top up"})
		return
	}

	c.JSON(http.StatusOK, TopUpResponse{
		Message: "wallet recharged",
		Wallet:  *w,
	})
}

// @Summary      List wallet transactions
// @Tags         wallet
// @Produce      json
// @Security     BearerAuth
// @Param        limit  query int false "Limit" default(50)
// @Param        offset query int false "Offset" default(0)
// @Success      200 {array} wallet.Transaction
// @Failure      401 {object} api.ErrorResponse
// @Failure      500 {object} api.ErrorResponse
// @Router       /wallet/transactions [get]
func (h *Handler) ListTransactions(c *gin.Context) {
	userID, ok := auth.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, api.ErrorResponse{Error: "user not authenticated"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	txs, err := h.repo.GetTransactions(c.Request.Context(), userID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, api.ErrorResponse{Error: "failed to load transactions"})
		return
	}

	c.JSON(http.StatusOK, txs)
}
