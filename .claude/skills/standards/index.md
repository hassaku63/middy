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

## 和声・分析の記法

### MusicXML の `<harmony>`（和音記号）

| 欄 | 内容 |
| --- | --- |
| 発行元 | W3C Music Notation Community Group（MusicXML の項と同じ） |
| 何を決めているか | 楽譜の中の和音記号と機能和声の分析を書く要素。根音（音名）か度数（numeral）、和音の種類（`<kind>`）、転回、ベース、足す・除く・変える音（`<degree>`）を持つ |
| 一次情報 | https://www.w3.org/2021/06/musicxml40/musicxml-reference/elements/harmony/ 。和音の種類の一覧は https://www.w3.org/2021/06/musicxml40/musicxml-reference/data-types/kind-value/ 、手引きは https://www.w3.org/2021/06/musicxml40/tutorial/chord-symbols-and-diagrams/ 。登録なしで読める |
| 状態 | MusicXML 4.0 の最終報告の一部。ローマ数字の文字で書く `<function>` の要素は 4.0 で非推奨になったとある。編集中の版の https://w3c.github.io/musicxml/ は https://w3c-cg.github.io/musicxml/ に移ったと示すが、移った先にはこの環境から届かなかった |
| 当たるべき場面 | 和音記号を譜面の形式で書き出す・読み込むとき。和音の質を足すとき（種類の一覧と比べる）。セカンダリードミナントを外の形式でどう書くか決めるとき |
| middy との接点 | 今はない。和音の質（`ChordQuality`、`backend/internal/music/chord.go`）と和音名の接尾辞（`frontend/src/shared/lib/notation.ts`）と比べる候補 |
| middy との違い | セカンダリードミナントを、度数の数字（5 と 2）を持つ和音を並べて表す（V of II）。middy はラベルの文字列（`V7/ii`）で書き、内部では `Step.Applied` で持つ。種類の一覧に 7sus4 に当たる値はなく、`<degree>` で種類に音を足して表す。middy は `dom7sus4` を 1 つの質として持つ（ADR-0025） |
| 確認日 | 2026-11-09 |

### Harte らの和音表記（ISMIR 2005）

| 欄 | 内容 |
| --- | --- |
| 発行元 | Christopher Harte、Mark Sandler、Samer Abdallah（Queen Mary, University of London）と Emilia Gómez（Universitat Pompeu Fabra）の ISMIR 2005 の論文。標準化の団体ではない |
| 何を決めているか | 和音のラベルを、文脈に依らず機械が一意に読めるように書く文字の文法（根音、コロン、略号か度数の一覧、スラッシュとベース）。和音のないことを示す記号も持つ |
| 一次情報 | https://archives.ismir.net/ismir2005/paper/000080.pdf （"Symbolic Representation of Musical Chords: A Proposed Syntax for Text Annotations"）。無償で読める |
| 状態 | 提案の論文。文法は BNF で示されている。JAMS の `chord_harte` 名前空間がこの文法に従い、`chord` 名前空間がそれを広げている（JAMS の項） |
| 当たるべき場面 | 和音のラベルを、研究用のデータセットや音響からの和音推定の道具とやりとりするとき。機械向けの和音名の表記を決めるとき |
| middy との接点 | 今はない。和音名（`Chord.String()`、`G7`）と比べる候補 |
| middy との違い | 略号が違う（短三和音は `min`、半減七は `hdim7`。middy の和音名は `m`、`m7b5`）。略号の一覧に 7sus4 とパワーコードはない（middy は `7sus4`、`5`）。スラッシュのあとのベースは和音の根音からの度数で書く（`/3`）。middy の和音名はベースを音名で書き（`C/E`）、ローマ数字のラベルは `/` のあとに調の度数を書く（`IV/1`、ADR-0020）ので、同じ `/3` でも数える基準が違う |
| 確認日 | 2026-11-09 |

### JAMS（JSON Annotated Music Specification）

| 欄 | 内容 |
| --- | --- |
| 発行元 | JAMS development team（ドキュメントの著作権の表示による）。ISMIR 2014 の論文（Humphrey ほか、New York University の Music and Audio Research Lab）が提案した |
| 何を決めているか | 音楽の注釈（拍、和音、区間、タグなど）を JSON で持つ形式。注釈の種類ごとの名前空間のスキーマと、その検証 |
| 一次情報 | https://jams.readthedocs.io/en/stable/ 。和音の名前空間は https://jams.readthedocs.io/en/stable/namespaces/chord.html 。論文は https://archives.ismir.net/ismir2014/paper/000355.pdf （CC BY 4.0）。どれも無償で読める |
| 状態 | ドキュメントに変更履歴がある。最新の情報は GitHub（https://github.com/marl/jams ）を見るようにとあるが、この session からは開けなかった。和音の名前空間は `chord_harte`、それを広げた `chord`、ローマ数字の `chord_roman` がある |
| 当たるべき場面 | 時刻つきの和音・調・拍の注釈を書き出す・読み込むとき。mir_eval で評価するとき（ドキュメントに橋渡しがあるとある） |
| middy との接点 | 今はない。`Timeline` の `Harmony`（和音の時刻と分析）が近い |
| middy との違い | 和音の注釈は時刻と長さを秒で持つ（middy は tick。`DefaultPPQ`）。`chord_roman` は質を記号で書き（`o` が減三和音、`h` が半減七、`x` が減七、`d` が属七、`s` が掛留）、転回を数字（`V6`、`V64`）で書く。middy は `°`、`ø7`、`°7` と、`/` のあとの調の度数（ADR-0020） |
| 確認日 | 2026-11-09 |

### RomanText（music21）

| 欄 | 内容 |
| --- | --- |
| 発行元 | Dmitri Tymoczko、Mark Gotham、Michael Scott Cuthbert、Christopher Ariza の ISMIR 2019 の論文。読み書きは music21（Cuthbert らの Python のライブラリ）が実装している |
| 何を決めているか | ローマ数字の和声分析を、小節・拍・調と一緒に文字で書く形式（`.rntxt`） |
| 一次情報 | https://archives.ismir.net/ismir2019/paper/000012.pdf （"The RomanText Format: A Flexible and Standard Method for Representing Roman Numeral Analyses"、CC BY 4.0）。music21 の実装は https://raw.githubusercontent.com/cuthbertLab/music21/master/music21/romanText/translate.py ほか（BSD）。music21 のドキュメントの https://web.mit.edu/music21/doc/ は 404 だった |
| 状態 | 論文が仕様を定める。ローマ数字の全一覧の参照として https://github.com/MarkGotham/When-in-Rome を示す（開いていない）。music21 の `romanText` は、Clercq と Temperley の形式と DCML の表の形式も扱う |
| 当たるべき場面 | ローマ数字の分析を外へ書き出す・読み込むとき。ローマ数字の記号（減・半減、転回、セカンダリードミナント、増六・ナポリ）を足すとき。転調を例に入れることを考えるとき |
| middy との接点 | ローマ数字のラベル（`RomanNumeral`、`ParseRomanNumeral`、`backend/internal/music/chord.go` と `roman.go`） |
| middy との違い | 減三和音を英字の `o`、半減七を `/o` で書く（middy は `°`、`ø7`。`vii°7/x` では `o` も読む、ADR-0024）。転回を数字付き低音（`6`、`6/4`、`6/5`）で書き、`/` を数字の区切りとセカンダリードミナントに使う。middy のラベルは `/` のあとの数字をベースの調の度数と読む（ADR-0020）ので、`I6/4` の `/4` は middy の文法ではベースの位置に当たる。ナポリを `N`、増六を `It`・`Ger`・`Fr` と略す |
| 確認日 | 2026-11-09 |

### DCML の和声注釈の標準

| 欄 | 内容 |
| --- | --- |
| 発行元 | Digital and Cognitive Musicology Lab（DCML、École Polytechnique Fédérale de Lausanne） |
| 何を決めているか | 楽譜に和声分析のラベルを書き込むための注釈の規約と、ラベルを要素（調、ローマ数字、形、数字付き低音、変化、向かう先、終止、フレーズ）に分ける正規表現 |
| 一次情報 | https://raw.githubusercontent.com/DCMLab/standards/main/README.md と、同じリポジトリの正規表現 https://raw.githubusercontent.com/DCMLab/standards/main/harmony.py 。README が示す手引き（https://dcmlab.github.io/standards/build/html/tutorial/ ）と参照（https://dcmlab.github.io/standards ）にはこの環境から届かず、リポジトリのページ（https://github.com/DCMLab/standards ）もこの session からは開けなかった。ライセンスは確かめていない |
| 状態 | README に版ごとのリリースノートがある。手引きは、MuseScore 3 で楽譜にラベルを入れ、DCML のパーサーで取り出す使い方を説明しているとある。使用例に The Annotated Beethoven Corpus（ABC）を挙げる |
| 当たるべき場面 | 楽譜に付いた和声分析のコーパスを読むとき。終止やフレーズの区切りを注釈として持つことを考えるとき |
| middy との接点 | ローマ数字のラベルと終止の分類（`Cadence`、`docs/architecture/domain-model.md`） |
| middy との違い | 和音の形を記号 `%`、`o`、`+`、`M` で書く（README に、`+M7` は増三和音に長七とある。ほかの記号の意味は開けなかった手引きで確かめる）。middy は `ø7`、`°`、`+`、`maj7` のような記号で書く。転回は数字付き低音（`7`、`65`、`43`、`42`、`2`、`64`、`6`）で書く（middy は `/` のあとの調の度数、ADR-0020）。終止は `PAC`・`IAC`・`HC` などの略号をラベルに付け、middy の `Cadence`（`authentic` など 5 つ）とは名前も数も違う。対応は開けなかった手引きで確かめる |
| 確認日 | 2026-11-09 |

### MIREX の和音の語彙（Audio Chord Estimation）

| 欄 | 内容 |
| --- | --- |
| 発行元 | MIREX（Music Information Retrieval Evaluation eXchange）。一次情報を開けず、確かめていない |
| 何を決めているか | 一次情報を開けず、確かめていない。音響からの和音推定の評価の課題が、どの和音の語彙で比べるかに当たる |
| 一次情報 | https://www.music-ir.org/mirex/wiki/ の年ごとの Audio Chord Estimation のページ（試したのは 2021、2024、2025 と MIREX_HOME）。サーバーの TLS 証明書が期限切れで、この環境から開けなかった |
| 状態 | 未確認 |
| 当たるべき場面 | 和音推定の結果や評価のデータと、middy の和音の質を対応させるとき |
| middy との接点 | 今はない |
| 確認日 | 未確認（一次情報のサーバーの証明書が期限切れで開けなかった） |

## 音高・調律

### ISO 16（標準の音高 A = 440 Hz）

| 欄 | 内容 |
| --- | --- |
| 発行元 | ISO（国際標準化機構）。一次情報を開けず、確かめていない |
| 何を決めているか | 一次情報を開けず、確かめていない。索引には、標準の音高を A = 440 Hz とする規格として置く |
| 一次情報 | ISO のサイトの規格のページ。試した https://www.iso.org/standard/3601.html （ISO 16 のページかどうかも確かめていない）は、Cloudflare の確認の画面（HTTP 403）で開けなかった。入手の条件は確かめていない |
| 状態 | 未確認 |
| 当たるべき場面 | 再生の基準の高さ（A4 の周波数）を決める・変えられるようにするとき |
| middy との接点 | 再生は A4（MIDI 69）= 440 Hz（`frontend/src/features/player/WebAudioSynth.ts` の `midiToHz`） |
| 確認日 | 未確認（ISO のサイトが確認の画面を返した） |

### 科学的音高表記（scientific pitch notation）

| 欄 | 内容 |
| --- | --- |
| 発行元 | 記法を定めた一次の文書を見つけられず、確かめていない |
| 何を決めているか | 音名にオクターブの番号を付けて音高を書く記法（`C4` など） |
| 一次情報 | 記法そのものを定めた文書は、この環境から開けるホストで見つからなかった（候補として試した pubs.aip.org と acousticalsociety.org は proxy に断られた）。同じ番号の付け方を使う標準として、MusicXML 4.0 の octave のデータ型（https://www.w3.org/2021/06/musicxml40/musicxml-reference/data-types/octave/ 、登録なし）が、中央の C から始まるオクターブを 4 とし、0〜9 を使うとある |
| 状態 | 未確認 |
| 当たるべき場面 | 音高を文字で表示する・書き出すとき。外の形式や機器とオクターブの番号を合わせるとき |
| middy との接点 | `Pitch.String()`（`"C#4"`）と `NewPitch`。C4 = 60（`docs/architecture/domain-model.md`、`backend/internal/music/pitch.go`） |
| 確認日 | 未確認（記法そのものの一次情報を開けなかった。MusicXML の octave のデータ型は 2026-11-09 に開いた） |

### Scala のスケールファイル（`.scl`）

| 欄 | 内容 |
| --- | --- |
| 発行元 | Scala（Manuel Op de Coul が作った調律のプログラム）。文書は Huygens-Fokker 財団のサイトにある |
| 何を決めているか | 1 つの音階（調律）を、比か cents の値の並びとして書く文字のファイル形式。1 ファイルに 1 つの音階 |
| 一次情報 | https://www.huygens-fokker.org/scala/scl_format.html 。登録なしで読める。Scala 本体は https://www.huygens-fokker.org/scala/ （無償のフリーウェア） |
| 状態 | ページに版の記載はない。スケールのアーカイブの大きさと Scala の普及から、音階の交換の標準になっているとある。対応するソフトの一覧がある。鍵盤（ノート番号）と音階の度数の対応は、別の形式（`.kbm`、Scala Keyboard Mappings）で書くとある |
| 当たるべき場面 | 12 平均律以外の調律を読み込む・書き出すとき |
| middy との接点 | 今はない。再生は A4 = 440 Hz の 12 平均律（`frontend/src/features/player/WebAudioSynth.ts` の `midiToHz`） |
| 確認日 | 2026-11-09 |

## 音源・音声

### SoundFont 2

| 欄 | 内容 |
| --- | --- |
| 発行元 | 一次情報を開けず、確かめていない |
| 何を決めているか | 一次情報では確かめていない。SFZ の参照サイト（https://sfzformat.com/ ）は、サンプルと鳴らし方の定義を同じ 1 つのバイナリファイルに入れる形式で、SFZ とは別物と説明している（SoundFont の一次情報ではない） |
| 一次情報 | 公式の置き場所を確かめられなかった。http://www.soundfont.com/ と https://www.creative.com/ はこの環境のプロキシに拒まれた（403）。MIDI.org のサイト内検索（https://midi.org/?s=soundfont ）に SoundFont の仕様のページはなかった。仕様の PDF を載せている第三者のサイト（www.synthfont.com）にも届かなかった |
| 状態 | 一次情報では確かめていない |
| 当たるべき場面 | サンプルの音源で鳴らすことを考えるとき。外部の音色の集まり（.sf2）を読み込むとき。DLS・SFZ と並べて形式を選ぶとき |
| middy との接点 | 今はない。再生は発振器による合成で、`frontend/src/features/player/WebAudioSynth.ts` のコメントはサンプルの音源を backlog の案としている。`docs/product/vision.md` の「やらないこと」にサンプル音源の配布がある |
| 確認日 | 未確認（一次情報のホストに届かなかった） |

### DLS（Downloadable Sounds）

| 欄 | 内容 |
| --- | --- |
| 発行元 | MIDI Association（MIDI.org）。ページは「MMA の仕様」と書いている |
| 何を決めているか | 音のデータ（サンプル）の形式と、DLS に対応する音源がそれを正しく鳴らすために要る機能。Level 1 が基準の構成を、Level 2 がその拡張を決める |
| 一次情報 | https://midi.org/dls 。Level 1 は https://midi.org/dls-level-1 、Level 2 は https://midi.org/dls-level-2 。どちらもダウンロードには MIDI.org への参加（Join）が要る。chunk の ID の表（DLS Proprietary Chunk IDs）は https://midi.org/spec-detail にある |
| 状態 | MIDI.org の仕様の一覧（https://midi.org/specs ）で、SMF・XMF と並ぶ公式のファイル形式に入っている。Level 2 は Level 1 の論理的な拡張とある。後継や廃止の記載はない |
| 当たるべき場面 | サンプルの音源で鳴らすことを考えるとき。SoundFont 2・SFZ と並べて音色の形式を選ぶとき |
| middy との接点 | 今はない。再生は発振器による合成（`frontend/src/features/player/WebAudioSynth.ts`）。`docs/product/vision.md` の「やらないこと」にサンプル音源の配布がある |
| 確認日 | 2026-11-09 |

### SFZ

| 欄 | 内容 |
| --- | --- |
| 発行元 | 標準化の団体はない。sfzformat.com が参照の場を名乗り、サイトはオープンソース（GitHub の sfz organization）。最初の版は rgc:audio のサイトで公開されたとある |
| 何を決めているか | サンプルの集まりを演奏のためにどう並べるかを定義する、文字のファイル形式。サンプルそのものは含まない |
| 一次情報 | https://sfzformat.com/ 。版の説明は https://sfzformat.com/versions 。登録なしで読める。形式は無償で、無償・商用のどちらにも使えるとある |
| 状態 | 第 2 版は明確に書かれた仕様がなく、サイトは Cakewalk の製品の解説書に載るものを第 2 版とみなすとある。ARIA などの拡張がある。新しい再生ソフトはすべての opcode に対応しなくてよいとある |
| 当たるべき場面 | サンプルの音源を文字のデータで定義することを考えるとき。SoundFont 2・DLS と並べて音色の形式を選ぶとき |
| middy との接点 | 今はない。再生は発振器による合成（`frontend/src/features/player/WebAudioSynth.ts`） |
| 確認日 | 2026-11-09 |

### WAV（RIFF WAVE）と BWF（Broadcast Wave Format）

| 欄 | 内容 |
| --- | --- |
| 発行元 | WAVE は Microsoft（RIFF の一種のファイル）。BWF は EBU（文書は Tech 3285） |
| 何を決めているか | 音声のデータのファイル形式。RIFF は chunk を並べた入れ物で、WAVE はそのうち音声を入れるもの。BWF は WAVE に放送用のメタデータの chunk（Broadcast Audio Extension）を足し、放送の環境や機種の間で音声を交換できるようにする |
| 一次情報 | BWF: https://tech.ebu.ch/docs/tech/tech3285.pdf （登録なしで開けた）。出版物のページ https://tech.ebu.ch/publications/tech3285 はサイトの防御（Cloudflare）に 403 で拒まれ、開けていない。WAVE: BWF が参照する Microsoft の文書（http://support.microsoft.com/kb/q120253/ 。Multimedia Registration Kit）はプロキシに拒まれた（403）。Microsoft Learn の RIFF の概要（https://learn.microsoft.com/en-us/windows/win32/xaudio2/resource-interchange-file-format--riff- 。XAudio2 の文書の一部）と、形式の構造体（https://learn.microsoft.com/en-us/windows/win32/api/mmreg/ns-mmreg-waveformatextensible ）は開けた。どれも無償 |
| 状態 | 開いた BWF の PDF は Version 2（2011 年）で、前の版と互換とある。それより新しい版の有無は、出版物のページが開けず確かめていない。BWF の付録の WAVE の説明は参考で、詳しくは Microsoft の最新の文書に当たれとある |
| 当たるべき場面 | 演奏を音声のファイルに書き出すことを考えるとき。サンプルの音源の音声を読み込むとき |
| middy との接点 | 今はない。書き出しは SMF だけ（`frontend/src/features/export`）。`docs/product/vision.md` の「やらないこと」にサーバー側のオーディオレンダリングがある |
| 確認日 | 2026-11-09（BWF の PDF と Microsoft Learn。WAVE の元の文書は未確認） |

### ラウドネスの測定（ITU-R BS.1770、EBU R 128）

| 欄 | 内容 |
| --- | --- |
| 発行元 | BS.1770 は ITU-R（Study Group 6 が管理）。R 128 は EBU |
| 何を決めているか | BS.1770 は番組の音の大きさ（ラウドネス）と true-peak のレベルを測るアルゴリズム。R 128 は、BS.1770 で測った番組のラウドネスで音声を正規化することと、許される最大のレベルを勧める |
| 一次情報 | BS.1770: https://www.itu.int/rec/R-REC-BS.1770 。現行の版のページに「Free Download」とある。R 128: https://tech.ebu.ch/docs/r/r128.pdf （登録なしで開けた）。出版物のページ https://tech.ebu.ch/publications/r128 はサイトの防御（Cloudflare）に 403 で拒まれ、開けていない |
| 状態 | BS.1770 は最新の版が「In force」で、前の版はすべて「Superseded」。R 128 の PDF は「EBU Recommendation」の 2023 年の版。測り方の細部は EBU Tech 3341・3342 に、配信（streaming）向けの案内は R 128 s2 にあるとある。R 128 の新しい版の有無は、出版物のページが開けず確かめていない |
| 当たるべき場面 | 例や楽器の間で音量を揃えるとき。音声を書き出すときの音量の目標を決めるとき |
| middy との接点 | 今はない。音量はマスターの固定の gain と、velocity からの曲線で決めていて、測っていない（`frontend/src/features/player/WebAudioSynth.ts`） |
| 確認日 | 2026-11-09 |

## ブラウザの音と PWA

### Web Audio API（AudioWorklet を含む）

| 欄 | 内容 |
| --- | --- |
| 発行元 | W3C Web Audio Working Group |
| 何を決めているか | Web アプリで音を処理・合成するための API。AudioNode をつないだ経路のグラフで音を作る。スクリプトで音を処理する仕組み（AudioWorklet）と、実時間より速く描き出す仕組み（OfflineAudioContext）も含む |
| 一次情報 | 勧告は https://www.w3.org/TR/webaudio-1.0/ 。https://www.w3.org/TR/webaudio/ は Web Audio API 1.1 の草案を返す。編集中の版は、このページが示す https://webaudio.github.io/web-audio-api/ （開いていない）。登録なしで読める |
| 状態 | Web Audio API（1.0）は W3C 勧告（2021-06-17）。1.1 は Working Draft（2026-09-22）で、勧告を目指している |
| 当たるべき場面 | 再生の合成やスケジューリングを変えるとき。AudioWorklet で自前の音の処理を書くことを考えるとき。演奏を音声に描き出すことを考えるとき |
| middy との接点 | `frontend/src/features/player/usePlayer.ts`（`AudioContext` を再生ボタンの操作から作る）、`WebAudioSynth.ts`（発振器・フィルター・gain のノード）、`TimelinePlayer.ts`（`AudioContext.currentTime` を時計にした先読みのスケジューラ）。AudioWorklet と OfflineAudioContext は使っていない |
| 確認日 | 2026-11-09 |

### Media Session API

| 欄 | 内容 |
| --- | --- |
| 発行元 | W3C Media Working Group |
| 何を決めているか | 再生中のメディアの情報を OS の画面に出し、OS のメディアの操作やメディアキー（キーボード、ヘッドセット、リモコン、モバイルのロック画面や通知のボタン）を Web アプリにつなぐ API |
| 一次情報 | https://www.w3.org/TR/mediasession/ 。編集中の版は、このページが示す https://w3c.github.io/mediasession/ （開いていない）。登録なしで読める |
| 状態 | Working Draft（2026-11-09）。勧告を目指している |
| 当たるべき場面 | ロック画面や通知、イヤホンのボタンから再生・停止させるとき。再生中の例の名前を OS に見せるとき |
| middy との接点 | 今はない。再生と停止は画面のボタンだけ（`frontend/src/features/player/PlayerControls.tsx`） |
| 確認日 | 2026-11-09 |

### Web Application Manifest

| 欄 | 内容 |
| --- | --- |
| 発行元 | W3C Web Applications Working Group |
| 何を決めているか | Web アプリのメタデータ（名前、アイコン、起動の URL、表示モードなど）を 1 か所に書く、JSON のファイル形式 |
| 一次情報 | https://www.w3.org/TR/appmanifest/ 。編集中の版は、このページが示す https://w3c.github.io/manifest/ （開いていない）。登録なしで読める |
| 状態 | Working Draft（2026-11-06）。仕様は安定しておらず、互換のない変更がありうるが、一部はすでにブラウザに出荷されているとある |
| 当たるべき場面 | manifest の項目（名前、アイコン、表示モード、色など）を足す・変えるとき |
| middy との接点 | `frontend/vite.config.ts` の `VitePWA` の `manifest`（vite-plugin-pwa が生成する）。使っている項目（`name`、`short_name`、`description`、`lang`、`theme_color`、`background_color`、`display`、`start_url`、`icons` と `purpose`）は、どれも今の草案にある。配信は nginx が `application/manifest+json` で `no-cache`（`deploy/nginx/templates/default.conf.template`、ADR-0027） |
| 確認日 | 2026-11-09 |

### Service Workers

| 欄 | 内容 |
| --- | --- |
| 発行元 | W3C Web Applications Working Group |
| 何を決めているか | イベントを受けて起きる worker。ページの取得のしかたを決めたり、push を受けたりする先になり、オフラインで開ける Web アプリの土台になる |
| 一次情報 | https://www.w3.org/TR/service-workers/ （題は「Service Workers Nightly」）。編集中の版は、このページが示す https://w3c.github.io/ServiceWorker/ （開いていない）。登録なしで読める |
| 状態 | Candidate Recommendation Draft（2026-10-29）。living document で、Service Workers 1 という版が勧告に向かっているとある |
| 当たるべき場面 | オフラインの振る舞い、キャッシュの方針、更新の流れを変えるとき。HTTP のキャッシュとの関係は「Web・API・安全」の HTTP Caching の項目 |
| middy との接点 | Service Worker は vite-plugin-pwa（Workbox）が生成する（`frontend/vite.config.ts`。`registerType: 'prompt'`、アプリシェルだけをプリキャッシュし、`/api/` はナビゲーションのフォールバックから外す）。登録と更新の案内は `frontend/src/shared/lib/pwa/wire.ts`。nginx は `sw.js` を `no-cache` で配る（`deploy/nginx/templates/default.conf.template`、ADR-0027） |
| 確認日 | 2026-11-09 |

## 同期・連携

### MIDI Clock と MIDI Time Code（MTC）

| 欄 | 内容 |
| --- | --- |
| 発行元 | MIDI Association（MIDI.org）。どちらも MIDI 1.0 の中核の仕様群（MIDI 1.0 Core Specifications）に入る |
| 何を決めているか | 機器の間で演奏の速さと位置を合わせるためのメッセージ。MIDI Clock は Timing Clock・Start・Continue・Stop と Song Position Pointer のメッセージ。MTC は Quarter Frame と Full の 2 種のメッセージ（と任意の SMPTE user bits のメッセージ）で時刻を送る |
| 一次情報 | MIDI Clock のメッセージは MIDI 1.0 Detailed Specification（https://midi.org/midi-1-0-detailed-specification ）の全メッセージの中核の配置に入る。ステータスバイトの表は https://midi.org/expanded-midi-1-0-messages-list でページの上で読める（表のダウンロードには参加（Join）が要る）。MTC は https://midi.org/midi-time-code 。どちらのページのダウンロードのボタンにも参加の表示はない（リンク先の Google Drive はプロキシに拒まれ、開いていない） |
| 状態 | 現行の MIDI 1.0 の一部。MTC は MIDI 1.0 Core Specifications のカテゴリにある |
| 当たるべき場面 | 外部の機器や DAW とテンポ・再生の位置を合わせるとき。Web MIDI API（演奏データの章）で外へ送ることを考えるとき |
| middy との接点 | 今はない。再生の時計は `AudioContext.currentTime`（`frontend/src/features/player/TimelinePlayer.ts`）で、外と合わせない。SMF の拍子のメタイベントは拍を MIDI クロックの数で書く（`backend/internal/music/smf/smf.go` の `timeSignatureBytes`。4 分音符を 24 クロックとして書く）。`docs/product/vision.md` の「やらないこと」に Web MIDI 出力がある |
| 確認日 | 2026-11-09 |

### Open Sound Control（OSC）

| 欄 | 内容 |
| --- | --- |
| 発行元 | opensoundcontrol.stanford.edu（OSC を CNMAT で作った Matt Wright と Adrian Freed による。サイトは編集者が保守する）。標準化の団体を名乗っていない |
| 何を決めているか | アプリや機器の間でリアルタイムにメッセージを送るための、伝送に依存しない符号化。アドレスは利用者が URL のような階層で決め、時刻のタグを持てる |
| 一次情報 | https://opensoundcontrol.stanford.edu/ 。仕様は https://opensoundcontrol.stanford.edu/spec-1_0.html 。登録なしで読める（サイトは CC BY） |
| 状態 | 形式の仕様は 1.0（2002 年）。1.1 は仕様の形では出ておらず、NIME 2009 の論文を 1.1 の考えの正式な記述とするとある（https://opensoundcontrol.stanford.edu/spec-1_1.html ） |
| 当たるべき場面 | 他のアプリや機器と、ネットワーク越しに制御や同期のメッセージをやりとりすることを考えるとき |
| middy との接点 | 今はない |
| 確認日 | 2026-11-09 |

### Ableton Link

| 欄 | 内容 |
| --- | --- |
| 発行元 | Ableton（会社）。標準化の団体の仕様ではなく、Ableton が配る技術とライブラリ |
| 何を決めているか | ローカルネットワーク上の複数のアプリで、拍・テンポ・位相と開始・停止を合わせる技術。参加するアプリは自動で見つけ合い、どれでもテンポを変えられる。Link Audio は、さらに音声のチャンネルを共有する |
| 一次情報 | 概念の説明は https://ableton.github.io/link/ 。ソースは https://github.com/Ableton/link （README は https://raw.githubusercontent.com/Ableton/link/master/README.md で開いた）。iOS 向けは LinkKit SDK（https://ableton.github.io/linkkit 。開いていない）。GPLv2+ と独自のライセンスの二重ライセンスで、独自のアプリに組み込むには Ableton に連絡するとある |
| 状態 | ライブラリは C++ のソースで配られている。Link Audio が Link を広げたものとしてある |
| 当たるべき場面 | 他のアプリや機器とテンポ・拍を合わせて一緒に鳴らすことを考えるとき |
| middy との接点 | 今はない。再生の時計は `AudioContext.currentTime`（`frontend/src/features/player/TimelinePlayer.ts`）で、外と合わせない |
| 確認日 | 2026-11-09 |

## Web・API・安全

### OpenAPI Specification（OAS）

| 欄 | 内容 |
| --- | --- |
| 発行元 | OpenAPI Initiative（spec.openapis.org。仕様のページの著作権表示は Linux Foundation） |
| 何を決めているか | HTTP API のインターフェースを、プログラミング言語に依らずに記述する形式。人と計算機が、ソースコードや通信を見ずに API の能力を知れるようにする |
| 一次情報 | https://spec.openapis.org/oas/latest.html （最新の公開版）。各版の一覧は https://spec.openapis.org/oas/ 。登録なしで読める。仕様は Apache License 2.0 |
| 状態 | 現行。最新の公開版は 3.2.1（2026-09-10）。版の patch の部分は機能を変えず、道具は 3.1.0 と 3.1.1 を区別しないとある |
| 当たるべき場面 | API の契約の書き方で迷うとき。`openapi` の版を上げるか考えるとき。型の生成や整合性テストの振る舞いが仕様どおりか確かめるとき |
| middy との接点 | `docs/api/openapi.yaml`（`openapi: 3.1.0`）が API の正典。Go の整合性テスト（`backend/internal/httpapi/openapi_test.go`）と、`npm run gen:api`（openapi-typescript）が作る `frontend/src/shared/api/schema.gen.ts` |
| 確認日 | 2026-11-09 |

### JSON Schema

| 欄 | 内容 |
| --- | --- |
| 発行元 | JSON Schema（json-schema.org のプロジェクト） |
| 何を決めているか | JSON の文書の構造を記述し、検証するための語彙。仕様は Core と Validation の 2 部に分かれる |
| 一次情報 | https://json-schema.org/specification 。Core、Validation、メタスキーマの各文書はこのページから行ける（各文書は開いていない）。登録なしで読める |
| 状態 | 現行は 2020-12。前の版は 2019-09 とある |
| 当たるべき場面 | LLM に渡す出力の形を変えるとき。`docs/api/openapi.yaml` のスキーマのキーワードの意味を確かめるとき |
| middy との接点 | `backend/internal/intent/schema.go` の `Schema` と `DiagnosticSchema`（LLM の構造化出力に渡す）。OAS のページに、Schema Object は既定で JSON Schema 2020-12 の上位集合とあるので、`docs/api/openapi.yaml` のスキーマもこれに当たる |
| middy との違い | `intent.Schema` は Messages API が受け付ける部分集合だけを使う（`maxItems` と `minimum`/`maximum` を使わず、`minItems` は 0 か 1、すべての object に `additionalProperties: false`）。部分集合で書けない制限は `description` に書き、`Validate` が強制する。`$schema` は宣言していない |
| 確認日 | 2026-11-09 |

### HTTP Caching（RFC 9111）

| 欄 | 内容 |
| --- | --- |
| 発行元 | IETF（RFC Editor が公開。STD 98） |
| 何を決めているか | HTTP のキャッシュと、キャッシュの振る舞いを制御する・応答がキャッシュできることを示すヘッダー（`Cache-Control` など） |
| 一次情報 | https://www.rfc-editor.org/rfc/rfc9111.html 。書誌の情報は https://www.rfc-editor.org/info/rfc9111 。無償 |
| 状態 | Internet Standard（STD 98）。2022 年 6 月の発行で、RFC 7234 を廃止して置き換えた。info のページに、これを置き換える RFC の表示はない |
| 当たるべき場面 | nginx が付ける `Cache-Control` を変えるとき。API の応答にキャッシュの指示を足すとき。Service Worker のキャッシュと HTTP のキャッシュの関係を考えるとき |
| middy との接点 | `deploy/nginx/templates/default.conf.template`（`/assets/` は `public, max-age=31536000, immutable`、アプリの殻・manifest・ほかのファイルは `no-cache`。ADR-0022、ADR-0027）。`scripts/smoke.sh` が `no-cache` を確かめる。`immutable` は RFC 9111 の本文には出てこない。Service Worker は API の応答をキャッシュしない（`frontend/vite.config.ts` の `workbox`） |
| 確認日 | 2026-11-09 |

### Problem Details for HTTP APIs（RFC 9457）

| 欄 | 内容 |
| --- | --- |
| 発行元 | IETF（RFC Editor が公開） |
| 何を決めているか | HTTP API のエラーの詳細を、機械が読める形で応答に入れるための「problem detail」。API ごとに新しいエラーの形を作らずに済むようにする |
| 一次情報 | https://www.rfc-editor.org/rfc/rfc9457.html 。書誌の情報は https://www.rfc-editor.org/info/rfc9457 。無償 |
| 状態 | Proposed Standard。2023 年 7 月の発行で、RFC 7807 を廃止して置き換えた。info のページに、これを置き換える RFC の表示はない |
| 当たるべき場面 | エラー応答の形を変えるとき。外部の API クライアントや道具とエラーをやりとりするとき |
| middy との接点 | `backend/internal/httpapi/json.go` の `writeError`、`docs/api/openapi.yaml` の `Error`、nginx が自分で答えるエラー（`deploy/nginx/templates/default.conf.template` の `@not_found` など） |
| middy との違い | middy のエラーは `{"error":{"code":…,"message":…}}` の共通エンベロープで、`Content-Type` は `application/json`。RFC 9457 は、メンバーを `type`・`status`・`title`・`detail`・`instance` とし、JSON では `application/problem+json` のメディアタイプを使うとある |
| 確認日 | 2026-11-09 |

### Content Security Policy（CSP）

| 欄 | 内容 |
| --- | --- |
| 発行元 | W3C Web Application Security Working Group |
| 何を決めているか | Web の開発者が、ページが取得・実行できる資源と、安全に関わるいくつかの方針を制御する仕組み |
| 一次情報 | https://www.w3.org/TR/CSP3/ （Level 3）。登録なしで読める。編集中の版は、このページが示す https://w3c.github.io/webappsec-csp/ （開いていない） |
| 状態 | Level 3 は Working Draft（2026-10-27）。W3C 勧告への過程（Recommendation track）にある。Level 2 を発展させたものとある |
| 当たるべき場面 | nginx に安全のためのヘッダーを足すとき。外部のスクリプト、フォント、接続先を足すとき |
| middy との接点 | 今はない。nginx の設定は `Content-Security-Policy` のヘッダーを付けていない（`deploy/nginx/templates/default.conf.template`） |
| 確認日 | 2026-11-09 |

### OWASP ASVS（Application Security Verification Standard）

| 欄 | 内容 |
| --- | --- |
| 発行元 | OWASP（ASVS Project） |
| 何を決めているか | Web アプリと Web サービスを設計・開発・テストするための、安全の要件を並べた開かれた標準 |
| 一次情報 | https://github.com/OWASP/ASVS の README（https://raw.githubusercontent.com/OWASP/ASVS/master/README.md で開いた）。PDF・Word・CSV の各版へのリンクがある。CC BY-SA 4.0 で無償。プロジェクトのページ https://owasp.org/www-project-application-security-verification-standard/ は https://asvs.owasp.org/ に転送され、この環境からは届かなかった。github.com のページもこの session からは開けなかった |
| 状態 | 最新の安定版は 5.0.0（2025 年 5 月）。master は開発中の版で、次は patch 版を目指すとある |
| 当たるべき場面 | 入力の検証、秘密情報、HTTP のヘッダーなど、安全の点検をするとき。外部に公開する前に要件を洗うとき。要件を引くときは、識別子が版で変わるので版を付けて書くよう勧めている |
| middy との接点 | 今はない（ASVS に照らした点検の記録はリポジトリにない） |
| 確認日 | 2026-11-09 |

## アクセシビリティと国際化

### WCAG（Web Content Accessibility Guidelines）

| 欄 | 内容 |
| --- | --- |
| 発行元 | W3C Accessibility Guidelines Working Group |
| 何を決めているか | Web のコンテンツを、障害のある人を含む幅広い人にとって使えるものにするための推奨（原則・ガイドライン・達成基準） |
| 一次情報 | https://www.w3.org/TR/WCAG22/ （https://www.w3.org/TR/WCAG/ も 2.2 を示す）。次の世代は https://www.w3.org/TR/wcag-3.0/ 。登録なしで読める |
| 状態 | 2.2 が W3C Recommendation（2024-12-12）。2.2 に適合するコンテンツは 2.0 と 2.1 にも適合するとある。3.0 は Working Draft（2026-09-10）で、WCAG 2 を置き換えないとある |
| 当たるべき場面 | 画面や操作を足す・変えるとき（コントラスト、キーボード、フォーカスの順序、見出しとラベル、状態のメッセージ、ページの言語など）。ux-principles の 6 章の確かめ方を行うとき |
| middy との接点 | `frontend/index.html` の `lang="ja"`。`docs/dev/conventions.md` のアクセシビリティの規約（`button`/`role`、`getByRole`）。`docs/design/ux-principles.md` の 6 章（アクセシビリティツリーで要素を数える、Tab の回数と見出しの構造）。ux-principles は見た目（色など）を扱わない |
| 確認日 | 2026-11-09 |

### WAI-ARIA

| 欄 | 内容 |
| --- | --- |
| 発行元 | W3C Accessible Rich Internet Applications Working Group |
| 何を決めているか | 支援技術にウィジェット・構造・振る舞いの意味を伝えるための、役割（role）・状態・属性の体系 |
| 一次情報 | https://www.w3.org/TR/wai-aria/ （1.2）。編集中の版は、このページが示す https://w3c.github.io/aria/ （1.3）。登録なしで読める |
| 状態 | 1.2 が W3C Recommendation（2023-06-06）。1.3 は Editor's Draft（2026-11-07） |
| 当たるべき場面 | `role` や `aria-*` の属性を足す・変えるとき。読み上げの量を確かめるとき（backlog の Ideas の、再生中の解説の読み上げ） |
| middy との接点 | `role="status"`・`role="alert"`（`frontend/src/app/App.tsx`、`frontend/src/features/projects/ProjectList.tsx` など）、`aria-live="polite"`（`frontend/src/features/explanation/ExplanationPanel.tsx`、`frontend/src/features/chat/ChatPanel.tsx`）、`aria-label`・`aria-current`（`frontend/src/shared/ui/DegreeMap.tsx`） |
| 確認日 | 2026-11-09 |

### Unicode の音楽記号（Musical Symbols、♭ ♮ ♯）

| 欄 | 内容 |
| --- | --- |
| 発行元 | Unicode Consortium（The Unicode Standard の符号表） |
| 何を決めているか | 音楽の記号の文字と符号位置。Musical Symbols のブロック（U+1D100–1D1FF）と、Miscellaneous Symbols のブロックの中の音楽記号（♩ ♪ ♫ ♬ ♭ ♮ ♯） |
| 一次情報 | 符号表の一覧は https://www.unicode.org/charts/ （Unicode 18.0 の一覧 https://www.unicode.org/Public/18.0.0/charts/ に転送される。転送は http を経るので、https の転送先を直接開いた）。Musical Symbols は https://www.unicode.org/Public/18.0.0/charts/PDF/U1D100.pdf 、Miscellaneous Symbols は https://www.unicode.org/Public/18.0.0/charts/PDF/U2600.pdf 。登録なしで読める。Musical Symbols の表は、よく使う臨時記号（♭ ♮ ♯）が Miscellaneous Symbols にあると示している |
| 状態 | 開いた符号表は Unicode 18.0 のもの。PDF は errata や文字の追加では更新しないとあり、最新の一覧は https://www.unicode.org/charts/ から引くよう示している |
| 当たるべき場面 | 臨時記号や音符の記号を画面や書き出しに使うとき。音名を ASCII の `b`・`#` で書くか ♭・♯ で書くか決めるとき |
| middy との接点 | 和音名と音名は ASCII の `b`・`#`（`Bb`、`C#4`。`frontend/src/shared/lib/notation.ts`、`docs/architecture/domain-model.md`）。ローマ数字の度数と解説の文は ♭（`♭VII`。`RomanNumeralAlt`）。分数コードのベースのラベルは ♭・♯ を `b`・`#` で書く（`docs/api/openapi.yaml`）。ピアノロールの音名の綴りは backlog の Ideas |
| 確認日 | 2026-11-09 |

### CLDR（Unicode Common Locale Data Repository）

| 欄 | 内容 |
| --- | --- |
| 発行元 | Unicode CLDR Project（Unicode, Inc.） |
| 何を決めているか | 言語や地域に合わせて表示するためのデータ（日付・数・単位の書式、並べ替えの順序など）を集めたリポジトリ |
| 一次情報 | https://cldr.unicode.org/ 。版の一覧とダウンロードは https://cldr.unicode.org/index/downloads 。データの形式の仕様（LDML）は、このページが示す https://www.unicode.org/reports/tr35/ （開いていない）。無償。利用条件は Unicode の Terms of Use とある |
| 状態 | 現行。最新の版は CLDR 49（2026-10-22）。版ごとにデータ・表・仕様がある |
| 当たるべき場面 | 日付・数・言語名などの表示を言語に合わせるとき。UI を日本語以外に広げることを考えるとき |
| middy との接点 | `frontend/src/features/projects/ProjectList.tsx` が `toLocaleString('ja-JP')` で日時を出す。CLDR のページに、多くの開発者は ICU などのライブラリを通して間接に CLDR を使うとある（ブラウザがどのデータを使うかは確かめていない）。UI の言語は日本語（`frontend/index.html` の `lang="ja"`） |
| 確認日 | 2026-11-09 |

## コンテナとビルド

### Dockerfile（BuildKit の Dockerfile フロントエンドの構文）

| 欄 | 内容 |
| --- | --- |
| 発行元 | Docker, Inc. と Moby プロジェクト（BuildKit。フロントエンドのイメージは `docker/dockerfile`） |
| 何を決めているか | コンテナイメージを組み立てる命令の書き方。パーサー指令（`syntax`・`escape`・`check`）、命令（`FROM`・`RUN`・`COPY` など）、BuildKit の拡張（`RUN --mount` の bind・cache・tmpfs・secret・ssh、`RUN --network`、ヒアドキュメント、`COPY --link`・`--parents`・`--exclude`）と、それぞれが要るフロントエンドの最小の版 |
| 一次情報 | https://docs.docker.com/reference/dockerfile/ （正典。登録なしで読める）。同じ文書の元は https://github.com/moby/buildkit/blob/master/frontend/dockerfile/docs/reference.md 。以前の `syntax.md`（例: https://github.com/moby/buildkit/blob/dockerfile/1.4.3/frontend/dockerfile/docs/syntax.md ）は master では「reference.md に移った」とだけある。版ごとの変更は https://github.com/moby/buildkit/releases （`dockerfile/` で始まるタグ。docs.docker.com の release notes のページはここへ転送される）。書き方の指針は https://docs.docker.com/build/building/best-practices/ |
| 状態 | 現行。最新の安定版は `dockerfile/1.28.0`（2026-10-07）。1.29.0 は rc。`syntax` 指令は「多くの利用者は `docker/dockerfile:1` にするとよい」（最新の安定版の 1.x を取ってくる）とある。`labs` の系列は、major を上げずに機能を外すことがある |
| 当たるべき場面 | `Dockerfile` を書く・変えるとき。`RUN --mount`、ヒアドキュメント、`COPY` の新しい旗を使う前に、要る最小の版を確かめるとき。`syntax` の版を決めるとき。ベースイメージの固定（タグかダイジェストか）を決めるとき |
| middy との接点 | `Dockerfile`（`# syntax=docker/dockerfile:1`。リファレンスの勧めどおり、最新の安定版の 1.x を使う。#408、`RUN --mount=type=cache`、多段ビルド、`--target web` と `api`。ADR-0027）。CI は hadolint で検査する（`docs/dev/harness.md`） |
| 確認日 | 2026-11-11 |

### Build checks（BuildKit の組み込みの検査）

| 欄 | 内容 |
| --- | --- |
| 発行元 | Docker, Inc.（BuildKit） |
| 何を決めているか | Dockerfile とビルドの設定を、決まった規則（`StageNameCasing`、`JSONArgsRecommended`、`UndefinedVar`、`SecretsUsedInArgOrEnv`、`CopyIgnoredFile` など）で検査する仕組み。`docker build --check` で、イメージを作らずに検査だけ走らせる。`# check=skip=…`・`# check=error=true` の指令で、外す規則と、失敗をエラーにするかを決める。既定では警告 |
| 一次情報 | https://docs.docker.com/reference/build-checks/ （規則の一覧。登録なしで読める）。指令の書き方は https://docs.docker.com/reference/dockerfile/ の `check` の節 |
| 状態 | 現行。規則は増えていく（一覧に experimental の印が付くものがある） |
| 当たるべき場面 | Dockerfile の検査を足す・変えるとき。hadolint と役割が重なるか確かめるとき |
| middy との接点 | 今はない。Dockerfile の静的な検査は hadolint だけ（`Makefile` の `dockerfile-lint`、CI の `shell, workflows and Dockerfile` のジョブ） |
| 確認日 | 2026-11-11 |

### Compose Specification

| 欄 | 内容 |
| --- | --- |
| 発行元 | Compose Specification のプロジェクト（compose-spec。Docker Compose がその実装） |
| 何を決めているか | 複数のコンテナからなるアプリを 1 つのファイル（`compose.yaml`）で定める形式。サービス、ネットワーク、ボリューム、設定、秘密情報の書き方 |
| 一次情報 | https://github.com/compose-spec/compose-spec の `spec.md`（https://raw.githubusercontent.com/compose-spec/compose-spec/main/spec.md で開いた）。Docker Compose の実装としての参照は https://docs.docker.com/reference/compose-file/ 。どちらも登録なしで読める |
| 状態 | 現行。旧い 2.x と 3.x の形式は Compose Specification に統合された。最上位の `version` は後方互換のためだけに残り、書くと廃止の警告が出る。既定のファイル名は `compose.yaml`（`compose.yml` も可。`docker-compose.yaml` は後方互換） |
| 当たるべき場面 | `compose.yaml` を書く・変えるとき。ローカルの検証の構成（`make up`）を変えるとき |
| middy との接点 | `compose.yaml`（`web` と `api` の 2 つのコンテナ。ADR-0027、ADR-0008）。`version` は書いていない |
| 確認日 | 2026-11-11 |
