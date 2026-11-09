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
