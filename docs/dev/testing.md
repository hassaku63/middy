# テスト戦略

## 原則

1. **決定論的なものは網羅的に。** 音楽理論エンジンは入力空間が小さい（12 調 × モード × 7 度数）ので、既知の事実を表にして全部確認する。
2. **LLM はテストしない、境界をテストする。** LLM の出力は fake の台本で与え、「その出力をどう検証し、どう使うか」をテストする。実 API を叩くテストは `integration` ビルドタグで分離し、CI では走らせない。
3. **UI はユーザーの視点で。** role / text で要素を取り、スナップショットは使わない。
4. **CI はネットワークなしで緑。** 外部サービス依存のテストは書かない。

## レイヤー別

### `backend/internal/music`

- テーブル駆動。ケース名に音楽的な意味を書く（`"A minor sevenths"`）。
- 新しい型を足すときは「JSON ラウンドトリップ」「`Validate` の各失敗ケース」「境界値（degree 0/8、pitch -1/128）」を最低限含める。
- 将来の SMF エクスポートはゴールデンファイル（`testdata/*.mid`）で検証し、差分は `-update` フラグで更新する。

### ユースケース（M1 以降）

- `llm/fake` に台本を与え、入力 → 出力の end-to-end を検証する。
- 不正な LLM 出力（スキーマ違反、理論的に不成立）に対して **落ちずにエラーまたは補正** することをテストする。

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

## 実行

```bash
make test                                   # 全部
cd backend && go test ./internal/music/ -v  # パッケージ単位
cd backend && go test -run TestDiatonic ./... # ケース単位
cd frontend && npx vitest run src/app       # ディレクトリ単位
cd frontend && npx vitest --coverage        # カバレッジ
```

## カバレッジの目安

- `internal/music`: 90% 以上を維持する（実測値は `make backend-test` の出力で見る。この文書に数値を書かない）。下がる PR は理由を書く。
- その他: 数値目標なし。重要な分岐が網羅されているかをレビューで見る。
