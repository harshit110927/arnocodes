package skeleton

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
)

type EndpointSpec struct {
	Name           string `json:"name"`
	Method         string `json:"method"`
	Path           string `json:"path"`
	ExpectedStatus int    `json:"expected_status"`
	Body           string `json:"body,omitempty"`
}

type EndpointCheckResult struct {
	Name           string `json:"name"`
	Method         string `json:"method"`
	Path           string `json:"path"`
	ExpectedStatus int    `json:"expected_status"`
	ActualStatus   int    `json:"actual_status"`
	Passed         bool   `json:"passed"`
}

type SmokeCheckReport struct {
	Total   int                   `json:"total"`
	Passed  int                   `json:"passed"`
	Failed  int                   `json:"failed"`
	Results []EndpointCheckResult `json:"results"`
}

func APICatalog() []EndpointSpec {
	return []EndpointSpec{
		{Name: "Health", Method: http.MethodGet, Path: "/api/v1/health", ExpectedStatus: http.StatusOK},
		{Name: "Get profile", Method: http.MethodGet, Path: "/api/v1/profiles/me", ExpectedStatus: http.StatusOK},
		{Name: "Update profile", Method: http.MethodPatch, Path: "/api/v1/profiles/me", ExpectedStatus: http.StatusAccepted},
		{Name: "List platforms", Method: http.MethodGet, Path: "/api/v1/profiles/me/platform-connections", ExpectedStatus: http.StatusOK},
		{Name: "Connect platform", Method: http.MethodPost, Path: "/api/v1/profiles/me/platform-connections", ExpectedStatus: http.StatusAccepted},
		{Name: "Disconnect platform", Method: http.MethodDelete, Path: "/api/v1/profiles/me/platform-connections/leetcode", ExpectedStatus: http.StatusAccepted},
		{Name: "Dashboard summary", Method: http.MethodGet, Path: "/api/v1/dashboard/summary", ExpectedStatus: http.StatusOK},
		{Name: "Dashboard heatmap", Method: http.MethodGet, Path: "/api/v1/dashboard/heatmap?from=2026-01-01&to=2026-01-31", ExpectedStatus: http.StatusOK},
		{Name: "Dashboard leaderboard", Method: http.MethodGet, Path: "/api/v1/dashboard/leaderboards?scope=global&window=weekly", ExpectedStatus: http.StatusOK},
		{Name: "Course structure", Method: http.MethodGet, Path: "/api/v1/course/structure", ExpectedStatus: http.StatusOK},
		{Name: "Topics", Method: http.MethodGet, Path: "/api/v1/topics", ExpectedStatus: http.StatusOK},
		{Name: "Topic by id", Method: http.MethodGet, Path: "/api/v1/topics/topic-1", ExpectedStatus: http.StatusOK},
		{Name: "Topic unlock status", Method: http.MethodGet, Path: "/api/v1/topics/topic-1/unlock-status", ExpectedStatus: http.StatusOK},
		{Name: "Subtopic by id", Method: http.MethodGet, Path: "/api/v1/subtopics/subtopic-1", ExpectedStatus: http.StatusOK},
		{Name: "Complete learning question", Method: http.MethodPost, Path: "/api/v1/learning/questions/question-1/complete", ExpectedStatus: http.StatusAccepted},
		{Name: "Complete subtopic", Method: http.MethodPost, Path: "/api/v1/subtopics/subtopic-1/complete", ExpectedStatus: http.StatusAccepted, Body: `{"mastery_score":0.9}`},
		{Name: "Test by id", Method: http.MethodGet, Path: "/api/v1/tests/test-1", ExpectedStatus: http.StatusOK},
		{Name: "Start test", Method: http.MethodPost, Path: "/api/v1/tests/test-1/start", ExpectedStatus: http.StatusAccepted},
		{Name: "Test session", Method: http.MethodGet, Path: "/api/v1/tests/test-1/session", ExpectedStatus: http.StatusOK},
		{Name: "Submit answer", Method: http.MethodPost, Path: "/api/v1/test-attempts/attempt-1/answers", ExpectedStatus: http.StatusAccepted},
		{Name: "Submit attempt", Method: http.MethodPost, Path: "/api/v1/test-attempts/attempt-1/submit", ExpectedStatus: http.StatusAccepted},
		{Name: "Attempt result", Method: http.MethodGet, Path: "/api/v1/test-attempts/attempt-1/result", ExpectedStatus: http.StatusOK},
		{Name: "Next question", Method: http.MethodGet, Path: "/api/v1/test-attempts/attempt-1/next-question", ExpectedStatus: http.StatusOK},
		{Name: "Expire attempt", Method: http.MethodPost, Path: "/api/v1/test-attempts/attempt-1/expire", ExpectedStatus: http.StatusAccepted},
		{Name: "Resume attempt", Method: http.MethodPost, Path: "/api/v1/test-attempts/attempt-1/resume", ExpectedStatus: http.StatusAccepted},
		{Name: "Trigger platform sync", Method: http.MethodPost, Path: "/api/v1/platform-sync/trigger", ExpectedStatus: http.StatusAccepted},
		{Name: "Platform sync job", Method: http.MethodGet, Path: "/api/v1/platform-sync/jobs/job-1", ExpectedStatus: http.StatusOK},
		{Name: "AI query", Method: http.MethodPost, Path: "/api/v1/ai/query", ExpectedStatus: http.StatusAccepted},
		{Name: "AI code helper", Method: http.MethodPost, Path: "/api/v1/ai/code-helper/step", ExpectedStatus: http.StatusAccepted},
		{Name: "AI usage", Method: http.MethodGet, Path: "/api/v1/ai/usage", ExpectedStatus: http.StatusOK},
		{Name: "Internal dashboard recompute", Method: http.MethodPost, Path: "/api/v1/internal/recompute-dashboard", ExpectedStatus: http.StatusAccepted},
		{Name: "Internal leaderboard refresh", Method: http.MethodPost, Path: "/api/v1/internal/refresh-leaderboard", ExpectedStatus: http.StatusAccepted},
	}
}

func RunSmokeCheck(handler http.Handler) SmokeCheckReport {
	specs := APICatalog()
	results := make([]EndpointCheckResult, 0, len(specs))
	passed := 0

	for _, spec := range specs {
		body := bytes.NewBuffer(nil)
		if spec.Body != "" {
			body = bytes.NewBufferString(spec.Body)
		}
		req := httptest.NewRequest(spec.Method, spec.Path, body)
		if spec.Body != "" {
			req.Header.Set("Content-Type", "application/json")
		}
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		ok := rr.Code == spec.ExpectedStatus
		if ok {
			passed++
		}

		results = append(results, EndpointCheckResult{
			Name:           spec.Name,
			Method:         spec.Method,
			Path:           spec.Path,
			ExpectedStatus: spec.ExpectedStatus,
			ActualStatus:   rr.Code,
			Passed:         ok,
		})
	}

	return SmokeCheckReport{
		Total:   len(specs),
		Passed:  passed,
		Failed:  len(specs) - passed,
		Results: results,
	}
}

func (r SmokeCheckReport) JSON() ([]byte, error) {
	return json.MarshalIndent(r, "", "  ")
}
