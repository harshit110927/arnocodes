package skeleton

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
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
		{Name: "Test by id", Method: http.MethodGet, Path: "/api/v1/tests/diagnostic-1?topics=arrays,strings", ExpectedStatus: http.StatusOK},
		{Name: "Start test", Method: http.MethodPost, Path: "/api/v1/tests/diagnostic-1/start", ExpectedStatus: http.StatusAccepted, Body: `{"topics_known":["arrays","strings"]}`},
		{Name: "Test session", Method: http.MethodGet, Path: "/api/v1/tests/diagnostic-1/session?attempt_id={attempt_id}", ExpectedStatus: http.StatusOK},
		{Name: "Next question", Method: http.MethodGet, Path: "/api/v1/test-attempts/{attempt_id}/next-question", ExpectedStatus: http.StatusOK},
		{Name: "Submit answer", Method: http.MethodPost, Path: "/api/v1/test-attempts/{attempt_id}/answers", ExpectedStatus: http.StatusAccepted, Body: `{"question_id":"{question_id}","selected_option":2}`},
		{Name: "Submit attempt", Method: http.MethodPost, Path: "/api/v1/test-attempts/{attempt_id}/submit", ExpectedStatus: http.StatusAccepted},
		{Name: "Attempt result", Method: http.MethodGet, Path: "/api/v1/test-attempts/{attempt_id}/result", ExpectedStatus: http.StatusOK},
		{Name: "Expire attempt", Method: http.MethodPost, Path: "/api/v1/test-attempts/{attempt_id}/expire", ExpectedStatus: http.StatusConflict},
		{Name: "Resume attempt", Method: http.MethodPost, Path: "/api/v1/test-attempts/{attempt_id}/resume", ExpectedStatus: http.StatusConflict},
		{Name: "Trigger platform sync", Method: http.MethodPost, Path: "/api/v1/platform-sync/trigger", ExpectedStatus: http.StatusAccepted},
		{Name: "Platform sync job", Method: http.MethodGet, Path: "/api/v1/platform-sync/jobs/job-1", ExpectedStatus: http.StatusOK},
		{Name: "AI query", Method: http.MethodPost, Path: "/api/v1/ai/query", ExpectedStatus: http.StatusAccepted},
		{Name: "AI code helper", Method: http.MethodPost, Path: "/api/v1/ai/code-helper/step", ExpectedStatus: http.StatusAccepted},
		{Name: "AI usage", Method: http.MethodGet, Path: "/api/v1/ai/usage", ExpectedStatus: http.StatusOK},
		{Name: "Internal dashboard recompute", Method: http.MethodPost, Path: "/api/v1/internal/recompute-dashboard", ExpectedStatus: http.StatusAccepted},
		{Name: "Internal leaderboard refresh", Method: http.MethodPost, Path: "/api/v1/internal/refresh-leaderboard", ExpectedStatus: http.StatusAccepted},
	}
}

func replaceVars(s, attemptID, questionID string) string {
	s = strings.ReplaceAll(s, "{attempt_id}", attemptID)
	s = strings.ReplaceAll(s, "{question_id}", questionID)
	return s
}

func RunSmokeCheck(handler http.Handler) SmokeCheckReport {
	specs := APICatalog()
	results := make([]EndpointCheckResult, 0, len(specs))
	passed := 0
	attemptID := ""
	questionID := "q-1"

	for _, spec := range specs {
		path := replaceVars(spec.Path, attemptID, questionID)
		bodyText := replaceVars(spec.Body, attemptID, questionID)
		body := bytes.NewBuffer(nil)
		if bodyText != "" {
			body = bytes.NewBufferString(bodyText)
		}
		req := httptest.NewRequest(spec.Method, path, body)
		req.Header.Set("X-User-ID", "smoke-user")
		if bodyText != "" {
			req.Header.Set("Content-Type", "application/json")
		}
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		if spec.Name == "Start test" && rr.Code == http.StatusAccepted {
			var payload struct {
				Data map[string]interface{} `json:"data"`
			}
			_ = json.NewDecoder(rr.Body).Decode(&payload)
			if v, ok := payload.Data["attempt_id"].(string); ok {
				attemptID = v
			}
		}
		if spec.Name == "Next question" && rr.Code == http.StatusOK {
			var payload struct {
				Data struct {
					Question struct {
						ID string `json:"id"`
					} `json:"question"`
				} `json:"data"`
			}
			_ = json.NewDecoder(rr.Body).Decode(&payload)
			if payload.Data.Question.ID != "" {
				questionID = payload.Data.Question.ID
			}
		}

		ok := rr.Code == spec.ExpectedStatus
		if ok {
			passed++
		}

		results = append(results, EndpointCheckResult{
			Name:           spec.Name,
			Method:         spec.Method,
			Path:           path,
			ExpectedStatus: spec.ExpectedStatus,
			ActualStatus:   rr.Code,
			Passed:         ok,
		})
	}

	return SmokeCheckReport{Total: len(specs), Passed: passed, Failed: len(specs) - passed, Results: results}
}

func (r SmokeCheckReport) JSON() ([]byte, error) {
	return json.MarshalIndent(r, "", "  ")
}
