# docs

middy のドキュメント索引。**新しいドキュメントを追加したらここに 1 行足す。**

## プロダクト（何を・なぜ）

| ファイル | 内容 |
| --- | --- |
| [product/vision.md](product/vision.md) | ビジョン、ペルソナ、コアシナリオ、価値仮説、やらないこと |
| [product/roadmap.md](product/roadmap.md) | マイルストーン（M0〜M4）と各マイルストーンの出口条件 |
| [product/backlog.md](product/backlog.md) | 優先順位付きの機能スライス一覧（Now / Next / Later / Ideas）。AI エージェントはここから着手項目を選ぶ |

## アーキテクチャ（どう作るか）

| ファイル | 内容 |
| --- | --- |
| [architecture/overview.md](architecture/overview.md) | システム構成、境界、ディレクトリ配置、データフロー |
| [architecture/domain-model.md](architecture/domain-model.md) | 音楽ドメインの語彙と型（Pitch, Key, Chord, Timeline, ...） |
| [architecture/llm-integration.md](architecture/llm-integration.md) | LLM の役割、プロンプト/構造化出力の方針、実装メモ |
| [architecture/adr/](architecture/adr/README.md) | Architecture Decision Records |
| [api/openapi.yaml](api/openapi.yaml) | HTTP API 契約（OpenAPI 3.1） |

## 開発（どう進めるか）

| ファイル | 内容 |
| --- | --- |
| [dev/workflow.md](dev/workflow.md) | AI/人間共通の開発ループ、完了の定義、ガードレール、PR の作法 |
| [dev/conventions.md](dev/conventions.md) | Go / TypeScript のコーディング規約、命名、エラー処理、ログ |
| [dev/testing.md](dev/testing.md) | テスト戦略（何をどのレイヤーでテストするか） |
| [dev/docs-hygiene.md](dev/docs-hygiene.md) | ドキュメントを腐らせないための原則、機械的検査、自己点検手順、月次の庭仕事 |
| [glossary.md](glossary.md) | 用語集（音楽用語・プロダクト用語） |
