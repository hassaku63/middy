# コーディング規約

ツールで強制できるものはツールに任せる（gofmt, golangci-lint, ESLint, Prettier）。ここにはツールで強制できない判断の規約を書く。

## 共通

- 言語: コード・識別子・コードコメントは英語。ドキュメント・UI 文言・コミットの本文は日本語可。コミットの件名は英語（Conventional Commits）。
- 命名はドメイン用語（`docs/glossary.md`）に揃える。`Chord` を `Harmony` と呼び替えない。
- コメントは「何を」ではなく「なぜ」。公開 API には doc コメント必須（Go は revive が強制）。
- マジックナンバーは名前付き定数に（`DefaultPPQ`, `MaxPitch`）。
- エラーは 1 か所で 1 回だけ扱う。扱うとは、記録する、利用者への答え（API の封筒、画面の文言）に変える、別の振る舞いに切り替える、のどれか。呼び出し側へ返すなら記録せず、何をしていたかを足して返す。記録は、失敗を吸収する場所か最後の境界でだけ行う。境界は、Go では `httpapi` のハンドラ・ミドルウェアと `cmd/server/main.go`、フロントでは hook、nginx では自分が答える `error_page`。出典: [Uber Go Style Guide「Handle Errors Once」](https://github.com/uber-go/guide/blob/master/style.md#handle-errors-once)、[Google Go Style Guide「Logging errors」](https://google.github.io/styleguide/go/best-practices.html#logging-errors)、[Dave Cheney「Only handle errors once」](https://dave.cheney.net/2016/04/27/dont-just-check-errors-handle-them-gracefully)（#228）。

## Go

- パッケージは小さく、`internal/` 配下。パッケージ名は単数・小文字（`music`, `httpapi`）。
- エラー: `fmt.Errorf("music: ...")` のようにパッケージ名で始める。複数の検証エラーは `errors.Join`。センチネルは `ErrXxx`、呼び出し側は `errors.Is/As`。
  - 返すときは何をしていたかを足す（`failed to` は書かない）。呼び出し側が分岐に使う原因は `%w` で包む。
  - 下の層の原因を見せたくないとき（プロバイダの型など実装の詳細）は、分類のセンチネルを `%w`、原因を `%v` にする（`fmt.Errorf("%w: %v", ErrLLMUnavailable, err)`）。だから `errorlint` は使わない（ADR-0029、#228）。
  - 出典: [Go blog「Working with Errors in Go 1.13」の「Whether to Wrap」](https://go.dev/blog/go1.13-errors)、[Uber Go Style Guide「Error Wrapping」](https://github.com/uber-go/guide/blob/master/style.md#error-wrapping)、[Go Code Review Comments「Error Strings」](https://go.dev/wiki/CodeReviewComments#error-strings)。
- 値型を優先。`Pitch`, `Key`, `Chord` は比較可能な値。スライスを返す関数はコピーを返し、内部配列を漏らさない。
- テスト: 同じパッケージ内（`package music`）でテーブル駆動。`t.Run` にケース名。`t.Helper()` をヘルパーに。外部テストパッケージ（`music_test`）は公開 API の使用例を示したいときだけ。
- HTTP: ハンドラは `func(w, r)`。JSON の書き出しは `writeJSON`/`writeError` に集約。ルーティングは `"GET /api/v1/..."` パターン。
- ログ: `slog` の構造化フィールド。文字列連結でログを組み立てない。logger は `Deps` から受け取る。エラーの属性名は `error`。秘密情報・ユーザー入力全文を出さない。
  - レベル: ERROR は直すべき自分の不具合（500、panic）、WARN は外部の失敗（上流の 502 など）。利用者の誤り（4xx）は記録せず、アクセスログの 1 行で足りる。
  - 出典: [Google Go Style Guide「Logging errors」](https://google.github.io/styleguide/go/best-practices.html#logging-errors)（"messages at the error level should be actionable"）。
- 設定は `config.Load` だけが環境変数を読む。他の場所で `os.Getenv` を呼ばない。
- 依存注入は `httpapi.Deps` のような構造体で。グローバル変数で共有しない（`httpapi.Version` のようなビルド時定数は例外）。

## TypeScript / React

- `strict`, `noUncheckedIndexedAccess`, `exactOptionalPropertyTypes` を前提に書く。`any` と `as unknown as` は禁止（テストでも）。
- `import type` を使う（ESLint が強制）。
- コンポーネントは関数コンポーネント + 名前付きエクスポート。1 ファイル 1 コンポーネントを基本に、小さな private コンポーネントは同居可。
- 状態は最も近い場所に置く。feature をまたぐ状態は `src/app` の context に昇格させ、3 つを超えたら ADR。
- 副作用（fetch, AudioContext, ストレージ, タイマー）は hook かサービス（`.ts`）に閉じ込め、コンポーネント（`.tsx`）は表示に集中する。ESLint が `*.tsx` での `@/shared/api` の値 import と上記グローバルを禁止する（型 import は可）。hook は引数で依存（`fetchImpl`, `saveFile` など）を受け取れるようにしてテスト可能にする。
- feature 内の参照は `./`、共有は `@/shared`。`../` で隣の feature に届く書き方と `@/features/<他>` は ESLint が禁止する。
- API 呼び出しは `src/shared/api` 経由のみ。コンポーネントから直接 `fetch` しない。
- アクセシビリティ: インタラクティブ要素は `button`/`role` を正しく。テストは `getByRole` で取れることを前提にする。
- CSS: プレーン CSS。グローバルは `src/app/global.css` のみ。feature 固有は CSS Modules（`*.module.css`）。
- 非同期の未処理 Promise 禁止（`no-floating-promises`）。
- API とストレージの失敗は、それを呼ぶ hook の `catch` で画面の状態に変えて終える。`console` に出して投げ直さない。`try` は失敗しうる呼び出しだけを囲み、結果を組み立てる自分の関数は外に置く（`useIdeas.ts`。自分のバグを「〜できませんでした」に変えて隠さないため）。

## API（HTTP/JSON）

- フィールドは camelCase（`tempoBpm`, `romanNumeral`）。
- 列挙はスネークケースではなく、ドメインの識別子（`"major"`, `"dom7"`）。
- 成功は 200/201、検証エラーは 400 + エラーエンベロープ、未知のルートは 404。
- 破壊的変更は `v2`。

## コミットメッセージ

```
<type>(<scope>): <subject in English, imperative>

<body: 日本語可。なぜ・何を。backlog ID を refs BL-NNN の形で>
```

type: `feat`, `fix`, `docs`, `test`, `refactor`, `chore`, `ci`, `build`。scope: `music`, `httpapi`, `llm`, `config`, `frontend`, `docs`, `ci`。
