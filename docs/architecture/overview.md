# アーキテクチャ概要

## 全体像

```
┌──────────────────────────── ブラウザ (PWA) ────────────────────────────┐
│  React + TypeScript                                                    │
│  features/chat ── features/timeline ── features/player ── explanation  │
│         │                 │                 │                          │
│         └──── shared/api（fetch, OpenAPI 準拠の型） ─────┘              │
│                           │   Web Audio（再生はブラウザ内で完結）       │
└───────────────────────────┼────────────────────────────────────────────┘
                            │ HTTPS  /api/v1/*  (JSON)
┌───────────────────────────┼─── Go サーバー ────────────────────────────┐
│  httpapi（薄いハンドラ） ──┼─→ ユースケース ──→ music（理論エンジン）   │
│                           │         │                                  │
│                           │         └──→ llm.Client ──→ Anthropic API  │
│                           │              （fake で差し替え可能）        │
└────────────────────────────────────────────────────────────────────────┘
```

- **音符を作るのは常に `music` パッケージ。** LLM は `MusicalIntent`（構造化 JSON）を返すだけ（ADR-0003）。
- **再生はブラウザ内。** サーバーはオーディオを扱わない。`Timeline` JSON を渡し、フロントの Web Audio シンセが鳴らす。
- **API 契約は `docs/api/openapi.yaml`。** フロントの型とハンドラはこれに追従する。

## データフロー（M2 時点の想定）

1. ユーザーがイメージ文を入力 → `POST /api/v1/ideas`
2. `ideas.Service.Imagine` が `llm.Client.Complete` を構造化出力（`intent.Schema`）つきで呼ぶ
3. `intent.Parse` が復号と検証を行う。不正なら、その答えと検証エラーを添えて **1 回だけ** 再試行する（エラー文は直し方を添える。1 文字消せば読めるラベルにはその候補、別の拍子の長さや型にはその拍子。#133）。それでも不正なら `ErrInvalidOutput`
4. 候補ごとに `Candidate.ToProgression` → `music.FromProgression` → `music.Analyze` で分析済み `Timeline` を作る。返事と候補の理由に書かれた和音の主張（「ローマ数字（和音名）」と「X へ解決する・進む」）と、理由に書かれたベースの並び（「ベースが F→E→D→C」。和音のベースか伴奏の最低音の並びに、ループとして現れるか。`ideas/bass.go`）を、その候補と照合し（`ideas/claims.go`、ADR-0034）、合わなければ 3 と同じく 1 回だけ再試行する。再試行した答えもまだ合わなければ、その文だけを除いて返す
5. JSON で返す。フロントは候補カードを表示し、選択されたものを `player` が再生、`timeline` が描画、`explanation` が解説
6. 相談（チャット）は会話履歴（ユーザー発言と、assistant 側は前回の `summary` に続く各候補の要点。BL-310）を往復させ、同じ流れを繰り返す。サーバーは状態を持たない
7. `LLM_PROVIDER=fake` のとき、台本のない fake は入力の気分語（雨・悲・夜…）から決定論的な見本 Intent を返すので、鍵なしで開発・E2E ができる

`POST /api/v1/timelines/from-progression`（vision の S1、土台の能力）は 1〜3 と 6 を通らず、調・ローマ数字の進行・テンポを直接 4 に渡す。画面からは使わない（入口はイメージだけ。`docs/product/experience.md`）。

## バックエンドの層

```
cmd/server          配線のみ（config → deps → http.Server）
cmd/converse        会話をまたいだ答えの計測（所有者の手元で。CI には入れない。#453）
internal/httpapi    HTTP 境界。decode / validate / encode。ビジネスロジック禁止
internal/ideas      アプリケーションサービス（S2）: 会話 → LLM → intent → music。1 回だけ再試行
internal/music      ドメイン。純粋関数、標準ライブラリのみ、I/O なし
internal/intent     LLM が返す MusicalIntent の型・検証・JSON Schema（music にのみ依存）
internal/music/smf  Timeline → Standard MIDI File のエンコーダ（music にのみ依存）
internal/music/pattern  伴奏・ドラムの 1 小節テンプレートとスタイルのカタログ（data/*.json、ADR-0035）と、Harmony からトラックを作り直す Apply（music にのみ依存）
internal/llm        ポート（interface）。adapters: fake, anthropic（公式 SDK）
internal/config     環境変数
```

依存の向きは上から下へのみ。`music` は何にも依存しない。`intent` は `music` に依存する（ローマ数字の検証）。`llm` はどちらにも依存しない（プロバイダ境界のみ）。逆方向は禁止。

### なぜフレームワークを使わないか

Go 1.22+ の `net/http` はメソッド付きパターンルーティングを持ち、middy の API 規模（十数エンドポイント）には十分。依存が少ないほど AI エージェントが推測で誤った API を書くリスクが減り、アップグレード作業も減る。必要になったら ADR を書いて導入する。

## フロントエンドの層

```
src/app        アプリシェル、ルーティング（必要になったら）、グローバル CSS
src/features   機能ごとの縦スライス。各 feature は
               components/ hooks/ state/ を自己完結させ、他 feature を import しない
src/shared     api/（fetch クライアント、OpenAPI から生成した型 schema.gen.ts）、ui/（汎用コンポーネント）、lib/（純粋ユーティリティ: timing, notation, storage, pwa）
src/test       Vitest セットアップ
```

- 状態管理ライブラリは **まだ入れない**。`useState`/`useReducer` + context で始め、feature をまたぐ状態（現在のタイムライン、再生位置）が 3 つを超えたら ADR を書いて検討する。
- 再生は `features/player`: `TimelinePlayer`（先読みスケジューラ、Web Audio 非依存でテスト可能）が `Synth` と `Clock` を注入され、`WebAudioSynth` が音を出す（和音のトラックは GM の音色から鍵盤・撥弦・オルガン・弦・笛の系統を選んで、波形と減衰を変える。`instrumentVoice`。リアルさは追わない。ch.9 は簡易打楽器音。キックとタムは音程が下がる太鼓、サイドスティックは短いクリック、スネアはノイズと音、ハイハットとシンバルは長さと帯域の違うノイズ（閉じたハイハットは短く明るく、足のハイハットは短く暗く、オープンは長めに、クラッシュは長く、ライドは「チン」という音を重ねる。#404 の S4）、それ以外は閉じたハイハットと同じ音。どれで鳴らすかは `drumVoice` が決める）。`AudioContext` は最初の再生操作で作る。再生は途中から始められる（`play(fromTick)`、再生中の `seek`）。始めた位置より前に鳴り始めて伸びている音は残りの長さで鳴らし、打楽器は打ち直さない。`pause` はその場で止め、`stop` と鳴り終わりは先頭へ戻る（BL-327、ADR-0070）。
- 表示は `features/timeline`: `TimelineView` がコードレーン（`listbox`/`option`、クリックで選択。名前は見えている 2 行のとおり `V7 B7`。押した和音は、鳴り始めの位置から枠までを灰色のつまみで示す。分析の時刻は小節線のまま。#385）とピアノロール（支援技術からは隠す）、ch.9 のノートだけを集めたドラムレーンを SVG（`role="group"`、名前は「タイムライン」）で描き、`positionTick` に応じて鳴っているコード・ノート・打楽器をハイライトする（`playing` が偽のあいだはハイライトしない）。再生中に描き直すのは動くものだけで、ピアノロール・ドラムレーン・コードレーンは鳴っている音や和音が変わったときだけ、小節線・ルーラー・グリッドは配置が変わったときだけ描く（`useSounding`、`memo`。#500）。1 小節 = `BAR_WIDTH` px。 長いタイムライン（4 小節を超える）は `TimelineOverview` が、全体を 1 枚の `TimelineView` で描いて横スクロールの箱（`.timeline-overview__scroll`）に入れ、その下に全体の縮小図の帯（`stretch`、支援技術からは隠し、コードはタブで止まらない）を置く（#360）。帯は今見えている範囲を枠で示す。箱のスクロールは `useTimelineScroll` が持つ。再生中は再生位置を、行名の列を除いた見えている絵の中央に保つ（`scrollFor` の `centre`。動きを減らす設定〔`useReducedMotion`〕ではページめくりの `page`。ADR-0066）。再生中に手でスクロールすると（フックが置いた位置から離れると）追随を止め、「再生位置へ戻る」を出す（`paused`、`resume`。次の再生でも戻る。ADR-0068）。停止中は固定した和音を中央に置き、固定がなければ動かさない。本体の絵の和音以外を押すと、押した拍の頭（`beatAt`）へ再生位置を動かす（`useSeekByClick` → `usePlayer.seek`。マウスとペンでは、動く先を破線で示し、押した瞬間に動かして、押したまま動かすとついてくる。再生中はそのあいだ一時停止し、離すと続きから鳴らす〔`app/useScrub`〕。指は線のまわりの帯〔`useHeadGrip`〕をつかむと、線を中央に保って譜面を引く。どのドラッグも最初のポインタだけを追ってほかの指は無視し、ポインタを失ったとき〔`lostpointercapture`〕や絵が外れたときにも終える。別の例を開いたあとに終わったドラッグは、始めた例を鳴らし直さない〔#500〕。ADR-0067、ADR-0071、ADR-0072）。縮小図の帯は見る場所を選ぶ道具で、押すと箱をそこへスクロールし、枠（今見えている範囲）をドラッグすると箱が連動する（`useStripDrag` → `look`。再生位置は動かさず、固定もしない。再生中は追随を一時停止する。ADR-0069）。キーボードの同等手段は、和音を固定したときに解説の枠に出る「ここから再生」（`ChordFrame` の `onPlayFrom`。固定した和音の頭へシークして鳴らす）。箱の高さは、4 小節が箱の幅に収まる大きさで、画面の 45 % まで（ADR-0041）。 画面の配置は wide（所有者の判断、#279）。広い画面では会話を左の列に置き、結果をその右に置く。狭い画面では縦に並べる（しきい値は `docs/dev/conventions.md` の「配置」の表）。横向きで高さの低い画面では、例を開くと会話を引き出しにたたむ（`app/useChatDrawer`、`ChatDrawer.tsx`。ADR-0073）。左の列は全体でスクロールし、入力欄はやり取りのすぐ下に置く。やり取りが短いうちは上に寄り、列からあふれたら列の下に張り付く（BL-222）。ページは 1 画面の高さで、会話と結果はそれぞれの列の中でスクロールし、入力欄は列の下に見えたままにする（BL-184、#279）。和音 1 つの解説は、大きさが固定の枠（`ChordFrame`）に出す。再生に追随しているか固定しているかは枠の縁の色と、コードレーンの ▶／■（`marks`）で示し、←／→ で前後の和音へ移って固定し、Esc で追随に戻る（`app/useChordKeys`）。読み上げる（`aria-live`）のは和音名と機能の 1 行だけ（BL-185）。左の列では、前の往復の例を 1 行ずつ（名前・調・BPM・進行、往復の番号つき）並べ、最新の往復だけカードで出す（`features/chat/ExampleList`）。1／2／3 で最新の答えの例を開く（`app/useCandidateKeys`）。例ごとに固定した和音を覚えていて、前の例を開くとその和音が固定された状態で戻る（`app/usePinMemory`。再生は例を替えると止まり、頭から始まる）（BL-186）。タイムラインは結果の部分の横いっぱいに描き（高さに上限があり、縦横比を保つ。ADR-0041）、和音の解説の枠と、ほかの解説を並べる（BL-197、#314）。結果の幅が広ければ、枠をタイムラインの上の帯にし（`order`）、ほかの解説を下に 1 列で並べる。狭ければ、どちらもタイムラインの下に並べる（結果を container にした container query。#353、#360、ADR-0044）。利用者が候補を選ぶ（カード、前の例の行、1／2／3）と開いて頭から鳴らし、鳴っている例をもう一度選ぶと一時停止する。Space でも再生と一時停止を切り替える。会話の入力欄の下の「候補を選んだら再生」で切れて、選んだ値はこのブラウザに覚える（`app/useChooseToHear`、#315）。答えが届いたときは鳴らさない。
- 会話は `features/chat`: `useIdeas` が履歴（ユーザー発言と、assistant の `reply` と各候補の要点。`history.ts`）を保持して `POST /api/v1/ideas` を呼び、`ChatPanel` がメッセージと候補カードを表示する。答えが届いても例は自動では開かず、ユーザーが候補を開くと `App` が `Timeline` を差し替え、再生・表示・解説の部品に載る（BL-219）。`usePatterns` が伴奏・ドラムのカタログを取得し、候補カードに名前を出す。
- `src/app/App.tsx` が画面 1 枚の結線: 候補または保存したプロジェクトから開いた `Timeline` を 1 つの状態に持ち、再生位置と固定中のコードを player / timeline / explanation に配る。再生位置は再生中に毎秒何度も変わるので、`App` の状態にしない。`usePlayer` の `position`（購読する値、`features/player/position`）にし、位置を描く部品（`PlayerControls` と `app/Stage`）だけが `usePosition` で読む。再生中に描き直すのはそこから下だけで、会話の列やツールバーのほかの部品は描き直さない（ADR-0077）。
- キーの変更は `features/key-change`: `shared/lib/rebuildRequest` の `rebuildRequest` が開いているタイムライン（ローマ数字・拍数・押し・テンポ・モード・リズムの id かドラムの格子とノリ）から `POST /api/v1/timelines/from-progression` のリクエストを組み立て、主音だけを替える。LLM は呼ばない。ローマ数字が同じ和音に読み戻せることは `music` のテストが保証する。
- キーとスタイルの変更は `app/useExampleChanges` がまとめる。変更が返る前に別の例を開いたら、その結果は捨てる（古い例で新しい例を上書きしない）。
- スタイルの選び直しは `features/regroove`: 型のカタログが各スタイルの拍子ごとのグルーヴ（`pattern.Style.Groove`。その拍子で使える最初の伴奏・ドラム・ノリ）を返し、`regrooveRequest` が和声をそのままにそのグルーヴでリクエストを組み立てる。`grooveChanges` が前後で変わった部分（伴奏・ドラム・ノリ・楽器）だけを述べる。LLM は呼ばない。
- 保存は `features/projects`: `useProjects` が `shared/lib/storage/projects` の `ProjectStore`（ブラウザに IndexedDB があればそれ、なければメモリ）を使い、開いている例があるときだけ、結果の上の題名の横に `KeepToggle`（☆／★）が出る。押すと、タイムラインとその出どころ（イメージと候補名。以前の版で保存したものは進行の入力）を名前つきで保存し、もう一度押すと外す。イメージから開いた例は、候補のカード（`ProjectSource.candidate`）も一緒に残し、とっておいた例を開いたときも解説の先頭に出す（BL-299。前に残した例にはない）。とっておいたかどうかは、同じタイムラインを持つものがあるかで決める（`keptAs`）。`ProjectList` はとっておいたものを折りたたんで一覧し、開く・名前を変える（`ProjectName`。Enter か欄を離れると決まり、Esc と空の名前は元のまま）・削除を受け持つ。サーバーには何も保存しない。
- 書き出しは `features/export`: `DownloadMidiButton` が `POST /api/v1/export/midi` の Blob を、調と進行を含むファイル名で保存する。
- 解説は `features/explanation`: `ExplanationPanel` が固定中のコード、なければ再生位置のコードの `analysis`（機能・終止・解説文）を表示し、`timeline.rhythm` があれば `RhythmNotes` が伴奏・ドラムの説明と打点の分析を出す。例ごとに 1 回読む「この例」と「リズム」は、`shared/ui/Fold`（`details`）で個別にたたむ（ADR-0075）。度数の地図（`shared/ui/DegreeMap`、`Overview.degrees` を並べ、解説中の和音を光らせる。BL-099）は、結果の画面では 1 和音の解説の枠（`ChordFrame`）の中に置く（#360、ADR-0044）。長調と短調では `flow` で機能の列（T → S → D）に並べる。モードの例では（`shared/lib/notation` の `isFunctional`）、解説の機能の札、読み上げの機能名、地図の機能の色を出さず、地図の主和音に印を付ける（`functions`、ADR-0045）。代わりに `roles` で地図を役割の列（`DegreeChord.role`）に分け、解説の札と読み上げを役割の言葉（`notation` の `roleName`）にする（ADR-0048）。地図は置き場所を持たない部品なので、ほかの画面にも置ける（`ExplanationPanel` が 1 和音の解説も出すときは、パネルの中に出す）。文章はバックエンドの決定論的な分析をそのまま出し、フロントで生成しない。例外は先頭の「この例（AI の提案）」の段で、開いている候補のカードの中身（AI の提案）を `features/chat` の `ExampleNotes` が組み、`app` が `ExplanationPanel` の `example` に差し込む。エンジンの文とは線で分ける（BL-298）。右上の `CopyButton`（`useCopyText`）は、`app` が組んだ文字（カードの `cardLines` と、`explanationLines` のコード名・全体・流れ・リズム）をクリップボードに写す（BL-300）。
- 描画はまず SVG。ピアノロールの性能問題が出たら Canvas を検討する。
- スタイルはプレーン CSS（CSS Modules 可）。UI ライブラリは導入しない。

## PWA

- `vite-plugin-pwa` で manifest と Service Worker を生成。アプリシェルのみプリキャッシュし、`/api/*` はキャッシュしない。
- Service Worker は `prompt` モード: 新しい版は待機し、ユーザーがバナーの「更新して再読み込み」を押したときだけ切り替わる（作業中の再読み込みを防ぐ）。開いたままのアプリも、オンラインのあいだは 1 時間ごとに新しい版を探す（`shared/lib/pwa/wire` の `CHECK_EVERY_MS`、#500）。
- `shared/lib/pwa` の `PwaStatus` が SW の状態・インストールの申し出・オンライン状態を持つ小さな外部ストアで、`main.tsx` の `wirePwa` が SW 登録と `window` のイベントをつなぐ。`features/pwa` の `PwaBanner` が「オフラインです」「新しいバージョンがあります」「インストール」を表示する。オフラインで使える準備ができたことは正常な状態なので、何も言わない（原則 7、BL-071）。
- オフラインの価値は「保存済みプロジェクトの再生」。保存は IndexedDB（`features/projects`）、再生はブラウザ内で完結するので、ネットワークなしで開いて鳴らせる（E2E `offline.spec.ts` が検証）。生成と AI の提案はオンラインが必要。

## 配備形態（ADR-0027）

`web`（nginx とビルド済みの SPA）と `api`（Go）の 2 つのコンテナを、同一オリジンの 1 つのサービスとして配る。外に開くのは `web` だけで、`/api` を `api` に渡す。

- **Go は API だけを答える。** `/api/v1` のルート以外は、画面の URL も含め、メソッドによらず JSON の 404（`not_found`）。既知のパスに別のメソッドが来たときは 405。
- **配信の規則は nginx が持つ**（`deploy/nginx/templates/`）。ファイルがあればそれを返す。なければ、画面の URL には `index.html`（SPA フォールバック）、ファイルを指す URL（`/assets/*` や拡張子のあるもの）には JSON の 404 を返す（ADR-0022 の規則）。`/assets/*`（ハッシュ付き）は `immutable` で長期キャッシュ、`index.html` / `sw.js` / manifest は `no-cache`。すべての応答に、セキュリティヘッダと、同じオリジンだけを許す CSP を付ける（`deploy/nginx/snippets/security-headers.conf`、ADR-0040）。
- 開発時は `make dev-backend` と `make dev-frontend`（Vite が `/api` を `API_UPSTREAM` へプロキシ）。`make build` は `vite build` と `go build`（`-ldflags` でバージョン埋め込み）をそれぞれ行い、`make run` は API だけを起動する。
- **コンテナ**: `Dockerfile` から 2 つのイメージを `--target` で作る。
  - `web`: `nginx-unprivileged` とビルド済みの SPA。`deploy/nginx/templates/` を起動時に `PORT` と `API_UPSTREAM`（URL）で埋め、上の配信の規則に gzip と `/healthz` を足して配り、`/api` を Go へ渡す（読み取りの待ち 180 s、本文 4 MiB まで）。
  - `api`: Go のバイナリを `distroless/static` 非 root で。`HEALTHCHECK` はバイナリの `-healthcheck` フラグが自分の `/api/v1/health` を叩く。
  - `compose.yaml` は `web`（`WEB_PORT`、既定 8080 で公開）と `api`（`PORT=8081`、公開しない）。`LLM_PROVIDER` と鍵は `api` にだけ、起動するコマンドの環境変数から渡す（ADR-0011）。CI の `image` ジョブと `e2e` ワークフローが 2 つのイメージを作り、compose で起動して両方の healthy を待ち、`scripts/smoke.sh` を当てる（`e2e` はさらに Playwright を当てる）。

## 横断的関心事

| 関心事 | 方針 |
| --- | --- |
| エラー | RFC 9457 の problem detail（`application/problem+json`。`type`・`title`・`status`・`detail` と拡張メンバー `code`）。種類の一覧は `docs/api/problems.md`（ADR-0030）。Go 側は `errors.Join` で検証エラーをまとめる |
| ログ | `log/slog` JSON。リクエストログはミドルウェアで 1 行（状態と所要時間）。ハンドラの panic は同じミドルウェアが problem detail（500 `internal_error`）に変えて記録する。`/ideas` が LLM の失敗で 502 を返すときは、原因のエラーを `WARN` で 1 行残す（応答はコードだけ） |
| 設定 | 環境変数のみ。`config.Load(getenv)` で全部読む。既定値で起動できる |
| 秘密情報 | `ANTHROPIC_API_KEY` は環境変数。ログに出さない（アダプタのエラー文に鍵が入らないことをテストで確かめる） |
| 認証 | なし（アカウントは Ideas）。導入時は httpapi のミドルウェアとして追加 |
| 可観測性 | M4 以降。`slog` のハンドラ差し替えで対応できる形にしておく |

## 将来の検討事項（ADR 化の候補）

- 音楽エンジンの一部をブラウザでも使いたくなった場合（即時フィードバック）: TinyGo/WASM、または TS への限定的な移植。ADR-0002 の再検討条件を参照
- 会話状態のサーバー側保持（現状はステートレスに往復）
- 永続化先（IndexedDB → サーバー）
