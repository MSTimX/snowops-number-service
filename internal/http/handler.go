package http

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	"number-service/internal/service"
)

type Handler struct {
	numberService *service.NumberService
	log           zerolog.Logger
}

func NewHandler(
	numberService *service.NumberService,
	log zerolog.Logger,
) *Handler {
	return &Handler{
		numberService: numberService,
		log:           log,
	}
}

func (h *Handler) Register(r *gin.Engine, authMiddleware gin.HandlerFunc) {
	public := r.Group("/api/v1/numbers")
	{
		public.POST("/check", h.checkNumber)
		public.POST("/whitelist", h.addToWhitelist)
		public.POST("/blacklist", h.addToBlacklist)
		public.DELETE("/whitelist", h.removeFromWhitelist)
		public.DELETE("/blacklist", h.removeFromBlacklist)
	}
}

func (h *Handler) checkNumber(c *gin.Context) {
	var req struct {
		Plate string `json:"plate" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err.Error()))
		return
	}

	result, err := h.numberService.NormalizeAndCheck(c.Request.Context(), req.Plate)
	if err != nil {
		if errors.Is(err, service.ErrInvalidInput) {
			c.JSON(http.StatusBadRequest, errorResponse(err.Error()))
			return
		}
		h.log.Error().Err(err).Msg("failed to check number")
		c.JSON(http.StatusInternalServerError, errorResponse("internal error"))
		return
	}

	c.JSON(http.StatusOK, successResponse(result))
}

func (h *Handler) addToWhitelist(c *gin.Context) {
	var req struct {
		Plate string `json:"plate" binding:"required"`
		Note  string `json:"note,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err.Error()))
		return
	}

	var note *string
	if req.Note != "" {
		note = &req.Note
	}

	if err := h.numberService.AddToWhitelist(c.Request.Context(), req.Plate, note); err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, successResponse(gin.H{"message": "added to whitelist"}))
}

func (h *Handler) addToBlacklist(c *gin.Context) {
	var req struct {
		Plate string `json:"plate" binding:"required"`
		Note  string `json:"note,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err.Error()))
		return
	}

	var note *string
	if req.Note != "" {
		note = &req.Note
	}

	if err := h.numberService.AddToBlacklist(c.Request.Context(), req.Plate, note); err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, successResponse(gin.H{"message": "added to blacklist"}))
}

func (h *Handler) removeFromWhitelist(c *gin.Context) {
	plate := strings.TrimSpace(c.Query("plate"))
	if plate == "" {
		c.JSON(http.StatusBadRequest, errorResponse("plate parameter is required"))
		return
	}

	if err := h.numberService.RemoveFromWhitelist(c.Request.Context(), plate); err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, successResponse(gin.H{"message": "removed from whitelist"}))
}

func (h *Handler) removeFromBlacklist(c *gin.Context) {
	plate := strings.TrimSpace(c.Query("plate"))
	if plate == "" {
		c.JSON(http.StatusBadRequest, errorResponse("plate parameter is required"))
		return
	}

	if err := h.numberService.RemoveFromBlacklist(c.Request.Context(), plate); err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, successResponse(gin.H{"message": "removed from blacklist"}))
}

func (h *Handler) handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidInput):
		c.JSON(http.StatusBadRequest, errorResponse(err.Error()))
	case errors.Is(err, service.ErrNotFound):
		c.JSON(http.StatusNotFound, errorResponse(err.Error()))
	default:
		h.log.Error().Err(err).Msg("handler error")
		c.JSON(http.StatusInternalServerError, errorResponse("internal error"))
	}
}

func successResponse(data interface{}) gin.H {
	return gin.H{
		"data": data,
	}
}

func errorResponse(message string) gin.H {
	return gin.H{
		"error": message,
	}
}

