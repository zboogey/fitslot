package booking

import (
	"net/http"
	"strconv"

	"fitslot/internal/api"
	"fitslot/internal/auth"
	"fitslot/internal/logger"
	"fitslot/internal/metrics"
	"fitslot/internal/subscription"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{
		service: service,
	}
}

// @Summary      Book a time slot
// @Description  Create a booking for the current user (paid with wallet or subscription)
// @Tags         bookings
// @Produce      json
// @Security     BearerAuth
// @Param        slotID path int true "Time slot ID"
// @Success      201 {object} BookSlotResponse
// @Failure      400 {object} api.ErrorResponse
// @Failure      401 {object} api.ErrorResponse
// @Failure      402 {object} api.ErrorResponse
// @Failure      404 {object} api.ErrorResponse
// @Failure      409 {object} api.ErrorResponse
// @Failure      500 {object} api.ErrorResponse
// @Router       /slots/{slotID}/book [post]
func (h *Handler) BookSlot(c *gin.Context) {
	userID, exists := auth.GetUserID(c)
	if !exists {
		logger.Error("Unauthorized booking attempt")
		c.JSON(http.StatusUnauthorized, api.ErrorResponse{Error: "User not authenticated"})
		return
	}

	slotIDStr := c.Param("slotID")
	slotID, err := strconv.Atoi(slotIDStr)
	if err != nil {
		logger.Errorf("Invalid slot ID: %s", slotIDStr)
		c.JSON(http.StatusBadRequest, api.ErrorResponse{Error: "Invalid slot ID"})
		return
	}
	logger.Infof("User %d booking slot %d", userID, slotID)

	ctx := c.Request.Context()
	booking, paymentMethod, paymentDetails, err := h.service.BookSlot(ctx, userID, slotID)
	if err != nil {
		switch err.Error() {
		case "time slot not found":
			c.JSON(http.StatusNotFound, api.ErrorResponse{Error: "Time slot not found"})
		case "cannot book a slot in the past":
			c.JSON(http.StatusBadRequest, api.ErrorResponse{Error: "Cannot book a slot in the past"})
		case "time slot is full":
			c.JSON(http.StatusConflict, api.ErrorResponse{Error: "Time slot is full"})
		case "user already has a booking for this slot":
			c.JSON(http.StatusConflict, api.ErrorResponse{Error: "You already have a booking for this slot"})
		case "insufficient wallet balance":
			c.JSON(http.StatusPaymentRequired, api.ErrorResponse{Error: "insufficient wallet balance"})
		default:
			logger.Errorf("Failed to create booking for user %d, slot %d: %v", userID, slotID, err)
			c.JSON(http.StatusInternalServerError, api.ErrorResponse{Error: "Failed to create booking"})
		}
		return
	}

	logger.Infof("Booking created: ID=%d, User=%d, Slot=%d", booking.ID, userID, slotID)
	metrics.RecordBooking("success", paymentMethod)

	var resp BookSlotResponse
	resp.Booking = booking
	resp.PaidWith = paymentMethod

	if paymentMethod == "subscription" {
		if sub, ok := paymentDetails.(*subscription.Subscription); ok {
			resp.Subscription = sub
		}
	} else if paymentMethod == "wallet" {
		if details, ok := paymentDetails.(map[string]interface{}); ok {
			switch v := details["amount_cents"].(type) {
			case int64:
				resp.AmountCents = &v
			case int:
				vv := int64(v)
				resp.AmountCents = &vv
			case float64:
				vv := int64(v)
				resp.AmountCents = &vv
			}
		}
	}

	c.JSON(http.StatusCreated, resp)
}

// @Summary      Cancel booking
// @Description  Cancel a booking owned by the current user
// @Tags         bookings
// @Produce      json
// @Security     BearerAuth
// @Param        bookingID path int true "Booking ID"
// @Success      200 {object} booking.CancelBookingResponse
// @Failure      400 {object} api.ErrorResponse
// @Failure      401 {object} api.ErrorResponse
// @Failure      403 {object} api.ErrorResponse
// @Failure      404 {object} api.ErrorResponse
// @Failure      500 {object} api.ErrorResponse
// @Router       /bookings/{bookingID}/cancel [post]
func (h *Handler) CancelBooking(c *gin.Context) {
	userID, exists := auth.GetUserID(c)
	if !exists {
		logger.Error("Unauthorized cancellation attempt")
		c.JSON(http.StatusUnauthorized, api.ErrorResponse{Error: "User not authenticated"})
		return
	}

	bookingIDStr := c.Param("bookingID")
	bookingID, err := strconv.Atoi(bookingIDStr)
	if err != nil {
		logger.Errorf("Invalid booking ID: %s", bookingIDStr)
		c.JSON(http.StatusBadRequest, api.ErrorResponse{Error: "Invalid booking ID"})
		return
	}

	logger.Infof("User %d cancelling booking %d", userID, bookingID)
	ctx := c.Request.Context()
	err = h.service.CancelBooking(ctx, userID, bookingID)
	if err != nil {
		logger.Errorf("Failed to cancel booking %d: %v", bookingID, err)
		switch err.Error() {
		case "booking not found":
			c.JSON(http.StatusNotFound, api.ErrorResponse{Error: "Booking not found"})
		case "unauthorized: can only cancel own bookings":
			c.JSON(http.StatusForbidden, api.ErrorResponse{Error: "You can only cancel your own bookings"})
		default:
			c.JSON(http.StatusInternalServerError, api.ErrorResponse{Error: "Failed to cancel booking"})
		}
		return
	}

	c.JSON(http.StatusOK, CancelBookingResponse{Message: "Booking cancelled successfully"})
	metrics.RecordBookingCancellation()
}

// @Summary      List my bookings
// @Tags         bookings
// @Produce      json
// @Security     BearerAuth
// @Success      200 {array} booking.Booking
// @Failure      401 {object} api.ErrorResponse
// @Failure      500 {object} api.ErrorResponse
// @Router       /bookings [get]
func (h *Handler) ListMyBookings(c *gin.Context) {
	userID, exists := auth.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, api.ErrorResponse{Error: "User not authenticated"})
		return
	}

	ctx := c.Request.Context()
	bookings, err := h.service.GetUserBookings(ctx, userID)
	if err != nil {
		logger.Errorf("Failed to fetch bookings for user %d: %v", userID, err)
		c.JSON(http.StatusInternalServerError, api.ErrorResponse{Error: "Failed to fetch bookings"})
		return
	}

	c.JSON(http.StatusOK, bookings)
}

// @Summary      List bookings by time slot (admin)
// @Tags         admin,bookings
// @Produce      json
// @Security     BearerAuth
// @Param        slotID path int true "Time slot ID"
// @Success      200 {array} booking.BookingWithDetails
// @Failure      400 {object} api.ErrorResponse
// @Failure      401 {object} api.ErrorResponse
// @Failure      403 {object} api.ErrorResponse
// @Failure      500 {object} api.ErrorResponse
// @Router       /admin/slots/{slotID}/bookings [get]
func (h *Handler) ListBookingsBySlot(c *gin.Context) {
	slotIDStr := c.Param("slotID")
	slotID, err := strconv.Atoi(slotIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.ErrorResponse{Error: "Invalid slot ID"})
		return
	}

	ctx := c.Request.Context()
	bookings, err := h.service.GetBookingsByTimeSlot(ctx, slotID)
	if err != nil {
		logger.Errorf("Failed to fetch bookings for slot %d: %v", slotID, err)
		c.JSON(http.StatusInternalServerError, api.ErrorResponse{Error: "Failed to fetch bookings"})
		return
	}

	c.JSON(http.StatusOK, bookings)
}

// @Summary      List bookings by gym (admin)
// @Tags         admin,bookings
// @Produce      json
// @Security     BearerAuth
// @Param        gymID path int true "Gym ID"
// @Success      200 {array} booking.BookingWithDetails
// @Failure      400 {object} api.ErrorResponse
// @Failure      401 {object} api.ErrorResponse
// @Failure      403 {object} api.ErrorResponse
// @Failure      500 {object} api.ErrorResponse
// @Router       /admin/gyms/{gymID}/bookings [get]
func (h *Handler) ListBookingsByGym(c *gin.Context) {
	gymIDStr := c.Param("gymID")
	gymID, err := strconv.Atoi(gymIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.ErrorResponse{Error: "Invalid gym ID"})
		return
	}

	ctx := c.Request.Context()
	bookings, err := h.service.GetBookingsByGym(ctx, gymID)
	if err != nil {
		logger.Errorf("Failed to fetch bookings for gym %d: %v", gymID, err)
		c.JSON(http.StatusInternalServerError, api.ErrorResponse{Error: "Failed to fetch bookings"})
		return
	}

	c.JSON(http.StatusOK, bookings)
}
