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
2. ユースケースが `llm.Client.Complete` を構造化出力つきで呼び、`MusicalIntent` を得る
3. `intent.Validate()` で理論的に成立する値だけを通す（不正なローマ数字などは弾く／補正）
4. `music.FromProgression(...)` で候補ごとに `Timeline` を生成、`music.Analyze` で解説データを付与
5. JSON で返す。フロントは候補カードを表示し、選択されたものを `player` が再生、`timeline` が描画、`explanation` が解説
6. 相談（チャット）は会話履歴を往復させ、同じ流れを繰り返す

M1 では 1〜3 と 6 がなく、ユーザー入力（調・ローマ数字の進行・テンポ）が `POST /api/v1/timelines/from-progression` から直接 4 に入る。

## バックエンドの層

```
cmd/server          配線のみ（config → deps → http.Server）
internal/httpapi    HTTP 境界。decode / validate / encode。ビジネスロジック禁止
internal/<usecase>  アプリケーションサービス。LLM と music を組み合わせる段階（BL-012）で追加する。それまでハンドラが music を直接呼ぶ
internal/music      ドメイン。純粋関数、標準ライブラリのみ、I/O なし
internal/intent     LLM が返す MusicalIntent の型・検証・JSON Schema（music にのみ依存）
internal/music/smf  Timeline → Standard MIDI File のエンコーダ（music にのみ依存）
internal/llm        ポート（interface）。adapters: fake（実装済み）, anthropic（BL-011）
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
src/shared     api/（fetch クライアントと型）、ui/（汎用コンポーネント）、lib/（純粋ユーティリティ: timing, notation）
src/test       Vitest セットアップ
```

- 状態管理ライブラリは **まだ入れない**。`useState`/`useReducer` + context で始め、feature をまたぐ状態（現在のタイムライン、再生位置）が 3 つを超えたら ADR を書いて検討する。
- 再生は `features/player`: `TimelinePlayer`（先読みスケジューラ、Web Audio 非依存でテスト可能）が `Synth` と `Clock` を注入され、`WebAudioSynth` が音を出す。`AudioContext` は最初の再生操作で作る。
- 表示は `features/timeline`: `TimelineView` がコードレーン（`listbox`/`option`、クリックで選択）とピアノロールを SVG で描き、`positionTick` に応じて鳴っているコード・ノートをハイライトする。1 小節 = `BAR_WIDTH` px。
- 入力は `features/progression-input`: `ProgressionForm` がテキスト（`I - V - vi - IV`）をトークン化し、API リクエストを組み立てる。ローマ数字の妥当性はバックエンドが判定する。
- `src/app/App.tsx` が S1 シナリオの結線: フォーム → API → `Timeline` を状態に持ち、再生位置と固定中のコードを player / timeline / explanation に配る。
- 書き出しは `features/export`: `DownloadMidiButton` が `POST /api/v1/export/midi` の Blob を、調と進行を含むファイル名で保存する。
- 解説は `features/explanation`: `ExplanationPanel` が固定中のコード、なければ再生位置のコードの `analysis`（機能・終止・解説文）を表示する。文章はバックエンドの決定論的な分析をそのまま出し、フロントで生成しない。
- 描画はまず SVG。ピアノロールの性能問題が出たら Canvas を検討する。
- スタイルはプレーン CSS（CSS Modules 可）。UI ライブラリは導入しない。

## PWA

- `vite-plugin-pwa` で manifest と Service Worker を生成。アプリシェルのみプリキャッシュし、`/api/*` はキャッシュしない。
- オフラインの価値は「保存済みプロジェクトの再生」。これは M3 のローカル永続化（IndexedDB）で実現する。

## 配備形態（ADR-0008）

- **単一の Go バイナリ** が API と UI を同一オリジンで配信する。`make build` が `vite build` → `backend/internal/web/dist` へコピー → `go build`（`-ldflags` でバージョン埋め込み）の順に行う。
- `httpapi` の静的ハンドラ: `/api` 配下以外で、ファイルがあればそれを、なければ `index.html`（SPA フォールバック）。`/assets/*`（ハッシュ付き）は `immutable` で長期キャッシュ、`index.html` / `sw.js` / manifest は `no-cache`。未知の `/api/*` は JSON の 404。
- フロント未同梱のバイナリでも API は動き、UI は 503 `frontend_not_built` を返す。開発時は `make dev-frontend`（Vite のプロキシ）を使う。
- **コンテナ**: `Dockerfile` はマルチステージ（`node` でフロント → `golang` でバイナリ → `distroless/static` 非 root）。`HEALTHCHECK` はシェルのない distroless のため、バイナリの `-healthcheck` フラグが自分の `/api/v1/health` を叩く。`compose.yaml` は `app` 1 サービスで、`LLM_PROVIDER` と鍵を環境変数または `backend/.env` から渡す。CI の `image` ジョブが同じ Dockerfile をビルドし、ヘルス・`/`・`sw.js` のヘッダ・`HEALTHCHECK` の healthy をスモークする。

## 横断的関心事

| 関心事 | 方針 |
| --- | --- |
| エラー | API は常に `{"error":{"code","message"}}`。Go 側は `errors.Join` で検証エラーをまとめる |
| ログ | `log/slog` JSON。リクエストログはミドルウェアで 1 行 |
| 設定 | 環境変数のみ。`config.Load(getenv)` で全部読む。既定値で起動できる |
| 秘密情報 | `ANTHROPIC_API_KEY` は環境変数。ログに出さない |
| 認証 | M3 までなし。導入時は httpapi のミドルウェアとして追加 |
| 可観測性 | M4 以降。`slog` のハンドラ差し替えで対応できる形にしておく |

## 将来の検討事項（ADR 化の候補）

- 音楽エンジンの一部をブラウザでも使いたくなった場合（即時フィードバック）: TinyGo/WASM、または TS への限定的な移植。ADR-0002 の再検討条件を参照
- 会話状態のサーバー側保持（現状はステートレスに往復）
- 永続化先（IndexedDB → サーバー）
