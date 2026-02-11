package handlers

import "net/http"

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/health", h.HealthHandler)
	mux.HandleFunc(apiV1BasePath+"/health", h.HealthHandler)

	mux.HandleFunc(apiV1BasePath+"/profiles/me", h.ProfileMeHandler)
	mux.HandleFunc(apiV1BasePath+"/profiles/me/platform-connections", h.PlatformConnectionsHandler)
	mux.HandleFunc(apiV1BasePath+"/profiles/me/platform-connections/", h.PlatformConnectionByNameHandler)

	mux.HandleFunc(apiV1BasePath+"/dashboard/summary", h.DashboardSummaryHandler)
	mux.HandleFunc(apiV1BasePath+"/dashboard/heatmap", h.DashboardHeatmapHandler)
	mux.HandleFunc(apiV1BasePath+"/dashboard/leaderboards", h.DashboardLeaderboardsHandler)

	mux.HandleFunc(apiV1BasePath+"/course/structure", h.CourseStructureHandler)
	mux.HandleFunc(apiV1BasePath+"/topics", h.TopicsRouter)
	mux.HandleFunc(apiV1BasePath+"/topics/", h.TopicsRouter)
	mux.HandleFunc(apiV1BasePath+"/subtopics/", h.SubtopicsRouter)
	mux.HandleFunc(apiV1BasePath+"/learning/questions/", h.LearningQuestionsRouter)

	mux.HandleFunc(apiV1BasePath+"/tests/", h.TestsRouter)
	mux.HandleFunc(apiV1BasePath+"/test-attempts/", h.TestAttemptsRouter)

	mux.HandleFunc(apiV1BasePath+"/platform-sync/trigger", h.PlatformSyncTriggerHandler)
	mux.HandleFunc(apiV1BasePath+"/platform-sync/jobs/", h.PlatformSyncJobHandler)

	mux.HandleFunc(apiV1BasePath+"/ai/query", h.AIQueryHandler)
	mux.HandleFunc(apiV1BasePath+"/ai/code-helper/step", h.AICodeHelperStepHandler)
	mux.HandleFunc(apiV1BasePath+"/ai/usage", h.AIUsageHandler)

	mux.HandleFunc(apiV1BasePath+"/internal/recompute-dashboard", h.InternalRecomputeDashboardHandler)
	mux.HandleFunc(apiV1BasePath+"/internal/refresh-leaderboard", h.InternalRefreshLeaderboardHandler)
}
