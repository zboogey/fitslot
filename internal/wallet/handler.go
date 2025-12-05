package wallet

import (
	"net/http"
	"strconv"

	"fitslot/internal/auth"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

type Handler struct {
	repo *Repository
}

func NewHandler(db *sqlx.DB) *Handler {
	return &Handler{
		repo: NewRepository(db),
	}
}

type TopUpRequest struct {
	AmountCents int64 `json:"amount_cents" binding:"required"`
}

func (h *Handler) GetBalance(c *gin.Context) {
	userID, ok := auth.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	w, err := h.repo.GetOrCreateWallet(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load wallet"})
		return
	}

	c.JSON(http.StatusOK, w)
}

func (h *Handler) TopUp(c *gin.Context) {
	userID, ok := auth.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	var req TopUpRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.AmountCents <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "amount_cents must be positive"})
		return
	}

	if err := h.repo.TopUp(c.Request.Context(), userID, req.AmountCents); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to top up wallet"})
		return
	}

	w, err := h.repo.GetOrCreateWallet(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load wallet after top up"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "wallet recharged",
		"wallet":  w,
	})
}

func (h *Handler) ListTransactions(c *gin.Context) {
	userID, ok := auth.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	txs, err := h.repo.GetTransactions(c.Request.Context(), userID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load transactions"})
		return
	}

	c.JSON(http.StatusOK, txs)
}