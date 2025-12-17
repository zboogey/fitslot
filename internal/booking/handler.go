package booking

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"fitslot/internal/auth"
	"fitslot/internal/gym"
	"fitslot/internal/subscription"
	"fitslot/internal/wallet"
	"fitslot/internal/email" 
	"fitslot/internal/user" 
	"fitslot/internal/logger"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

var (
	ErrSlotNotFound  = errors.New("time slot not found")
	ErrSlotInPast    = errors.New("cannot book a slot in the past")
	ErrSlotFull      = errors.New("time slot is full")
	ErrAlreadyBooked = errors.New("user already has a booking for this slot")
)

type Handler struct {
	repo             *Repository
	gymRepo          *gym.Repository
	subscriptionRepo *subscription.Repository
	walletRepo       *wallet.Repository
	userRepo         *user.Repository
	email            *email.Service
}

func NewHandler(db *sqlx.DB, emailService *email.Service) *Handler {
	return &Handler{
		repo:             NewRepository(db),
		gymRepo:          gym.NewRepository(db),
		subscriptionRepo: subscription.NewRepository(db),
		walletRepo:       wallet.NewRepository(db),
		userRepo:         user.NewRepository(db),
		email:            emailService,
	}
}

func (h *Handler) BookSlot(c *gin.Context) {
	userID, exists := auth.GetUserID(c)
	if !exists {
		logger.Error("Unauthorized booking attempt")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	slotIDStr := c.Param("slotID")
	slotID, err := strconv.Atoi(slotIDStr)
	if err != nil {
		logger.Errorf("Invalid slot ID: %s", slotIDStr)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid slot ID"})
		return
	}
	logger.Infof("User %d booking slot %d", userID, slotID)

	slot, err := h.gymRepo.GetTimeSlotByID(slotID)
	if err != nil {
		logger.Errorf("Slot %d not found: %v", slotID, err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Time slot not found"})
		return
	}

	if slot.StartTime.Before(time.Now()) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot book a slot in the past"})
		return
	}

	bookedCount, err := h.repo.CountActiveBookingsForSlot(slotID)
	if err != nil {
		logger.Errorf("Database error checking bookings for slot %d: %v", slotID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	if bookedCount >= slot.Capacity {
		c.JSON(http.StatusConflict, gin.H{"error": "Time slot is full"})
		return
	}

	hasBooking, err := h.repo.UserHasBookingForSlot(userID, slotID)
	if err != nil {
		logger.Errorf("Database error checking user booking: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	if hasBooking {
		c.JSON(http.StatusConflict, gin.H{"error": "You already have a booking for this slot"})
		return
	}

	ctx := c.Request.Context()

	useSubscription := false
	var activeSub *subscription.Subscription

	sub, err := h.subscriptionRepo.GetActiveForUserAndGym(ctx, userID, slot.GymID)
	if err == nil && sub.Status == subscription.StatusActive {
		if sub.VisitsLimit == nil {
			useSubscription = true
			activeSub = sub
		} else if sub.VisitsUsed < *sub.VisitsLimit {
			useSubscription = true
			activeSub = sub
		}
	}

	booking, err := h.repo.CreateBooking(userID, slotID)
	if err != nil {
		logger.Errorf("Failed to create booking for user %d, slot %d: %v", userID, slotID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create booking"})
		return
	}

	logger.Infof("Booking created: ID=%d, User=%d, Slot=%d", booking.ID, userID, slotID)

	if useSubscription && activeSub != nil {
		if err := h.subscriptionRepo.IncrementVisits(ctx, activeSub.ID); err != nil {
			// Бронь уже есть, поэтому просто вернём warning
			logger.Errorf("Failed to increment subscription visits: %v", err)
			c.JSON(http.StatusCreated, gin.H{
				"booking":      booking,
				"paid_with":    "subscription",
				"subscription": activeSub,
				"warning":      "booking created, but failed to update subscription usage",
			})
			return
		}
		logger.Infof("Booking %d paid with subscription %d", booking.ID, activeSub.ID)
		c.JSON(http.StatusCreated, gin.H{
			"booking":      booking,
			"paid_with":    "subscription",
			"subscription": activeSub,
		})
		return
	}
	
	const priceCents int64 = 1000

	if err := h.walletRepo.AddTransaction(ctx, userID, -priceCents, "booking_payment"); err != nil {
		logger.Errorf("Wallet transaction failed for user %d: %v", userID, err)
		if err.Error() == "insufficient balance" {
			c.JSON(http.StatusPaymentRequired, gin.H{"error": "insufficient wallet balance"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to charge wallet"})
		return
	}

	logger.Infof("Booking %d paid with wallet: %d cents", booking.ID, priceCents)
	c.JSON(http.StatusCreated, gin.H{
		"booking":      booking,
		"paid_with":    "wallet",
		"amount_cents": priceCents,
	})
}

func (h *Handler) CancelBooking(c *gin.Context) {
	userID, exists := auth.GetUserID(c)
	if !exists {
		logger.Error("Unauthorized cancellation attempt")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	bookingIDStr := c.Param("bookingID")
	bookingID, err := strconv.Atoi(bookingIDStr)
	if err != nil {
		logger.Errorf("Invalid booking ID: %s", bookingIDStr)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid booking ID"})
		return
	}

	logger.Infof("User %d cancelling booking %d", userID, bookingID)
	booking, err := h.repo.GetBookingByID(bookingID)
	if err != nil {
		logger.Errorf("Booking %d not found: %v", bookingID, err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Booking not found"})
		return
	}

	if booking.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "You can only cancel your own bookings"})
		return
	}

	err = h.repo.CancelBooking(bookingID)
	if err != nil {
		logger.Errorf("Failed to cancel booking %d: %v", bookingID, err)
		if err == ErrBookingNotFoundOrAlreadyCancelled {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Booking not found or already cancelled"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cancel booking"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Booking cancelled successfully"})
}

func (h *Handler) ListMyBookings(c *gin.Context) {
	userID, exists := auth.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	bookings, err := h.repo.GetUserBookings(userID)
	if err != nil {
		logger.Errorf("Failed to fetch bookings for user %d: %v", userID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch bookings"})
		return
	}

	c.JSON(http.StatusOK, bookings)
}

func (h *Handler) ListBookingsBySlot(c *gin.Context) {
	slotIDStr := c.Param("slotID")
	slotID, err := strconv.Atoi(slotIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid slot ID"})
		return
	}

	bookings, err := h.repo.GetBookingsByTimeSlot(slotID)
	if err != nil {
		logger.Errorf("Failed to fetch bookings for slot %d: %v", slotID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch bookings"})
		return
	}

	c.JSON(http.StatusOK, bookings)
}

func (h *Handler) ListBookingsByGym(c *gin.Context) {
	gymIDStr := c.Param("gymID")
	gymID, err := strconv.Atoi(gymIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid gym ID"})
		return
	}

	bookings, err := h.repo.GetBookingsByGym(gymID)
	if err != nil {
		logger.Errorf("Failed to fetch bookings for gym %d: %v", gymID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch bookings"})
		return
	}

	c.JSON(http.StatusOK, bookings)
}
