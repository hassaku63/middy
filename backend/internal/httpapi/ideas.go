package httpapi

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/hassaku63/middy/backend/internal/ideas"
)

// maxIdeasBody fits the largest conversation the contract allows:
// MaxMessages messages of MaxMessageRunes characters each, every one a
// character outside the BMP written as an escaped surrogate pair (12
// bytes, "\ud83c\udfb5"), plus room for the roles and the JSON around
// them (#499). Over it the client gets a 413.
const maxIdeasBody = ideas.MaxMessages*ideas.MaxMessageRunes*12 + 16<<10

// Request shape for POST /api/v1/ideas (docs/api/openapi.yaml).
type ideasRequest struct {
	Messages []ideas.Message `json:"messages"`
}

// handleIdeas runs one conversational turn of scenario S2 and returns the
// reply with rendered candidates. Provider trouble is a 502, not a 500:
// the server is fine, the upstream is not. A safety refusal is a 422.
// Upstream failures are logged with their cause, so a deployment can
// tell a timeout from a rejected key; the client only gets the code. A
// client that went away is neither: nothing is written to it and nothing
// is logged above DEBUG (#499).
func handleIdeas(svc *ideas.Service, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req ideasRequest
		if err := decodeJSONUpTo(r, &req, maxIdeasBody); err != nil {
			writeDecodeError(w, err)
			return
		}
		res, err := svc.Imagine(r.Context(), req.Messages)
		if err != nil && r.Context().Err() != nil {
			logger.DebugContext(r.Context(), "ideas: client went away", "error", err)
			return
		}
		switch {
		case err == nil:
			writeJSON(w, res)
		case errors.Is(err, ideas.ErrLLMUnavailable):
			logger.WarnContext(r.Context(), "ideas: upstream failed", "code", "llm_unavailable", "error", err)
			writeError(w, http.StatusBadGateway, "llm_unavailable", "the language model could not be reached; try again")
		case errors.Is(err, ideas.ErrInvalidOutput):
			logger.WarnContext(r.Context(), "ideas: upstream failed", "code", "llm_invalid_output", "error", err)
			writeError(w, http.StatusBadGateway, "llm_invalid_output", "the language model gave an unusable answer twice; try rephrasing")
		case errors.Is(err, ideas.ErrDeclined):
			// Not an upstream fault: the request was understood and refused.
			writeError(w, http.StatusUnprocessableEntity, "llm_declined", "the language model declined this request; try rephrasing")
		case errors.Is(err, ideas.ErrInvalidConversation):
			writeError(w, http.StatusBadRequest, "invalid_request", err.Error())
		default:
			// An answer that passed validation but could not be built is
			// the engine's fault, not the caller's.
			logger.ErrorContext(r.Context(), "ideas: could not build the examples", "error", err)
			writeError(w, http.StatusInternalServerError, "internal_error", "the examples could not be built; try again")
		}
	}
}
