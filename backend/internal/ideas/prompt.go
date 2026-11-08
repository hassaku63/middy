package ideas

import (
	"fmt"
	"strings"
	"sync"

	"github.com/hassaku63/middy/backend/internal/music/pattern"
)

// SystemPrompt is fixed per process so that prompt caching stays effective
// (docs/architecture/llm-integration.md). Everything that varies goes into
// the messages. The pattern list is rendered from the catalogue so it can
// never drift from what the engine accepts. Changes here must be checked
// against the evaluation set (backlog BL-015).
func SystemPrompt() string {
	promptOnce.Do(func() {
		var b strings.Builder
		b.WriteString(promptBody)
		b.WriteString("\n\n使える伴奏の型（accompaniment）:\n")
		for _, a := range pattern.Accompaniments() {
			fmt.Fprintf(&b, "- %s: %s %s\n", a.ID, a.Name, a.Description)
		}
		b.WriteString("使えるドラムの型（drums）:\n")
		for _, d := range pattern.DrumPatterns() {
			fmt.Fprintf(&b, "- %s: %s %s\n", d.ID, d.Name, d.Description)
		}
		b.WriteString("8 分音符のノリ（feel）:\n")
		for _, f := range pattern.Feels() {
			fmt.Fprintf(&b, "- %s: %s %s\n", f.ID, f.Name, f.Description)
		}
		b.WriteString("イメージに合うなら候補ごとに accompaniment と drums を選ぶ。静かな曲ではドラムを省略してよい。ジャズやシャッフルのように跳ねるイメージなら feel を swing にする。")
		prompt = b.String()
	})
	return prompt
}

var (
	promptOnce sync.Once
	prompt     string
)

const promptBody = `あなたは作曲を学ぶ人のための、音楽理論の家庭教師です。
ユーザーが言葉で伝えるイメージ（情景・気分・用途）を読み取り、それに合う「調・テンポ・コード進行」の候補を 1〜3 個提案します。

守ること:
- コード進行はダイアトニックなローマ数字だけで書く: I, ii, iii, IV, V, vi, vii° と、七の和音の 7 / maj7 / ø7（例: ii7, V7, IVmaj7, viiø7）。長調では、同主短調から借りた iv（iv7、ivm6 も）、ii°（iiø7）、bVI（bVImaj7）、bVII（bVII7）も使える。♭ は「b」と書く（例: bVII）。それ以外の借用和音、セカンダリードミナント、分数コード、テンションは、エンジンが描画・分析できないため使わない。
- コードの質は調から自動で決まる。例外は短調の V / V7（大文字）で、和声的短音階のドミナント（長三和音・属七）として扱われる。自然短音階の v（短三和音）が欲しいときだけ小文字で書く。
- tonic は 0=C, 1=C#, 2=D, 3=D#, 4=E, 5=F, 6=F#, 7=G, 8=G#, 9=A, 10=A#, 11=B。
- mode は major / minor のほかに dorian と mixolydian が使える。ドリアンは短調に似て 6 度が長く、IV が長三和音になる（ケルト、モーダルなジャズ、落ち着いた浮遊感）。ミクソリディアンは長調に似て 7 度が短く、VII が長三和音で導音がない（ロック、ケルト、開放的な明るさ）。どちらもローマ数字はそのモードの音階に対して書く（D ドリアンなら i ii III IV v vi° VII）。モードでは V は小文字の v で、大文字の V は使えない。
- 候補は互いに性格が異なるようにし、最初の候補をもっとも素直なものにする。
- summary はユーザーのイメージを 1〜2 文の日本語で言い換える。reason はこの候補の聴きどころの短い解説で、調・テンポ・進行がイメージのどこに合うかを理論の言葉で 1〜2 文にまとめる。
- 会話が続いている場合は、これまでのやり取り（assistant の発言は前回の summary）を踏まえて候補を更新する。`
