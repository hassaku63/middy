package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
)

// problemTypeBase is where each problem type is documented; a type URI is
// this plus the type's code (ADR-0030).
const problemTypeBase = "https://github.com/hassaku63/middy/blob/main/docs/api/problems.md#"

// problemTitles are the problem types the API answers with, by code, and
// their fixed titles. docs/api/problems.md documents every one of them (a
// test checks), together with the one nginx answers on its own
// (api_unavailable).
var problemTitles = map[string]string{
	"invalid_json":       "Invalid JSON",
	"invalid_request":    "Invalid request",
	"not_found":          "Not found",
	"method_not_allowed": "Method not allowed",
	"request_too_large":  "Request too large",
	"llm_declined":       "Model declined",
	"internal_error":     "Internal error",
	"llm_unavailable":    "Model unavailable",
	"llm_invalid_output": "Model output invalid",
}

// problem is the error body every endpoint uses: an RFC 9457 problem
// detail, with the type's code as an extension member (ADR-0030).
type problem struct {
	Type   string `json:"type"`
	Title  string `json:"title"`
	Status int    `json:"status"`
	Detail string `json:"detail"`
	Code   string `json:"code"`
}

// writeJSON answers 200 with v as JSON.
func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	writeBody(w, http.StatusOK, v)
}

// writeBody encodes v before anything goes out (#499). A value
// encoding/json cannot write (a NaN, say) is a programming error: it
// panics, so requestLogger logs it and answers 500, instead of a 200
// with a cut-off body and no trace.
func writeBody(w http.ResponseWriter, status int, v any) {
	b, err := json.Marshal(v)
	if err != nil {
		panic(fmt.Errorf("httpapi: encode the response: %w", err))
	}
	b = append(b, '\n')
	w.Header().Set("Content-Length", strconv.Itoa(len(b)))
	w.WriteHeader(status)
	_, _ = w.Write(b)
}

// writeError answers with the problem detail of code; detail says what
// went wrong this time, for developers.
func writeError(w http.ResponseWriter, status int, code, detail string) {
	w.Header().Set("Content-Type", "application/problem+json")
	writeBody(w, status, problem{
		Type:   problemTypeBase + code,
		Title:  problemTitles[code],
		Status: status,
		Detail: detail,
		Code:   code,
	})
}

// writeDecodeError answers a body that could not be read: 413 when it is
// over the endpoint's limit, else 400 invalid_json.
func writeDecodeError(w http.ResponseWriter, err error) {
	if errors.Is(err, errBodyTooLarge) {
		writeError(w, http.StatusRequestEntityTooLarge, "request_too_large", err.Error())
		return
	}
	writeError(w, http.StatusBadRequest, "invalid_json", err.Error())
}
