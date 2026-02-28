package handlers

import "net/http"

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	h.router = mux
	mux.HandleFunc("/health", h.HealthHandler)
	mux.HandleFunc(apiV1BasePath+"/health", h.HealthHandler)

	protect := func(fn http.HandlerFunc) http.HandlerFunc {
		handler := http.Handler(fn)
		if h.authMiddleware != nil {
			handler = h.authMiddleware.Middleware(handler)
		}
		return func(w http.ResponseWriter, r *http.Request) {
			handler.ServeHTTP(w, r)
		}
	}

	mux.HandleFunc(apiV1BasePath+"/profiles/me", protect(h.ProfileMeHandler))
	mux.HandleFunc(apiV1BasePath+"/profiles/me/status", protect(h.ProfileStatusHandler))
	mux.HandleFunc(apiV1BasePath+"/profiles/me/platform-connections", protect(h.PlatformConnectionsHandler))
	mux.HandleFunc(apiV1BasePath+"/profiles/me/platform-connections/", protect(h.PlatformConnectionByNameHandler))

	mux.HandleFunc(apiV1BasePath+"/dashboard", protect(h.DashboardSummaryHandler))
	mux.HandleFunc(apiV1BasePath+"/dashboard/summary", protect(h.DashboardSummaryHandler))
	mux.HandleFunc(apiV1BasePath+"/dashboard/heatmap", protect(h.DashboardHeatmapHandler))
	mux.HandleFunc(apiV1BasePath+"/dashboard/leaderboards", protect(h.DashboardLeaderboardsHandler))

	mux.HandleFunc(apiV1BasePath+"/course", protect(h.CourseRouter))
	mux.HandleFunc(apiV1BasePath+"/course/", protect(h.CourseRouter))
	mux.HandleFunc(apiV1BasePath+"/course/structure", protect(h.CourseStructureHandler))
	mux.HandleFunc(apiV1BasePath+"/topics", protect(h.TopicsRouter))
	mux.HandleFunc(apiV1BasePath+"/topics/", protect(h.TopicsRouter))
	mux.HandleFunc(apiV1BasePath+"/subtopics/", protect(h.SubtopicsRouter))
	mux.HandleFunc(apiV1BasePath+"/learning/questions/", protect(h.LearningQuestionsRouter))

	mux.HandleFunc(apiV1BasePath+"/diagnostic/start", protect(h.DiagnosticRouter))
	mux.HandleFunc(apiV1BasePath+"/diagnostic/", protect(h.DiagnosticRouter))

	mux.HandleFunc(apiV1BasePath+"/platform-sync/trigger", protect(h.PlatformSyncTriggerHandler))
	mux.HandleFunc(apiV1BasePath+"/platform-sync/overview", protect(h.PlatformSyncOverviewHandler))
	mux.HandleFunc(apiV1BasePath+"/platform-sync/jobs/", protect(h.PlatformSyncJobHandler))
	mux.HandleFunc(apiV1BasePath+"/ide/submit", protect(h.IDESubmitHandler))
	mux.HandleFunc(apiV1BasePath+"/ide/status", protect(h.IDEStatusHandler))
	mux.HandleFunc(apiV1BasePath+"/ide/run", protect(h.IDERunHandler))

	mux.HandleFunc(apiV1BasePath+"/ai/query", protect(h.AIQueryHandler))
	mux.HandleFunc(apiV1BasePath+"/ai/code-helper/step", protect(h.AICodeHelperStepHandler))
	mux.HandleFunc(apiV1BasePath+"/ai/usage", protect(h.AIUsageHandler))

	mux.HandleFunc(apiV1BasePath+"/internal/recompute-dashboard", protect(h.InternalRecomputeDashboardHandler))
	mux.HandleFunc(apiV1BasePath+"/internal/refresh-leaderboard", protect(h.InternalRefreshLeaderboardHandler))
	mux.HandleFunc(apiV1BasePath+"/internal/api-catalog", protect(h.APICatalogHandler))
	mux.HandleFunc(apiV1BasePath+"/internal/api-smoke-check", protect(h.APISmokeCheckHandler))
}
