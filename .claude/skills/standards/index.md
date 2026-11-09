# 業界標準の索引

項目の形と書かないものは `SKILL.md`。分野は #213 の種まきに従う。

## 演奏データ

### Standard MIDI File（SMF）

| 欄 | 内容 |
| --- | --- |
| 発行元 | MIDI Association（MIDI.org）。文書は RP-001 |
| 何を決めているか | 時刻つきの MIDI データ（トラック、テンポ、拍子、トラック名などのメタ情報）を、プログラムの間で交換するためのファイル形式 |
| 一次情報 | https://midi.org/standard-midi-files 。仕様の PDF のダウンロードには MIDI.org への参加（Join）が要る |
| 状態 | 現行。関連する推奨実践（歌詞、デバイス名・プログラム名のメタイベント、言語と表示の拡張）が同じページにある。MIDI 2.0 の Universal MIDI Packet を使う後継の形式として MIDI Clip File Specification がある |
| 当たるべき場面 | MIDI の書き出しを変えるとき。SMF を読み込むとき（#207 の (c)）。テンポ・拍子・調・トラック名のメタイベントを足すとき |
| middy との接点 | `backend/internal/music/smf`（書き出しだけ。読み込みはない） |
| 確認日 | 2026-11-09 |

### General MIDI（GM 1、GM 2、GM Lite）

| 欄 | 内容 |
| --- | --- |
| 発行元 | MIDI Association（MIDI.org） |
| 何を決めているか | 準拠する音源の最低限の構成と振る舞い。共通の音色の並び（GM Sound Set）で、機器が違っても同じように鳴るようにする |
| 一次情報 | https://midi.org/general-midi 。仕様のダウンロードには MIDI.org への参加（Join）が要る |
| 状態 | GM 1 は 1999 年に GM 2 に置き換えられた。ただし GM 1 は、SMF で配られる音楽で今も広く使われている |
| 当たるべき場面 | 楽器（プログラム番号）や打楽器（ノート番号）を足す・変えるとき。楽器の系統を決めるとき |
| middy との接点 | 伴奏の型の `Program`、ドラムの型の打音のノート番号（`backend/internal/music/pattern`）、楽器の系統（BL-110a） |
| 確認日 | 2026-11-09 |

## 譜面・リードシート

### MusicXML

| 欄 | 内容 |
| --- | --- |
| 発行元 | W3C Music Notation Community Group。W3C 標準ではなく、W3C の標準化の過程にもない（Community Group の最終報告） |
| 何を決めているか | 楽譜をアプリの間で交換し、保存するための開かれた形式 |
| 一次情報 | https://www.w3.org/2021/06/musicxml40/ （Version 4.0、2021-06-01 の最終報告）。編集中の版は、このページが示す https://w3c.github.io/musicxml/ 。登録なしで読める |
| 状態 | このページは 4.0 の最終報告。より新しい版の有無は確かめていない。ページから、同じ Community Group の MNX との比較の文書に行ける |
| 当たるべき場面 | 楽譜の書き出し・読み込みを考えるとき。和音記号の表記を決めるとき（「Chord Symbols and Diagrams」の章） |
| middy との接点 | 今はない。和音の表記（`docs/architecture/domain-model.md`）と比べる候補 |
| 確認日 | 2026-11-09 |
