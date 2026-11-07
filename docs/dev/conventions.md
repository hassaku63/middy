# コーディング規約

ツールで強制できるものはツールに任せる（gofmt, golangci-lint, ESLint, Prettier）。ここにはツールで強制できない判断の規約を書く。

## 共通

- 言語: コード・識別子・コードコメントは英語。ドキュメント・UI 文言・コミットの本文は日本語可。コミットの件名は英語（Conventional Commits）。
- 命名はドメイン用語（`docs/glossary.md`）に揃える。`Chord` を `Harmony` と呼び替えない。
- コメントは「何を」ではなく「なぜ」。公開 API には doc コメント必須（Go は revive が強制）。
- マジックナンバーは名前付き定数に（`DefaultPPQ`, `MaxPitch`）。

## Go

- パッケージは小さく、`internal/` 配下。パッケージ名は単数・小文字（`music`, `httpapi`）。
- エラー: `fmt.Errorf("music: ...")` のようにパッケージ名で始める。複数の検証エラーは `errors.Join`。センチネルは `ErrXxx`、呼び出し側は `errors.Is/As`。
- 値型を優先。`Pitch`, `Key`, `Chord` は比較可能な値。スライスを返す関数はコピーを返し、内部配列を漏らさない。
- テスト: 同じパッケージ内（`package music`）でテーブル駆動。`t.Run` にケース名。`t.Helper()` をヘルパーに。外部テストパッケージ（`music_test`）は公開 API の使用例を示したいときだけ。
- HTTP: ハンドラは `func(w, r)`。JSON の書き出しは `writeJSON`/`writeError` に集約。ルーティングは `"GET /api/v1/..."` パターン。
- ログ: `slog` の構造化フィールド。文字列連結でログを組み立てない。秘密情報・ユーザー入力全文を出さない。
- 設定は `config.Load` だけが環境変数を読む。他の場所で `os.Getenv` を呼ばない。
- 依存注入は `httpapi.Deps` のような構造体で。グローバル変数で共有しない（`httpapi.Version` のようなビルド時定数は例外）。

## TypeScript / React

- `strict`, `noUncheckedIndexedAccess`, `exactOptionalPropertyTypes` を前提に書く。`any` と `as unknown as` は禁止（テストでも）。
- `import type` を使う（ESLint が強制）。
- コンポーネントは関数コンポーネント + 名前付きエクスポート。1 ファイル 1 コンポーネントを基本に、小さな private コンポーネントは同居可。
- 状態は最も近い場所に置く。feature をまたぐ状態は `src/app` の context に昇格させ、3 つを超えたら ADR。
- 副作用（fetch, AudioContext）は hook に閉じ込め、コンポーネントは表示に集中する。hook は引数で依存（`fetchImpl` など）を受け取れるようにしてテスト可能にする。
- API 呼び出しは `src/shared/api` 経由のみ。コンポーネントから直接 `fetch` しない。
- アクセシビリティ: インタラクティブ要素は `button`/`role` を正しく。テストは `getByRole` で取れることを前提にする。
- CSS: プレーン CSS。グローバルは `src/app/global.css` のみ。feature 固有は CSS Modules（`*.module.css`）。
- 非同期の未処理 Promise 禁止（`no-floating-promises`）。

## API（HTTP/JSON）

- フィールドは camelCase（`tempoBpm`, `romanNumeral`）。
- 列挙はスネークケースではなく、ドメインの識別子（`"major"`, `"dom7"`）。
- 成功は 200/201、検証エラーは 400 + エラーエンベロープ、未知のルートは 404。
- 破壊的変更は `v2`。

## コミットメッセージ

```
<type>(<scope>): <subject in English, imperative>

<body: 日本語可。なぜ・何を。backlog ID を refs BL-001 の形で>
```

type: `feat`, `fix`, `docs`, `test`, `refactor`, `chore`, `ci`, `build`。scope: `music`, `httpapi`, `llm`, `config`, `frontend`, `docs`, `ci`。
