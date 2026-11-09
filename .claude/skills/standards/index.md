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

### MIDI 1.0

| 欄 | 内容 |
| --- | --- |
| 発行元 | MIDI Association（MIDI.org） |
| 何を決めているか | 楽器や機器がデジタルのメッセージでやりとりするためのプロトコル。メッセージの形と、5 ピン DIN の伝送 |
| 一次情報 | https://midi.org/midi-1-0 。中核の仕様は https://midi.org/midi-1-0-core-specifications 。詳細仕様のダウンロードのボタンには、SMF のページのような参加（Join）の表示がない（リンク先の Google Drive は開いていない）。メッセージの参照表（https://midi.org/spec-detail ）のダウンロードには参加が要る |
| 状態 | 現行。中核の仕様は 1996 年の改訂版。その後の追加は Addenda（https://midi.org/midi-1-0-addenda ）にあり、合わせて読むとある。MIDI 2.0 の土台でもある |
| 当たるべき場面 | SMF に書くチャンネルのメッセージを足す・変えるとき（コントロールチェンジ、ピッチベンドなど）。外部の機器と MIDI でつなぐとき |
| middy との接点 | `backend/internal/music/smf` が書くノートのオン・オフとプログラムチェンジ。ピッチの MIDI ノート番号（`docs/architecture/domain-model.md`） |
| 確認日 | 2026-11-09 |

### MIDI 2.0（MIDI-CI、Universal MIDI Packet、Property Exchange、MIDI Clip File）

| 欄 | 内容 |
| --- | --- |
| 発行元 | MIDI Association（MIDI.org） |
| 何を決めているか | MIDI 1.0 を広げる仕様の一群。双方向の機器が拡張の機能を取り決める仕組み（MIDI-CI。Profile、Property Exchange、Process Inquiry を含む）。MIDI 1.0 と 2.0 のメッセージを入れる共通の入れ物（UMP）と MIDI 2.0 Protocol。UMP の列を入れるファイル形式（MIDI Clip File） |
| 一次情報 | https://midi.org/midi-2-0 。コア仕様の 6 文書の zip（https://midi.org/midi-2-0-core-specification-collection ）のダウンロードには参加（Join）が要る。個々の文書のページは、参加を求める表示があるもの（UMP と MIDI 2.0 Protocol）と、ないもの（MIDI-CI、MIDI Clip File）がある。Property Exchange は https://midi.org/midi-2-0-property-exchange で、参加が要る |
| 状態 | 現行。MIDI 1.0 を置き換えず、その上に作るとある。MIDI Clip File は SMF の format 0 に当たり、format 1 に当たる MIDI Container File の仕様がこれから出るとある |
| 当たるべき場面 | MIDI の書き出しを UMP や MIDI Clip File にするか考えるとき。音ごとの制御や高い分解能が要るとき。外部の機器の能力を問い合わせるとき |
| middy との接点 | 今はない。書き出しは MIDI 1.0 の SMF（`backend/internal/music/smf`） |
| 確認日 | 2026-11-09 |

### MPE（MIDI Polyphonic Expression）

| 欄 | 内容 |
| --- | --- |
| 発行元 | MIDI Association と AMEI（ページに「MMA/AMEI Recommended Practice」とある） |
| 何を決めているか | 和音を弾きながら音ごとに音高や音色を変えられるコントローラーのための推奨実践。MIDI 1.0 のメッセージで、チャンネルを音ごとに使って音ごとの制御を伝える |
| 一次情報 | https://midi.org/mpe-midi-polyphonic-expression 。ダウンロードには MIDI.org への参加（Join）が要る |
| 状態 | 現行。前の版（RP-053）を置き換えた。技術的な設計の変更はないとある |
| 当たるべき場面 | 音ごとのピッチベンドや表現を書き出す・受け取るとき。表現力のあるコントローラーからの入力を考えるとき |
| middy との接点 | 今はない。SMF はノートのオン・オフとプログラムチェンジだけを書く（`backend/internal/music/smf`） |
| 確認日 | 2026-11-09 |

### MIDI Tuning（MIDI Tuning Standard）

| 欄 | 内容 |
| --- | --- |
| 発行元 | MIDI Association（MIDI.org）。MIDI 1.0 の Addenda に入っている |
| 何を決めているか | 12 平均律以外の調律（マイクロチューニング）を機器の間で共有し、演奏中に切り替えるためのメッセージ |
| 一次情報 | https://midi.org/midi-tuning-updated-specification 。ダウンロードには MIDI.org への参加（Join）が要る |
| 状態 | 「Updated Specification」として、あとから足された拡張（Bank/Dump、Scale/Octave など）を 1 つにまとめた文書 |
| 当たるべき場面 | 12 平均律以外の調律を鳴らす・書き出すとき |
| middy との接点 | 今はない。再生は A4 = 440 Hz の 12 平均律（`frontend/src/features/player/WebAudioSynth.ts` の `midiToHz`）。調律の分野は BL-138 |
| 確認日 | 2026-11-09 |

### Web MIDI API

| 欄 | 内容 |
| --- | --- |
| 発行元 | W3C Audio Working Group |
| 何を決めているか | Web アプリが MIDI の入力・出力の機器を列挙して選び、MIDI メッセージを送受信するための API。SMF の再生はこの仕様の範囲外とある |
| 一次情報 | https://www.w3.org/TR/webmidi/ 。登録なしで読める。編集中の版は、このページが示す https://webaudio.github.io/web-midi-api/ （開いていない） |
| 状態 | Working Draft（2026-10-31）。W3C 勧告への過程（Recommendation track）にある |
| 当たるべき場面 | 外部の音源やコントローラーと MIDI でつなぐとき（backlog の Ideas の「Web MIDI 出力、外部音源」） |
| middy との接点 | 今はない。再生は Web Audio による自前の合成（`frontend/src/features/player`）、書き出しは SMF のダウンロード（`frontend/src/features/export`） |
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

### MNX

| 欄 | 内容 |
| --- | --- |
| 発行元 | W3C Music Notation Community Group（MusicXML と同じ） |
| 何を決めているか | 楽譜の形式の新しい仕様。Community Group のページに、MNX という仮の名前で開発しているとある |
| 一次情報 | https://w3c-cg.github.io/mnx/ 。前の https://w3c.github.io/mnx/ は「Moved」とここを示す。この環境からは届かなかった。Community Group のページは https://www.w3.org/community/music-notation/ |
| 状態 | 開発中（Community Group のページによる）。一次情報では確かめていない |
| 当たるべき場面 | MusicXML と並べて、楽譜の書き出し・読み込みを考えるとき |
| middy との接点 | 今はない |
| 確認日 | 未確認（一次情報のホストに届かなかった） |

### MEI（Music Encoding Initiative）

| 欄 | 内容 |
| --- | --- |
| 発行元 | Music Encoding Initiative（コミュニティ。マインツの科学・文学アカデミーと RISM Digital Center が共同で受け入れている）。スキーマは MEI Technical Team が保守する |
| 何を決めているか | 楽譜の文書を、機械が読める構造で書くための XML スキーマと、その使い方のガイドライン |
| 一次情報 | https://music-encoding.org/ 。ガイドラインは https://music-encoding.org/guidelines/v5/content/ 。登録なしで読める。Educational Community License 2.0 で無償とある（https://music-encoding.org/about/ ） |
| 状態 | 現行のガイドラインは MEI 5。開発版と過去の版もサイトにある。about のページに MusicXML との違いの節がある |
| 当たるべき場面 | 楽譜の符号化を、分析や研究の用途で考えるとき。MusicXML と比べるとき |
| middy との接点 | 今はない |
| 確認日 | 2026-11-09 |

### ABC 記法

| 欄 | 内容 |
| --- | --- |
| 発行元 | abcnotation.com の wiki。ページは発行の団体を名乗っていない。変更は abcusers のメーリングリストかフォーラムで話し合うとある |
| 何を決めているか | 文字だけで書く、人にも計算機にも読める楽譜の記法 |
| 一次情報 | https://abcnotation.com/wiki/abc:standard （各版への入り口。リンクはここに張るよう勧めている）。登録なしで読める |
| 状態 | 現行は 2.1（2011 年 12 月）。新しい草案として 2.2 がある |
| 当たるべき場面 | 文字の譜面の書き出し・読み込みを考えるとき。伴奏やドラムの型をデータにするとき（#207）。和音記号の表記を比べるとき（2.1 の「Chord symbols」の節） |
| middy との接点 | 今はない |
| 確認日 | 2026-11-09 |

### LilyPond

| 欄 | 内容 |
| --- | --- |
| 発行元 | LilyPond（GNU Project の一部の自由ソフトウェア） |
| 何を決めているか | 標準の形式ではなく、楽譜を浄書するプログラム。文字で書いた入力ファイルをコンパイルして楽譜にする。和音と歌詞を合わせてリードシートも作れるとある |
| 一次情報 | https://lilypond.org/ 。入力の書き方の紹介は https://lilypond.org/text-input.html 、詳しくは Manuals。無償（GPL） |
| 状態 | 安定版と開発版がある |
| 当たるべき場面 | 楽譜を印刷向けに書き出すことを考えるとき |
| middy との接点 | 今はない。楽譜の表示は backlog の Ideas にある |
| 確認日 | 2026-11-09 |

### Humdrum（**kern）

| 欄 | 内容 |
| --- | --- |
| 発行元 | Humdrum（David Huron が 1980 年代に作った、計算による音楽分析の道具一式）。サイトは標準化の団体を名乗っていない |
| 何を決めているか | **kern は、共通慣習期の西洋音楽の音高と長さを中心に書く表現。楽譜の印刷や音の生成ではなく、分析のための表現とある |
| 一次情報 | https://www.humdrum.org/rep/kern/ 。道具一式は https://www.humdrum.org/ 。登録なしで読める |
| 状態 | ページに版の記載はない。元の Humdrum Toolkit のほかに、Humdrum Extras や humlib などの新しい実装がある |
| 当たるべき場面 | 分析のための楽譜データを読み書きするとき |
| middy との接点 | 今はない |
| 確認日 | 2026-11-09 |

### ChordPro

| 欄 | 内容 |
| --- | --- |
| 発行元 | ChordPro.ORG（2015 年から、元の作者に代わって標準と参照実装を保守する利用者のチーム） |
| 何を決めているか | 歌詞と和音を書く、文字のファイル形式。歌詞の中に角括弧で和音を、波括弧で指示（directive）を書く |
| 一次情報 | https://www.chordpro.org/chordpro/home/ （「ChordPro 6: The File Format Specification」）。無償で、オープンソース |
| 状態 | 現行の仕様は ChordPro 6 |
| 当たるべき場面 | 和音つきの歌詞やリードシートの書き出し・読み込みを考えるとき。音名の体系を比べるとき（Nashville やローマ数字の表記を扱えるとある） |
| middy との接点 | 今はない |
| 確認日 | 2026-11-09 |

### iReal Pro の Custom Chord Chart Protocol

| 欄 | 内容 |
| --- | --- |
| 発行元 | Technimo LLC（iReal Pro のアプリ。ページのフッターによる） |
| 何を決めているか | iReal Pro の曲（題名、作曲者、スタイル、調、和音の進行）を HTML のリンクの URL（`irealbook://`）に入れて、アプリに取り込ませる書き方。和音記号、小節線、拍子、リハーサルマークなどの記号 |
| 一次情報 | https://www.irealpro.com/ireal-pro-custom-chord-chart-protocol/ 。https://www.irealpro.com/ireal-pro-file-format はここに転送される。登録なしで読める |
| 状態 | ページに版の記載はない。URL の 5 番目の要素は今は使わないとある |
| 当たるべき場面 | 進行をリードシートのアプリに渡すことを考えるとき。和音記号の表記を比べるとき |
| middy との接点 | 今はない。ADR-0012 が、進行とスタイルを別々に持つアプリの例に挙げている。middy との違い: 和音記号を `-`・`^7`・`h7`・`o` などの記号で書き、middy の和音名（`m`・`maj7`・`m7b5`・`dim`、`backend/internal/music/chord.go`）とは表記が違う |
| 確認日 | 2026-11-09 |
