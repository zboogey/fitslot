package gym

import (
	"net/http"
	"strconv"
	"strings"

	"fitslot/internal/api"

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

// @Summary      Create a gym
// @Description  Admin-only: create a new gym
// @Tags         admin,gyms
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body gym.CreateGymRequest true "Gym payload"
// @Success      201 {object} gym.Gym
// @Failure      400 {object} api.ErrorResponse
// @Failure      401 {object} api.ErrorResponse
// @Failure      403 {object} api.ErrorResponse
// @Failure      500 {object} api.ErrorResponse
// @Router       /admin/gyms [post]
func (h *Handler) CreateGym(c *gin.Context) {
	var req CreateGymRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, api.ErrorResponse{Error: err.Error()})
		return
	}

	ctx := c.Request.Context()
	gym, err := h.service.CreateGym(ctx, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, api.ErrorResponse{Error: "Failed to create gym"})
		return
	}

	c.JSON(http.StatusCreated, gym)
}

// @Summary      List gyms
// @Tags         gyms,admin
// @Produce      json
// @Security     BearerAuth
// @Success      200 {array} gym.Gym
// @Failure      401 {object} api.ErrorResponse
// @Failure      500 {object} api.ErrorResponse
// @Router       /gyms [get]
// @Router       /admin/gyms [get]
func (h *Handler) ListGyms(c *gin.Context) {
	ctx := c.Request.Context()
	gyms, err := h.service.GetAllGyms(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, api.ErrorResponse{Error: "Failed to fetch gyms"})
		return
	}

	c.JSON(http.StatusOK, gyms)
}

// @Summary      Create a time slot
// @Description  Admin-only: create a time slot for a gym
// @Tags         admin,gyms
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        gymID path int true "Gym ID"
// @Param        request body gym.CreateTimeSlotRequest true "Time slot payload"
// @Success      201 {object} gym.TimeSlot
// @Failure      400 {object} api.ErrorResponse
// @Failure      401 {object} api.ErrorResponse
// @Failure      403 {object} api.ErrorResponse
// @Failure      404 {object} api.ErrorResponse
// @Failure      500 {object} api.ErrorResponse
// @Router       /admin/gyms/{gymID}/slots [post]
func (h *Handler) CreateTimeSlot(c *gin.Context) {
	gymIDStr := c.Param("gymID")
	gymID, err := strconv.Atoi(gymIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.ErrorResponse{Error: "Invalid gym ID"})
		return
	}

	var req CreateTimeSlotRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, api.ErrorResponse{Error: err.Error()})
		return
	}

	ctx := c.Request.Context()
	slot, err := h.service.CreateTimeSlot(ctx, gymID, req)
	if err != nil {
		switch err {
		case ErrGymNotFound:
			c.JSON(http.StatusNotFound, api.ErrorResponse{Error: "Gym not found"})
		case ErrTimeSlotInvalid:
			c.JSON(http.StatusBadRequest, api.ErrorResponse{Error: "Invalid time slot data"})
		default:
			c.JSON(http.StatusInternalServerError, api.ErrorResponse{Error: "Failed to create time slot"})
		}
		return
	}

	c.JSON(http.StatusCreated, slot)
}

// @Summary      List time slots for a gym
// @Tags         gyms,admin
// @Produce      json
// @Security     BearerAuth
// @Param        gymID path int true "Gym ID"
// @Success      200 {array} gym.TimeSlotWithAvailability
// @Failure      400 {object} api.ErrorResponse
// @Failure      401 {object} api.ErrorResponse
// @Failure      404 {object} api.ErrorResponse
// @Failure      500 {object} api.ErrorResponse
// @Router       /gyms/{gymID}/slots [get]
// @Router       /admin/gyms/{gymID}/slots [get]
func (h *Handler) ListTimeSlots(c *gin.Context) {
	gymIDStr := c.Param("gymID")
	gymID, err := strconv.Atoi(gymIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.ErrorResponse{Error: "Invalid gym ID"})
		return
	}

	ctx := c.Request.Context()
	onlyFuture := !strings.Contains(c.Request.URL.Path, "/admin/")
	slots, err := h.service.GetTimeSlots(ctx, gymID, onlyFuture)
	if err != nil {
		switch err {
		case ErrGymNotFound:
			c.JSON(http.StatusNotFound, api.ErrorResponse{Error: "Gym not found"})
		default:
			c.JSON(http.StatusInternalServerError, api.ErrorResponse{Error: "Failed to fetch time slots"})
		}
		return
	}

	c.JSON(http.StatusOK, slots)
}
