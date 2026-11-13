# docs

middy のドキュメント索引。**新しいドキュメントを追加したらここに 1 行足す。**

## プロダクト（何を・なぜ）

| ファイル | 内容 |
| --- | --- |
| [product/vision.md](product/vision.md) | ビジョン、ペルソナ、コアシナリオと土台となる能力、価値仮説、やらないこと |
| [product/intent.md](product/intent.md) | プロダクトの意思（「X を Y より優先する」）と、中心の体験と入口。UI の判断の拠り所 |
| [product/experience.md](product/experience.md) | 体験の設計: 概念、流れ、画面の約束。UI を変える PR はここを根拠にする |
| [product/roadmap.md](product/roadmap.md) | マイルストーン（M0〜M4）と各マイルストーンの出口条件 |
| [product/backlog.md](product/backlog.md) | 優先順位付きの機能スライス一覧（Now / Next / Later / Ideas）。AI エージェントはここから着手項目を選ぶ |

## アーキテクチャ（どう作るか）

| ファイル | 内容 |
| --- | --- |
| [architecture/overview.md](architecture/overview.md) | システム構成、境界、ディレクトリ配置、データフロー |
| [architecture/domain-model.md](architecture/domain-model.md) | 音楽ドメインの語彙と型（Pitch, Key, Chord, Timeline, ...） |
| [architecture/pattern-catalogue.schema.json](architecture/pattern-catalogue.schema.json) | 型のカタログの JSON Schema（エディタ向け。決まりの正典は `pattern.Catalogue.Validate`） |
| [architecture/llm-integration.md](architecture/llm-integration.md) | LLM の役割、プロンプト/構造化出力の方針、実装メモ |
| [architecture/adr/](architecture/adr/README.md) | Architecture Decision Records |
| [api/openapi.yaml](api/openapi.yaml) | HTTP API 契約（OpenAPI 3.1） |
| [api/problems.md](api/problems.md) | エラー応答（RFC 9457 の problem detail）の種類の一覧。`type` が指す先（ADR-0030） |

## 設計（画面と体験をどう決めるか）

| ファイル | 内容 |
| --- | --- |
| [design/ux-principles.md](design/ux-principles.md) | UI/UX の層（意思・概念・流れ・画面の約束）、意思の書き方、引き算の原則、新しい画面や操作を決める手順 |

## 開発（どう進めるか）

| ファイル | 内容 |
| --- | --- |
| [dev/workflow.md](dev/workflow.md) | AI/人間共通の開発ループ、完了の定義、ガードレール、PR の作法 |
| [dev/conventions.md](dev/conventions.md) | Go / TypeScript のコーディング規約、命名、エラー処理、ログ |
| [dev/testing.md](dev/testing.md) | テスト戦略（何をどのレイヤーでテストするか） |
| [dev/patterns.md](dev/patterns.md) | 型の足し方（伴奏・ドラム・スタイルのカタログの JSON、検査、ゴールデン、評価） |
| [dev/harness.md](dev/harness.md) | 開発ハーネスの全検査の一覧（目的・実行方法・場所）と見直しの規則 |
| [dev/docs-hygiene.md](dev/docs-hygiene.md) | ドキュメントを腐らせないための原則、機械的検査、自己点検手順、月次の庭仕事 |
| [glossary.md](glossary.md) | 用語集（音楽用語・プロダクト用語・画面の部品） |
