# Architecture Decision Records

設計上の判断とその理由を記録する。**新しい判断は [`0000-template.md`](0000-template.md) をコピーして次の番号で追加する。** 過去の ADR は書き換えず、覆す場合は新しい ADR で「Supersedes」を書く。

| 番号 | タイトル | 状態 |
| --- | --- | --- |
| [0001](0001-record-architecture-decisions.md) | 設計判断を ADR として記録する | Accepted |
| [0002](0002-tech-stack.md) | Go バックエンド + React/TypeScript フロントエンドのモノレポ | Accepted |
| [0003](0003-llm-interprets-engine-generates.md) | LLM は解釈、理論エンジンが生成 | Accepted |
| [0004](0004-api-contract-openapi.md) | API 契約は OpenAPI を唯一の情報源とする | Accepted |
| [0005](0005-vertical-slices-and-backlog-discipline.md) | 縦スライスとバックログ規律による開発 | Accepted |
| [0006](0006-documentation-freshness.md) | ドキュメントは現在の状態だけを書き、機械的検査と定期点検で鮮度を保つ | Accepted |
| [0007](0007-decay-prevention.md) | 開発サイクルの劣化を、矯正・検知・削減の三層で防ぐ | Accepted |
| [0008](0008-local-verification-and-containers.md) | 本番相当の配信形態を「単一 Go バイナリ」に決め、コンテナで検証可能にする | Accepted |
| [0009](0009-generated-api-types.md) | TypeScript の API 型を OpenAPI から生成する | Accepted |
| [0010](0010-ux-decisions-follow-product-intent.md) | UI の判断は、プロダクトの意思を根拠に共通の原則と手順で行う | Accepted |
| [0011](0011-settings-only-from-launch-environment.md) | ローカル実行の設定と秘密情報は、起動時の環境変数だけで渡す | Accepted |
| [0012](0012-harmony-groove-style.md) | 例を「和声 × グルーヴ」に分け、「スタイル」で結ぶ（付録: [和声](0012/harmony.md)、[リズム](0012/rhythm.md)、[結びつき](0012/combination.md)、[類似アプリ](0012/apps.md)） | Proposed |
| [0013](0013-breadth-of-proposals.md) | 提案の幅は「表せる空間」と「選ぶ振る舞い」に分けて、測ってから広げる | Accepted |
| [0014](0014-issues-in-the-loop.md) | 観察と企画を GitHub Issue に置き、backlog と ADR はそのまま正典に保つ | Accepted |
| [0015](0015-modes.md) | モードを調の属性として足し、特徴音で解説する | Accepted |
| [0016](0016-ci-speed.md) | CI は、アプリと同じ版をキャッシュで速く取り、検査を 1 回ずつ走らせ、差分でジョブを選ぶ | Accepted |

## いつ ADR を書くか

- 依存ライブラリ・フレームワークを追加／置換するとき
- データモデルや API の互換性を壊す変更をするとき
- 複数の選択肢があり、後から「なぜこうしたか」を聞かれそうなとき
- 既存 ADR の前提が崩れたとき
