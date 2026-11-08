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
3. `intent.Parse` が復号と検証を行う。不正なら、その答えと検証エラーを添えて **1 回だけ** 再試行し、それでも不正なら `ErrInvalidOutput`
4. 候補ごとに `Candidate.ToProgression` → `music.FromProgression` → `music.Analyze` で分析済み `Timeline` を作る
5. JSON で返す。フロントは候補カードを表示し、選択されたものを `player` が再生、`timeline` が描画、`explanation` が解説
6. 相談（チャット）は会話履歴（ユーザー発言と、assistant 側は前回の `summary`）を往復させ、同じ流れを繰り返す。サーバーは状態を持たない
7. `LLM_PROVIDER=fake` のとき、台本のない fake は入力の気分語（雨・悲・夜…）から決定論的な見本 Intent を返すので、鍵なしで開発・E2E ができる

`POST /api/v1/timelines/from-progression`（vision の S1、土台の能力）は 1〜3 と 6 を通らず、調・ローマ数字の進行・テンポを直接 4 に渡す。画面からは使わない（入口はイメージだけ。`docs/product/experience.md`）。

## バックエンドの層

```
cmd/server          配線のみ（config → deps → http.Server）
internal/httpapi    HTTP 境界。decode / validate / encode。ビジネスロジック禁止
internal/ideas      アプリケーションサービス（S2）: 会話 → LLM → intent → music。1 回だけ再試行
internal/music      ドメイン。純粋関数、標準ライブラリのみ、I/O なし
internal/intent     LLM が返す MusicalIntent の型・検証・JSON Schema（music にのみ依存）
internal/music/smf  Timeline → Standard MIDI File のエンコーダ（music にのみ依存）
internal/music/pattern  伴奏・ドラムの 1 小節テンプレートと、Harmony からトラックを作り直す Apply（music にのみ依存）
internal/llm        ポート（interface）。adapters: fake, anthropic（公式 SDK）
internal/web        ビルド済みフロント（frontend/dist）の embed。未同梱なら Dist() が false
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
- 再生は `features/player`: `TimelinePlayer`（先読みスケジューラ、Web Audio 非依存でテスト可能）が `Synth` と `Clock` を注入され、`WebAudioSynth` が音を出す（ch.9 はキック・スネア・ハイハットの簡易打楽器音）。`AudioContext` は最初の再生操作で作る。
- 表示は `features/timeline`: `TimelineView` がコードレーン（`listbox`/`option`、クリックで選択）とピアノロール、ch.9 のノートだけを集めたドラムレーンを SVG で描き、`positionTick` に応じて鳴っているコード・ノート・打楽器をハイライトする。1 小節 = `BAR_WIDTH` px。
- 会話は `features/chat`: `useIdeas` が履歴（ユーザー発言と assistant の `reply`）を保持して `POST /api/v1/ideas` を呼び、`ChatPanel` がメッセージと候補カードを表示する。答えが届くと `useIdeas` の `onAnswer` で最初の候補が開き、別の候補を開くと `App` が `Timeline` を差し替え、再生・表示・解説の部品に載る。`usePatterns` が伴奏・ドラムのカタログを取得し、候補カードに名前を出す。
- `src/app/App.tsx` が画面 1 枚の結線: 候補または保存したプロジェクトから開いた `Timeline` を 1 つの状態に持ち、再生位置と固定中のコードを player / timeline / explanation に配る。
- キーの変更は `features/key-change`: `rebuildRequest` が開いているタイムライン（ローマ数字・拍数・テンポ・モード・リズムの id とノリ）から `POST /api/v1/timelines/from-progression` のリクエストを組み立て、主音だけを替える。LLM は呼ばない。ローマ数字が同じ和音に読み戻せることは `music` のテストが保証する。
- 保存は `features/projects`: `useProjects` が `shared/lib/storage/projects` の `ProjectStore`（ブラウザに IndexedDB があればそれ、なければメモリ）を使い、開いている例があるときだけ結果のツールバーに `SaveProjectForm` が出て、タイムラインとその出どころ（イメージと候補名。以前の版で保存したものは進行の入力）を名前つきで保存する。`ProjectList` は保存したものを折りたたんで一覧し、開く・削除を受け持つ。サーバーには何も保存しない。
- 書き出しは `features/export`: `DownloadMidiButton` が `POST /api/v1/export/midi` の Blob を、調と進行を含むファイル名で保存する。
- 解説は `features/explanation`: `ExplanationPanel` が固定中のコード、なければ再生位置のコードの `analysis`（機能・終止・解説文）を表示し、`timeline.rhythm` があれば `RhythmNotes` が伴奏・ドラムの説明と打点の分析を出す。文章はバックエンドの決定論的な分析をそのまま出し、フロントで生成しない。
- 描画はまず SVG。ピアノロールの性能問題が出たら Canvas を検討する。
- スタイルはプレーン CSS（CSS Modules 可）。UI ライブラリは導入しない。

## PWA

- `vite-plugin-pwa` で manifest と Service Worker を生成。アプリシェルのみプリキャッシュし、`/api/*` はキャッシュしない。
- Service Worker は `prompt` モード: 新しい版は待機し、ユーザーがバナーの「更新して再読み込み」を押したときだけ切り替わる（作業中の再読み込みを防ぐ）。
- `shared/lib/pwa` の `PwaStatus` が SW の状態・インストールの申し出・オンライン状態を持つ小さな外部ストアで、`main.tsx` の `wirePwa` が SW 登録と `window` のイベントをつなぐ。`features/pwa` の `PwaBanner` が「オフラインです」「新しいバージョンがあります」「オフラインでも開けるようになりました」「インストール」を表示する。
- オフラインの価値は「保存済みプロジェクトの再生」。保存は IndexedDB（`features/projects`）、再生はブラウザ内で完結するので、ネットワークなしで開いて鳴らせる（E2E `offline.spec.ts` が検証）。生成と AI の提案はオンラインが必要。

## 配備形態（ADR-0008）

- **単一の Go バイナリ** が API と UI を同一オリジンで配信する。`make build` が `vite build` → `backend/internal/web/dist` へコピー → `go build`（`-ldflags` でバージョン埋め込み）の順に行う。
- `httpapi` の静的ハンドラ: `/api` 配下以外で、ファイルがあればそれを、なければ `index.html`（SPA フォールバック）。`/assets/*`（ハッシュ付き）は `immutable` で長期キャッシュ、`index.html` / `sw.js` / manifest は `no-cache`。未知の `/api/*` は JSON の 404。
- フロント未同梱のバイナリでも API は動き、UI は 503 `frontend_not_built` を返す。開発時は `make dev-frontend`（Vite のプロキシ）を使う。
- **コンテナ**: `Dockerfile` はマルチステージ（`node` でフロント → `golang` でバイナリ → `distroless/static` 非 root）。`HEALTHCHECK` はシェルのない distroless のため、バイナリの `-healthcheck` フラグが自分の `/api/v1/health` を叩く。`compose.yaml` は `app` 1 サービスで、`LLM_PROVIDER` と鍵を、起動するコマンドの環境変数だけから渡す（ADR-0011）。CI の `image` ジョブが同じ Dockerfile をビルドし、ヘルス・`/`・`sw.js` のヘッダ・`HEALTHCHECK` の healthy をスモークする。

## 横断的関心事

| 関心事 | 方針 |
| --- | --- |
| エラー | API は常に `{"error":{"code","message"}}`。Go 側は `errors.Join` で検証エラーをまとめる |
| ログ | `log/slog` JSON。リクエストログはミドルウェアで 1 行 |
| 設定 | 環境変数のみ。`config.Load(getenv)` で全部読む。既定値で起動できる |
| 秘密情報 | `ANTHROPIC_API_KEY` は環境変数。ログに出さない |
| 認証 | なし（アカウントは Ideas）。導入時は httpapi のミドルウェアとして追加 |
| 可観測性 | M4 以降。`slog` のハンドラ差し替えで対応できる形にしておく |

## 将来の検討事項（ADR 化の候補）

- 音楽エンジンの一部をブラウザでも使いたくなった場合（即時フィードバック）: TinyGo/WASM、または TS への限定的な移植。ADR-0002 の再検討条件を参照
- 会話状態のサーバー側保持（現状はステートレスに往復）
- 永続化先（IndexedDB → サーバー）
