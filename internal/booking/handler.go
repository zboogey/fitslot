package booking

import (
    "errors"
    "net/http"
    "strconv"
    "time"

    "fitslot/internal/auth"
    "fitslot/internal/gym"

    "github.com/gin-gonic/gin"
    "github.com/jmoiron/sqlx"
)

var (
    ErrSlotNotFound             = errors.New("time slot not found")
    ErrSlotInPast               = errors.New("cannot book a slot in the past")
    ErrSlotFull                 = errors.New("time slot is full")
    ErrAlreadyBooked            = errors.New("user already has a booking for this slot")
    
)

type Handler struct {
    repo    *Repository
    gymRepo *gym.Repository
}

func NewHandler(db *sqlx.DB) *Handler {
    return &Handler{
        repo:    NewRepository(db),
        gymRepo: gym.NewRepository(db),
    }
}

// BookSlot godoc
// @Summary      Book time slot
// @Description  Creates a booking for the given time slot.
// @Tags         bookings
// @Security     BearerAuth
// @Produce      json
// @Param        slotID  path      int  true  "Time slot ID"
// @Success      201     {object}  Booking
// @Failure      400     {object}  gin.H
// @Failure      404     {object}  gin.H
// @Failure      409     {object}  gin.H
// @Failure      500     {object}  gin.H
// @Router       /slots/{slotID}/book [post]
func (h *Handler) BookSlot(c *gin.Context) {
    userID, exists := auth.GetUserID(c)
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
        return
    }

    slotIDStr := c.Param("slotID")
    slotID, err := strconv.Atoi(slotIDStr)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid slot ID"})
        return
    }

    slot, err := h.gymRepo.GetTimeSlotByID(slotID)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Time slot not found"})
        return
    }

    if slot.StartTime.Before(time.Now()) {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot book a slot in the past"})
        return
    }

    bookedCount, err := h.repo.CountActiveBookingsForSlot(slotID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
        return
    }

    if bookedCount >= slot.Capacity {
        c.JSON(http.StatusConflict, gin.H{"error": "Time slot is full"})
        return
    }

    hasBooking, err := h.repo.UserHasBookingForSlot(userID, slotID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
        return
    }

    if hasBooking {
        c.JSON(http.StatusConflict, gin.H{"error": "You already have a booking for this slot"})
        return
    }

    booking, err := h.repo.CreateBooking(userID, slotID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create booking"})
        return
    }

    c.JSON(http.StatusCreated, booking)
}

// CancelBooking godoc
// @Summary      Cancel booking
// @Description  Cancels an existing booking of the current user.
// @Tags         bookings
// @Security     BearerAuth
// @Produce      json
// @Param        bookingID  path      int  true  "Booking ID"
// @Success      200        {object}  gin.H
// @Failure      400        {object}  gin.H
// @Failure      403        {object}  gin.H
// @Failure      404        {object}  gin.H
// @Failure      500        {object}  gin.H
// @Router       /bookings/{bookingID}/cancel [post]
func (h *Handler) CancelBooking(c *gin.Context) {
    userID, exists := auth.GetUserID(c)
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
        return
    }

    bookingIDStr := c.Param("bookingID")
    bookingID, err := strconv.Atoi(bookingIDStr)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid booking ID"})
        return
    }

    booking, err := h.repo.GetBookingByID(bookingID)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Booking not found"})
        return
    }

    if booking.UserID != userID {
        c.JSON(http.StatusForbidden, gin.H{"error": "You can only cancel your own bookings"})
        return
    }

    err = h.repo.CancelBooking(bookingID)
    if err != nil {
        if err == ErrBookingNotFoundOrAlreadyCancelled {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Booking not found or already cancelled"})
            return
        }
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cancel booking"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Booking cancelled successfully"})
}

// ListMyBookings godoc
// @Summary      List my bookings
// @Description  Returns bookings of the authenticated user.
// @Tags         bookings
// @Security     BearerAuth
// @Produce      json
// @Success      200  {array}   Booking
// @Failure      500  {object}  gin.H
// @Router       /bookings [get]
func (h *Handler) ListMyBookings(c *gin.Context) {
    userID, exists := auth.GetUserID(c)
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
        return
    }

    bookings, err := h.repo.GetUserBookings(userID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch bookings"})
        return
    }

    c.JSON(http.StatusOK, bookings)
}

// ListBookingsBySlot godoc
// @Summary      List bookings by slot
// @Description  Returns all bookings for a specific time slot. Admin only.
// @Tags         bookings
// @Security     BearerAuth
// @Produce      json
// @Param        slotID  path      int  true  "Time slot ID"
// @Success      200     {array}   Booking
// @Failure      400     {object}  gin.H
// @Failure      500     {object}  gin.H
// @Router       /admin/slots/{slotID}/bookings [get]
func (h *Handler) ListBookingsBySlot(c *gin.Context) {
    slotIDStr := c.Param("slotID")
    slotID, err := strconv.Atoi(slotIDStr)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid slot ID"})
        return
    }

    bookings, err := h.repo.GetBookingsByTimeSlot(slotID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch bookings"})
        return
    }

    c.JSON(http.StatusOK, bookings)
}

// ListBookingsByGym godoc
// @Summary      List bookings by gym
// @Description  Returns all bookings for a specific gym. Admin only.
// @Tags         bookings
// @Security     BearerAuth
// @Produce      json
// @Param        gymID  path      int  true  "Gym ID"
// @Success      200    {array}   Booking
// @Failure      400    {object}  gin.H
// @Failure      500    {object}  gin.H
// @Router       /admin/gyms/{gymID}/bookings [get]
func (h *Handler) ListBookingsByGym(c *gin.Context) {
    gymIDStr := c.Param("gymID")
    gymID, err := strconv.Atoi(gymIDStr)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid gym ID"})
        return
    }

    bookings, err := h.repo.GetBookingsByGym(gymID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch bookings"})
        return
    }

    c.JSON(http.StatusOK, bookings)
}

// GetBookingAnalytics godoc
// @Summary      Booking analytics
// @Description  Returns aggregated booking analytics. Admin only.
// @Tags         bookings
// @Security     BearerAuth
// @Produce      json
// @Param        group_by  query     string  false  "Group by dimension (day or gym)"
// @Param        from      query     string  true   "Start datetime (RFC3339)"
// @Param        to        query     string  true   "End datetime (RFC3339)"
// @Success      200       {object}  map[string]interface{}
// @Failure      400       {object}  gin.H
// @Failure      500       {object}  gin.H
// @Router       /admin/analytics/bookings [get]
func (h *Handler) GetBookingAnalytics(c *gin.Context) {
    groupBy := c.DefaultQuery("group_by", "day")
    fromStr := c.Query("from")
    toStr := c.Query("to")

    if fromStr == "" || toStr == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "from and to query params are required"})
        return
    }

    from, err := time.Parse(time.RFC3339, fromStr)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid from format, use RFC3339"})
        return
    }

    to, err := time.Parse(time.RFC3339, toStr)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid to format, use RFC3339"})
        return
    }

    switch groupBy {
    case "day":
        stats, err := h.repo.GetBookingStatsByDay(from, to)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch stats"})
            return
        }
        c.JSON(http.StatusOK, gin.H{
            "group_by": "day",
            "from":     from,
            "to":       to,
            "data":     stats,
        })
    case "gym":
        stats, err := h.repo.GetBookingStatsByGym(from, to)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch stats"})
            return
        }
        c.JSON(http.StatusOK, gin.H{
            "group_by": "gym",
            "from":     from,
            "to":       to,
            "data":     stats,
        })
    default:
        c.JSON(http.StatusBadRequest, gin.H{"error": "group_by must be 'day' or 'gym'"})
    }
}
