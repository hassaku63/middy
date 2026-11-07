package httpapi

import (
	"net/http"
	"strconv"

	"github.com/hassaku63/middy/backend/internal/music"
	"github.com/hassaku63/middy/backend/internal/music/smf"
)

// Request shape for POST /api/v1/export/midi (docs/api/openapi.yaml).
type exportMIDIRequest struct {
	Timeline music.Timeline `json:"timeline"`
}

// handleExportMIDI returns the timeline as a Standard MIDI File. The file
// name is fixed; the client chooses a descriptive one for the download.
func handleExportMIDI(w http.ResponseWriter, r *http.Request) {
	var req exportMIDIRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}
	if err := req.Timeline.Validate(); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	data, err := smf.Write(req.Timeline)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	w.Header().Set("Content-Type", "audio/midi")
	w.Header().Set("Content-Disposition", `attachment; filename="middy.mid"`)
	w.Header().Set("Content-Length", strconv.Itoa(len(data)))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}
