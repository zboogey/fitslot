package gym

import (
    "net/http"
    "strconv"
    "strings"
    "time"

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

// CreateGym godoc
// @Summary      Create gym
// @Description  Creates a new gym. Admin only.
// @Tags         gyms
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request  body      CreateGymRequest  true  "Gym data"
// @Success      201      {object}  Gym
// @Failure      400      {object}  gin.H
// @Failure      500      {object}  gin.H
// @Router       /admin/gyms [post]
func (h *Handler) CreateGym(c *gin.Context) {
    var req CreateGymRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    gym, err := h.repo.CreateGym(req.Name, req.Location)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create gym"})
        return
    }

    c.JSON(http.StatusCreated, gym)
}

// ListGyms godoc
// @Summary      List gyms
// @Description  Returns list of gyms.
// @Tags         gyms
// @Security     BearerAuth
// @Produce      json
// @Success      200  {array}   Gym
// @Failure      500  {object}  gin.H
// @Router       /gyms [get]
func (h *Handler) ListGyms(c *gin.Context) {
    gyms, err := h.repo.GetAllGyms()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch gyms"})
        return
    }

    c.JSON(http.StatusOK, gyms)
}

// CreateTimeSlot godoc
// @Summary      Create time slot
// @Description  Creates a new time slot for a gym. Admin only.
// @Tags         gyms
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        gymID    path      int                   true  "Gym ID"
// @Param        request  body      CreateTimeSlotRequest true  "Time slot data"
// @Success      201      {object}  TimeSlot
// @Failure      400      {object}  gin.H
// @Failure      404      {object}  gin.H
// @Failure      500      {object}  gin.H
// @Router       /admin/gyms/{gymID}/slots [post]
func (h *Handler) CreateTimeSlot(c *gin.Context) {
    gymIDStr := c.Param("gymID")
    gymID, err := strconv.Atoi(gymIDStr)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid gym ID"})
        return
    }

    _, err = h.repo.GetGymByID(gymID)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Gym not found"})
        return
    }

    var req CreateTimeSlotRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    startTime, err := time.Parse(time.RFC3339, req.StartTime)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start_time format. Use RFC3339 (e.g., 2024-01-15T10:00:00Z)"})
        return
    }

    endTime, err := time.Parse(time.RFC3339, req.EndTime)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end_time format. Use RFC3339 (e.g., 2024-01-15T11:00:00Z)"})
        return
    }

    if endTime.Before(startTime) || endTime.Equal(startTime) {
        c.JSON(http.StatusBadRequest, gin.H{"error": "end_time must be after start_time"})
        return
    }

    slot, err := h.repo.CreateTimeSlot(gymID, startTime, endTime, req.Capacity)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create time slot"})
        return
    }

    c.JSON(http.StatusCreated, slot)
}

// ListTimeSlots godoc
// @Summary      List time slots
// @Description  Returns time slots for a gym with availability info.
// @Tags         gyms
// @Security     BearerAuth
// @Produce      json
// @Param        gymID  path      int  true  "Gym ID"
// @Success      200    {array}   TimeSlotWithAvailability
// @Failure      400    {object}  gin.H
// @Failure      404    {object}  gin.H
// @Failure      500    {object}  gin.H
// @Router       /gyms/{gymID}/slots [get]
func (h *Handler) ListTimeSlots(c *gin.Context) {
    gymIDStr := c.Param("gymID")
    gymID, err := strconv.Atoi(gymIDStr)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid gym ID"})
        return
    }

    _, err = h.repo.GetGymByID(gymID)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Gym not found"})
        return
    }

    onlyFuture := !strings.Contains(c.Request.URL.Path, "/admin/")
    slots, err := h.repo.GetTimeSlotsWithAvailability(gymID, onlyFuture)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch time slots"})
        return
    }

    c.JSON(http.StatusOK, slots)
}
