package handlers

import (
	"net/http"
	"strconv"

	"github.com/controlewise/backend/internal/services"
	"github.com/controlewise/backend/internal/utils"
)

type AdminDashboardHandler struct {
	statsService *services.AdminStatsService
}

func NewAdminDashboardHandler(statsService *services.AdminStatsService) *AdminDashboardHandler {
	return &AdminDashboardHandler{
		statsService: statsService,
	}
}

func (h *AdminDashboardHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.statsService.GetPlatformStats(r.Context())
	if err != nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, "Failed to get platform stats")
		return
	}

	utils.SuccessResponse(w, http.StatusOK, stats)
}

func (h *AdminDashboardHandler) GetRecentActivity(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 50 {
		limit = 10
	}

	activities, err := h.statsService.GetRecentActivity(r.Context(), limit)
	if err != nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, "Failed to get recent activity")
		return
	}

	utils.SuccessResponse(w, http.StatusOK, activities)
}
