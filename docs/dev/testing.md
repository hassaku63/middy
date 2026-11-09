# テスト戦略

## 原則

1. **決定論的なものは網羅的に。** 音楽理論エンジンは入力空間が小さい（12 調 × モード × 7 度数）ので、既知の事実を表にして全部確認する。
2. **LLM はテストしない、境界をテストする。** LLM の出力は fake の台本で与え、「その出力をどう検証し、どう使うか」をテストする。実 API を叩くテストは `integration` ビルドタグで分離し、CI では走らせない。
3. **UI はユーザーの視点で。** role / text で要素を取り、スナップショットは使わない（`toMatchSnapshot` は ESLint が禁止する）。
4. **CI はネットワークなしで緑。** 外部サービス依存のテストは書かない。

## レイヤー別

### `backend/internal/music`

- テーブル駆動。ケース名に音楽的な意味を書く（`"A minor sevenths"`）。
- 新しい型を足すときは「JSON ラウンドトリップ」「`Validate` の各失敗ケース」「境界値（degree 0/8、pitch -1/128）」を最低限含める。
- `music/smf` は最小のファイルのバイト列を期待値として固定し、構造は小さな読み戻しパーサで検証する。外部ツール（`mido` 等）での読み戻しは UI スライス同様、手動で 1 回行う。
- パターンのカタログは、描画の結果をゴールデンファイルで固定する（`internal/golden`。`music` は OS に触れないので、テストは外のパッケージに置く。全部の伴奏とドラムを拍子ごとに、全部のスタイルのグルーヴを拍子ごとに、決まった進行で）。`GET /patterns` の応答（`httpapi/testdata/golden/`）とシステムプロンプト（`ideas/testdata/golden/`）も同じく固定する。意図して変えたときは `go test ./internal/golden ./internal/httpapi ./internal/ideas -run Golden -update` で書き直し、差分を PR で見せる（BL-176、#207）。ゴールデンファイルは PR の行数に数えない（`scripts/check-pr-size.sh`）。

### ユースケース（`internal/ideas`）

- `llm/fake` に台本を与え、入力 → 出力の end-to-end を検証する。
- 不正な LLM 出力（スキーマ違反、理論的に不成立）に対して **落ちずにエラーまたは補正** することをテストする。

### 言語またぎ（`backend/internal/fixtures` ↔ `frontend/src/shared/api/fixtures`）

- バックエンドの表記（コード記号、ローマ数字）や分析結果をフロントが写している箇所は、Go が生成した JSON フィクスチャに対してテストする。フィクスチャは `make fixtures` で再生成してコミットし、古くなると Go のテストが落ちる。

### `backend/internal/httpapi`

- `httptest.NewServer(NewHandler(deps))` で実際に HTTP を通す。
- 正常系 1 つ + 検証エラー 1 つ + 未知ルートを最低限。
- レスポンスの JSON を構造体にデコードして検証し、`docs/api/openapi.yaml` と目視で突き合わせる。

### `frontend`（ロジック）

- 純粋関数と hook は Vitest で直接。hook は `renderHook`。
- 時間依存（再生位置）は `vi.useFakeTimers()`。AudioContext はモック。

### `frontend`（UI）

- Testing Library。`findByRole` で非同期を待つ。
- `fetch` は `vi.stubGlobal` で差し替え、`afterEach` で `vi.unstubAllGlobals()`。

### 通しの E2E（Playwright, `frontend/e2e/`）

本番相当のスタック（`make up`。nginx と Go の 2 つのコンテナ）に対して `make e2e` を走らせる。イメージから例を開いて再生・追従・固定・書き出しする通しのシナリオ（fake の LLM 経由）、会話での更新、画面幅ごとの配置（再生中にタイムラインと成り立ちが同時に見える、横スクロールが出ない、ライトテーマでカードが読める）、言い直しを重ねても再生と入力欄が最初の画面に残ること、深いリンクのフォールバックと Service Worker の起動、ローカル保存（再読み込み後の復元）、オフラインでの保存済みプロジェクトの再生を検証する。主な画面（最初の画面、例を開いた画面、コードを固定した解説）には axe-core（`@axe-core/playwright`）を当て、WCAG 2.2 の A・AA の違反があれば落とす（`e2e/a11y.spec.ts`。機械で見えるのは一部で、残りは `docs/design/ux-principles.md` の確かめ方による）。単体テストは `fetch` や `AudioContext` をモックするため、ブラウザ固有の失敗（例: `fetch` を束縛せずに呼ぶと "Illegal invocation"）はここでしか見つからない。**UI に触れる PR では `make e2e` を 1 回通す。** 手元の Chromium を使うときは `E2E_CHROMIUM_PATH` で実行ファイルを指す。 CI では `e2e` ワークフロー（手動と週次）が、compose で起動した 2 つのコンテナに nginx 越しで実行する（ADR-0027）。安定したら PR ごとの実行に昇格させる。

- 検証はロール・ラベル・属性で行い、CSS クラスは解説パネルの文章など role のない要素に限る。
- 時間依存（再生位置）は `expect.poll` / `toHaveText` の待機で吸収し、固定の `sleep` を使わない。
- ブラウザが別の場所にある環境では `E2E_CHROMIUM_PATH` または `PLAYWRIGHT_BROWSERS_PATH` を使う。

## 実行

```bash
make test                                   # 全部
cd backend && go test ./internal/music/ -v  # パッケージ単位
cd backend && go test -run TestDiatonic ./... # ケース単位
cd frontend && npx vitest run src/app       # ディレクトリ単位
```

## 所要時間

`make test-report` が両スイートの遅いテスト上位と合計を出す（CI でも毎回出力）。予算は backend 60 秒、frontend 30 秒で、超えたら警告になる。遅いテストは「価値に見合うか」を問い、分割・高速化・削除のいずれかを backlog に書く。

## カバレッジの目安

- `internal/music`: 90% 以上を維持する（実測値は `make backend-test` の出力で見る。この文書に数値を書かない）。下がる PR は理由を書く。
- その他: 数値目標なし。重要な分岐が網羅されているかをレビューで見る。
