package music

import (
	"encoding/json"
	"testing"
)

func TestKeyScale(t *testing.T) {
	tests := []struct {
		key  Key
		want [7]PitchClass
	}{
		{Key{C, Major}, [7]PitchClass{C, D, E, F, G, A, B}},
		{Key{A, Minor}, [7]PitchClass{A, B, C, D, E, F, G}},
		{Key{G, Major}, [7]PitchClass{G, A, B, C, D, E, Fs}},
		{Key{F, Major}, [7]PitchClass{F, G, A, As, C, D, E}},
		{Key{E, Minor}, [7]PitchClass{E, Fs, G, A, B, C, D}},
	}
	for _, tt := range tests {
		t.Run(tt.key.String(), func(t *testing.T) {
			if got := tt.key.Scale(); got != tt.want {
				t.Fatalf("Scale() = %v, want %v", got, tt.want)
			}
			for i, pc := range tt.want {
				got, err := tt.key.Degree(i + 1)
				if err != nil || got != pc {
					t.Fatalf("Degree(%d) = %v, %v; want %v", i+1, got, err, pc)
				}
				if !tt.key.Contains(pc) {
					t.Fatalf("Contains(%v) should be true", pc)
				}
			}
		})
	}
}

func TestKeyDegreeBounds(t *testing.T) {
	k := Key{C, Major}
	for _, d := range []int{0, 8, -1} {
		if _, err := k.Degree(d); err == nil {
			t.Fatalf("Degree(%d) should fail", d)
		}
	}
	if k.Contains(Cs) {
		t.Fatal("C# is not in C major")
	}
}

func TestKeyJSON(t *testing.T) {
	k := Key{Tonic: A, Mode: Minor}
	b, err := json.Marshal(k)
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != `{"tonic":9,"mode":"minor"}` {
		t.Fatalf("unexpected JSON %s", b)
	}
	var back Key
	if err := json.Unmarshal(b, &back); err != nil || back != k {
		t.Fatalf("round trip = %+v, %v", back, err)
	}
	if err := json.Unmarshal([]byte(`{"tonic":9,"mode":"Minor"}`), &back); err != nil || back != k {
		t.Fatalf("mode must be read case-insensitively: %+v, %v", back, err)
	}
	if err := json.Unmarshal([]byte(`{"tonic":0,"mode":"lydian"}`), &back); err == nil {
		t.Fatal("unknown mode must fail to unmarshal")
	}
}
