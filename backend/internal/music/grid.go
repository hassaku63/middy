package music

// DrumGrid is a drum groove written out on a one-bar grid (ADR-0012
// decision 2, BL-122): each beat is cut into Division steps, and each voice
// is a string with one character per step. "X" is an accent, "x" a normal
// stroke, "o" a ghost note, "." silence; spaces and "|" may separate beats
// and are ignored. The engine validates it, renders it and explains it;
// the voices are a fixed vocabulary (pattern.GridVoices).
type DrumGrid struct {
	Division int               `json:"division"`
	Voices   map[string]string `json:"voices"`
	// BasedOn, when set, is the catalogue drums the grid starts from.
	BasedOn string `json:"basedOn,omitempty"`
}
