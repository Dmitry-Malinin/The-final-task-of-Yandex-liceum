package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCalculateHandler(t *testing.T) {
	type want struct {
		statusCode int
		response   Response
	}
	tests := []struct {
		name   string
		method string
		body   Request
		want   want
	}{
		{
			name:   "valid expression",
			method: http.MethodPost,
			body:   Request{Expression: "2 + 3 * 4"},
			want: want{
				statusCode: http.StatusOK,
				response:   Response{Result: "14.000000"},
			},
		},
		{
			name:   "invalid expression (invalid character)",
			method: http.MethodPost,
			body:   Request{Expression: "2 + 3 * 4 "},
			want: want{
				statusCode: http.StatusUnprocessableEntity,
				response:   Response{Error: "Expression is not valid"},
			},
		},
		{
			name:   "invalid expression (mismatched parentheses)",
			method: http.MethodPost,
			body:   Request{Expression: "(2 + 3 * 4"},
			want: want{
				statusCode: http.StatusUnprocessableEntity,
				response:   Response{Error: "Expression is not valid"},
			},
		},
		{
			name:   "invalid method",
			method: http.MethodGet,
			body:   Request{Expression: "2 + 3 * 4"},
			want: want{
				statusCode: http.StatusMethodNotAllowed,
			},
		},
		{
			name:   "empty expression",
			method: http.MethodPost,
			body:   Request{Expression: ""},
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqBody, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(tt.method, "/api/v1/calculate", bytes.NewBuffer(reqBody))
			w := httptest.NewRecorder()

			calculateHandler(w, req)

			resp := Response{}
			err := json.NewDecoder(w.Body).Decode(&resp)
			if err != nil {
				t.Fatalf("Failed to decode response: %v", err)
			}

			if w.Code != tt.want.statusCode {
				t.Errorf("Expected status code %d, got %d", tt.want.statusCode, w.Code)
			}

			if !compareResponses(resp, tt.want.response) {
				t.Errorf("Expected response %+v, got %+v", tt.want.response, resp)
			}
		})
	}
}
func compareResponses(a, b Response) bool {
	if a.Result != b.Result || a.Error != b.Error {
		return false
	}
	return true
}