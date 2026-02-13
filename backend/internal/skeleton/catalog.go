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
		{Name: "Diagnostic start requires user", Method: http.MethodPost, Path: "/api/v1/diagnostic/start", ExpectedStatus: http.StatusUnauthorized, Body: `{"selected_topics":["22222222-2222-2222-2222-222222222221"]}`},
		{Name: "Protected dashboard requires user", Method: http.MethodGet, Path: "/api/v1/dashboard/summary", ExpectedStatus: http.StatusUnauthorized},
		{Name: "Internal catalog", Method: http.MethodGet, Path: "/api/v1/internal/api-catalog", ExpectedStatus: http.StatusOK},
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
		results = append(results, EndpointCheckResult{Name: spec.Name, Method: spec.Method, Path: spec.Path, ExpectedStatus: spec.ExpectedStatus, ActualStatus: rr.Code, Passed: ok})
	}
	return SmokeCheckReport{Total: len(specs), Passed: passed, Failed: len(specs) - passed, Results: results}
}

func (r SmokeCheckReport) JSON() ([]byte, error) {
	return json.MarshalIndent(r, "", "  ")
}
