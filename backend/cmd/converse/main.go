// Command converse measures what survives a conversation (#453): it sends
// the same turns to a running server as the app does, keeps every answer,
// and reports, turn by turn, each example's length in bars, whether it
// keeps sus4, and how many of its chords sound a tension next to whether
// its words claim one. A script (scenarios/*.json) may also say what each
// turn should change; the report then checks every example against it
// (#473, BL-314). With LLM_PROVIDER=fake it runs without a key.
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
	Expect *Expect       `json:"expect,omitempty"`
	Answer *ideas.Result `json:"answer,omitempty"`
	Error  string        `json:"error,omitempty"`
}

func collect(args []string) error {
	fs := flag.NewFlagSet("collect", flag.ContinueOnError)
	base := fs.String("base", "http://localhost:8080", "the running server")
	turnsPath := fs.String("turns", "", "a script: a JSON array of the user's messages, or {name, turns: [{say, expect}]}")
	runs := fs.Int("runs", 3, "how many times to send the whole conversation")
	out := fs.String("out", "converse-out", "directory for one JSON file per run")
	if err := fs.Parse(args); err != nil {
		return err
	}
	data, err := os.ReadFile(*turnsPath)
	if err != nil {
		return err
	}
	script, err := readScript(data)
	if err != nil {
		return fmt.Errorf("%s: %w", *turnsPath, err)
	}
	if err := os.MkdirAll(*out, 0o750); err != nil {
		return err
	}
	client := &http.Client{Timeout: 5 * time.Minute}
	for r := 1; r <= *runs; r++ {
		got := conversation(client, *base, script.Steps, progress(os.Stderr, r))
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
func conversation(client *http.Client, base string, steps []Step, done func(turn, of int, d time.Duration, err error)) []Turn {
	var history []message
	var out []Turn
	for i, st := range steps {
		history = append(history, message{Role: "user", Content: st.Say})
		start := time.Now()
		res, err := imagine(client, base, history)
		done(i+1, len(steps), time.Since(start), err)
		t := Turn{Sent: st.Say, Expect: st.Expect}
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

// progress writes a line per turn as it comes back, so a slow run on the
// real API (a turn can take a minute or more) is seen to move.
func progress(w io.Writer, run int) func(turn, of int, d time.Duration, err error) {
	return func(turn, of int, d time.Duration, err error) {
		status := "ok"
		if err != nil {
			status = "error: " + err.Error()
		}
		_, _ = fmt.Fprintf(w, "run %d, turn %d/%d: %s (%.0fs)\n", run, turn, of, status, d.Seconds())
	}
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
	runs, expects, err := readRuns(paths)
	if err != nil {
		return err
	}
	var b strings.Builder
	// A script with expectations adds a column; one without keeps the
	// table as it was (BL-314).
	head, rule := "", ""
	if expects {
		head, rule = " expected |", " --- |"
	}
	b.WriteString("| run | turn | example | meter | bars | sus4 | tension sounded | words claim a tension |" + head + "\n")
	b.WriteString("| --- | --- | --- | --- | --- | --- | --- | --- |" + rule + "\n")
	for i, turns := range runs {
		writeRun(&b, runName(paths, i), turns, expects)
	}
	_, err = io.WriteString(w, b.String())
	return err
}

// readRuns reads the collected runs, and whether any turn has expectations.
func readRuns(paths []string) ([][]Turn, bool, error) {
	runs := make([][]Turn, len(paths))
	expects := false
	for i, path := range paths {
		data, err := os.ReadFile(path) //nolint:gosec // files the owner names on the command line
		if err != nil {
			return nil, false, err
		}
		if err := json.Unmarshal(data, &runs[i]); err != nil {
			return nil, false, fmt.Errorf("%s: %w", path, err)
		}
		for _, t := range runs[i] {
			expects = expects || t.Expect != nil
		}
	}
	return runs, expects, nil
}

// writeRun writes one run's rows: one per example, checked against its
// turn's expectations and the previous answer's tempo.
func writeRun(b *strings.Builder, run string, turns []Turn, expects bool) {
	blank := ""
	if expects {
		blank = " |"
	}
	prevTempo := 0.0
	for n, t := range turns {
		if t.Answer == nil {
			fmt.Fprintf(b, "| %s | %d | (error: %s) | | | | | |%s\n", run, n+1, t.Error, blank)
			continue
		}
		for _, c := range t.Answer.Candidates {
			m := measure(c, t.Answer.Reply)
			fmt.Fprintf(b, "| %s | %d | %s | %s | %s | %s | %d / %d | %s |",
				run, n+1, c.Title, m.Meter, m.Bars, yes(m.Sus4), m.Tensioned, m.Chords, yes(m.Claims))
			if expects {
				fmt.Fprintf(b, " %s |", strings.Join(check(t.Expect, c, prevTempo), ", "))
			}
			b.WriteString("\n")
		}
		if len(t.Answer.Candidates) > 0 {
			prevTempo = t.Answer.Candidates[0].TempoBPM
		}
	}
}

// runName is how the report names a run file: "run-1", or with its
// directory ("length/run-1") when the runs come from several directories,
// one per script.
func runName(paths []string, i int) string {
	name := strings.TrimSuffix(filepath.Base(paths[i]), ".json")
	for _, p := range paths {
		if filepath.Dir(p) != filepath.Dir(paths[i]) {
			return filepath.Base(filepath.Dir(paths[i])) + "/" + name
		}
	}
	return name
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
