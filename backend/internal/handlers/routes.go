package handlers

import "net/http"

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/health", h.HealthHandler)

	mux.HandleFunc("/profiles/me", h.ProfileMeHandler)
	mux.HandleFunc("/profiles/me/platform-connections", h.PlatformConnectionsHandler)
	mux.HandleFunc("/profiles/me/platform-connections/", h.PlatformConnectionByNameHandler)

	mux.HandleFunc("/dashboard/summary", h.DashboardSummaryHandler)
	mux.HandleFunc("/dashboard/heatmap", h.DashboardHeatmapHandler)
	mux.HandleFunc("/dashboard/leaderboards", h.DashboardLeaderboardsHandler)

	mux.HandleFunc("/topics", h.TopicsRouter)
	mux.HandleFunc("/topics/", h.TopicsRouter)
	mux.HandleFunc("/subtopics/", h.SubtopicsRouter)
	mux.HandleFunc("/learning/questions/", h.LearningQuestionsRouter)

	mux.HandleFunc("/tests/", h.TestsRouter)
	mux.HandleFunc("/test-attempts/", h.TestAttemptsRouter)

	mux.HandleFunc("/platform-sync/trigger", h.PlatformSyncTriggerHandler)
	mux.HandleFunc("/platform-sync/jobs/", h.PlatformSyncJobHandler)

	mux.HandleFunc("/ai/query", h.AIQueryHandler)
	mux.HandleFunc("/ai/code-helper/step", h.AICodeHelperStepHandler)
	mux.HandleFunc("/ai/usage", h.AIUsageHandler)

	mux.HandleFunc("/internal/recompute-dashboard", h.InternalRecomputeDashboardHandler)
	mux.HandleFunc("/internal/refresh-leaderboard", h.InternalRefreshLeaderboardHandler)
}
