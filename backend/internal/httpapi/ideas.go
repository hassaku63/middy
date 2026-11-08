package httpapi

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/hassaku63/middy/backend/internal/ideas"
)

// Request shape for POST /api/v1/ideas (docs/api/openapi.yaml).
type ideasRequest struct {
	Messages []ideas.Message `json:"messages"`
}

// handleIdeas runs one conversational turn of scenario S2 and returns the
// reply with rendered candidates. Provider trouble is a 502, not a 500:
// the server is fine, the upstream is not. A safety refusal is a 422.
func handleIdeas(svc *ideas.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req ideasRequest
		if err := decodeJSON(r, &req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid_json", err.Error())
			return
		}
		res, err := svc.Imagine(r.Context(), req.Messages)
		switch {
		case err == nil:
			writeJSON(w, http.StatusOK, res)
		case errors.Is(err, ideas.ErrLLMUnavailable):
			writeError(w, http.StatusBadGateway, "llm_unavailable", "the language model could not be reached; try again")
		case errors.Is(err, ideas.ErrInvalidOutput):
			writeError(w, http.StatusBadGateway, "llm_invalid_output", "the language model gave an unusable answer twice; try rephrasing")
		case errors.Is(err, ideas.ErrDeclined):
			// Not an upstream fault: the request was understood and refused.
			writeError(w, http.StatusUnprocessableEntity, "llm_declined", "the language model declined this request; try rephrasing")
		case errors.Is(err, ideas.ErrInvalidConversation):
			writeError(w, http.StatusBadRequest, "invalid_request", err.Error())
		default:
			// An answer that passed validation but could not be built is
			// the engine's fault, not the caller's.
			slog.ErrorContext(r.Context(), "ideas: could not build the examples", "error", err)
			writeError(w, http.StatusInternalServerError, "internal_error", "the examples could not be built; try again")
		}
	}
}
