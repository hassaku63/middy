// Command converse measures what survives a conversation (#453): it sends
// the same turns to a running server as the app does, keeps every answer,
// and reports, turn by turn, each example's length in bars, whether it
// keeps sus4, and how many of its chords sound a tension next to whether
// its words claim one. It is a measuring tool for the owner's machine,
// not part of CI; with LLM_PROVIDER=fake it runs without a key.
//
//	go run ./cmd/converse collect -base http://localhost:8080 -turns turns.json -runs 3 -out out/
//	go run ./cmd/converse report out/*.json
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/hassaku63/middy/backend/internal/ideas"
	"github.com/hassaku63/middy/backend/internal/music"
	"github.com/hassaku63/middy/backend/internal/music/pattern"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: converse collect|report ...")
		os.Exit(2)
	}
	var err error
	switch os.Args[1] {
	case "collect":
		err = collect(os.Args[2:])
	case "report":
		err = report(os.Stdout, os.Args[2:])
	default:
		err = fmt.Errorf("unknown command %q", os.Args[1])
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Turn is one exchange as collected: what was sent and what came back.
type Turn struct {
	Sent   string        `json:"sent"`
	Answer *ideas.Result `json:"answer,omitempty"`
	Error  string        `json:"error,omitempty"`
}

func collect(args []string) error {
	fs := flag.NewFlagSet("collect", flag.ContinueOnError)
	base := fs.String("base", "http://localhost:8080", "the running server")
	turnsPath := fs.String("turns", "", "JSON array of the user's messages, in order")
	runs := fs.Int("runs", 3, "how many times to send the whole conversation")
	out := fs.String("out", "converse-out", "directory for one JSON file per run")
	if err := fs.Parse(args); err != nil {
		return err
	}
	data, err := os.ReadFile(*turnsPath)
	if err != nil {
		return err
	}
	var turns []string
	if err := json.Unmarshal(data, &turns); err != nil {
		return fmt.Errorf("%s: %w", *turnsPath, err)
	}
	if err := os.MkdirAll(*out, 0o750); err != nil {
		return err
	}
	client := &http.Client{Timeout: 5 * time.Minute}
	for r := 1; r <= *runs; r++ {
		got := conversation(client, *base, turns)
		b, err := json.MarshalIndent(got, "", "  ")
		if err != nil {
			return err
		}
		path := filepath.Join(*out, fmt.Sprintf("run-%d.json", r))
		if err := os.WriteFile(path, b, 0o600); err != nil {
			return err
		}
		fmt.Fprintln(os.Stderr, "wrote", path)
	}
	return nil
}

// conversation sends the turns one by one with the history the app sends:
// the user's words, and for each answer its reply and the gist of its
// examples (frontend/src/features/chat/history.ts). Keep the two in step.
func conversation(client *http.Client, base string, turns []string) []Turn {
	var history []message
	var out []Turn
	for _, text := range turns {
		history = append(history, message{Role: "user", Content: text})
		res, err := imagine(client, base, history)
		t := Turn{Sent: text}
		if err != nil {
			t.Error = err.Error()
			out = append(out, t)
			history = history[:len(history)-1] // the app keeps no answer it did not get
			continue
		}
		t.Answer = res
		out = append(out, t)
		history = append(history, message{Role: "assistant", Content: answerMessage(res)})
	}
	return out
}

// answerMessage is an answer as the next request carries it, as the app's
// historyMessage writes it (BL-310): the reply, then one line per example.
func answerMessage(res *ideas.Result) string {
	if len(res.Candidates) == 0 {
		return res.Reply
	}
	var b strings.Builder
	b.WriteString(res.Reply + "\n前回の候補:")
	for _, c := range res.Candidates {
		b.WriteString("\n" + gist(c))
	}
	return string([]rune(b.String())[:min(len([]rune(b.String())), 1000)])
}

// gist is one example in a line, as candidateGist writes it in the app.
func gist(c ideas.Candidate) string {
	meter := c.Meter
	if meter == "" {
		meter = "4/4"
	}
	parts := []string{c.Key.String(), meter, strconv.FormatFloat(c.TempoBPM, 'f', -1, 64) + " BPM"}
	if c.Bars > 0 {
		parts = append(parts, strconv.Itoa(c.Bars)+" bars")
	}
	chords := make([]string, len(c.Progression))
	for i, l := range c.Progression {
		beats := c.BeatsPerChord
		if i < len(c.Beats) {
			beats = c.Beats[i]
		}
		chords[i] = l + ":" + strconv.Itoa(beats)
	}
	parts = append(parts, strings.Join(chords, " "))
	for _, kv := range [][2]string{{"style", c.Style}, {"accompaniment", c.Accompaniment}, {"drums", c.Drums}, {"feel", c.Feel}, {"tensions", c.Tensions}} {
		if kv[1] != "" {
			parts = append(parts, kv[0]+" "+kv[1])
		}
	}
	return "- " + c.Title + ": " + strings.Join(parts, ", ")
}

func imagine(client *http.Client, base string, history []message) (*ideas.Result, error) {
	body, err := json.Marshal(map[string]any{"messages": history})
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, strings.TrimSuffix(base, "/")+"/api/v1/ideas", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close() //nolint:errcheck // read-only body
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status %d: %s", resp.StatusCode, strings.TrimSpace(string(data)))
	}
	var res ideas.Result
	if err := json.Unmarshal(data, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

func report(w io.Writer, paths []string) error {
	if len(paths) == 0 {
		return errors.New("report: give the collected run files")
	}
	var b strings.Builder
	b.WriteString("| run | turn | example | meter | bars | sus4 | tension sounded | words claim a tension |\n")
	b.WriteString("| --- | --- | --- | --- | --- | --- | --- | --- |\n")
	for _, path := range paths {
		data, err := os.ReadFile(path) //nolint:gosec // files the owner names on the command line
		if err != nil {
			return err
		}
		var turns []Turn
		if err := json.Unmarshal(data, &turns); err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}
		run := strings.TrimSuffix(filepath.Base(path), ".json")
		for i, t := range turns {
			if t.Answer == nil {
				fmt.Fprintf(&b, "| %s | %d | (error: %s) | | | | | |\n", run, i+1, t.Error)
				continue
			}
			for _, c := range t.Answer.Candidates {
				m := measure(c, t.Answer.Reply)
				fmt.Fprintf(&b, "| %s | %d | %s | %s | %s | %s | %d / %d | %s |\n",
					run, i+1, c.Title, m.Meter, m.Bars, yes(m.Sus4), m.Tensioned, m.Chords, yes(m.Claims))
			}
		}
	}
	_, err := io.WriteString(w, b.String())
	return err
}

// Measure is what the report says about one example.
type Measure struct {
	Meter     string
	Bars      string // whole bars, or a fraction when it ends mid-bar
	Sus4      bool
	Chords    int
	Tensioned int  // chords under which a non-chord tone sounds on a pitched track
	Claims    bool // the reason or the reply names a 9th, 13th or tension
}

func measure(c ideas.Candidate, reply string) Measure {
	tl := c.Timeline
	ts := tl.TimeSignature
	bar := music.BarTicks(ts, tl.PPQ)
	var end music.Tick
	for _, h := range tl.Harmony {
		end = max(end, h.Start+h.Duration)
	}
	m := Measure{Meter: fmt.Sprintf("%d/%d", ts.Numerator, ts.Denominator), Chords: len(tl.Harmony)}
	if bar > 0 {
		if end%bar == 0 {
			m.Bars = strconv.Itoa(int(end / bar))
		} else {
			m.Bars = fmt.Sprintf("%.2f", float64(end)/float64(bar))
		}
	}
	for _, label := range c.Progression {
		if strings.Contains(label, "sus4") {
			m.Sus4 = true
		}
	}
	for _, h := range tl.Harmony {
		if sounds(tl, h) {
			m.Tensioned++
		}
	}
	words := c.Reason + " " + reply
	for _, w := range []string{"9 度", "9度", "13 度", "13度", "テンション", "add9", "9th"} {
		if strings.Contains(words, w) {
			m.Claims = true
		}
	}
	return m
}

// sounds reports a note on a pitched track, inside the chord's span, whose
// pitch class is neither a chord tone nor the slash bass: a tension. A
// note running past the chord's end is the next chord pushed early, not
// this chord's colour.
func sounds(tl music.Timeline, h music.HarmonyEvent) bool {
	tones := map[music.PitchClass]bool{}
	for _, pc := range h.Chord.PitchClasses() {
		tones[pc.Normalize()] = true
	}
	if h.Bass != nil {
		tones[h.Bass.Normalize()] = true
	}
	for _, tr := range tl.Tracks {
		if tr.Channel == pattern.DrumChannel {
			continue
		}
		for _, n := range tr.Notes {
			if n.Start >= h.Start && n.Start+n.Duration <= h.Start+h.Duration && !tones[music.PitchClass(n.Pitch%12).Normalize()] {
				return true
			}
		}
	}
	return false
}

func yes(b bool) string {
	if b {
		return "yes"
	}
	return ""
}
