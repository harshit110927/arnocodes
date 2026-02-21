package handlers

import "net/http"

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	h.router = mux
	mux.HandleFunc("/health", h.HealthHandler)
	mux.HandleFunc(apiV1BasePath+"/health", h.HealthHandler)

	mux.HandleFunc(apiV1BasePath+"/profiles/me", h.ProfileMeHandler)
	mux.HandleFunc(apiV1BasePath+"/profiles/me/status", h.ProfileStatusHandler)
	mux.HandleFunc(apiV1BasePath+"/profiles/me/platform-connections", h.PlatformConnectionsHandler)
	mux.HandleFunc(apiV1BasePath+"/profiles/me/platform-connections/", h.PlatformConnectionByNameHandler)

	mux.HandleFunc(apiV1BasePath+"/dashboard", h.DashboardSummaryHandler)
	mux.HandleFunc(apiV1BasePath+"/dashboard/summary", h.DashboardSummaryHandler)
	mux.HandleFunc(apiV1BasePath+"/dashboard/heatmap", h.DashboardHeatmapHandler)
	mux.HandleFunc(apiV1BasePath+"/dashboard/leaderboards", h.DashboardLeaderboardsHandler)

	mux.HandleFunc(apiV1BasePath+"/course", h.CourseRouter)
	mux.HandleFunc(apiV1BasePath+"/course/", h.CourseRouter)
	mux.HandleFunc(apiV1BasePath+"/course/structure", h.CourseStructureHandler)
	mux.HandleFunc(apiV1BasePath+"/topics", h.TopicsRouter)
	mux.HandleFunc(apiV1BasePath+"/topics/", h.TopicsRouter)
	mux.HandleFunc(apiV1BasePath+"/subtopics/", h.SubtopicsRouter)
	mux.HandleFunc(apiV1BasePath+"/learning/questions/", h.LearningQuestionsRouter)

	mux.HandleFunc(apiV1BasePath+"/diagnostic/start", h.DiagnosticRouter)
	mux.HandleFunc(apiV1BasePath+"/diagnostic/", h.DiagnosticRouter)

	mux.HandleFunc(apiV1BasePath+"/platform-sync/trigger", h.PlatformSyncTriggerHandler)
	mux.HandleFunc(apiV1BasePath+"/platform-sync/jobs/", h.PlatformSyncJobHandler)
	mux.HandleFunc(apiV1BasePath+"/ide/submit", h.IDESubmitHandler)
	mux.HandleFunc(apiV1BasePath+"/ide/status", h.IDEStatusHandler)
	mux.HandleFunc(apiV1BasePath+"/ide/run", h.IDERunHandler)

	mux.HandleFunc(apiV1BasePath+"/ai/query", h.AIQueryHandler)
	mux.HandleFunc(apiV1BasePath+"/ai/code-helper/step", h.AICodeHelperStepHandler)
	mux.HandleFunc(apiV1BasePath+"/ai/usage", h.AIUsageHandler)

	mux.HandleFunc(apiV1BasePath+"/internal/recompute-dashboard", h.InternalRecomputeDashboardHandler)
	mux.HandleFunc(apiV1BasePath+"/internal/refresh-leaderboard", h.InternalRefreshLeaderboardHandler)
	mux.HandleFunc(apiV1BasePath+"/internal/api-catalog", h.APICatalogHandler)
	mux.HandleFunc(apiV1BasePath+"/internal/api-smoke-check", h.APISmokeCheckHandler)
}
